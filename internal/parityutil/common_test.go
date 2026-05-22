package common

import "testing"

func TestUniformDataUsesDeterministicPCGStream(t *testing.T) {
	got := UniformData(19680801, 0, 3, 23, 32)
	want := UniformData(19680801, 0, 3, 23, 32)
	if len(got) != 3 {
		t.Fatalf("UniformData length = %d, want 3", len(got))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("UniformData is not deterministic at %d: %v != %v", i, got[i], want[i])
		}
		if got[i] < 23 || got[i] >= 32 {
			t.Fatalf("UniformData[%d] = %v outside [23, 32)", i, got[i])
		}
	}
}

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
