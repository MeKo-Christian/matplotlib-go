package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func TestAxesMatShowConfiguresMatrixView(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())

	img := ax.MatShow([][]float64{
		{1, 2, 3},
		{4, 5, 6},
	})
	if img == nil {
		t.Fatal("MatShow() returned nil")
	}
	if img.Origin != ImageOriginUpper {
		t.Fatalf("image origin = %v, want %v", img.Origin, ImageOriginUpper)
	}
	if img.XMin != -0.5 || img.XMax != 2.5 || img.YMin != -0.5 || img.YMax != 1.5 {
		t.Fatalf("image extent = [%v,%v]x[%v,%v], want [-0.5,2.5]x[-0.5,1.5]", img.XMin, img.XMax, img.YMin, img.YMax)
	}
	if !ax.YInverted() {
		t.Fatal("MatShow() should invert the y-axis")
	}
	if ax.XAxis == nil || ax.XAxis.ShowTicks || ax.XAxis.ShowLabels {
		t.Fatal("MatShow() should hide bottom x ticks and labels")
	}
	if ax.XAxisTop == nil || !ax.XAxisTop.ShowTicks || !ax.XAxisTop.ShowLabels {
		t.Fatal("MatShow() should show top x ticks and labels")
	}
	if ax.xLabelSide != AxisBottom {
		t.Fatalf("x label side = %v, want AxisBottom", ax.xLabelSide)
	}
	if _, ok := ax.XAxis.Locator.(MaxNLocator); !ok {
		t.Fatalf("x locator = %T, want MaxNLocator", ax.XAxis.Locator)
	}
	if _, ok := ax.XAxisTop.Locator.(MaxNLocator); !ok {
		t.Fatalf("top x locator = %T, want MaxNLocator", ax.XAxisTop.Locator)
	}
	if _, ok := ax.YAxis.Locator.(MaxNLocator); !ok {
		t.Fatalf("y locator = %T, want MaxNLocator", ax.YAxis.Locator)
	}
}

func TestAxesMatShowAcceptsSharedNorm(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())

	img := ax.MatShow([][]float64{
		{1, 10},
		{100, 1000},
	}, MatShowOptions{Norm: LogNorm{VMin: 1, VMax: 1000}})
	if img == nil {
		t.Fatal("MatShow() returned nil")
	}
	if img.Norm == nil || img.Norm.NormName() != "log" {
		t.Fatalf("MatShow norm = %#v, want log norm", img.Norm)
	}
}

func TestAxesMatShowInterpolationPropagatesToImage(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	bicubic := "bicubic"

	img := ax.MatShow([][]float64{{0, 1}}, MatShowOptions{Interpolation: &bicubic})
	if img == nil {
		t.Fatal("MatShow() returned nil")
	}
	if img.Interpolation != "bicubic" {
		t.Fatalf("MatShow interpolation = %q, want bicubic", img.Interpolation)
	}
}

func TestAxesImShowKeepsBottomXAxis(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())

	img := ax.ImShow([][]float64{
		{1, 2, 3},
		{4, 5, 6},
	})
	if img == nil {
		t.Fatal("ImShow() returned nil")
	}
	if img.Origin != ImageOriginUpper {
		t.Fatalf("image origin = %v, want %v", img.Origin, ImageOriginUpper)
	}
	if img.XMin != -0.5 || img.XMax != 2.5 || img.YMin != -0.5 || img.YMax != 1.5 {
		t.Fatalf("image extent = [%v,%v]x[%v,%v], want [-0.5,2.5]x[-0.5,1.5]", img.XMin, img.XMax, img.YMin, img.YMax)
	}
	if !ax.YInverted() {
		t.Fatal("ImShow() with upper origin should invert the y-axis")
	}
	if ax.XAxis == nil || !ax.XAxis.ShowTicks || !ax.XAxis.ShowLabels {
		t.Fatal("ImShow() should keep bottom x ticks and labels visible")
	}
	if ax.XAxisTop != nil {
		t.Fatal("ImShow() should not create a top x-axis")
	}
}

