// Package animation provides Matplotlib-style FuncAnimation and
// ArtistAnimation drivers on top of the canvas.FigureCanvas and event-loop
// scheduling primitives from package canvas.
//
// The animation engine is deterministic: tests can call Step to advance one
// frame without spinning up real timers. Interactive use cases supply a
// canvas.EventLoop, and the engine starts a timer that fires at the
// configured interval.
//
// Artists marked with SetAnimated(true) are skipped by the figure's default
// draw pass and drawn only on the animation's overlay pass. When a canvas
// reports both BlitCanvas and BufferRegioner the animation captures a
// background snapshot during Start, then on each frame restores the snapshot
// and replays only the animated artists before blitting the affected region.
// Backends without blit support fall back to a full canvas redraw per frame.
//
// GIF and APNG output are dependency-free through GifWriter and APNGWriter.
// MP4 and WebM output are optional: builds compiled with -tags ffmpeg register
// ffmpeg-backed writers that stream raw RGBA frames to an ffmpeg executable
// discovered at runtime.
package animation
