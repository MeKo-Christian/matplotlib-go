package pyplot

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

func TestFigureRegistryTracksCurrentFigureAndAxes(t *testing.T) {
	resetForTests()

	fig1 := Figure()
	if got := GCF(); got != fig1 {
		t.Fatalf("GCF() = %p, want %p", got, fig1)
	}

	ax1 := GCA()
	if ax1 == nil {
		t.Fatal("GCA() returned nil")
	}
	if len(fig1.Children) != 1 {
		t.Fatalf("len(fig1.Children) = %d, want 1", len(fig1.Children))
	}

	fig2 := FigureSized(900, 700)
	if got := GCF(); got != fig2 {
		t.Fatalf("after FigureSized, GCF() = %p, want %p", got, fig2)
	}
	if fig2.SizePx.X != 900 || fig2.SizePx.Y != 700 {
		t.Fatalf("FigureSized dimensions = %.0fx%.0f, want 900x700", fig2.SizePx.X, fig2.SizePx.Y)
	}
}

func TestSubplotReusesAxesForSameSlot(t *testing.T) {
	resetForTests()

	fig := Figure()
	ax1 := Subplot(2, 2, 3)
	ax2 := Subplot(2, 2, 3)
	if ax1 == nil || ax2 == nil {
		t.Fatal("Subplot returned nil axes")
	}
	if ax1 != ax2 {
		t.Fatalf("Subplot did not reuse axes: %p != %p", ax1, ax2)
	}
	if got := len(fig.Children); got != 1 {
		t.Fatalf("len(fig.Children) = %d, want 1", got)
	}
	if got := GCA(); got != ax1 {
		t.Fatalf("GCA() = %p, want %p", got, ax1)
	}
}

func TestSubplotsCreatesNewFigureAndCurrentAxes(t *testing.T) {
	resetForTests()

	fig, grid := Subplots(2, 2, core.WithSubplotShareX())
	if fig == nil {
		t.Fatal("Subplots returned nil figure")
	}
	if len(grid) != 2 || len(grid[0]) != 2 {
		t.Fatalf("Subplots grid dimensions = %dx%d, want 2x2", len(grid), len(grid[0]))
	}
	if got := GCF(); got != fig {
		t.Fatalf("GCF() = %p, want %p", got, fig)
	}
	if got := GCA(); got != grid[0][0] {
		t.Fatalf("GCA() = %p, want %p", got, grid[0][0])
	}

	left := 0.2
	right := 0.85
	bottom := 0.15
	top := 0.88
	SubplotsAdjust(core.SubplotAdjust{
		Left:   &left,
		Right:  &right,
		Bottom: &bottom,
		Top:    &top,
	})
	if !approxFloat(grid[0][0].RectFraction.Min.X, left, 1e-12) || !approxFloat(grid[0][0].RectFraction.Max.Y, top, 1e-12) {
		t.Fatalf("top-left subplot rect after adjust = %+v", grid[0][0].RectFraction)
	}
	if !approxFloat(grid[1][1].RectFraction.Max.X, right, 1e-12) || !approxFloat(grid[1][1].RectFraction.Min.Y, bottom, 1e-12) {
		t.Fatalf("bottom-right subplot rect after adjust = %+v", grid[1][1].RectFraction)
	}
}

func TestSubplot2GridAddsCurrentSpanningAxes(t *testing.T) {
	resetForTests()

	fig := Figure()
	ax := Subplot2Grid([2]int{3, 3}, [2]int{1, 1}, 2, 2)
	if ax == nil {
		t.Fatal("Subplot2Grid() returned nil")
	}
	if got := GCA(); got != ax {
		t.Fatalf("GCA() = %p, want %p", got, ax)
	}
	if got := GCF(); got != fig {
		t.Fatalf("GCF() = %p, want %p", got, fig)
	}
	if len(fig.Children) != 1 || fig.Children[0] != ax {
		t.Fatalf("figure children = %+v, want subplot2grid axes", fig.Children)
	}
	if ax.RectFraction.Min.X <= 0.1 || ax.RectFraction.Max.X <= 0.8 {
		t.Fatalf("subplot2grid rect = %+v, want spanning right-side axes", ax.RectFraction)
	}
}

func TestSubplotMosaicAddsNamedAxesAndSetsCurrent(t *testing.T) {
	resetForTests()

	fig := Figure()
	axes, err := SubplotMosaic([][]string{
		{"A", "A", "B"},
		{"C", ".", "B"},
	})
	if err != nil {
		t.Fatalf("SubplotMosaic() error = %v", err)
	}
	if len(axes) != 3 {
		t.Fatalf("mosaic axes count = %d, want 3", len(axes))
	}
	if axes["A"] == nil || axes["B"] == nil || axes["C"] == nil {
		t.Fatalf("mosaic axes = %+v, want A/B/C", axes)
	}
	if got := GCF(); got != fig {
		t.Fatalf("GCF() = %p, want %p", got, fig)
	}
	if got := GCA(); got != axes["A"] {
		t.Fatalf("GCA() = %p, want first mosaic axes %p", got, axes["A"])
	}
	if len(fig.Children) != 3 {
		t.Fatalf("figure children = %d, want 3", len(fig.Children))
	}
	if _, err := SubplotMosaic([][]string{{"A", "B"}, {"C", "A"}}); err == nil {
		t.Fatal("SubplotMosaic() with non-rectangular region returned nil error")
	}
}

