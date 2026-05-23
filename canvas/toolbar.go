package canvas

import (
	"errors"
	"sync"
)

// ToolbarAction is the stable identifier of one entry in the standard
// navigation toolbar. The set mirrors Matplotlib's NavigationToolbar2: the
// four primary actions (home, pan, zoom-to-rect, save) plus the optional
// history and subplot-config actions. Backend implementations may show a
// subset.
type ToolbarAction string

const (
	// ToolbarHome resets every axes to the view limits captured at attach
	// time. Mirrors NavigationToolbar2.home.
	ToolbarHome ToolbarAction = "home"
	// ToolbarBack steps one entry back through the view history. Mirrors
	// NavigationToolbar2.back.
	ToolbarBack ToolbarAction = "back"
	// ToolbarForward steps one entry forward through the view history.
	// Mirrors NavigationToolbar2.forward.
	ToolbarForward ToolbarAction = "forward"
	// ToolbarPan toggles drag-to-pan. Mirrors NavigationToolbar2.pan.
	ToolbarPan ToolbarAction = "pan"
	// ToolbarZoom toggles rubber-band zoom-to-rect. Mirrors
	// NavigationToolbar2.zoom.
	ToolbarZoom ToolbarAction = "zoom"
	// ToolbarSave saves the current figure. Mirrors
	// NavigationToolbar2.save_figure.
	ToolbarSave ToolbarAction = "save_figure"
	// ToolbarConfigure opens the subplots-configuration panel. Mirrors
	// NavigationToolbar2.configure_subplots.
	ToolbarConfigure ToolbarAction = "configure_subplots"
)

// ToolbarMode enumerates the toggleable modes a NavigationToolbar can be in.
// At most one of pan / zoom is active; ToolbarModeNone means clicks fall
// through to the dispatcher without modification.
type ToolbarMode uint8

const (
	// ToolbarModeNone disables drag-driven view manipulation.
	ToolbarModeNone ToolbarMode = iota
	// ToolbarModePan activates drag-to-pan.
	ToolbarModePan
	// ToolbarModeZoom activates drag-to-rubber-band-zoom.
	ToolbarModeZoom
)

// NavigationToolbar is the backend-agnostic contract for the standard
// figure toolbar. A backend implementation owns the widget layout and
// rendering; this interface defines how the rest of matplotlib-go drives
// it.
//
// The contract is intentionally narrow so that Gio, a hypothetical Qt
// binding, and the web (WebAgg) backend can all satisfy it.
type NavigationToolbar interface {
	// Mode returns the active interaction mode.
	Mode() ToolbarMode
	// SetMode switches the active interaction mode. Implementations should
	// keep the underlying canvas.Navigation in sync.
	SetMode(ToolbarMode)
	// Trigger executes the named action as if the user clicked the button.
	Trigger(ToolbarAction) error
	// SetActionEnabled hides or disables an action in the UI. Disabled
	// actions still appear in the toolbar but do not respond to clicks.
	SetActionEnabled(ToolbarAction, bool)
	// ActionEnabled reports whether an action is currently enabled.
	ActionEnabled(ToolbarAction) bool
	// SetMessage shows a transient status message (typically the
	// hover-driven coordinate readout).
	SetMessage(string)
}

// ToolbarHandler runs the side-effect of a toolbar action — saving the
// figure, replaying history, opening the subplot configurator, etc. The
// pan/zoom handlers are wired by NewToolbarController and do not need to
// be supplied.
type ToolbarHandler func() error

// ToolbarController is the backend-agnostic implementation of
// NavigationToolbar. Concrete backends embed (or wrap) it and add the
// platform-specific widget rendering.
//
// Pan / zoom are wired through canvas.Navigation; other actions are
// dispatched via per-action handlers supplied by the host backend. Both
// halves are optional — a Save handler is meaningless until the host
// knows where to save, for example.
//
// The zero value is unusable; call NewToolbarController.
type ToolbarController struct {
	mu       sync.Mutex
	mode     ToolbarMode
	enabled  map[ToolbarAction]bool
	handlers map[ToolbarAction]ToolbarHandler
	message  string

	nav *Navigation
}

