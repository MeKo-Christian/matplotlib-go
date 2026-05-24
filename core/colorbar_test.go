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

func TestFigureAddHorizontalColorbarConfiguresBottomAxes(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	img := ax.Image([][]float64{
		{0, 1},
		{2, 3},
	})

	cbAx := fig.AddColorbar(ax, img, ColorbarOptions{Location: "bottom", Label: "Intensity"})
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}

	base := geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	}
	wantHeight := resolvedColorbarThickness(fig, base, 0, defaultColorbarAspect, "bottom")
	wantSlotHeight := resolvedColorbarSlotThickness(base, 0, "bottom")
	wantPadding := resolvedColorbarPadding(base, 0, "bottom")
	if got, want := ax.RectFraction.Min.Y, base.Min.Y+wantSlotHeight+wantPadding; !floatApprox(got, want, 1e-12) {
		t.Fatalf("expected parent to reserve bottom colorbar space: got bottom=%v want %v", got, want)
	}
	if got, want := cbAx.RectFraction.H(), wantHeight; !floatApprox(got, want, 1e-12) {
		t.Fatalf("expected colorbar height to follow default aspect: got %v want %v", got, want)
	}
	if got, want := cbAx.RectFraction.Min.Y, base.Min.Y; !floatApprox(got, want, 1e-12) {
		t.Fatalf("expected bottom colorbar to start at base bottom: got %v want %v", got, want)
	}
	if cbAx.RectFraction.Max.Y >= ax.RectFraction.Min.Y {
		t.Fatalf("expected colorbar below parent axes, got colorbar=%+v parent=%+v", cbAx.RectFraction, ax.RectFraction)
	}
	if cbAx.RectFraction.Min.X != ax.RectFraction.Min.X || cbAx.RectFraction.Max.X != ax.RectFraction.Max.X {
		t.Fatalf("expected horizontal colorbar to share parent horizontal extent, got %+v", cbAx.RectFraction)
	}
	if cbAx.XAxis == nil || !cbAx.XAxis.ShowTicks || !cbAx.XAxis.ShowLabels {
		t.Fatalf("expected visible bottom x-axis ticks and labels, got %+v", cbAx.XAxis)
	}
	if cbAx.YAxis.ShowSpine || cbAx.YAxis.ShowTicks || cbAx.YAxis.ShowLabels {
		t.Fatalf("expected hidden horizontal colorbar y-axis, got %+v", cbAx.YAxis)
	}
	if cbAx.effectiveXLabelSide() != AxisBottom {
		t.Fatalf("expected horizontal colorbar label on bottom side")
	}
	if cbAx.XLabel != "Intensity" {
		t.Fatalf("unexpected colorbar x label %q", cbAx.XLabel)
	}
	xMin, xMax := cbAx.XScale.Domain()
	if xMin != 0 || xMax != 3 {
		t.Fatalf("unexpected horizontal colorbar limits %v..%v", xMin, xMax)
	}
	cb, ok := cbAx.Artists[0].(*Colorbar)
	if !ok {
		t.Fatalf("colorbar artist = %T, want *Colorbar", cbAx.Artists[0])
	}
	if cb.Orientation != "horizontal" {
		t.Fatalf("colorbar orientation = %q, want horizontal", cb.Orientation)
	}

	gotLayout := cbAx.adjustedLayout(fig)
	wantRect := cbAx.layout(fig)
	if !floatApprox(gotLayout.W(), wantRect.W(), 1e-12) {
		t.Fatalf("horizontal colorbar adjusted width = %v, want full slot width %v", gotLayout.W(), wantRect.W())
	}
	if !floatApprox(gotLayout.H(), wantRect.H(), 1e-12) {
		t.Fatalf("horizontal colorbar adjusted height = %v, want slot height %v", gotLayout.H(), wantRect.H())
	}
}

