package animation

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// fakeCanvas is a deterministic FigureCanvas used by the animation tests.
// It records every Draw call so the tests can assert frame counts without
// running a real backend.
type fakeCanvas struct {
	fig       *core.Figure
	drawCount int
	drawErr   error
}

func newFakeCanvas() *fakeCanvas {
	return &fakeCanvas{fig: &core.Figure{SizePx: geom.Pt{X: 100, Y: 100}}}
}

func (f *fakeCanvas) Figure() *core.Figure { return f.fig }
func (f *fakeCanvas) Draw() error {
	f.drawCount++
	return f.drawErr
}

func (f *fakeCanvas) Resize(_, _ int) error                                            { return nil }
func (f *fakeCanvas) Connect(_ canvas.EventType, _ canvas.Handler) canvas.ConnectionID { return 0 }
func (f *fakeCanvas) Disconnect(_ canvas.ConnectionID)                                 {}
func (f *fakeCanvas) Close() error                                                     { return nil }

// blitFakeCanvas adds blit capability and records overlay calls.
type blitFakeCanvas struct {
	*fakeCanvas
	region       *render.BufferRegion
	captureCount int
	restoreCount int
	blitCount    int
}

func newBlitFakeCanvas() *blitFakeCanvas {
	return &blitFakeCanvas{fakeCanvas: newFakeCanvas()}
}

func (b *blitFakeCanvas) CopyFromBBox(bbox geom.Rect) *render.BufferRegion {
	b.captureCount++
	b.region = &render.BufferRegion{Rect: bbox}
	return b.region
}

func (b *blitFakeCanvas) RestoreRegion(_ *render.BufferRegion, _ *geom.Rect, _ geom.Pt) error {
	b.restoreCount++
	return nil
}

func (b *blitFakeCanvas) Blit(_ geom.Rect) error {
	b.blitCount++
	return nil
}

// fakeArtist supports SetAnimated / Animated / SetVisible / Visible to mirror
// the embedded ArtistRasterization that real core artists carry.
type fakeArtist struct {
	core.ArtistRasterization

	zOrder    float64
	drawCalls int
	mu        sync.Mutex
}

func (a *fakeArtist) Draw(_ render.Renderer, _ *core.DrawContext) {
	a.mu.Lock()
	a.drawCalls++
	a.mu.Unlock()
}
func (a *fakeArtist) Z() float64                           { return a.zOrder }
func (a *fakeArtist) Bounds(_ *core.DrawContext) geom.Rect { return geom.Rect{} }

type pickableFakeArtist struct {
	fakeArtist
}

func (a *pickableFakeArtist) Contains(geom.Pt, *core.DrawContext) (bool, core.PickInfo) {
	return true, core.PickInfo{}
}

func TestNewFuncAnimationRejectsBadConfig(t *testing.T) {
	if _, err := NewFuncAnimation(Config{}, func(int) ([]core.Artist, error) { return nil, nil }, nil); !errors.Is(err, ErrNilCanvas) {
		t.Fatalf("expected ErrNilCanvas, got %v", err)
	}
	cnv := newFakeCanvas()
	if _, err := NewFuncAnimation(Config{Canvas: cnv}, nil, nil); !errors.Is(err, ErrNilUpdate) {
		t.Fatalf("expected ErrNilUpdate, got %v", err)
	}
}

func TestFuncAnimationStepInvokesUpdateAndDraws(t *testing.T) {
	cnv := newFakeCanvas()
	updates := 0
	anim, err := NewFuncAnimation(Config{Canvas: cnv, Frames: 3}, func(frame int) ([]core.Artist, error) {
		updates++
		if frame != updates-1 {
			t.Fatalf("update saw frame %d, expected %d", frame, updates-1)
		}
		return nil, nil
	}, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := anim.Step(); err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
	}
	if updates != 3 || cnv.drawCount != 3 || anim.TotalFramesDrawn() != 3 {
		t.Fatalf("updates=%d draws=%d total=%d, want 3/3/3", updates, cnv.drawCount, anim.TotalFramesDrawn())
	}
	if _, err := anim.Step(); !errors.Is(err, ErrNoFramesToRun) {
		t.Fatalf("4th Step: expected ErrNoFramesToRun, got %v", err)
	}
}

