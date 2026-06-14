package canvas

import (
	"sync"

	"github.com/cwbudde/matplotlib-go/geom"
)

// AxesHoverTracker synthesizes axes enter / leave lifecycle events from a
// stream of figure-level mouse events.
type AxesHoverTracker struct {
	figure     *Figure
	dispatcher *Dispatcher
	mu         sync.Mutex
	current    *Axes
}

// NewAxesHoverTracker creates a tracker bound to one figure and dispatcher.
func NewAxesHoverTracker(fig *Figure, dispatcher *Dispatcher) *AxesHoverTracker {
	return &AxesHoverTracker{figure: fig, dispatcher: dispatcher}
}

// Update observes a mouse lifecycle event and emits axes enter / leave events
// when the resolved axes under the cursor changes.
func (t *AxesHoverTracker) Update(event Event) {
	if t == nil || t.dispatcher == nil {
		return
	}
	fig := event.Figure
	if fig == nil {
		fig = t.figure
	}
	var next *Axes
	dataPosition := event.DataPosition
	hasDataPosition := event.HasDataPosition
	if event.Type != EventFigureLeave {
		next, dataPosition, hasDataPosition = ResolveEventTarget(fig, event.Position)
	}

	t.mu.Lock()
	current := t.current
	if next == current {
		t.mu.Unlock()
		return
	}
	t.current = next
	t.mu.Unlock()

	if current != nil {
		leave := event
		leave.Type = EventAxesLeave
		leave.Figure = fig
		leave.Axes = current
		leave.HasDataPosition = false
		leave.DataPosition = geom.Pt{}
		_ = t.dispatcher.Emit(leave)
	}
	if next != nil {
		enter := event
		enter.Type = EventAxesEnter
		enter.Figure = fig
		enter.Axes = next
		enter.DataPosition = dataPosition
		enter.HasDataPosition = hasDataPosition
		_ = t.dispatcher.Emit(enter)
	}
}
