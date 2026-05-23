package webagg

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"sync"
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	plotcanvas "github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// newTestManager builds a manager backed by the GoBasic renderer at a
// small fixed size so tests stay fast.
func newTestManager(t *testing.T) *Manager {
	t.Helper()
	return newTestManagerWithLoop(t, nil)
}

func newTestManagerWithLoop(t *testing.T, loop plotcanvas.EventLoop) *Manager {
	t.Helper()
	fig := core.NewFigure(80, 60)
	mgr, err := NewManager(Options{
		Figure: fig,
		Renderer: func(w, h int, bg render.Color) (RasterRenderer, error) {
			r := gobasic.New(w, h, bg)
			if r == nil {
				return nil, fmt.Errorf("gobasic.New returned nil")
			}
			return r, nil
		},
		HasBackground: true,
		Background:    render.Color{R: 1, G: 1, B: 1, A: 1},
		EventLoop:     loop,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return mgr
}

// captureSink records every JSON and binary frame for assertions. It
// implements clientSink so tests bypass the real WebSocket layer.
type captureSink struct {
	mu       sync.Mutex
	jsonMsgs [][]byte
	binMsgs  [][]byte
	clientID uint64
	closed   bool
}

type manualEventLoop struct {
	mu    sync.Mutex
	queue []func() error
}

func (l *manualEventLoop) CallSoon(callback func() error) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.queue = append(l.queue, callback)
	return nil
}

func (l *manualEventLoop) NewTimer(time.Duration, func() error) plotcanvas.Timer {
	return &manualTimer{}
}

func (l *manualEventLoop) RunAll() error {
	for {
		l.mu.Lock()
		if len(l.queue) == 0 {
			l.mu.Unlock()
			return nil
		}
		callback := l.queue[0]
		copy(l.queue, l.queue[1:])
		l.queue = l.queue[:len(l.queue)-1]
		l.mu.Unlock()
		if callback != nil {
			if err := callback(); err != nil {
				return err
			}
		}
	}
}

type manualTimer struct{}

func (*manualTimer) Start() error  { return nil }
func (*manualTimer) Stop() error   { return nil }
func (*manualTimer) Running() bool { return false }

type testStaleArtist struct {
	core.ArtistLifecycle
}

func newTestStaleArtist() *testStaleArtist {
	artist := &testStaleArtist{}
	artist.BindArtist(artist)
	return artist
}

func (*testStaleArtist) Draw(render.Renderer, *core.DrawContext) {}
func (*testStaleArtist) Z() float64                              { return 0 }
func (*testStaleArtist) Bounds(*core.DrawContext) geom.Rect      { return geom.Rect{} }

type blitTestRenderer struct {
	img    *image.RGBA
	begins int
}

func newBlitTestRenderer(w, h int, _ render.Color) (*blitTestRenderer, error) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetRGBA(x, y, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
		}
	}
	return &blitTestRenderer{img: img}, nil
}

func (r *blitTestRenderer) Begin(geom.Rect) error {
	r.begins++
	return nil
}
func (*blitTestRenderer) End() error                             { return nil }
func (*blitTestRenderer) Save()                                  {}
func (*blitTestRenderer) Restore()                               {}
func (*blitTestRenderer) ClipRect(geom.Rect)                     {}
func (*blitTestRenderer) ClipPath(geom.Path)                     {}
func (*blitTestRenderer) Path(geom.Path, *render.Paint)          {}
func (*blitTestRenderer) Image(render.Image, geom.Rect)          {}
func (*blitTestRenderer) GlyphRun(render.GlyphRun, render.Color) {}
func (*blitTestRenderer) MeasureText(string, float64, string) render.TextMetrics {
	return render.TextMetrics{}
}
func (r *blitTestRenderer) GetImage() *image.RGBA { return r.img }
func (r *blitTestRenderer) CopyFromBBox(bbox geom.Rect) *render.BufferRegion {
	minX, minY := int(bbox.Min.X), int(bbox.Min.Y)
	maxX, maxY := int(bbox.Max.X), int(bbox.Max.Y)
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > r.img.Rect.Dx() {
		maxX = r.img.Rect.Dx()
	}
	if maxY > r.img.Rect.Dy() {
		maxY = r.img.Rect.Dy()
	}
	if minX >= maxX || minY >= maxY {
		return nil
	}
	out := image.NewRGBA(image.Rect(0, 0, maxX-minX, maxY-minY))
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			out.SetRGBA(x-minX, y-minY, r.img.RGBAAt(x, y))
		}
	}
	return &render.BufferRegion{
		Image: out,
		Rect:  geom.Rect{Min: geom.Pt{X: float64(minX), Y: float64(minY)}, Max: geom.Pt{X: float64(maxX), Y: float64(maxY)}},
	}
}
func (r *blitTestRenderer) RestoreRegion(region *render.BufferRegion, _ *geom.Rect, offset geom.Pt) {
	if region == nil || region.Image == nil {
		return
	}
	for y := 0; y < region.Image.Rect.Dy(); y++ {
		for x := 0; x < region.Image.Rect.Dx(); x++ {
			r.img.SetRGBA(int(region.Rect.Min.X)+x+int(offset.X), int(region.Rect.Min.Y)+y+int(offset.Y), region.Image.RGBAAt(x, y))
		}
	}
}