func TestAxesImShowAcceptsSharedNorm(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())

	img := ax.ImShow([][]float64{
		{1, 10},
		{100, 1000},
	}, ImShowOptions{Norm: LogNorm{VMin: 1, VMax: 1000}})
	if img == nil {
		t.Fatal("ImShow() returned nil")
	}
	if img.Norm == nil || img.Norm.NormName() != "log" {
		t.Fatalf("ImShow norm = %#v, want log norm", img.Norm)
	}
}

func TestAxesSpySupportsMarkerAndImageModes(t *testing.T) {
	data := [][]float64{
		{0, 1, 0},
		{2, 0, 0},
		{0, 0, 3},
	}

	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	result := ax.Spy(data, SpyOptions{Precision: 0.5})
	if result == nil {
		t.Fatal("Spy() returned nil for default image mode")
	}
	if result.Image == nil {
		t.Fatal("default Spy() should create an Image2D like matplotlib")
	}
	if result.Image.Colormap != "binary" {
		t.Fatalf("default Spy() image colormap = %q, want %q", result.Image.Colormap, "binary")
	}
	if result.Image.Interpolation != "nearest" {
		t.Fatalf("default Spy() image interpolation = %q, want %q", result.Image.Interpolation, "nearest")
	}
	if result.Markers != nil {
		t.Fatal("default Spy() should not create marker collection")
	}
	if got := len(result.Indices); got != 3 {
		t.Fatalf("len(indices) = %d, want 3", got)
	}
	if !ax.YInverted() {
		t.Fatal("Spy() should invert the y-axis")
	}
	if ax.XAxisTop == nil || !ax.XAxisTop.ShowTicks || !ax.XAxisTop.ShowLabels {
		t.Fatal("Spy() should show matrix-style top x ticks and labels")
	}

	fig = NewFigure(400, 300)
	ax = fig.AddAxes(unitRect())
	useImage := false
	result = ax.Spy(data, SpyOptions{UseImage: &useImage})
	if result == nil {
		t.Fatal("Spy() returned nil for marker mode")
	}
	if result.Markers == nil {
		t.Fatal("marker mode should create a PathCollection")
	}
	if result.Image != nil {
		t.Fatal("marker mode should not create an Image2D")
	}
}

func TestAxesSpyLeavesXLabelAtBottomWithTopTicks(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	ax.SetXLabel("column")
	ax.Spy([][]float64{{1, 0}, {0, 1}}, SpyOptions{MarkerSize: 10})

	if ax.effectiveXLabelSide() != AxisBottom {
		t.Fatalf("spy xlabel side = %v, want bottom like Matplotlib", ax.effectiveXLabelSide())
	}
	if ax.XAxis == nil || !ax.XAxis.ShowTicks || ax.XAxis.ShowLabels {
		t.Fatal("spy should keep bottom x tick marks visible while moving tick labels to the top")
	}
	if ax.XAxisTop == nil || !ax.XAxisTop.ShowTicks || !ax.XAxisTop.ShowLabels {
		t.Fatal("spy should place ticks and tick labels on the top axis")
	}

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	ctx.RC.DPI = 72
	px := ax.adjustedLayout(fig)
	r := &axesLabelRecordingRenderer{}
	drawAxesLabels(ax, r, ctx, px, figureTextAlignment{})
	if len(r.texts) != 1 || r.texts[0] != "column" {
		t.Fatalf("drawn texts = %v, want only xlabel", r.texts)
	}
	if r.origins[0].Y >= xLabelSpinePixelY(AxisBottom, px) {
		t.Fatalf("spy xlabel origin Y = %.3f, want below bottom spine %.3f", r.origins[0].Y, xLabelSpinePixelY(AxisBottom, px))
	}
}

func TestAxesSpyDrawFigureLeavesXLabelAtBottomOnce(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	ax.SetXLabel("column")
	ax.Spy([][]float64{{1, 0}, {0, 1}}, SpyOptions{MarkerSize: 10})

	r := &axesLabelRecordingRenderer{}
	DrawFigure(fig, r)

	var origins []geom.Pt
	for i, drawn := range r.texts {
		if drawn == "column" {
			origins = append(origins, r.origins[i])
		}
	}
	if len(origins) != 1 {
		t.Fatalf("drawn column origins = %v, want one xlabel", origins)
	}

	px := ax.adjustedLayout(fig)
	if origins[0].Y >= xLabelSpinePixelY(AxisBottom, px) {
		t.Fatalf("spy xlabel origin Y = %.3f, want below bottom spine %.3f", origins[0].Y, xLabelSpinePixelY(AxisBottom, px))
	}
}

