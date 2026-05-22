//go:build skia && !js

package webdemo

import "testing"

func TestSkiaBackendSelectableWithBuildTag(t *testing.T) {
	if !ValidBackendID("skia") {
		t.Fatal("ValidBackendID(\"skia\") = false with skia build tag")
	}
	img, descriptor, err := RenderWithBackend("axes", "skia", 320, 180)
	if err != nil {
		t.Fatalf("RenderWithBackend(skia) error = %v", err)
	}
	if descriptor.ID != "axes" {
		t.Fatalf("descriptor.ID = %q, want axes", descriptor.ID)
	}
	if img == nil || img.Bounds().Empty() {
		t.Fatal("RenderWithBackend(skia) returned empty image")
	}
}
