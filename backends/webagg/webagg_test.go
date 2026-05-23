package webagg

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"sync"
	"testing"

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
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
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
