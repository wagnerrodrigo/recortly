package video

import (
	"math"
	"strings"

	"github.com/smashcut/smashcut/api/internal/video/template"
)

type VariationPlan struct {
	Index      int
	Headline   string
	Zoom       float64
	Saturation float64
	Brightness float64
	Speed      float64
	OffsetX    float64
	OffsetY    float64
	Mirror     bool
}

func mix(seed int64, idx int, salt int) float64 {
	x := math.Sin(float64(seed*7919)+float64(idx*104729)+float64(salt*1299709)) * 43758.5453
	return x - math.Floor(x)
}

func pickRange(values []float64, seed int64, idx, salt int) float64 {
	if len(values) == 0 {
		return 1
	}
	if len(values) == 1 {
		return values[0]
	}
	min, max := template.SortedRanges(values)
	if max <= min {
		return min
	}
	return min + (max-min)*mix(seed, idx, salt)
}

// spread força valores diferentes entre variações mesmo com faixas estreitas.
func spread(seed int64, idx, count int, salt int) float64 {
	if count <= 1 {
		return 0.5
	}
	base := float64(idx) / float64(count-1)
	jitter := (mix(seed, idx, salt) - 0.5) * (0.6 / float64(count))
	return math.Min(1, math.Max(0, base+jitter))
}

// Plan devolve variações determinísticas por render (seed do render + índice).
func Plan(spec template.Spec, renderSeed int64) []VariationPlan {
	spec.Normalize()
	seed := spec.Variations.SeedBase + renderSeed

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

	plan := make([]VariationPlan, 0, count)
	for i := 0; i < count; i++ {
		v := spec.Variations
		zoomMin, zoomMax := template.SortedRanges(v.Zoom)
		satMin, satMax := template.SortedRanges(v.Saturation)
		briMin, briMax := template.SortedRanges(v.Brightness)
		spdMin, spdMax := template.SortedRanges(v.Speed)
		oxMin, oxMax := template.SortedRanges(v.OffsetX)
		oyMin, oyMax := template.SortedRanges(v.OffsetY)

		p := VariationPlan{
			Index:    i,
			Headline: phrases[i%len(phrases)],
		}
		if zoomMax > zoomMin {
			p.Zoom = zoomMin + (zoomMax-zoomMin)*spread(seed, i, count, 1)
		} else {
			p.Zoom = zoomMin
		}
		if satMax > satMin {
			p.Saturation = satMin + (satMax-satMin)*spread(seed, i, count, 2)
		} else {
			p.Saturation = satMin
		}
		if briMax > briMin {
			p.Brightness = briMin + (briMax-briMin)*spread(seed, i, count, 3)
		} else {
			p.Brightness = briMin
		}
		if spdMax > spdMin {
			p.Speed = spdMin + (spdMax-spdMin)*mix(seed, i, 4)
		} else {
			p.Speed = spdMin
		}
		if oxMax > oxMin {
			p.OffsetX = oxMin + (oxMax-oxMin)*spread(seed, i, count, 5)
		} else {
			p.OffsetX = 0.5
		}
		if oyMax > oyMin {
			p.OffsetY = oyMin + (oyMax-oyMin)*spread(seed, i, count, 6)
		} else {
			p.OffsetY = 0.5
		}
		if p.Speed <= 0 {
			p.Speed = 1
		}
		if p.Zoom < 1 {
			p.Zoom = 1
		}
		if p.Saturation <= 0 {
			p.Saturation = 1
		}
		if p.Brightness <= 0 {
			p.Brightness = 1
		}
		if v.Mirror && i%2 == 1 {
			p.Mirror = true
		}
		plan = append(plan, p)
	}
	return plan
}
