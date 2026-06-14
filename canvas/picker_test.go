package canvas

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

func newPickerFigure(t *testing.T) (*Figure, *Axes, *core.DrawContext) {
	t.Helper()
	fig := core.NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	ctx := core.AxesDrawContext(ax, fig)
	return fig, ax, ctx
}

func TestPickReturnsTopmostArtist(t *testing.T) {
	fig, ax, ctx := newPickerFigure(t)

	bottom := &core.Rectangle{XY: geom.Pt{X: 0, Y: 0}, Width: 10, Height: 10}
	top := &core.Rectangle{XY: geom.Pt{X: 2, Y: 2}, Width: 4, Height: 4}
	ax.Add(bottom)
	ax.Add(top)

	pos := (&ctx.DataToPixel).Apply(geom.Pt{X: 4, Y: 4})
	hits := Pick(fig, pos)
	if len(hits) == 0 {
		t.Fatalf("expected at least one hit, got none")
	}
	if hits[0].Artist != top {
		t.Fatalf("topmost hit = %v, want top rectangle", hits[0].Artist)
	}
}

func TestPickNoHit(t *testing.T) {
	fig, ax, ctx := newPickerFigure(t)
	rect := &core.Rectangle{XY: geom.Pt{X: 1, Y: 1}, Width: 2, Height: 2}
	ax.Add(rect)

	pos := (&ctx.DataToPixel).Apply(geom.Pt{X: 8, Y: 8})
	if hits := Pick(fig, pos); len(hits) != 0 {
		t.Fatalf("expected no hits, got %d", len(hits))
	}
}

func TestEmitPickFiresPickEvent(t *testing.T) {
	fig, ax, ctx := newPickerFigure(t)
	rect := &core.Rectangle{XY: geom.Pt{X: 0, Y: 0}, Width: 10, Height: 10}
	ax.Add(rect)

	var dispatcher Dispatcher
	var picked PickEvent
	dispatcher.Connect(EventPick, func(ev Event) error {
		picked = PickEvent{Event: ev, Artist: rect}
		return nil
	})

	pos := (&ctx.DataToPixel).Apply(geom.Pt{X: 5, Y: 5})
	mouse := NewMouseEvent(EventMousePress, fig, pos, MouseButtonLeft)
	result, ok := EmitPick(&dispatcher, fig, mouse)
	if !ok {
		t.Fatalf("expected pick result")
	}
	if result.Artist != rect {
		t.Fatalf("result.Artist mismatch")
	}
	if picked.Type != EventPick {
		t.Fatalf("dispatcher did not see pick event")
	}
}
