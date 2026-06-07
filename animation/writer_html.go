package animation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"os"
	"strconv"

	"github.com/cwbudde/matplotlib-go/canvas"
)

// HTMLWriter is a deterministic, dependency-free writer for self-contained
// browser playback. It captures RGBA frames and writes a standalone HTML file
// with each frame embedded as a PNG data URI.
type HTMLWriter struct {
	FPS int

	cnv    canvas.FigureCanvas
	out    string
	dpi    float64
	w, h   int
	frames []*image.RGBA
}

// NewHTMLWriter returns an HTML writer at the given frame rate. fps <= 0 falls
// back to the AbstractMovieWriter default of 5.
func NewHTMLWriter(fps int) *HTMLWriter {
	if fps <= 0 {
		fps = defaultWriterFPS
	}
	return &HTMLWriter{FPS: fps}
}

// Setup stores the canvas, output path, and dpi and resets the frame buffer.
func (w *HTMLWriter) Setup(c canvas.FigureCanvas, outfile string, dpi float64) error {
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
func (w *HTMLWriter) FrameSize() (width, height int) {
	return w.w, w.h
}

// GrabFrame captures the canvas's most recently rendered RGBA buffer and
// appends a private copy.
func (w *HTMLWriter) GrabFrame() error {
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

// Finish writes the accumulated frames into a standalone HTML animation.
func (w *HTMLWriter) Finish() (err error) {
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
	return encodeHTMLAnimation(f, w.frames, w.FPS)
}

type htmlAnimationData struct {
	FPS        int      `json:"fps"`
	FrameDelay int      `json:"frameDelayMS"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	Frames     []string `json:"frames"`
}

func encodeHTMLAnimation(out interface{ Write([]byte) (int, error) }, frames []*image.RGBA, fps int) error {
	if fps <= 0 {
		fps = defaultWriterFPS
	}
	firstBounds := frames[0].Bounds()
	width, height := firstBounds.Dx(), firstBounds.Dy()
	if width <= 0 || height <= 0 {
		return errors.New("animation: html writer requires a positive frame size")
	}

	data := htmlAnimationData{
		FPS:        fps,
		FrameDelay: int(1000.0 / float64(fps)),
		Width:      width,
		Height:     height,
		Frames:     make([]string, 0, len(frames)),
	}
	for _, frame := range frames {
		if b := frame.Bounds(); b.Dx() != width || b.Dy() != height {
			return errors.New("animation: html frames must all have the same size")
		}
		var encoded bytes.Buffer
		if err := png.Encode(&encoded, frame); err != nil {
			return err
		}
		data.Frames = append(data.Frames, "data:image/png;base64,"+base64.StdEncoding.EncodeToString(encoded.Bytes()))
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = out.Write([]byte(htmlAnimationDocument(string(payload), width, height)))
	return err
}

func htmlAnimationDocument(payload string, width, height int) string {
	widthText := strconv.Itoa(width)
	heightText := strconv.Itoa(height)
	return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Animation</title>
<style>
:root { color-scheme: light dark; font-family: system-ui, sans-serif; }
body { margin: 0; min-height: 100vh; display: grid; place-items: center; background: Canvas; color: CanvasText; }
.animation { display: grid; gap: 12px; justify-items: center; padding: 16px; }
.viewport { width: min(100vw - 32px, ` + widthText + `px); }
img { display: block; width: 100%; height: auto; image-rendering: auto; }
.controls { display: grid; grid-template-columns: repeat(4, auto) minmax(140px, 1fr); align-items: center; gap: 8px; width: min(100vw - 32px, ` + widthText + `px); }
button { font: inherit; padding: 6px 10px; }
input[type="range"] { width: 100%; }
</style>
</head>
<body>
<main class="animation" data-width="` + widthText + `" data-height="` + heightText + `">
<div class="viewport"><img id="frame" alt="Animation frame"></div>
<div class="controls">
<button type="button" id="first" aria-label="First frame">|&lt;</button>
<button type="button" id="prev" aria-label="Previous frame">&lt;</button>
<button type="button" id="play" aria-label="Pause">Pause</button>
<button type="button" id="next" aria-label="Next frame">&gt;</button>
<input type="range" id="seek" min="0" value="0" step="1" aria-label="Frame">
</div>
</main>
<script type="application/json" id="animation-data">` + payload + `</script>
<script>
(() => {
  const data = JSON.parse(document.getElementById("animation-data").textContent);
  const image = document.getElementById("frame");
  const seek = document.getElementById("seek");
  const play = document.getElementById("play");
  let frame = 0;
  let timer = 0;
  seek.max = Math.max(0, data.frames.length - 1);
  function show(index) {
    frame = (index + data.frames.length) % data.frames.length;
    image.src = data.frames[frame];
    seek.value = String(frame);
  }
  function start() {
    stop();
    play.textContent = "Pause";
    play.setAttribute("aria-label", "Pause");
    timer = window.setInterval(() => show(frame + 1), data.frameDelayMS);
  }
  function stop() {
    if (timer) window.clearInterval(timer);
    timer = 0;
    play.textContent = "Play";
    play.setAttribute("aria-label", "Play");
  }
  document.getElementById("first").addEventListener("click", () => show(0));
  document.getElementById("prev").addEventListener("click", () => show(frame - 1));
  document.getElementById("next").addEventListener("click", () => show(frame + 1));
  play.addEventListener("click", () => timer ? stop() : start());
  seek.addEventListener("input", () => show(Number(seek.value)));
  show(0);
  if (data.frames.length > 1) start();
})();
</script>
</body>
</html>
`
}

var _ MovieWriter = (*HTMLWriter)(nil)
