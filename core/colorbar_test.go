package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestFigureAddColorbarConfiguresAxes(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	img := ax.Image([][]float64{
		{0, 1},
		{2, 3},
	})

	cbAx := fig.AddColorbar(ax, img, ColorbarOptions{Label: "Intensity"})
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}

	base := geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	}
	wantWidth := resolvedColorbarWidth(fig, base, 0, defaultColorbarAspect)
	wantPadding := resolvedColorbarPadding(base, 0)
	if wantPadding != base.W()*0.05 {
		t.Fatalf("default colorbar padding = %v, want 5%% of parent width", wantPadding)
	}
	if got, want := ax.RectFraction.Max.X, base.Min.X+base.W()*(1-defaultColorbarFraction-defaultColorbarPadding); !floatApprox(got, want, 1e-12) {
		t.Fatalf("expected parent to reserve colorbar space: got right=%v want %v", got, want)
	}
	if got, want := cbAx.RectFraction.W(), wantWidth; !floatApprox(got, want, 1e-12) {
		t.Fatalf("expected colorbar width to follow default aspect: got %v want %v", got, want)
	}
	if got, want := cbAx.RectFraction.Min.X, base.Min.X+base.W()*(1-defaultColorbarFraction); !floatApprox(got, want, 1e-12) {
		t.Fatalf("expected colorbar to start at matplotlib default slot: got left=%v want %v", got, want)
	}
	if cbAx.RectFraction.Min.X <= ax.RectFraction.Max.X {
		t.Fatalf("expected colorbar to be placed to the right, got %+v", cbAx.RectFraction)
	}
	if cbAx.RectFraction.Min.Y != ax.RectFraction.Min.Y || cbAx.RectFraction.Max.Y != ax.RectFraction.Max.Y {
		t.Fatalf("expected colorbar to share vertical extent, got %+v", cbAx.RectFraction)
	}
	if cbAx.XAxis.ShowSpine || cbAx.XAxis.ShowTicks || cbAx.XAxis.ShowLabels {
		t.Fatalf("expected hidden colorbar x-axis, got %+v", cbAx.XAxis)
	}
	if cbAx.YAxis.ShowSpine || cbAx.YAxis.ShowTicks || cbAx.YAxis.ShowLabels {
		t.Fatalf("expected hidden left-side y-axis, got %+v", cbAx.YAxis)
	}
	if cbAx.YAxisRight == nil {
		t.Fatal("expected explicit right-side y-axis")
	}
	if cbAx.YAxisRight.Side != AxisRight {
		t.Fatalf("expected right-side y-axis, got %v", cbAx.YAxisRight.Side)
	}
	if !cbAx.YAxisRight.ShowLabels || !cbAx.YAxisRight.ShowTicks {
		t.Fatalf("expected visible right-side colorbar ticks and labels, got %+v", cbAx.YAxisRight)
	}
	if cbAx.YAxis.ShowLabels || cbAx.YAxis.ShowTicks {
		t.Fatalf("expected hidden left-side colorbar ticks and labels, got %+v", cbAx.YAxis)
	}
	if cbAx.effectiveYLabelSide() != AxisRight {
		t.Fatalf("expected colorbar label on right side")
	}
	if cbAx.YLabel != "Intensity" {
		t.Fatalf("unexpected colorbar label %q", cbAx.YLabel)
	}

	yMin, yMax := cbAx.YScale.Domain()
	if yMin != 0 || yMax != 3 {
		t.Fatalf("unexpected colorbar limits %v..%v", yMin, yMax)
	}
}

func TestColorbarDrawRendersGradientAndTickLabels(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	img := ax.Image([][]float64{
		{0, 1},
		{2, 3},
	})

	cbAx := fig.AddColorbar(ax, img, ColorbarOptions{Label: "Value"})
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}

	var r colorbarRecordingRenderer
	DrawFigure(fig, &r)

	if r.imageCount < 1 {
		t.Fatalf("expected heatmap image, got %d", r.imageCount)
	}
	if r.pathCount == 0 {
		t.Fatal("expected colorbar border/axes paths to be rendered")
	}
	if len(r.texts) == 0 {
		t.Fatal("expected tick labels or colorbar label to be rendered")
	}
}

