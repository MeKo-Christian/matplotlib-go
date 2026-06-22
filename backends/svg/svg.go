package svg

import (
	"errors"

	"github.com/cwbudde/matplotlib-go/backends/internal/mixedraster"
	"github.com/cwbudde/matplotlib-go/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	quantizationGrid  = 1e-6
	defaultFontHeight = 13.0
)

type state struct {
	clipRect      *geom.Rect
	clipPathDepth int
}

type svgNode struct {
	content string
	// clipIDs lists active clip-path defs in outer-to-inner order. An empty
	// slice means the node is unclipped. The serializer wraps the content in
	// nested <g clip-path="url(#…)"> groups, opening outer-most first.
	clipIDs []string
	// filterIDs lists active SVG filter defs in outer-to-inner order.
	filterIDs []string
}

type clipDef struct {
	id        string
	rect      *geom.Rect
	path      string
	transform string
}

type markerDef struct {
	id   string
	data string
}

type pathCollectionDef struct {
	id   string
	data string
}

type hatchDef struct {
	id        string
	hatch     string
	faceColor render.Color
	lineColor render.Color
	lineWidth float64
	spacing   float64
	forced    bool
}

type fontFaceDef struct {
	family string
	data   string
	mime   string
	format string
}

type filterDef struct {
	id     string
	name   string
	radius float64
}

type Renderer struct {
	width      int
	height     int
	viewport   geom.Rect
	began      bool
	stack      []state
	clipRect   *geom.Rect
	resolution uint
	background render.Color

	nodes           []svgNode
	clipDefs        map[string]string // rect-key → clipDef ID
	clipPathDefs    map[string]string // path data → clipDef ID
	clipOrder       []clipDef         // registration-order defs (rects and paths interleaved)
	clipPathStack   []string          // active path-clip IDs in outer-to-inner order
	clipPaths       []geom.Path       // active path clips in display coordinates for mixed raster replay
	clipIDCounter   int
	markerDefs      map[string]string // marker path data → markerDef ID
	markerOrder     []markerDef       // registration-order marker defs
	markerIDCounter int
	pathDefs        map[string]string // collection path data → pathCollectionDef ID
	pathOrder       []pathCollectionDef
	pathIDCounter   int
	hatchDefs       map[string]string // hatch paint metadata → hatchDef ID
	hatchOrder      []hatchDef
	hatchIDCounter  int

	gradientDefs         map[string]string // gradient key → gradientDef ID
	gradientOrder        []gradientDef     // registration-order gradient defs
	gradientIDCounter    int
	patternFillDefs      map[string]string // pattern key → patternFillDef ID
	patternFillOrder     []patternFillDef  // registration-order pattern defs
	patternFillIDCounter int

	fontFaces     map[string]fontFaceDef
	fontFaceOrder []fontFaceDef

	filterDefs      map[string]string // filter metadata → filterDef ID
	filterOrder     []filterDef
	filterIDCounter int
	filterStack     []string

	lastFontKey   string
	texManager    *tex.Manager
	texErr        error
	options       render.SVGOptions
	raster        *mixedraster.Session
	defaultSketch render.SketchParams
}

// SetDefaultSketch sets the sketch/xkcd perturbation applied to paths whose
// paint does not carry its own. Implements render.SketchAware.
func (r *Renderer) SetDefaultSketch(params render.SketchParams) { r.defaultSketch = params }

var (
	_ render.Renderer                = (*Renderer)(nil)
	_ render.DPIAware                = (*Renderer)(nil)
	_ render.TextDrawer              = (*Renderer)(nil)
	_ render.RotatedTextDrawer       = (*Renderer)(nil)
	_ render.VerticalTextDrawer      = (*Renderer)(nil)
	_ render.TextPather              = (*Renderer)(nil)
	_ render.TeXMetricer             = (*Renderer)(nil)
	_ render.TeXDrawer               = (*Renderer)(nil)
	_ render.RotatedTeXDrawer        = (*Renderer)(nil)
	_ render.ImageTransformer        = (*Renderer)(nil)
	_ render.ClipPathTransformer     = (*Renderer)(nil)
	_ render.MarkerDrawer            = (*Renderer)(nil)
	_ render.PathCollectionDrawer    = (*Renderer)(nil)
	_ render.NativeHatcher           = (*Renderer)(nil)
	_ render.GradientFiller          = (*Renderer)(nil)
	_ render.PatternFiller           = (*Renderer)(nil)
	_ render.PathEffectFilterDrawer  = (*Renderer)(nil)
	_ render.RasterizationController = (*Renderer)(nil)
	_ render.SVGExporter             = (*Renderer)(nil)
)

