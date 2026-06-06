//go:build ffmpeg

package animation

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFFmpegWritersRegisterWhenBuildTagged(t *testing.T) {
	t.Setenv("PATH", fakeFFmpegDir(t, "ffmpeg"))

	mp4, err := WriterByName("ffmpeg", 12)
	if err != nil {
		t.Fatalf("WriterByName(ffmpeg): %v", err)
	}
	if _, ok := mp4.(*FFmpegWriter); !ok {
		t.Fatalf("WriterByName(ffmpeg) = %T, want *FFmpegWriter", mp4)
	}

	webm, err := WriterByName("ffmpeg-webm", 12)
	if err != nil {
		t.Fatalf("WriterByName(ffmpeg-webm): %v", err)
	}
	if got, ok := webm.(*FFmpegWriter); !ok || got.Format != "webm" {
		t.Fatalf("WriterByName(ffmpeg-webm) = %#v, want WebM *FFmpegWriter", webm)
	}
}

func TestSaveMP4StreamsRGBAFramesToFFmpeg(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "ffmpeg.log")
	t.Setenv("PATH", fakeFFmpegDir(t, "ffmpeg"))
	t.Setenv("MATPLOTLIB_GO_FAKE_FFMPEG_LOG", logPath)

	cnv := newRasterFakeCanvas(3, 2)
	anim := newGIFAnimation(t, cnv, 3)
	out := filepath.Join(t.TempDir(), "anim.mp4")

	if err := anim.Save(out, WithFPS(8)); err != nil {
		t.Fatalf("Save(.mp4): %v", err)
	}

	payload := readFakeFFmpegLog(t, logPath)
	if payload.Width != 3 || payload.Height != 2 {
		t.Fatalf("fake ffmpeg size = %dx%d, want 3x2", payload.Width, payload.Height)
	}
	if payload.FPS != 8 {
		t.Fatalf("fake ffmpeg fps = %d, want 8", payload.FPS)
	}
	if payload.Codec != "libx264" {
		t.Fatalf("fake ffmpeg codec = %q, want libx264", payload.Codec)
	}
	if payload.Bytes != 3*2*4*3 {
		t.Fatalf("fake ffmpeg received %d bytes, want one RGBA payload per frame", payload.Bytes)
	}
	if payload.Output != out {
		t.Fatalf("fake ffmpeg output = %q, want %q", payload.Output, out)
	}
}

func TestSaveWebMSelectsVP9Codec(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "ffmpeg.log")
	t.Setenv("PATH", fakeFFmpegDir(t, "ffmpeg"))
	t.Setenv("MATPLOTLIB_GO_FAKE_FFMPEG_LOG", logPath)

	cnv := newRasterFakeCanvas(2, 2)
	anim := newGIFAnimation(t, cnv, 1)
	out := filepath.Join(t.TempDir(), "anim.webm")

	if err := anim.Save(out, WithFPS(6)); err != nil {
		t.Fatalf("Save(.webm): %v", err)
	}

	payload := readFakeFFmpegLog(t, logPath)
	if payload.Codec != "libvpx-vp9" {
		t.Fatalf("fake ffmpeg codec = %q, want libvpx-vp9", payload.Codec)
	}
	if payload.Output != out {
		t.Fatalf("fake ffmpeg output = %q, want %q", payload.Output, out)
	}
}

func TestFFmpegWriterReportsUnavailableAtRuntime(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	if FFmpegAvailable() {
		t.Fatal("FFmpegAvailable() = true with empty PATH, want false")
	}

	cnv := newRasterFakeCanvas(2, 2)
	w := NewFFmpegWriter(5)
	err := w.Setup(cnv, filepath.Join(t.TempDir(), "anim.mp4"), 0)
	if !errors.Is(err, ErrWriterUnsupported) {
		t.Fatalf("Setup without ffmpeg = %v, want ErrWriterUnsupported", err)
	}
}

type fakeFFmpegPayload struct {
	Width  int
	Height int
	FPS    int
	Codec  string
	Bytes  int
	Output string
}

func fakeFFmpegDir(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	exe := filepath.Join(dir, name)
	if runtime.GOOS == "windows" {
		exe += ".bat"
	}
	script := `#!/bin/sh
set -eu
width=0
height=0
fps=0
codec=""
prev=""
for arg in "$@"; do
  case "$prev" in
    -s) width=${arg%x*}; height=${arg#*x};;
    -framerate) fps="$arg";;
    -vcodec) codec="$arg";;
  esac
  prev="$arg"
done
bytes=$(wc -c | tr -d ' ')
out=""
for arg in "$@"; do
  out="$arg"
done
printf '%s\n%s\n%s\n%s\n%s\n%s\n' "$width" "$height" "$fps" "$codec" "$bytes" "$out" > "$MATPLOTLIB_GO_FAKE_FFMPEG_LOG"
`
	if err := os.WriteFile(exe, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake ffmpeg: %v", err)
	}
	return dir + string(os.PathListSeparator) + os.Getenv("PATH")
}

func readFakeFFmpegLog(t *testing.T, path string) fakeFFmpegPayload {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fake ffmpeg log: %v", err)
	}
	lines := splitLines(string(data))
	if len(lines) < 6 {
		t.Fatalf("fake ffmpeg log has %d lines, want 6: %q", len(lines), data)
	}
	return fakeFFmpegPayload{
		Width:  atoiLog(t, lines[0]),
		Height: atoiLog(t, lines[1]),
		FPS:    atoiLog(t, lines[2]),
		Codec:  lines[3],
		Bytes:  atoiLog(t, lines[4]),
		Output: lines[5],
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func atoiLog(t *testing.T, s string) int {
	t.Helper()
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			t.Fatalf("invalid integer %q", s)
		}
		n = n*10 + int(s[i]-'0')
	}
	return n
}