func TestFigureAddColorbarUsesOriginalSubplotPositionAfterSetPosition(t *testing.T) {
	fig := NewFigure(640, 360)
	gs := fig.GridSpec(
		1,
		1,
		WithGridSpecPadding(0.125, 0.9, 0.11, 0.88),
		WithGridSpecSpacing(0, 0),
	)
	ax := gs.Cell(0, 0).AddAxes()
	ax.SetPosition(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.90, Y: 0.88},
	})
	img := ax.Image([][]float64{
		{0, 1},
		{2, 3},
	})

	cbAx := fig.AddColorbar(ax, img)
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}
	DrawFigure(fig, &colorbarRecordingRenderer{})

	base := geom.Rect{
		Min: geom.Pt{X: 0.125, Y: 0.11},
		Max: geom.Pt{X: 0.9, Y: 0.88},
	}
	wantParent := colorbarParentRect(
		base,
		resolvedColorbarWidth(fig, base, 0, defaultColorbarAspect),
		resolvedColorbarPadding(base, 0),
		false,
	)
	if !rectsApprox(ax.RectFraction, wantParent, 1e-12) {
		t.Fatalf("parent rect = %+v, want %+v", ax.RectFraction, wantParent)
	}
	if got, want := cbAx.RectFraction.Min.Y, base.Min.Y; !floatApprox(got, want, 1e-12) {
		t.Fatalf("colorbar bottom = %v, want original subplot bottom %v", got, want)
	}
	if got, want := cbAx.RectFraction.Max.Y, base.Max.Y; !floatApprox(got, want, 1e-12) {
		t.Fatalf("colorbar top = %v, want original subplot top %v", got, want)
	}
}

func rectsApprox(got, want geom.Rect, tol float64) bool {
	return floatApprox(got.Min.X, want.Min.X, tol) &&
		floatApprox(got.Min.Y, want.Min.Y, tol) &&
		floatApprox(got.Max.X, want.Max.X, tol) &&
		floatApprox(got.Max.Y, want.Max.Y, tol)
}

func TestFigureAddColorbarUsesLogNormTicks(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	img := ax.Image([][]float64{
		{1, 10},
		{10, 100},
	}, ImageOptions{Norm: LogNorm{VMin: 1, VMax: 100}})

	cbAx := fig.AddColorbar(ax, img)
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}
	if _, ok := cbAx.YAxisRight.Locator.(LogLocator); !ok {
		t.Fatalf("right colorbar locator = %T, want LogLocator", cbAx.YAxisRight.Locator)
	}
	if _, ok := cbAx.YAxisRight.Formatter.(LogFormatter); !ok {
		t.Fatalf("right colorbar formatter = %T, want LogFormatter", cbAx.YAxisRight.Formatter)
	}
	yMin, yMax := cbAx.YScale.Domain()
	if yMin != 1 || yMax != 100 {
		t.Fatalf("log colorbar domain = %v..%v, want 1..100", yMin, yMax)
	}
}

func TestFigureAddColorbarUsesBoundaryNormTicks(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	mesh := ax.PColorMesh([][]float64{{0.5, 1.5}}, MeshOptions{
		Norm: BoundaryNorm{Boundaries: []float64{0, 1, 2}, NColors: 3},
	})

	cbAx := fig.AddColorbar(ax, mesh)
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}
	loc, ok := cbAx.YAxisRight.Locator.(FixedLocator)
	if !ok {
		t.Fatalf("right colorbar locator = %T, want FixedLocator", cbAx.YAxisRight.Locator)
	}
	want := []float64{0, 1, 2}
	if len(loc.TicksList) != len(want) {
		t.Fatalf("boundary ticks = %v, want %v", loc.TicksList, want)
	}
	for i := range want {
		if loc.TicksList[i] != want[i] {
			t.Fatalf("boundary tick %d = %v, want %v", i, loc.TicksList[i], want[i])
		}
	}
}

