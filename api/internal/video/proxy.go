package video

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// NeedsProxy indica se vale converter a origem antes de renderizar.
// Fontes VP9/AV1/4K são lentas e podem travar o decodificador.
func NeedsProxy(info ProbeInfo) bool {
	if info.Codec != "h264" {
		return true
	}
	if info.Width*info.Height > 1920*1080 {
		return true
	}
	if info.FPS > 31 {
		return true
	}
	return false
}

// Proxy cria uma versão 1080p H.264 apenas do trecho usado no render.
// Assim uma origem de 2 horas não precisa ser convertida inteira.
func Proxy(ctx context.Context, srcPath, outPath string, startMS, endMS int64, hasAudio bool, onProgress func(int)) error {
	if _, err := os.Stat(outPath); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	tmp := outPath + ".tmp.mp4"
	durMS := endMS - startMS
	if durMS <= 0 {
		return errors.New("trecho inválido para o proxy")
	}
	durSec := float64(durMS) / 1000

	args := []string{
		"-hide_banner", "-loglevel", "error", "-nostdin", "-progress", "pipe:1",
		"-ss", fmt.Sprintf("%.3f", float64(startMS)/1000),
		"-i", srcPath,
		"-t", fmt.Sprintf("%.3f", durSec),
		"-vf", "scale='min(1920,iw)':-2:force_original_aspect_ratio=decrease,fps=30",
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "22",
		"-pix_fmt", "yuv420p", "-movflags", "+faststart",
	}
	if hasAudio {
		args = append(args, "-c:a", "aac", "-b:a", "128k")
	} else {
		args = append(args, "-an")
	}
	args = append(args, "-y", tmp)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Cancel = func() error { return cmd.Process.Kill() }
	stderr := &limitedBuffer{limit: 4000}
	cmd.Stderr = stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("proxy: ffmpeg não iniciou: %w", err)
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
			if pct > 99 {
				pct = 99
			}
			if pct > int(last.Load()) {
				last.Store(int64(pct))
				onProgress(pct)
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("proxy falhou: %s", lastLines(strings.TrimSpace(stderr.String()), 4))
	}
	if err := os.Rename(tmp, outPath); err != nil {
		return err
	}
	if onProgress != nil {
		onProgress(100)
	}
	return nil
}

// Render executa o FFmpeg com watchdog: se não houver progresso por
// stallTimeout, o processo é encerrado para não travar a fila.
func Render(ctx context.Context, in RenderInput, onProgress func(int)) error {
	return renderWithWatchdog(ctx, in, onProgress, 90*time.Second)
}

func renderWithWatchdog(ctx context.Context, in RenderInput, onProgress func(int), stallTimeout time.Duration) error {
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
	var lastPct atomic.Int64
	var lastActivity atomic.Int64
	lastActivity.Store(time.Now().UnixNano())

	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			line := sc.Text()
			if !strings.HasPrefix(line, "out_time_us=") {
				continue
			}
			us, err := strconv.ParseInt(strings.TrimPrefix(line, "out_time_us="), 10, 64)
			if err != nil {
				continue
			}
			lastActivity.Store(time.Now().UnixNano())
			if onProgress == nil {
				continue
			}
			pct := int(float64(us) / 1e6 / durSec * 100)
			if pct < 0 {
				pct = 0
			}
			if pct > 100 {
				pct = 100
			}
			if pct > int(lastPct.Load()) {
				lastPct.Store(int64(pct))
				onProgress(pct)
			}
		}
	}()

	done := make(chan struct{})
	var killed atomic.Bool
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				idle := time.Since(time.Unix(0, lastActivity.Load()))
				if idle > stallTimeout {
					killed.Store(true)
					_ = cmd.Process.Kill()
					return
				}
			}
		}
	}()

	err = cmd.Wait()
	close(done)
	if killed.Load() {
		return errors.New("render travou sem progresso e foi encerrado (veja o log do worker)")
	}
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("ffmpeg falhou: %s", lastLines(msg, 6))
	}
	return nil
}
