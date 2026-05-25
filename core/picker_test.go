package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func newPickerTestAxes(t *testing.T) (*Figure, *Axes) {
	t.Helper()
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	return fig, ax
}

func TestLine2DContainsPickRadius(t *testing.T) {
	fig, ax := newPickerTestAxes(t)
	line := &Line2D{XY: []geom.Pt{{X: 0, Y: 0}, {X: 10, Y: 10}}}
	ax.Add(line)
	ctx := AxesDrawContext(ax, fig)

	pxMid, ok := dataToPx(ctx, geom.Pt{X: 5, Y: 5})
	if !ok {
		t.Fatalf("could not project midpoint")
	}
	if hit, _ := line.Contains(pxMid, ctx); !hit {
		t.Fatalf("expected hit at midpoint %v", pxMid)
	}
	// Display space is y-up, so a data slope +1 stays slope +1 in pixels. Step
	// along the line (+X,+Y) to stay within the pick radius.
	offset := geom.Pt{X: pxMid.X + 4, Y: pxMid.Y + 4}
	if hit, _ := line.Contains(offset, ctx); !hit {
		t.Fatalf("expected hit within default pick radius at %v", offset)
	}
	// Step perpendicular to the diagonal (+X,-Y) to leave the line.
	farAway := geom.Pt{X: pxMid.X + 30, Y: pxMid.Y - 30}
	if hit, _ := line.Contains(farAway, ctx); hit {
		t.Fatalf("did not expect hit far from line at %v", farAway)
	}
}

func TestRectangleContains(t *testing.T) {
	fig, ax := newPickerTestAxes(t)
	rect := &Rectangle{XY: geom.Pt{X: 2, Y: 2}, Width: 4, Height: 4}
	ax.Add(rect)
	ctx := AxesDrawContext(ax, fig)

	inside, ok := dataToPx(ctx, geom.Pt{X: 4, Y: 4})
	if !ok {
		t.Fatalf("data-to-pixel failed")
	}
	if hit, _ := rect.Contains(inside, ctx); !hit {
		t.Fatalf("expected hit at rectangle interior")
	}
	outside, _ := dataToPx(ctx, geom.Pt{X: 9, Y: 9})
	if hit, _ := rect.Contains(outside, ctx); hit {
		t.Fatalf("did not expect hit outside rectangle")
	}
}

func TestCircleContains(t *testing.T) {
	fig, ax := newPickerTestAxes(t)
	c := &Circle{Center: geom.Pt{X: 5, Y: 5}, Radius: 2}
	ax.Add(c)
	ctx := AxesDrawContext(ax, fig)

	center, _ := dataToPx(ctx, geom.Pt{X: 5, Y: 5})
	if hit, _ := c.Contains(center, ctx); !hit {
		t.Fatalf("expected hit at circle center")
	}
	outside, _ := dataToPx(ctx, geom.Pt{X: 9, Y: 9})
	if hit, _ := c.Contains(outside, ctx); hit {
		t.Fatalf("did not expect hit outside circle")
	}
}

func TestPolygonContains(t *testing.T) {
	fig, ax := newPickerTestAxes(t)
	poly := &Polygon{XY: []geom.Pt{{X: 1, Y: 1}, {X: 5, Y: 1}, {X: 3, Y: 5}}}
	ax.Add(poly)
	ctx := AxesDrawContext(ax, fig)

	inside, _ := dataToPx(ctx, geom.Pt{X: 3, Y: 2})
	if hit, _ := poly.Contains(inside, ctx); !hit {
		t.Fatalf("expected hit inside polygon")
	}
	outside, _ := dataToPx(ctx, geom.Pt{X: 6, Y: 5})
	if hit, _ := poly.Contains(outside, ctx); hit {
		t.Fatalf("did not expect hit outside polygon")
	}
}

func TestFormatCoordRoundTrip(t *testing.T) {
	fig, ax := newPickerTestAxes(t)
	ctx := AxesDrawContext(ax, fig)
	pt, _ := dataToPx(ctx, geom.Pt{X: 5, Y: 5})
	got := ax.FormatCoord(pt)
	if got != "x=5 y=5" {
		t.Fatalf("FormatCoord = %q, want %q", got, "x=5 y=5")
	}
}

func TestFormatCoordCustom(t *testing.T) {
	fig, ax := newPickerTestAxes(t)
	ax.SetFormatCoord(func(p geom.Pt) string {
		return "custom"
	})
	ctx := AxesDrawContext(ax, fig)
	pt, _ := dataToPx(ctx, geom.Pt{X: 1, Y: 1})
	if got := ax.FormatCoord(pt); got != "custom" {
		t.Fatalf("FormatCoord = %q, want %q", got, "custom")
	}
}

func dataToPx(ctx *DrawContext, p geom.Pt) (geom.Pt, bool) {
	if ctx == nil {
		return geom.Pt{}, false
	}
	return (&ctx.DataToPixel).Apply(p), true
}
