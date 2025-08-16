package render

import (
    "testing"
    "matplotlib-go/internal/geom"
)

// Compile-time assertion also present in render.go, kept here to guard against accidental changes.
var _ Renderer = (*NullRenderer)(nil)

type fakeImage struct{ w, h int }
func (f fakeImage) Size() (int, int) { return f.w, f.h }

func TestNullRenderer_NoPanicAndStackBalance(t *testing.T) {
    var r NullRenderer
    vp := geom.Rect{Min: geom.Pt{0,0}, Max: geom.Pt{100,100}}
    if err := r.Begin(vp); err != nil { t.Fatalf("begin: %v", err) }

    // Save/Restore balance
    r.Save(); r.Save()
    if d := r.depth(); d != 2 { t.Fatalf("depth want 2 got %d", d) }
    r.Restore(); r.Restore(); r.Restore() // extra restore should clamp to 0
    if d := r.depth(); d != 0 { t.Fatalf("depth want 0 got %d", d) }

    // Drawing verbs should not panic
    var p geom.Path
    p.MoveTo(geom.Pt{0,0}); p.LineTo(geom.Pt{10,10})
    r.ClipRect(vp)
    r.ClipPath(p)
    r.Path(p, Paint{})
    r.Image(fakeImage{w: 10, h: 10}, vp)
    r.GlyphRun(GlyphRun{}, Color{})
    _ = r.MeasureText("hi", 12, "default")

    if err := r.End(); err != nil { t.Fatalf("end: %v", err) }
}

func TestNullRenderer_BeginEndOrder(t *testing.T) {
    var r NullRenderer
    // End before begin should error
    if err := r.End(); err == nil { t.Fatalf("expected error on End before Begin") }
    if err := r.Begin(geom.Rect{}); err != nil { t.Fatalf("begin: %v", err) }
    // Double begin should error
    if err := r.Begin(geom.Rect{}); err == nil { t.Fatalf("expected error on double Begin") }
    if err := r.End(); err != nil { t.Fatalf("end: %v", err) }
}