func TestFigureAddColorbarUsesFunctionScaleForTwoSlopeNorm(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	img := ax.Image([][]float64{
		{-3, 0},
		{0, 6},
	}, ImageOptions{Norm: TwoSlopeNorm{VMin: -3, VCenter: 0, VMax: 6}})

	cbAx := fig.AddColorbar(ax, img)
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}
	if _, ok := cbAx.YScale.(transform.FuncScale); !ok {
		t.Fatalf("colorbar y scale = %T, want transform.FuncScale for TwoSlopeNorm", cbAx.YScale)
	}
	if got := cbAx.YScale.Fwd(0); got < 0.49 || got > 0.51 {
		t.Fatalf("TwoSlopeNorm colorbar center maps to %v, want about 0.5", got)
	}
}

func TestFigureAddColorbarShrinksAxesForExtensions(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.90, Y: 0.88},
	})
	mesh := ax.PColorMesh([][]float64{{-1, 0, 2}}, MeshOptions{
		XEdges: []float64{0, 1, 2, 3},
		YEdges: []float64{0, 1},
	})
	base := colorbarBaseRect(ax)

	cbAx := fig.AddColorbar(ax, mesh, ColorbarOptions{Extend: "both"})
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}
	if !(cbAx.RectFraction.Min.Y > base.Min.Y && cbAx.RectFraction.Max.Y < base.Max.Y) {
		t.Fatalf("extended colorbar rect = %+v, want vertical inset within base %+v", cbAx.RectFraction, base)
	}
}

func TestExtendedColorbarAdjustedLayoutKeepsMatplotlibWidth(t *testing.T) {
	fig := NewFigure(640, 360)
	gs := fig.GridSpec(
		1,
		1,
		WithGridSpecPadding(0.125, 0.9, 0.11, 0.88),
		WithGridSpecSpacing(0, 0),
	)
	ax := gs.Cell(0, 0).AddAxes()
	ax.SetPosition(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.90, Y: 0.88},
	})
	mesh := ax.PColorMesh([][]float64{{-1, 0, 2}}, MeshOptions{
		XEdges: []float64{0, 1, 2, 3},
		YEdges: []float64{0, 1},
	})

	cbAx := fig.AddColorbar(ax, mesh, ColorbarOptions{Extend: "both"})
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}

	got := cbAx.adjustedLayout(fig)
	base := colorbarBaseRect(ax)
	wantWidth := resolvedColorbarWidth(fig, base, 0, defaultColorbarAspect) * fig.SizePx.X
	if !floatApprox(got.W(), wantWidth, 1e-12) {
		t.Fatalf("extended colorbar adjusted pixel width = %v, want Matplotlib non-shrunk width %v", got.W(), wantWidth)
	}
}

func TestColorbarDrawAddsExtensionTriangles(t *testing.T) {
	var r colorbarRecordingRenderer
	clip := geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 42, Y: 80},
	}

	cb := &Colorbar{
		Mapping:     ScalarMapInfo{Colormap: "gray", VMin: 0, VMax: 1}.Resolved(),
		Extend:      "both",
		Alpha:       1,
		BorderColor: render.Color{A: 1},
		BorderWidth: 1,
	}
	ctx := &DrawContext{Clip: clip}
	cb.Draw(&r, ctx)
	cb.DrawOverlay(&r, ctx)

	if len(r.paths) != 259 {
		t.Fatalf("path count = %d, want 256 cells plus 2 extensions plus outline", len(r.paths))
	}
	if len(r.paths[256].V) < 3 || len(r.paths[257].V) < 3 {
		t.Fatalf("extension paths should be triangles, got lens %d and %d", len(r.paths[256].V), len(r.paths[257].V))
	}
}

func TestColorbarExtensionsDrawOutsideAxesClip(t *testing.T) {
	var r clipTrackingRenderer
	clip := geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 42, Y: 80},
	}
	cb := &Colorbar{
		Mapping:     ScalarMapInfo{Colormap: "gray", VMin: 0, VMax: 1}.Resolved(),
		Extend:      "both",
		Alpha:       1,
		BorderColor: render.Color{A: 1},
		BorderWidth: 1,
	}
	ctx := &DrawContext{Clip: clip}

	r.Save()
	r.ClipRect(clip)
	cb.Draw(&r, ctx)
	r.Restore()

	if r.clippedOutsidePaths != 0 {
		t.Fatalf("extension paths drawn while axes clip was active = %d, want 0", r.clippedOutsidePaths)
	}
	overlay, ok := any(cb).(OverlayArtist)
	if !ok {
		t.Fatal("colorbar with extensions should draw overlay outside the axes clip")
	}
	overlay.DrawOverlay(&r, ctx)
	if r.unclippedOutsidePaths == 0 {
		t.Fatal("expected colorbar extension overlay paths outside the axes clip")
	}
}

