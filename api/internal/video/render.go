package video

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/smashcut/smashcut/api/internal/model"
	"github.com/smashcut/smashcut/api/internal/video/template"
)

type ProbeInfo struct {
	DurationMS int64
	Width      int
	Height     int
	HasAudio   bool
	FPS        float64
	Codec      string
}

func Probe(ctx context.Context, path string) (ProbeInfo, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error", "-print_format", "json", "-show_format", "-show_streams", path)
	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = string(ee.Stderr)
		}
		return ProbeInfo{}, fmt.Errorf("ffprobe falhou: %s", strings.TrimSpace(stderr))
	}
	var parsed struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
		Streams []struct {
			CodecType  string `json:"codec_type"`
			CodecName  string `json:"codec_name"`
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			RFrameRate string `json:"r_frame_rate"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		return ProbeInfo{}, err
	}
	info := ProbeInfo{}
	if d, err := strconv.ParseFloat(parsed.Format.Duration, 64); err == nil {
		info.DurationMS = int64(d * 1000)
	}
	for _, s := range parsed.Streams {
		switch s.CodecType {
		case "video":
			if info.Width == 0 {
				info.Width, info.Height = s.Width, s.Height
				info.Codec = s.CodecName
				if parts := strings.Split(s.RFrameRate, "/"); len(parts) == 2 {
					n, _ := strconv.ParseFloat(parts[0], 64)
					d, _ := strconv.ParseFloat(parts[1], 64)
					if d > 0 {
						info.FPS = n / d
					}
				}
			}
		case "audio":
			info.HasAudio = true
		}
	}
	if info.Width == 0 {
		return info, errors.New("arquivo não contém trilha de vídeo")
	}
	if info.DurationMS == 0 {
		return info, errors.New("não foi possível ler a duração do vídeo")
	}
	return info, nil
}

type RenderInput struct {
	SourcePath string
	OutPath    string
	Clip       model.Clip
	Spec       TemplateSpec
	Plan       VariationPlan
	HasAudio   bool
	FontPath   string
	LogoPath   string
}

// BuildArgs monta o comando de FFmpeg para uma variação.
func BuildArgs(in RenderInput) ([]string, error) {
	spec := in.Spec
	spec.Normalize()

	dur := float64(in.Clip.EndMS-in.Clip.StartMS) / 1000
	if dur <= 0 {
		return nil, errors.New("clip com duração inválida")
	}
	layout := template.BuildLayout(spec)

	fps := spec.Canvas.FPS
	if fps <= 0 {
		fps = 30
	}
	W, H := spec.Canvas.Width, spec.Canvas.Height

	speed := orOne(in.Plan.Speed)
	zoom := orOne(in.Plan.Zoom)
	sat := orOne(in.Plan.Saturation)
	bri := orOne(in.Plan.Brightness)
	offX := clamp01(in.Plan.OffsetX)
	offY := clamp01(in.Plan.OffsetY)

	srcVideoCount := 0
	for _, z := range spec.Zones {
		if z.Type == "video" {
			srcVideoCount++
		}
	}
	if srcVideoCount == 0 {
		return nil, errors.New("template sem zona de vídeo")
	}

	chain := []string{}
	labels := make([]string, len(layout.Zones))

	// Cada zona de vídeo recebe uma cópia do stream original.
	videoSource := "0:v"
	if srcVideoCount > 1 {
		splits := make([]string, 0, srcVideoCount)
		for i := 0; i < srcVideoCount; i++ {
			splits = append(splits, fmt.Sprintf("vs%d", i))
		}
		chain = append(chain, fmt.Sprintf("[0:v]split=%d[%s]", srcVideoCount, strings.Join(splits, "][")))
	}

	videoN := 0
	for i, z := range layout.Zones {
		if z.Type != "video" {
			label := fmt.Sprintf("blank%d", i)
			chain = append(chain, fmt.Sprintf("color=c=0x1A1A1E:s=%dx%d:d=%.3f:r=%d[%s]",
				z.W, z.H, dur, fps, label))
			labels[i] = label
			continue
		}
		label := fmt.Sprintf("z%d", i)
		// Folga para o zoom, garantindo sempre uma área >= à zona (evita crop inválido).
		factor := 1.12 / zoom
		if factor < 1.02 {
			factor = 1.02
		}
		outW := int(float64(z.W) * factor)
		outH := int(float64(z.H) * factor)
		if outW%2 != 0 {
			outW++
		}
		if outH%2 != 0 {
			outH++
		}
		filter := fmt.Sprintf(
			"fps=%d,scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d:(iw-%d)*%.4f:(ih-%d)*%.4f",
			fps, outW, outH, z.W, z.H, z.W, offX, z.H, offY)
		if in.Plan.Mirror && videoN%2 == 1 {
			filter += ",hflip"
		}
		filter += fmt.Sprintf(",eq=saturation=%.3f:brightness=%.3f", sat, bri-1)
		input := videoSource
		if srcVideoCount > 1 {
			input = fmt.Sprintf("vs%d", videoN)
		}
		chain = append(chain, fmt.Sprintf("[%s]%s[%s]", input, filter, label))
		labels[i] = label
		videoN++
	}

	prev := "base"
	chain = append(chain,
		fmt.Sprintf("color=c=%s:s=%dx%d:d=%.3f:r=%d[%s]",
			hexToFFColor(spec.Canvas.Background), W, H, dur, fps, prev))

	for i, z := range layout.Zones {
		next := fmt.Sprintf("o%d", i)
		chain = append(chain, fmt.Sprintf("[%s][%s]overlay=%d:%d:shortest=0[%s]", prev, labels[i], z.X, z.Y, next))
		prev = next
	}

	useLogo := spec.Logo.Enabled && in.LogoPath != ""
	if useLogo {
		if _, err := os.Stat(in.LogoPath); err != nil {
			useLogo = false
		}
	}
	if useLogo {
		opacity := spec.Logo.Opacity
		if opacity <= 0 || opacity > 1 {
			opacity = 1
		}
		pos := spec.Logo.Position
		margin := spec.Logo.Margin
		x, y := logoPosition(pos, margin)
		chain = append(chain, fmt.Sprintf(
			"[1:v]scale=%d:-2,format=rgba,colorchannelmixer=aa=%.2f[logo]",
			spec.Logo.Size, opacity))
		chain = append(chain, fmt.Sprintf("[%s][logo]overlay=%s:%s[logoo]", prev, x, y))
		prev = "logoo"
	}

	// Faixa de fundo da headline: retângulo de largura total.
	if spec.Text.Band && spec.Text.Position != "hidden" && layout.TextH > 0 {
		bandH := layout.TextH
		if bandH > H {
			bandH = H
		}
		bandY := layout.TextY
		chain = append(chain, fmt.Sprintf(
			"color=c=%s:s=%dx%d:d=%.3f:r=%d,format=rgba,colorchannelmixer=aa=0.9[band]",
			hexToFFColor(spec.Text.BandColor), W, bandH, dur, fps))
		chain = append(chain, fmt.Sprintf("[%s][band]overlay=0:%d[bandout]", prev, bandY))
		prev = "bandout"
	}

	stem := strings.TrimSuffix(in.OutPath, filepath.Ext(in.OutPath))
	fontPath := in.FontPath
	if fontPath == "" {
		fontPath = "/System/Library/Fonts/Helvetica.ttc"
	}
	textFilters := []string{}

	if spec.Profile.Position != "hidden" {
		name := strings.TrimSpace(spec.Profile.Name)
		if name == "" {
			name = "perfil"
		}
		nameFile := stem + ".name.txt"
		if err := os.WriteFile(nameFile, []byte(name), 0o644); err != nil {
			return nil, err
		}
		av := spec.Profile.AvatarSize
		align := spec.Profile.Align
		nameX := W / 2
		avatarX := (W - av) / 2
		nameY := layout.ProfileY + av + 12
		handleY := nameY + spec.Profile.NameSize + 6
		if align == "left" {
			avatarX = 40
			nameX = 40 + av + 24
			nameY = layout.ProfileY + 4
			handleY = nameY + spec.Profile.NameSize + 2
		}

		// Avatar circular com a inicial do perfil.
		if av > 0 {
			initial := strings.ToUpper(string([]rune(name)[0]))
			initFile := stem + ".initial.txt"
			if err := os.WriteFile(initFile, []byte(initial), 0o644); err != nil {
				return nil, err
			}
			chain = append(chain, fmt.Sprintf(
				"color=c=0x2B2B31:s=%dx%d:d=%.3f:r=%d,format=rgba,geq=r='r(X,Y)':g='g(X,Y)':b='b(X,Y)':a='if(lte(hypot(X-%d,Y-%d),%d),255,0)'[avatar]",
				av, av, dur, fps, av/2, av/2, av/2))
			chain = append(chain, fmt.Sprintf("[%s][avatar]overlay=%d:%d[%s]", prev, avatarX, layout.ProfileY, "withav"))
			prev = "withav"
			textFilters = append(textFilters, fmt.Sprintf(
				"drawtext=fontfile='%s':textfile='%s':fontsize=%d:fontcolor=%s:x=%d:y=%d",
				escapeFFPath(fontPath), escapeFFPath(initFile), av/2,
				hexToFFColor("#FFFFFF"), avatarX+av/2-av/8, layout.ProfileY+av/2-av/4))
		}

		drawName := fmt.Sprintf(
			"drawtext=fontfile='%s':textfile='%s':fontsize=%d:fontcolor=%s:x=%s:y=%d",
			escapeFFPath(fontPath), escapeFFPath(nameFile), spec.Profile.NameSize,
			hexToFFColor(spec.Text.Color), xExpr(align, nameX), nameY)
		if spec.Text.StrokeEnabled {
			drawName += fmt.Sprintf(":borderw=%d:bordercolor=%s", spec.Text.StrokeWidth, hexToFFColor(spec.Text.StrokeColor))
		}
		textFilters = append(textFilters, drawName)

		handle := strings.TrimSpace(spec.Profile.Handle)
		if handle != "" {
			handleFile := stem + ".handle.txt"
			if err := os.WriteFile(handleFile, []byte(handle), 0o644); err != nil {
				return nil, err
			}
			textFilters = append(textFilters, fmt.Sprintf(
				"drawtext=fontfile='%s':textfile='%s':fontsize=%d:fontcolor=%s@0.72:x=%s:y=%d",
				escapeFFPath(fontPath), escapeFFPath(handleFile), spec.Profile.HandleSize,
				hexToFFColor(spec.Text.Color), xExpr(align, nameX), handleY))
		}
	}

	headline := strings.TrimSpace(in.Plan.Headline)
	if headline != "" && spec.Text.Position != "hidden" {
		lines := template.WrapText(headline, spec.Text.Size, spec.Canvas.Width-spec.Text.SideSpacing*2, 3)
		headline = strings.Join(lines, "\n")
		hlineFile := stem + ".headline.txt"
		if err := os.WriteFile(hlineFile, []byte(headline), 0o644); err != nil {
			return nil, err
		}
		textAlign := "center"
		if spec.Text.Align == "left" {
			textAlign = "left"
		}
		draw := fmt.Sprintf(
			"drawtext=fontfile='%s':textfile='%s':fontsize=%d:fontcolor=%s:x=%s:y=%d:line_spacing=%d:text_align=%s",
			escapeFFPath(fontPath), escapeFFPath(hlineFile), spec.Text.Size,
			hexToFFColor(spec.Text.Color), xExpr(spec.Text.Align, 0),
			layout.TextY+spec.Text.Offset, spec.Text.Size/4, textAlign)
		if spec.Text.StrokeEnabled && !spec.Text.Band {
			draw += fmt.Sprintf(":borderw=%d:bordercolor=%s", spec.Text.StrokeWidth, hexToFFColor(spec.Text.StrokeColor))
		}
		textFilters = append(textFilters, draw)
	}

	if len(textFilters) > 0 {
		chain = append(chain, fmt.Sprintf("[%s]%s[vout]", prev, strings.Join(textFilters, ",")))
	} else {
		chain = append(chain, fmt.Sprintf("[%s]null[vout]", prev))
	}

	args := []string{
		"-hide_banner", "-loglevel", "error", "-nostdin",
		"-progress", "pipe:1",
		"-ss", fmt.Sprintf("%.3f", float64(in.Clip.StartMS)/1000),
		"-i", in.SourcePath,
	}
	if useLogo {
		args = append(args, "-i", in.LogoPath)
	}
	args = append(args,
		"-t", fmt.Sprintf("%.3f", dur),
		"-filter_complex", strings.Join(chain, ";"),
		"-map", "[vout]",
	)
	if in.HasAudio {
		if speed != 1 {
			args = append(args, "-af", "atempo="+strconv.FormatFloat(speed, 'f', 4, 64))
		}
		args = append(args, "-map", "0:a:0", "-c:a", "aac", "-b:a", "128k")
	} else {
		args = append(args, "-an")
	}
	args = append(args,
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "20",
		"-pix_fmt", "yuv420p", "-r", strconv.Itoa(fps),
		"-movflags", "+faststart", "-t", fmt.Sprintf("%.3f", dur),
		"-y", in.OutPath,
	)
	return args, nil
}

func Thumbnail(ctx context.Context, videoPath, outPath string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", "error", "-nostdin",
		"-ss", "0.5", "-i", videoPath, "-frames:v", "1", "-vf", "scale=540:-2", "-q:v", "4", "-y", outPath)
	return cmd.Run()
}

type limitedBuffer struct {
	buf   strings.Builder
	limit int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.buf.Len() < b.limit {
		remaining := b.limit - b.buf.Len()
		if remaining > len(p) {
			remaining = len(p)
		}
		b.buf.Write(p[:remaining])
	}
	return len(p), nil
}

func (b *limitedBuffer) String() string { return b.buf.String() }

func lastLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, " | ")
}

func hexToFFColor(hex string) string {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(hex) != 6 {
		return "0x0E0E10"
	}
	for _, r := range hex {
		if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
			return "0x0E0E10"
		}
	}
	return "0x" + strings.ToUpper(hex)
}

func escapeFFPath(p string) string {
	p = strings.ReplaceAll(p, `\`, `\\`)
	p = strings.ReplaceAll(p, ":", `\:`)
	p = strings.ReplaceAll(p, "'", `\'`)
	return p
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// orOne garante valores neutros quando o plano vem zerado.
func orOne(v float64) float64 {
	if v <= 0 {
		return 1
	}
	return v
}

func logoPosition(pos string, margin int) (string, string) {
	switch pos {
	case "top-left":
		return strconv.Itoa(margin), strconv.Itoa(margin)
	case "bottom-left":
		return strconv.Itoa(margin), fmt.Sprintf("H-h-%d", margin)
	case "bottom-right":
		return fmt.Sprintf("W-w-%d", margin), fmt.Sprintf("H-h-%d", margin)
	default: // top-right
		return fmt.Sprintf("W-w-%d", margin), strconv.Itoa(margin)
	}
}

func xExpr(align string, leftX int) string {
	if align == "left" && leftX > 0 {
		return strconv.Itoa(leftX)
	}
	return "(w-text_w)/2"
}
