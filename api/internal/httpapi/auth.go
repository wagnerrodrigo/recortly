package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/smashcut/smashcut/api/internal/model"
	"github.com/smashcut/smashcut/api/internal/store"
)

const sessionCookie = "sc_session"

type ctxKey string

const userKey ctxKey = "user"

func userFrom(r *http.Request) model.User {
	u, _ := r.Context().Value(userKey).(model.User)
	return u
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	user, err := s.Store.UserByEmail(r.Context(), strings.ToLower(strings.TrimSpace(in.Email)))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)) != nil {
		writeErr(w, http.StatusUnauthorized, "email ou senha inválidos")
		return
	}
	token, err := newToken()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "erro ao criar sessão")
		return
	}
	ttl := time.Duration(s.Cfg.SessionTTLHours) * time.Hour
	if err := s.Store.CreateSession(r.Context(), hashToken(token), user.ID, ttl); err != nil {
		writeErr(w, http.StatusInternalServerError, "erro ao criar sessão")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(ttl),
	})
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		_ = s.Store.DeleteSession(r.Context(), hashToken(c.Value))
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"user": userFrom(r)})
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookie)
		if err != nil || c.Value == "" {
			writeErr(w, http.StatusUnauthorized, "não autenticado")
			return
		}
		user, err := s.Store.UserBySession(r.Context(), hashToken(c.Value))
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeErr(w, http.StatusUnauthorized, "sessão expirada")
				return
			}
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		next.ServeHTTP(w, r.WithContext(contextWithUser(r, user)))
	})
}
