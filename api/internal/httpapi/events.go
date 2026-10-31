package httpapi

import (
	"net/http"
	"strconv"
)

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	level := r.URL.Query().Get("level")
	source := r.URL.Query().Get("source")
	list, err := s.Store.ListEvents(r.Context(), level, source, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": list})
}

func (s *Server) handleClearEvents(w http.ResponseWriter, r *http.Request) {
	var in struct {
		OlderThanDays int `json:"olderThanDays"`
	}
	if r.ContentLength > 0 {
		if err := decodeJSON(w, r, &in); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	n, err := s.Store.ClearEvents(r.Context(), in.OlderThanDays)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": n})
}
