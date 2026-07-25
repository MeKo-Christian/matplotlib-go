package desktop

import (
	"errors"
	"sync"
	"testing"

	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestNewRequiresFigure(t *testing.T) {
	if _, err := New(Options{}); !errors.Is(err, ErrNoFigure) {
		t.Fatalf("New() err = %v, want ErrNoFigure", err)
	}
}

func TestNewWithoutConstructorReturnsNotImplemented(t *testing.T) {
	saved := DefaultConstructor()
	Register(nil)
	t.Cleanup(func() { Register(saved) })

	fig := &core.Figure{}
	_, err := New(Options{Figure: fig})
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("New() err = %v, want ErrNotImplemented", err)
	}
}

func TestRegisterAndNew(t *testing.T) {
	saved := DefaultConstructor()
	t.Cleanup(func() { Register(saved) })

	got := Options{}
	Register(func(o Options) (Backend, error) {
		got = o
		return mockBackend{}, nil
	})

	fig := &core.Figure{}
	b, err := New(Options{Figure: fig})
	if err != nil {
		t.Fatal(err)
	}
	if b == nil {
		t.Fatal("New returned nil backend")
	}
	if got.Title != "Figure" || got.Width != 640 || got.Height != 480 {
		t.Fatalf("defaults not applied: %+v", got)
	}
	if (got.Background != render.Color{R: 1, G: 1, B: 1, A: 1}) {
		t.Fatalf("background default = %+v, want opaque white", got.Background)
	}
	if got.Renderer == nil {
		t.Fatal("default renderer factory not installed")
	}
}

func TestConstructorRegistryConcurrentAccess(t *testing.T) {
	saved := DefaultConstructor()
	t.Cleanup(func() { Register(saved) })

	constructors := []Constructor{
		func(Options) (Backend, error) { return mockBackend{}, nil },
		func(Options) (Backend, error) { return mockBackend{}, nil },
	}
	Register(constructors[0])

	start := make(chan struct{})
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			<-start
			for i := 0; i < 500; i++ {
				if worker%2 == 0 {
					Register(constructors[i%len(constructors)])
					continue
				}
				if DefaultConstructor() == nil {
					t.Error("DefaultConstructor returned nil during registered access")
					return
				}
				backend, err := New(Options{Figure: &core.Figure{}})
				if err != nil || backend == nil {
					t.Errorf("New() = %v, %v; want a backend", backend, err)
					return
				}
			}
		}(worker)
	}
	close(start)
	wg.Wait()
}

func TestWithDefaultsPreservesNonZero(t *testing.T) {
	opts := Options{
		Figure:     &core.Figure{},
		Title:      "Custom",
		Width:      1024,
		Height:     768,
		Background: render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
	}
	out := WithDefaults(opts)
	if out.Title != "Custom" || out.Width != 1024 || out.Height != 768 {
		t.Fatalf("non-zero fields overwritten: %+v", out)
	}
	if (out.Background != render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1}) {
		t.Fatalf("background overwritten: %+v", out.Background)
	}
}

func TestDefaultRendererFactoryErrors(t *testing.T) {
	opts := WithDefaults(Options{Figure: &core.Figure{}})
	if _, err := opts.Renderer(10, 10, render.Color{}); !errors.Is(err, ErrRendererNotConfigured) {
		t.Fatalf("default factory err = %v, want ErrRendererNotConfigured", err)
	}
}

// mockBackend is the minimal Backend satisfying the interface for tests.
type mockBackend struct{}

func (mockBackend) Canvas() canvas.DrawIdleCanvas     { return nil }
func (mockBackend) Toolbar() canvas.NavigationToolbar { return nil }
func (mockBackend) Navigation() *canvas.Navigation    { return nil }
func (mockBackend) Run() error                        { return nil }
func (mockBackend) Close() error                      { return nil }