func (s *captureSink) sendJSON(payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errClientClosed
	}
	c := make([]byte, len(payload))
	copy(c, payload)
	s.jsonMsgs = append(s.jsonMsgs, c)
	return nil
}

func (s *captureSink) sendBinary(payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errClientClosed
	}
	c := make([]byte, len(payload))
	copy(c, payload)
	s.binMsgs = append(s.binMsgs, c)
	return nil
}

func (s *captureSink) id() uint64 { return s.clientID }

func (s *captureSink) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
}

func (s *captureSink) snapshot() (jsonMsgs, binMsgs [][]byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	jsonMsgs = make([][]byte, len(s.jsonMsgs))
	for i, m := range s.jsonMsgs {
		jsonMsgs[i] = append([]byte{}, m...)
	}
	binMsgs = make([][]byte, len(s.binMsgs))
	for i, m := range s.binMsgs {
		binMsgs[i] = append([]byte{}, m...)
	}
	return jsonMsgs, binMsgs
}

// TestRegisterHandshake checks that a freshly-attached client receives
// the resize / label / image_mode / refresh sequence and a first
// binary PNG frame.
func TestRegisterHandshake(t *testing.T) {
	mgr := newTestManager(t)
	sink := &captureSink{}
	id, err := mgr.hub.register(sink)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if id == 0 {
		t.Fatalf("register returned id 0")
	}
	jsonMsgs, binMsgs := sink.snapshot()
	if len(jsonMsgs) < 4 {
		t.Fatalf("expected at least 4 JSON frames, got %d", len(jsonMsgs))
	}
	want := []string{"resize", "figure_label", "image_mode", "refresh"}
	for i, w := range want {
		typ := decodeType(t, jsonMsgs[i])
		if typ != w {
			t.Errorf("frame %d: want type %q, got %q", i, w, typ)
		}
	}
	if len(binMsgs) != 1 {
		t.Fatalf("expected 1 binary frame after register, got %d", len(binMsgs))
	}
	if !isPNG(binMsgs[0]) {
		t.Errorf("binary frame is not a PNG: % x", binMsgs[0][:8])
	}
}

// TestDiffFrameSecondPaint verifies the manager switches to "diff" mode
// after the first frame when the figure background is opaque.
func TestDiffFrameSecondPaint(t *testing.T) {
	mgr := newTestManager(t)
	sink := &captureSink{}
	if _, err := mgr.hub.register(sink); err != nil {
		t.Fatalf("register: %v", err)
	}
	// Force another paint with no figure changes — should still be a
	// diff frame (an all-transparent image, but a valid PNG).
	if err := mgr.Draw(); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	mode := mgr.currentImageMode()
	if mode != imageModeDiff {
		t.Errorf("expected diff mode after second paint, got %s", mode)
	}
	jsonMsgs, binMsgs := sink.snapshot()
	// Must have an additional image_mode announcement and binary frame.
	if len(binMsgs) < 2 {
		t.Fatalf("expected at least 2 binary frames, got %d", len(binMsgs))
	}
	if !containsType(jsonMsgs, "image_mode", "diff") {
		t.Errorf("expected image_mode=diff announcement; got %v", typesOf(jsonMsgs))
	}
}

