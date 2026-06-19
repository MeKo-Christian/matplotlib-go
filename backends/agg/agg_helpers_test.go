package agg

import (
	"image"
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

var white = render.Color{R: 1, G: 1, B: 1, A: 1}

// mustNew creates a renderer or fails the test.
func mustNew(t *testing.T, w, h int) *Renderer {
	t.Helper()
	r, err := New(w, h, white)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func upperLeftTriangleClip() geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.LineTo(geom.Pt{X: 70, Y: 0})
	p.LineTo(geom.Pt{X: 0, Y: 70})
	p.Close()
	return p
}

func fullRectPath(w, h float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.LineTo(geom.Pt{X: w, Y: 0})
	p.LineTo(geom.Pt{X: w, Y: h})
	p.LineTo(geom.Pt{X: 0, Y: h})
	p.Close()
	return p
}

func aggTestRectPath(x0, y0, x1, y1 float64) geom.Path {
	p := geom.Path{}
	p.MoveTo(geom.Pt{X: x0, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y1})
	p.LineTo(geom.Pt{X: x0, Y: y1})
	p.Close()
	return p
}

func inkBounds(img *image.RGBA, background color.RGBA) (geom.Rect, int, bool) {
	if img == nil {
		return geom.Rect{}, 0, false
	}

	b := img.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X, b.Min.Y
	pixels := 0

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y) == background {
				continue
			}
			pixels++
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x >= maxX {
				maxX = x + 1
			}
			if y >= maxY {
				maxY = y + 1
			}
		}
	}

	if pixels == 0 {
		return geom.Rect{}, 0, false
	}

	return geom.Rect{
		Min: geom.Pt{X: float64(minX), Y: float64(minY)},
		Max: geom.Pt{X: float64(maxX), Y: float64(maxY)},
	}, pixels, true
}
