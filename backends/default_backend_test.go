package backends_test

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/agg"
	_ "github.com/cwbudde/matplotlib-go/backends/gobasic"
)

func TestDefaultBackendPreference(t *testing.T) {
	backend, err := backends.BestBackend(nil)
	if err != nil {
		t.Fatalf("BestBackend failed: %v", err)
	}
	if backend != backends.GoBasic {
		t.Fatalf("expected default backend %s, got %s", backends.GoBasic, backend)
	}
}

func TestCreateAllAvailableBackends(t *testing.T) {
	available := backends.Available()
	if len(available) == 0 {
		t.Fatal("expected at least one available backend")
	}

	for _, backend := range available {
		backend := backend
		t.Run(string(backend), func(t *testing.T) {
			renderer, err := backends.Create(backend, backends.TestDefaultConfig(100, 100))
			if err != nil {
				t.Fatalf("Create(%s) failed: %v", backend, err)
			}
			if renderer == nil {
				t.Fatalf("Create(%s) returned nil renderer", backend)
			}
		})
	}
}
