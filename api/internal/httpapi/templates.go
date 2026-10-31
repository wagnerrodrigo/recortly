package httpapi

import (
	"net/http"
	"strings"

	"github.com/recortly/recortly/api/internal/video"
)

func (s *Server) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	list, err := s.Store.ListTemplates(r.Context(), u.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"templates": list})
}

func (s *Server) handleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	var in struct {
		Name string         `json:"name"`
		Spec *video.TemplateSpec `json:"spec"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	spec := video.TemplateDefault()
	if in.Spec != nil {
		spec = *in.Spec
	}
	spec.Normalize()
	if err := video.ValidateTemplate(spec); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = spec.Name
	}
	if name == "" {
		name = "Novo template"
	}
	spec.Name = name
	created, err := s.Store.CreateTemplate(r.Context(), u.ID, name, spec)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"template": created})
}

func (s *Server) handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	tpl, err := s.Store.GetTemplate(r.Context(), u.ID, id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"template": tpl})
}

func (s *Server) handleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	var in struct {
		Name string         `json:"name"`
		Spec *video.TemplateSpec `json:"spec"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	spec := video.TemplateDefault()
	if in.Spec != nil {
		spec = *in.Spec
	}
	spec.Normalize()
	if err := video.ValidateTemplate(spec); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = spec.Name
	}
	if name == "" {
		name = "Template"
	}
	spec.Name = name
	if err := s.Store.UpdateTemplate(r.Context(), u.ID, id, name, spec); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	tpl, err := s.Store.GetTemplate(r.Context(), u.ID, id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"template": tpl})
}

func (s *Server) handleDeleteTemplate(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	id, ok := urlID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := s.Store.DeleteTemplate(r.Context(), u.ID, id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
