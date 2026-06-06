package animation

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// DefaultInterval matches Matplotlib's FuncAnimation default frame interval.
const DefaultInterval = 200 * time.Millisecond

// Errors returned by the animation engine.
var (
	ErrNilCanvas         = errors.New("animation: canvas is required")
	ErrNilUpdate         = errors.New("animation: update function is required")
	ErrAlreadyRunning    = errors.New("animation: already running")
	ErrNotRunning        = errors.New("animation: not running")
	ErrNoFramesToRun     = errors.New("animation: no frames remain to step")
	ErrEmptyFrameSet     = errors.New("animation: artist animation needs at least one frame")
	ErrWriterUnsupported = errors.New("animation: movie writers are not supported")
)

// UpdateFunc returns the animated artists for the given frame index. Artists
// not previously marked animated will have SetAnimated(true) applied
// automatically so the figure's default draw skips them. Returning nil is
// allowed; the engine will redraw whatever is currently marked animated.
type UpdateFunc func(frame int) ([]core.Artist, error)

// InitFunc is called once before frame 0 and returns the initial set of
// animated artists. It may be nil.
type InitFunc func() ([]core.Artist, error)

// Config carries the shared animation parameters used by FuncAnimation and
// ArtistAnimation. The zero value of Interval is replaced with DefaultInterval.
type Config struct {
	Canvas      canvas.FigureCanvas
	Frames      int
	Interval    time.Duration
	Repeat      bool
	RepeatDelay time.Duration
	Blit        bool
	EventLoop   canvas.EventLoop
}

func (c Config) normalize() Config {
	if c.Interval <= 0 {
		c.Interval = DefaultInterval
	}
	if c.EventLoop == nil {
		c.EventLoop = canvas.SynchronousEventLoop{}
	}
	if c.Frames < 0 {
		c.Frames = 0
	}
	return c
}

// Animation drives a frame loop on a FigureCanvas. The same struct backs both
// FuncAnimation and ArtistAnimation.
type Animation struct {
	cfg    Config
	init   InitFunc
	update UpdateFunc

	mu          sync.Mutex
	frame       int
	totalFrames int // running count of frames produced
	running     bool
	timer       canvas.Timer
	// awaitingRepeat is true between the last frame of a finite run and the
	// next repeat cycle. It exists so RepeatDelay can be honored.
	awaitingRepeat bool

	// background snapshot for the blit fast path. Populated lazily on the
	// first frame when Blit is enabled and the canvas supports both
	// BlitCanvas and a BufferRegioner-aware drawing path.
	background *render.BufferRegion
	blitReady  bool

	// animatedArtists tracks artists registered as animated by InitFunc or
	// any UpdateFunc invocation. This is used for the blit overlay pass and
	// to know which artists to keep marked SetAnimated(true).
	animatedArtists []core.Artist
}

// NewFuncAnimation returns an Animation that calls update on each frame.
// init, when non-nil, is invoked once before frame 0.
func NewFuncAnimation(cfg Config, update UpdateFunc, init InitFunc) (*Animation, error) {
	if cfg.Canvas == nil {
		return nil, ErrNilCanvas
	}
	if update == nil {
		return nil, ErrNilUpdate
	}
	return &Animation{
		cfg:    cfg.normalize(),
		init:   init,
		update: update,
	}, nil
}

// NewArtistAnimation returns an Animation that shows frames[i] on frame i by
// toggling artist visibility. All artists in any frame are marked animated so
// the figure default-draw never picks them up; only the engine's overlay pass
// shows them. The engine sets the artists in frames[i] visible and the rest
// hidden before each draw.
func NewArtistAnimation(cfg Config, frames [][]core.Artist) (*Animation, error) {
	if cfg.Canvas == nil {
		return nil, ErrNilCanvas
	}
	if len(frames) == 0 {
		return nil, ErrEmptyFrameSet
	}
	// Build the union of every artist across frames; mark all animated and
	// hidden up front. The update function picks the visible set per frame.
	union := uniqueArtists(frames)
	for _, art := range union {
		setAnimated(art, true)
		setVisible(art, false)
	}

	frameSets := make([][]core.Artist, len(frames))
	for i, set := range frames {
		clone := make([]core.Artist, len(set))
		copy(clone, set)
		frameSets[i] = clone
	}

	cfg.Frames = len(frameSets)
	update := func(frame int) ([]core.Artist, error) {
		for _, art := range union {
			setVisible(art, false)
		}
		visible := frameSets[frame]
		for _, art := range visible {
			setVisible(art, true)
		}
		return union, nil
	}
	init := func() ([]core.Artist, error) {
		return union, nil
	}
	return NewFuncAnimation(cfg, update, init)
}

