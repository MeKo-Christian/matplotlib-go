package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestContourRCDefaultsAndExplicitOptionsPrecedence(t *testing.T) {
	fig := NewFigure(320, 240)
	fig.RC.Contour.Algorithm = "serial"
	fig.RC.Contour.CornerMask = false
	fig.RC.Contour.LineWidth = 2.75
	fig.RC.Contour.LineWidthSet = true
	fig.RC.Contour.NegativeLineStyle = "dotted"
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	black := render.Color{A: 1}
	data := [][]float64{{-2, -1, 0}, {-1, 0, 1}, {0, 1, 2}}
	fromRC := ax.Contour(data, ContourOptions{Levels: []float64{-1, 1}, Color: &black})
	if fromRC == nil || fromRC.Lines == nil {
		t.Fatal("expected contour lines")
	}
	if fromRC.Algorithm != "serial" || fromRC.CornerMask {
		t.Fatalf("resolved structured defaults = algorithm %q, corner mask %v", fromRC.Algorithm, fromRC.CornerMask)
	}
	if got := fromRC.Lines.LineWidth; got != 2.75 {
		t.Fatalf("rc contour linewidth = %v, want 2.75", got)
	}
	assertNegativeContourDash(t, fromRC, lineStyleToDashes("dotted", 2.75))

	width := 4.0
	cornerMask := true
	solid := "solid"
	explicit := ax.Contour(data, ContourOptions{
		Levels:             []float64{-1, 1},
		Color:              &black,
		Algorithm:          "mpl2014",
		CornerMask:         &cornerMask,
		LineWidth:          &width,
		NegativeLineStyles: &solid,
	})
	if explicit == nil || explicit.Lines == nil {
		t.Fatal("expected explicit contour lines")
	}
	if explicit.Algorithm != "mpl2014" || !explicit.CornerMask || explicit.Lines.LineWidth != width {
		t.Fatalf("explicit contour defaults did not win: %+v / %+v", explicit, explicit.Lines)
	}
	assertNegativeContourDash(t, explicit, nil)
}

func TestTriContourUsesContourLineRCDefaults(t *testing.T) {
	fig := NewFigure(320, 240)
	fig.RC.Contour.LineWidth = 3.25
	fig.RC.Contour.LineWidthSet = true
	fig.RC.Contour.NegativeLineStyle = "dashdot"
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	black := render.Color{A: 1}
	tri := Triangulation{
		X:         []float64{0, 1, 0, 1},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {1, 3, 2}},
	}
	cs := ax.TriContour(tri, []float64{-2, 2, 0, 2}, ContourOptions{
		Levels: []float64{-1, 1},
		Color:  &black,
	})
	if cs == nil || cs.Lines == nil {
		t.Fatal("expected triangular contour lines")
	}
	if got := cs.Lines.LineWidth; got != 3.25 {
		t.Fatalf("tri contour rc linewidth = %v, want 3.25", got)
	}
	assertNegativeContourDash(t, cs, lineStyleToDashes("dashdot", 3.25))
}

func TestStructuredContourCornerMaskControlsMaskedCornerGeometry(t *testing.T) {
	data := [][]float64{{math.NaN(), 1}, {0, 2}}
	levels := []float64{0, 2}

	makeContourf := func(cornerMask bool) *ContourSet {
		fig := NewFigure(200, 200)
		fig.RC.Contour.CornerMask = cornerMask
		ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
		return ax.Contourf(data, ContourOptions{Levels: levels})
	}
	if got := makeContourf(false); got != nil {
		t.Fatalf("corner_mask=False retained masked cell geometry: %+v", got.Fills)
	}
	masked := makeContourf(true)
	if masked == nil || masked.Fills == nil || len(masked.Fills.Polygons) == 0 {
		t.Fatal("corner_mask=True did not retain the valid triangular portion")
	}
	for _, polygon := range masked.Fills.Polygons {
		for _, point := range polygon {
			if point == (geom.Pt{X: 0, Y: 0}) {
				t.Fatalf("masked corner leaked into contourf polygon: %+v", polygon)
			}
		}
	}
}

