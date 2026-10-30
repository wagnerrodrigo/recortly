package httpapi

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/smashcut/smashcut/api/internal/model"
	"github.com/smashcut/smashcut/api/internal/store"
	"github.com/smashcut/smashcut/api/internal/video"
)

const maxUploadBytes = 2 << 30 // 2 GiB

func (s *Server) handleListSources(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	list, err := s.Store.ListSources(r.Context(), u.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i := range list {
		list[i].MediaURL = fmt.Sprintf("/api/sources/%d/media", list[i].ID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"sources": list})
}

func (s *Server) handleUploadSource(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeErr(w, http.StatusBadRequest, "envio inválido: "+err.Error())
		return
	}
	file, header, err := r.FormFile("video")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "campo 'video' é obrigatório")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".mp4"
	}
	name := uuid.NewString() + ext
	dest := filepath.Join(s.StorageDir, "sources", name)
	out, err := os.Create(dest)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	size, err := io.Copy(out, file)
	closeErr := out.Close()
	if err != nil || closeErr != nil {
		_ = os.Remove(dest)
		writeErr(w, http.StatusInternalServerError, "falha ao salvar arquivo")
		return
	}

	info, err := video.Probe(r.Context(), dest)
	if err != nil {
		_ = os.Remove(dest)
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		title = strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	}
	src := model.Source{
		UserID:       u.ID,
		Title:        title,
		FilePath:     dest,
		OriginalName: header.Filename,
		SizeBytes:    size,
		DurationMS:   info.DurationMS,
		Width:        info.Width,
		Height:       info.Height,
		HasAudio:     info.HasAudio,
	}
	created, err := s.Store.CreateSource(r.Context(), src)
	if err != nil {
		_ = os.Remove(dest)
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	created.MediaURL = fmt.Sprintf("/api/sources/%d/media", created.ID)
	s.Store.Log(r.Context(), &u.ID, "source.upload", map[string]any{"sourceId": created.ID, "name": header.Filename})
	writeJSON(w, http.StatusCreated, map[string]any{"source": created})
}

func (s *Server) handleGetSource(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	src, err := s.Store.GetSource(r.Context(), u.ID, id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	clips, err := s.Store.ListClips(r.Context(), src.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	src.Clips = clips
	src.MediaURL = fmt.Sprintf("/api/sources/%d/media", src.ID)
	writeJSON(w, http.StatusOK, map[string]any{"source": src})
}

func (s *Server) handleUpdateSource(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	src, err := s.Store.GetSource(r.Context(), u.ID, id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	var in struct {
		Title           *string `json:"title"`
		AuthorName      *string `json:"authorName"`
		AuthorHandle    *string `json:"authorHandle"`
		OriginalURL     *string `json:"originalUrl"`
		LicenseType     *string `json:"licenseType"`
		PermissionProof *string `json:"permissionProof"`
		Notes           *string `json:"notes"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if in.Title != nil {
		src.Title = *in.Title
	}
	if in.AuthorName != nil {
		src.AuthorName = *in.AuthorName
	}
	if in.AuthorHandle != nil {
		src.AuthorHandle = *in.AuthorHandle
	}
	if in.OriginalURL != nil {
		src.OriginalURL = *in.OriginalURL
	}
	if in.LicenseType != nil {
		lt := strings.TrimSpace(*in.LicenseType)
		switch lt {
		case "":
			src.LicenseType = nil
		case "own", "permission", "credit":
			src.LicenseType = &lt
		default:
			writeErr(w, http.StatusBadRequest, "licenseType deve ser own, permission ou credit")
			return
		}
	}
	if in.PermissionProof != nil {
		src.PermissionProof = *in.PermissionProof
	}
	if in.Notes != nil {
		src.Notes = *in.Notes
	}
	if err := s.Store.UpdateSourceMeta(r.Context(), u.ID, src.ID, src); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, _ := s.Store.GetSource(r.Context(), u.ID, src.ID)
	updated.MediaURL = fmt.Sprintf("/api/sources/%d/media", updated.ID)
	writeJSON(w, http.StatusOK, map[string]any{"source": updated})
}

func (s *Server) handleDeleteSource(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	src, err := s.Store.GetSource(r.Context(), u.ID, id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	if err := s.Store.DeleteSource(r.Context(), u.ID, id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = os.Remove(src.FilePath)
	s.Store.Log(r.Context(), &u.ID, "source.delete", map[string]any{"sourceId": id})
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (s *Server) handleSourceMedia(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	src, err := s.Store.GetSource(r.Context(), u.ID, id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	serveVideoFile(w, r, src.FilePath)
}

func (s *Server) handleCreateClip(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	var in struct {
		SourceID int64   `json:"sourceId"`
		Name     string  `json:"name"`
		StartMS  int64   `json:"startMs"`
		EndMS    int64   `json:"endMs"`
		Layout   string  `json:"layout"`
		Zoom     float64 `json:"zoom"`
		OffsetX  *float64 `json:"offsetX"`
		OffsetY  *float64 `json:"offsetY"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	src, err := s.Store.GetSource(r.Context(), u.ID, in.SourceID)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	if in.StartMS < 0 || in.EndMS <= in.StartMS {
		writeErr(w, http.StatusBadRequest, "intervalo inválido")
		return
	}
	if in.EndMS > src.DurationMS+500 {
		writeErr(w, http.StatusBadRequest, "o fim do corte passa da duração do vídeo")
		return
	}
	switch in.Layout {
	case "", "blur", "fill":
	default:
		writeErr(w, http.StatusBadRequest, "layout deve ser blur ou fill")
		return
	}
	layout := in.Layout
	if layout == "" {
		layout = "blur"
	}
	if in.Zoom <= 0 {
		in.Zoom = 1
	}
	offsetX, offsetY := 0.5, 0.5
	if in.OffsetX != nil {
		offsetX = *in.OffsetX
	}
	if in.OffsetY != nil {
		offsetY = *in.OffsetY
	}
	created, err := s.Store.CreateClip(r.Context(), model.Clip{
		SourceID: src.ID, Name: in.Name, StartMS: in.StartMS, EndMS: in.EndMS,
		Layout: layout, Zoom: in.Zoom, OffsetX: offsetX, OffsetY: offsetY,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"clip": created})
}

func (s *Server) handleDeleteClip(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := s.Store.DeleteClip(r.Context(), u.ID, id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func serveVideoFile(w http.ResponseWriter, r *http.Request, path string) {
	f, err := os.Open(path)
	if err != nil {
		writeErr(w, http.StatusNotFound, "arquivo não encontrado")
		return
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		ctype = "video/mp4"
	}
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Cache-Control", "private, max-age=3600")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
}

func writeStoreErr(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "não encontrado")
		return
	}
	writeErr(w, http.StatusInternalServerError, err.Error())
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	dec := jsonNewDecoder(r)
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("json inválido: %w", err)
	}
	return nil
}
