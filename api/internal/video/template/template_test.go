package template

import (
	"strings"
	"testing"
)

func TestDefaultValidates(t *testing.T) {
	spec := Default()
	spec.Normalize()
	if err := Validate(spec); err != nil {
		t.Fatalf("default deveria ser válido: %v", err)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Spec)
		want   string
	}{
		{"canvas errado", func(s *Spec) { s.Canvas.Width = 720 }, "canvas"},
		{"cor inválida", func(s *Spec) { s.Text.Color = "branco" }, "text.color"},
		{"fonte pequena", func(s *Spec) { s.Text.Size = 5 }, "text.size"},
		{"sem frases", func(s *Spec) { s.Phrases = []string{"  "} }, "frase"},
		{"variações demais", func(s *Spec) { s.Variations.Count = 99 }, "variations.count"},
		{"velocidade fora", func(s *Spec) { s.Variations.Speed = []float64{3} }, "variations.speed"},
		{"zoom fora", func(s *Spec) { s.Variations.Zoom = []float64{0.5} }, "variations.zoom"},
		{"sem zona de vídeo", func(s *Spec) { s.Zones = []Zone{{Type: "blank", Weight: 1}} }, "zona de vídeo"},
		{"zona sem peso", func(s *Spec) { s.Zones = []Zone{{Type: "video", Weight: 0}} }, "peso"},
		{"zona de imagem", func(s *Spec) {
			s.Zones = []Zone{{Type: "video", Weight: 1}, {Type: "image", Weight: 1}}
		}, "imagem"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := Default()
			tt.mutate(&spec)
			err := Validate(spec)
			if err == nil {
				t.Fatalf("esperava erro contendo %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("erro %q não contém %q", err.Error(), tt.want)
			}
		})
	}
}

func TestNormalizePreencheCampos(t *testing.T) {
	var s Spec
	s.Normalize()
	if s.Canvas.Width != 1080 || s.Canvas.Height != 1920 {
		t.Fatalf("canvas não normalizado: %+v", s.Canvas)
	}
	if s.Variations.Count != 1 {
		t.Fatalf("count esperado 1, veio %d", s.Variations.Count)
	}
	if len(s.Phrases) == 0 {
		t.Fatal("frases deveriam ser preenchidas")
	}
	if len(s.Zones) == 0 || s.Zones[0].Weight <= 0 {
		t.Fatalf("zonas inválidas: %+v", s.Zones)
	}
}

func TestParseInvalido(t *testing.T) {
	if _, err := Parse([]byte("{não é json")); err == nil {
		t.Fatal("esperava erro de parse")
	}
}

func TestPresetsValidos(t *testing.T) {
	for _, p := range Presets() {
		spec := p.Spec
		spec.Normalize()
		if err := Validate(spec); err != nil {
			t.Fatalf("preset %s inválido: %v", p.ID, err)
		}
		if preset, ok := PresetByID(p.ID); !ok || preset.Name == "" {
			t.Fatalf("preset %s não encontrado por id", p.ID)
		}
	}
}

func TestBuildLayoutOrdem(t *testing.T) {
	spec := Default()
	spec.Profile.Position = "top"
	spec.Text.Position = "bottom"
	spec.Zones = []Zone{
		{Type: "video", Weight: 0.6},
		{Type: "blank", Weight: 0.4},
	}
	l := BuildLayout(spec)
	if len(l.Zones) != 2 {
		t.Fatalf("esperava 2 zonas, veio %d", len(l.Zones))
	}
	video, blank := l.Zones[0], l.Zones[1]
	if video.Y <= l.ProfileY {
		t.Fatalf("vídeo deveria começar abaixo do perfil: perfil=%d vídeo=%d", l.ProfileY, video.Y)
	}
	if blank.Y <= video.Y {
		t.Fatalf("zonas fora de ordem: %d >= %d", video.Y, blank.Y)
	}
	if blank.Y+blank.H > l.TextY {
		t.Fatalf("zonas não devem invadir a headline: fim=%d textY=%d", blank.Y+blank.H, l.TextY)
	}
	if l.TextY+l.TextH > 1920 {
		t.Fatalf("headline fora do canvas: %d", l.TextY+l.TextH)
	}
	if video.X != 40 || video.W != 1000 {
		t.Fatalf("margem inesperada: x=%d w=%d", video.X, video.W)
	}
}

func TestBuildLayoutEscondido(t *testing.T) {
	spec := Default()
	spec.Profile.Position = "hidden"
	spec.Text.Position = "hidden"
	l := BuildLayout(spec)
	if !l.ProfileOff {
		t.Fatal("perfil deveria estar oculto")
	}
	if l.TextH != 0 {
		t.Fatalf("headline oculta não deveria reservar altura: %d", l.TextH)
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		size    int
		width   int
		max     int
		wantMin int
	}{
		{"curto", "Oi", 50, 900, 3, 1},
		{"longo", strings.Repeat("palavra ", 20), 54, 900, 3, 2},
		{"vazio", "   ", 50, 900, 3, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := WrapText(tt.text, tt.size, tt.width, tt.max)
			if tt.wantMin == 0 {
				if len(lines) != 0 {
					t.Fatalf("esperava 0 linhas, veio %v", lines)
				}
				return
			}
			if len(lines) < tt.wantMin {
				t.Fatalf("esperava pelo menos %d linhas, veio %v", tt.wantMin, lines)
			}
			if len(lines) > tt.max {
				t.Fatalf("mais linhas que o máximo: %v", lines)
			}
		})
	}
}

func TestMeasureTextLines(t *testing.T) {
	spec := Default()
	spec.Phrases = []string{"curta"}
	if n := MeasureTextLines(spec); n != 1 {
		t.Fatalf("esperava 1 linha, veio %d", n)
	}
	spec.Phrases = []string{strings.Repeat("frase comprida ", 12)}
	if n := MeasureTextLines(spec); n < 2 {
		t.Fatalf("esperava várias linhas, veio %d", n)
	}
}