func TestAxesAddsCurrentAxesToCurrentFigure(t *testing.T) {
	resetForTests()

	fig := Figure()
	rect := geom.Rect{Min: geom.Pt{X: 0.2, Y: 0.3}, Max: geom.Pt{X: 0.8, Y: 0.9}}
	ax := Axes(rect)
	if ax == nil {
		t.Fatal("Axes() returned nil")
	}
	if got := GCA(); got != ax {
		t.Fatalf("GCA() = %p, want %p", got, ax)
	}
	if got := GCF(); got != fig {
		t.Fatalf("GCF() = %p, want %p", got, fig)
	}
	if len(fig.Children) != 1 || fig.Children[0] != ax {
		t.Fatalf("figure children = %+v, want added axes", fig.Children)
	}
	if ax.RectFraction != rect {
		t.Fatalf("axes rect = %+v, want %+v", ax.RectFraction, rect)
	}
}

func TestSCAAndDelAxesUpdateCurrentAxesRegistry(t *testing.T) {
	resetForTests()

	fig, grid := Subplots(1, 2)
	left := grid[0][0]
	right := grid[0][1]

	if err := SCA(right); err != nil {
		t.Fatalf("SCA(right) error = %v", err)
	}
	if got := GCA(); got != right {
		t.Fatalf("GCA() after SCA = %p, want right %p", got, right)
	}
	if got := GCF(); got != fig {
		t.Fatalf("GCF() after SCA = %p, want fig %p", got, fig)
	}

	if err := DelAxes(right); err != nil {
		t.Fatalf("DelAxes(right) error = %v", err)
	}
	if len(fig.Children) != 1 || fig.Children[0] != left {
		t.Fatalf("figure children after DelAxes = %+v, want only left", fig.Children)
	}
	if got := GCA(); got != left {
		t.Fatalf("GCA() after deleting current axes = %p, want left %p", got, left)
	}

	otherFig := core.NewFigure(100, 100)
	foreign := otherFig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	if err := SCA(foreign); err == nil {
		t.Fatal("SCA(foreign) returned nil error")
	}
	if err := DelAxes(foreign); err == nil {
		t.Fatal("DelAxes(foreign) returned nil error")
	}
}

func TestFigureLayoutAndTwinAxesWrappersDelegateToCurrentState(t *testing.T) {
	resetForTests()

	fig := GCF()
	if text := FigText(0.1, 0.9, "figure note", core.TextOptions{}); text == nil {
		t.Fatal("FigText() returned nil")
	}
	if got := len(fig.Artists); got != 1 {
		t.Fatalf("len(fig.Artists) = %d, want 1", got)
	}

	TightLayout()
	if got := fig.LayoutEngine(); got != core.LayoutEngineTight {
		t.Fatalf("layout engine = %v, want LayoutEngineTight", got)
	}

	base := GCA()
	xTwin := TwinX()
	if xTwin == nil {
		t.Fatal("TwinX() returned nil")
	}
	if got := GCA(); got != xTwin {
		t.Fatalf("after TwinX, GCA() = %p, want %p", got, xTwin)
	}
	if xTwin.YAxisRight == nil || xTwin.YAxisRight.Side != core.AxisRight {
		t.Fatalf("TwinX right axis = %#v, want right", xTwin.YAxisRight)
	}

	SCA(base)
	yTwin := TwinY()
	if yTwin == nil {
		t.Fatal("TwinY() returned nil")
	}
	if got := GCA(); got != yTwin {
		t.Fatalf("after TwinY, GCA() = %p, want %p", got, yTwin)
	}
	if yTwin.XAxisTop == nil || yTwin.XAxisTop.Side != core.AxisTop {
		t.Fatalf("TwinY top axis = %#v, want top", yTwin.XAxisTop)
	}
	if got := len(fig.Children); got != 3 {
		t.Fatalf("len(fig.Children) = %d, want base axes plus two twins", got)
	}
}

func TestColorbarUsesCurrentAxesAndFigure(t *testing.T) {
	resetForTests()

	img := Image([][]float64{
		{0, 1},
		{2, 3},
	}, core.ImageOptions{})
	cb := Colorbar(img, core.ColorbarOptions{Label: "Intensity"})
	if cb == nil {
		t.Fatal("Colorbar() returned nil")
	}

	fig := GCF()
	if len(fig.Children) != 2 {
		t.Fatalf("len(fig.Children) = %d, want 2", len(fig.Children))
	}
	if cb.YLabel() != "Intensity" {
		t.Fatalf("colorbar label = %q, want %q", cb.YLabel(), "Intensity")
	}
}
