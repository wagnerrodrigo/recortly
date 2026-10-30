package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const SchemaVersion = 1

// Spec é o template salvo no banco e editado pelo wizard.
type Spec struct {
	SchemaVersion int       `json:"schemaVersion"`
	Name          string    `json:"name"`
	Canvas        Canvas    `json:"canvas"`
	Profile       Profile   `json:"profile"`
	Text          TextStyle `json:"text"`
	Phrases       []string  `json:"phrases"`
	Variations    Variations `json:"variations"`
}

type Canvas struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	FPS        int    `json:"fps"`
	Background string `json:"background"`
}

type Profile struct {
	Name     string `json:"name"`
	Handle   string `json:"handle"`
	Verified bool   `json:"verified"`
	Avatar   string `json:"avatar"`
}

type TextStyle struct {
	Font            string `json:"font"`
	Size            int    `json:"size"`
	Color           string `json:"color"`
	Align           string `json:"align"`
	SideSpacing     int    `json:"sideSpacing"`
	OffsetFromProfile int  `json:"offsetFromProfile"`
	StrokeEnabled   bool   `json:"strokeEnabled"`
	StrokeWidth     int    `json:"strokeWidth"`
	StrokeColor     string `json:"strokeColor"`
	ShowHandle      bool   `json:"showHandle"`
}

type Variations struct {
	Count      int     `json:"count"`
	Zoom       []float64 `json:"zoom"`
	Saturation []float64 `json:"saturation"`
	Speed      []float64 `json:"speed"`
	Mirror     bool    `json:"mirror"`
	SeedBase   int64   `json:"seedBase"`
}

func Default() Spec {
	return Spec{
		SchemaVersion: SchemaVersion,
		Name:          "Novo template",
		Canvas:        Canvas{Width: 1080, Height: 1920, FPS: 30, Background: "#0E0E10"},
		Profile:       Profile{Name: "Seu perfil", Handle: "@seuperfil", Verified: true},
		Text: TextStyle{
			Font:            "Inter Bold",
			Size:            53,
			Color:           "#FFFFFF",
			Align:           "center",
			SideSpacing:     84,
			OffsetFromProfile: 156,
			StrokeWidth:     4,
			StrokeColor:     "#000000",
			ShowHandle:      true,
		},
		Phrases:    []string{"Coloque sua headline aqui"},
		Variations: Variations{Count: 1, Zoom: []float64{1.0, 1.06}, Saturation: []float64{1.0, 1.08}, Speed: []float64{1.0, 1.0}, Mirror: false, SeedBase: 42},
	}
}

var hexRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func (s *Spec) Normalize() {
	d := Default()
	if s.SchemaVersion == 0 {
		s.SchemaVersion = SchemaVersion
	}
	if s.Canvas.Width <= 0 || s.Canvas.Height <= 0 {
		s.Canvas.Width, s.Canvas.Height = d.Canvas.Width, d.Canvas.Height
	}
	if s.Canvas.FPS <= 0 {
		s.Canvas.FPS = d.Canvas.FPS
	}
	if s.Canvas.Background == "" {
		s.Canvas.Background = d.Canvas.Background
	}
	if s.Text.Size <= 0 {
		s.Text.Size = d.Text.Size
	}
	if s.Text.Color == "" {
		s.Text.Color = d.Text.Color
	}
	if s.Text.StrokeColor == "" {
		s.Text.StrokeColor = d.Text.StrokeColor
	}
	if s.Text.SideSpacing <= 0 {
		s.Text.SideSpacing = d.Text.SideSpacing
	}
	if s.Text.OffsetFromProfile <= 0 {
		s.Text.OffsetFromProfile = d.Text.OffsetFromProfile
	}
	if s.Text.StrokeWidth <= 0 {
		s.Text.StrokeWidth = d.Text.StrokeWidth
	}
	if s.Text.Align == "" {
		s.Text.Align = "center"
	}
	if s.Variations.Count <= 0 {
		s.Variations.Count = 1
	}
	if len(s.Variations.Zoom) == 0 {
		s.Variations.Zoom = d.Variations.Zoom
	}
	if len(s.Variations.Saturation) == 0 {
		s.Variations.Saturation = d.Variations.Saturation
	}
	if len(s.Variations.Speed) == 0 {
		s.Variations.Speed = d.Variations.Speed
	}
	if len(s.Phrases) == 0 {
		s.Phrases = d.Phrases
	}
}

func Validate(s Spec) error {
	var errs []string
	if s.SchemaVersion != SchemaVersion {
		errs = append(errs, fmt.Sprintf("schemaVersion deve ser %d", SchemaVersion))
	}
	if s.Canvas.Width != 1080 || s.Canvas.Height != 1920 {
		errs = append(errs, "canvas deve ser 1080x1920")
	}
	if !hexRe.MatchString(s.Canvas.Background) {
		errs = append(errs, "canvas.background inválido")
	}
	if !hexRe.MatchString(s.Text.Color) {
		errs = append(errs, "text.color inválido")
	}
	if s.Text.Size < 20 || s.Text.Size > 140 {
		errs = append(errs, "text.size deve estar entre 20 e 140")
	}
	if s.Text.SideSpacing < 0 || s.Text.SideSpacing > 400 {
		errs = append(errs, "text.sideSpacing inválido")
	}
	if s.Text.OffsetFromProfile < 0 || s.Text.OffsetFromProfile > 900 {
		errs = append(errs, "text.offsetFromProfile inválido")
	}
	if s.Variations.Count < 1 || s.Variations.Count > 30 {
		errs = append(errs, "variations.count deve estar entre 1 e 30")
	}
	for _, v := range s.Variations.Speed {
		if v < 0.5 || v > 2.0 {
			errs = append(errs, "variations.speed deve estar entre 0.5 e 2.0")
			break
		}
	}
	for _, v := range s.Variations.Zoom {
		if v < 1.0 || v > 1.5 {
			errs = append(errs, "variations.zoom deve estar entre 1.0 e 1.5")
			break
		}
	}
	frases := 0
	for _, p := range s.Phrases {
		if strings.TrimSpace(p) != "" {
			frases++
		}
	}
	if frases == 0 {
		errs = append(errs, "informe pelo menos uma frase")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func Parse(raw []byte) (Spec, error) {
	var s Spec
	if err := json.Unmarshal(raw, &s); err != nil {
		return s, fmt.Errorf("template inválido: %w", err)
	}
	s.Normalize()
	return s, nil
}
