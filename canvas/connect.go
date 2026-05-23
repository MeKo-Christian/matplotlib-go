// Package canvas — convenience wrappers for the mpl_connect / mpl_disconnect
// callback API.

package canvas

// Connect installs a handler for the named event on the canvas. Returns the
// connection ID required by [Disconnect]. This is the Go-friendly equivalent
// of Matplotlib's FigureCanvasBase.mpl_connect.
//
// A nil canvas returns the zero ConnectionID, mirroring the no-op semantics
// the Dispatcher itself uses when handed a nil handler.
func Connect(c FigureCanvas, eventType EventType, handler Handler) ConnectionID {
	if c == nil {
		return 0
	}
	return c.Connect(eventType, handler)
}

// Disconnect removes a previously installed handler. Calling with the zero
// ConnectionID, or after the canvas was closed, is a no-op. Mirrors
// Matplotlib's FigureCanvasBase.mpl_disconnect.
func Disconnect(c FigureCanvas, id ConnectionID) {
	if c == nil || id == 0 {
		return
	}
	c.Disconnect(id)
}