func TestFigureAddHorizontalColorbarConfiguresTopAxes(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	img := ax.Image([][]float64{{0, 1}, {2, 3}})

	cbAx := fig.AddColorbar(ax, img, ColorbarOptions{Location: "top", Label: "Intensity"})
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}

	base := geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.12}, Max: geom.Pt{X: 0.78, Y: 0.88}}
	wantSlotHeight := resolvedColorbarSlotThickness(base, 0, "top")
	wantPadding := resolvedColorbarPadding(base, 0, "top")
	if got, want := ax.RectFraction.Max.Y, base.Max.Y-wantSlotHeight-wantPadding; !floatApprox(got, want, 1e-12) {
		t.Fatalf("expected parent to reserve top colorbar space: got top=%v want %v", got, want)
	}
	if cbAx.RectFraction.Min.Y <= ax.RectFraction.Max.Y {
		t.Fatalf("expected top colorbar above parent axes, got colorbar=%+v parent=%+v", cbAx.RectFraction, ax.RectFraction)
	}
	if cbAx.XAxisTop == nil || !cbAx.XAxisTop.ShowTicks || !cbAx.XAxisTop.ShowLabels {
		t.Fatalf("expected visible top x-axis ticks and labels, got %+v", cbAx.XAxisTop)
	}
	if cbAx.XAxis.ShowTicks || cbAx.XAxis.ShowLabels {
		t.Fatalf("expected hidden bottom x-axis ticks and labels, got %+v", cbAx.XAxis)
	}
	if cbAx.effectiveXLabelSide() != AxisTop {
		t.Fatalf("expected horizontal top colorbar label on top side")
	}
}

func TestHorizontalColorbarDrawsGradientLeftToRight(t *testing.T) {
	var r colorbarRecordingRenderer
	clip := geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 42, Y: 80},
	}

	cb := &Colorbar{
		Mapping:     ScalarMapInfo{Colormap: "gray", VMin: 0, VMax: 1}.Resolved(),
		Orientation: "horizontal",
		Alpha:       1,
		BorderColor: render.Color{A: 1},
		BorderWidth: 1,
	}
	cb.Draw(&r, &DrawContext{Clip: clip})

	if len(r.paths) != 257 {
		t.Fatalf("path count = %d, want 256 color cells plus outline", len(r.paths))
	}
	first := r.paths[0]
	last := r.paths[255]
	if len(first.V) < 4 || len(last.V) < 4 {
		t.Fatalf("colorbar cell paths are malformed")
	}
	if first.V[0].X >= last.V[0].X {
		t.Fatalf("horizontal colorbar cells should advance left-to-right: first=%+v last=%+v", first.V[0], last.V[0])
	}
	if first.V[0].Y != last.V[0].Y || first.V[2].Y != last.V[2].Y {
		t.Fatalf("horizontal colorbar cells should share vertical span: first=%+v last=%+v", first.V, last.V)
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

func TestFigureAddColorbarUsesExplicitBoundariesAsTicks(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	img := ax.Image([][]float64{{0, 2}, {4, 5}})

	cbAx := fig.AddColorbar(ax, img, ColorbarOptions{Boundaries: []float64{0, 2, 5}, Values: []float64{1, 4}})
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}
	loc, ok := cbAx.YAxisRight.Locator.(FixedLocator)
	if !ok {
		t.Fatalf("right colorbar locator = %T, want FixedLocator", cbAx.YAxisRight.Locator)
	}
	want := []float64{0, 2, 5}
	if len(loc.TicksList) != len(want) {
		t.Fatalf("explicit boundary ticks = %v, want %v", loc.TicksList, want)
	}
	for i := range want {
		if loc.TicksList[i] != want[i] {
			t.Fatalf("explicit boundary tick %d = %v, want %v", i, loc.TicksList[i], want[i])
		}
	}
}

func TestFigureAddLeftColorbarUsesLeftBoundaryTicks(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.20, Y: 0.12},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	mesh := ax.PColorMesh([][]float64{{0.5, 1.5}}, MeshOptions{
		Norm: BoundaryNorm{Boundaries: []float64{0, 1, 2}, NColors: 3},
	})

	cbAx := fig.AddColorbar(ax, mesh, ColorbarOptions{Location: "left", Label: "band"})
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}
	if cbAx.RectFraction.Max.X >= ax.RectFraction.Min.X {
		t.Fatalf("expected left colorbar before parent axes, got colorbar=%+v parent=%+v", cbAx.RectFraction, ax.RectFraction)
	}
	if cbAx.YAxis == nil || !cbAx.YAxis.ShowTicks || !cbAx.YAxis.ShowLabels {
		t.Fatalf("expected visible left colorbar y-axis, got %+v", cbAx.YAxis)
	}
	if cbAx.YAxisRight != nil && (cbAx.YAxisRight.ShowTicks || cbAx.YAxisRight.ShowLabels) {
		t.Fatalf("expected hidden right y-axis for left colorbar, got %+v", cbAx.YAxisRight)
	}
	loc, ok := cbAx.YAxis.Locator.(FixedLocator)
	if !ok {
		t.Fatalf("left colorbar locator = %T, want FixedLocator", cbAx.YAxis.Locator)
	}
	want := []float64{0, 1, 2}
	if len(loc.TicksList) != len(want) {
		t.Fatalf("left boundary ticks = %v, want %v", loc.TicksList, want)
	}
	for i := range want {
		if loc.TicksList[i] != want[i] {
			t.Fatalf("left boundary tick %d = %v, want %v", i, loc.TicksList[i], want[i])
		}
	}
	if cbAx.effectiveYLabelSide() != AxisLeft {
		t.Fatalf("expected left colorbar label on left side")
	}
}

