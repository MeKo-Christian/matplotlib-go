package common

import "testing"

func TestGet3DWireframeTestDataMatchesMatplotlibScaledXY(t *testing.T) {
	x, y, z := Get3DWireframeTestData(0.05)
	if len(x) == 0 || len(y) == 0 || len(z) == 0 {
		t.Fatal("Get3DWireframeTestData returned empty data")
	}
	if got, want := x[0], -30.0; got != want {
		t.Fatalf("first x = %v, want Matplotlib-scaled %v", got, want)
	}
	if got, want := y[0], -30.0; got != want {
		t.Fatalf("first y = %v, want Matplotlib-scaled %v", got, want)
	}
	if got, want := x[len(x)-1], 29.5; got != want {
		t.Fatalf("last x = %v, want Matplotlib-scaled %v", got, want)
	}
	if got, want := y[len(y)-1], 29.5; got != want {
		t.Fatalf("last y = %v, want Matplotlib-scaled %v", got, want)
	}
}
