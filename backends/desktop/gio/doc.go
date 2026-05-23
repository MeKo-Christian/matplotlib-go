// Package gio is the Gio (gioui.org) adapter for the desktop interactive
// backend selected in docs/adr/0001-desktop-interactive-backend.md.
//
// Status: scaffolded. The actual window / event-loop integration lands
// in a follow-up PR for PLAN.md §4.2. Once that PR adds gioui.org to
// go.mod, this package will:
//
//   - import gioui.org/app, gioui.org/io/pointer, gioui.org/io/key,
//     gioui.org/op/paint;
//   - open one app.Window per figure;
//   - map pointer.Event / key.Event / app.FrameEvent into canvas
//     events through the figure's canvas.Dispatcher;
//   - blit the AGG renderer's *image.RGBA into the window via
//     paint.NewImageOp;
//   - host a canvas.ToolbarController-backed NavigationToolbar widget
//     above the canvas;
//   - register itself with desktop.Register from an init function so
//     callers of desktop.New get the Gio backend automatically.
//
// Until then, importing this package is a no-op and desktop.New returns
// desktop.ErrNotImplemented.
package gio
