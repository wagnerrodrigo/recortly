package video

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/smashcut/smashcut/api/internal/model"
)

// ProbeInfo são os metadados lidos do arquivo de vídeo.
type ProbeInfo struct {
	DurationMS int64
	Width      int
	Height     int
	HasAudio   bool
	FPS        float64
}

// Probe lê metadados com ffprobe.
func Probe(ctx context.Context, path string) (ProbeInfo, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_format", "-show_streams", path)
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

// RenderInput reúne tudo que o worker precisa para gerar uma variação.
type RenderInput struct {
	SourcePath string
	OutPath    string
	Clip       model.Clip
	Spec       TemplateSpec
	Plan       VariationPlan
	HasAudio   bool
	FontPath   string
}

const (
	canvasW      = 1080
	canvasH      = 1920
	videoWidth   = 984
	profileBlock = 220
	footerPad    = 60
)

// BuildArgs monta os argumentos do FFmpeg (sem o binário).
func BuildArgs(in RenderInput) ([]string, error) {
	spec := in.Spec
	spec.Normalize()
	fps := spec.Canvas.FPS
	if fps <= 0 {
		fps = 30
	}

	bg := hexToFFColor(spec.Canvas.Background)
	textColor := hexToFFColor(spec.Text.Color)
	strokeColor := hexToFFColor(spec.Text.StrokeColor)

	start := float64(in.Clip.StartMS) / 1000
	end := float64(in.Clip.EndMS) / 1000
	dur := end - start
	if dur <= 0 {
		return nil, errors.New("clip com duração inválida")
	}

	// Região do vídeo dentro do canvas.
	textBlock := spec.Text.Size*2 + spec.Text.OffsetFromProfile + 20
	videoTop := profileBlock + textBlock
	vidH := canvasH - videoTop - footerPad
	if vidH%2 != 0 {
		vidH--
	}
	videoX := (canvasW - videoWidth) / 2

	zoom := in.Plan.Zoom
	if zoom <= 0 {
		zoom = 1
	}
	sat := in.Plan.Saturation
	if sat <= 0 {
		sat = 1
	}
	speed := in.Plan.Speed
	if speed <= 0 {
		speed = 1
	}

	// 6% de folga para permitir o zoom sem bordas.
	outW := int(float64(videoWidth) * 1.06 / zoom)
	outH := int(float64(vidH) * 1.06 / zoom)
	if outW%2 != 0 {
		outW++
	}
	if outH%2 != 0 {
		outH++
	}

	offsetX := in.Clip.OffsetX
	offsetY := in.Clip.OffsetY
	if offsetX < 0 || offsetX > 1 {
		offsetX = 0.5
	}
	if offsetY < 0 || offsetY > 1 {
		offsetY = 0.5
	}

	base := fmt.Sprintf(
		"fps=%d,scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d:(iw-%d)*%.4f:(ih-%d)*%.4f",
		fps, outW, outH, videoWidth, vidH, videoWidth, offsetX, vidH, offsetY)
	if in.Plan.Mirror {
		base += ",hflip"
	}
	base += fmt.Sprintf(",eq=saturation=%.3f", sat)

	var chain string
	switch in.Clip.Layout {
	case "fill":
		chain = fmt.Sprintf("[0:v]%s[s];[s]pad=%d:%d:%d:%d:color=%s[v]",
			base, canvasW, canvasH, videoX, videoTop, bg)
	default: // blur
		chain = fmt.Sprintf(
			"[0:v]%s,split[fg][bgsrc];"+
				"[bgsrc]scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d,boxblur=24:3[bgl];"+
				"[fg]scale=%d:%d[fgx];"+
				"[bgl][fgx]overlay=%d:%d[v]",
			base, canvasW, canvasH, canvasW, canvasH, videoWidth, vidH, videoX, videoTop)
	}

	fontPath := in.FontPath
	if fontPath == "" {
		fontPath = "/System/Library/Fonts/Helvetica.ttc"
	}
	stem := strings.TrimSuffix(in.OutPath, filepath.Ext(in.OutPath))

	textFilters := make([]string, 0, 3)

	name := strings.TrimSpace(spec.Profile.Name)
	if name == "" {
		name = "perfil"
	}
	nameFile := stem + ".name.txt"
	if err := os.WriteFile(nameFile, []byte(name), 0o644); err != nil {
		return nil, err
	}
	drawName := fmt.Sprintf(
		"drawtext=fontfile='%s':textfile='%s':fontsize=46:fontcolor=%s:x=(w-text_w)/2:y=76",
		escapeFFPath(fontPath), escapeFFPath(nameFile), textColor)
	if spec.Text.StrokeEnabled {
		drawName += fmt.Sprintf(":borderw=%d:bordercolor=%s", spec.Text.StrokeWidth, strokeColor)
	}
	textFilters = append(textFilters, drawName)

	handle := strings.TrimSpace(spec.Profile.Handle)
	if spec.Text.ShowHandle && handle != "" {
		handleFile := stem + ".handle.txt"
		if err := os.WriteFile(handleFile, []byte(handle), 0o644); err != nil {
			return nil, err
		}
		textFilters = append(textFilters, fmt.Sprintf(
			"drawtext=fontfile='%s':textfile='%s':fontsize=34:fontcolor=%s@0.75:x=(w-text_w)/2:y=136",
			escapeFFPath(fontPath), escapeFFPath(handleFile), textColor))
	}

	headline := strings.TrimSpace(in.Plan.Headline)
	if headline != "" {
		hlineFile := stem + ".headline.txt"
		if err := os.WriteFile(hlineFile, []byte(headline), 0o644); err != nil {
			return nil, err
		}
		textY := profileBlock + spec.Text.OffsetFromProfile
		drawHeadline := fmt.Sprintf(
			"drawtext=fontfile='%s':textfile='%s':fontsize=%d:fontcolor=%s:x=%d:y=%d:line_spacing=%d",
			escapeFFPath(fontPath), escapeFFPath(hlineFile), spec.Text.Size, textColor,
			spec.Text.SideSpacing, textY, spec.Text.Size/4)
		if spec.Text.StrokeEnabled {
			drawHeadline += fmt.Sprintf(":borderw=%d:bordercolor=%s", spec.Text.StrokeWidth, strokeColor)
		}
		textFilters = append(textFilters, drawHeadline)
	}

	chain += ";[v]" + strings.Join(textFilters, ",") + "[vout]"

	args := []string{
		"-hide_banner", "-loglevel", "error", "-nostdin",
		"-progress", "pipe:1",
		"-ss", fmt.Sprintf("%.3f", start),
		"-i", in.SourcePath,
		"-t", fmt.Sprintf("%.3f", dur),
		"-filter_complex", chain,
		"-map", "[vout]",
	}
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
		"-movflags", "+faststart",
		"-y", in.OutPath,
	)
	return args, nil
}

