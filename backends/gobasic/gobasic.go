package gobasic

import (
	"image"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/vector"
)

// state represents a saved graphics state.
type state struct {
	clipRect  *geom.Rect
	clipPaths []geom.Path
}

// Renderer implements render.Renderer using pure Go dependencies.
type Renderer struct {
	dst         *image.RGBA
	viewport    geom.Rect
	began       bool
	stack       []state
	clipRect    *geom.Rect
	clipPaths   []geom.Path
	clipMaskMap map[clipMaskKey]*image.Alpha
	rasterizer  *vector.Rasterizer
	lastFontKey string
	resolution  uint
}

type clipMaskKey struct {
	width  int
	height int
	hash   uint64
}

var (
	_ render.Renderer           = (*Renderer)(nil)
	_ render.DPIAware           = (*Renderer)(nil)
	_ render.ImageTransformer   = (*Renderer)(nil)
	_ render.TextDrawer         = (*Renderer)(nil)
	_ render.TextPather         = (*Renderer)(nil)
	_ render.RotatedTextDrawer  = (*Renderer)(nil)
	_ render.VerticalTextDrawer = (*Renderer)(nil)
	_ render.PNGExporter        = (*Renderer)(nil)
)
