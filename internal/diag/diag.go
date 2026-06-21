// Package diag provides a small, shared diagnostic-warning facility so that
// library code can surface non-fatal problems (unknown colormap, silent
// capability downgrades, ignored input) instead of failing silently.
//
// Matplotlib emits these through Python's logging/warnings machinery; the Go
// port routes them through a single swappable handler. It lives in internal/
// so both low-level packages (color) and high-level packages (core) can use it
// without coupling to each other and without enlarging the public API surface.
package diag

import (
	"fmt"
	"log"
	"sync"
)

var (
	mu      sync.RWMutex
	handler = defaultHandler
)

func defaultHandler(msg string) { log.Printf("matplotlib-go: %s", msg) }

// SetHandler installs fn as the warning sink and returns a function that
// restores the previously installed handler. A nil fn silences warnings.
// It is safe for concurrent use; tests use it to capture or mute output.
func SetHandler(fn func(string)) (restore func()) {
	if fn == nil {
		fn = func(string) {}
	}
	mu.Lock()
	prev := handler
	handler = fn
	mu.Unlock()
	return func() {
		mu.Lock()
		handler = prev
		mu.Unlock()
	}
}

// Warnf formats and emits a non-fatal warning through the active handler.
func Warnf(format string, args ...any) {
	mu.RLock()
	h := handler
	mu.RUnlock()
	h(fmt.Sprintf(format, args...))
}
