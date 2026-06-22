package render

import "testing"

func TestSketchActive(t *testing.T) {
	if SketchActive(SketchParams{}) {
		t.Fatal("zero params should be inactive")
	}
	if SketchActive(SketchParams{Length: 100, Randomness: 2}) {
		t.Fatal("scale==0 should be inactive regardless of length/randomness")
	}
	if !SketchActive(SketchParams{Scale: 1}) {
		t.Fatal("nonzero scale should be active")
	}
}

func TestEffectiveSketch(t *testing.T) {
	def := SketchParams{Scale: 1, Length: 100, Randomness: 2}
	override := SketchParams{Scale: 3, Length: 40, Randomness: 5}

	if got := EffectiveSketch(SketchParams{}, def); got != def {
		t.Fatalf("zero paint should fall back to default: got %+v", got)
	}
	if got := EffectiveSketch(override, def); got != override {
		t.Fatalf("explicit paint should win: got %+v", got)
	}
	if got := EffectiveSketch(SketchParams{}, SketchParams{}); got != (SketchParams{}) {
		t.Fatalf("no default and no override should stay zero: got %+v", got)
	}
}
