package gio

import (
	"errors"
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"github.com/cwbudde/matplotlib-go/backends/desktop"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// stubRenderer is a minimal render.Renderer used to keep these tests
// cgo-free. The AGG backend is exercised through the embedding example.
// Only the verbs DrawFigure touches on an empty figure are interesting
// — everything else is a no-op.
type stubRenderer struct {
	w, h int
	img  *image.RGBA
}

func newStubRenderer(w, h int, _ render.Color) (render.Renderer, error) {
	return &stubRenderer{w: w, h: h, img: image.NewRGBA(image.Rect(0, 0, w, h))}, nil
}

func (r *stubRenderer) Begin(geom.Rect) error                  { return nil }
func (r *stubRenderer) End() error                             { return nil }
func (r *stubRenderer) Save()                                  {}
func (r *stubRenderer) Restore()                               {}
func (r *stubRenderer) ClipRect(geom.Rect)                     {}
func (r *stubRenderer) ClipPath(geom.Path)                     {}
func (r *stubRenderer) Path(geom.Path, *render.Paint)          {}
func (r *stubRenderer) Image(render.Image, geom.Rect)          {}
func (r *stubRenderer) GlyphRun(render.GlyphRun, render.Color) {}
func (r *stubRenderer) MeasureText(string, float64, string) render.TextMetrics {
	return render.TextMetrics{}
}

func (r *stubRenderer) GetImage() *image.RGBA { return r.img }

func newTestOptions() desktop.Options {
	fig := core.NewFigure(320, 240)
	return desktop.WithDefaults(desktop.Options{
		Figure:   fig,
		Title:    "test",
		Width:    320,
		Height:   240,
		Renderer: newStubRenderer,
	})
}

func TestRegisteredWithDesktop(t *testing.T) {
	if desktop.DefaultConstructor() == nil {
		t.Fatal("gio.init() did not register a desktop.Constructor")
	}
}

func TestNewBindsNavigationAndToolbar(t *testing.T) {
	b, err := New(newTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	if b.Canvas() == nil {
		t.Fatal("Canvas() returned nil")
	}
	if b.Navigation() == nil {
		t.Fatal("Navigation() returned nil")
	}
	if b.Toolbar() == nil {
		t.Fatal("Toolbar() returned nil")
	}
	// Toolbar pan toggles Navigation.
	tb := b.Toolbar()
	if err := tb.Trigger(canvas.ToolbarPan); err != nil {
		t.Fatal(err)
	}
	if got := b.Navigation().Mode(); got != canvas.NavPan {
		t.Fatalf("nav mode = %v, want NavPan", got)
	}
}

func TestCanvasDrawIdleEmitsEventOnDraw(t *testing.T) {
	b, err := New(newTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	got := 0
	b.Canvas().Connect(canvas.EventDraw, func(canvas.Event) error {
		got++
		return nil
	})
	if err := b.Canvas().Draw(); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("draw events = %d, want 1", got)
	}
}

func TestCanvasResizeEmitsEvent(t *testing.T) {
	b, err := New(newTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	got := 0
	b.Canvas().Connect(canvas.EventResize, func(ev canvas.Event) error {
		got++
		if ev.Width != 800 || ev.Height != 600 {
			t.Fatalf("resize event size = %d×%d, want 800×600", ev.Width, ev.Height)
		}
		return nil
	})
	if err := b.Canvas().Resize(800, 600); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("resize events = %d, want 1", got)
	}
	if b.Canvas().(*gioCanvas).Width() != 800 {
		t.Fatal("width not updated after Resize")
	}
}

func TestDrawIdleMarksDirtyWithoutImmediateDrawEvent(t *testing.T) {
	b, err := New(newTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	got := 0
	b.Canvas().Connect(canvas.EventDraw, func(canvas.Event) error {
		got++
		return nil
	})
	if err := b.Canvas().DrawIdle(); err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Fatalf("draw events after DrawIdle = %d, want 0 before Gio frame draw", got)
	}
	if err := b.Canvas().Draw(); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("draw events after Draw = %d, want 1", got)
	}
}

func TestDrawPropagatesRendererError(t *testing.T) {
	want := errors.New("renderer failed")
	opts := newTestOptions()
	opts.Renderer = func(int, int, render.Color) (render.Renderer, error) {
		return nil, want
	}
	b, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Canvas().Draw(); !errors.Is(err, want) {
		t.Fatalf("Draw error = %v, want %v", err, want)
	}
}

func TestPointerPressEmitsPick(t *testing.T) {
	opts := newTestOptions()
	fig := opts.Figure
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	rect := &core.Rectangle{XY: geom.Pt{X: 0, Y: 0}, Width: 10, Height: 10}
	ax.Add(rect)
	ctx := core.AxesDrawContext(ax, fig)
	pos := (&ctx.DataToPixel).Apply(geom.Pt{X: 5, Y: 5})

	b, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}
	got := make(chan canvas.EventType, 2)
	b.Canvas().Connect(canvas.EventMousePress, func(ev canvas.Event) error {
		got <- ev.Type
		return nil
	})
	b.Canvas().Connect(canvas.EventPick, func(ev canvas.Event) error {
		got <- ev.Type
		return nil
	})
	b.dispatchPointer(pointer.Event{
		Kind:     pointer.Press,
		Position: f32.Pt(float32(pos.X), float32(pos.Y)),
		Buttons:  pointer.ButtonPrimary,
	})

	want := []canvas.EventType{canvas.EventMousePress, canvas.EventPick}
	for i, typ := range want {
		select {
		case gotType := <-got:
			if gotType != typ {
				t.Fatalf("event %d = %s, want %s", i, gotType, typ)
			}
		default:
			t.Fatalf("missing event %d (%s)", i, typ)
		}
	}
}

