package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/smashcut/smashcut/api/internal/model"
	"github.com/smashcut/smashcut/api/internal/publishers"
	"github.com/smashcut/smashcut/api/internal/store"
	"github.com/smashcut/smashcut/api/internal/video"
	"github.com/smashcut/smashcut/api/internal/video/template"
)

type Worker struct {
	Store      *store.Store
	StorageDir string
	FontPath   string
	RedisAddr  string
}

func (w *Worker) Server(concurrency int) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: w.RedisAddr},
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				QueueRenders:     6,
				QueuePublish:     3,
				QueueMaintenance: 1,
			},
		},
	)
}

// Recover encerra renders que ficaram presos em "running" e reenfileira os
// que estavam na fila. Deve rodar uma única vez, no boot do worker.
func (w *Worker) Recover(ctx context.Context, queue *Enqueuer) {
	n, err := w.Store.MarkStaleRendersFailed(ctx,
		"render interrompido (worker reiniciado). Gere novamente.")
	if err != nil {
		slog.Error("recuperar renders", "err", err)
	} else if n > 0 {
		slog.Warn("renders interrompidos marcados como falha", "total", n)
	}

	ids, err := w.Store.QueuedRenders(ctx)
	if err != nil {
		slog.Error("renders na fila", "err", err)
		return
	}
	for _, id := range ids {
		if err := queue.EnqueueRender(id); err != nil {
			slog.Warn("reenfileirar render", "render", id, "err", err)
		}
	}
	if len(ids) > 0 {
		slog.Info("renders reenfileirados", "total", len(ids))
	}
}

func (w *Worker) Mux() *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeRender, recoverTask("render", w.handleRender))
	mux.HandleFunc(TypePublish, recoverTask("publish", w.handlePublish))
	mux.HandleFunc(TypeCleanup, recoverTask("cleanup", w.handleCleanup))
	return mux
}

// recoverTask evita que um panic derrube o worker sem diagnóstico.
func recoverTask(kind string, fn func(context.Context, *asynq.Task) error) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) (err error) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic na task", "tipo", kind, "err", rec, "stack", string(debug.Stack()))
				err = fmt.Errorf("panic: %v", rec)
			}
		}()
		return fn(ctx, t)
	}
}

