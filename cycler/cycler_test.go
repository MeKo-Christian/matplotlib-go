package cycler

import (
	"reflect"
	"testing"
)

func TestNewAndByKey(t *testing.T) {
	c := New("color", "r", "g", "b")
	if c.Len() != 3 {
		t.Fatalf("Len = %d, want 3", c.Len())
	}
	if got, want := c.ByKey("color"), []any{"r", "g", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ByKey(color) = %v, want %v", got, want)
	}
	if c.ByKey("missing") != nil {
		t.Fatalf("ByKey(missing) should be nil")
	}
}

func TestRowWrapAround(t *testing.T) {
	c := New("color", "r", "g")
	if got := c.Row(0)["color"]; got != "r" {
		t.Fatalf("Row(0) = %v, want r", got)
	}
	if got := c.Row(3)["color"]; got != "g" {
		t.Fatalf("Row(3) = %v, want g (wrap)", got)
	}
	if got := c.Row(-1)["color"]; got != "g" {
		t.Fatalf("Row(-1) = %v, want g", got)
	}
}

func TestConcatZipsEqualLength(t *testing.T) {
	c, err := New("color", "r", "g").Concat(New("linestyle", "-", "--"))
	if err != nil {
		t.Fatalf("Concat: %v", err)
	}
	if c.Len() != 2 {
		t.Fatalf("Len = %d, want 2", c.Len())
	}
	if got, want := c.Keys(), []string{"color", "linestyle"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys = %v, want %v", got, want)
	}
	want := []map[string]any{
		{"color": "r", "linestyle": "-"},
		{"color": "g", "linestyle": "--"},
	}
	for i, w := range want {
		if got := c.Row(i); !reflect.DeepEqual(got, w) {
			t.Fatalf("Row(%d) = %v, want %v", i, got, w)
		}
	}
}

func TestConcatUnequalLengthErrors(t *testing.T) {
	if _, err := New("color", "r", "g", "b").Concat(New("linestyle", "-", "--")); err == nil {
		t.Fatalf("expected length-mismatch error")
	}
}

func TestConcatOverlappingKeysErrors(t *testing.T) {
	if _, err := New("color", "r").Concat(New("color", "g")); err == nil {
		t.Fatalf("expected overlapping-key error")
	}
}

func TestMultiplyOuterProductLeftSlowest(t *testing.T) {
	c, err := New("color", "r", "g").Multiply(New("linestyle", "-", "--"))
	if err != nil {
		t.Fatalf("Multiply: %v", err)
	}
	want := []map[string]any{
		{"color": "r", "linestyle": "-"},
		{"color": "r", "linestyle": "--"},
		{"color": "g", "linestyle": "-"},
		{"color": "g", "linestyle": "--"},
	}
	if c.Len() != len(want) {
		t.Fatalf("Len = %d, want %d", c.Len(), len(want))
	}
	for i, w := range want {
		if got := c.Row(i); !reflect.DeepEqual(got, w) {
			t.Fatalf("Row(%d) = %v, want %v", i, got, w)
		}
	}
}

func TestMultiplyOverlappingKeysErrors(t *testing.T) {
	if _, err := New("color", "r").Multiply(New("color", "g")); err == nil {
		t.Fatalf("expected overlapping-key error")
	}
}

func TestCloneIsIndependent(t *testing.T) {
	c := New("color", "r", "g")
	clone := c.Clone()
	clone.rows[0]["color"] = "x"
	if c.Row(0)["color"] != "r" {
		t.Fatalf("clone mutated original")
	}
}

func TestNilCyclerSafe(t *testing.T) {
	var c *Cycler
	if c.Len() != 0 || c.Keys() != nil || c.Has("color") || c.Row(0) != nil {
		t.Fatalf("nil cycler should behave as empty")
	}
}