func TestFuncAnimationInitRunsOnce(t *testing.T) {
	cnv := newFakeCanvas()
	initCalls := 0
	anim, err := NewFuncAnimation(Config{Canvas: cnv, Frames: 3},
		func(int) ([]core.Artist, error) { return nil, nil },
		func() ([]core.Artist, error) { initCalls++; return nil, nil })
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := anim.Step(); err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
	}
	if initCalls != 1 {
		t.Fatalf("init calls = %d, want 1", initCalls)
	}
}

func TestFuncAnimationRepeatWrapsFrameCounter(t *testing.T) {
	cnv := newFakeCanvas()
	anim, err := NewFuncAnimation(Config{Canvas: cnv, Frames: 2, Repeat: true},
		func(int) ([]core.Artist, error) { return nil, nil }, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	seen := make([]int, 0, 5)
	for i := 0; i < 5; i++ {
		f, err := anim.Step()
		if err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
		seen = append(seen, f)
	}
	want := []int{0, 1, 0, 1, 0}
	for i, v := range want {
		if seen[i] != v {
			t.Fatalf("frame sequence %v, want %v", seen, want)
		}
	}
}

func TestFuncAnimationRegistersAnimatedArtists(t *testing.T) {
	cnv := newFakeCanvas()
	art := &fakeArtist{}
	anim, err := NewFuncAnimation(Config{Canvas: cnv, Frames: 1},
		func(int) ([]core.Artist, error) { return []core.Artist{art}, nil }, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	if _, err := anim.Step(); err != nil {
		t.Fatalf("Step: %v", err)
	}
	if !art.Animated() {
		t.Fatal("artist returned from update should be marked animated")
	}
}

func TestFuncAnimationDoesNotStealWidgetPickLayer(t *testing.T) {
	cnv := newFakeCanvas()
	cnv.fig = core.NewFigure(120, 80)
	ax := cnv.fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	button := ax.Button("Run")
	art := &pickableFakeArtist{fakeArtist: fakeArtist{zOrder: 10000}}
	ax.Add(art)

	anim, err := NewFuncAnimation(Config{Canvas: cnv, Frames: 1},
		func(int) ([]core.Artist, error) { return []core.Artist{art}, nil }, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	if _, err := anim.Step(); err != nil {
		t.Fatalf("Step: %v", err)
	}
	if !art.Animated() {
		t.Fatal("animated data artist should be marked animated")
	}

	hits := canvas.Pick(cnv.fig, geom.Pt{X: 60, Y: 40})
	if len(hits) == 0 {
		t.Fatal("expected pick hits")
	}
	if hits[0].Artist != button {
		t.Fatalf("top pick = %T, want button widget", hits[0].Artist)
	}
}

func TestArtistAnimationTogglesVisibilityPerFrame(t *testing.T) {
	cnv := newFakeCanvas()
	a := &fakeArtist{}
	b := &fakeArtist{}
	c := &fakeArtist{}
	anim, err := NewArtistAnimation(Config{Canvas: cnv}, [][]core.Artist{
		{a},
		{b, c},
		{c},
	})
	if err != nil {
		t.Fatalf("NewArtistAnimation: %v", err)
	}

	// All artists start hidden and animated.
	for _, art := range []*fakeArtist{a, b, c} {
		if art.Visible() {
			t.Fatalf("artist visible before any frame")
		}
		if !art.Animated() {
			t.Fatalf("artist not marked animated")
		}
	}

	if _, err := anim.Step(); err != nil {
		t.Fatalf("Step 0: %v", err)
	}
	if !a.Visible() || b.Visible() || c.Visible() {
		t.Fatalf("frame 0 visibility a=%v b=%v c=%v, want true/false/false",
			a.Visible(), b.Visible(), c.Visible())
	}

	if _, err := anim.Step(); err != nil {
		t.Fatalf("Step 1: %v", err)
	}
	if a.Visible() || !b.Visible() || !c.Visible() {
		t.Fatalf("frame 1 visibility a=%v b=%v c=%v, want false/true/true",
			a.Visible(), b.Visible(), c.Visible())
	}

	if _, err := anim.Step(); err != nil {
		t.Fatalf("Step 2: %v", err)
	}
	if a.Visible() || b.Visible() || !c.Visible() {
		t.Fatalf("frame 2 visibility a=%v b=%v c=%v, want false/false/true",
			a.Visible(), b.Visible(), c.Visible())
	}
}

func TestArtistAnimationRejectsEmptyFrames(t *testing.T) {
	cnv := newFakeCanvas()
	if _, err := NewArtistAnimation(Config{Canvas: cnv}, nil); !errors.Is(err, ErrEmptyFrameSet) {
		t.Fatalf("expected ErrEmptyFrameSet, got %v", err)
	}
}

func TestAnimationSaveReturnsUnsupportedWriterError(t *testing.T) {
	cnv := newFakeCanvas()
	anim, err := NewFuncAnimation(Config{Canvas: cnv, Frames: 1}, func(int) ([]core.Artist, error) {
		return nil, nil
	}, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}

	err = anim.Save("out.gif")
	if !errors.Is(err, ErrWriterUnsupported) {
		t.Fatalf("Save error = %v, want ErrWriterUnsupported", err)
	}
}

func TestBlitFastPathCapturesBackgroundAndOverlays(t *testing.T) {
	cnv := newBlitFakeCanvas()
	anim, err := NewFuncAnimation(Config{Canvas: cnv, Frames: 3, Blit: true},
		func(int) ([]core.Artist, error) { return nil, nil }, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := anim.Step(); err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
	}
	if cnv.captureCount != 1 {
		t.Fatalf("background capture count = %d, want 1", cnv.captureCount)
	}
	if cnv.restoreCount != 2 {
		t.Fatalf("restore count = %d, want 2 (frames 2 and 3)", cnv.restoreCount)
	}
	if cnv.blitCount != 3 {
		t.Fatalf("blit count = %d, want 3", cnv.blitCount)
	}
}

func TestBlitFallsBackForNonBlitCanvas(t *testing.T) {
	cnv := newFakeCanvas()
	anim, err := NewFuncAnimation(Config{Canvas: cnv, Frames: 2, Blit: true},
		func(int) ([]core.Artist, error) { return nil, nil }, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	for i := 0; i < 2; i++ {
		if _, err := anim.Step(); err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
	}
	if cnv.drawCount != 2 {
		t.Fatalf("draw count = %d, want 2 (blit fallback path)", cnv.drawCount)
	}
}

// recordingTimer captures Start/Stop calls and exposes a Fire method that
// invokes the registered callback. It models the canvas.Timer contract
// without spinning a real ticker.
type recordingTimer struct {
	mu       sync.Mutex
	cb       func() error
	interval time.Duration
	started  bool
}

func (t *recordingTimer) Start() error {
	t.mu.Lock()
	t.started = true
	t.mu.Unlock()
	return nil
}

func (t *recordingTimer) Stop() error {
	t.mu.Lock()
	t.started = false
	t.mu.Unlock()
	return nil
}

func (t *recordingTimer) Running() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.started
}

func (t *recordingTimer) fire() error {
	t.mu.Lock()
	cb := t.cb
	t.mu.Unlock()
	if cb == nil {
		return nil
	}
	return cb()
}

// recordingEventLoop captures CallSoon callbacks and exposes the most recent
// timer so tests can fire it deterministically.
type recordingEventLoop struct {
	mu        sync.Mutex
	soonCalls atomic.Int32
	lastTimer *recordingTimer
}

func (l *recordingEventLoop) CallSoon(cb func() error) error {
	l.soonCalls.Add(1)
	if cb == nil {
		return nil
	}
	return cb()
}

func (l *recordingEventLoop) NewTimer(interval time.Duration, callback func() error) canvas.Timer {
	t := &recordingTimer{cb: callback, interval: interval}
	l.mu.Lock()
	l.lastTimer = t
	l.mu.Unlock()
	return t
}

func (l *recordingEventLoop) current() *recordingTimer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.lastTimer
}

type failingStartTimer struct {
	err error
}

func (t failingStartTimer) Start() error { return t.err }
func (t failingStartTimer) Stop() error  { return nil }
func (t failingStartTimer) Running() bool {
	return false
}

type failingStartEventLoop struct {
	err error
}

func (l failingStartEventLoop) CallSoon(cb func() error) error {
	if cb == nil {
		return nil
	}
	return cb()
}

func (l failingStartEventLoop) NewTimer(time.Duration, func() error) canvas.Timer {
	return failingStartTimer{err: l.err}
}

type sequenceStartTimer struct {
	recordingTimer
	err error
}

func (t *sequenceStartTimer) Start() error {
	if t.err != nil {
		return t.err
	}
	return t.recordingTimer.Start()
}

type sequenceStartEventLoop struct {
	mu     sync.Mutex
	timers []*sequenceStartTimer
	errs   []error
}

func (l *sequenceStartEventLoop) CallSoon(cb func() error) error {
	if cb == nil {
		return nil
	}
	return cb()
}

func (l *sequenceStartEventLoop) NewTimer(interval time.Duration, callback func() error) canvas.Timer {
	l.mu.Lock()
	defer l.mu.Unlock()
	var err error
	if len(l.errs) > len(l.timers) {
		err = l.errs[len(l.timers)]
	}
	t := &sequenceStartTimer{
		recordingTimer: recordingTimer{cb: callback, interval: interval},
		err:            err,
	}
	l.timers = append(l.timers, t)
	return t
}

func (l *sequenceStartEventLoop) timer(index int) *sequenceStartTimer {
	l.mu.Lock()
	defer l.mu.Unlock()
	if index < 0 || index >= len(l.timers) {
		return nil
	}
	return l.timers[index]
}

func TestStartStopUsesEventLoopTimer(t *testing.T) {
	cnv := newFakeCanvas()
	loop := &recordingEventLoop{}
	anim, err := NewFuncAnimation(Config{Canvas: cnv, Frames: 2, EventLoop: loop, Interval: 50 * time.Millisecond},
		func(int) ([]core.Artist, error) { return nil, nil }, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	if err := anim.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !anim.Running() {
		t.Fatal("Running() = false after Start")
	}
	if err := anim.Start(); !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("expected ErrAlreadyRunning, got %v", err)
	}
	tm := loop.current()
	if tm == nil {
		t.Fatal("event loop did not receive a NewTimer call")
	}
	if tm.interval != 50*time.Millisecond {
		t.Fatalf("timer interval = %v, want 50ms", tm.interval)
	}
	// Fire two ticks; the animation should auto-stop after Frames=2.
	if err := tm.fire(); err != nil {
		t.Fatalf("tick 1: %v", err)
	}
	if err := tm.fire(); err != nil {
		t.Fatalf("tick 2: %v", err)
	}
	if anim.Running() {
		t.Fatal("finite non-repeating run should have stopped itself")
	}
	if cnv.drawCount != 2 {
		t.Fatalf("draw count after 2 ticks = %d, want 2", cnv.drawCount)
	}
	if err := anim.Stop(); err != nil {
		t.Fatalf("Stop after auto-stop: %v", err)
	}
}

func TestStartRollsBackRunningStateWhenTimerStartFails(t *testing.T) {
	cnv := newFakeCanvas()
	startErr := errors.New("timer start failed")
	anim, err := NewFuncAnimation(Config{
		Canvas:    cnv,
		Frames:    1,
		EventLoop: failingStartEventLoop{err: startErr},
	}, func(int) ([]core.Artist, error) { return nil, nil }, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}

	if err := anim.Start(); !errors.Is(err, startErr) {
		t.Fatalf("Start error = %v, want %v", err, startErr)
	}
	if anim.Running() {
		t.Fatal("animation should not remain running after timer start failure")
	}
	if err := anim.Stop(); err != nil {
		t.Fatalf("Stop after failed Start: %v", err)
	}
}

func TestRepeatDelayInsertsSkipTick(t *testing.T) {
	cnv := newFakeCanvas()
	loop := &recordingEventLoop{}
	anim, err := NewFuncAnimation(Config{
		Canvas:      cnv,
		Frames:      2,
		Repeat:      true,
		RepeatDelay: 250 * time.Millisecond,
		Interval:    50 * time.Millisecond,
		EventLoop:   loop,
	}, func(int) ([]core.Artist, error) { return nil, nil }, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	if err := anim.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Fire frames 0 and 1.
	if err := loop.current().fire(); err != nil {
		t.Fatalf("tick frame 0: %v", err)
	}
	if err := loop.current().fire(); err != nil {
		t.Fatalf("tick frame 1: %v", err)
	}
	// Next tick should be a repeat-delay skip; no new draw happens, but the
	// animation reschedules using the delay duration.
	prevDraws := cnv.drawCount
	if err := loop.current().fire(); err != nil {
		t.Fatalf("repeat-delay skip tick: %v", err)
	}
	delayTimer := loop.current()
	if delayTimer.interval != 250*time.Millisecond {
		t.Fatalf("repeat-delay timer interval = %v, want 250ms", delayTimer.interval)
	}
	if cnv.drawCount != prevDraws {
		t.Fatalf("draw count changed during repeat-delay skip: %d -> %d", prevDraws, cnv.drawCount)
	}
	// Firing the delay timer draws one frame and reinstalls the interval timer.
	if err := delayTimer.fire(); err != nil {
		t.Fatalf("delay timer fire: %v", err)
	}
	if cnv.drawCount != prevDraws+1 {
		t.Fatalf("frame after repeat-delay = %d, want %d", cnv.drawCount, prevDraws+1)
	}
	if err := anim.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestStopDuringRepeatDelayLetsRestartDrawImmediately(t *testing.T) {
	cnv := newFakeCanvas()
	loop := &recordingEventLoop{}
	anim, err := NewFuncAnimation(Config{
		Canvas:      cnv,
		Frames:      1,
		Repeat:      true,
		RepeatDelay: 250 * time.Millisecond,
		EventLoop:   loop,
	}, func(int) ([]core.Artist, error) { return nil, nil }, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	if err := anim.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := loop.current().fire(); err != nil {
		t.Fatalf("first tick: %v", err)
	}
	if cnv.drawCount != 1 {
		t.Fatalf("draw count after first tick = %d, want 1", cnv.drawCount)
	}
	if err := anim.Stop(); err != nil {
		t.Fatalf("Stop during repeat delay: %v", err)
	}
	if err := anim.Start(); err != nil {
		t.Fatalf("restart: %v", err)
	}
	if err := loop.current().fire(); err != nil {
		t.Fatalf("first tick after restart: %v", err)
	}
	if cnv.drawCount != 2 {
		t.Fatalf("draw count after restart tick = %d, want immediate draw 2", cnv.drawCount)
	}
}

func TestRepeatDelayRestartFailureStopsAnimation(t *testing.T) {
	cnv := newFakeCanvas()
	restartErr := errors.New("restart failed")
	loop := &sequenceStartEventLoop{
		errs: []error{nil, nil, restartErr},
	}
	anim, err := NewFuncAnimation(Config{
		Canvas:      cnv,
		Frames:      1,
		Repeat:      true,
		RepeatDelay: 250 * time.Millisecond,
		EventLoop:   loop,
	}, func(int) ([]core.Artist, error) { return nil, nil }, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	if err := anim.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	intervalTimer := loop.timer(0)
	if intervalTimer == nil {
		t.Fatal("missing initial interval timer")
	}
	if err := intervalTimer.fire(); err != nil {
		t.Fatalf("first tick: %v", err)
	}
	if err := intervalTimer.fire(); err != nil {
		t.Fatalf("repeat-delay scheduling tick: %v", err)
	}
	delayTimer := loop.timer(1)
	if delayTimer == nil {
		t.Fatal("missing repeat-delay timer")
	}
	if err := delayTimer.fire(); !errors.Is(err, restartErr) {
		t.Fatalf("delay timer fire error = %v, want %v", err, restartErr)
	}
	if anim.Running() {
		t.Fatal("animation should stop after interval restart failure")
	}
	if err := anim.Stop(); err != nil {
		t.Fatalf("Stop after restart failure: %v", err)
	}
}

func TestStopBeforeStartIsNoop(t *testing.T) {
	cnv := newFakeCanvas()
	anim, err := NewFuncAnimation(Config{Canvas: cnv, Frames: 1},
		func(int) ([]core.Artist, error) { return nil, nil }, nil)
	if err != nil {
		t.Fatalf("NewFuncAnimation: %v", err)
	}
	if err := anim.Stop(); err != nil {
		t.Fatalf("Stop before Start: %v", err)
	}
}
