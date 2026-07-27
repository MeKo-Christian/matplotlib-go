// Package desktop defines the host-agnostic contract for matplotlib-go's
// desktop interactive backends.
//
// See docs/adr/0001-desktop-interactive-backend.md for the toolkit
// decision (Gio). This package owns the surface that the rest of
// matplotlib-go talks to; concrete toolkit adapters (Gio today, Qt /
// GTK / WebGPU in the future) live in subpackages such as
// backends/desktop/gio.
//
// The actual window / event-loop implementation is deferred to a
// follow-up commit; this file scaffolds the types so
// callers can be written and unit-tested against the contract before
// any GUI dependency lands in go.mod.
//
// The default-constructor registry is safe for concurrent registration and
// lookup. Backend instances retain their documented event-loop ownership rules;
// see docs/concurrency.md for the shared figure and renderer contract.
package desktop

import (
	"errors"
	"sync"

	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

// RendererFactory builds a fresh raster renderer sized for the given
// pixel dimensions. The backend calls this once at startup and again on
// every resize. The returned renderer must satisfy the standard
// render.Renderer contract; callers typically pass backends/agg.New.
type RendererFactory func(width, height int, bg render.Color) (render.Renderer, error)

// Options configures a desktop backend.
//
// All fields are optional except Figure. Sensible defaults are documented
// per field; the zero value of Options other than Figure produces a
// 640×480 window with a white background.
type Options struct {
	// Figure is the figure to display. Required.
	Figure *core.Figure

	// Title is the initial window title. Defaults to "Figure".
	Title string

	// Width and Height are the initial window size in logical pixels. If
	// zero, defaults to the figure's pixel size or 640×480.
	Width, Height int

	// DPI overrides the renderer DPI. Zero uses the figure default.
	DPI float64

	// Background is the renderer clear color. The zero value (transparent
	// black) is treated as opaque white.
	Background render.Color

	// Renderer constructs the raster renderer used to paint each frame.
	// Defaults to a stub factory returning ErrRendererNotConfigured; real
	// callers pass backends/agg.New.
	Renderer RendererFactory

	// Navigation, when non-nil, is reused as the backend's navigation
	// helper. If nil, the backend creates one via canvas.NewNavigation.
	Navigation *canvas.Navigation

	// Toolbar, when non-nil, is reused as the backend's toolbar
	// controller. If nil, the backend creates one via
	// canvas.NewToolbarController bound to Navigation.
	Toolbar *canvas.ToolbarController

	// SaveHandler is wired to ToolbarSave when the toolbar is created
	// internally. Ignored if the caller supplies their own Toolbar.
	SaveHandler canvas.ToolbarHandler
}

// Backend is the host-agnostic contract for a desktop interactive
// backend. A Backend owns one OS window, one event-pump goroutine, and
// one frame producer.
//
// Run blocks until the window is closed (or the supplied context cancels
// in implementations that accept one). Show is the non-blocking variant
// for embedding into an existing event loop.
type Backend interface {
	// Canvas returns the FigureCanvas presented in this window. The
	// returned canvas is wired to the backend's dispatcher; callers may
	// register handlers with canvas.Connect.
	Canvas() canvas.DrawIdleCanvas

	// Toolbar returns the navigation toolbar attached to this window.
	Toolbar() canvas.NavigationToolbar

	// Navigation returns the navigation helper attached to this window.
	Navigation() *canvas.Navigation

	// Run pumps the backend's event loop until the window closes. It is
	// safe to call Run only once per backend.
	Run() error

	// Close requests that the backend shut down. Idempotent.
	Close() error
}

// ErrRendererNotConfigured is returned by the default RendererFactory.
// Callers must supply Options.Renderer (typically backends/agg.New) to
// produce real pixels.
var ErrRendererNotConfigured = errors.New("desktop: renderer factory not configured (set Options.Renderer)")

// ErrNoFigure is returned when New is called without a figure.
var ErrNoFigure = errors.New("desktop: Options.Figure is required")

// ErrNotImplemented is returned by the placeholder constructor in this
// skeleton until the Gio adapter lands. Callers can use it to detect
// "the backend is wired but the host is not yet linked in" cases.
var ErrNotImplemented = errors.New("desktop: interactive backend not yet implemented; see docs/adr/0001-desktop-interactive-backend.md")

// Constructor builds a Backend from Options. Toolkit adapters register
// their constructor via Register; the default selection — currently the
// Gio adapter — is resolved by New.
type Constructor func(Options) (Backend, error)

var defaultConstructor = struct {
	sync.RWMutex
	constructor Constructor
}{}

// Register installs the default Backend constructor. Toolkit adapters
// call this from an init function. Registering twice replaces the
// previous constructor; the last init wins, which lets a test or example
// swap in a mock. Register is safe for concurrent use with New and
// DefaultConstructor.
func Register(c Constructor) {
	defaultConstructor.Lock()
	defaultConstructor.constructor = c
	defaultConstructor.Unlock()
}

// DefaultConstructor returns the currently-registered constructor, or
// nil if none has been registered yet.
func DefaultConstructor() Constructor {
	defaultConstructor.RLock()
	defer defaultConstructor.RUnlock()
	return defaultConstructor.constructor
}

// New constructs a desktop backend using the registered default
// constructor. Returns ErrNotImplemented when no adapter has been linked
// in — import _ "github.com/cwbudde/matplotlib-go/backends/desktop/gio"
// (once the adapter lands) to register one.
func New(opts Options) (Backend, error) {
	if opts.Figure == nil {
		return nil, ErrNoFigure
	}
	constructor := DefaultConstructor()
	if constructor == nil {
		return nil, ErrNotImplemented
	}
	return constructor(opts.withDefaults())
}

// withDefaults returns a copy of opts with zero fields replaced by
// sensible defaults. Backends should call this if they bypass New.
func (o Options) withDefaults() Options {
	out := o
	if out.Title == "" {
		out.Title = "Figure"
	}
	if out.Width <= 0 {
		out.Width = 640
	}
	if out.Height <= 0 {
		out.Height = 480
	}
	if (out.Background == render.Color{}) {
		out.Background = render.Color{R: 1, G: 1, B: 1, A: 1}
	}
	if out.Renderer == nil {
		out.Renderer = func(int, int, render.Color) (render.Renderer, error) {
			return nil, ErrRendererNotConfigured
		}
	}
	return out
}

// WithDefaults is exported for adapter packages that want to apply the
// shared defaults without reimplementing them.
func WithDefaults(o Options) Options { return o.withDefaults() }