func TestStructuredContourMPL2005CornerMaskRules(t *testing.T) {
	fig := NewFigure(200, 200)
	fig.RC.Contour.Algorithm = "mpl2005"
	fig.RC.Contour.CornerMask = true
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	data := [][]float64{{0, 1}, {1, 2}}

	cs := ax.Contour(data, ContourOptions{Levels: []float64{0.5}})
	if cs == nil {
		t.Fatal("expected mpl2005 contour")
	}
	if cs.Algorithm != "mpl2005" || cs.CornerMask {
		t.Fatalf("mpl2005 resolved algorithm/corner mask = %q/%v", cs.Algorithm, cs.CornerMask)
	}

	cornerMask := true
	if got := ax.Contour(data, ContourOptions{
		Levels:     []float64{0.5},
		Algorithm:  "mpl2005",
		CornerMask: &cornerMask,
	}); got != nil {
		t.Fatal("mpl2005 accepted unsupported explicit corner_mask=True")
	}
	if got := ax.Contour(data, ContourOptions{Levels: []float64{0.5}, Algorithm: "unknown"}); got != nil {
		t.Fatal("invalid explicit contour algorithm was accepted")
	}
}

func TestContourLinewidthNoneFallsBackToLinesLinewidth(t *testing.T) {
	fig := NewFigure(200, 200)
	fig.RC = style.Default
	fig.RC.LineWidth = 1.875
	fig.RC.Contour.LineWidthSet = false
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	cs := ax.Contour([][]float64{{0, 1}, {1, 2}}, ContourOptions{Levels: []float64{0.5}})
	if cs == nil || cs.Lines == nil {
		t.Fatal("expected contour lines")
	}
	if got := cs.Lines.LineWidth; got != 1.875 {
		t.Fatalf("inherited lines.linewidth = %v, want 1.875", got)
	}
}

func TestContourNamedDashesUseActiveLineRCAndDeviceScaling(t *testing.T) {
	fig := NewFigure(200, 200)
	fig.RC.Lines.DashedPattern = []float64{4, 2}
	fig.RC.Lines.ScaleDashes = false
	fig.RC.Contour.NegativeLineStyle = "dashed"
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	black := render.Color{A: 1}
	cs := ax.Contour(
		[][]float64{{-2, -1, 0}, {-1, 0, 1}, {0, 1, 2}},
		ContourOptions{Levels: []float64{-1, 1}, Color: &black},
	)
	if cs == nil || cs.Lines == nil {
		t.Fatal("expected contour lines")
	}
	var pattern []float64
	for i, level := range cs.lineLevels {
		if level < 0 && i < len(cs.Lines.DashPatterns) {
			pattern = cs.Lines.DashPatterns[i]
			break
		}
	}
	if len(pattern) != 2 || pattern[0] != 4 || pattern[1] != 2 {
		t.Fatalf("unscaled contour dash pattern = %v, want [4 2] points", pattern)
	}

	r := &recordingRenderer{}
	ctx := createTestDrawContext()
	ctx.RC = fig.RC
	cs.Lines.Draw(r, ctx)
	pointPx := pointsToPixels(ctx.RC, 1)
	found := false
	for _, call := range r.pathCalls {
		if len(call.paint.Dashes) == 2 {
			found = true
			if math.Abs(call.paint.Dashes[0]-4*pointPx) > 1e-12 ||
				math.Abs(call.paint.Dashes[1]-2*pointPx) > 1e-12 {
				t.Fatalf("device contour dashes = %v, want [%v %v]", call.paint.Dashes, 4*pointPx, 2*pointPx)
			}
		}
	}
	if !found {
		t.Fatal("did not draw a dashed negative contour")
	}
}

func assertNegativeContourDash(t *testing.T, cs *ContourSet, want []float64) {
	t.Helper()
	for i, level := range cs.lineLevels {
		if level >= 0 {
			continue
		}
		got := cs.Lines.DashPatterns
		if want == nil {
			if len(got) > i && len(got[i]) != 0 {
				t.Fatalf("negative level %v dashes = %v, want solid", level, got[i])
			}
			return
		}
		if len(got) <= i || len(got[i]) != len(want) {
			t.Fatalf("negative level %v dashes = %v, want %v", level, got, want)
		}
		for j := range want {
			if math.Abs(got[i][j]-want[j]) > 1e-12 {
				t.Fatalf("negative level %v dashes = %v, want %v", level, got[i], want)
			}
		}
		return
	}
	t.Fatal("test contour had no negative line level")
}