// TestResizeForcesFullFrame asserts that resizing flushes the diff
// cache and switches the next frame back to full.
func TestResizeForcesFullFrame(t *testing.T) {
	mgr := newTestManager(t)
	sink := &captureSink{}
	if _, err := mgr.hub.register(sink); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := mgr.Resize(96, 72); err != nil {
		t.Fatalf("Resize: %v", err)
	}
	if mgr.currentImageMode() != imageModeFull {
		t.Errorf("expected full mode after resize, got %s", mgr.currentImageMode())
	}
}

func TestDrawIdleCoalescesRequestsUntilIdleTick(t *testing.T) {
	loop := &manualEventLoop{}
	mgr := newTestManagerWithLoop(t, loop)
	sink := &captureSink{}
	if _, err := mgr.hub.register(sink); err != nil {
		t.Fatalf("register: %v", err)
	}

	draws := 0
	mgr.Connect(plotcanvas.EventDraw, func(plotcanvas.Event) error {
		draws++
		return nil
	})
	_, beforeBins := sink.snapshot()

	for range 5 {
		if err := mgr.DrawIdle(); err != nil {
			t.Fatalf("DrawIdle: %v", err)
		}
	}
	_, queuedBins := sink.snapshot()
	if draws != 0 {
		t.Fatalf("draw events before idle tick = %d, want 0", draws)
	}
	if len(queuedBins) != len(beforeBins) {
		t.Fatalf("binary frames before idle tick = %d, want %d", len(queuedBins), len(beforeBins))
	}

	if err := loop.RunAll(); err != nil {
		t.Fatalf("RunAll: %v", err)
	}
	_, afterBins := sink.snapshot()
	if draws != 1 {
		t.Fatalf("draw events after idle tick = %d, want 1", draws)
	}
	if len(afterBins) != len(beforeBins)+1 {
		t.Fatalf("binary frames after idle tick = %d, want %d", len(afterBins), len(beforeBins)+1)
	}

	if err := loop.RunAll(); err != nil {
		t.Fatalf("second RunAll: %v", err)
	}
	if draws != 1 {
		t.Fatalf("draw events after empty idle tick = %d, want 1", draws)
	}
}

