package transform

import "github.com/cwbudde/matplotlib-go/geom"

// Invalidation identifies which portion of a transform dependency is stale.
//
// Matplotlib distinguishes affine and non-affine invalidation so expensive
// projection stages can be cached independently from cheap display-affine
// updates. The Go port keeps that split explicit without adopting the full
// upstream transform class hierarchy.
type Invalidation uint8

const (
	InvalidNone   Invalidation = 0
	InvalidAffine Invalidation = 1 << iota
	InvalidNonAffine
)

// InvalidAll marks both affine and non-affine transform stages stale.
const InvalidAll = InvalidAffine | InvalidNonAffine

// Has reports whether the invalidation includes the requested stage.
func (i Invalidation) Has(stage Invalidation) bool {
	return i&stage != 0
}

// TransformNode is a small invalidation node for cache-friendly transform
// compositions.
//
// A node may have dependents. Invalidating a node marks it stale and propagates
// the same stale stage to each dependent. The type is intentionally small and
// zero-value ready so existing value-style transforms can opt in only where
// shared dynamic transform chains need cache invalidation.
type TransformNode struct {
	invalid    Invalidation
	version    uint64
	dependents map[*TransformNode]struct{}
}

// AddDependent registers a downstream transform node that should become stale
// when this node is invalidated.
func (n *TransformNode) AddDependent(dependent *TransformNode) {
	if n == nil || dependent == nil || dependent == n {
		return
	}
	if n.dependents == nil {
		n.dependents = make(map[*TransformNode]struct{})
	}
	n.dependents[dependent] = struct{}{}
}

// RemoveDependent unregisters a downstream transform node.
func (n *TransformNode) RemoveDependent(dependent *TransformNode) {
	if n == nil || dependent == nil || n.dependents == nil {
		return
	}
	delete(n.dependents, dependent)
}

// Invalidate marks this node and its dependents stale for the requested stage.
func (n *TransformNode) Invalidate(stage Invalidation) {
	if n == nil || stage == InvalidNone {
		return
	}
	n.invalidate(stage, make(map[*TransformNode]struct{}))
}

func (n *TransformNode) invalidate(stage Invalidation, seen map[*TransformNode]struct{}) {
	if n == nil {
		return
	}
	if _, ok := seen[n]; ok {
		return
	}
	seen[n] = struct{}{}

	n.invalid |= stage
	n.version++
	for dependent := range n.dependents {
		dependent.invalidate(stage, seen)
	}
}

// ClearInvalid marks this node's current cached value valid.
func (n *TransformNode) ClearInvalid() {
	if n == nil {
		return
	}
	n.invalid = InvalidNone
}

// Invalid reports which stages are stale on this node.
func (n *TransformNode) Invalid() Invalidation {
	if n == nil {
		return InvalidNone
	}
	return n.invalid
}

// Version reports a monotonic invalidation counter for cache bookkeeping.
func (n *TransformNode) Version() uint64 {
	if n == nil {
		return 0
	}
	return n.version
}

// CachedTransform wraps a transform builder with TransformNode invalidation.
// It satisfies T by rebuilding the wrapped transform only when its node is
// invalid or when no transform has been built yet.
type CachedTransform struct {
	TransformNode
	build         func() T
	cached        T
	cachedVersion uint64
}

// NewCachedTransform creates a cached transform and connects it to optional
// dependency nodes.
func NewCachedTransform(build func() T, dependencies ...*TransformNode) *CachedTransform {
	c := &CachedTransform{build: build}
	for _, dependency := range dependencies {
		if dependency != nil {
			dependency.AddDependent(&c.TransformNode)
		}
	}
	return c
}

// Current returns the current built transform, rebuilding it if stale.
func (c *CachedTransform) Current() T {
	if c == nil {
		return nil
	}
	if c.cached == nil || c.Invalid() != InvalidNone {
		if c.build != nil {
			c.cached = c.build()
		} else {
			c.cached = nil
		}
		c.cachedVersion = c.Version()
		c.ClearInvalid()
	}
	return c.cached
}

// CachedVersion reports the source node version used for the current cache.
func (c *CachedTransform) CachedVersion() uint64 {
	if c == nil {
		return 0
	}
	return c.cachedVersion
}

// Apply transforms a point through the current cached transform.
func (c *CachedTransform) Apply(p geom.Pt) geom.Pt {
	tr := c.Current()
	if tr == nil {
		return p
	}
	return tr.Apply(p)
}

// Invert transforms a point back through the current cached transform.
func (c *CachedTransform) Invert(p geom.Pt) (geom.Pt, bool) {
	tr := c.Current()
	if tr == nil {
		return p, true
	}
	return tr.Invert(p)
}
