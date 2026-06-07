package animation

import (
	"sort"
	"strings"

	"github.com/cwbudde/matplotlib-go/canvas"
)

// MovieWriter mirrors matplotlib.animation.AbstractMovieWriter: a way to grab
// frames by calling GrabFrame. Setup is called to start the process and Finish
// is called afterwards. The Saving helper sequences Setup and Finish the way
// matplotlib's saving() context manager does.
//
// matplotlib's writer.setup takes a Figure whose .canvas is implicit; in this
// port the canvas.FigureCanvas carries both the figure and its renderer, so
// Setup takes the canvas and reads the figure from it.
type MovieWriter interface {
	// Setup prepares the writer to grab frames from c into outfile at the given
	// dpi. It mirrors AbstractMovieWriter.setup(fig, outfile, dpi).
	Setup(c canvas.FigureCanvas, outfile string, dpi float64) error
	// FrameSize returns the (width, height) in pixels of a movie frame, mirroring
	// AbstractMovieWriter.frame_size.
	FrameSize() (w, h int)
	// GrabFrame grabs the current rendered frame, mirroring grab_frame.
	GrabFrame() error
	// Finish completes the output file, mirroring finish.
	Finish() error
}

// Saving is the Go analogue of AbstractMovieWriter.saving(): it runs Setup, then
// body, and always runs Finish afterwards (like the context manager's finally).
func Saving(w MovieWriter, c canvas.FigureCanvas, outfile string, dpi float64, body func() error) (err error) {
	if err = w.Setup(c, outfile, dpi); err != nil {
		return err
	}
	defer func() {
		if ferr := w.Finish(); err == nil {
			err = ferr
		}
	}()
	return body()
}

// WriterFactory constructs a MovieWriter at the given frame rate, mirroring how
// matplotlib instantiates a registered writer class as writer_cls(fps).
type WriterFactory func(fps int) MovieWriter

// writerRegistry mirrors matplotlib.animation.MovieWriterRegistry / the module
// level writers registry. The dependency-free GIF and APNG writers are always
// registered; optional external encoders may register themselves from
// build-tagged files.
var writerRegistry = map[string]WriterFactory{}

// RegisterWriter adds a named writer factory, mirroring
// MovieWriterRegistry.register.
func RegisterWriter(name string, factory WriterFactory) {
	writerRegistry[strings.ToLower(name)] = factory
}

// RegisteredWriters lists the registered writer names, mirroring the iterable
// behavior of MovieWriterRegistry.
func RegisteredWriters() []string {
	names := make([]string, 0, len(writerRegistry))
	for name := range writerRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// WriterByName resolves a registered writer by name at the given frame rate.
// Unknown or unavailable names return ErrWriterUnsupported.
//
// This diverges from matplotlib, which silently falls back to PillowWriter when
// a named writer is unavailable; the port surfaces an explicit error instead.
func WriterByName(name string, fps int) (MovieWriter, error) {
	factory, ok := writerRegistry[strings.ToLower(name)]
	if !ok {
		return nil, ErrWriterUnsupported
	}
	return factory(fps), nil
}

func init() {
	RegisterWriter("pillow", func(fps int) MovieWriter { return NewGifWriter(fps) })
}