func TestColorbarDrawSnapsRangeLegendToPixels(t *testing.T) {
	var r colorbarRecordingRenderer
	clip := geom.Rect{
		Min: geom.Pt{X: 10.4, Y: 20.6},
		Max: geom.Pt{X: 42.6, Y: 80.2},
	}

	(&Colorbar{Colormap: "inferno", Alpha: 1, BorderColor: render.Color{A: 1}, BorderWidth: 1}).Draw(&r, &DrawContext{Clip: clip})

	if len(r.imageRects) != 0 {
		t.Fatalf("image rect count = %d, want 0", len(r.imageRects))
	}
	if len(r.paths) != 257 {
		t.Fatalf("path count = %d, want 256 color cells plus outline", len(r.paths))
	}
	wantFirstCell := []geom.Pt{
		{X: 10, Y: 80},
		{X: 43, Y: 80},
		{X: 43, Y: 81},
		{X: 10, Y: 81},
	}
	for i, want := range wantFirstCell {
		if !floatApprox(r.paths[0].V[i].X, want.X, 1e-12) || !floatApprox(r.paths[0].V[i].Y, want.Y, 1e-12) {
			t.Fatalf("first cell vertex %d = %+v, want %+v", i, r.paths[0].V[i], want)
		}
	}
	wantOutline := []geom.Pt{
		{X: 10.5, Y: 21.5},
		{X: 43.5, Y: 21.5},
		{X: 43.5, Y: 80.5},
		{X: 10.5, Y: 80.5},
	}
	outline := r.paths[len(r.paths)-1]
	if len(outline.V) < len(wantOutline) {
		t.Fatalf("outline vertices = %v, want at least %d", outline.V, len(wantOutline))
	}
	for i, want := range wantOutline {
		if !floatApprox(outline.V[i].X, want.X, 1e-12) || !floatApprox(outline.V[i].Y, want.Y, 1e-12) {
			t.Fatalf("outline vertex %d = %+v, want %+v", i, outline.V[i], want)
		}
	}
}

type colorbarRecordingRenderer struct {
	render.NullRenderer
	imageCount int
	pathCount  int
	texts      []string
	imageRects []geom.Rect
	paths      []geom.Path
	fills      []render.Color
}

func (r *colorbarRecordingRenderer) Image(_ render.Image, dst geom.Rect) {
	r.imageCount++
	r.imageRects = append(r.imageRects, dst)
}

func (r *colorbarRecordingRenderer) Path(path geom.Path, paint *render.Paint) {
	r.pathCount++
	r.paths = append(r.paths, path)
	if paint != nil {
		r.fills = append(r.fills, paint.Fill)
	}
}

func (r *colorbarRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *colorbarRecordingRenderer) DrawText(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
}

type clipTrackingRenderer struct {
	render.NullRenderer
	clipActive            bool
	clip                  geom.Rect
	clippedOutsidePaths   int
	unclippedOutsidePaths int
}

func (r *clipTrackingRenderer) Save() {}

func (r *clipTrackingRenderer) Restore() {
	r.clipActive = false
}

func (r *clipTrackingRenderer) ClipRect(rect geom.Rect) {
	r.clipActive = true
	r.clip = rect
}

func (r *clipTrackingRenderer) Path(path geom.Path, _ *render.Paint) {
	if !pathOutsideRect(path, r.clip) {
		return
	}
	if r.clipActive {
		r.clippedOutsidePaths++
		return
	}
	r.unclippedOutsidePaths++
}

func pathOutsideRect(path geom.Path, rect geom.Rect) bool {
	const snapTolerance = 1.1
	for _, pt := range path.V {
		if pt.X < rect.Min.X-snapTolerance || pt.X > rect.Max.X+snapTolerance ||
			pt.Y < rect.Min.Y-snapTolerance || pt.Y > rect.Max.Y+snapTolerance {
			return true
		}
	}
	return false
}
