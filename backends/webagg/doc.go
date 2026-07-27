// Package webagg implements an interactive web backend for matplotlib-go
// figures that mirrors the upstream Matplotlib `webagg` protocol.
//
// A [Manager] wraps a [core.Figure] together with an AGG (or compatible)
// renderer and tracks a per-client PNG diff state so the wire can carry
// either a full PNG frame or a sparse diff frame after the first paint.
// A [Server] exposes the manager over HTTP and WebSockets: JSON text
// frames carry events in both directions; binary frames carry the PNG
// payload from server to client.
//
// The protocol shape — message names, payload field names, and the
// "full vs diff" image-mode handshake — intentionally matches
// `third_party/matplotlib/lib/matplotlib/backends/backend_webagg_core.py`
// so existing tooling and upstream documentation transfer over.
//
// The package ships its own lightweight HTML / JavaScript / CSS client
// in `static/`. Hosts are free to point the [Server] at a different
// document root; the wire protocol does not depend on the bundled
// client.
package webagg