func (w *Worker) handleRender(ctx context.Context, t *asynq.Task) error {
	var p RenderPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("payload inválido: %w", err)
	}
	r, err := w.Store.RenderForWorker(ctx, p.RenderID)
	if err != nil {
		slog.Error("render não encontrado", "render", p.RenderID, "err", err)
		return asynq.SkipRetry
	}
	slog.Info("processando render", "render", r.ID, "clip", r.ClipID, "status", r.Status)
	if r.Status == "cancelled" || r.Status == "done" {
		return nil
	}

	clip, err := w.Store.GetClipForWorker(ctx, r.ClipID)
	if err != nil {
		slog.Error("clip", "render", r.ID, "clip", r.ClipID, "err", err)
		_ = w.Store.SetRenderError(ctx, r.ID, "clip não encontrado")
		return asynq.SkipRetry
	}
	source, err := w.Store.SourceForClip(ctx, clip.ID)
	if err != nil {
		slog.Error("source", "render", r.ID, "clip", clip.ID, "err", err)
		_ = w.Store.SetRenderError(ctx, r.ID, "vídeo de origem não encontrado")
		return asynq.SkipRetry
	}
	slog.Info("fonte pronta", "render", r.ID, "path", source.FilePath, "audio", source.HasAudio)

	spec := template.Default()
	if r.TemplateID != nil {
		tpl, err := w.Store.TemplateForWorker(ctx, *r.TemplateID)
		if err != nil {
			_ = w.Store.SetRenderError(ctx, r.ID, "template não encontrado")
			return asynq.SkipRetry
		}
		spec = tpl.Spec
	}
	spec.Normalize()
	if err := template.Validate(spec); err != nil {
		_ = w.Store.SetRenderError(ctx, r.ID, err.Error())
		return asynq.SkipRetry
	}

	_ = w.Store.SetRenderStatus(ctx, r.ID, "running", 0)
	if err := w.Store.DeleteRenderOutputs(ctx, r.ID); err != nil {
		return err
	}

	// Fonte VP9/AV1/4K: converte só o trecho do corte para H.264 1080p.
	// Fica em cache por source+corte, então as variações reaproveitam.
	sourcePath := source.FilePath
	renderClip := clip
	if info, err := video.Probe(ctx, source.FilePath); err == nil {
		if video.NeedsProxy(info) {
			key := fmt.Sprintf("%d-%d-%d", source.ID, clip.StartMS, clip.EndMS)
			proxyPath := filepath.Join(w.StorageDir, "proxies", key+".mp4")
			proxyCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
			err := video.Proxy(proxyCtx, source.FilePath, proxyPath, clip.StartMS, clip.EndMS, info.HasAudio, func(pct int) {
				_ = w.Store.SetRenderStatus(context.WithoutCancel(ctx), r.ID, "running", pct/10)
			})
			cancel()
			if err != nil {
				slog.Error("proxy", "source", source.ID, "err", err)
				_ = w.Store.SetRenderError(context.WithoutCancel(ctx), r.ID, "preparação do vídeo falhou: "+err.Error())
				return asynq.SkipRetry
			}
			sourcePath = proxyPath
			source.HasAudio = true
			renderClip = model.Clip{
				ID: clip.ID, SourceID: clip.SourceID, Layout: clip.Layout,
				Zoom: clip.Zoom, OffsetX: clip.OffsetX, OffsetY: clip.OffsetY,
				StartMS: 0, EndMS: clip.EndMS - clip.StartMS,
			}
		}
	}

	plans := video.Plan(spec, r.Seed)
	if len(plans) > r.VariantCount && r.VariantCount > 0 {
		plans = plans[:r.VariantCount]
	}

	outDir := filepath.Join(w.StorageDir, "renders", fmt.Sprintf("%d", r.ID))
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		_ = w.Store.SetRenderError(ctx, r.ID, err.Error())
		return err
	}

	for i, plan := range plans {
		select {
		case <-ctx.Done():
			_ = w.Store.SetRenderStatus(ctx, r.ID, "cancelled", 0)
			return nil
		default:
		}

		outPath := filepath.Join(outDir, fmt.Sprintf("variant-%02d.mp4", i+1))
		logoPath := ""
		if spec.Logo.Enabled && spec.Logo.Path != "" {
			logoPath = filepath.Join(w.StorageDir, "logos", filepath.Base(spec.Logo.Path))
		}
		in := video.RenderInput{
			SourcePath: sourcePath,
			LogoPath:   logoPath,
			OutPath:    outPath,
			Clip:       renderClip,
			Spec:       spec,
			Plan:       plan,
			HasAudio:   source.HasAudio,
			FontPath:   w.FontPath,
		}
		base := 10 + i*90/len(plans)
		var lastWrite time.Time
		err := video.Render(ctx, in, func(pct int) {
			overall := base + pct*90/100/len(plans)
			if time.Since(lastWrite) < 1500*time.Millisecond && overall < 100 {
				return
			}
			lastWrite = time.Now()
			_ = w.Store.SetRenderStatus(context.WithoutCancel(ctx), r.ID, "running", overall)
		})
		if err != nil {
			slog.Error("ffmpeg falhou", "render", r.ID, "variante", i, "err", err)
			if errors.Is(err, context.Canceled) {
				_ = w.Store.SetRenderStatus(context.WithoutCancel(ctx), r.ID, "cancelled", 0)
				return nil
			}
			_ = w.Store.SetRenderError(context.WithoutCancel(ctx), r.ID, err.Error())
			return asynq.SkipRetry
		}
		if _, err := w.Store.AddRenderOutput(ctx, model.RenderOutput{
			RenderID:   r.ID,
			VariantIdx: i,
			Headline:   plan.Headline,
			FilePath:   outPath,
			DurationMS: clip.DurationMS(),
		}); err != nil {
			slog.Error("salvar output", "render", r.ID, "variante", i, "err", err)
			return err
		}
	}

	_ = w.Store.SetRenderStatus(ctx, r.ID, "done", 100)
	slog.Info("render concluído", "render", r.ID, "variacoes", len(plans))
	return nil
}

