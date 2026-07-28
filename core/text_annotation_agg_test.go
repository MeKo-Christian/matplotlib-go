package core_test

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestAnnotationCurvedArrowMatchesMatplotlibGalleryPath(t *testing.T) {
	fig := core.NewFigure(1040, 720)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.55}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetXLim(0, 8)
	ax.SetYLim(-1.25, 1.25)

	blue := render.Color{R: 0.12, G: 0.31, B: 0.68, A: 1}
	arrow, _ := core.ArrowStyleFromString("-|>,head_length=0.35,head_width=0.20")
	arc, _ := core.ConnectionStyleFromString("arc3,rad=0.25")
	ax.Annotate("curved arrow\nbbox label", math.Pi/2, 1, core.AnnotationOptions{
		OffsetX:         optional.Of(pointsToPixels(fig.RC.DPI, 68)),
		OffsetY:         optional.Of(pointsToPixels(fig.RC.DPI, -42)),
		OffsetUnits:     core.AnnotationOffsetPixels,
		FontSize:        10,
		HAlign:          core.TextAlignCenter,
		VAlign:          core.TextVAlignMiddle,
		ArrowStyle:      arrow,
		ConnectionStyle: arc,
		ArrowColor:      blue,
		// ArrowWidth is in points, like Matplotlib's lw; it must not be
		// pre-converted the way the pixel-unit offsets above are.
		ArrowWidth: optional.Of(1.2),
		BBox: optional.Of(core.TextBBoxOptions{
			Padding:      pointsToPixels(fig.RC.DPI, 0.28*10),
			FaceColor:    render.Color{R: 0.92, G: 0.97, B: 1, A: 0.9},
			EdgeColor:    blue,
			LineWidth:    pointsToPixels(fig.RC.DPI, 0.9),
			CornerRadius: 5,
		}),
	})

	r, err := agg.New(1040, 720, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("new agg renderer: %v", err)
	}
	rec := &annotationPathRecorder{Renderer: r, stroke: blue}
	core.DrawFigure(fig, rec)

	if len(rec.curves) == 0 {
		t.Fatalf("did not record curved annotation arrow path")
	}
	got := rec.curves[0]
	want := []geom.Pt{
		{X: 337.089, Y: 572.873},
		{X: 307.591, Y: 604.587},
		{X: 262.783, Y: 609.436},
	}
	if len(got.V) < len(want) {
		t.Fatalf("annotation arrow curve vertices = %+v, want at least %+v", got.V, want)
	}
	for i, pt := range want {
		if distance(got.V[i], pt) > 1.0 {
			t.Fatalf("annotation arrow vertex %d = %+v, want Matplotlib gallery %+v; full path=%+v", i, got.V[i], pt, got)
		}
	}
}

// Matplotlib clips the annotation arrow against the text's layout box
// (Text.get_window_extent), not against its glyph ink. This reproduces the
// transform_coordinates fixture, whose recorded Matplotlib geometry is
//
//	get_window_extent : x0=432.458 y0=249.968 x1=500.208 y1=263.968  (67.75 x 14.0)
//	posA -> posB      : (466.333, 256.968) -> (548.208, 289.968)
//	shaft             : (493.170, 267.785) (519.401, 278.357) (543.831, 288.204)
//
// which is the same y-up space the renderer receives paths in. The ink box is
// narrower and sits higher than the layout box, so using it tilts the shaft and
// walks its tail several pixels along the text.
//
// All three vertices are pinned. The tail and the control point are what the
// patch box decides; the tip additionally depends on how much the "->" head
// trims off the shaft, which is why the arrow style is set explicitly rather
// than left to Axes.Annotate's "-|>" default.
func TestAnnotationArrowClipsAgainstTextLayoutBoxNotInk(t *testing.T) {
	fig := core.NewFigure(720, 420)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.13, Y: 0.16}, Max: geom.Pt{X: 0.90, Y: 0.84}})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)

	ink := render.Color{R: 0.10, G: 0.10, B: 0.10, A: 1}
	arrowStyle, ok := core.ArrowStyleFromString("->")
	if !ok {
		t.Fatal(`ArrowStyleFromString("->")`)
	}
	ax.Annotate("axes note", 0.82, 0.78, core.AnnotationOptions{
		Coords:     core.Coords(core.CoordAxes),
		ArrowStyle: arrowStyle,
		OffsetX:    optional.Of(-48.0),
		OffsetY:    optional.Of(-26.0),
		FontSize:   10,
		Color:      ink,
		ArrowColor: ink,
		ArrowWidth: optional.Of(1.25),
		HAlign:     core.TextAlignRight,
		VAlign:     core.TextVAlignTop,
	})

	r, err := agg.New(720, 420, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("new agg renderer: %v", err)
	}
	rec := &annotationPathRecorder{Renderer: r, stroke: ink}
	core.DrawFigure(fig, rec)

	if len(rec.curves) == 0 {
		t.Fatalf("did not record annotation arrow shaft")
	}
	got := rec.curves[0]
	want := []struct {
		pt  geom.Pt
		tol float64
	}{
		{geom.Pt{X: 493.170, Y: 267.785}, 0.01},
		{geom.Pt{X: 519.401, Y: 278.357}, 0.01},
		{geom.Pt{X: 543.831, Y: 288.204}, 0.01},
	}
	if len(got.V) != len(want) {
		t.Fatalf("annotation arrow shaft vertices = %+v, want %d quadratic vertices", got.V, len(want))
	}
	for i, w := range want {
		if distance(got.V[i], w.pt) > w.tol {
			t.Fatalf("annotation arrow vertex %d = %+v, want Matplotlib %+v within %g; full path=%+v", i, got.V[i], w.pt, w.tol, got.V)
		}
	}
}

type annotationPathRecorder struct {
	*agg.Renderer
	stroke render.Color
	curves []geom.Path
}

func (r *annotationPathRecorder) Path(path geom.Path, paint *render.Paint) {
	if paint != nil && paint.Stroke == r.stroke && paint.Fill.A == 0 && containsPathCmd(path, geom.QuadTo) {
		r.curves = append(r.curves, geom.Path{
			C: append([]geom.Cmd(nil), path.C...),
			V: append([]geom.Pt(nil), path.V...),
		})
	}
	r.Renderer.Path(path, paint)
}

func containsPathCmd(path geom.Path, want geom.Cmd) bool {
	for _, cmd := range path.C {
		if cmd == want {
			return true
		}
	}
	return false
}

func pointsToPixels(dpi, points float64) float64 {
	return points * dpi / 72
}

func distance(a, b geom.Pt) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}
