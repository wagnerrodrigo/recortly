package httpapi

import (
	"net/http"
	"os"
)

func (s *Server) handleCreateRender(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	var in struct {
		ClipID      int64  `json:"clipId"`
		TemplateID  *int64 `json:"templateId"`
		VariantCount int   `json:"variantCount"`
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
