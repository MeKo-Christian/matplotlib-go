package gio

import (
	"image"
	"testing"

	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"github.com/cwbudde/matplotlib-go/backends/desktop"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// stubRenderer is a minimal render.Renderer used to keep these tests
// cgo-free. The AGG backend is exercised through the embedding example.
// Only the verbs DrawFigure touches on an empty figure are interesting
// — everything else is a no-op.
type stubRenderer struct {
	w, h int
	img  *image.RGBA
}

func newStubRenderer(w, h int, _ render.Color) (render.Renderer, error) {
	return &stubRenderer{w: w, h: h, img: image.NewRGBA(image.Rect(0, 0, w, h))}, nil
}

func (r *stubRenderer) Begin(geom.Rect) error                  { return nil }
func (r *stubRenderer) End() error                             { return nil }
func (r *stubRenderer) Save()                                  {}
func (r *stubRenderer) Restore()                               {}
func (r *stubRenderer) ClipRect(geom.Rect)                     {}
func (r *stubRenderer) ClipPath(geom.Path)                     {}
func (r *stubRenderer) Path(geom.Path, *render.Paint)          {}
func (r *stubRenderer) Image(render.Image, geom.Rect)          {}
func (r *stubRenderer) GlyphRun(render.GlyphRun, render.Color) {}
func (r *stubRenderer) MeasureText(string, float64, string) render.TextMetrics {
	return render.TextMetrics{}
}

func (r *stubRenderer) GetImage() *image.RGBA { return r.img }

func newTestOptions() desktop.Options {
	fig := core.NewFigure(320, 240)
	return desktop.WithDefaults(desktop.Options{
		Figure:   fig,
		Title:    "test",
		Width:    320,
		Height:   240,
		Renderer: newStubRenderer,
	})
}

func TestRegisteredWithDesktop(t *testing.T) {
	if desktop.DefaultConstructor() == nil {
		t.Fatal("gio.init() did not register a desktop.Constructor")
	}
}

func TestNewBindsNavigationAndToolbar(t *testing.T) {
	b, err := New(newTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	if b.Canvas() == nil {
		t.Fatal("Canvas() returned nil")
	}
	if b.Navigation() == nil {
		t.Fatal("Navigation() returned nil")
	}
	if b.Toolbar() == nil {
		t.Fatal("Toolbar() returned nil")
	}
	// Toolbar pan toggles Navigation.
	tb := b.Toolbar()
	if err := tb.Trigger(canvas.ToolbarPan); err != nil {
		t.Fatal(err)
	}
	if got := b.Navigation().Mode(); got != canvas.NavPan {
		t.Fatalf("nav mode = %v, want NavPan", got)
	}
}

func TestCanvasDrawIdleEmitsEventOnDraw(t *testing.T) {
	b, err := New(newTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	got := 0
	b.Canvas().Connect(canvas.EventDraw, func(canvas.Event) error {
		got++
		return nil
	})
	if err := b.Canvas().Draw(); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("draw events = %d, want 1", got)
	}
}

func TestCanvasResizeEmitsEvent(t *testing.T) {
	b, err := New(newTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	got := 0
	b.Canvas().Connect(canvas.EventResize, func(ev canvas.Event) error {
		got++
		if ev.Width != 800 || ev.Height != 600 {
			t.Fatalf("resize event size = %d×%d, want 800×600", ev.Width, ev.Height)
		}
		return nil
	})
	if err := b.Canvas().Resize(800, 600); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("resize events = %d, want 1", got)
	}
	if b.Canvas().(*gioCanvas).Width() != 800 {
		t.Fatal("width not updated after Resize")
	}
}

func TestMapButtons(t *testing.T) {
	cases := []struct {
		in   pointer.Buttons
		want canvas.MouseButton
	}{
		{0, canvas.MouseButtonNone},
		{pointer.ButtonPrimary, canvas.MouseButtonLeft},
		{pointer.ButtonSecondary, canvas.MouseButtonRight},
		{pointer.ButtonTertiary, canvas.MouseButtonMiddle},
		{pointer.ButtonPrimary | pointer.ButtonSecondary, canvas.MouseButtonLeft},
	}
	for _, c := range cases {
		if got := mapButtons(c.in); got != c.want {
			t.Fatalf("mapButtons(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestMapModifiers(t *testing.T) {
	cases := []struct {
		in   key.Modifiers
		want canvas.Modifier
	}{
		{0, 0},
		{key.ModShift, canvas.ModifierShift},
		{key.ModCtrl, canvas.ModifierControl},
		{key.ModAlt, canvas.ModifierAlt},
		{key.ModSuper, canvas.ModifierMeta},
		{key.ModCommand, canvas.ModifierMeta},
		{key.ModShift | key.ModCtrl, canvas.ModifierShift | canvas.ModifierControl},
	}
	for _, c := range cases {
		if got := mapModifiers(c.in); got != c.want {
			t.Fatalf("mapModifiers(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestCloseIdempotent(t *testing.T) {
	b, err := New(newTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
}
