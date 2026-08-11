package pgf

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/backends/internal/mixedraster"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const defaultFontHeight = 13.0

// colorDefsPlaceholder marks where collected \definecolor declarations are
// injected at the pgfpicture's outermost group. It is always substituted in End.
const colorDefsPlaceholder = "%mplgpgf-color-defs\n"

type state struct {
	inContent bool
	clipRect  *geom.Rect
	clipPaths []geom.Path
}

// Renderer implements render.Renderer by emitting PGF/TikZ commands.
type Renderer struct {
	width      float64
	height     float64
	background render.Color
	resolution uint

	began          bool
	viewport       geom.Rect
	content        strings.Builder
	colorDefs      strings.Builder
	document       []byte
	stack          []state
	clipRect       *geom.Rect
	clipPaths      []geom.Path
	raster         *mixedraster.Session
	colorNames     map[string]string
	pathIDs        map[string]string
	shadingCounter int
	pgfOpts        render.PGFOptions
	defaultSketch  render.SketchParams
}

// SetDefaultSketch sets the sketch/xkcd perturbation applied to paths whose
// paint does not carry its own. Implements render.SketchAware.
func (r *Renderer) SetDefaultSketch(params render.SketchParams) { r.defaultSketch = params }

var (
	_ render.Renderer                = (*Renderer)(nil)
	_ render.PGFExporter             = (*Renderer)(nil)
	_ render.DPIAware                = (*Renderer)(nil)
	_ render.ImageTransformer        = (*Renderer)(nil)
	_ render.MarkerDrawer            = (*Renderer)(nil)
	_ render.PathCollectionDrawer    = (*Renderer)(nil)
	_ render.FontTextDrawer          = (*Renderer)(nil)
	_ render.FontRotatedTextDrawer   = (*Renderer)(nil)
	_ render.NativeHatcher           = (*Renderer)(nil)
	_ render.GradientFiller          = (*Renderer)(nil)
	_ render.PatternFiller           = (*Renderer)(nil)
	_ render.ClipPathTransformer     = (*Renderer)(nil)
	_ render.FontVerticalTextDrawer  = (*Renderer)(nil)
	_ render.RasterizationController = (*Renderer)(nil)
	_ render.PGFOptionExporter       = (*Renderer)(nil)
	_ render.PGFOptionSetter         = (*Renderer)(nil)
)