// NewToolbarController builds a controller bound to the given Navigation
// helper. Pass nil for nav to build a controller with no pan/zoom wiring
// (useful for tests).
func NewToolbarController(nav *Navigation) *ToolbarController {
	c := &ToolbarController{
		nav:      nav,
		enabled:  make(map[ToolbarAction]bool),
		handlers: make(map[ToolbarAction]ToolbarHandler),
	}
	for _, a := range standardToolbarActions() {
		c.enabled[a] = true
	}
	return c
}

// standardToolbarActions lists the actions a default toolbar exposes.
func standardToolbarActions() []ToolbarAction {
	return []ToolbarAction{
		ToolbarHome, ToolbarBack, ToolbarForward,
		ToolbarPan, ToolbarZoom,
		ToolbarSave, ToolbarConfigure,
	}
}

// SetHandler registers a custom side-effect handler for one action. Use
// this to wire ToolbarSave, ToolbarHome, ToolbarBack, ToolbarForward, or
// ToolbarConfigure to backend-specific behavior. Pan and zoom toggles
// are handled internally and ignore handlers registered here.
func (c *ToolbarController) SetHandler(action ToolbarAction, handler ToolbarHandler) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if handler == nil {
		delete(c.handlers, action)
		return
	}
	c.handlers[action] = handler
}

// Mode returns the active interaction mode.
func (c *ToolbarController) Mode() ToolbarMode {
	if c == nil {
		return ToolbarModeNone
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.mode
}

// SetMode switches the active interaction mode and forwards the change
// to the underlying Navigation helper.
func (c *ToolbarController) SetMode(mode ToolbarMode) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.mode = mode
	nav := c.nav
	c.mu.Unlock()
	if nav == nil {
		return
	}
	switch mode {
	case ToolbarModePan:
		nav.SetMode(NavPan)
	case ToolbarModeZoom:
		nav.SetMode(NavZoom)
	default:
		nav.SetMode(NavNone)
	}
}

// SetActionEnabled hides or disables an action.
func (c *ToolbarController) SetActionEnabled(action ToolbarAction, enabled bool) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled[action] = enabled
}

// ActionEnabled reports whether the named action is enabled.
func (c *ToolbarController) ActionEnabled(action ToolbarAction) bool {
	if c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.enabled[action]
	return ok && v
}

// SetMessage updates the transient status message. Backends typically
// render this as a strip below the figure.
func (c *ToolbarController) SetMessage(msg string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.message = msg
}

// Message returns the current status message.
func (c *ToolbarController) Message() string {
	if c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.message
}

// Trigger executes an action as if the user clicked its toolbar button.
// Pan and zoom toggle the active mode; all other actions invoke the
// handler registered via SetHandler. Triggering an action that has no
// registered handler returns nil — clicking an unwired button is not an
// error.
func (c *ToolbarController) Trigger(action ToolbarAction) error {
	if c == nil {
		return errors.New("canvas: nil toolbar controller")
	}
	if !c.ActionEnabled(action) {
		return nil
	}
	switch action {
	case ToolbarPan:
		if c.Mode() == ToolbarModePan {
			c.SetMode(ToolbarModeNone)
		} else {
			c.SetMode(ToolbarModePan)
		}
		return nil
	case ToolbarZoom:
		if c.Mode() == ToolbarModeZoom {
			c.SetMode(ToolbarModeNone)
		} else {
			c.SetMode(ToolbarModeZoom)
		}
		return nil
	}
	c.mu.Lock()
	handler := c.handlers[action]
	c.mu.Unlock()
	if handler == nil {
		return nil
	}
	return handler()
}