func TestFigureAddColorbarUsesExplicitTicks(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	img := ax.Image([][]float64{{-1, 0}, {0.5, 1}})

	cbAx := fig.AddColorbar(ax, img, ColorbarOptions{Ticks: []float64{-1, 0, 1}})
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}
	loc, ok := cbAx.YAxisRight.Locator.(FixedLocator)
	if !ok {
		t.Fatalf("right colorbar locator = %T, want FixedLocator", cbAx.YAxisRight.Locator)
	}
	want := []float64{-1, 0, 1}
	if len(loc.TicksList) != len(want) {
		t.Fatalf("explicit ticks = %v, want %v", loc.TicksList, want)
	}
	for i := range want {
		if loc.TicksList[i] != want[i] {
			t.Fatalf("explicit tick %d = %v, want %v", i, loc.TicksList[i], want[i])
		}
	}
}

func TestHorizontalColorbarUsesExplicitTicks(t *testing.T) {
	fig := NewFigure(900, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	img := ax.Image([][]float64{{-1, 0}, {0.5, 1}})

	cbAx := fig.AddColorbar(ax, img, ColorbarOptions{Location: "bottom", Ticks: []float64{-1, 0, 1}})
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}
	loc, ok := cbAx.XAxis.Locator.(FixedLocator)
	if !ok {
		t.Fatalf("bottom colorbar locator = %T, want FixedLocator", cbAx.XAxis.Locator)
	}
	want := []float64{-1, 0, 1}
	if len(loc.TicksList) != len(want) {
		t.Fatalf("explicit ticks = %v, want %v", loc.TicksList, want)
	}
	for i := range want {
		if loc.TicksList[i] != want[i] {
			t.Fatalf("explicit tick %d = %v, want %v", i, loc.TicksList[i], want[i])
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

func TestFigureColorbarSyncsMutableCollectionMapping(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.78, Y: 0.88},
	})
	pc := &PathCollection{
		Collection: Collection{Colormap: "viridis"},
		Path:       polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}, true),
		Offsets:    []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}},
		FaceColor:  render.Color{A: 1},
	}
	if err := pc.SetArray([]float64{0, 1}); err != nil {
		t.Fatalf("SetArray: %v", err)
	}
	ax.Add(pc)

	cbAx := fig.AddColorbar(ax, pc)
	if cbAx == nil {
		t.Fatal("expected colorbar axes")
	}
	if err := pc.SetCLim(-1, 2); err != nil {
		t.Fatalf("SetCLim: %v", err)
	}
	pc.SetColormap("plasma")

	DrawFigure(fig, &colorbarRecordingRenderer{})

	yMin, yMax := cbAx.YScale.Domain()
	if yMin != -1 || yMax != 2 {
		t.Fatalf("synced colorbar limits = %v..%v, want -1..2", yMin, yMax)
	}
	cb, ok := cbAx.Artists[0].(*Colorbar)
	if !ok {
		t.Fatalf("colorbar artist = %T, want *Colorbar", cbAx.Artists[0])
	}
	if cb.Mapping.VMin != -1 || cb.Mapping.VMax != 2 {
		t.Fatalf("synced colorbar mapping = %+v, want -1..2", cb.Mapping)
	}
	if cb.Mapping.Colormap != "plasma" {
		t.Fatalf("synced colorbar colormap = %q, want plasma", cb.Mapping.Colormap)
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

func TestBoundaryColorbarDrawUsesUniformSpacingByDefault(t *testing.T) {
	var r colorbarRecordingRenderer
	clip := geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 30, Y: 80},
	}
	cb := &Colorbar{
		Mapping: ScalarMapInfo{
			Colormap: "viridis",
			Norm:     BoundaryNorm{Boundaries: []float64{0, 1, 3}, NColors: 3},
			VMin:     0,
			VMax:     3,
		}.Resolved(),
		Alpha:       1,
		BorderColor: render.Color{A: 1},
		BorderWidth: 1,
	}

	cb.Draw(&r, &DrawContext{Clip: clip})

	if len(r.paths) != 3 {
		t.Fatalf("path count = %d, want 2 boundary cells plus outline", len(r.paths))
	}
	first, _ := pathBounds(r.paths[0])
	if !floatApprox(first.Min.Y, 50, 1e-12) || !floatApprox(first.Max.Y, 80, 1e-12) {
		t.Fatalf("first uniform boundary cell bounds = %+v, want y 50..80", first)
	}
}

