package animation

import (
	"image"
	"image/gif"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// rasterFakeCanvas is a deterministic canvas.RasterCanvas: each Draw paints a
// solid frame whose color encodes the draw index, so tests can grab frames
// without a real backend.
type rasterFakeCanvas struct {
	fig   *core.Figure
	w, h  int
	count int
	last  *image.RGBA
}

func newRasterFakeCanvas(w, h int) *rasterFakeCanvas {
	return &rasterFakeCanvas{
		fig: &core.Figure{SizePx: geom.Pt{X: float64(w), Y: float64(h)}},
		w:   w,
		h:   h,
	}
}

func (c *rasterFakeCanvas) Figure() *core.Figure { return c.fig }

func (c *rasterFakeCanvas) Draw() error {
	img := image.NewRGBA(image.Rect(0, 0, c.w, c.h))
	shade := uint8((c.count * 40) % 256)
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = shade
		img.Pix[i+1] = shade
		img.Pix[i+2] = shade
		img.Pix[i+3] = 255
	}
	c.last = img
	c.count++
	return nil
}

func (c *rasterFakeCanvas) Resize(int, int) error { return nil }
func (c *rasterFakeCanvas) Connect(canvas.EventType, canvas.Handler) canvas.ConnectionID {
	return 0
}
func (c *rasterFakeCanvas) Disconnect(canvas.ConnectionID) {}
func (c *rasterFakeCanvas) Close() error                   { return nil }
func (c *rasterFakeCanvas) FrameRGBA() *image.RGBA         { return c.last }

var _ canvas.RasterCanvas = (*rasterFakeCanvas)(nil)

func newGIFAnimation(t *testing.T, cnv canvas.FigureCanvas, frames int) *Animation {
	t.Helper()
	anim, err := NewFuncAnimation(Config{Canvas: cnv, Frames: frames}, func(int) ([]core.Artist, error) {
		return nil, nil
	}, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	return anim
}

func TestSaveGIFEncodesDeterministicFrames(t *testing.T) {
	cnv := newRasterFakeCanvas(4, 3)
	anim := newGIFAnimation(t, cnv, 3)

	out := filepath.Join(t.TempDir(), "anim.gif")
	if err := anim.Save(out); err != nil {
		t.Fatalf("Save: %v", err)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("open gif: %v", err)
	}
	defer f.Close()
	decoded, err := gif.DecodeAll(f)
	if err != nil {
		t.Fatalf("decode gif: %v", err)
	}

	if got := len(decoded.Image); got != 3 {
		t.Fatalf("frame count = %d, want 3", got)
	}
	if decoded.LoopCount != 0 {
		t.Fatalf("loop count = %d, want 0 (infinite)", decoded.LoopCount)
	}
	// Default interval is 200ms → 5 fps → 20 centiseconds per frame.
	for i, d := range decoded.Delay {
		if d != 20 {
			t.Fatalf("frame %d delay = %d, want 20 centiseconds", i, d)
		}
	}
	for i, img := range decoded.Image {
		if b := img.Bounds(); b.Dx() != 4 || b.Dy() != 3 {
			t.Fatalf("frame %d bounds = %v, want 4x3", i, b)
		}
	}
}

func TestSaveGIFRespectsSaveCount(t *testing.T) {
	cnv := newRasterFakeCanvas(2, 2)
	anim := newGIFAnimation(t, cnv, 10)

	out := filepath.Join(t.TempDir(), "anim.gif")
	if err := anim.Save(out, WithSaveCount(4), WithFPS(10)); err != nil {
		t.Fatalf("Save: %v", err)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("open gif: %v", err)
	}
	defer f.Close()
	decoded, err := gif.DecodeAll(f)
	if err != nil {
		t.Fatalf("decode gif: %v", err)
	}
	if got := len(decoded.Image); got != 4 {
		t.Fatalf("frame count = %d, want 4 (save_count)", got)
	}
	// 10 fps → 10 centiseconds per frame.
	for i, d := range decoded.Delay {
		if d != 10 {
			t.Fatalf("frame %d delay = %d, want 10 centiseconds", i, d)
		}
	}
}

func TestSaveUnsupportedExtensionReturnsError(t *testing.T) {
	cnv := newRasterFakeCanvas(2, 2)
	anim := newGIFAnimation(t, cnv, 2)

	out := filepath.Join(t.TempDir(), "anim.unknownvideo")
	if err := anim.Save(out); err != ErrWriterUnsupported {
		t.Fatalf("Save(.unknownvideo) = %v, want ErrWriterUnsupported", err)
	}
}

func TestWriterByNameRegistry(t *testing.T) {
	w, err := WriterByName("pillow", 5)
	if err != nil {
		t.Fatalf("WriterByName(pillow): %v", err)
	}
	if _, ok := w.(*GifWriter); !ok {
		t.Fatalf("WriterByName(pillow) = %T, want *GifWriter", w)
	}

	for _, name := range []string{"imagemagick", "unknown"} {
		if _, err := WriterByName(name, 5); err != ErrWriterUnsupported {
			t.Fatalf("WriterByName(%q) = %v, want ErrWriterUnsupported", name, err)
		}
	}
}

func TestGifWriterRequiresRasterCanvas(t *testing.T) {
	// A plain canvas without RasterCanvas must be rejected at Setup.
	w := NewGifWriter(5)
	if err := w.Setup(plainCanvas{}, "x.gif", 0); err != ErrWriterUnsupported {
		t.Fatalf("Setup(non-raster) = %v, want ErrWriterUnsupported", err)
	}
}

// plainCanvas implements canvas.FigureCanvas but not canvas.RasterCanvas.
type plainCanvas struct{}

func (plainCanvas) Figure() *core.Figure { return &core.Figure{SizePx: geom.Pt{X: 2, Y: 2}} }
func (plainCanvas) Draw() error          { return nil }
func (plainCanvas) Resize(int, int) error {
	return nil
}
func (plainCanvas) Connect(canvas.EventType, canvas.Handler) canvas.ConnectionID { return 0 }
func (plainCanvas) Disconnect(canvas.ConnectionID)                               {}
func (plainCanvas) Close() error                                                 { return nil }
