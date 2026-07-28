package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

// The pointer and magic-zero option spellings this replaced could not
// tell "the caller wants zero" from "the caller said nothing". These tests pin
// the distinction for the four APIs migrated to optional.Value.

func TestAnnotateDistinguishesExplicitZeroOffsetFromUnset(t *testing.T) {
	ax := NewFigure(400, 300).AddAxes(geom.Rect{})

	defaulted := ax.Annotate("d", 0, 0, AnnotationOptions{})
	if defaulted.OffsetX != 28 || defaulted.OffsetY != -20 {
		t.Fatalf("unset offsets = (%v, %v), want Matplotlib (28, -20)", defaulted.OffsetX, defaulted.OffsetY)
	}

	// Previously unreachable: (0, 0) was the sentinel that selected the default.
	pinned := ax.Annotate("p", 0, 0, AnnotationOptions{
		OffsetX: optional.Of(0.0),
		OffsetY: optional.Of(0.0),
	})
	if pinned.OffsetX != 0 || pinned.OffsetY != 0 {
		t.Fatalf("explicit zero offsets = (%v, %v), want (0, 0)", pinned.OffsetX, pinned.OffsetY)
	}
}

func TestAnnotateDistinguishesExplicitZeroArrowSizeFromUnset(t *testing.T) {
	ax := NewFigure(400, 300).AddAxes(geom.Rect{})

	// An unset head size takes the annotation's font size, which is Matplotlib's
	// mutation_scale default; here that resolves through the axes rc.
	rcFontSize := ax.resolvedRC().FontSize
	defaulted := ax.Annotate("d", 0, 0, AnnotationOptions{})
	if defaulted.ArrowWidth != 1.25 || defaulted.ArrowHeadSize != rcFontSize {
		t.Fatalf("unset arrow = width %v head %v, want 1.25 and %v", defaulted.ArrowWidth, defaulted.ArrowHeadSize, rcFontSize)
	}

	sized := ax.Annotate("s", 0, 0, AnnotationOptions{FontSize: 14})
	if sized.ArrowHeadSize != 14 {
		t.Fatalf("unset arrow head for 14 pt text = %v, want the font size 14", sized.ArrowHeadSize)
	}

	pinned := ax.Annotate("p", 0, 0, AnnotationOptions{
		ArrowWidth:    optional.Of(0.0),
		ArrowHeadSize: optional.Of(0.0),
	})
	if pinned.ArrowWidth != 0 || pinned.ArrowHeadSize != 0 {
		t.Fatalf("explicit zero arrow = width %v head %v, want 0 and 0", pinned.ArrowWidth, pinned.ArrowHeadSize)
	}
}

func TestHLinesDistinguishesExplicitZeroFromUnset(t *testing.T) {
	y := []float64{1, 2}
	xMin := []float64{0}
	xMax := []float64{4}

	defaulted := NewFigure(400, 300).AddAxes(geom.Rect{}).HLines(y, xMin, xMax, LineCollectionOptions{})
	if defaulted.Alpha != 1 || defaulted.LineWidth != 1 {
		t.Fatalf("unset = alpha %v width %v, want 1 and 1", defaulted.Alpha, defaulted.LineWidth)
	}

	pinned := NewFigure(400, 300).AddAxes(geom.Rect{}).HLines(y, xMin, xMax, LineCollectionOptions{
		Alpha:     optional.Of(0.0),
		LineWidth: optional.Of(0.0),
	})
	if pinned.Alpha != 0 || pinned.LineWidth != 0 {
		t.Fatalf("explicit zero = alpha %v width %v, want 0 and 0", pinned.Alpha, pinned.LineWidth)
	}
}

// TestHLinesLeavesPropertyCycleAloneWhenColorsSupplied pins the fallback rule
// the migration had to preserve: the scalar Color only reaches for the property
// cycle when no per-segment Colors slice would override it.
func TestHLinesLeavesPropertyCycleAloneWhenColorsSupplied(t *testing.T) {
	ax := NewFigure(400, 300).AddAxes(geom.Rect{})
	first := ax.NextColor()

	perSegment := []render.Color{{R: 1, A: 1}, {G: 1, A: 1}}
	ax.HLines([]float64{1, 2}, []float64{0}, []float64{4}, LineCollectionOptions{Colors: perSegment})

	if next := ax.NextColor(); next == first {
		t.Fatal("property cycle did not advance across the two direct NextColor calls; test cannot detect a leak")
	}
	cycled := ax.HLines([]float64{3}, []float64{0}, []float64{4}, LineCollectionOptions{})
	if cycled.Color == (render.Color{}) {
		t.Fatal("HLines without Colors should take a color from the property cycle")
	}
}

func TestImShowOriginCanOverrideANonDefaultRC(t *testing.T) {
	fig := NewFigure(400, 300)
	fig.RC.Image.Origin = "lower"
	ax := fig.AddAxes(geom.Rect{})

	if img := ax.ImShow([][]float64{{0, 1}, {2, 3}}, ImShowOptions{}); img.Origin != ImageOriginLower {
		t.Fatalf("unset origin = %v, want the rc default lower", img.Origin)
	}

	// Previously unreachable: ImageOriginUpper is the zero value, which the old
	// merge treated as "unset" and replaced with the rc default.
	forced := ax.ImShow([][]float64{{0, 1}, {2, 3}}, ImShowOptions{
		Origin: optional.Of(ImageOriginUpper),
	})
	if forced.Origin != ImageOriginUpper {
		t.Fatalf("explicit origin = %v, want upper to override the rc default", forced.Origin)
	}
}

func TestStemDistinguishesExplicitZeroFromUnset(t *testing.T) {
	x := []float64{0, 1, 2}
	y := []float64{1, 3, 2}

	defaulted := NewFigure(400, 300).AddAxes(geom.Rect{}).Stem(x, y, StemOptions{})
	if defaulted.StemLines.LineWidth != 1.5 {
		t.Fatalf("unset line width = %v, want 1.5", defaulted.StemLines.LineWidth)
	}

	pinned := NewFigure(400, 300).AddAxes(geom.Rect{}).Stem(x, y, StemOptions{
		LineWidth: optional.Of(0.0),
	})
	if pinned.StemLines.LineWidth != 0 {
		t.Fatalf("explicit zero line width = %v, want 0", pinned.StemLines.LineWidth)
	}
}
