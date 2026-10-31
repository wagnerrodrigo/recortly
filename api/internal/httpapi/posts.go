package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/recortly/recortly/api/internal/model"
)

func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	var in struct {
		OutputID    int64   `json:"outputId"`
		Platform    string  `json:"platform"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		ScheduledAt *string `json:"scheduledAt"`
		PublishNow  bool    `json:"publishNow"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := s.Store.OutputForUser(r.Context(), u.ID, in.OutputID)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	src, err := s.Store.SourceForOutput(r.Context(), u.ID, out.ID)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	if !src.CanPublish() {
		writeErr(w, http.StatusUnprocessableEntity,
			"preencha autor, link e licença do vídeo de origem antes de publicar")
		return
	}
	switch in.Platform {
	case "youtube", "instagram", "tiktok":
	default:
		writeErr(w, http.StatusBadRequest, "plataforma inválida")
		return
	}

	post := model.Post{
		OutputID:    out.ID,
		Platform:    in.Platform,
		Title:       strings.TrimSpace(in.Title),
		Description: strings.TrimSpace(in.Description),
		Status:      "draft",
	}
	var schedule *time.Time
	if in.ScheduledAt != nil && strings.TrimSpace(*in.ScheduledAt) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*in.ScheduledAt))
		if err != nil {
			writeErr(w, http.StatusBadRequest, "scheduledAt deve estar em RFC3339")
			return
		}
		schedule = &t
		post.ScheduledAt = schedule
		post.Status = "scheduled"
	}
	created, err := s.Store.CreatePost(r.Context(), post)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	if in.PublishNow {
		created.ScheduledAt = nil
	}
	if in.PublishNow || schedule != nil {
		at := time.Now()
		if schedule != nil {
			at = *schedule
		}
		if err := s.Queue.EnqueuePublish(created.ID, at); err != nil {
			writeErr(w, http.StatusServiceUnavailable, "fila indisponível")
			return
		}
	}
	writeJSON(w, http.StatusCreated, map[string]any{"post": created})
}

func (s *Server) handlePublishPost(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	post, err := s.Store.GetPost(r.Context(), id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	if post.Status == "published" || post.Status == "publishing" {
		writeErr(w, http.StatusConflict, "publicação já em andamento ou concluída")
		return
	}
	if err := s.Queue.EnqueuePublish(id, time.Now()); err != nil {
		writeErr(w, http.StatusServiceUnavailable, "fila indisponível")
		return
	}
	_ = s.Store.SetPostStatus(r.Context(), id, "scheduled", "", "", "")
	_ = u
	writeJSON(w, http.StatusAccepted, map[string]bool{"queued": true})
}

func (s *Server) handleListPosts(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	list, err := s.Store.ListPosts(r.Context(), u.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"posts": list})
}

func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	list, err := s.Store.ListAccounts(r.Context(), u.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"accounts": list})
}

// --- YouTube OAuth ---

func (s *Server) signState(userID int64) string {
	return s.signStateWithExpiry(userID, time.Now().Add(10*time.Minute))
}

func (s *Server) signStateWithExpiry(userID int64, exp time.Time) string {
	raw := fmt.Sprintf("%d:%d", userID, exp.Unix())
	mac := hmac.New(sha256.New, []byte(s.Cfg.TokenKey))
	mac.Write([]byte(raw))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(raw)) + "." + sig
}

func (s *Server) verifyState(state string) (int64, bool) {
	parts := strings.Split(state, ".")
	if len(parts) != 2 {
		return 0, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, false
	}
	mac := hmac.New(sha256.New, []byte(s.Cfg.TokenKey))
	mac.Write(raw)
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return 0, false
	}
	fields := strings.Split(string(raw), ":")
	if len(fields) != 2 {
		return 0, false
	}
	uid, _ := strconv.ParseInt(fields[0], 10, 64)
	exp, _ := strconv.ParseInt(fields[1], 10, 64)
	if time.Now().Unix() > exp {
		return 0, false
	}
	return uid, true
}

func (s *Server) handleYouTubeStart(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	if s.Cfg.YouTubeClientID == "" || s.Cfg.YouTubeClientSecret == "" {
		writeErr(w, http.StatusNotImplemented,
			"configure RECORTLY_YOUTUBE_CLIENT_ID e RECORTLY_YOUTUBE_CLIENT_SECRET")
		return
	}
	q := url.Values{}
	q.Set("client_id", s.Cfg.YouTubeClientID)
	q.Set("redirect_uri", s.Cfg.YouTubeRedirect)
	q.Set("response_type", "code")
	q.Set("access_type", "offline")
	q.Set("prompt", "consent")
	q.Set("scope", "https://www.googleapis.com/auth/youtube.upload")
	q.Set("state", s.signState(u.ID))
	writeJSON(w, http.StatusOK, map[string]string{
		"url": "https://accounts.google.com/o/oauth2/v2/auth?" + q.Encode(),
	})
}

func (s *Server) handleYouTubeCallback(w http.ResponseWriter, r *http.Request) {
	if s.Cfg.YouTubeClientID == "" {
		writeErr(w, http.StatusNotImplemented, "YouTube não configurado")
		return
	}
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		writeErr(w, http.StatusBadRequest, "autorização negada: "+errParam)
		return
	}
	uid, ok := s.verifyState(r.URL.Query().Get("state"))
	if !ok {
		writeErr(w, http.StatusBadRequest, "state inválido ou expirado")
		return
	}
	code := r.URL.Query().Get("code")
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", s.Cfg.YouTubeClientID)
	form.Set("client_secret", s.Cfg.YouTubeClientSecret)
	form.Set("redirect_uri", s.Cfg.YouTubeRedirect)
	form.Set("grant_type", "authorization_code")

	req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost,
		"https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode >= 300 {
		writeErr(w, http.StatusBadGateway, "google: "+string(body))
		return
	}
	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		writeErr(w, http.StatusBadGateway, "resposta inválida do google")
		return
	}
	var expires *time.Time
	if tok.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
		expires = &t
	}
	if err := s.Store.UpsertAccount(r.Context(), uid, "youtube", "YouTube", tok.AccessToken, tok.RefreshToken, expires); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.Store.Log(r.Context(), &uid, "account.connect", map[string]any{"platform": "youtube"})
	http.Redirect(w, r, "/integracoes?connected=youtube", http.StatusFound)
}

// outputClipID resolve o clip de um render (usado para validar licença).
