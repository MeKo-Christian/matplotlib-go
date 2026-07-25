// Package optional provides the tri-state value used by every plotting option
// struct.
//
// Matplotlib keyword arguments distinguish three states: absent (fall back to
// the rc default), present with a value, and present with a value that happens
// to equal Go's zero value. A plain struct field collapses the first and third
// states, so this port historically spelled optional fields as pointers
// (*float64, *bool, *string) or leaned on a magic zero — "Alpha == 0 means
// unset", "ArrowWidth <= 0 means unset". Both spellings lose information:
// the pointer form needs an addressable temporary at every call site and shares
// mutable state between the caller and the artist, and the magic-zero form
// makes legitimate values such as alpha 0, line width 0, or an annotation
// offset of (0, 0) impossible to request.
//
// [Value] restores the three states in one comparable value whose zero value is
// "absent":
//
//	core.ImShowOptions{}                        // every field falls back
//	core.ImShowOptions{Alpha: optional.Of(0.0)} // explicitly fully transparent
//
// Resolving an option against its default is then a single expression, and no
// separate "was anything supplied?" flag has to be threaded alongside it:
//
//	alpha := opt.Alpha.Or(1)
//
// See docs/plans/phase2-options-model.md for how Value fits into the wider
// options model.
package optional

// Value holds a value that may or may not have been set.
//
// The zero Value is unset, so an options struct literal that omits a field
// requests the default for it. Value is comparable whenever T is comparable,
// and it is copied by value, so an option struct never shares storage with the
// artist built from it.
type Value[T any] struct {
	value T
	set   bool
}

// Of returns a Value that is set to v.
//
// Of(zero) is deliberately distinct from the unset Value: it requests the zero
// value rather than the default.
func Of[T any](v T) Value[T] {
	return Value[T]{value: v, set: true}
}

// None returns an unset Value. It is identical to the zero Value[T] and exists
// for call sites that want to say so explicitly.
func None[T any]() Value[T] {
	return Value[T]{}
}

// FromPtr converts the pointer spelling of an optional value, treating nil as
// unset. It bridges call sites and artist fields that still hold pointers.
func FromPtr[T any](p *T) Value[T] {
	if p == nil {
		return Value[T]{}
	}
	return Value[T]{value: *p, set: true}
}

// IsSet reports whether the value was set.
func (o Value[T]) IsSet() bool { return o.set }

// Get returns the value and whether it was set. The value is the zero T when it
// was not.
func (o Value[T]) Get() (T, bool) { return o.value, o.set }

// Or returns the value when it was set, and fallback otherwise.
func (o Value[T]) Or(fallback T) T {
	if !o.set {
		return fallback
	}
	return o.value
}

// OrZero returns the value when it was set, and the zero T otherwise.
func (o Value[T]) OrZero() T { return o.value }

// Ptr returns a pointer to a copy of the value, or nil when it was not set.
//
// It bridges to artist fields and renderer structs that still spell optional
// values as pointers. Because the copy is fresh, mutating the result never
// reaches back into the options that produced it.
func (o Value[T]) Ptr() *T {
	if !o.set {
		return nil
	}
	v := o.value
	return &v
}
