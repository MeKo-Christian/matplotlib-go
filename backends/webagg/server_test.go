package webagg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	plotcanvas "github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/net/websocket"
)

// newTestServer wraps newTestManager in a NewServer + httptest.Server
// pair so the WebSocket dial actually goes through the real handler.
func newTestServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	fig := core.NewFigure(40, 30)
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
	srv, err := NewServer(ServerOptions{Manager: mgr})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	return srv, ts
}

// TestServeStaticAssets verifies the bundled HTML, JS, and CSS files
// are served with the correct content types.
func TestServeStaticAssets(t *testing.T) {
	_, ts := newTestServer(t)
	cases := []struct {
		path        string
		wantType    string
		wantSnippet string
	}{
		{"/", "text/html; charset=utf-8", "MPLFigure"},
		{"/mpl.js", "application/javascript; charset=utf-8", "WebSocket"},
		{"/mpl.css", "text/css; charset=utf-8", "mpl-figure"},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			resp, err := http.Get(ts.URL + c.path)
			if err != nil {
				t.Fatalf("GET %s: %v", c.path, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status: want 200, got %d", resp.StatusCode)
			}
			gotType := resp.Header.Get("Content-Type")
			if gotType != c.wantType {
				t.Errorf("Content-Type: want %q, got %q", c.wantType, gotType)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if !strings.Contains(string(body), c.wantSnippet) {
				t.Errorf("response body does not contain %q", c.wantSnippet)
			}
		})
	}
}

// TestWebSocketHandshake dials the server's /ws endpoint and verifies
// the initial resize/figure_label/image_mode/refresh sequence arrives,
// followed by a binary PNG frame.
func TestWebSocketHandshake(t *testing.T) {
	_, ts := newTestServer(t)
	wsURL := "ws://" + strings.TrimPrefix(ts.URL, "http://") + "/ws"
	cfg, err := websocket.NewConfig(wsURL, ts.URL)
	if err != nil {
		t.Fatalf("ws config: %v", err)
	}
	conn, err := websocket.DialConfig(cfg)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	defer conn.Close()

	want := []string{"resize", "figure_label", "image_mode", "refresh", "history_buttons"}
	for _, w := range want {
		var data []byte
		if err := websocket.Message.Receive(conn, &data); err != nil {
			t.Fatalf("receive %s: %v", w, err)
		}
		var ev map[string]any
		if err := json.Unmarshal(data, &ev); err != nil {
			t.Fatalf("decode %s: %v (raw=%q)", w, err, string(data))
		}
		if ev["type"] != w {
			t.Errorf("want frame type %q, got %q", w, ev["type"])
		}
	}
	// Then a binary PNG payload.
	var bin []byte
	if err := websocket.Message.Receive(conn, &bin); err != nil {
		t.Fatalf("binary receive: %v", err)
	}
	if !isPNG(bin) {
		t.Errorf("expected PNG payload, got % x", bin[:min(8, len(bin))])
	}
}

func TestWebSocketDrivesPanScrollAndPick(t *testing.T) {
	fig := core.NewFigure(100, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	rect := &core.Rectangle{XY: geom.Pt{X: 0, Y: 0}, Width: 10, Height: 10}
	ax.Add(rect)
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
	srv, err := NewServer(ServerOptions{Manager: mgr})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)

	events := make(chan plotcanvas.EventType, 8)
	for _, typ := range []plotcanvas.EventType{plotcanvas.EventMouseRelease, plotcanvas.EventScroll, plotcanvas.EventPick} {
		typ := typ
		mgr.Connect(typ, func(plotcanvas.Event) error {
			events <- typ
			return nil
		})
	}

	wsURL := "ws://" + strings.TrimPrefix(ts.URL, "http://") + "/ws"
	cfg, err := websocket.NewConfig(wsURL, ts.URL)
	if err != nil {
		t.Fatalf("ws config: %v", err)
	}
	conn, err := websocket.DialConfig(cfg)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	defer conn.Close()
	for range 6 {
		var discard []byte
		if err := websocket.Message.Receive(conn, &discard); err != nil {
			t.Fatalf("initial receive: %v", err)
		}
	}

	center := geom.Pt{X: 50, Y: 40}
	for _, msg := range []map[string]any{
		{"type": "toolbar_button", "name": "pan"},
		{"type": "button_press", "x": center.X, "y": center.Y, "button": 0},
		{"type": "motion_notify", "x": center.X + 20, "y": center.Y, "buttons": 1},
		{"type": "button_release", "x": center.X + 20, "y": center.Y, "button": 0},
		{"type": "scroll", "x": center.X, "y": center.Y, "step": 1.0},
		{"type": "button_press", "x": center.X, "y": center.Y, "button": 0},
	} {
		if err := websocket.Message.Send(conn, mustJSON(t, msg)); err != nil {
			t.Fatalf("send %v: %v", msg, err)
		}
	}

	waitEvent(t, events, plotcanvas.EventMouseRelease)
	waitEvent(t, events, plotcanvas.EventScroll)
	waitEvent(t, events, plotcanvas.EventPick)
	xMin, xMax := ax.XScale.Domain()
	if xMin == 0 && xMax == 10 {
		t.Fatalf("pan/scroll left x limits unchanged")
	}
}

// TestUnknownAssetReturns404 exercises the fallback path of
// serveAsset for paths that don't resolve in the asset FS.
func TestUnknownAssetReturns404(t *testing.T) {
	_, ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/does-not-exist.txt")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status: want 404, got %d", resp.StatusCode)
	}
}

func waitEvent(t *testing.T, events <-chan plotcanvas.EventType, want plotcanvas.EventType) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case got := <-events:
			if got == want {
				return
			}
		case <-deadline:
			t.Fatalf("timed out waiting for %s", want)
		}
	}
}
