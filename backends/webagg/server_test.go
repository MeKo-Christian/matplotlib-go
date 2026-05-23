package webagg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	"github.com/cwbudde/matplotlib-go/core"
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

	want := []string{"resize", "figure_label", "image_mode", "refresh"}
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