func TestPointerEnterLeaveEmitsFigureLifecycleEvents(t *testing.T) {
	b, err := New(newTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	got := make(chan canvas.EventType, 2)
	b.Canvas().Connect(canvas.EventFigureEnter, func(ev canvas.Event) error {
		got <- ev.Type
		if ev.Position.X != 12 || ev.Position.Y != 8 {
			t.Fatalf("figure enter position = %+v, want (12,8)", ev.Position)
		}
		return nil
	})
	b.Canvas().Connect(canvas.EventFigureLeave, func(ev canvas.Event) error {
		got <- ev.Type
		if ev.Position.X != 20 || ev.Position.Y != 30 {
			t.Fatalf("figure leave position = %+v, want (20,30)", ev.Position)
		}
		return nil
	})

	b.dispatchPointer(pointer.Event{
		Kind:     pointer.Enter,
		Position: f32.Pt(12, 8),
	})
	b.dispatchPointer(pointer.Event{
		Kind:     pointer.Leave,
		Position: f32.Pt(20, 30),
	})

	want := []canvas.EventType{canvas.EventFigureEnter, canvas.EventFigureLeave}
	for i, typ := range want {
		select {
		case gotType := <-got:
			if gotType != typ {
				t.Fatalf("event %d = %s, want %s", i, gotType, typ)
			}
		default:
			t.Fatalf("missing event %d (%s)", i, typ)
		}
	}
}

func TestPointerMoveEmitsAxesEnterLeaveEvents(t *testing.T) {
	opts := newTestOptions()
	fig := opts.Figure
	left := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 0.5, Y: 1}})
	right := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.5, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})

	b, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}
	var got []canvas.Event
	b.Canvas().Connect(canvas.EventAxesEnter, func(ev canvas.Event) error {
		got = append(got, ev)
		return nil
	})
	b.Canvas().Connect(canvas.EventAxesLeave, func(ev canvas.Event) error {
		got = append(got, ev)
		return nil
	})

	for _, ev := range []pointer.Event{
		{Kind: pointer.Move, Position: f32.Pt(40, 120)},
		{Kind: pointer.Move, Position: f32.Pt(120, 120)},
		{Kind: pointer.Move, Position: f32.Pt(200, 120)},
		{Kind: pointer.Leave, Position: f32.Pt(400, 120)},
	} {
		b.dispatchPointer(ev)
	}

	wantTypes := []canvas.EventType{
		canvas.EventAxesEnter,
		canvas.EventAxesLeave,
		canvas.EventAxesEnter,
		canvas.EventAxesLeave,
	}
	wantAxes := []*core.Axes{left, left, right, right}
	if len(got) != len(wantTypes) {
		t.Fatalf("events = %d, want %d", len(got), len(wantTypes))
	}
	for i := range wantTypes {
		if got[i].Type != wantTypes[i] {
			t.Fatalf("event %d type = %s, want %s", i, got[i].Type, wantTypes[i])
		}
		if got[i].Axes != wantAxes[i] {
			t.Fatalf("event %d axes mismatch", i)
		}
	}
}

