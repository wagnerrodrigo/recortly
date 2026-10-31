package video

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/smashcut/smashcut/api/internal/model"
	"github.com/smashcut/smashcut/api/internal/video/template"
)

func TestPlanDeterministico(t *testing.T) {
	spec := template.Default()
	spec.Variations.Count = 5
	spec.Phrases = []string{"a", "b"}

	first := Plan(spec, 100)
	second := Plan(spec, 100)

	if len(first) != 5 {
		t.Fatalf("esperava 5 planos, veio %d", len(first))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("plano %d não é determinístico: %+v != %+v", i, first[i], second[i])
		}
	}
	if first[0].Headline != "a" || first[1].Headline != "b" || first[2].Headline != "a" {
		t.Fatalf("frases não ciclaram: %+v", first)
	}
}

func TestPlanSeedDiferenteGeraVariacoesDiferentes(t *testing.T) {
	spec := template.Default()
	spec.Variations.Count = 4

	a := Plan(spec, 1)
	b := Plan(spec, 2)

	iguais := 0
	for i := range a {
		if a[i].Zoom == b[i].Zoom && a[i].OffsetX == b[i].OffsetX && a[i].OffsetY == b[i].OffsetY {
			iguais++
		}
	}
	if iguais == len(a) {
		t.Fatal("seeds diferentes deveriam produzir variações diferentes")
	}
}

func TestPlanVariacoesDistintas(t *testing.T) {
	spec := template.Default()
	spec.Variations.Count = 5
	plans := Plan(spec, 42)

	vistos := map[float64]bool{}
	for _, p := range plans {
		chave := float64(int(p.OffsetX*1000))*1000 + float64(int(p.Zoom*1000))
		if vistos[chave] {
			t.Fatalf("variação repetida: %+v", p)
		}
		vistos[chave] = true
	}
}

func TestPlanMirror(t *testing.T) {
	spec := template.Default()
	spec.Variations.Mirror = true
	spec.Variations.Count = 4
	plans := Plan(spec, 7)
	for i, p := range plans {
		if p.Mirror != (i%2 == 1) {
			t.Fatalf("mirror errado no índice %d: %v", i, p.Mirror)
		}
	}
}

func TestPlanSemFrases(t *testing.T) {
	spec := template.Default()
	spec.Phrases = []string{" "}
	plans := Plan(spec, 1)
	if len(plans) == 0 || plans[0].Headline == "" {
		t.Fatalf("deveria gerar plano com headline não vazia: %+v", plans)
	}
}

func TestBuildArgsZonas(t *testing.T) {
	dir := t.TempDir()
	spec := template.Default()
	spec.Zones = []template.Zone{
		{Type: "video", Weight: 0.6, Radius: 16, Fit: "cover"},
		{Type: "blank", Weight: 0.4},
	}
	in := RenderInput{
		SourcePath: "/tmp/in.mp4",
		OutPath:    filepath.Join(dir, "variant-01.mp4"),
		Clip:       model.Clip{StartMS: 1000, EndMS: 5000, Layout: "blur", Zoom: 1},
		Spec:       spec,
		Plan:       VariationPlan{Headline: "Headline de teste", Zoom: 1.03, Saturation: 1.05, Brightness: 1, Speed: 1},
		HasAudio:   true,
		FontPath:   "/System/Library/Fonts/Helvetica.ttc",
	}
	args, err := BuildArgs(in)
	if err != nil {
		t.Fatalf("BuildArgs: %v", err)
	}
	joined := strings.Join(args, " ")
	for _, want := range []string{"-filter_complex", "overlay", "drawtext", "[vout]", "libx264", "-t 4.000"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("args sem %q: %s", want, joined)
		}
	}
	if !strings.Contains(joined, "-map 0:a:0") {
		t.Fatalf("áudio deveria ser mapeado: %s", joined)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) == 0 {
		t.Fatal("arquivos de drawtext não foram criados")
	}
}

func TestBuildArgsSemAudioEExport(t *testing.T) {
	dir := t.TempDir()
	in := RenderInput{
		SourcePath: "/tmp/in.mp4",
		OutPath:    filepath.Join(dir, "out.mp4"),
		Clip:       model.Clip{StartMS: 0, EndMS: 2000, Layout: "fill", Zoom: 1},
		Spec:       template.Default(),
		Plan:       VariationPlan{Headline: "x", Zoom: 1, Saturation: 1, Brightness: 1, Speed: 1},
		HasAudio:   false,
		FontPath:   "/System/Library/Fonts/Helvetica.ttc",
	}
	args, err := BuildArgs(in)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-an") {
		t.Fatalf("esperava -an: %s", joined)
	}
}

func TestBuildArgsClipInvalido(t *testing.T) {
	in := RenderInput{
		Clip: model.Clip{StartMS: 5000, EndMS: 1000},
		Spec: template.Default(),
	}
	if _, err := BuildArgs(in); err == nil {
		t.Fatal("esperava erro para clip inválido")
	}
}

func TestBuildArgsSemZonaDeVideo(t *testing.T) {
	spec := template.Default()
	spec.Zones = []template.Zone{{Type: "blank", Weight: 1}}
	in := RenderInput{
		Clip: model.Clip{StartMS: 0, EndMS: 2000},
		Spec: spec,
		Plan: VariationPlan{Headline: "x"},
	}
	if _, err := BuildArgs(in); err == nil {
		t.Fatal("esperava erro para template sem vídeo")
	}
}

func TestHexToFFColor(t *testing.T) {
	tests := map[string]string{
		"#ffffff": "0xFFFFFF",
		"000000":  "0x000000",
		"xyz":     "0x0E0E10",
		"":        "0x0E0E10",
	}
	for in, want := range tests {
		if got := hexToFFColor(in); got != want {
			t.Fatalf("hexToFFColor(%q) = %q, esperava %q", in, got, want)
		}
	}
}
