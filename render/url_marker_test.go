package render

import "testing"

func TestGraphicsContextWithURLGID(t *testing.T) {
	gc := NewGraphicsContext()
	if gc.URL != "" || gc.GID != "" {
		t.Fatalf("expected empty url/gid by default, got %q/%q", gc.URL, gc.GID)
	}

	gc = gc.WithURL("https://example.com").WithGID("node-7")
	if gc.URL != "https://example.com" {
		t.Fatalf("WithURL: got %q", gc.URL)
	}
	if gc.GID != "node-7" {
		t.Fatalf("WithGID: got %q", gc.GID)
	}

	// Clone preserves metadata; clearing one returned value leaves the other.
	clone := gc.Clone()
	if clone.URL != gc.URL || clone.GID != gc.GID {
		t.Fatalf("Clone dropped url/gid: %q/%q", clone.URL, clone.GID)
	}
	if cleared := gc.WithURL(""); cleared.URL != "" || cleared.GID != "node-7" {
		t.Fatalf("WithURL(\"\") should clear only url: %q/%q", cleared.URL, cleared.GID)
	}
}
