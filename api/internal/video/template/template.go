package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
)

const SchemaVersion = 2

// Spec é o template salvo no banco e editado pelo wizard.
type Spec struct {
	SchemaVersion int        `json:"schemaVersion"`
	Name          string     `json:"name"`
	Preset        string     `json:"preset"`
	Canvas        Canvas     `json:"canvas"`
	Profile       Profile    `json:"profile"`
	Text          TextStyle  `json:"text"`
	Logo          Logo       `json:"logo"`
	Zones         []Zone     `json:"zones"`
	Phrases       []string   `json:"phrases"`
	Variations    Variations `json:"variations"`
}

type Canvas struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	FPS        int    `json:"fps"`
	Background string `json:"background"`
}

type Profile struct {
	Name       string `json:"name"`
	Handle     string `json:"handle"`
	Verified   bool   `json:"verified"`
	Avatar     string `json:"avatar"`
	Position   string `json:"position"` // top | bottom | hidden
	Align      string `json:"align"`    // left | center
	Theme      string `json:"theme"`    // light | dark
	AvatarSize int    `json:"avatarSize"`
	NameSize   int    `json:"nameSize"`
	HandleSize int    `json:"handleSize"`
}

type TextStyle struct {
	Font          string `json:"font"`
	Size          int    `json:"size"`
	Color         string `json:"color"`
	Align         string `json:"align"`
	SideSpacing   int    `json:"sideSpacing"`
	Position      string `json:"position"` // top | bottom | hidden
	Band          bool   `json:"band"`
	BandColor     string `json:"bandColor"`
	StrokeEnabled bool   `json:"strokeEnabled"`
	StrokeWidth   int    `json:"strokeWidth"`
	StrokeColor   string `json:"strokeColor"`
	Offset        int    `json:"offset"`
}

// Zone é um bloco empilhado dentro do canvas.
// Type: video | blank | image (image ainda não renderiza, fica reservado).
type Zone struct {
	Type   string  `json:"type"`
	Weight float64 `json:"weight"`
	Radius int     `json:"radius"`
	Fit    string  `json:"fit"` // cover | contain
	// SourceID aponta para outro vídeo da biblioteca. Vazio = vídeo principal.
	SourceID *int64 `json:"sourceId,omitempty"`
	// ImagePath é o nome do arquivo em storage/images (zonas de imagem).
	ImagePath string `json:"imagePath,omitempty"`
}

// Logo é um selo/watermark opcional sobre o canvas.
type Logo struct {
	Enabled  bool    `json:"enabled"`
	Path     string  `json:"path"`     // nome do arquivo em storage/logos
	Position string  `json:"position"` // top-left | top-right | bottom-left | bottom-right
	Size     int     `json:"size"`     // largura em px do canvas
	Opacity  float64 `json:"opacity"`
	Margin   int     `json:"margin"`
}

type Variations struct {
	Count      int       `json:"count"`
	Zoom       []float64 `json:"zoom"`
	Saturation []float64 `json:"saturation"`
	Brightness []float64 `json:"brightness"`
	Speed      []float64 `json:"speed"`
	OffsetX    []float64 `json:"offsetX"`
	OffsetY    []float64 `json:"offsetY"`
	Mirror     bool      `json:"mirror"`
	SeedBase   int64     `json:"seedBase"`
}

func Default() Spec {
	return Spec{
		SchemaVersion: SchemaVersion,
		Name:          "Novo template",
		Preset:        "blank",
		Canvas:        Canvas{Width: 1080, Height: 1920, FPS: 30, Background: "#0E0E10"},
		Profile: Profile{
			Name:       "Seu perfil",
			Handle:     "@seuperfil",
			Verified:   true,
			Position:   "top",
			Align:      "center",
			Theme:      "light",
			AvatarSize: 170,
			NameSize:   50,
			HandleSize: 42,
		},
		Text: TextStyle{
			Font:        "Inter Bold",
			Size:        53,
			Color:       "#FFFFFF",
			Align:       "center",
			SideSpacing: 84,
			Position:    "top",
			Band:        false,
			BandColor:   "#000000",
			StrokeWidth: 4,
			StrokeColor: "#000000",
			Offset:      24,
		},
		Logo: Logo{
			Enabled:  false,
			Position: "top-right",
			Size:     140,
			Opacity:  0.9,
			Margin:   40,
		},
		Zones:   []Zone{{Type: "video", Weight: 1, Radius: 16, Fit: "cover"}},
		Phrases: []string{"Coloque sua headline aqui"},
		Variations: Variations{
			Count:      5,
			Zoom:       []float64{1.0, 1.18},
			Saturation: []float64{0.9, 1.15},
			Brightness: []float64{0.96, 1.06},
			Speed:      []float64{0.96, 1.04},
			OffsetX:    []float64{0.35, 0.65},
			OffsetY:    []float64{0.4, 0.6},
			SeedBase:   42,
		},
	}
}

var hexRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func (s *Spec) Normalize() {
	d := Default()
	s.SchemaVersion = SchemaVersion
	if s.Name == "" {
		s.Name = d.Name
	}
	if s.Preset == "" {
		s.Preset = d.Preset
	}
	if s.Canvas.Width <= 0 || s.Canvas.Height <= 0 {
		s.Canvas.Width, s.Canvas.Height = d.Canvas.Width, d.Canvas.Height
	}
	if s.Canvas.FPS <= 0 {
		s.Canvas.FPS = d.Canvas.FPS
	}
	if !hexRe.MatchString(s.Canvas.Background) {
		s.Canvas.Background = d.Canvas.Background
	}
	if s.Profile.Position == "" {
		s.Profile.Position = "top"
	}
	if s.Profile.Align == "" {
		s.Profile.Align = "center"
	}
	if s.Profile.Theme == "" {
		s.Profile.Theme = "light"
	}
	if s.Profile.AvatarSize <= 0 {
		s.Profile.AvatarSize = d.Profile.AvatarSize
	}
	if s.Profile.NameSize <= 0 {
		s.Profile.NameSize = d.Profile.NameSize
	}
	if s.Profile.HandleSize <= 0 {
		s.Profile.HandleSize = d.Profile.HandleSize
	}
	if s.Text.Size <= 0 {
		s.Text.Size = d.Text.Size
	}
	if !hexRe.MatchString(s.Text.Color) {
		s.Text.Color = d.Text.Color
	}
	if !hexRe.MatchString(s.Text.StrokeColor) {
		s.Text.StrokeColor = d.Text.StrokeColor
	}
	if !hexRe.MatchString(s.Text.BandColor) {
		s.Text.BandColor = d.Text.BandColor
	}
	if s.Text.SideSpacing <= 0 {
		s.Text.SideSpacing = d.Text.SideSpacing
	}
	if s.Text.StrokeWidth <= 0 {
		s.Text.StrokeWidth = d.Text.StrokeWidth
	}
	if s.Text.Align == "" {
		s.Text.Align = "center"
	}
	if s.Text.Position == "" {
		s.Text.Position = "top"
	}
	if s.Text.Offset < 0 {
		s.Text.Offset = d.Text.Offset
	}
	switch s.Logo.Position {
	case "top-left", "top-right", "bottom-left", "bottom-right":
	default:
		s.Logo.Position = "top-right"
	}
	if s.Logo.Size <= 0 {
		s.Logo.Size = d.Logo.Size
	}
	if s.Logo.Size > 600 {
		s.Logo.Size = 600
	}
	if s.Logo.Opacity <= 0 || s.Logo.Opacity > 1 {
		s.Logo.Opacity = d.Logo.Opacity
	}
	if s.Logo.Margin < 0 || s.Logo.Margin > 200 {
		s.Logo.Margin = d.Logo.Margin
	}
	if len(s.Zones) == 0 {
		s.Zones = d.Zones
	}
	total := 0.0
	for i := range s.Zones {
		z := &s.Zones[i]
		switch z.Type {
		case "video", "blank", "image":
		default:
			z.Type = "video"
		}
		if z.Weight <= 0 {
			z.Weight = 1
		}
		if z.Fit != "contain" {
			z.Fit = "cover"
		}
		if z.Radius < 0 {
			z.Radius = 0
		}
		total += z.Weight
	}
	if total == 0 {
		s.Zones[0].Weight = 1
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
	if len(s.Variations.Brightness) == 0 {
		s.Variations.Brightness = d.Variations.Brightness
	}
	if len(s.Variations.Speed) == 0 {
		s.Variations.Speed = d.Variations.Speed
	}
	if len(s.Variations.OffsetX) == 0 {
		s.Variations.OffsetX = d.Variations.OffsetX
	}
	if len(s.Variations.OffsetY) == 0 {
		s.Variations.OffsetY = d.Variations.OffsetY
	}
	hasPhrase := false
	for _, p := range s.Phrases {
		if strings.TrimSpace(p) != "" {
			hasPhrase = true
			break
		}
	}
	if !hasPhrase {
		s.Phrases = d.Phrases
	}
}

func Validate(s Spec) error {
	var errs []string
	if s.Canvas.Width != 1080 || s.Canvas.Height != 1920 {
		errs = append(errs, "canvas deve ser 1080x1920")
	}
	if !hexRe.MatchString(s.Canvas.Background) {
		errs = append(errs, "canvas.background inválido")
	}
	if !hexRe.MatchString(s.Text.Color) {
		errs = append(errs, "text.color inválido")
	}
	if s.Text.Size < 20 || s.Text.Size > 160 {
		errs = append(errs, "text.size deve estar entre 20 e 160")
	}
	if s.Text.SideSpacing < 0 || s.Text.SideSpacing > 400 {
		errs = append(errs, "text.sideSpacing inválido")
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
		if v < 1.0 || v > 2.0 {
			errs = append(errs, "variations.zoom deve estar entre 1.0 e 2.0")
			break
		}
	}
	videos := 0
	images := 0
	for _, z := range s.Zones {
		if z.Weight <= 0 {
			errs = append(errs, "toda zona precisa de peso maior que zero")
			break
		}
		switch z.Type {
		case "video":
			videos++
		case "image":
			images++
		}
	}
	if videos == 0 {
		errs = append(errs, "o template precisa de pelo menos uma zona de vídeo")
	}
	for _, z := range s.Zones {
		if z.Type == "image" && strings.TrimSpace(z.ImagePath) == "" {
			errs = append(errs, "toda zona de imagem precisa de uma imagem")
			break
		}
	}
	if s.Logo.Enabled && strings.TrimSpace(s.Logo.Path) == "" {
		errs = append(errs, "envie a logo ou desative a opção")
	}
	frases := 0
	for _, p := range s.Phrases {
		if strings.TrimSpace(p) != "" {
			frases++
		}
	}
	if frases == 0 && s.Text.Position != "hidden" {
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

// Geometry devolve a área útil de cada zona em pixels do canvas.
type ZoneBox struct {
	Type   string `json:"type"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	W      int    `json:"w"`
	H      int    `json:"h"`
	Radius int    `json:"radius"`
	Fit    string `json:"fit"`
	Index  int    `json:"index"`
}

type Layout struct {
	Canvas     Canvas    `json:"canvas"`
	Zones      []ZoneBox `json:"zones"`
	TextY      int       `json:"textY"`
	TextH      int       `json:"textH"`
	ProfileY   int       `json:"profileY"`
	ProfileH   int       `json:"profileH"`
	ProfileOff bool      `json:"profileOff"`
}

const (
	margin     = 40
	profileGap = 18
	textGap    = 14
	bottomSafe = 60
)

// BuildLayout calcula as posições de todas as zonas.
// Ordem vertical: [perfil topo] [headline topo] [zonas] [headline base] [perfil base]
func BuildLayout(s Spec) Layout {
	s.Normalize()
	W, H := s.Canvas.Width, s.Canvas.Height
	l := Layout{Canvas: s.Canvas}

	profH := 0
	if s.Profile.Position != "hidden" {
		if s.Profile.Align == "left" {
			stack := s.Profile.NameSize + s.Profile.HandleSize + 8
			if stack > s.Profile.AvatarSize {
				profH = stack
			} else {
				profH = s.Profile.AvatarSize
			}
		} else {
			profH = s.Profile.AvatarSize + profileGap + s.Profile.NameSize + s.Profile.HandleSize + 8
		}
		l.ProfileH = profH
	} else {
		l.ProfileOff = true
	}

	textH := 0
	if s.Text.Position != "hidden" {
		lines := MeasureTextLines(s)
		if lines < 1 {
			lines = 1
		}
		textH = int(float64(s.Text.Size)*1.3*float64(lines)) + s.Text.Offset*2
	}
	l.TextH = textH

	top := margin
	bottom := H - bottomSafe

	cursor := top
	if s.Profile.Position == "top" {
		l.ProfileY = cursor
		cursor += profH
	}
	if s.Text.Position == "top" {
		l.TextY = cursor
		cursor += textH
	}
	zonesTop := cursor

	cursor = bottom
	if s.Profile.Position == "bottom" {
		l.ProfileY = cursor - profH
		cursor -= profH
	}
	if s.Text.Position == "bottom" {
		l.TextY = cursor - textH
		cursor -= textH
	}
	zonesBottom := cursor
	if zonesBottom <= zonesTop {
		zonesBottom = zonesTop + 100
	}

	totalWeight := 0.0
	for _, z := range s.Zones {
		totalWeight += z.Weight
	}
	if totalWeight <= 0 {
		totalWeight = 1
	}

	x := margin
	w := W - margin*2
	avail := zonesBottom - zonesTop
	accY := zonesTop
	for i, z := range s.Zones {
		h := int(math.Round(float64(avail) * (z.Weight / totalWeight)))
		if i == len(s.Zones)-1 {
			h = zonesBottom - accY
		}
		if h%2 != 0 {
			h--
		}
		if h < 2 {
			h = 2
		}
		l.Zones = append(l.Zones, ZoneBox{
			Type: z.Type, X: x, Y: accY, W: w, H: h,
			Radius: z.Radius, Fit: z.Fit, Index: i,
		})
		accY += h
	}
	return l
}

// SortedRanges devolve min e max de uma faixa.
func SortedRanges(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 1, 1
	}
	cp := append([]float64(nil), values...)
	sort.Float64s(cp)
	return cp[0], cp[len(cp)-1]
}

// glyphRatio é uma estimativa conservadora da largura média de um caractere
// em relação ao tamanho da fonte (Helvetica/Inter).
const glyphRatio = 0.56

// WrapText quebra o texto em linhas que cabem em maxWidth.
func WrapText(text string, fontSize, maxWidth, maxLines int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if fontSize <= 0 {
		fontSize = 40
	}
	if maxWidth <= 0 {
		maxWidth = 1000
	}
	if maxLines <= 0 {
		maxLines = 3
	}
	limit := int(float64(maxWidth) / (float64(fontSize) * glyphRatio))
	if limit < 8 {
		limit = 8
	}
	words := strings.Fields(text)
	lines := []string{}
	current := ""
	for _, w := range words {
		if current == "" {
			current = w
			continue
		}
		if len([]rune(current))+1+len([]rune(w)) <= limit {
			current += " " + w
			continue
		}
		lines = append(lines, current)
		current = w
		if len(lines) == maxLines {
			break
		}
	}
	if current != "" && len(lines) < maxLines {
		lines = append(lines, current)
	}
	if len(lines) == maxLines {
		// se sobrou texto, indica reticências na última linha
		remaining := false
		for _, w := range words {
			if !strings.Contains(strings.Join(lines, " "), w) {
				remaining = true
				break
			}
		}
		if remaining {
			last := []rune(lines[len(lines)-1])
			if len(last) > 1 {
				lines[len(lines)-1] = string(last[:len(last)-1]) + "…"
			}
		}
	}
	return lines
}

// MeasureTextLines devolve quantas linhas o maior texto do template ocupa.
func MeasureTextLines(s Spec) int {
	if s.Text.Position == "hidden" || len(s.Phrases) == 0 {
		return 0
	}
	maxWidth := s.Canvas.Width - s.Text.SideSpacing*2
	best := 1
	for _, p := range s.Phrases {
		if n := len(WrapText(p, s.Text.Size, maxWidth, 3)); n > best {
			best = n
		}
	}
	return best
}
