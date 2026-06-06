//go:build ffmpeg

package animation

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"os/exec"

	"github.com/cwbudde/matplotlib-go/canvas"
)

const ffmpegExecutable = "ffmpeg"

// FFmpegAvailable reports whether the optional ffmpeg executable can be found
// on PATH. It is checked at runtime so binaries built with -tags ffmpeg still
// fail clearly on systems without ffmpeg installed.
func FFmpegAvailable() bool {
	_, err := exec.LookPath(ffmpegExecutable)
	return err == nil
}

// FFmpegWriter streams raw RGBA frames to an ffmpeg subprocess. It is available
// only in builds compiled with -tags ffmpeg and mirrors matplotlib's pipe-based
// FFMpegWriter rather than staging frame files.
type FFmpegWriter struct {
	FPS    int
	Format string

	cnv    canvas.FigureCanvas
	out    string
	dpi    float64
	w, h   int
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stderr bytes.Buffer
	frames int
}

// NewFFmpegWriter returns an MP4/H.264 ffmpeg writer at the given frame rate.
// fps <= 0 falls back to the AbstractMovieWriter default of 5.
func NewFFmpegWriter(fps int) *FFmpegWriter {
	return newFFmpegWriter(fps, "mp4")
}

// NewFFmpegWebMWriter returns a WebM/VP9 ffmpeg writer at the given frame rate.
// fps <= 0 falls back to the AbstractMovieWriter default of 5.
func NewFFmpegWebMWriter(fps int) *FFmpegWriter {
	return newFFmpegWriter(fps, "webm")
}

func newFFmpegWriter(fps int, format string) *FFmpegWriter {
	if fps <= 0 {
		fps = defaultWriterFPS
	}
	return &FFmpegWriter{FPS: fps, Format: format}
}

// Setup starts ffmpeg and prepares stdin for raw RGBA frames.
func (w *FFmpegWriter) Setup(c canvas.FigureCanvas, outfile string, dpi float64) error {
	if c == nil {
		return ErrNilCanvas
	}
	if _, ok := c.(canvas.RasterCanvas); !ok {
		return ErrWriterUnsupported
	}
	path, err := exec.LookPath(ffmpegExecutable)
	if err != nil {
		return fmt.Errorf("%w: %s executable not found", ErrWriterUnsupported, ffmpegExecutable)
	}

	width, height := 0, 0
	if fig := c.Figure(); fig != nil {
		width = int(fig.SizePx.X)
		height = int(fig.SizePx.Y)
	}
	if width <= 0 || height <= 0 {
		return errors.New("animation: ffmpeg writer requires a positive frame size")
	}

	w.cnv = c
	w.out = outfile
	w.dpi = dpi
	w.w = width
	w.h = height
	w.frames = 0
	w.stderr.Reset()

	cmd := exec.Command(path, w.args()...)
	cmd.Stderr = &w.stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return fmt.Errorf("animation: start ffmpeg: %w", err)
	}
	w.cmd = cmd
	w.stdin = stdin
	return nil
}

func (w *FFmpegWriter) args() []string {
	args := []string{
		"-y",
		"-f", "rawvideo",
		"-vcodec", "rawvideo",
		"-s", fmt.Sprintf("%dx%d", w.w, w.h),
		"-pix_fmt", "rgba",
		"-framerate", fmt.Sprintf("%d", w.FPS),
		"-i", "pipe:0",
		"-an",
	}
	switch w.Format {
	case "webm":
		args = append(args, "-vcodec", "libvpx-vp9", "-pix_fmt", "yuv420p")
	default:
		args = append(args, "-vcodec", "libx264", "-pix_fmt", "yuv420p")
	}
	return append(args, w.out)
}

// FrameSize returns the (width, height) in pixels of a movie frame, mirroring
// AbstractMovieWriter.frame_size.
func (w *FFmpegWriter) FrameSize() (width, height int) {
	return w.w, w.h
}

// GrabFrame writes the canvas's most recently rendered RGBA buffer to ffmpeg.
func (w *FFmpegWriter) GrabFrame() error {
	raster, ok := w.cnv.(canvas.RasterCanvas)
	if !ok {
		return ErrWriterUnsupported
	}
	if w.stdin == nil {
		return errors.New("animation: ffmpeg writer is not set up")
	}
	src := raster.FrameRGBA()
	if src == nil {
		return errors.New("animation: canvas produced no RGBA frame to grab")
	}
	if err := w.writeRGBA(src); err != nil {
		return err
	}
	w.frames++
	return nil
}

func (w *FFmpegWriter) writeRGBA(src *image.RGBA) error {
	b := src.Bounds()
	if b.Dx() != w.w || b.Dy() != w.h {
		return fmt.Errorf("animation: ffmpeg frame size is %dx%d, want %dx%d", b.Dx(), b.Dy(), w.w, w.h)
	}
	rowBytes := w.w * 4
	for y := b.Min.Y; y < b.Max.Y; y++ {
		offset := src.PixOffset(b.Min.X, y)
		if _, err := w.stdin.Write(src.Pix[offset : offset+rowBytes]); err != nil {
			return err
		}
	}
	return nil
}

// Finish closes ffmpeg stdin and waits for the encoder to finish.
func (w *FFmpegWriter) Finish() error {
	stdin := w.stdin
	cmd := w.cmd
	w.stdin = nil
	w.cmd = nil
	if stdin == nil || cmd == nil {
		if w.frames == 0 {
			return ErrNoFramesGrabbed
		}
		return nil
	}
	if w.frames == 0 {
		_ = stdin.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
		return ErrNoFramesGrabbed
	}
	if err := stdin.Close(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		if w.stderr.Len() > 0 {
			return fmt.Errorf("animation: ffmpeg failed: %w: %s", err, w.stderr.String())
		}
		return fmt.Errorf("animation: ffmpeg failed: %w", err)
	}
	return nil
}

func init() {
	RegisterWriter("ffmpeg", func(fps int) MovieWriter { return NewFFmpegWriter(fps) })
	RegisterWriter("ffmpeg-webm", func(fps int) MovieWriter { return NewFFmpegWebMWriter(fps) })
}

var _ MovieWriter = (*FFmpegWriter)(nil)
