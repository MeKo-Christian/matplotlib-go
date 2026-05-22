//go:build skia && !js

package main

import "testing"

func TestSelectedBackendIDAcceptsSkiaWithBuildTag(t *testing.T) {
	got, err := selectedBackendID("skia")
	if err != nil {
		t.Fatalf("selectedBackendID(skia) error = %v", err)
	}
	if got != "skia" {
		t.Fatalf("selectedBackendID(skia) = %q, want skia", got)
	}
}
