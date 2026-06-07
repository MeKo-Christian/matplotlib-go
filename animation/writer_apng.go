package animation

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"math"
	"os"

	"github.com/cwbudde/matplotlib-go/canvas"
)

// APNGWriter is a deterministic, dependency-free APNG writer. It accumulates
// full RGBA frames during GrabFrame and encodes them as a full-frame animated
// PNG on Finish.
type APNGWriter struct {
	FPS int

	cnv    canvas.FigureCanvas
	out    string
	dpi    float64
	w, h   int
	frames []*image.RGBA
}

// NewAPNGWriter returns an APNG writer at the given frame rate. fps <= 0 falls
// back to the AbstractMovieWriter default of 5.
func NewAPNGWriter(fps int) *APNGWriter {
	if fps <= 0 {
		fps = defaultWriterFPS
	}
	return &APNGWriter{FPS: fps}
}

// Setup stores the canvas, output path, and dpi and resets the frame buffer.
func (w *APNGWriter) Setup(c canvas.FigureCanvas, outfile string, dpi float64) error {
	if c == nil {
		return ErrNilCanvas
	}
	if _, ok := c.(canvas.RasterCanvas); !ok {
		return ErrWriterUnsupported
	}
	w.cnv = c
	w.out = outfile
	w.dpi = dpi
	w.frames = nil
	if fig := c.Figure(); fig != nil {
		w.w = int(fig.SizePx.X)
		w.h = int(fig.SizePx.Y)
	}
	return nil
}

// FrameSize returns the (width, height) in pixels of a movie frame, mirroring
// AbstractMovieWriter.frame_size.
func (w *APNGWriter) FrameSize() (width, height int) {
	return w.w, w.h
}

// GrabFrame captures the canvas's most recently rendered RGBA buffer and
// appends a private copy.
func (w *APNGWriter) GrabFrame() error {
	raster, ok := w.cnv.(canvas.RasterCanvas)
	if !ok {
		return ErrWriterUnsupported
	}
	src := raster.FrameRGBA()
	if src == nil {
		return errors.New("animation: canvas produced no RGBA frame to grab")
	}
	w.frames = append(w.frames, cloneRGBAImage(src))
	return nil
}

// Finish encodes the accumulated frames into an animated PNG at w.out.
func (w *APNGWriter) Finish() (err error) {
	if len(w.frames) == 0 {
		return ErrNoFramesGrabbed
	}

	f, err := os.Create(w.out)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	return encodeAPNG(f, w.frames, w.FPS)
}

func encodeAPNG(out interface{ Write([]byte) (int, error) }, frames []*image.RGBA, fps int) error {
	firstBounds := frames[0].Bounds()
	width, height := firstBounds.Dx(), firstBounds.Dy()
	if width <= 0 || height <= 0 {
		return errors.New("animation: apng writer requires a positive frame size")
	}

	if _, err := out.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10}); err != nil {
		return err
	}
	if err := writePNGChunk(out, "IHDR", apngIHDR(width, height)); err != nil {
		return err
	}
	if err := writePNGChunk(out, "acTL", apngACTL(len(frames))); err != nil {
		return err
	}

	delayNum, delayDen := apngDelay(fps)
	seq := uint32(0)
	for i, frame := range frames {
		if b := frame.Bounds(); b.Dx() != width || b.Dy() != height {
			return errors.New("animation: apng frames must all have the same size")
		}
		if err := writePNGChunk(out, "fcTL", apngFCTL(seq, width, height, delayNum, delayDen)); err != nil {
			return err
		}
		seq++

		compressed, err := apngFrameData(frame)
		if err != nil {
			return err
		}
		if i == 0 {
			if err := writePNGChunk(out, "IDAT", compressed); err != nil {
				return err
			}
			continue
		}
		payload := make([]byte, 4+len(compressed))
		binary.BigEndian.PutUint32(payload[0:4], seq)
		copy(payload[4:], compressed)
		seq++
		if err := writePNGChunk(out, "fdAT", payload); err != nil {
			return err
		}
	}
	return writePNGChunk(out, "IEND", nil)
}

func apngIHDR(width, height int) []byte {
	data := make([]byte, 13)
	binary.BigEndian.PutUint32(data[0:4], uint32(width))
	binary.BigEndian.PutUint32(data[4:8], uint32(height))
	data[8] = 8 // bit depth
	data[9] = 6 // truecolor with alpha
	return data
}

func apngACTL(frameCount int) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], uint32(frameCount))
	// num_plays=0 means loop forever.
	return data
}

func apngFCTL(seq uint32, width, height int, delayNum, delayDen uint16) []byte {
	data := make([]byte, 26)
	binary.BigEndian.PutUint32(data[0:4], seq)
	binary.BigEndian.PutUint32(data[4:8], uint32(width))
	binary.BigEndian.PutUint32(data[8:12], uint32(height))
	binary.BigEndian.PutUint16(data[20:22], delayNum)
	binary.BigEndian.PutUint16(data[22:24], delayDen)
	// dispose_op=0 (none), blend_op=0 (source).
	return data
}

func apngDelay(fps int) (uint16, uint16) {
	if fps <= 0 {
		fps = defaultWriterFPS
	}
	if fps > math.MaxUint16 {
		fps = math.MaxUint16
	}
	return 1, uint16(fps)
}

func apngFrameData(img *image.RGBA) ([]byte, error) {
	b := img.Bounds()
	rowBytes := b.Dx() * 4
	var raw bytes.Buffer
	for y := b.Min.Y; y < b.Max.Y; y++ {
		raw.WriteByte(0) // PNG filter type 0: none.
		offset := img.PixOffset(b.Min.X, y)
		raw.Write(img.Pix[offset : offset+rowBytes])
	}

	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write(raw.Bytes()); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return compressed.Bytes(), nil
}

func writePNGChunk(out interface{ Write([]byte) (int, error) }, typ string, data []byte) error {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(data)))
	if _, err := out.Write(length[:]); err != nil {
		return err
	}
	typeBytes := []byte(typ)
	if _, err := out.Write(typeBytes); err != nil {
		return err
	}
	if len(data) > 0 {
		if _, err := out.Write(data); err != nil {
			return err
		}
	}

	crc := crc32.NewIEEE()
	_, _ = crc.Write(typeBytes)
	_, _ = crc.Write(data)
	var sum [4]byte
	binary.BigEndian.PutUint32(sum[:], crc.Sum32())
	_, err := out.Write(sum[:])
	return err
}

func init() {
	RegisterWriter("apng", func(fps int) MovieWriter { return NewAPNGWriter(fps) })
}

var _ MovieWriter = (*APNGWriter)(nil)
