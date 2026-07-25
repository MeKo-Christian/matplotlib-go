package optarg

import (
	"errors"
	"testing"
)

type sampleOptions struct {
	Label string
}

func TestOneReturnsZeroAndSingleValue(t *testing.T) {
	if got := One("sample", []sampleOptions(nil)); got != (sampleOptions{}) {
		t.Fatalf("One(empty) = %+v, want zero value", got)
	}
	want := sampleOptions{Label: "only"}
	if got := One("sample", []sampleOptions{want}); got != want {
		t.Fatalf("One(one) = %+v, want %+v", got, want)
	}
}

func TestOnePanicsOnExtraOptions(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("One(two) did not panic")
		}
		var tooManyErr *TooManyError
		err, ok := recovered.(error)
		if !ok || !errors.As(err, &tooManyErr) {
			t.Fatalf("One(two) panicked with %#v, want *TooManyError", recovered)
		}
		if tooManyErr.Call != "sample" || tooManyErr.Type != "sampleOptions" || tooManyErr.Count != 2 {
			t.Fatalf("unexpected error fields: %+v", tooManyErr)
		}
		const want = "sample accepts at most one sampleOptions value (got 2)"
		if tooManyErr.Error() != want {
			t.Fatalf("Error() = %q, want %q", tooManyErr.Error(), want)
		}
	}()
	One("sample", []sampleOptions{{Label: "a"}, {Label: "b"}})
}

func TestOnlyReportsExtraOptionsAsError(t *testing.T) {
	got, err := Only("sample", []sampleOptions{{Label: "a"}, {Label: "b"}})
	if err == nil {
		t.Fatalf("Only(two) returned no error")
	}
	if got != (sampleOptions{}) {
		t.Fatalf("Only(two) = %+v, want zero value", got)
	}
	var tooManyErr *TooManyError
	if !errors.As(err, &tooManyErr) {
		t.Fatalf("Only(two) error = %v, want *TooManyError", err)
	}

	if got, err := Only("sample", []sampleOptions(nil)); err != nil || got != (sampleOptions{}) {
		t.Fatalf("Only(empty) = (%+v, %v), want (zero, nil)", got, err)
	}
	want := sampleOptions{Label: "only"}
	if got, err := Only("sample", []sampleOptions{want}); err != nil || got != want {
		t.Fatalf("Only(one) = (%+v, %v), want (%+v, nil)", got, err, want)
	}
}

// TestNamesUnnamedOptionTypes covers option parameters whose element type has
// no declared name, so the message falls back to the generic wording.
func TestNamesUnnamedOptionTypes(t *testing.T) {
	_, err := Only("sample", []struct{ A int }{{A: 1}, {A: 2}})
	if err == nil {
		t.Fatalf("Only(two) returned no error")
	}
	const want = "sample accepts at most one option value (got 2)"
	if err.Error() != want {
		t.Fatalf("Error() = %q, want %q", err.Error(), want)
	}
}
