//go:build !skia

package webdemo

import "testing"

func TestSkiaBackendHiddenWithoutBuildTag(t *testing.T) {
	if ValidBackendID("skia") {
		t.Fatal("ValidBackendID(\"skia\") = true without skia build tag")
	}
}
