// Package pyplot provides an optional stateful convenience layer on top of the
// explicit core Figure/Axes API.
//
// The package keeps the object-oriented API first-class by delegating directly
// to core.Figure and core.Axes methods. Its registry tracks the "current"
// figure and axes so migration-style helpers such as GCA, Title, Savefig, and
// Show can be used without replacing the underlying object model.
//
// By default Show performs a headless render through the selected backend so
// callers can validate that open figures draw successfully. Hosts that want a
// real window or browser integration can install a custom handler with
// SetShowHandler.
//
// Registry bookkeeping is synchronized, but pyplot's current figure and axes
// are process-wide mutable state. Concurrent callers can interfere logically,
// and returned core objects require their own ownership or synchronization.
// See docs/concurrency.md.
package pyplot
