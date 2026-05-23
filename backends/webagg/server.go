package webagg

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

// Server hosts one [Manager] over HTTP plus WebSockets. It implements
// http.Handler so callers can mount it under any prefix (or wrap it in
// middleware for auth, logging, etc.).
//
// Routes (relative to the mount point):
//
//	GET  /            → index.html (or the host's custom page)
//	GET  /mpl.js      → bundled JS client
//	GET  /mpl.css     → bundled stylesheet
//	WS   /ws          → bi-directional event stream
//
// The zero value is unusable — call [NewServer].
type Server struct {
	manager *Manager
	logger  *log.Logger
	assets  http.FileSystem
	wsRoute string
}

// ServerOptions tunes how the server presents itself. Zero values
// pick sensible defaults.
type ServerOptions struct {
	// Manager is the figure manager to serve. Required.
	Manager *Manager
	// Assets, if non-nil, replaces the bundled static document root.
	// Useful for hosts who want a custom HTML shell. The file system
	// must contain at least index.html, mpl.js, and mpl.css.
	Assets http.FileSystem
	// Logger receives errors from the WebSocket loops. nil disables
	// logging.
	Logger *log.Logger
	// WSRoute is the WebSocket path relative to the mount point. Empty
	// defaults to "/ws". Useful when embedding under a router that
	// rewrites paths.
	WSRoute string
}

// NewServer builds a server that hosts manager. Pass non-nil Assets
// to override the bundled static files.
func NewServer(opts ServerOptions) (*Server, error) {
	if opts.Manager == nil {
		return nil, errors.New("webagg: nil manager")
	}
	assets := opts.Assets
	if assets == nil {
		assetFS, err := bundledAssets()
		if err != nil {
			return nil, err
		}
		assets = http.FS(assetFS)
	}
	wsRoute := opts.WSRoute
	if wsRoute == "" {
		wsRoute = "/ws"
	}
	if !strings.HasPrefix(wsRoute, "/") {
		wsRoute = "/" + wsRoute
	}
	return &Server{
		manager: opts.Manager,
		logger:  opts.Logger,
		assets:  assets,
		wsRoute: wsRoute,
	}, nil
}

// Manager returns the wrapped figure manager.
func (s *Server) Manager() *Manager { return s.manager }

// WSRoute returns the relative WebSocket path. Useful when templating
// HTML that the host serves outside the bundled index.html.
func (s *Server) WSRoute() string { return s.wsRoute }

// ServeHTTP routes one HTTP request. The router strips the server's
// mount prefix before dispatching, so r.URL.Path is the asset-relative
// path.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == s.wsRoute {
		websocket.Server{Handler: s.handleWebSocket}.ServeHTTP(w, r)
		return
	}
	switch path {
	case "", "/":
		s.serveAsset(w, r, "index.html")
		return
	case "/mpl.js":
		s.serveAsset(w, r, "mpl.js")
		return
	case "/mpl.css":
		s.serveAsset(w, r, "mpl.css")
		return
	}
	// Anything else: try to serve it from the asset FS.
	s.serveAsset(w, r, strings.TrimPrefix(path, "/"))
}

func (s *Server) serveAsset(w http.ResponseWriter, r *http.Request, name string) {
	f, err := s.assets.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, "stat: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if stat.IsDir() {
		http.NotFound(w, r)
		return
	}

	// Pick the content type explicitly for our three known assets to
	// avoid relying on file extension sniffing on hosts where that
	// table is incomplete.
	switch {
	case strings.HasSuffix(name, ".js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case strings.HasSuffix(name, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case strings.HasSuffix(name, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	http.ServeContent(w, r, name, stat.ModTime(), f.(readSeekCloser))
}

// readSeekCloser is what fs.File satisfies for the standard library
// files we open out of an embed.FS.
type readSeekCloser interface {
	fs.File
	Read(p []byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
}

// handleWebSocket is the websocket.Handler invoked once per client
// connection. It runs a read loop that decodes inbound JSON events and
// pushes them through the manager's dispatch; outbound frames are
// pushed by the manager via the hub.
func (s *Server) handleWebSocket(ws *websocket.Conn) {
	defer ws.Close()

	sink := &wsSink{ws: ws}
	id, err := s.manager.hub.register(sink)
	if err != nil {
		s.logf("register: %v", err)
		return
	}
	sink.assignedID = id
	defer s.manager.hub.unregister(id)

	// Per upstream's docs, JS payloads are always text frames; we set
	// PayloadType explicitly to avoid being interpreted as binary.
	ws.PayloadType = websocket.TextFrame
	for {
		var data []byte
		if err := websocket.Message.Receive(ws, &data); err != nil {
			// EOF / close → end the connection silently.
			return
		}
		if err := s.manager.HandleClientMessage(data); err != nil {
			s.logf("client %d: %v", id, err)
		}
	}
}

func (s *Server) logf(format string, args ...any) {
	if s.logger == nil {
		return
	}
	s.logger.Printf("webagg: "+format, args...)
}

// wsSink adapts an x/net/websocket connection to the clientSink
// interface. The websocket package itself is not goroutine-safe across
// writes, so we serialize them through a mutex.
type wsSink struct {
	ws         *websocket.Conn
	assignedID uint64

	mu sync.Mutex
}

func (s *wsSink) sendJSON(payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ws.PayloadType = websocket.TextFrame
	_, err := s.ws.Write(payload)
	return err
}

func (s *wsSink) sendBinary(payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ws.PayloadType = websocket.BinaryFrame
	_, err := s.ws.Write(payload)
	return err
}

func (s *wsSink) id() uint64 { return s.assignedID }

func (s *wsSink) close() { rwClose(s.ws) }

// Listen starts an HTTP server on addr that serves this WebAgg server
// at the root. Convenience for hosts that don't need their own router.
func (s *Server) Listen(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/", s)
	return http.ListenAndServe(addr, mux) //nolint:gosec // host opts in via Listen
}

// String returns a human-readable summary, mostly for logs.
func (s *Server) String() string {
	return fmt.Sprintf("webagg.Server{ws=%s, clients=%d}", s.wsRoute, s.manager.hub.clientCount())
}