// Frame returns the index of the next frame Step will produce.
func (a *Animation) Frame() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.frame
}

// TotalFramesDrawn returns the running count of frames produced across all
// Step / Start cycles. It is intended for tests and observability.
func (a *Animation) TotalFramesDrawn() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.totalFrames
}

// Running reports whether a timer-driven loop is currently active.
func (a *Animation) Running() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running
}

// Config returns the normalized configuration in effect.
func (a *Animation) Config() Config {
	return a.cfg
}

// SaveOption configures Save, mirroring the writer/fps/dpi/save_count keyword
// arguments of matplotlib.animation.Animation.save.
type SaveOption func(*saveConfig)

type saveConfig struct {
	writer    MovieWriter
	writerSet bool
	fps       int
	dpi       float64
	saveCount int
}

// WithWriter saves with an explicit writer instead of selecting one by filename
// extension, mirroring save(writer=...).
func WithWriter(w MovieWriter) SaveOption {
	return func(c *saveConfig) { c.writer = w; c.writerSet = true }
}

// WithFPS sets the saved frame rate, mirroring save(fps=...). When unset, the
// rate is derived from the animation interval as 1000/interval(ms).
func WithFPS(fps int) SaveOption { return func(c *saveConfig) { c.fps = fps } }

// WithDPI sets the dots-per-inch passed to the writer, mirroring save(dpi=...).
func WithDPI(dpi float64) SaveOption { return func(c *saveConfig) { c.dpi = dpi } }

// WithSaveCount sets how many frames to write, mirroring FuncAnimation's
// save_count. When unset it defaults to the configured Frames count. The Go port
// has no cache_frame_data knob because Save regenerates every frame
// deterministically from the update function.
func WithSaveCount(n int) SaveOption { return func(c *saveConfig) { c.saveCount = n } }

// Save draws every frame and writes them to filename, mirroring
// matplotlib.animation.Animation.save.
//
// The writer is chosen from the filename extension (".gif" -> GifWriter) unless
// one is supplied with WithWriter. Extensions and writer names that require
// external encoders (mp4/ffmpeg/imagemagick) or HTML output return
// ErrWriterUnsupported.
func (a *Animation) Save(filename string, opts ...SaveOption) error {
	cfg := saveConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	// fps defaults to 1000/interval(ms), matching save()'s fps fallback.
	fps := cfg.fps
	if fps <= 0 {
		fps = int(float64(time.Second)/float64(a.cfg.Interval) + 0.5)
		if fps < 1 {
			fps = 1
		}
	}

	writer := cfg.writer
	if !cfg.writerSet {
		name, ok := writerNameForFile(filename)
		if !ok {
			return ErrWriterUnsupported
		}
		w, err := WriterByName(name, fps)
		if err != nil {
			return err
		}
		writer = w
	}
	if writer == nil {
		return ErrWriterUnsupported
	}

	// save_count defaults to the configured frame count.
	saveCount := cfg.saveCount
	if saveCount <= 0 {
		saveCount = a.cfg.Frames
	}
	if saveCount <= 0 {
		return ErrNoFramesToRun
	}

	return Saving(writer, a.cfg.Canvas, filename, cfg.dpi, func() error {
		for frame := 0; frame < saveCount; frame++ {
			if err := a.drawSaveFrame(frame, frame == 0); err != nil {
				return err
			}
			if err := writer.GrabFrame(); err != nil {
				return err
			}
		}
		return nil
	})
}

// drawSaveFrame renders one frame for Save without blitting, mirroring
// Animation._draw_next_frame(data, blit=False): run the one-time init pass, call
// the update function, mark the produced artists animated, and draw the figure.
func (a *Animation) drawSaveFrame(frame int, first bool) error {
	if first && a.init != nil {
		initial, err := a.init()
		if err != nil {
			return err
		}
		a.registerAnimated(initial)
	}
	updated, err := a.update(frame)
	if err != nil {
		return err
	}
	a.registerAnimated(updated)
	return a.cfg.Canvas.Draw()
}

// writerNameForFile maps a filename extension to a registered writer name. Only
// ".gif" is supported; every other extension is an intentional omission.
func writerNameForFile(filename string) (string, bool) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".gif":
		return "pillow", true
	default:
		return "", false
	}
}

