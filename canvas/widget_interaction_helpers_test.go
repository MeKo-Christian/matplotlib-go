package canvas

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func assertCloseEnough(t *testing.T, got, want float64) {
	t.Helper()
	const eps = 1e-9
	if math.Abs(got-want) > eps {
		t.Fatalf("value = %v, want %v", got, want)
	}
}

func sortedPair(a, b float64) (float64, float64) {
	if a <= b {
		return a, b
	}
	return b, a
}

type widgetPickDataArtist struct{}

func (widgetPickDataArtist) Draw(render.Renderer, *core.DrawContext) {}

func (widgetPickDataArtist) Z() float64 { return 10000 }

func (widgetPickDataArtist) Bounds(*core.DrawContext) geom.Rect { return geom.Rect{} }

func (widgetPickDataArtist) Contains(geom.Pt, *core.DrawContext) (bool, core.PickInfo) {
	return true, core.PickInfo{}
}
