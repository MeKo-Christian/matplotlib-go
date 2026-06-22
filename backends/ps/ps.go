package ps

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/backends/internal/mixedraster"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const defaultFontHeight = 13.0

type state struct {
	inContent bool
	clipRect  *geom.Rect
	clipPaths []geom.Path
}

// Renderer implements render.Renderer by emitting deterministic Level-2
// PostScript content.
type Renderer struct {
	width      int
	height     int
	background render.Color
	resolution uint

	began         bool
	viewport      geom.Rect
	content       strings.Builder
	document      []byte
	stack         []state
	clipRect      *geom.Rect
	clipPaths     []geom.Path
	raster        *mixedraster.Session
	markerIDs     map[string]string
	imageIDs      map[string]string
	psOpts        render.PSOptions
	lastFontKey   string
	defaultSketch render.SketchParams
}

// SetDefaultSketch sets the sketch/xkcd perturbation applied to paths whose
// paint does not carry its own. Implements render.SketchAware.
func (r *Renderer) SetDefaultSketch(params render.SketchParams) { r.defaultSketch = params }

var (
	_ render.Renderer                = (*Renderer)(nil)
	_ render.PSExporter              = (*Renderer)(nil)
	_ render.DPIAware                = (*Renderer)(nil)
	_ render.ImageTransformer        = (*Renderer)(nil)
	_ render.MarkerDrawer            = (*Renderer)(nil)
	_ render.NativeHatcher           = (*Renderer)(nil)
	_ render.GradientFiller          = (*Renderer)(nil)
	_ render.PatternFiller           = (*Renderer)(nil)
	_ render.PathCollectionDrawer    = (*Renderer)(nil)
	_ render.RasterizationController = (*Renderer)(nil)
	_ render.TextPather              = (*Renderer)(nil)
	_ render.FontTextDrawer          = (*Renderer)(nil)
	_ render.FontRotatedTextDrawer   = (*Renderer)(nil)
	_ render.FontVerticalTextDrawer  = (*Renderer)(nil)
	_ render.PSOptionSetter          = (*Renderer)(nil)
)
