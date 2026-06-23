// Package cycler is a small, faithful port of Matplotlib's "cycler" library
// (https://matplotlib.org/cycler/). A Cycler is a finite, ordered sequence of
// property maps. It backs Matplotlib's axes.prop_cycle, which combines artist
// properties such as color, linestyle, marker and linewidth so that successive
// plotted artists draw their styling from one shared, repeating cycle.
//
// Values are stored as any: color entries hold render.Color, linestyle/marker
// entries hold string, and linewidth entries hold float64. Consumers
// type-assert per key.
package cycler

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

// Cycler is an ordered, finite sequence of property maps sharing the same keys.
//
// The zero value is an empty cycle; use New to construct one.
type Cycler struct {
	keys []string
	rows []map[string]any
}

// New builds a single-key cycler from one property name and its values, mirroring
// Matplotlib's cycler(key, values). A cycler with no values is valid but empty.
func New(key string, values ...any) *Cycler {
	rows := make([]map[string]any, len(values))
	for i, v := range values {
		rows[i] = map[string]any{key: v}
	}
	return &Cycler{keys: []string{key}, rows: rows}
}

// Len reports the number of entries (one repetition of the cycle).
func (c *Cycler) Len() int {
	if c == nil {
		return 0
	}
	return len(c.rows)
}

// Keys returns the property names in insertion order.
func (c *Cycler) Keys() []string {
	if c == nil {
		return nil
	}
	return append([]string(nil), c.keys...)
}

// Has reports whether the cycler carries the named property.
func (c *Cycler) Has(key string) bool {
	if c == nil {
		return false
	}
	return slices.Contains(c.keys, key)
}

// Row returns a copy of the i-th entry, wrapping modulo Len so any index is
// valid for a non-empty cycler. It returns nil for an empty or nil cycler.
func (c *Cycler) Row(i int) map[string]any {
	if c.Len() == 0 {
		return nil
	}
	idx := i % len(c.rows)
	if idx < 0 {
		idx += len(c.rows)
	}
	return maps.Clone(c.rows[idx])
}

// ByKey returns the values of one property in cycle order, matching
// Matplotlib's Cycler.by_key()[key]. It returns nil when the key is absent.
func (c *Cycler) ByKey(key string) []any {
	if !c.Has(key) {
		return nil
	}
	out := make([]any, len(c.rows))
	for i, row := range c.rows {
		out[i] = row[key]
	}
	return out
}

// Concat implements the "+" operator: it zips two equal-length cyclers with
// disjoint key sets into one cycler of the same length, merging each pair of
// entries. It mirrors Matplotlib, which raises ValueError on a length mismatch
// or on overlapping keys.
func (c *Cycler) Concat(other *Cycler) (*Cycler, error) {
	if c.Len() != other.Len() {
		return nil, fmt.Errorf("cycler: can only add equal-length cyclers (%d and %d)", c.Len(), other.Len())
	}
	if err := disjoint(c, other); err != nil {
		return nil, err
	}
	rows := make([]map[string]any, c.Len())
	for i := range rows {
		merged := make(map[string]any, len(c.keys)+len(other.keys))
		maps.Copy(merged, c.rows[i])
		maps.Copy(merged, other.rows[i])
		rows[i] = merged
	}
	return &Cycler{keys: mergeKeys(c.keys, other.keys), rows: rows}, nil
}

// Multiply implements the "*" operator: the outer product of two disjoint-key
// cyclers. The left cycler varies slowest, matching Matplotlib's ordering, so
// cycler(color=[r,g]) * cycler(ls=[-,--]) yields (r,-),(r,--),(g,-),(g,--).
func (c *Cycler) Multiply(other *Cycler) (*Cycler, error) {
	if err := disjoint(c, other); err != nil {
		return nil, err
	}
	rows := make([]map[string]any, 0, c.Len()*other.Len())
	for _, left := range c.rows {
		for _, right := range other.rows {
			merged := make(map[string]any, len(left)+len(right))
			maps.Copy(merged, left)
			maps.Copy(merged, right)
			rows = append(rows, merged)
		}
	}
	return &Cycler{keys: mergeKeys(c.keys, other.keys), rows: rows}, nil
}

// Clone returns a deep copy of the cycler. Property values are copied by
// assignment (the stored color/string/float64 values are themselves immutable).
func (c *Cycler) Clone() *Cycler {
	if c == nil {
		return nil
	}
	rows := make([]map[string]any, len(c.rows))
	for i, row := range c.rows {
		rows[i] = maps.Clone(row)
	}
	return &Cycler{keys: append([]string(nil), c.keys...), rows: rows}
}

func disjoint(a, b *Cycler) error {
	for _, k := range a.Keys() {
		if b.Has(k) {
			return fmt.Errorf("cycler: cannot combine cyclers with overlapping key %q", k)
		}
	}
	return nil
}

func mergeKeys(a, b []string) []string {
	out := make([]string, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

// ErrEmpty is returned by helpers that require a non-empty cycler.
var ErrEmpty = errors.New("cycler: empty cycler")
