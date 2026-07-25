package optional_test

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/optional"
)

func TestZeroValueIsUnset(t *testing.T) {
	var v optional.Value[float64]
	if v.IsSet() {
		t.Fatal("zero Value reports set")
	}
	if got, ok := v.Get(); ok || got != 0 {
		t.Fatalf("Get() = (%v, %v), want (0, false)", got, ok)
	}
	if got := v.Or(7); got != 7 {
		t.Fatalf("Or(7) = %v, want 7", got)
	}
	if got := v.OrZero(); got != 0 {
		t.Fatalf("OrZero() = %v, want 0", got)
	}
	if v.Ptr() != nil {
		t.Fatal("Ptr() on unset Value is non-nil")
	}
	if v != optional.None[float64]() {
		t.Fatal("None differs from the zero Value")
	}
}

// TestSetZeroDiffersFromUnset covers the reason Value exists: the magic-zero
// spelling it replaces could not tell "alpha 0" from "alpha unspecified".
func TestSetZeroDiffersFromUnset(t *testing.T) {
	set := optional.Of(0.0)
	if !set.IsSet() {
		t.Fatal("Of(0) reports unset")
	}
	if got := set.Or(1); got != 0 {
		t.Fatalf("Or(1) on Of(0) = %v, want 0", got)
	}
	if set == optional.None[float64]() {
		t.Fatal("Of(0) compares equal to the unset Value")
	}
}

func TestOfRoundTrips(t *testing.T) {
	v := optional.Of("bilinear")
	got, ok := v.Get()
	if !ok || got != "bilinear" {
		t.Fatalf("Get() = (%q, %v), want (%q, true)", got, ok, "bilinear")
	}
	if got := v.Or("nearest"); got != "bilinear" {
		t.Fatalf("Or() = %q, want %q", got, "bilinear")
	}
}

func TestFromPtr(t *testing.T) {
	if got := optional.FromPtr[int](nil); got.IsSet() {
		t.Fatal("FromPtr(nil) reports set")
	}
	n := 3
	got, ok := optional.FromPtr(&n).Get()
	if !ok || got != 3 {
		t.Fatalf("FromPtr(&3).Get() = (%v, %v), want (3, true)", got, ok)
	}
}

// TestPtrCopiesTheValue guards the aliasing hazard that the pointer spelling
// had: an artist holding the returned pointer must not observe later edits to
// the value it came from, and vice versa.
func TestPtrCopiesTheValue(t *testing.T) {
	v := optional.Of(2.5)
	p := v.Ptr()
	if p == nil || *p != 2.5 {
		t.Fatalf("Ptr() = %v, want pointer to 2.5", p)
	}
	*p = 9
	if got := v.OrZero(); got != 2.5 {
		t.Fatalf("mutating Ptr() changed the source Value to %v", got)
	}
	if second := v.Ptr(); second == p {
		t.Fatal("successive Ptr() calls alias one another")
	}
}

func TestValueIsComparable(t *testing.T) {
	type options struct {
		Alpha optional.Value[float64]
		Label optional.Value[string]
	}
	if (options{}) != (options{}) {
		t.Fatal("zero option structs compare unequal")
	}
	if (options{Alpha: optional.Of(1.0)}) == (options{}) {
		t.Fatal("a set option compares equal to an unset one")
	}
}
