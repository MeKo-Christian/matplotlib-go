package core

import (
	"errors"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/optarg"
)

// Every plotting entry point takes its options as a variadic tail so callers can
// omit them. Passing two or more option values silently discarded everything
// after the first; Phase 2.3 rejects it instead. Entry points that already
// return an error report it there, and the rest panic, because an extra option
// value can only come from a literal at the call site.

func TestErrorReturningEntryPointsRejectExtraOptions(t *testing.T) {
	x := []float64{0, 1, 2}
	y := []float64{0, 1, 2}

	tests := []struct {
		name string
		call func(*Axes) (any, error)
	}{
		{"Plot", func(a *Axes) (any, error) { return a.Plot(x, y, PlotOptions{}, PlotOptions{}) }},
		{"Scatter", func(a *Axes) (any, error) { return a.Scatter(x, y, ScatterOptions{}, ScatterOptions{}) }},
		{"Bar", func(a *Axes) (any, error) { return a.Bar(x, y, BarOptions{}, BarOptions{}) }},
		{"BarH", func(a *Axes) (any, error) { return a.BarH(x, y, BarOptions{}, BarOptions{}) }},
		{"FillBetween", func(a *Axes) (any, error) {
			return a.FillBetween(x, y, y, FillOptions{}, FillOptions{})
		}},
		{"FillBetweenX", func(a *Axes) (any, error) {
			return a.FillBetweenX(x, y, y, FillOptions{}, FillOptions{})
		}},
		{"FillBetweenPlot", func(a *Axes) (any, error) {
			return a.FillBetweenPlot(x, y, y, FillOptions{}, FillOptions{})
		}},
		{"Hist", func(a *Axes) (any, error) { return a.Hist(x, HistOptions{}, HistOptions{}) }},
		{"ErrorBar", func(a *Axes) (any, error) {
			return a.ErrorBar(x, y, nil, nil, ErrorBarOptions{}, ErrorBarOptions{})
		}},
		{"ErrorBarContainer", func(a *Axes) (any, error) {
			return a.ErrorBarContainer(x, y, nil, nil, ErrorBarOptions{}, ErrorBarOptions{})
		}},
		{"ImShowRGB", func(a *Axes) (any, error) {
			data := [][][]float64{{{0, 0, 0}, {1, 1, 1}}, {{1, 0, 0}, {0, 1, 0}}}
			return a.ImShowRGB(data, ImShowRGBOptions{}, ImShowRGBOptions{})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, restore := captureWarnings()
			defer restore()

			ax := newAlphaTestAxes()
			before := len(ax.Artists)

			got, err := tt.call(ax)
			if err == nil {
				t.Fatalf("%s with two option values returned no error", tt.name)
			}
			var tooMany *optarg.TooManyError
			if !errors.As(err, &tooMany) {
				t.Fatalf("%s error = %v, want *optarg.TooManyError", tt.name, err)
			}
			if tooMany.Count != 2 {
				t.Fatalf("%s reported Count = %d, want 2", tt.name, tooMany.Count)
			}
			if !strings.Contains(err.Error(), "at most one") {
				t.Fatalf("%s error = %q, want it to mention the limit", tt.name, err)
			}
			if got != nil && !isNilArtist(got) {
				t.Fatalf("%s returned artist %v, want nil", tt.name, got)
			}
			if len(ax.Artists) != before {
				t.Fatalf("%s added %d artists to a rejected call", tt.name, len(ax.Artists)-before)
			}
			if len(*warnings) != 0 {
				t.Fatalf("%s should return an error, not warn: %v", tt.name, *warnings)
			}
		})
	}
}

