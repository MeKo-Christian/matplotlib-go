package canvas

import (
	"image"
	"sync"
	"time"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// DrawIdleCanvas is an optional FigureCanvas extension for draw_idle-style
// scheduling.
type DrawIdleCanvas interface {
	FigureCanvas
	DrawIdle() error
}

// BlitCanvas is an optional FigureCanvas extension for Matplotlib-style
// copy_from_bbox / restore_region / blit redraw paths.
type BlitCanvas interface {
	FigureCanvas
	CopyFromBBox(bbox geom.Rect) *render.BufferRegion
	RestoreRegion(region *render.BufferRegion, bbox *geom.Rect, offset geom.Pt) error
	Blit(bbox geom.Rect) error
}

// RasterCanvas is an optional FigureCanvas extension exposing the most recently
// rendered frame as an RGBA buffer in display pixel order. It is the Go analogue
// of reading matplotlib's Agg canvas buffer_rgba after a draw, and is used by the
// animation movie writers to grab frames without a concrete backend dependency.
type RasterCanvas interface {
	FigureCanvas
	// FrameRGBA returns the pixels produced by the last Draw, or nil when the
	// canvas has not drawn yet or its renderer does not expose RGBA output.
	FrameRGBA() *image.RGBA
}

// Timer represents a backend or event-loop timer.
type Timer interface {
	Start() error
	Stop() error
	Running() bool
}

// EventLoop defines the minimal scheduling behavior interactive backends need.
type EventLoop interface {
	CallSoon(func() error) error
	NewTimer(interval time.Duration, callback func() error) Timer
}

// SynchronousEventLoop is the headless fallback loop. It executes queued work
// immediately and uses time.Ticker for timers.
type SynchronousEventLoop struct{}

// CallSoon runs callback immediately.
func (SynchronousEventLoop) CallSoon(callback func() error) error {
	if callback == nil {
		return nil
	}
	return callback()
}

// NewTimer creates a ticker-backed timer.
func (SynchronousEventLoop) NewTimer(interval time.Duration, callback func() error) Timer {
	return NewTimer(interval, callback)
}

type tickerTimer struct {
	mu       sync.Mutex
	interval time.Duration
	callback func() error
	stop     chan struct{}
	running  bool
}

// NewTimer creates a ticker-backed timer for headless or simple backends.
func NewTimer(interval time.Duration, callback func() error) Timer {
	return &tickerTimer{interval: interval, callback: callback}
}

func (t *tickerTimer) Start() error {
	if t == nil || t.callback == nil || t.interval <= 0 {
		return nil
	}

	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return nil
	}
	stop := make(chan struct{})
	t.stop = stop
	t.running = true
	interval := t.interval
	callback := t.callback
	t.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = callback()
			case <-stop:
				return
			}
		}
	}()
	return nil
}

func (t *tickerTimer) Stop() error {
	if t == nil {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.running {
		return nil
	}
	close(t.stop)
	t.stop = nil
	t.running = false
	return nil
}

func (t *tickerTimer) Running() bool {
	if t == nil {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.running
}
