// Package diag exposes matplotlib-go's diagnostic-warning hook to consumers.
//
// matplotlib-go surfaces non-fatal problems (unknown colormap, ignored or
// unhonored rcParam, silent capability downgrades) as warnings routed through
// a single process-global handler; the default handler writes each message to
// the standard library log prefixed with "matplotlib-go: ". Consumers use
// SetHandler to capture, filter, redirect, or silence those diagnostics.
package diag

import "github.com/cwbudde/matplotlib-go/internal/diag"

// SetHandler installs fn as the sink for all matplotlib-go diagnostic
// warnings and returns a function that restores the previously installed
// handler. A nil fn silences warnings entirely.
//
// The handler is process-global and shared by every package in the library,
// so a single call affects all matplotlib-go diagnostics. It is safe for
// concurrent use.
func SetHandler(fn func(string)) (restore func()) {
	return diag.SetHandler(fn)
}