func TestAxesSpyXLabelNotAffectedByEarlierTopXLabel(t *testing.T) {
	fig := NewFigure(1240, 620)

	topLabelAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.14},
		Max: geom.Pt{X: 0.31, Y: 0.88},
	})
	topLabelAx.SetXLabel("column")
	topLabelAx.Image([][]float64{
		{1, 2, 3, 4, 5},
		{2, 3, 4, 5, 6},
		{3, 4, 5, 6, 7},
		{4, 5, 6, 7, 8},
	}, ImageOptions{Origin: ImageOriginUpper})
	topLabelAx.SetXLim(-0.5, 4.5)
	topLabelAx.SetYLim(3.5, -0.5)
	_ = topLabelAx.SetAspect("equal")
	if topLabelAx.XAxis != nil {
		topLabelAx.XAxis.ShowTicks = false
		topLabelAx.XAxis.ShowLabels = false
	}
	if top := topLabelAx.TopAxis(); top != nil {
		top.ShowTicks = true
		top.ShowLabels = true
	}
	_ = topLabelAx.SetXLabelPosition("top")

	spyAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.69, Y: 0.14},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})
	spyAx.SetXLabel("column")
	spyData := make([][]float64, 18)
	for row := range spyData {
		spyData[row] = make([]float64, 18)
		for col := range spyData[row] {
			if row == col || row+col == 17 {
				spyData[row][col] = 1
			}
		}
	}
	spyAx.Spy(spyData, SpyOptions{MarkerSize: 10})
	if spyAx.effectiveXLabelSide() != AxisBottom {
		t.Fatalf("spy xlabel side before draw = %v, want bottom", spyAx.effectiveXLabelSide())
	}

	r := &axesLabelRecordingRenderer{}
	DrawFigure(fig, r)

	var spyOrigins []geom.Pt
	for i, drawn := range r.texts {
		if drawn == "column" && r.origins[i].X > 800 {
			spyOrigins = append(spyOrigins, r.origins[i])
		}
	}
	if len(spyOrigins) != 1 {
		t.Fatalf("spy column origins = %v, want one xlabel", spyOrigins)
	}
	px := spyAx.adjustedLayout(fig)
	if spyOrigins[0].Y >= xLabelSpinePixelY(AxisBottom, px) {
		t.Fatalf("spy xlabel origin Y = %.3f, want below bottom spine %.3f", spyOrigins[0].Y, xLabelSpinePixelY(AxisBottom, px))
	}
}

func TestAxesSpyMarkerSizeUsesMatplotlibPointDiameter(t *testing.T) {
	data := [][]float64{{1}}

	fig := NewFigure(400, 300)
	small := fig.AddAxes(unitRect()).Spy(data, SpyOptions{MarkerSize: 5})
	large := fig.AddAxes(unitRect()).Spy(data, SpyOptions{MarkerSize: 10})

	if small == nil || small.Markers == nil || large == nil || large.Markers == nil {
		t.Fatal("Spy(marker size) should use marker mode")
	}
	if !(large.Markers.Size > small.Markers.Size) {
		t.Fatalf("larger MarkerSize should increase rendered marker scale, got small=%v large=%v", small.Markers.Size, large.Markers.Size)
	}
	if got, want := large.Markers.Size/small.Markers.Size, 2.0; !almostEqualFloat(got, want) {
		t.Fatalf("marker scale ratio = %v, want %v", got, want)
	}
}

func TestAxesSpyMarkerSizeRoundsToPixelFootprintLikeLine2D(t *testing.T) {
	data := [][]float64{{1}}

	fig := NewFigure(400, 300)
	result := fig.AddAxes(unitRect()).Spy(data, SpyOptions{MarkerSize: 8})
	if result == nil || result.Markers == nil {
		t.Fatal("Spy(marker size) should use marker mode")
	}
	if got, want := result.Markers.Size, math.Ceil(pointsToPixels(fig.RC, 8)); !almostEqualFloat(got, want) {
		t.Fatalf("marker size = %v, want %v", got, want)
	}
}

