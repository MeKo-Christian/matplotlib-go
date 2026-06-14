package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestAddAxesUsesThemeDefaults(t *testing.T) {
	fig := NewFigure(400, 300, style.WithTheme(style.ThemePublication))
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	if ax.XAxis.Color != fig.RC.AxesEdgeColor {
		t.Fatalf("x axis color = %+v, want %+v", ax.XAxis.Color, fig.RC.AxesEdgeColor)
	}
	if ax.YAxis.LineWidth != fig.RC.AxisLineWidth {
		t.Fatalf("y axis width = %v, want %v", ax.YAxis.LineWidth, fig.RC.AxisLineWidth)
	}
	if got, want := ax.PeekColor(), fig.RC.Palette()[0]; got != want {
		t.Fatalf("palette head = %+v, want %+v", got, want)
	}
}

func TestAddGridAndLegendUseThemeDefaults(t *testing.T) {
	fig := NewFigure(400, 300, style.WithTheme(style.ThemeGGPlot))
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	grid := ax.AddXGrid()
	if grid.Color != fig.RC.GridColor {
		t.Fatalf("grid color = %+v, want %+v", grid.Color, fig.RC.GridColor)
	}
	if grid.MinorColor != fig.RC.MinorGridColor {
		t.Fatalf("minor grid color = %+v, want %+v", grid.MinorColor, fig.RC.MinorGridColor)
	}
	if grid.LineWidth != fig.RC.GridLineWidth {
		t.Fatalf("grid width = %v, want %v", grid.LineWidth, fig.RC.GridLineWidth)
	}

	legend := ax.AddLegend()
	if legend.BackgroundColor != fig.RC.LegendBackground {
		t.Fatalf("legend background = %+v, want %+v", legend.BackgroundColor, fig.RC.LegendBackground)
	}
	if legend.BorderColor != fig.RC.LegendBorderColor {
		t.Fatalf("legend border = %+v, want %+v", legend.BorderColor, fig.RC.LegendBorderColor)
	}
	if legend.TextColor != fig.RC.LegendTextColor {
		t.Fatalf("legend text color = %+v, want %+v", legend.TextColor, fig.RC.LegendTextColor)
	}
}

func TestNewFigureUsesRuntimeDefaults(t *testing.T) {
	style.ResetDefaults()
	t.Cleanup(style.ResetDefaults)

	if _, err := style.UpdateParams(style.Params{
		"figure.dpi":     "144",
		"axes.edgecolor": "#224466",
	}); err != nil {
		t.Fatalf("UpdateParams() error = %v", err)
	}

	fig := NewFigure(400, 300)
	if fig.RC.DPI != 144 {
		t.Fatalf("figure DPI = %v, want 144", fig.RC.DPI)
	}

	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	if got := ax.XAxis.Color; got.R != 0x22/255.0 || got.G != 0x44/255.0 || got.B != 0x66/255.0 {
		t.Fatalf("axis edge color = %+v", got)
	}
}

func TestAddAxesAppliesDefaultGridConfiguration(t *testing.T) {
	fig := NewFigure(400, 300)
	fig.RC.GridVisible = true
	fig.RC.GridAxis = "y"
	fig.RC.GridWhich = "both"
	fig.RC.GridDashes = []float64{6, 6}
	fig.RC.MinorGridDashes = []float64{1.2, 2.4}

	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	if len(ax.Artists) != 1 {
		t.Fatalf("expected one default grid artist, got %d", len(ax.Artists))
	}
	grid, ok := ax.Artists[0].(*Grid)
	if !ok {
		t.Fatalf("expected first artist to be Grid, got %T", ax.Artists[0])
	}
	if grid.Axis != AxisLeft || !grid.Major || !grid.Minor {
		t.Fatalf("unexpected grid configuration: axis=%v major=%v minor=%v", grid.Axis, grid.Major, grid.Minor)
	}
	if len(grid.Dashes) != 2 || len(grid.MinorDashes) != 2 {
		t.Fatalf("expected default grid dash styles, got major=%v minor=%v", grid.Dashes, grid.MinorDashes)
	}
}
