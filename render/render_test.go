package render

import (
	"image"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

// Compile-time assertion also present in render.go, kept here to guard against accidental changes.
var _ Renderer = (*NullRenderer)(nil)

type fakeImage struct{ w, h int }

func (f fakeImage) Size() (int, int)      { return f.w, f.h }
func (f fakeImage) Interpolation() string { return "" }

func TestNullRenderer_NoPanicAndStackBalance(t *testing.T) {
	var r NullRenderer
	vp := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	if err := r.Begin(vp); err != nil {
		t.Fatalf("begin: %v", err)
	}

	// Save/Restore balance
	r.Save()
	r.Save()
	if d := r.depth(); d != 2 {
		t.Fatalf("depth want 2 got %d", d)
	}
	r.Restore()
	r.Restore()
	r.Restore() // extra restore should clamp to 0
	if d := r.depth(); d != 0 {
		t.Fatalf("depth want 0 got %d", d)
	}

	// Drawing verbs should not panic
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.LineTo(geom.Pt{X: 10, Y: 10})
	r.ClipRect(vp)
	r.ClipPath(p)
	r.Path(p, &Paint{})
	r.DrawImage(fakeImage{w: 10, h: 10}, vp)
	r.GlyphRun(GlyphRun{}, Color{})
	_ = r.MeasureText("hi", 12, "default")

	if err := r.End(); err != nil {
		t.Fatalf("end: %v", err)
	}
}

func TestNullRenderer_BeginEndOrder(t *testing.T) {
	var r NullRenderer
	// End before begin should error
	if err := r.End(); err == nil {
		t.Fatalf("expected error on End before Begin")
	}
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("begin: %v", err)
	}
	// Double begin should error
	if err := r.Begin(geom.Rect{}); err == nil {
		t.Fatalf("expected error on double Begin")
	}
	if err := r.End(); err != nil {
		t.Fatalf("end: %v", err)
	}
}

func TestColorPremultiply(t *testing.T) {
	c := Color{R: 0.3, G: 0.7, B: 0.9, A: 0.6}

	got := c.Premultiply()
	want := Color{R: 0.18, G: 0.42, B: 0.54, A: 0.6}

	if got != want {
		t.Fatalf("Premultiply() = %+v, want %+v", got, want)
	}
}

func TestColorWithAlphaMultiplier(t *testing.T) {
	color := Color{R: 0.2, G: 0.4, B: 0.6, A: 0.8}
	got := color.WithAlphaMultiplier(0.25)
	want := Color{R: 0.2, G: 0.4, B: 0.6, A: 0.2}
	if got != want {
		t.Fatalf("WithAlphaMultiplier() = %+v, want %+v", got, want)
	}
	if color.A != 0.8 {
		t.Fatalf("WithAlphaMultiplier mutated receiver alpha to %v", color.A)
	}
}

func TestColorToPremultipliedRGBA(t *testing.T) {
	c := Color{R: 0.3, G: 0.7, B: 0.9, A: 0.6}

	r, g, b, a := c.ToPremultipliedRGBA()
	if r != 46 || g != 107 || b != 138 || a != 153 {
		t.Fatalf("ToPremultipliedRGBA() = (%d,%d,%d,%d), want (46,107,138,153)", r, g, b, a)
	}
}

func TestImage_InterpolationIsPartOfInterface(t *testing.T) {
	var img Image = fakeImage{w: 10, h: 5}
	if img.Interpolation() != "" {
		t.Fatalf("default Interpolation = %q, want empty", img.Interpolation())
	}
}

func TestImageData_SetInterpolation(t *testing.T) {
	img := NewImageData(image.NewRGBA(image.Rect(0, 0, 4, 4)))
	if img.Interpolation() != "" {
		t.Fatalf("zero-value Interpolation = %q, want empty", img.Interpolation())
	}
	img.SetInterpolation("bilinear")
	if img.Interpolation() != "bilinear" {
		t.Fatalf("after SetInterpolation: %q, want bilinear", img.Interpolation())
	}
}

func TestImageData_DefaultAlphaIsOne(t *testing.T) {
	img := NewImageData(image.NewRGBA(image.Rect(0, 0, 4, 4)))
	if got := img.Alpha(); got != 1 {
		t.Fatalf("NewImageData alpha = %v, want 1", got)
	}
}

func TestImageData_SetAlpha(t *testing.T) {
	img := NewImageData(image.NewRGBA(image.Rect(0, 0, 4, 4)))
	img.SetAlpha(0.37)
	if got := img.Alpha(); got != 0.37 {
		t.Fatalf("after SetAlpha: %v, want 0.37", got)
	}
}

func TestImageData_AlphaClamped(t *testing.T) {
	img := NewImageData(image.NewRGBA(image.Rect(0, 0, 4, 4)))
	img.SetAlpha(1.5)
	if got := img.Alpha(); got != 1 {
		t.Fatalf("SetAlpha(1.5) = %v, want 1", got)
	}
	img.SetAlpha(-0.2)
	if got := img.Alpha(); got != 0 {
		t.Fatalf("SetAlpha(-0.2) = %v, want 0", got)
	}
}
