package diag_test

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/diag"
	internaldiag "github.com/cwbudde/matplotlib-go/internal/diag"
)

func TestSetHandlerReceivesLibraryWarnings(t *testing.T) {
	var got []string
	restore := diag.SetHandler(func(msg string) { got = append(got, msg) })
	defer restore()

	internaldiag.Warnf("unknown colormap %q; falling back to %s", "foo", "viridis")

	if len(got) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(got), got)
	}
	want := `unknown colormap "foo"; falling back to viridis`
	if got[0] != want {
		t.Fatalf("warning = %q, want %q", got[0], want)
	}
}

func TestSetHandlerRestoresPrevious(t *testing.T) {
	var first []string
	restore := diag.SetHandler(func(msg string) { first = append(first, msg) })

	// Nested handler.
	var second []string
	restoreInner := diag.SetHandler(func(msg string) { second = append(second, msg) })
	internaldiag.Warnf("inner")
	restoreInner()

	internaldiag.Warnf("outer")
	restore()

	if len(second) != 1 || second[0] != "inner" {
		t.Fatalf("inner handler = %v, want [inner]", second)
	}
	if len(first) != 1 || first[0] != "outer" {
		t.Fatalf("outer handler = %v, want [outer]", first)
	}
}

func TestSetHandlerNilSilences(t *testing.T) {
	var got []string
	restoreCapture := diag.SetHandler(func(msg string) { got = append(got, msg) })
	defer restoreCapture()

	restoreNil := diag.SetHandler(nil)
	internaldiag.Warnf("silenced")
	restoreNil()

	if len(got) != 0 {
		t.Fatalf("nil handler leaked warnings: %v", got)
	}

	internaldiag.Warnf("audible")
	if len(got) != 1 || got[0] != "audible" {
		t.Fatalf("restored handler = %v, want [audible]", got)
	}
}
