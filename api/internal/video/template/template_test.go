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
		{"schema errado", func(s *Spec) { s.SchemaVersion = 99 }, "schemaVersion"},
		{"canvas errado", func(s *Spec) { s.Canvas.Width = 720 }, "canvas"},
		{"cor inválida", func(s *Spec) { s.Text.Color = "branco" }, "text.color"},
		{"fonte pequena", func(s *Spec) { s.Text.Size = 5 }, "text.size"},
		{"sem frases", func(s *Spec) { s.Phrases = []string{"  "} }, "frase"},
		{"variações demais", func(s *Spec) { s.Variations.Count = 99 }, "variations.count"},
		{"velocidade fora", func(s *Spec) { s.Variations.Speed = []float64{3} }, "variations.speed"},
		{"zoom fora", func(s *Spec) { s.Variations.Zoom = []float64{0.5} }, "variations.zoom"},
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
}

func TestParseInvalido(t *testing.T) {
	if _, err := Parse([]byte("{não é json")); err == nil {
		t.Fatal("esperava erro de parse")
	}
}
