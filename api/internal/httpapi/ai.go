package httpapi

import (
	"net/http"

	"github.com/recortly/recortly/api/internal/ai"
	"github.com/recortly/recortly/api/internal/video"
	"github.com/recortly/recortly/api/internal/video/template"
)

func (s *Server) handlePresets(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"presets": template.Presets()})
}

// handleLayout devolve a geometria calculada para o preview do wizard.
func (s *Server) handleLayout(w http.ResponseWriter, r *http.Request) {
	var in struct {
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
	writeJSON(w, http.StatusOK, map[string]any{"layout": template.BuildLayout(spec)})
}

func (s *Server) handleAIPhrases(w http.ResponseWriter, r *http.Request) {
	var in ai.PhrasesRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	client := ai.New(s.Cfg.AIBaseURL, s.Cfg.AIModel)
	client.MaxTokens = s.Cfg.AIMaxTokens
	client.Reasoning = s.Cfg.AIReasoning
	phrases, err := client.Phrases(r.Context(), in)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"phrases": phrases})
}

func (s *Server) handleAIStatus(w http.ResponseWriter, r *http.Request) {
	client := ai.New(s.Cfg.AIBaseURL, s.Cfg.AIModel)
	writeJSON(w, http.StatusOK, map[string]any{
		"configured": client.Configured(),
		"baseUrl":    s.Cfg.AIBaseURL,
		"model":      s.Cfg.AIModel,
		"maxTokens":  s.Cfg.AIMaxTokens,
		"reasoning":  s.Cfg.AIReasoning,
	})
}
