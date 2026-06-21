package core_test

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
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
		OffsetX:         pointsToPixels(fig.RC.DPI, 68),
		OffsetY:         pointsToPixels(fig.RC.DPI, -42),
		OffsetUnits:     core.AnnotationOffsetPixels,
		FontSize:        10,
		HAlign:          core.TextAlignCenter,
		VAlign:          core.TextVAlignMiddle,
		ArrowStyle:      arrow,
		ConnectionStyle: arc,
		ArrowColor:      blue,
		ArrowWidth:      pointsToPixels(fig.RC.DPI, 1.2),
		BBox: &core.TextBBoxOptions{
			Padding:      pointsToPixels(fig.RC.DPI, 0.28*10),
			FaceColor:    render.Color{R: 0.92, G: 0.97, B: 1, A: 0.9},
			EdgeColor:    blue,
			LineWidth:    pointsToPixels(fig.RC.DPI, 0.9),
			CornerRadius: 5,
		},
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
