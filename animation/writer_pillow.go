package animation

import (
	"errors"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"

	"github.com/cwbudde/matplotlib-go/canvas"
)

// defaultWriterFPS mirrors AbstractMovieWriter's default fps of 5.
const defaultWriterFPS = 5

// ErrNoFramesGrabbed is returned when Finish is called before any frame was
// grabbed, mirroring matplotlib indexing self._frames[0] on an empty list.
var ErrNoFramesGrabbed = errors.New("animation: no frames were grabbed")

// GifWriter is a deterministic, dependency-free GIF writer. It is the Go
// analogue of matplotlib.animation.PillowWriter: it accumulates frames during
// GrabFrame and encodes them into an animated GIF on Finish.
//
// matplotlib's PillowWriter grabs each frame with fig.savefig(format="rgba") and
// hands the frames to Pillow's save(save_all=True, append_images=..., loop=0).
// This port grabs the canvas RGBA buffer (canvas.RasterCanvas) and encodes with
// the standard library image/gif, quantizing against a fixed palette so the
// output is byte-for-byte deterministic across platforms.
type GifWriter struct {
	FPS int

	cnv    canvas.FigureCanvas
	out    string
	dpi    float64
	w, h   int
	frames []*image.RGBA // mirrors self._frames
}

// NewGifWriter returns a GifWriter at the given frame rate. fps <= 0 falls
// back to the AbstractMovieWriter default of 5.
func NewGifWriter(fps int) *GifWriter {
	if fps <= 0 {
		fps = defaultWriterFPS
	}
	return &GifWriter{FPS: fps}
}

// PillowWriter is a compatibility alias for GifWriter. Matplotlib names the
// equivalent class PillowWriter, but this port uses the standard library
// image/gif encoder rather than Pillow.
type PillowWriter = GifWriter

// NewPillowWriter returns a GifWriter. It is kept for callers using the earlier
// Matplotlib-derived name.
func NewPillowWriter(fps int) *GifWriter { return NewGifWriter(fps) }

// Setup stores the canvas, output path, and dpi and resets the frame buffer,
// mirroring PillowWriter.setup.
func (p *GifWriter) Setup(c canvas.FigureCanvas, outfile string, dpi float64) error {
	if c == nil {
		return ErrNilCanvas
	}
	if _, ok := c.(canvas.RasterCanvas); !ok {
		return ErrWriterUnsupported
	}
	p.cnv = c
	p.out = outfile
	p.dpi = dpi
	p.frames = nil
	if fig := c.Figure(); fig != nil {
		p.w = int(fig.SizePx.X)
		p.h = int(fig.SizePx.Y)
	}
	return nil
}

// FrameSize returns the (width, height) in pixels of a movie frame, mirroring
// AbstractMovieWriter.frame_size.
func (p *GifWriter) FrameSize() (w, h int) {
	return p.w, p.h
}

// GrabFrame captures the canvas's most recently rendered RGBA buffer and appends
// a private copy, mirroring PillowWriter.grab_frame (which reads the figure's
// RGBA buffer and stores it).
func (p *GifWriter) GrabFrame() error {
	raster, ok := p.cnv.(canvas.RasterCanvas)
	if !ok {
		return ErrWriterUnsupported
	}
	src := raster.FrameRGBA()
	if src == nil {
		return errors.New("animation: canvas produced no RGBA frame to grab")
	}
	p.frames = append(p.frames, cloneRGBAImage(src))
	return nil
}

// Finish encodes the accumulated frames into an animated GIF at p.out, mirroring
// PillowWriter.finish (save_all=True, duration=int(1000/fps), loop=0).
func (p *GifWriter) Finish() (err error) {
	if len(p.frames) == 0 {
		return ErrNoFramesGrabbed
	}
	bounds := image.Rect(0, 0, p.frames[0].Bounds().Dx(), p.frames[0].Bounds().Dy())
	// GIF delay is in centiseconds; matplotlib uses duration=int(1000/fps) ms.
	delay := int(100.0 / float64(p.FPS))

	out := &gif.GIF{
		Image:     make([]*image.Paletted, 0, len(p.frames)),
		Delay:     make([]int, 0, len(p.frames)),
		LoopCount: 0, // 0 == loop forever, matching Pillow loop=0
	}
	for _, frame := range p.frames {
		paletted := image.NewPaletted(bounds, palette.Plan9)
		// draw.Src maps every pixel to its nearest palette entry with no
		// dithering, keeping the encode deterministic.
		draw.Draw(paletted, bounds, frame, frame.Bounds().Min, draw.Src)
		out.Image = append(out.Image, paletted)
		out.Delay = append(out.Delay, delay)
	}

	f, err := os.Create(p.out)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	return gif.EncodeAll(f, out)
}

// cloneRGBAImage returns a deep copy of img so later canvas redraws cannot mutate
// an already-grabbed frame.
func cloneRGBAImage(img *image.RGBA) *image.RGBA {
	clone := image.NewRGBA(img.Bounds())
	copy(clone.Pix, img.Pix)
	return clone
}

var _ MovieWriter = (*GifWriter)(nil)
