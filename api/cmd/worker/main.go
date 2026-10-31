package main

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/hibiken/asynq"
	"github.com/smashcut/smashcut/api/internal/config"
	"github.com/smashcut/smashcut/api/internal/jobs"
	"github.com/smashcut/smashcut/api/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	cfg := config.Load()

	if err := store.MustStorage(cfg.StorageDir); err != nil {
		slog.Error("storage", "err", err)
		os.Exit(1)
	}
	st, err := store.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("banco", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		slog.Error("ffmpeg não encontrado no PATH")
		os.Exit(1)
	}

	w := &jobs.Worker{
		Store:      st,
		StorageDir: cfg.StorageDir,
		FontPath:   findFont(),
		RedisAddr:  cfg.RedisAddr,
	}
	// Garante que só exista um worker por vez. Dois workers em paralelo
	// causavam renders duplicados e travados.
	lock, err := acquireLock(cfg.StorageDir)
	if err != nil {
		slog.Error("já existe um worker rodando", "err", err)
		os.Exit(1)
	}
	defer lock.Release()

	queue := jobs.NewEnqueuer(cfg.RedisAddr)
	defer queue.Close()
	w.Recover(context.Background(), queue)

	srv := w.Server(max(1, cfg.RenderConcurrency))

	scheduler := asynq.NewScheduler(asynq.RedisClientOpt{Addr: cfg.RedisAddr}, nil)
	_, err = scheduler.Register("@daily", asynq.NewTask(jobs.TypeCleanup, nil), asynq.Queue(jobs.QueueMaintenance))
	if err != nil {
		slog.Warn("scheduler", "err", err)
	}
	go func() {
		if err := scheduler.Run(); err != nil {
			slog.Error("scheduler", "err", err)
		}
	}()

	slog.Info("worker no ar", "redis", cfg.RedisAddr, "concorrencia", cfg.RenderConcurrency)
	if err := srv.Run(w.Mux()); err != nil {
		slog.Error("worker", "err", err)
		os.Exit(1)
	}
	_ = time.Second
}

func findFont() string {
	candidates := []string{
		"/System/Library/Fonts/Helvetica.ttc",
		"/System/Library/Fonts/Supplemental/Arial Bold.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
		"/usr/share/fonts/dejavu/DejaVuSans-Bold.ttf",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