// TestPlainEntryPointsPanicOnExtraOptions covers the entry points that still
// take their options as a variadic tail. Axes.ImShow, Axes.Stem, Axes.Annotate,
// Axes.HLines, and Axes.VLines are absent on purpose: the Phase 2.3 options
// model made them take exactly one option value, so a second one no longer
// compiles and there is nothing left to reject at run time.
func TestPlainEntryPointsPanicOnExtraOptions(t *testing.T) {
	x := []float64{0, 1, 2}
	y := []float64{0, 1, 2}
	grid := [][]float64{{0, 1}, {1, 2}}

	tests := []struct {
		name string
		call func(*Axes)
	}{
		{"Image", func(a *Axes) { a.Image(grid, ImageOptions{}, ImageOptions{}) }},
		{"MatShow", func(a *Axes) { a.MatShow(grid, MatShowOptions{}, MatShowOptions{}) }},
		{"PColor", func(a *Axes) { a.PColor(grid, MeshOptions{}, MeshOptions{}) }},
		{"PColorMesh", func(a *Axes) { a.PColorMesh(grid, MeshOptions{}, MeshOptions{}) }},
		{"Contour", func(a *Axes) { a.Contour(grid, ContourOptions{}, ContourOptions{}) }},
		{"Contourf", func(a *Axes) { a.Contourf(grid, ContourOptions{}, ContourOptions{}) }},
		{"SemilogX", func(a *Axes) { a.SemilogX(x, y, PlotOptions{}, PlotOptions{}) }},
		{"SemilogY", func(a *Axes) { a.SemilogY(x, y, PlotOptions{}, PlotOptions{}) }},
		{"LogLog", func(a *Axes) { a.LogLog(x, y, PlotOptions{}, PlotOptions{}) }},
		{"Step", func(a *Axes) { a.Step(x, y, StepOptions{}, StepOptions{}) }},
		{"Fill", func(a *Axes) { a.Fill(x, y, FillOptions{}, FillOptions{}) }},
		{"Text", func(a *Axes) { a.Text(0, 0, "t", TextOptions{}, TextOptions{}) }},
		{"AxHLine", func(a *Axes) { a.AxHLine(0, HLineOptions{}, HLineOptions{}) }},
		{"AxVSpan", func(a *Axes) { a.AxVSpan(0, 1, VSpanOptions{}, VSpanOptions{}) }},
		{"Pie", func(a *Axes) { a.Pie(x, PieOptions{}, PieOptions{}) }},
		{"BoxPlot", func(a *Axes) { a.BoxPlot(x, BoxPlotOptions{}, BoxPlotOptions{}) }},
		{"Hexbin", func(a *Axes) { a.Hexbin(x, y, HexbinOptions{}, HexbinOptions{}) }},
		{"Quiver", func(a *Axes) { a.Quiver(x, y, x, y, QuiverOptions{}, QuiverOptions{}) }},
		{"Barbs", func(a *Axes) { a.Barbs(x, y, x, y, BarbsOptions{}, BarbsOptions{}) }},
		{"ECDF", func(a *Axes) { a.ECDF(x, ECDFOptions{}, ECDFOptions{}) }},
		{"StackPlot", func(a *Axes) { a.StackPlot(x, [][]float64{y}, StackPlotOptions{}, StackPlotOptions{}) }},
		{"Table", func(a *Axes) { a.Table(TableOptions{}, TableOptions{}) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ax := newAlphaTestAxes()
			before := len(ax.Artists)

			defer func() {
				recovered := recover()
				if recovered == nil {
					t.Fatalf("%s with two option values did not panic", tt.name)
				}
				err, ok := recovered.(error)
				if !ok {
					t.Fatalf("%s panicked with %#v, want an error", tt.name, recovered)
				}
				var tooMany *optarg.TooManyError
				if !errors.As(err, &tooMany) {
					t.Fatalf("%s panicked with %v, want *optarg.TooManyError", tt.name, err)
				}
				if len(ax.Artists) != before {
					t.Fatalf("%s added %d artists before rejecting", tt.name, len(ax.Artists)-before)
				}
			}()
			tt.call(ax)
		})
	}
}

// TestSingleOptionValueIsStillAccepted guards against the rejection firing one
// value too early.
func TestSingleOptionValueIsStillAccepted(t *testing.T) {
	x := []float64{0, 1, 2}
	y := []float64{0, 1, 2}

	ax := newAlphaTestAxes()
	if line, err := ax.Plot(x, y, PlotOptions{}); err != nil || line == nil {
		t.Fatalf("Plot() = (%v, %v), want an artist and no error", line, err)
	}
	if line := ax.SemilogX(x, y, PlotOptions{}); line == nil {
		t.Fatal("SemilogX() = nil, want an artist")
	}
	if img := ax.ImShow([][]float64{{0, 1}, {1, 2}}, ImShowOptions{}); img == nil {
		t.Fatal("ImShow() = nil, want an artist")
	}
	if txt := ax.Text(0, 0, "t", TextOptions{}); txt == nil {
		t.Fatal("Text() = nil, want an artist")
	}
}

// isNilArtist reports whether an interface holds a typed nil pointer, which is
// how the error-returning entry points spell "no artist".
func isNilArtist(v any) bool {
	switch a := v.(type) {
	case *Line2D:
		return a == nil
	case *Scatter2D:
		return a == nil
	case *Bar2D:
		return a == nil
	case *Fill2D:
		return a == nil
	case *Hist2D:
		return a == nil
	case *ErrorBar:
		return a == nil
	case *ErrorbarContainer:
		return a == nil
	case *Image2D:
		return a == nil
	default:
		return false
	}
}
