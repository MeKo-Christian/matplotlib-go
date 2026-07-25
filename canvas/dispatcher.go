package canvas

import (
	"sync"
	"sync/atomic"
)

type handlerEntry struct {
	id      ConnectionID
	handler Handler
}

// Dispatcher manages event subscribers for a figure canvas.
type Dispatcher struct {
	next     atomic.Int64
	mu       sync.RWMutex
	handlers map[EventType][]handlerEntry
	index    map[ConnectionID]EventType
}

// Connect registers a handler for one event type.
func (d *Dispatcher) Connect(eventType EventType, handler Handler) ConnectionID {
	if handler == nil {
		return 0
	}

	id := nextConnectionID(&d.next)
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.handlers == nil {
		d.handlers = make(map[EventType][]handlerEntry)
	}
	if d.index == nil {
		d.index = make(map[ConnectionID]EventType)
	}
	d.handlers[eventType] = append(d.handlers[eventType], handlerEntry{id: id, handler: handler})
	d.index[id] = eventType
	return id
}

// Disconnect removes a registered handler.
func (d *Dispatcher) Disconnect(id ConnectionID) {
	if id == 0 {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	eventType, ok := d.index[id]
	if !ok {
		return
	}
	handlers := d.handlers[eventType]
	for i, entry := range handlers {
		if entry.id != id {
			continue
		}
		d.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
		if len(d.handlers[eventType]) == 0 {
			delete(d.handlers, eventType)
		}
		delete(d.index, id)
		return
	}
}

// Emit dispatches an event to all handlers registered for its type.
//
//nolint:gocritic // Event dispatch snapshots the caller's value by design.
func (d *Dispatcher) Emit(event Event) error {
	d.mu.RLock()
	handlers := make([]Handler, 0, len(d.handlers[event.Type]))
	for _, entry := range d.handlers[event.Type] {
		handlers = append(handlers, entry.handler)
	}
	d.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			return err
		}
	}
	return nil
}