func TestBoundaryColorbarDrawCanUseProportionalSpacing(t *testing.T) {
	var r colorbarRecordingRenderer
	clip := geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 30, Y: 80},
	}
	cb := &Colorbar{
		Mapping: ScalarMapInfo{
			Colormap: "viridis",
			Norm:     BoundaryNorm{Boundaries: []float64{0, 1, 3}, NColors: 3},
			VMin:     0,
			VMax:     3,
		}.Resolved(),
		Spacing:     "proportional",
		Alpha:       1,
		BorderColor: render.Color{A: 1},
		BorderWidth: 1,
	}

	cb.Draw(&r, &DrawContext{Clip: clip})

	if len(r.paths) != 3 {
		t.Fatalf("path count = %d, want 2 boundary cells plus outline", len(r.paths))
	}
	first, _ := pathBounds(r.paths[0])
	if !floatApprox(first.Min.Y, 60, 1e-12) || !floatApprox(first.Max.Y, 80, 1e-12) {
		t.Fatalf("first proportional boundary cell bounds = %+v, want y 60..80", first)
	}
}

func TestBoundaryColorbarDrawEdgesAddsInternalDividers(t *testing.T) {
	var r colorbarRecordingRenderer
	clip := geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 30, Y: 80},
	}
	cb := &Colorbar{
		Mapping: ScalarMapInfo{
			Colormap: "viridis",
			Norm:     BoundaryNorm{Boundaries: []float64{0, 1, 3, 6}, NColors: 3},
			VMin:     0,
			VMax:     6,
		}.Resolved(),
		DrawEdges:   true,
		Alpha:       1,
		BorderColor: render.Color{A: 1},
		BorderWidth: 1,
	}

	cb.Draw(&r, &DrawContext{Clip: clip})

	if len(r.strokes) != 3 {
		t.Fatalf("stroke count = %d, want 2 internal dividers plus outline", len(r.strokes))
	}
	firstDivider, _ := pathBounds(r.strokePaths[0])
	if !floatApprox(firstDivider.Min.Y, 60.5, 1e-12) || !floatApprox(firstDivider.Max.Y, 60.5, 1e-12) {
		t.Fatalf("first divider bounds = %+v, want y 60.5", firstDivider)
	}
}

func TestColorbarExtendRectDrawsRectangularExtensions(t *testing.T) {
	var r colorbarRecordingRenderer
	clip := geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 42, Y: 80},
	}
	cb := &Colorbar{
		Mapping:     ScalarMapInfo{Colormap: "gray", VMin: 0, VMax: 1}.Resolved(),
		Extend:      "both",
		ExtendRect:  true,
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
	lower, _ := pathBounds(r.paths[256])
	if !floatApprox(lower.Min.X, clip.Min.X, 1e-12) || !floatApprox(lower.Max.X, clip.Max.X, 1e-12) {
		t.Fatalf("lower rectangular extension bounds = %+v, want full colorbar width", lower)
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
	imageCount  int
	pathCount   int
	texts       []string
	imageRects  []geom.Rect
	paths       []geom.Path
	fills       []render.Color
	strokes     []render.Color
	strokePaths []geom.Path
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
		if paint.LineWidth > 0 {
			r.strokes = append(r.strokes, paint.Stroke)
			r.strokePaths = append(r.strokePaths, path)
		}
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
