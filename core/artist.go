package core

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type Artist interface {
	Draw(r render.Renderer, ctx *DrawContext)
	Z() float64
	Bounds(ctx *DrawContext) geom.Rect
}

type WidgetArtist interface {
	Artist
	WidgetLayer()
}

const (
	defaultPatchZ = 1.0
	defaultLineZ  = 2.0
)

func zOrDefault(z, fallback float64) float64 {
	if z != 0 {
		return z
	}
	return fallback
}

type StickyEdgeArtist interface {
	StickyEdges() (x []float64, y []float64)
}

type OverlayArtist interface {
	DrawOverlay(r render.Renderer, ctx *DrawContext)
}

type ArtistFunc func(r render.Renderer, ctx *DrawContext)

func (f ArtistFunc) Draw(r render.Renderer, ctx *DrawContext) { f(r, ctx) }

func (f ArtistFunc) Z() float64 { return 0 }

func (f ArtistFunc) Bounds(_ *DrawContext) geom.Rect { return geom.Rect{} }