func TestPointerReleaseScrollAndKeyPayloads(t *testing.T) {
	b, err := New(newTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	var events []canvas.Event
	for _, typ := range []canvas.EventType{
		canvas.EventMouseRelease,
		canvas.EventPick,
		canvas.EventScroll,
		canvas.EventKeyPress,
		canvas.EventKeyRelease,
	} {
		typ := typ
		b.Canvas().Connect(typ, func(ev canvas.Event) error {
			events = append(events, ev)
			return nil
		})
	}

	b.dispatchPointer(pointer.Event{
		Kind:      pointer.Release,
		Position:  f32.Pt(1, 2),
		Buttons:   pointer.ButtonSecondary,
		Modifiers: key.ModCtrl,
	})
	b.dispatchPointer(pointer.Event{
		Kind:      pointer.Scroll,
		Position:  f32.Pt(3, 4),
		Scroll:    f32.Pt(1, -2),
		Modifiers: key.ModAlt,
	})
	b.dispatchKey(key.Event{
		Name:      key.Name("A"),
		State:     key.Press,
		Modifiers: key.ModShift,
	})
	b.dispatchKey(key.Event{
		Name:      key.NameEscape,
		State:     key.Release,
		Modifiers: key.ModCtrl,
	})

	wantTypes := []canvas.EventType{
		canvas.EventMouseRelease,
		canvas.EventScroll,
		canvas.EventKeyPress,
		canvas.EventKeyRelease,
	}
	if len(events) != len(wantTypes) {
		t.Fatalf("events = %d, want %d: %+v", len(events), len(wantTypes), events)
	}
	for i, want := range wantTypes {
		if events[i].Type != want {
			t.Fatalf("event %d type = %s, want %s", i, events[i].Type, want)
		}
	}
	if events[0].Button != canvas.MouseButtonRight || events[0].Modifiers != canvas.ModifierControl {
		t.Fatalf("release payload = %+v, want right with ctrl", events[0])
	}
	if events[1].DeltaX != 1 || events[1].DeltaY != -2 || events[1].Modifiers != canvas.ModifierAlt {
		t.Fatalf("scroll payload = %+v, want (1,-2) with alt", events[1])
	}
	if events[2].Key != "A" || events[2].Modifiers != canvas.ModifierShift {
		t.Fatalf("key press payload = %+v, want shift+A", events[2])
	}
	if events[3].Key != string(key.NameEscape) || events[3].Modifiers != canvas.ModifierControl {
		t.Fatalf("key release payload = %+v, want ctrl+escape", events[3])
	}
}

func TestMapButtons(t *testing.T) {
	cases := []struct {
		in   pointer.Buttons
		want canvas.MouseButton
	}{
		{0, canvas.MouseButtonNone},
		{pointer.ButtonPrimary, canvas.MouseButtonLeft},
		{pointer.ButtonSecondary, canvas.MouseButtonRight},
		{pointer.ButtonTertiary, canvas.MouseButtonMiddle},
		{pointer.ButtonPrimary | pointer.ButtonSecondary, canvas.MouseButtonLeft},
	}
	for _, c := range cases {
		if got := mapButtons(c.in); got != c.want {
			t.Fatalf("mapButtons(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestMapModifiers(t *testing.T) {
	cases := []struct {
		in   key.Modifiers
		want canvas.Modifier
	}{
		{0, 0},
		{key.ModShift, canvas.ModifierShift},
		{key.ModCtrl, canvas.ModifierControl},
		{key.ModAlt, canvas.ModifierAlt},
		{key.ModSuper, canvas.ModifierMeta},
		{key.ModCommand, canvas.ModifierMeta},
		{key.ModShift | key.ModCtrl, canvas.ModifierShift | canvas.ModifierControl},
	}
	for _, c := range cases {
		if got := mapModifiers(c.in); got != c.want {
			t.Fatalf("mapModifiers(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestCloseIdempotent(t *testing.T) {
	b, err := New(newTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	closes := 0
	b.Canvas().Connect(canvas.EventClose, func(canvas.Event) error {
		closes++
		return nil
	})
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	if closes != 1 {
		t.Fatalf("close events = %d, want 1", closes)
	}
}
