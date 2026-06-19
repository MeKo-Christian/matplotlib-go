package color

import "testing"

func TestTab10MatchesMatplotlibExactRGB(t *testing.T) {
	want := Palette{
		{R: 31.0 / 255.0, G: 119.0 / 255.0, B: 180.0 / 255.0, A: 1},
		{R: 1, G: 127.0 / 255.0, B: 14.0 / 255.0, A: 1},
		{R: 44.0 / 255.0, G: 160.0 / 255.0, B: 44.0 / 255.0, A: 1},
		{R: 214.0 / 255.0, G: 39.0 / 255.0, B: 40.0 / 255.0, A: 1},
		{R: 148.0 / 255.0, G: 103.0 / 255.0, B: 189.0 / 255.0, A: 1},
		{R: 140.0 / 255.0, G: 86.0 / 255.0, B: 75.0 / 255.0, A: 1},
		{R: 227.0 / 255.0, G: 119.0 / 255.0, B: 194.0 / 255.0, A: 1},
		{R: 127.0 / 255.0, G: 127.0 / 255.0, B: 127.0 / 255.0, A: 1},
		{R: 188.0 / 255.0, G: 189.0 / 255.0, B: 34.0 / 255.0, A: 1},
		{R: 23.0 / 255.0, G: 190.0 / 255.0, B: 207.0 / 255.0, A: 1},
	}
	if len(Tab10) != len(want) {
		t.Fatalf("Tab10 length = %d, want %d", len(Tab10), len(want))
	}
	for i := range want {
		if Tab10[i] != want[i] {
			t.Fatalf("Tab10[%d] = %+v, want exact Matplotlib color %+v", i, Tab10[i], want[i])
		}
	}
}