// Step advances the animation by one frame deterministically. It performs the
// init pass on first invocation (if configured), invokes the update function,
// and draws the result through the canvas. Step is safe to call without
// Start and is the path used by tests and frame writers.
//
// Step returns the frame index just rendered. When the configured Frames count
// is reached and Repeat is false, Step returns ErrNoFramesToRun and does not
// draw.
func (a *Animation) Step() (int, error) {
	a.mu.Lock()
	first := a.totalFrames == 0
	if a.cfg.Frames > 0 && a.frame >= a.cfg.Frames {
		if !a.cfg.Repeat {
			a.mu.Unlock()
			return -1, ErrNoFramesToRun
		}
		a.frame = 0
	}
	frame := a.frame
	a.mu.Unlock()

	if first && a.init != nil {
		initial, err := a.init()
		if err != nil {
			return -1, err
		}
		a.registerAnimated(initial)
	}

	updated, err := a.update(frame)
	if err != nil {
		return -1, err
	}
	a.registerAnimated(updated)

	if err := a.drawFrame(first); err != nil {
		return -1, err
	}

	a.mu.Lock()
	a.frame++
	a.totalFrames++
	a.mu.Unlock()
	return frame, nil
}

// Start begins driving frames on the configured EventLoop. It returns
// immediately; frames fire on the loop's timer until Frames is reached
// (when Repeat is false) or Stop is called.
func (a *Animation) Start() error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return ErrAlreadyRunning
	}
	a.running = true
	interval := a.cfg.Interval
	loop := a.cfg.EventLoop
	a.mu.Unlock()

	timer := loop.NewTimer(interval, a.onTimerTick)
	a.mu.Lock()
	a.timer = timer
	a.mu.Unlock()
	if err := timer.Start(); err != nil {
		a.mu.Lock()
		if a.timer == timer {
			a.timer = nil
			a.running = false
			a.awaitingRepeat = false
		}
		a.mu.Unlock()
		_ = timer.Stop()
		return err
	}
	return nil
}

// Stop halts a running animation. It is a no-op when not running.
func (a *Animation) Stop() error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return nil
	}
	a.running = false
	timer := a.timer
	a.timer = nil
	a.awaitingRepeat = false
	a.mu.Unlock()
	if timer == nil {
		return nil
	}
	return timer.Stop()
}

func (a *Animation) onTimerTick() error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return nil
	}
	repeatDelay := a.cfg.RepeatDelay
	awaitingRepeat := a.awaitingRepeat
	a.awaitingRepeat = false
	a.mu.Unlock()

	// Honor RepeatDelay by inserting a single "skipped" tick at the
	// boundary between the last frame and the next repeat cycle. Tests
	// rely on this being deterministic, so we model it as a one-shot skip
	// when the previous tick ended a finite run.
	if awaitingRepeat && repeatDelay > 0 {
		// reschedule timer with the repeat delay, then resume normal interval.
		return a.rescheduleAfterRepeatDelay(repeatDelay)
	}

	_, err := a.Step()
	if err != nil {
		// End of a non-repeating run; stop the timer.
		if errors.Is(err, ErrNoFramesToRun) {
			return a.Stop()
		}
		return err
	}

	a.mu.Lock()
	finiteDone := a.cfg.Frames > 0 && a.frame >= a.cfg.Frames
	if finiteDone {
		if a.cfg.Repeat {
			a.awaitingRepeat = repeatDelay > 0
			a.frame = 0
		} else {
			a.running = false
			t := a.timer
			a.timer = nil
			a.mu.Unlock()
			if t != nil {
				return t.Stop()
			}
			return nil
		}
	}
	a.mu.Unlock()
	return nil
}

func (a *Animation) rescheduleAfterRepeatDelay(delay time.Duration) error {
	a.mu.Lock()
	loop := a.cfg.EventLoop
	interval := a.cfg.Interval
	current := a.timer
	a.mu.Unlock()

	if current != nil {
		_ = current.Stop()
	}

	delayTimer := loop.NewTimer(delay, func() error {
		// fire one tick after the delay, then resume the interval timer.
		if err := a.fireSingleFrame(); err != nil {
			return err
		}
		next := loop.NewTimer(interval, a.onTimerTick)
		a.mu.Lock()
		a.timer = next
		a.mu.Unlock()
		if err := next.Start(); err != nil {
			a.mu.Lock()
			if a.timer == next {
				a.timer = nil
				a.running = false
				a.awaitingRepeat = false
			}
			a.mu.Unlock()
			_ = next.Stop()
			return err
		}
		return nil
	})

	a.mu.Lock()
	a.timer = delayTimer
	a.mu.Unlock()
	return delayTimer.Start()
}