func TestStaleArtistCallbackSchedulesOneIdleDraw(t *testing.T) {
	loop := &manualEventLoop{}
	fig := core.NewFigure(80, 60)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	artist := newTestStaleArtist()
	ax.Add(artist)
	mgr, err := NewManager(Options{
		Figure: fig,
		Renderer: func(w, h int, bg render.Color) (RasterRenderer, error) {
			return gobasic.New(w, h, bg), nil
		},
		HasBackground: true,
		Background:    render.Color{R: 1, G: 1, B: 1, A: 1},
		EventLoop:     loop,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	sink := &captureSink{}
	if _, err := mgr.hub.register(sink); err != nil {
		t.Fatalf("register: %v", err)
	}

	draws := 0
	mgr.Connect(plotcanvas.EventDraw, func(plotcanvas.Event) error {
		draws++
		return nil
	})
	_, beforeBins := sink.snapshot()

	artist.MarkStale()
	artist.MarkStale()
	if !artist.Stale() {
		t.Fatalf("artist not stale after MarkStale")
	}
	if draws != 0 {
		t.Fatalf("draw events before idle tick = %d, want 0", draws)
	}
	if err := loop.RunAll(); err != nil {
		t.Fatalf("RunAll: %v", err)
	}
	_, afterBins := sink.snapshot()
	if draws != 1 {
		t.Fatalf("draw events after idle tick = %d, want 1", draws)
	}
	if len(afterBins) != len(beforeBins)+1 {
		t.Fatalf("binary frames after idle tick = %d, want %d", len(afterBins), len(beforeBins)+1)
	}
	if artist.Stale() {
		t.Fatalf("artist still stale after redraw")
	}
}

func TestBlitBroadcastsRendererDamageWithoutFullDraw(t *testing.T) {
	var renderer *blitTestRenderer
	fig := core.NewFigure(16, 12)
	mgr, err := NewManager(Options{
		Figure: fig,
		Renderer: func(w, h int, bg render.Color) (RasterRenderer, error) {
			r, err := newBlitTestRenderer(w, h, bg)
			renderer = r
			return r, err
		},
		HasBackground: true,
		Background:    render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	sink := &captureSink{}
	if _, err := mgr.hub.register(sink); err != nil {
		t.Fatalf("register: %v", err)
	}
	if renderer == nil || renderer.begins != 1 {
		t.Fatalf("initial draw begins = %d, want 1", renderer.begins)
	}

	damage := geom.Rect{Min: geom.Pt{X: 2, Y: 2}, Max: geom.Pt{X: 5, Y: 5}}
	region := mgr.CopyFromBBox(damage)
	if region == nil {
		t.Fatalf("CopyFromBBox returned nil")
	}
	renderer.img.SetRGBA(3, 3, color.RGBA{R: 0xff, A: 0xff})
	_, beforeBins := sink.snapshot()
	if err := mgr.Blit(damage); err != nil {
		t.Fatalf("Blit: %v", err)
	}
	_, afterBins := sink.snapshot()
	if len(afterBins) != len(beforeBins)+1 {
		t.Fatalf("binary frames after blit = %d, want %d", len(afterBins), len(beforeBins)+1)
	}
	if renderer.begins != 1 {
		t.Fatalf("full draw begins after blit = %d, want 1", renderer.begins)
	}
}

// TestHandleClientMessageMouse round-trips a button_press JSON event
// into a canvas mouse-press dispatch. The press should show up at the
// position the JSON event carried.
func TestHandleClientMessageMouse(t *testing.T) {
	mgr := newTestManager(t)
	got := make(chan plotcanvas.Event, 1)
	mgr.Connect(plotcanvas.EventMousePress, func(ev plotcanvas.Event) error {
		got <- ev
		return nil
	})
	raw := mustJSON(t, map[string]any{
		"type":   "button_press",
		"x":      12.0,
		"y":      8.0,
		"button": 0,
	})
	if err := mgr.HandleClientMessage(raw); err != nil {
		t.Fatalf("HandleClientMessage: %v", err)
	}
	select {
	case ev := <-got:
		if ev.Position.X != 12 || ev.Position.Y != 8 {
			t.Errorf("position: want (12,8), got %+v", ev.Position)
		}
		if ev.Button != plotcanvas.MouseButtonLeft {
			t.Errorf("button: want left, got %v", ev.Button)
		}
	default:
		t.Fatalf("expected EventMousePress to be dispatched")
	}
}

func TestHandleClientMessageDoubleClickPayload(t *testing.T) {
	mgr := newTestManager(t)
	got := make(chan plotcanvas.Event, 1)
	mgr.Connect(plotcanvas.EventMousePress, func(ev plotcanvas.Event) error {
		got <- ev
		return nil
	})

	if err := mgr.HandleClientMessage(mustJSON(t, map[string]any{
		"type":      "dblclick",
		"x":         12.0,
		"y":         8.0,
		"button":    2,
		"modifiers": []string{"ctrl", "shift"},
	})); err != nil {
		t.Fatalf("HandleClientMessage: %v", err)
	}

	select {
	case ev := <-got:
		if ev.Type != plotcanvas.EventMousePress {
			t.Fatalf("type = %s, want %s", ev.Type, plotcanvas.EventMousePress)
		}
		if !ev.DoubleClick {
			t.Fatal("DoubleClick = false, want true")
		}
		if ev.Button != plotcanvas.MouseButtonRight {
			t.Fatalf("button = %v, want right", ev.Button)
		}
		wantMods := plotcanvas.ModifierControl | plotcanvas.ModifierShift
		if ev.Modifiers != wantMods {
			t.Fatalf("modifiers = %v, want %v", ev.Modifiers, wantMods)
		}
	default:
		t.Fatalf("expected double-click EventMousePress to be dispatched")
	}
}

func TestHandleClientMessageReleaseScrollAndKeyPayloads(t *testing.T) {
	mgr := newTestManager(t)
	var events []plotcanvas.Event
	for _, typ := range []plotcanvas.EventType{
		plotcanvas.EventMouseRelease,
		plotcanvas.EventPick,
		plotcanvas.EventScroll,
		plotcanvas.EventKeyPress,
		plotcanvas.EventKeyRelease,
	} {
		typ := typ
		mgr.Connect(typ, func(ev plotcanvas.Event) error {
			events = append(events, ev)
			return nil
		})
	}

	for _, msg := range []map[string]any{
		{"type": "button_release", "x": 1.0, "y": 2.0, "button": 1},
		{"type": "scroll", "x": 3.0, "y": 4.0, "step": -2.0, "modifiers": []string{"alt"}},
		{"type": "key_press", "key": "ctrl+a", "modifiers": []string{"ctrl"}},
		{"type": "key_release", "key": "shift+escape", "modifiers": []string{"shift"}},
	} {
		if err := mgr.HandleClientMessage(mustJSON(t, msg)); err != nil {
			t.Fatalf("HandleClientMessage %v: %v", msg, err)
		}
	}

	wantTypes := []plotcanvas.EventType{
		plotcanvas.EventMouseRelease,
		plotcanvas.EventScroll,
		plotcanvas.EventKeyPress,
		plotcanvas.EventKeyRelease,
	}
	if len(events) != len(wantTypes) {
		t.Fatalf("events = %d, want %d: %+v", len(events), len(wantTypes), events)
	}
	for i, want := range wantTypes {
		if events[i].Type != want {
			t.Fatalf("event %d type = %s, want %s", i, events[i].Type, want)
		}
	}
	if events[0].Button != plotcanvas.MouseButtonMiddle {
		t.Fatalf("release button = %v, want middle", events[0].Button)
	}
	if events[1].DeltaY != -2 || events[1].Modifiers != plotcanvas.ModifierAlt {
		t.Fatalf("scroll payload = %+v, want step -2 with alt", events[1])
	}
	if events[2].Key != "a" || events[2].Modifiers != plotcanvas.ModifierControl {
		t.Fatalf("key press payload = %+v, want ctrl+a normalized to a", events[2])
	}
	if events[3].Key != "escape" || events[3].Modifiers != plotcanvas.ModifierShift {
		t.Fatalf("key release payload = %+v, want shift+escape normalized to escape", events[3])
	}
}

func TestHandleClientMessagePressEmitsPick(t *testing.T) {
	fig := core.NewFigure(100, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	rect := &core.Rectangle{XY: geom.Pt{X: 0, Y: 0}, Width: 10, Height: 10}
	ax.Add(rect)
	ctx := core.AxesDrawContext(ax, fig)
	pos := (&ctx.DataToPixel).Apply(geom.Pt{X: 5, Y: 5})
	mgr, err := NewManager(Options{
		Figure: fig,
		Renderer: func(w, h int, bg render.Color) (RasterRenderer, error) {
			return gobasic.New(w, h, bg), nil
		},
		HasBackground: true,
		Background:    render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	got := make(chan plotcanvas.EventType, 2)
	mgr.Connect(plotcanvas.EventMousePress, func(ev plotcanvas.Event) error {
		got <- ev.Type
		return nil
	})
	mgr.Connect(plotcanvas.EventPick, func(ev plotcanvas.Event) error {
		got <- ev.Type
		return nil
	})
	raw := mustJSON(t, map[string]any{
		"type":   "button_press",
		"x":      pos.X,
		"y":      pos.Y,
		"button": 0,
	})
	if err := mgr.HandleClientMessage(raw); err != nil {
		t.Fatalf("HandleClientMessage: %v", err)
	}
	want := []plotcanvas.EventType{plotcanvas.EventMousePress, plotcanvas.EventPick}
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

func TestHandleClientMessageFigureEnterLeave(t *testing.T) {
	mgr := newTestManager(t)
	got := make(chan plotcanvas.EventType, 2)
	mgr.Connect(plotcanvas.EventFigureEnter, func(ev plotcanvas.Event) error {
		got <- ev.Type
		if ev.Position.X != 12 || ev.Position.Y != 8 {
			t.Fatalf("figure enter position = %+v, want (12,8)", ev.Position)
		}
		return nil
	})
	mgr.Connect(plotcanvas.EventFigureLeave, func(ev plotcanvas.Event) error {
		got <- ev.Type
		if ev.Position.X != 20 || ev.Position.Y != 30 {
			t.Fatalf("figure leave position = %+v, want (20,30)", ev.Position)
		}
		return nil
	})

	if err := mgr.HandleClientMessage(mustJSON(t, map[string]any{
		"type": "figure_enter",
		"x":    12.0,
		"y":    8.0,
	})); err != nil {
		t.Fatalf("HandleClientMessage enter: %v", err)
	}
	if err := mgr.HandleClientMessage(mustJSON(t, map[string]any{
		"type": "figure_leave",
		"x":    20.0,
		"y":    30.0,
	})); err != nil {
		t.Fatalf("HandleClientMessage leave: %v", err)
	}

	want := []plotcanvas.EventType{plotcanvas.EventFigureEnter, plotcanvas.EventFigureLeave}
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

func TestHandleClientMessageAxesEnterLeave(t *testing.T) {
	fig := core.NewFigure(200, 100)
	left := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 0.5, Y: 1}})
	right := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.5, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	mgr, err := NewManager(Options{
		Figure: fig,
		Renderer: func(w, h int, bg render.Color) (RasterRenderer, error) {
			return gobasic.New(w, h, bg), nil
		},
		HasBackground: true,
		Background:    render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	var got []plotcanvas.Event
	mgr.Connect(plotcanvas.EventAxesEnter, func(ev plotcanvas.Event) error {
		got = append(got, ev)
		return nil
	})
	mgr.Connect(plotcanvas.EventAxesLeave, func(ev plotcanvas.Event) error {
		got = append(got, ev)
		return nil
	})

	for _, msg := range []map[string]any{
		{"type": "motion_notify", "x": 25.0, "y": 50.0},
		{"type": "motion_notify", "x": 75.0, "y": 50.0},
		{"type": "motion_notify", "x": 125.0, "y": 50.0},
		{"type": "motion_notify", "x": 250.0, "y": 50.0},
	} {
		if err := mgr.HandleClientMessage(mustJSON(t, msg)); err != nil {
			t.Fatalf("HandleClientMessage %v: %v", msg, err)
		}
	}

	wantTypes := []plotcanvas.EventType{
		plotcanvas.EventAxesEnter,
		plotcanvas.EventAxesLeave,
		plotcanvas.EventAxesEnter,
		plotcanvas.EventAxesLeave,
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

// TestToolbarPanEcho asserts that a toolbar_button:pan event flips the
// controller into pan mode and broadcasts a navigate_mode message so
// other clients can update their UI.
func TestToolbarPanEcho(t *testing.T) {
	mgr := newTestManager(t)
	sink := &captureSink{}
	if _, err := mgr.hub.register(sink); err != nil {
		t.Fatalf("register: %v", err)
	}
	raw := mustJSON(t, map[string]any{
		"type": "toolbar_button",
		"name": "pan",
	})
	if err := mgr.HandleClientMessage(raw); err != nil {
		t.Fatalf("HandleClientMessage: %v", err)
	}
	if mgr.Toolbar().Mode() != plotcanvas.ToolbarModePan {
		t.Errorf("expected pan mode, got %v", mgr.Toolbar().Mode())
	}
	jsonMsgs, _ := sink.snapshot()
	if !containsType(jsonMsgs, "navigate_mode", "PAN") {
		t.Errorf("expected navigate_mode=PAN broadcast; got %v", typesOf(jsonMsgs))
	}
}

// TestPanShiftsAxes drives a pan drag through the HandleClientMessage
// path and asserts the underlying axes view limits moved.
func TestPanShiftsAxes(t *testing.T) {
	fig := core.NewFigure(100, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	mgr, err := NewManager(Options{
		Figure: fig,
		Renderer: func(w, h int, bg render.Color) (RasterRenderer, error) {
			return gobasic.New(w, h, bg), nil
		},
		HasBackground: true,
		Background:    render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if err := mgr.Toolbar().Trigger(plotcanvas.ToolbarPan); err != nil {
		t.Fatalf("Trigger pan: %v", err)
	}

	steps := []map[string]any{
		{"type": "button_press", "x": 50.0, "y": 40.0, "button": 0},
		{"type": "motion_notify", "x": 70.0, "y": 40.0, "buttons": 1},
		{"type": "button_release", "x": 70.0, "y": 40.0, "button": 0},
	}
	for _, s := range steps {
		if err := mgr.HandleClientMessage(mustJSON(t, s)); err != nil {
			t.Fatalf("HandleClientMessage %v: %v", s["type"], err)
		}
	}
	xMin, xMax := ax.XScale.Domain()
	if xMin >= 0 || xMax >= 10 {
		t.Errorf("pan did not shift X domain: got [%g, %g], expected both < their initials", xMin, xMax)
	}
}

// TestPackAndDiff exercises the pixel-packing and diff routines on a
// hand-built 2x2 image.
func TestPackAndDiff(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 0xff, A: 0xff})
	img.Set(1, 0, color.RGBA{G: 0xff, A: 0xff})
	img.Set(0, 1, color.RGBA{B: 0xff, A: 0xff})
	img.Set(1, 1, color.RGBA{R: 0xff, G: 0xff, A: 0xff})

	packed := packRGBA(img, 2, 2)
	if len(packed) != 4 {
		t.Fatalf("packed length: want 4, got %d", len(packed))
	}
	// Distinct pixels — every pack value should be unique.
	seen := map[uint32]bool{}
	for _, v := range packed {
		if seen[v] {
			t.Fatalf("duplicate packed value %#x", v)
		}
		seen[v] = true
	}

	// Build a "prior" buffer where (0,0) matches and (1,1) differs.
	prior := append([]uint32{}, packed...)
	prior[3] = 0 // (1,1) "changed"
	diff := diffImage(img, prior, packed, 2, 2)
	// (1,1) should carry the current pixel; the rest should be 0.
	for y := range 2 {
		for x := range 2 {
			got := diff.RGBAAt(x, y)
			if x == 1 && y == 1 {
				if got != (color.RGBA{R: 0xff, G: 0xff, A: 0xff}) {
					t.Errorf("diff(1,1) want yellow, got %+v", got)
				}
			} else if got.A != 0 {
				t.Errorf("diff(%d,%d) want transparent, got %+v", x, y, got)
			}
		}
	}
}

// TestImageHasTransparency exercises the alpha-aware fast path.
func TestImageHasTransparency(t *testing.T) {
	opaque := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for i := 0; i < len(opaque.Pix); i++ {
		opaque.Pix[i] = 0xff
	}
	if imageHasTransparency(opaque) {
		t.Errorf("opaque image reported as transparent")
	}
	blank := image.NewRGBA(image.Rect(0, 0, 2, 2))
	if !imageHasTransparency(blank) {
		t.Errorf("transparent image reported as opaque")
	}
}

// TestNormalizeKeyStripsModifiers verifies the prefix-trimmer matches
// upstream's "ctrl+a" → "a" reduction.
func TestNormalizeKeyStripsModifiers(t *testing.T) {
	cases := map[string]string{
		"a":           "a",
		"ctrl+a":      "a",
		"shift+x":     "x",
		"alt+meta+z":  "z", // every known modifier prefix is stripped
		"unknown+key": "unknown+key",
	}
	for in, want := range cases {
		got := normalizeKey(in)
		if got != want {
			t.Errorf("normalizeKey(%q) = %q, want %q", in, got, want)
		}
	}
}

// ---- helpers ---------------------------------------------------------

func decodeType(t *testing.T, payload []byte) string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	s, _ := m["type"].(string)
	return s
}

func containsType(msgs [][]byte, typ, modeOrLabel string) bool {
	for _, m := range msgs {
		var ev map[string]any
		if err := json.Unmarshal(m, &ev); err != nil {
			continue
		}
		if ev["type"] != typ {
			continue
		}
		if modeOrLabel == "" {
			return true
		}
		if ev["mode"] == modeOrLabel || ev["label"] == modeOrLabel {
			return true
		}
	}
	return false
}

func typesOf(msgs [][]byte) []string {
	out := make([]string, 0, len(msgs))
	for _, m := range msgs {
		var ev map[string]any
		_ = json.Unmarshal(m, &ev)
		s, _ := ev["type"].(string)
		out = append(out, s)
	}
	return out
}

func isPNG(b []byte) bool {
	return len(b) > 8 &&
		b[0] == 0x89 && b[1] == 0x50 && b[2] == 0x4e && b[3] == 0x47 &&
		b[4] == 0x0d && b[5] == 0x0a && b[6] == 0x1a && b[7] == 0x0a
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}
