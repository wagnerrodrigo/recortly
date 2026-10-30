package video

import (
	"math"
	"strings"

	"github.com/smashcut/smashcut/api/internal/video/template"
)

// VariationPlan descreve uma variação determinística de um clip com template.
type VariationPlan struct {
	Index      int
	Headline   string
	Zoom       float64
	Saturation float64
	Speed      float64
	Mirror     bool
}

// mix imita um hash determinístico simples a partir da seed e do índice.
func mix(seed int64, idx int, salt int) float64 {
	x := math.Sin(float64(seed*7919+int64(idx*104729)+int64(salt*1299709))) * 43758.5453
	return x - math.Floor(x)
}

func pickRange(values []float64, seed int64, idx, salt int) float64 {
	if len(values) == 0 {
		return 1
	}
	if len(values) == 1 {
		return values[0]
	}
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min + (max-min)*mix(seed, idx, salt)
}

// Plan devolve uma variação por render, determinística pela seed.
func Plan(spec template.Spec) []VariationPlan {
	spec.Normalize()
	phrases := make([]string, 0, len(spec.Phrases))
	for _, p := range spec.Phrases {
		if strings.TrimSpace(p) != "" {
			phrases = append(phrases, strings.TrimSpace(p))
		}
	}
	if len(phrases) == 0 {
		phrases = []string{" "}
	}

	count := spec.Variations.Count
	if count < 1 {
		count = 1
	}
	plans := make([]VariationPlan, 0, count)
	for i := 0; i < count; i++ {
		p := VariationPlan{
			Index:      i,
			Headline:   phrases[i%len(phrases)],
			Zoom:       pickRange(spec.Variations.Zoom, spec.Variations.SeedBase, i, 1),
			Saturation: pickRange(spec.Variations.Saturation, spec.Variations.SeedBase, i, 2),
			Speed:      pickRange(spec.Variations.Speed, spec.Variations.SeedBase, i, 3),
		}
		if spec.Variations.Mirror && i%2 == 1 {
			p.Mirror = true
		}
		if p.Speed <= 0 {
			p.Speed = 1
		}
		if p.Zoom <= 0 {
			p.Zoom = 1
		}
		if p.Saturation <= 0 {
			p.Saturation = 1
		}
		plans = append(plans, p)
	}
	return plans
}