func (a *Animation) fireSingleFrame() error {
	_, err := a.Step()
	if errors.Is(err, ErrNoFramesToRun) {
		return nil
	}
	return err
}

func (a *Animation) registerAnimated(artists []core.Artist) {
	if len(artists) == 0 {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, art := range artists {
		if art == nil {
			continue
		}
		setAnimated(art, true)
		if !containsArtist(a.animatedArtists, art) {
			a.animatedArtists = append(a.animatedArtists, art)
		}
	}
}

func (a *Animation) drawFrame(first bool) error {
	a.mu.Lock()
	useBlit := a.cfg.Blit
	cnv := a.cfg.Canvas
	a.mu.Unlock()

	// Try the blit fast path: requires BlitCanvas + a renderer that
	// implements BufferRegioner. For simplicity in Phase 6.1 we trigger
	// a full canvas Draw when blit is disabled or unsupported.
	if useBlit {
		if done, err := a.drawBlit(cnv, first); done {
			return err
		}
	}
	return cnv.Draw()
}

// drawBlit returns done=true when the blit fast path handled the frame.
// done=false means the caller should fall back to a regular Draw.
func (a *Animation) drawBlit(cnv canvas.FigureCanvas, first bool) (bool, error) {
	blitter, ok := cnv.(canvas.BlitCanvas)
	if !ok {
		return false, nil
	}
	// Capture a background snapshot on the first frame by drawing the
	// figure once with animated artists suppressed, then copying the full
	// figure rect into a BufferRegion. After that, every frame restores
	// the snapshot and draws only the animated overlay before blitting.
	if first || !a.blitReady {
		if err := cnv.Draw(); err != nil {
			return true, err
		}
		fig := cnv.Figure()
		if fig == nil {
			return false, nil
		}
		bbox := geom.Rect{Max: geom.Pt{X: fig.SizePx.X, Y: fig.SizePx.Y}}
		bg := blitter.CopyFromBBox(bbox)
		if bg == nil {
			// renderer does not support buffer regions; fall back.
			return false, nil
		}
		a.mu.Lock()
		a.background = bg
		a.blitReady = true
		a.mu.Unlock()
		return true, blitter.Blit(bbox)
	}
	// Restore background then overlay animated artists by drawing the
	// figure with AnimatedFilterOnlyAnimated through whichever Draw entry
	// point we have. canvas.Draw is the public entry — for Phase 6.1 we
	// rely on the canvas's regular Draw which redraws everything, then
	// composite. Backends that want a true overlay pass must opt in via
	// their own animation-aware Draw; until then, this is correctness with
	// the blit overhead of a full redraw, which is still useful as the API
	// surface and lets tests observe blit decisions.
	fig := cnv.Figure()
	bbox := geom.Rect{Max: geom.Pt{X: fig.SizePx.X, Y: fig.SizePx.Y}}
	a.mu.Lock()
	bg := a.background
	a.mu.Unlock()
	if bg != nil {
		blitter.RestoreRegion(bg, &bbox, geom.Pt{})
	}
	if err := cnv.Draw(); err != nil {
		return true, err
	}
	return true, blitter.Blit(bbox)
}

func containsArtist(list []core.Artist, target core.Artist) bool {
	for _, a := range list {
		if a == target {
			return true
		}
	}
	return false
}

func uniqueArtists(frames [][]core.Artist) []core.Artist {
	seen := make(map[core.Artist]struct{})
	var out []core.Artist
	for _, set := range frames {
		for _, art := range set {
			if art == nil {
				continue
			}
			if _, ok := seen[art]; ok {
				continue
			}
			seen[art] = struct{}{}
			out = append(out, art)
		}
	}
	return out
}

type animatedSetter interface {
	SetAnimated(bool)
}

type visibleSetter interface {
	SetVisible(bool)
}

func setAnimated(art core.Artist, animated bool) {
	if s, ok := art.(animatedSetter); ok {
		s.SetAnimated(animated)
	}
}

func setVisible(art core.Artist, visible bool) {
	if s, ok := art.(visibleSetter); ok {
		s.SetVisible(visible)
	}
}
