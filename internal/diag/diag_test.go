package diag

import "testing"

func TestWarnfRoutesFormattedMessageToHandler(t *testing.T) {
	var got []string
	restore := SetHandler(func(msg string) { got = append(got, msg) })
	defer restore()

	Warnf("unknown colormap %q; falling back to %s", "foo", "viridis")

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
	restore := SetHandler(func(msg string) { first = append(first, msg) })

	// Nested handler.
	var second []string
	restoreInner := SetHandler(func(msg string) { second = append(second, msg) })
	Warnf("inner")
	restoreInner()

	Warnf("outer")
	restore()

	if len(second) != 1 || second[0] != "inner" {
		t.Fatalf("inner handler = %v, want [inner]", second)
	}
	if len(first) != 1 || first[0] != "outer" {
		t.Fatalf("outer handler = %v, want [outer]", first)
	}
}

func TestSetHandlerNilSilences(t *testing.T) {
	restore := SetHandler(nil)
	defer restore()
	// Must not panic with a nil handler.
	Warnf("ignored")
}
