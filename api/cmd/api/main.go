package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/smashcut/smashcut/api/internal/config"
	"github.com/smashcut/smashcut/api/internal/httpapi"
	"github.com/smashcut/smashcut/api/internal/jobs"
	"github.com/smashcut/smashcut/api/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	cfg := config.Load()
	ctx := context.Background()

	if err := store.MustStorage(cfg.StorageDir); err != nil {
		slog.Error("storage", "err", err)
		os.Exit(1)
	}
	if err := store.Migrate(ctx, cfg.DatabaseURL); err != nil {
		slog.Error("migração", "err", err)
		os.Exit(1)
	}
	st, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("banco", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	seedUsers(ctx, st, cfg)

	queue := jobs.NewEnqueuer(cfg.RedisAddr)
	defer queue.Close()

	srv := &httpapi.Server{Cfg: cfg, Store: st, Queue: queue, StorageDir: cfg.StorageDir}
	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      0, // uploads e streams longos
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		slog.Info("api no ar", "addr", cfg.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	slog.Info("api encerrada")
}

func seedUsers(ctx context.Context, st *store.Store, cfg config.Config) {
	for _, u := range cfg.SeedUsers {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("seed bcrypt", "err", err)
			continue
		}
		if _, err := st.CreateUser(ctx, u.Name, u.Email, string(hash)); err != nil {
			slog.Warn("seed usuário", "email", u.Email, "err", err)
		}
	}
}
