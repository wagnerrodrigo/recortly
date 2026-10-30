package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/smashcut/smashcut/api/internal/config"
	"github.com/smashcut/smashcut/api/internal/jobs"
	"github.com/smashcut/smashcut/api/internal/model"
	"github.com/smashcut/smashcut/api/internal/store"
)

type Server struct {
	Cfg       config.Config
	Store     *store.Store
	Queue     *jobs.Enqueuer
	StorageDir string
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(recoverer, requestLogger)

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Post("/api/auth/login", s.handleLogin)
	r.Post("/api/auth/logout", s.handleLogout)

	// Callback do OAuth não pode exigir cookie (vem do Google), valida state.
	r.Get("/api/oauth/youtube/callback", s.handleYouTubeCallback)

	r.Group(func(r chi.Router) {
		r.Use(s.requireAuth)

		r.Get("/api/auth/me", s.handleMe)

		r.Get("/api/sources", s.handleListSources)
		r.Post("/api/sources", s.handleUploadSource)
		r.Get("/api/sources/{id}", s.handleGetSource)
		r.Patch("/api/sources/{id}", s.handleUpdateSource)
		r.Delete("/api/sources/{id}", s.handleDeleteSource)
		r.Get("/api/sources/{id}/media", s.handleSourceMedia)

		r.Post("/api/clips", s.handleCreateClip)
		r.Delete("/api/clips/{id}", s.handleDeleteClip)

		r.Get("/api/templates", s.handleListTemplates)
		r.Post("/api/templates", s.handleCreateTemplate)
		r.Get("/api/templates/{id}", s.handleGetTemplate)
		r.Put("/api/templates/{id}", s.handleUpdateTemplate)
		r.Delete("/api/templates/{id}", s.handleDeleteTemplate)

		r.Get("/api/renders", s.handleListRenders)
		r.Post("/api/renders", s.handleCreateRender)
		r.Get("/api/renders/{id}", s.handleGetRender)
		r.Get("/api/outputs/{id}/media", s.handleOutputMedia)

		r.Get("/api/posts", s.handleListPosts)
		r.Post("/api/posts", s.handleCreatePost)
		r.Post("/api/posts/{id}/publish", s.handlePublishPost)

		r.Get("/api/accounts", s.handleListAccounts)
		r.Get("/api/oauth/youtube/start", s.handleYouTubeStart)
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func contextWithUser(r *http.Request, u model.User) context.Context {
	return context.WithValue(r.Context(), userKey, u)
}

func urlID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	return id, err == nil && id > 0
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic", "err", rec, "path", r.URL.Path)
				writeErr(w, http.StatusInternalServerError, "erro interno")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("http", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

var errBadInput = errors.New("entrada inválida")
