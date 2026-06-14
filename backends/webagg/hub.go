package webagg

import (
	"encoding/json"
	"errors"
	"io"
	"sync"

	"github.com/cwbudde/matplotlib-go/geom"
)

// clientSink is the narrow contract a hub uses to push bytes to one
// connected client. Implementations exist for x/net/websocket
// connections and for tests that capture frames in-memory.
type clientSink interface {
	sendJSON(payload []byte) error
	sendBinary(payload []byte) error
	id() uint64
	close()
}

// hub fans frames out to every currently-connected client. The hub
// owns the broadcast lock; sends to each client are synchronized so a
// slow client cannot interleave a binary PNG with a text frame.
type hub struct {
	mgr *Manager

	mu      sync.Mutex
	nextID  uint64
	clients map[uint64]clientSink
}

func newHub(m *Manager) *hub {
	return &hub{mgr: m, clients: make(map[uint64]clientSink)}
}

// register adds a client and sends the initial handshake (refresh +
// image_mode + figure_label + first frame). The returned uint64 is the
// per-hub client ID; pass it to unregister on disconnect.
func (h *hub) register(sink clientSink) (uint64, error) {
	h.mu.Lock()
	h.nextID++
	id := h.nextID
	h.clients[id] = sink
	h.mu.Unlock()

	// Initial handshake — order mirrors upstream
	// FigureManagerWebAgg.add_web_socket / handle_refresh.
	h.mgr.mu.Lock()
	mode := string(h.mgr.imageMode)
	label := h.mgr.windowTitle
	w := h.mgr.figure.SizePx.X
	hh := h.mgr.figure.SizePx.Y
	h.mgr.forceFull = true
	h.mgr.mu.Unlock()

	forwardFalse := false
	resize := &outboundEvent{
		Type:      MsgResize,
		Size:      []float64{w, hh},
		ResizeFwd: &forwardFalse,
	}
	if err := sendJSONTo(sink, resize); err != nil {
		return id, err
	}
	if err := sendJSONTo(sink, &outboundEvent{Type: MsgFigureLabel, Label: label}); err != nil {
		return id, err
	}
	if err := sendJSONTo(sink, &outboundEvent{Type: MsgImageMode, Mode: mode}); err != nil {
		return id, err
	}
	if err := sendJSONTo(sink, &outboundEvent{Type: MsgRefresh}); err != nil {
		return id, err
	}
	// Push the first frame so the client has something to paint with.
	return id, h.broadcastFrame()
}

// unregister removes a client and closes its sink.
func (h *hub) unregister(id uint64) {
	h.mu.Lock()
	sink, ok := h.clients[id]
	delete(h.clients, id)
	h.mu.Unlock()
	if ok {
		sink.close()
	}
}

// closeAll boots every client. Called from Manager.Close.
func (h *hub) closeAll() {
	h.mu.Lock()
	clients := make([]clientSink, 0, len(h.clients))
	for _, c := range h.clients {
		clients = append(clients, c)
	}
	h.clients = make(map[uint64]clientSink)
	h.mu.Unlock()
	for _, c := range clients {
		c.close()
	}
}

// clientCount returns the number of currently-connected clients.
// Exposed for tests.
func (h *hub) clientCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

// broadcastJSON marshals payload once and sends it to every client.
// Errors are swallowed per-client because one slow client should not
// stop the broadcast — upstream WebAgg has the same policy.
func (h *hub) broadcastJSON(payload *outboundEvent) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.mu.Lock()
	clients := make([]clientSink, 0, len(h.clients))
	for _, c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.Unlock()
	for _, c := range clients {
		_ = c.sendJSON(data)
	}
}

// broadcastFrame paints the current figure and sends the resulting
// PNG (with a preceding image_mode announce if the mode changed) to
// every client.
func (h *hub) broadcastFrame() error {
	if len(h.clientsSnapshot()) == 0 {
		return nil
	}
	prevMode := h.mgr.currentImageMode()
	mode, payload, err := h.mgr.renderFrame()
	if err != nil {
		return err
	}
	if mode != prevMode {
		h.broadcastJSON(&outboundEvent{Type: MsgImageMode, Mode: string(mode)})
	}
	clients := h.clientsSnapshot()
	for _, c := range clients {
		_ = c.sendBinary(payload)
	}
	return nil
}

func (h *hub) broadcastBlitFrame(damage *geom.Rect) error {
	if len(h.clientsSnapshot()) == 0 {
		return nil
	}
	prevMode := h.mgr.currentImageMode()
	mode, payload, err := h.mgr.renderCurrentFrame(damage)
	if err != nil {
		return err
	}
	if mode != prevMode {
		h.broadcastJSON(&outboundEvent{Type: MsgImageMode, Mode: string(mode)})
	}
	clients := h.clientsSnapshot()
	for _, c := range clients {
		_ = c.sendBinary(payload)
	}
	return nil
}

func (h *hub) clientsSnapshot() []clientSink {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]clientSink, 0, len(h.clients))
	for _, c := range h.clients {
		out = append(out, c)
	}
	return out
}

// currentImageMode returns the manager's most-recently-broadcast image
// mode. The hub uses this to decide whether to announce a mode change
// before pushing a frame.
func (m *Manager) currentImageMode() imageMode {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.imageMode
}

func sendJSONTo(sink clientSink, payload *outboundEvent) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return sink.sendJSON(data)
}

// errClientClosed is returned by clientSink implementations after
// close() has been called.
var errClientClosed = errors.New("webagg: client closed")

// rwClose closes an io.WriteCloser, ignoring io.EOF and similar
// already-closed signals. Helper kept here to avoid duplication
// between the websocket and test sinks.
func rwClose(c io.Closer) {
	if c == nil {
		return
	}
	_ = c.Close()
}