// Render executa o FFmpeg e reporta progresso (0-100) via callback.
func Render(ctx context.Context, in RenderInput, onProgress func(int)) error {
	args, err := BuildArgs(in)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Cancel = func() error { return cmd.Process.Kill() }
	stderr := &limitedBuffer{limit: 8000}
	cmd.Stderr = stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg não iniciou: %w", err)
	}

	durSec := float64(in.Clip.EndMS-in.Clip.StartMS) / 1000
	if durSec <= 0 {
		durSec = 1
	}
	var last atomic.Int64
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			line := sc.Text()
			if !strings.HasPrefix(line, "out_time_us=") {
				continue
			}
			us, err := strconv.ParseInt(strings.TrimPrefix(line, "out_time_us="), 10, 64)
			if err != nil || onProgress == nil {
				continue
			}
			pct := int(float64(us) / 1e6 / durSec * 100)
			if pct < 0 {
				pct = 0
			}
			if pct > 100 {
				pct = 100
			}
			if pct > int(last.Load()) {
				last.Store(int64(pct))
				onProgress(pct)
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("ffmpeg falhou: %s", lastLines(msg, 6))
	}
	return nil
}

// Thumbnail extrai um frame do vídeo gerado.
func Thumbnail(ctx context.Context, videoPath, outPath string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", "error", "-nostdin",
		"-ss", "1", "-i", videoPath, "-frames:v", "1", "-q:v", "3", "-y", outPath)
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

// escapeFFPath escapa caminhos para uso dentro de filtros do FFmpeg.
func escapeFFPath(p string) string {
	p = strings.ReplaceAll(p, `\`, `\\`)
	p = strings.ReplaceAll(p, ":", `\:`)
	p = strings.ReplaceAll(p, "'", `\'`)
	return p
}