func (w *Worker) handlePublish(ctx context.Context, t *asynq.Task) error {
	var p PublishPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("payload inválido: %w", err)
	}
	post, source, output, err := w.Store.PostForWorker(ctx, p.PostID)
	if err != nil {
		slog.Error("post não encontrado", "post", p.PostID, "err", err)
		return asynq.SkipRetry
	}
	slog.Info("publicando", "post", post.ID, "plataforma", post.Platform, "status", post.Status)
	if post.Status == "published" {
		return nil
	}
	if !source.CanPublish() {
		slog.Warn("publicação bloqueada por licença", "post", post.ID)
		_ = w.Store.SetPostStatus(context.WithoutCancel(ctx), post.ID, "failed", "origem e licença não preenchidas", "", "")
		return asynq.SkipRetry
	}

	title := post.Title
	if title == "" {
		title = output.Headline
	}
	desc := post.Description
	credit := creditLine(source)
	if credit != "" && !contains(desc, credit) {
		if desc != "" {
			desc += "\n\n"
		}
		desc += credit
	}

	_ = w.Store.SetPostStatus(context.WithoutCancel(ctx), post.ID, "publishing", "", "", "")
	pub, err := w.publisherFor(ctx, post.Platform, source.UserID)
	if err != nil {
		_ = w.Store.SetPostStatus(context.WithoutCancel(ctx), post.ID, "failed", err.Error(), "", "")
		return asynq.SkipRetry
	}
	res, err := pub.Publish(ctx, publishers.PublishInput{
		VideoPath:   output.FilePath,
		Title:       title,
		Description: desc,
	})
	if err != nil {
		slog.Error("falha ao publicar", "post", post.ID, "err", err)
		if setErr := w.Store.SetPostStatus(context.WithoutCancel(ctx), post.ID, "failed", err.Error(), "", ""); setErr != nil {
			slog.Error("salvar status do post", "post", post.ID, "err", setErr)
		}
		return nil
	}
	slog.Info("publicado", "post", post.ID, "url", res.ExternalURL)
	_ = w.Store.SetPostStatus(context.WithoutCancel(ctx), post.ID, "published", "", res.ExternalID, res.ExternalURL)
	return nil
}

func (w *Worker) publisherFor(ctx context.Context, platform string, userID int64) (publishers.Publisher, error) {
	switch platform {
	case "youtube":
		_, token, err := w.Store.Account(ctx, userID, "youtube")
		if err != nil || token == "" {
			return publishers.Unsupported{Name: "youtube", Reason: "conecte sua conta em Integrações"}, nil
		}
		return publishers.YouTube{AccessToken: token}, nil
	case "instagram":
		return publishers.Unsupported{Name: "instagram", Reason: "configure a Meta Graph API e o app aprovado"}, nil
	case "tiktok":
		return publishers.Unsupported{Name: "tiktok", Reason: "configure a Content Posting API e a auditoria do app"}, nil
	}
	return nil, fmt.Errorf("plataforma desconhecida: %s", platform)
}

func (w *Worker) handleCleanup(ctx context.Context, _ *asynq.Task) error {
	deadline := time.Now().Add(-7 * 24 * time.Hour)
	dir := filepath.Join(w.StorageDir, "tmp")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	removed := 0
	for _, e := range entries {
		info, err := e.Info()
		if err != nil || info.ModTime().After(deadline) {
			continue
		}
		_ = os.RemoveAll(filepath.Join(dir, e.Name()))
		removed++
	}
	slog.Info("limpeza concluída", "arquivos", removed)
	return nil
}

func creditLine(s model.Source) string {
	if s.AuthorName == "" && s.AuthorHandle == "" && s.OriginalURL == "" {
		return ""
	}
	parts := s.AuthorName
	if s.AuthorHandle != "" {
		parts += " (" + s.AuthorHandle + ")"
	}
	if s.OriginalURL != "" {
		parts += " · " + s.OriginalURL
	}
	return "Crédito: " + parts
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}
