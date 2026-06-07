package animation

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestSaveHTMLEmbedsSelfContainedPNGFrames(t *testing.T) {
	cnv := newRasterFakeCanvas(4, 3)
	anim := newGIFAnimation(t, cnv, 3)

	out := filepath.Join(t.TempDir(), "anim.html")
	if err := anim.Save(out); err != nil {
		t.Fatalf("Save: %v", err)
	}

	html := readHTMLFile(t, out)
	if !strings.Contains(html, "<!doctype html>") {
		t.Fatalf("html missing doctype")
	}
	if strings.Contains(html, "http://") || strings.Contains(html, "https://") {
		t.Fatalf("html should be self-contained, got external URL")
	}
	if !strings.Contains(html, `"fps":5`) {
		t.Fatalf("html missing default fps metadata: %s", html)
	}
	if !strings.Contains(html, `"width":4`) || !strings.Contains(html, `"height":3`) {
		t.Fatalf("html missing frame size metadata: %s", html)
	}

	frames := embeddedHTMLFrames(t, html)
	if got := len(frames); got != 3 {
		t.Fatalf("embedded frame count = %d, want 3", got)
	}
	img, err := png.Decode(bytes.NewReader(frames[0]))
	if err != nil {
		t.Fatalf("decode first embedded frame: %v", err)
	}
	if b := img.Bounds(); b.Dx() != 4 || b.Dy() != 3 {
		t.Fatalf("first embedded frame bounds = %v, want 4x3", b)
	}
}

func TestSaveHTMLRespectsSaveCountAndFPS(t *testing.T) {
	cnv := newRasterFakeCanvas(2, 2)
	anim := newGIFAnimation(t, cnv, 10)

	out := filepath.Join(t.TempDir(), "anim.htm")
	if err := anim.Save(out, WithSaveCount(4), WithFPS(10)); err != nil {
		t.Fatalf("Save: %v", err)
	}

	html := readHTMLFile(t, out)
	if !strings.Contains(html, `"fps":10`) {
		t.Fatalf("html missing fps metadata: %s", html)
	}
	if !strings.Contains(html, `"frameDelayMS":100`) {
		t.Fatalf("html missing frame delay metadata: %s", html)
	}
	if got := len(embeddedHTMLFrames(t, html)); got != 4 {
		t.Fatalf("embedded frame count = %d, want 4", got)
	}
}

func TestWriterByNameHTMLRegistry(t *testing.T) {
	w, err := WriterByName("html", 5)
	if err != nil {
		t.Fatalf("WriterByName(html): %v", err)
	}
	if _, ok := w.(*HTMLWriter); !ok {
		t.Fatalf("WriterByName(html) = %T, want *HTMLWriter", w)
	}
}

func TestHTMLWriterRequiresRasterCanvas(t *testing.T) {
	w := NewHTMLWriter(5)
	if err := w.Setup(plainCanvas{}, "x.html", 0); err != ErrWriterUnsupported {
		t.Fatalf("Setup(non-raster) = %v, want ErrWriterUnsupported", err)
	}
}

func TestHTMLWriterFinishWithoutFrames(t *testing.T) {
	w := NewHTMLWriter(5)
	if err := w.Setup(newRasterFakeCanvas(2, 2), filepath.Join(t.TempDir(), "empty.html"), 0); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if err := w.Finish(); err != ErrNoFramesGrabbed {
		t.Fatalf("Finish without frames = %v, want ErrNoFramesGrabbed", err)
	}
}

func readHTMLFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read html: %v", err)
	}
	return string(data)
}

func embeddedHTMLFrames(t *testing.T, html string) [][]byte {
	t.Helper()
	re := regexp.MustCompile(`data:image/png;base64,([^"]+)`)
	matches := re.FindAllStringSubmatch(html, -1)
	frames := make([][]byte, 0, len(matches))
	for i, match := range matches {
		payload, err := strconv.Unquote(`"` + match[1] + `"`)
		if err != nil {
			t.Fatalf("unquote embedded frame %d: %v", i, err)
		}
		data, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			t.Fatalf("decode embedded frame %d: %v", i, err)
		}
		frames = append(frames, data)
	}
	return frames
}
