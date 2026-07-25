// Package optarg centralizes how the plotting API unpacks its variadic option
// parameters.
//
// Almost every plotting entry point takes its options as a variadic tail
// (`opts ...FooOptions`) so that callers can omit them entirely. The tail is a
// stand-in for an optional argument, so exactly zero or one value is
// meaningful; passing two or more is an API misuse that silently discarded
// every value after the first.
//
// Phase 2.3 of the API rework makes that misuse impossible or rejected. Until
// the final options representation lands, this package is the single place
// where the "at most one" rule is enforced, so every entry point reports it the
// same way:
//
//   - [One] serves calls whose signature has no error result. An extra option
//     value cannot come from data — it has to be written literally at the call
//     site — so it is a bug in the caller, and One panics.
//   - [Optional] is the same rule for calls that merge the supplied options into
//     a prepared default value and so must distinguish "absent" from "zero".
//   - [Only] serves calls that already report rejected input as an error, and
//     keeps the rejection on that same channel.
package optarg

import (
	"fmt"
	"reflect"
)

// TooManyError reports a call that supplied more than one option value.
type TooManyError struct {
	// Call names the entry point as a user would write it, e.g. "imshow".
	Call string
	// Type is the option struct's name without its package qualifier,
	// e.g. "ImShowOptions". It is empty when the name cannot be determined.
	Type string
	// Count is the number of option values the caller supplied.
	Count int
}

func (e *TooManyError) Error() string {
	if e.Type == "" {
		return fmt.Sprintf("%s accepts at most one option value (got %d)", e.Call, e.Count)
	}
	return fmt.Sprintf("%s accepts at most one %s value (got %d)", e.Call, e.Type, e.Count)
}

// One returns the single option value in values, or the zero value of T when
// values is empty.
//
// It panics with a [*TooManyError] when values holds more than one entry. Such
// a call cannot arise from plot data; it requires the caller to have written
// two option literals in one call, which the final options model will reject at
// compile time.
func One[T any](call string, values []T) T {
	switch len(values) {
	case 0:
		var zero T
		return zero
	case 1:
		return values[0]
	default:
		panic(tooMany[T](call, len(values)))
	}
}

// Optional reports whether the caller supplied an option value, and returns it
// when they did. It panics with a [*TooManyError] on the same condition as
// [One].
//
// Use it where the entry point merges the caller's options into a prepared
// default value: the zero value of T is a meaningful input there, so "absent"
// has to stay distinguishable from "zero".
func Optional[T any](call string, values []T) (T, bool) {
	switch len(values) {
	case 0:
		var zero T
		return zero, false
	case 1:
		return values[0], true
	default:
		panic(tooMany[T](call, len(values)))
	}
}

// Only returns the single option value in values, or the zero value of T when
// values is empty. It reports a [*TooManyError] when values holds more than one
// entry.
//
// Use it from entry points that already return an error, so that extra option
// values reach the caller on the same channel as every other rejected input.
func Only[T any](call string, values []T) (T, error) {
	var zero T
	switch len(values) {
	case 0:
		return zero, nil
	case 1:
		return values[0], nil
	default:
		return zero, tooMany[T](call, len(values))
	}
}

// tooMany builds the shared rejection, naming the option type through
// reflection so callers do not have to repeat it.
func tooMany[T any](call string, count int) *TooManyError {
	return &TooManyError{Call: call, Type: reflect.TypeFor[T]().Name(), Count: count}
}
