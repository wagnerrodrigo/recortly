package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/recortly/recortly/api/internal/store"
)

func (s *Server) handleCreateRender(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	var in struct {
		ClipID       int64  `json:"clipId"`
		TemplateID   *int64 `json:"templateId"`
		VariantCount int    `json:"variantCount"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	clip, err := s.Store.GetClip(r.Context(), u.ID, in.ClipID)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	variants := in.VariantCount
	if variants <= 0 {
		variants = 1
	}
	if variants > 30 {
		writeErr(w, http.StatusBadRequest, "máximo de 30 variações")
		return
	}
	if in.TemplateID != nil {
		if _, err := s.Store.GetTemplate(r.Context(), u.ID, *in.TemplateID); err != nil {
			writeStoreErr(w, err)
			return
		}
	}
	render, err := s.Store.CreateRender(r.Context(), clip.ID, in.TemplateID, variants)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := s.Queue.EnqueueRender(render.ID); err != nil {
		_ = s.Store.SetRenderError(r.Context(), render.ID, "fila indisponível: "+err.Error())
		writeErr(w, http.StatusServiceUnavailable, "fila indisponível")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"render": render})
}

func (s *Server) handleGetRender(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	render, err := s.Store.GetRender(r.Context(), u.ID, id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	for i := range render.Outputs {
		render.Outputs[i].MediaURL = outputURL(render.Outputs[i].ID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"render": render})
}

func (s *Server) handleListRenders(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	list, err := s.Store.ListRenders(r.Context(), u.ID, 50)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i := range list {
		outs, err := s.Store.RenderOutputs(r.Context(), list[i].ID)
		if err == nil {
			for j := range outs {
				outs[j].MediaURL = outputURL(outs[j].ID)
			}
			list[i].Outputs = outs
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"renders": list})
}

// handleRecoverRenders encerra renders presos em "running" (útil se o worker morreu).
func (s *Server) handleRecoverRenders(w http.ResponseWriter, r *http.Request) {
	n, err := s.Store.MarkStaleRendersFailed(r.Context(),
		"render interrompido. Gere novamente.")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"recovered": n})
}

func (s *Server) handleRetryRender(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	old, err := s.Store.GetRender(r.Context(), u.ID, id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	if old.Status == "queued" || old.Status == "running" {
		writeErr(w, http.StatusConflict, "esse render ainda está na fila")
		return
	}
	render, err := s.Store.CreateRender(r.Context(), old.ClipID, old.TemplateID, old.VariantCount)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := s.Queue.EnqueueRender(render.ID); err != nil {
		_ = s.Store.SetRenderError(r.Context(), render.ID, "fila indisponível: "+err.Error())
		writeErr(w, http.StatusServiceUnavailable, "fila indisponível")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"render": render})
}

// handleDeleteRender apaga um render (arquivos + registros) preservando o
// histórico de publicações já concluídas.
func (s *Server) handleDeleteRender(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	files, err := s.Store.DeleteRenders(r.Context(), u.ID, []int64{id})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(files) == 0 {
		writeErr(w, http.StatusConflict, "render em execução ou não encontrado")
		return
	}
	s.removeRenderFiles(files)
	writeJSON(w, http.StatusOK, map[string]any{"deleted": len(files)})
}

// handlePurgeRenders apaga em lote por status e/ou idade.
func (s *Server) handlePurgeRenders(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	var in struct {
		OlderThanDays int      `json:"olderThanDays"`
		Statuses      []string `json:"statuses"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	allowed := map[string]bool{"done": true, "failed": true, "cancelled": true}
	statuses := []string{}
	for _, st := range in.Statuses {
		if allowed[st] {
			statuses = append(statuses, st)
		}
	}
	if len(statuses) == 0 {
		statuses = []string{"failed", "cancelled"}
	}
	ids, err := s.Store.RenderIDsForCleanup(r.Context(), u.ID, statuses, in.OlderThanDays)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	files, err := s.Store.DeleteRenders(r.Context(), u.ID, ids)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.removeRenderFiles(files)
	s.cleanupOrphanDirs(r)
	writeJSON(w, http.StatusOK, map[string]any{"deleted": len(files)})
}

// cleanupOrphanDirs apaga pastas de storage/renders sem registro no banco.
func (s *Server) cleanupOrphanDirs(r *http.Request) {
	existing, err := s.Store.ExistingRenderIDs(r.Context())
	if err != nil {
		return
	}
	entries, err := os.ReadDir(filepath.Join(s.StorageDir, "renders"))
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id, err := strconv.ParseInt(e.Name(), 10, 64)
		if err != nil || existing[id] {
			continue
		}
		_ = os.RemoveAll(filepath.Join(s.StorageDir, "renders", e.Name()))
	}
}

func (s *Server) removeRenderFiles(files []store.RenderFiles) {
	for _, f := range files {
		if f.RenderID > 0 {
			_ = os.RemoveAll(filepath.Join(s.StorageDir, "renders", strconv.FormatInt(f.RenderID, 10)))
		}
		for _, p := range f.Paths {
			_ = os.Remove(p)
		}
	}
}

func (s *Server) handleOutputMedia(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	out, err := s.Store.OutputForUser(r.Context(), u.ID, id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	if _, err := os.Stat(out.FilePath); err != nil {
		writeErr(w, http.StatusNotFound, "arquivo ainda não gerado")
		return
	}
	serveVideoFile(w, r, out.FilePath)
}

func outputURL(id int64) string {
	return "/api/outputs/" + itoa(id) + "/media"
}

func itoa(i int64) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}
