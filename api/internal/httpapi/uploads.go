package httpapi

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxLogoBytes = 8 << 20

// handleUploadLogo recebe a logo usada nos templates.
func (s *Server) handleUploadLogo(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxLogoBytes)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeErr(w, http.StatusBadRequest, "envio inválido: "+err.Error())
		return
	}
	file, header, err := r.FormFile("logo")
	if err != nil {
		file, header, err = r.FormFile("file")
	}
	if err != nil {
		writeErr(w, http.StatusBadRequest, "campo 'logo' é obrigatório")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
	default:
		writeErr(w, http.StatusBadRequest, "use PNG, JPG ou WEBP")
		return
	}

	dir := filepath.Join(s.StorageDir, "logos")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	name := uuid.NewString() + ext
	path := filepath.Join(dir, name)
	out, err := os.Create(path)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		_ = out.Close()
		_ = os.Remove(path)
		writeErr(w, http.StatusInternalServerError, "falha ao salvar a logo")
		return
	}
	_ = out.Close()

	writeJSON(w, http.StatusCreated, map[string]string{
		"path": name,
		"url":  "/api/logos/" + name,
	})
}

// handleLogoMedia serve a logo salva (usada no preview do wizard).
func (s *Server) handleLogoMedia(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(chi.URLParam(r, "name"))
	path := filepath.Join(s.StorageDir, "logos", name)
	f, err := os.Open(path)
	if err != nil {
		writeErr(w, http.StatusNotFound, "logo não encontrada")
		return
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
}