func TestAxesSpyMarkerEdgeWidthUsesMatplotlibPointWidth(t *testing.T) {
	data := [][]float64{{1}}

	fig := NewFigure(400, 300)
	result := fig.AddAxes(unitRect()).Spy(data, SpyOptions{MarkerSize: 8})
	if result == nil || result.Markers == nil {
		t.Fatal("Spy(marker size) should use marker mode")
	}
	if got, want := result.Markers.EdgeWidth, pointsToPixels(fig.RC, 1); !almostEqualFloat(got, want) {
		t.Fatalf("marker edge width = %v, want %v", got, want)
	}
}

func TestAxesAnnotatedHeatmapAddsLabels(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	threshold := 2.5

	result := ax.AnnotatedHeatmap([][]float64{
		{1, 2},
		{3, 4},
	}, AnnotatedHeatmapOptions{
		Format:    "%.1f",
		Threshold: &threshold,
	})
	if result == nil {
		t.Fatal("AnnotatedHeatmap() returned nil")
	}
	if result.Image == nil {
		t.Fatal("AnnotatedHeatmap() should create an image")
	}
	if got := len(result.Labels); got != 4 {
		t.Fatalf("label count = %d, want 4", got)
	}
	if result.Labels[0].Content != "1.0" {
		t.Fatalf("first label text = %q, want %q", result.Labels[0].Content, "1.0")
	}
	if result.Labels[0].Color == result.Labels[3].Color {
		t.Fatal("expected low and high cells to use different text colors")
	}
}

func unitRect() geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	}
}

func TestImShow_ExtentOverridesCenteredPixelDefault(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	img := ax.ImShow([][]float64{{0, 1}, {2, 3}}, ImShowOptions{
		Extent: &[4]float64{-2, 2, -1, 1},
	})
	if img == nil {
		t.Fatal("ImShow returned nil")
	}
	if img.XMin != -2 || img.XMax != 2 || img.YMin != -1 || img.YMax != 1 {
		t.Fatalf("extent = [%v,%v]x[%v,%v], want [-2,2]x[-1,1]",
			img.XMin, img.XMax, img.YMin, img.YMax)
	}
}

func TestImShow_ExtentDrivesAxesLimits(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	_ = ax.ImShow([][]float64{{0, 1}}, ImShowOptions{
		Extent: &[4]float64{10, 20, 30, 40},
	})
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	// Origin is upper by default → ImShow inverts Y, so domain comes back swapped.
	if xMin != 10 || xMax != 20 {
		t.Fatalf("x domain = [%v,%v], want [10,20]", xMin, xMax)
	}
	if !(yMin == 30 && yMax == 40) && !(yMin == 40 && yMax == 30) {
		t.Fatalf("y domain = [%v,%v], want {30,40}", yMin, yMax)
	}
}

func TestImShow_ExplicitExtentOriginUpperDoesNotInvertYLimits(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	img := ax.ImShow([][]float64{{0, 1}, {2, 3}}, ImShowOptions{
		Extent: &[4]float64{10, 20, 30, 40},
		Origin: ImageOriginUpper,
	})
	if img == nil {
		t.Fatal("ImShow returned nil")
	}

	yMin, yMax := ax.YScale.Domain()
	if yMin != 30 || yMax != 40 {
		t.Fatalf("explicit extent with origin upper y domain = [%v,%v], want Matplotlib [30,40]", yMin, yMax)
	}
	if ax.YInverted() {
		t.Fatal("explicit extent with origin upper should not invert the y-axis")
	}
}

func TestImShow_DefaultInterpolationUsesMatplotlibAntialiased(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	img := ax.ImShow([][]float64{{0, 1}, {2, 3}})
	if img == nil {
		t.Fatal("ImShow returned nil")
	}
	if img.Interpolation != "antialiased" {
		t.Fatalf("default ImShow interpolation = %q, want matplotlib rc default antialiased", img.Interpolation)
	}
}

func TestImShow_InterpolationPropagatesToImage(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	bilinear := "bilinear"
	img := ax.ImShow([][]float64{{0, 1}}, ImShowOptions{Interpolation: &bilinear})
	if img.Interpolation != "bilinear" {
		t.Fatalf("Interpolation = %q, want %q", img.Interpolation, "bilinear")
	}
}
