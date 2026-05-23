package webagg

import (
	"encoding/json"

	plotcanvas "github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// HandleClientMessage decodes a JSON event from a client and applies
// its side effects on the manager — repainting, mutating navigation
// state, or echoing announcements back to all clients. Exported so
// tests can drive the manager without going through a WebSocket.
func (m *Manager) HandleClientMessage(raw []byte) error {
	if m.closed.Load() {
		return errClientClosed
	}
	var ev inboundEvent
	if err := json.Unmarshal(raw, &ev); err != nil {
		return err
	}
	return m.dispatchInbound(&ev)
}

func (m *Manager) dispatchInbound(ev *inboundEvent) error {
	switch ev.Type {
	case MsgAck:
		// Upstream uses ack messages as a flow-control heartbeat; the
		// server side does not need to act on them.
		return nil
	case MsgDrawCmd:
		return m.DrawIdle()
	case MsgButtonPress:
		return m.emitMouse(plotcanvas.EventMousePress, ev)
	case MsgButtonRelease:
		return m.emitMouse(plotcanvas.EventMouseRelease, ev)
	case MsgDblClick:
		return m.emitMouse(plotcanvas.EventMousePress, ev)
	case MsgMotionNotify:
		return m.emitMouse(plotcanvas.EventMouseMove, ev)
	case MsgFigureEnter:
		return m.emitMouse(plotcanvas.EventFigureEnter, ev)
	case MsgFigureLeave:
		return m.emitMouse(plotcanvas.EventFigureLeave, ev)
	case MsgScroll:
		return m.emitScroll(ev)
	case MsgKeyPress:
		return m.emitKey(plotcanvas.EventKeyPress, ev)
	case MsgKeyRelease:
		return m.emitKey(plotcanvas.EventKeyRelease, ev)
	case MsgToolbarButton:
		return m.handleToolbarButton(ev.Name)
	case MsgRefreshCmd:
		return m.handleRefresh()
	case MsgResizeCmd:
		return m.Resize(int(ev.Width+0.5), int(ev.Height+0.5))
	case MsgSendImageMode:
		mode := m.currentImageMode()
		m.hub.broadcastJSON(&outboundEvent{Type: MsgImageMode, Mode: string(mode)})
		return nil
	case MsgSetDevicePixelRatio:
		return m.handleSetDPR(ev.DevicePixelRatio)
	case MsgSetDPIRatio:
		return m.handleSetDPR(ev.DPIRatio)
	}
	// Unknown message types are not fatal — upstream logs and skips.
	return nil
}

func (m *Manager) emitMouse(eventType plotcanvas.EventType, ev *inboundEvent) error {
	m.mu.Lock()
	m.lastMouseX = ev.X
	m.lastMouseY = ev.Y
	m.mu.Unlock()
	event := plotcanvas.Event{
		Type:      eventType,
		Figure:    m.figure,
		Position:  geom.Pt{X: ev.X, Y: ev.Y},
		Button:    mouseButtonFromJS(ev.Button),
		Modifiers: modifiersFromJS(ev.Modifiers),
	}
	if ev.Type == MsgDblClick {
		event.DoubleClick = true
	}
	if err := m.dispatch.Emit(event); err != nil {
		return err
	}
	if eventType == plotcanvas.EventMousePress {
		plotcanvas.EmitPick(m.dispatch, m.figure, plotcanvas.MouseEvent{Event: event})
	}
	switch eventType {
	case plotcanvas.EventMouseMove, plotcanvas.EventFigureEnter, plotcanvas.EventFigureLeave:
		m.hover.Update(event)
	}
	return nil
}

func (m *Manager) emitScroll(ev *inboundEvent) error {
	m.mu.Lock()
	m.lastMouseX = ev.X
	m.lastMouseY = ev.Y
	m.mu.Unlock()
	// Step direction matches upstream: positive = scroll up = zoom in.
	// Canvas.Navigation.handleScroll already maps a positive DeltaY to
	// a zoom-in factor, so passing Step straight through preserves the
	// upstream semantic.
	return m.dispatch.Emit(plotcanvas.Event{
		Type:      plotcanvas.EventScroll,
		Figure:    m.figure,
		Position:  geom.Pt{X: ev.X, Y: ev.Y},
		DeltaY:    ev.Step,
		Modifiers: modifiersFromJS(ev.Modifiers),
	})
}

func (m *Manager) emitKey(eventType plotcanvas.EventType, ev *inboundEvent) error {
	return m.dispatch.Emit(plotcanvas.Event{
		Type:      eventType,
		Figure:    m.figure,
		Key:       plotcanvas.NormalizeKey(ev.Key),
		Modifiers: modifiersFromJS(ev.Modifiers),
	})
}

// handleToolbarButton maps the upstream tool name onto our
// ToolbarController action set and triggers it.
func (m *Manager) handleToolbarButton(name string) error {
	action, ok := toolbarActionByName(name)
	if !ok {
		return nil
	}
	if err := m.tb.Trigger(action); err != nil {
		return err
	}
	// Mirror the mode toggle back to clients so their toolbar UI can
	// stay in sync.
	if action == plotcanvas.ToolbarPan || action == plotcanvas.ToolbarZoom {
		mode := navigateModeName(m.tb.Mode())
		m.hub.broadcastJSON(&outboundEvent{Type: MsgNavigateMode, Mode: mode})
	}
	return nil
}

// handleRefresh re-sends the initial handshake to every client and
// forces a full frame on the next paint.
func (m *Manager) handleRefresh() error {
	m.mu.Lock()
	m.forceFull = true
	label := m.windowTitle
	m.mu.Unlock()
	m.hub.broadcastJSON(&outboundEvent{Type: MsgFigureLabel, Label: label})
	return m.DrawIdle()
}

func (m *Manager) handleSetDPR(ratio float64) error {
	if ratio <= 0 {
		return nil
	}
	m.mu.Lock()
	if m.devPxRatio == ratio {
		m.mu.Unlock()
		return nil
	}
	m.devPxRatio = ratio
	m.forceFull = true
	m.mu.Unlock()
	return m.DrawIdle()
}

// toolbarActionByName maps the upstream tool name to our action enum.
// Names match NavigationToolbar2.toolitems entries.
func toolbarActionByName(name string) (plotcanvas.ToolbarAction, bool) {
	switch name {
	case "home":
		return plotcanvas.ToolbarHome, true
	case "back":
		return plotcanvas.ToolbarBack, true
	case "forward":
		return plotcanvas.ToolbarForward, true
	case "pan":
		return plotcanvas.ToolbarPan, true
	case "zoom":
		return plotcanvas.ToolbarZoom, true
	case "download", "save_figure", "save":
		return plotcanvas.ToolbarSave, true
	case "configure_subplots":
		return plotcanvas.ToolbarConfigure, true
	}
	return "", false
}

// navigateModeName mirrors the strings upstream's NavigationToolbar2
// emits ("PAN", "ZOOM", or "" for none).
func navigateModeName(mode plotcanvas.ToolbarMode) string {
	switch mode {
	case plotcanvas.ToolbarModePan:
		return "PAN"
	case plotcanvas.ToolbarModeZoom:
		return "ZOOM"
	}
	return ""
}

// mouseButtonFromJS maps the JS button index (left=0, middle=1,
// right=2) to our canvas-side enum.
func mouseButtonFromJS(b int) plotcanvas.MouseButton {
	return plotcanvas.MouseButtonFromJSIndex(b)
}

// modifiersFromJS converts the upstream string list (["ctrl", "shift",
// "alt", "meta"]) to our bitfield. Case is normalized to lower.
func modifiersFromJS(mods []string) plotcanvas.Modifier {
	return plotcanvas.ModifiersFromNames(mods)
}

// normalizeKey strips the upstream "ctrl+", "shift+", "alt+" prefixes
// (which our dispatcher carries via Modifiers) so the dispatched Key
// field holds just the key value.
func normalizeKey(key string) string {
	return plotcanvas.NormalizeKey(key)
}
