// Package canvas provides the interactive layer that sits between a [Figure]
// and a windowing or browser host: event delivery, hit testing, and the
// figure-canvas/figure-manager contracts that GUI backends implement.
//
// [Figure] and [Axes] are aliases for the core types, so a canvas works
// directly with the object model from package core. A [Dispatcher] routes
// [Event] values — draw, resize, close, mouse, key, and pick events — to
// [Handler] callbacks registered with [Dispatcher.Connect] and removed with
// [Dispatcher.Disconnect]. [ResolveEventTarget] maps a pixel position to the
// [Axes] under the cursor for picking.
//
// [FigureCanvas] and [FigureManager] are the interfaces a GUI or web backend
// implements to drive rendering and window lifecycle; [DrawFigure] performs a
// plain render pass of a figure into any [render.Renderer].
package canvas
