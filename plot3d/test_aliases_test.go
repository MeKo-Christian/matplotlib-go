package plot3d

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

type (
	Artist               = core.Artist
	Colorbar             = core.Colorbar
	ContourOptions       = core.ContourOptions
	DrawContext          = core.DrawContext
	Line2D               = core.Line2D
	LineCollection       = core.LineCollection
	LogNorm              = core.LogNorm
	MarkerCollection     = core.PathCollection
	Normalize            = core.Normalize
	PathCollection       = core.PathCollection
	PlotOptions          = core.PlotOptions
	ColorbarOptions      = core.ColorbarOptions
	TextOptions          = core.TextOptions
	PolyCollection       = core.PolyCollection
	Pt                   = geom.Pt
	ScalarMapInfo        = core.ScalarMapInfo
	ScalarMappable       = core.ScalarMappable
	Scatter2D            = core.Scatter2D
	ScatterOptions       = core.ScatterOptions
	TextAlign            = core.TextAlign
	Triangulation        = core.Triangulation
	singleLineTextLayout = render.TextLineLayout
)

const (
	CoordData           = core.CoordData
	TextAlignCenter     = core.TextAlignCenter
	textLayoutVAlignTop = iota
	defaultTickPadPt    = 3.5
)

var (
	Coords     = core.Coords
	DrawFigure = core.DrawFigure
	NewFigure  = core.NewFigure
)

func unitRect() geom.Rect {
	return geom.Rect{Max: geom.Pt{X: 1, Y: 1}}
}

func approx(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func (a *Axes3D) adjustedLayout(*core.Figure) geom.Rect {
	if a == nil || a.Axes == nil {
		return geom.Rect{}
	}
	return a.DisplayRect()
}

func (a *Axes3D) resolvedRC() style.RC {
	if a == nil || a.Axes == nil {
		return style.Default
	}
	return a.ResolvedRC()
}

func (a *Axes3D) effectiveXScale() transform.Scale {
	if a == nil || a.Axes == nil {
		return nil
	}
	return a.XScale
}

func (a *Axes3D) effectiveYScale() transform.Scale {
	if a == nil || a.Axes == nil {
		return nil
	}
	return a.YScale
}

func newAxesDrawContext(ax *core.Axes, fig *core.Figure, _, _ geom.Rect) *core.DrawContext {
	return core.AxesDrawContext(ax, fig)
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func sortedArtistDrawOrder(artists []Artist) []Artist {
	if len(artists) < 2 {
		return artists
	}
	out := append([]Artist(nil), artists...)
	sort.SliceStable(out, func(i, j int) bool {
		zi, zj := out[i].Z(), out[j].Z()
		if zi == zj {
			return i < j
		}
		return zi < zj
	})
	return out
}

func captureWarnings() (*[]string, func()) {
	var warnings []string
	restore := diag.SetHandler(func(message string) {
		warnings = append(warnings, message)
	})
	return &warnings, restore
}

func measureSingleLineTextLayout(r render.Renderer, text string, size float64, fontKey string, _ ...bool) singleLineTextLayout {
	return render.MeasureTextLineLayout(r, text, size, fontKey)
}

//nolint:gocritic // Compatibility shim mirrors the former core test helper's value signature.
func textHorizontalOriginOffset(layout singleLineTextLayout, align TextAlign) float64 {
	switch align {
	case core.TextAlignLeft:
		return 0
	case core.TextAlignRight:
		return layout.Width
	default:
		return layout.Width / 2
	}
}

//nolint:gocritic // Compatibility shim mirrors the former core test helper's value signature.
func textBaselineOffset(layout singleLineTextLayout, align int) float64 {
	if align == textLayoutVAlignTop {
		return -layout.Ascent
	}
	return 0
}

//nolint:gocritic // Compatibility shim mirrors the former core test helper's value signature.
func alignedSingleLineOrigin(anchor geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign int) geom.Pt {
	return geom.Pt{
		X: anchor.X - textHorizontalOriginOffset(layout, hAlign),
		Y: anchor.Y + textBaselineOffset(layout, vAlign),
	}
}

func resolveScalarMapValues(values []float64, cmap string, vmin, vmax *float64) core.ScalarMapInfo {
	mapping, err := core.ResolveScalarMapValues(values, core.ScalarMapConfig{
		Colormap: cmap,
		VMin:     optional.FromPtr(vmin),
		VMax:     optional.FromPtr(vmax),
	})
	if err != nil {
		return core.ScalarMapInfo{Colormap: cmap}.Resolved()
	}
	return mapping
}

//nolint:gocritic // Compatibility shim preserves the moved contour tests' value-shaped call sites.
func contourGridBandPolygons(
	x, y []float64,
	data [][]float64,
	levels []float64,
	_ core.ContourOptions,
	_ core.ScalarMapInfo,
	_ float64,
) ([][]geom.Pt, []render.Color, []string) {
	var polygons [][]geom.Pt
	for i := 0; i+1 < len(levels); i++ {
		polygons = append(polygons, contourGridBandPolygonsForLevel(x, y, data, levels[i], levels[i+1])...)
	}
	return polygons, nil, nil
}

type colorbarRecordingRenderer struct {
	render.NullRenderer
}

func (*colorbarRecordingRenderer) DrawImage(render.Image, geom.Rect) {}

func (*colorbarRecordingRenderer) Path(geom.Path, *render.Paint) {}

func (*colorbarRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (*colorbarRecordingRenderer) DrawText(string, geom.Pt, float64, render.Color) {}

type batchRecordingRenderer struct {
	render.NullRenderer
	pathCalls []geom.Path
}

func (r *batchRecordingRenderer) Path(path geom.Path, _ *render.Paint) {
	r.pathCalls = append(r.pathCalls, path)
}
