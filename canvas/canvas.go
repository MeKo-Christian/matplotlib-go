package canvas

import (
	"sync/atomic"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

// Figure aliases the plotting figure type for runtime-facing APIs.
type Figure = core.Figure

// Axes aliases the plotting axes type for runtime-facing APIs.
type Axes = core.Axes

// EventType identifies a canvas runtime event.
type EventType string

const (
	EventDraw         EventType = "draw"
	EventResize       EventType = "resize"
	EventClose        EventType = "close"
	EventMousePress   EventType = "mouse_press"
	EventMouseRelease EventType = "mouse_release"
	EventMouseMove    EventType = "mouse_move"
	EventFigureEnter  EventType = "figure_enter"
	EventFigureLeave  EventType = "figure_leave"
	EventAxesEnter    EventType = "axes_enter"
	EventAxesLeave    EventType = "axes_leave"
	EventScroll       EventType = "scroll"
	EventKeyPress     EventType = "key_press"
	EventKeyRelease   EventType = "key_release"
	EventPick         EventType = "pick"
)

// MouseButton identifies a mouse button in a runtime event.
type MouseButton uint8

const (
	MouseButtonNone MouseButton = iota
	MouseButtonLeft
	MouseButtonMiddle
	MouseButtonRight
)

// Modifier tracks active key modifiers for an event.
type Modifier uint8

const (
	ModifierShift Modifier = 1 << iota
	ModifierControl
	ModifierAlt
	ModifierMeta
)

// Event carries normalized runtime input and lifecycle information.
type Event struct {
	Type            EventType
	Figure          *Figure
	Axes            *Axes
	Position        geom.Pt
	DataPosition    geom.Pt
	HasDataPosition bool
	Width           int
	Height          int
	Button          MouseButton
	Key             string
	Modifiers       Modifier
	DeltaX          float64
	DeltaY          float64
	DoubleClick     bool
	Native          any
}

// DrawEvent represents a completed draw lifecycle event.
type DrawEvent struct{ Event }

// ResizeEvent represents a figure canvas resize lifecycle event.
type ResizeEvent struct{ Event }

// CloseEvent represents a figure canvas close lifecycle event.
type CloseEvent struct{ Event }

// MouseEvent represents mouse press, release, move, and scroll events.
type MouseEvent struct{ Event }

// KeyEvent represents key press and release events.
type KeyEvent struct{ Event }

// PickEvent represents a picked artist. Backends may emit this after applying
// their hit-testing policy to a mouse event.
type PickEvent struct {
	Event
	Artist core.Artist
}

// NewDrawEvent creates a normalized draw event payload.
func NewDrawEvent(fig *Figure, width, height int) DrawEvent {
	return DrawEvent{Event: Event{Type: EventDraw, Figure: fig, Width: width, Height: height}}
}

// NewResizeEvent creates a normalized resize event payload.
func NewResizeEvent(fig *Figure, width, height int) ResizeEvent {
	return ResizeEvent{Event: Event{Type: EventResize, Figure: fig, Width: width, Height: height}}
}

// NewCloseEvent creates a normalized close event payload.
func NewCloseEvent(fig *Figure) CloseEvent {
	return CloseEvent{Event: Event{Type: EventClose, Figure: fig}}
}

// NewMouseEvent creates a normalized mouse event payload.
func NewMouseEvent(eventType EventType, fig *Figure, position geom.Pt, button MouseButton) MouseEvent {
	return MouseEvent{Event: Event{Type: eventType, Figure: fig, Position: position, Button: button}}
}

// NewKeyEvent creates a normalized key event payload.
func NewKeyEvent(eventType EventType, fig *Figure, key string, modifiers Modifier) KeyEvent {
	return KeyEvent{Event: Event{Type: eventType, Figure: fig, Key: key, Modifiers: modifiers}}
}

// NewPickEvent creates a normalized pick event payload.
//
//nolint:gocritic // Public event constructors intentionally preserve value semantics.
func NewPickEvent(fig *Figure, artist core.Artist, mouse MouseEvent) PickEvent {
	event := mouse.Event
	event.Type = EventPick
	event.Figure = fig
	return PickEvent{Event: event, Artist: artist}
}

// ConnectionID identifies a registered event handler.
type ConnectionID int64

// Handler is invoked for a normalized runtime event.
type Handler func(Event) error

// FigureCanvas exposes drawing, sizing, and event subscription for one figure.
type FigureCanvas interface {
	Figure() *Figure
	Draw() error
	Resize(width, height int) error
	Connect(EventType, Handler) ConnectionID
	Disconnect(ConnectionID)
	Close() error
}

// FigureManager exposes presentation and tooling for one figure canvas.
type FigureManager interface {
	Canvas() FigureCanvas
	Show() error
	Close() error
	SetTitle(string)
	ToolManager() *ToolManager
}

// ResolveEventTarget resolves the topmost axes under a figure-pixel position.
func ResolveEventTarget(fig *Figure, position geom.Pt) (*Axes, geom.Pt, bool) {
	if fig == nil {
		return nil, geom.Pt{}, false
	}

	for i := len(fig.Children) - 1; i >= 0; i-- {
		ax := fig.Children[i]
		if ax == nil || !ax.ContainsDisplayPoint(position) {
			continue
		}
		data, ok := ax.PixelToData(position)
		return ax, data, ok
	}

	return nil, geom.Pt{}, false
}

func nextConnectionID(counter *atomic.Int64) ConnectionID {
	return ConnectionID(counter.Add(1))
}