func New(w, h int, bg render.Color) (*Renderer, error) {
	if w <= 0 || h <= 0 {
		return nil, errors.New("svg: width and height must be positive")
	}

	return &Renderer{
		width:           w,
		height:          h,
		background:      bg,
		resolution:      72,
		clipDefs:        map[string]string{},
		clipPathDefs:    map[string]string{},
		markerDefs:      map[string]string{},
		pathDefs:        map[string]string{},
		hatchDefs:       map[string]string{},
		gradientDefs:    map[string]string{},
		patternFillDefs: map[string]string{},
		fontFaces:       map[string]fontFaceDef{},
		filterDefs:      map[string]string{},
		texManager:      tex.NewManager(tex.ManagerConfig{}),
		options:         render.DefaultSVGOptions(),
	}, nil
}

func (r *Renderer) SetSVGOptions(opts render.SVGOptions) {
	if r == nil {
		return
	}
	r.options = normalizeSVGOptions(opts)
}

func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("Begin called twice")
	}

	r.began = true
	r.viewport = viewport
	r.nodes = nil
	r.stack = r.stack[:0]
	r.clipRect = nil
	r.clipDefs = map[string]string{}
	r.clipPathDefs = map[string]string{}
	r.clipOrder = nil
	r.clipPathStack = nil
	r.clipPaths = nil
	r.clipIDCounter = 0
	r.markerDefs = map[string]string{}
	r.markerOrder = nil
	r.markerIDCounter = 0
	r.pathDefs = map[string]string{}
	r.pathOrder = nil
	r.pathIDCounter = 0
	r.hatchDefs = map[string]string{}
	r.hatchOrder = nil
	r.hatchIDCounter = 0
	r.fontFaces = map[string]fontFaceDef{}
	r.fontFaceOrder = nil
	r.filterDefs = map[string]string{}
	r.filterOrder = nil
	r.filterIDCounter = 0
	r.filterStack = nil
	r.lastFontKey = ""
	return nil
}

func (r *Renderer) End() error {
	if !r.began {
		return errors.New("End called before Begin")
	}

	r.began = false
	r.stack = r.stack[:0]
	r.clipRect = nil
	return nil
}

func (r *Renderer) StartRasterized(options render.Rasterization) bool {
	if r == nil || !r.began || r.raster != nil {
		return false
	}
	session, ok := mixedraster.Start(r.width, r.height, r.viewport, options, r.resolution, r.clipRect, r.clipPaths)
	if !ok {
		return false
	}
	r.raster = session
	return true
}

func (r *Renderer) StopRasterized() bool {
	if r == nil || r.raster == nil {
		return false
	}
	session := r.raster
	r.raster = nil
	img, rect, ok := session.Stop()
	if !ok {
		return false
	}
	r.Image(img, rect)
	return true
}

func (r *Renderer) activeRaster() render.Renderer {
	if r == nil || r.raster == nil {
		return nil
	}
	return r.raster.Renderer()
}

func (r *Renderer) Save() {
	if rr := r.activeRaster(); rr != nil {
		rr.Save()
		return
	}
	var clipCopy *geom.Rect
	if r.clipRect != nil {
		copyRect := *r.clipRect
		clipCopy = &copyRect
	}
	r.stack = append(r.stack, state{
		clipRect:      clipCopy,
		clipPathDepth: len(r.clipPathStack),
	})
}

func (r *Renderer) Restore() {
	if rr := r.activeRaster(); rr != nil {
		rr.Restore()
		return
	}
	if len(r.stack) == 0 {
		return
	}

	s := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	r.clipRect = s.clipRect
	if s.clipPathDepth < len(r.clipPathStack) {
		r.clipPathStack = r.clipPathStack[:s.clipPathDepth]
	}
	if s.clipPathDepth < len(r.clipPaths) {
		r.clipPaths = r.clipPaths[:s.clipPathDepth]
	}
}

func (r *Renderer) SetResolution(dpi uint) {
	if dpi > 0 {
		r.resolution = dpi
	}
}
