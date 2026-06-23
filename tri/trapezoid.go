package tri

import (
	"fmt"
	"math"
	"math/rand"
)

// TrapezoidMapTriFinder locates points using the trapezoid-map algorithm of
// de Berg et al., a faithful port of matplotlib's C++ TrapezoidMapTriFinder
// (src/tri/_tri.cpp). Construction is O(n log n) expected; queries are
// O(log n) expected. It returns the same containing triangle as the
// brute-force finder for interior points and -1 outside the mesh.
//
// The C++ implementation manages node lifetimes with reference counting; this
// port relies on Go's garbage collector and keeps only the parent links needed
// to splice replacement subtrees into the search tree.
type TrapezoidMapTriFinder struct {
	tri      Triangulation
	points   []tmPoint
	edges    []*tmEdge
	tree     *tmNode
	fallback *BruteForceTriFinder // used if the map could not be built
}

// NewTrapezoidMapTriFinder builds the search structure for t. If the
// triangulation cannot be processed (e.g. it is geometrically invalid), the
// finder transparently falls back to brute-force point location.
func NewTrapezoidMapTriFinder(t Triangulation) *TrapezoidMapTriFinder {
	f := &TrapezoidMapTriFinder{tri: t}
	if err := f.initialize(); err != nil {
		f.tree = nil
		f.fallback = NewBruteForceTriFinder(t)
	}
	return f
}

// Find returns the index of the triangle containing (x, y), or -1.
func (f *TrapezoidMapTriFinder) Find(x, y float64) int {
	if f.fallback != nil {
		return f.fallback.Find(x, y)
	}
	node := f.tree.search(x, y)
	if node == nil {
		return -1
	}
	return node.getTri()
}

// --- geometry primitives (mirroring XY in _tri.h) ---

func isRightOf(ax, ay, bx, by float64) bool {
	if ax == bx {
		return ay > by
	}
	return ax > bx
}

func pointEqual(ax, ay, bx, by float64) bool { return ax == bx && ay == by }

// tmPoint is a triangulation point plus the index of an incident triangle.
type tmPoint struct {
	x, y float64
	tri  int
}

func (p *tmPoint) rightOf(q *tmPoint) bool { return isRightOf(p.x, p.y, q.x, q.y) }

// tmEdge is a directed edge (left -> right, with right to the right of left).
type tmEdge struct {
	left, right            *tmPoint
	triangleBelow          int
	triangleAbove          int
	pointBelow, pointAbove *tmPoint
}

// getPointOrientation returns +1 if (x,y) is left of the edge (above), -1 if
// right (below), 0 if on the line. Sign matches matplotlib: cross_z of
// (xy-left) with (right-left).
func (e *tmEdge) getPointOrientation(x, y float64) int {
	cz := (x-e.left.x)*(e.right.y-e.left.y) - (y-e.left.y)*(e.right.x-e.left.x)
	switch {
	case cz > 0:
		return +1
	case cz < 0:
		return -1
	default:
		return 0
	}
}

func (e *tmEdge) getSlope() float64 {
	return (e.right.y - e.left.y) / (e.right.x - e.left.x)
}

func (e *tmEdge) hasPoint(p *tmPoint) bool { return e.left == p || e.right == p }

// --- search tree nodes ---

type nodeKind int

const (
	xNode nodeKind = iota
	yNode
	trapNode
)

type tmNode struct {
	kind nodeKind

	// xNode: split by point x (left < point, right >= point).
	point       *tmPoint
	left, right *tmNode

	// yNode: split by edge (below/above).
	edge         *tmEdge
	below, above *tmNode

	// trapNode
	trapezoid *tmTrapezoid

	parents []*tmNode
}

func newXNode(point *tmPoint, left, right *tmNode) *tmNode {
	n := &tmNode{kind: xNode, point: point, left: left, right: right}
	left.addParent(n)
	right.addParent(n)
	return n
}

func newYNode(edge *tmEdge, below, above *tmNode) *tmNode {
	n := &tmNode{kind: yNode, edge: edge, below: below, above: above}
	below.addParent(n)
	above.addParent(n)
	return n
}

func newTrapNode(t *tmTrapezoid) *tmNode {
	n := &tmNode{kind: trapNode, trapezoid: t}
	t.node = n
	return n
}

func (n *tmNode) addParent(p *tmNode) { n.parents = append(n.parents, p) }

func (n *tmNode) removeParent(p *tmNode) {
	for i, q := range n.parents {
		if q == p {
			n.parents = append(n.parents[:i], n.parents[i+1:]...)
			return
		}
	}
}

func (n *tmNode) replaceChild(oldChild, newChild *tmNode) {
	switch n.kind {
	case xNode:
		if n.left == oldChild {
			n.left = newChild
		} else {
			n.right = newChild
		}
	case yNode:
		if n.below == oldChild {
			n.below = newChild
		} else {
			n.above = newChild
		}
	}
	oldChild.removeParent(n)
	newChild.addParent(n)
}

func (n *tmNode) replaceWith(newNode *tmNode) {
	for len(n.parents) > 0 {
		n.parents[0].replaceChild(n, newNode)
	}
}

func (n *tmNode) getTri() int {
	switch n.kind {
	case xNode:
		return n.point.tri
	case yNode:
		if n.edge.triangleAbove != -1 {
			return n.edge.triangleAbove
		}
		return n.edge.triangleBelow
	default:
		return n.trapezoid.below.triangleAbove
	}
}

// search locates the leaf node whose trapezoid contains (x, y).
func (n *tmNode) search(x, y float64) *tmNode {
	switch n.kind {
	case xNode:
		if pointEqual(x, y, n.point.x, n.point.y) {
			return n
		}
		if isRightOf(x, y, n.point.x, n.point.y) {
			return n.right.search(x, y)
		}
		return n.left.search(x, y)
	case yNode:
		orient := n.edge.getPointOrientation(x, y)
		if orient == 0 {
			return n
		}
		if orient < 0 {
			return n.above.search(x, y)
		}
		return n.below.search(x, y)
	default:
		return n
	}
}

// searchEdge locates the trapezoid containing the left endpoint of edge, used
// when inserting edges. Mirrors Node::search(const Edge&).
func (n *tmNode) searchEdge(edge *tmEdge) (*tmTrapezoid, error) {
	switch n.kind {
	case xNode:
		if edge.left == n.point {
			return n.right.searchEdge(edge)
		}
		if edge.left.rightOf(n.point) {
			return n.right.searchEdge(edge)
		}
		return n.left.searchEdge(edge)
	case yNode:
		switch {
		case edge.left == n.edge.left:
			// Coinciding left edge points.
			if edge.getSlope() == n.edge.getSlope() {
				if n.edge.triangleAbove == edge.triangleBelow {
					return n.above.searchEdge(edge)
				} else if n.edge.triangleBelow == edge.triangleAbove {
					return n.below.searchEdge(edge)
				}
				return nil, fmt.Errorf("invalid triangulation, common left points")
			}
			if edge.getSlope() > n.edge.getSlope() {
				return n.above.searchEdge(edge)
			}
			return n.below.searchEdge(edge)
		case edge.right == n.edge.right:
			// Coinciding right edge points.
			if edge.getSlope() == n.edge.getSlope() {
				if n.edge.triangleAbove == edge.triangleBelow {
					return n.above.searchEdge(edge)
				} else if n.edge.triangleBelow == edge.triangleAbove {
					return n.below.searchEdge(edge)
				}
				return nil, fmt.Errorf("invalid triangulation, common right points")
			}
			if edge.getSlope() > n.edge.getSlope() {
				return n.below.searchEdge(edge)
			}
			return n.above.searchEdge(edge)
		default:
			orient := n.edge.getPointOrientation(edge.left.x, edge.left.y)
			if orient == 0 {
				switch {
				case n.edge.pointAbove != nil && edge.hasPoint(n.edge.pointAbove):
					orient = -1
				case n.edge.pointBelow != nil && edge.hasPoint(n.edge.pointBelow):
					orient = +1
				default:
					return nil, fmt.Errorf("invalid triangulation, point on edge")
				}
			}
			if orient < 0 {
				return n.above.searchEdge(edge)
			}
			return n.below.searchEdge(edge)
		}
	default:
		return n.trapezoid, nil
	}
}

// --- trapezoids ---

type tmTrapezoid struct {
	left, right  *tmPoint
	below, above *tmEdge

	lowerLeft, lowerRight *tmTrapezoid
	upperLeft, upperRight *tmTrapezoid

	node *tmNode
}

func newTrapezoid(left, right *tmPoint, below, above *tmEdge) *tmTrapezoid {
	return &tmTrapezoid{left: left, right: right, below: below, above: above}
}

func (t *tmTrapezoid) setLowerLeft(o *tmTrapezoid) {
	t.lowerLeft = o
	if o != nil {
		o.lowerRight = t
	}
}

func (t *tmTrapezoid) setLowerRight(o *tmTrapezoid) {
	t.lowerRight = o
	if o != nil {
		o.lowerLeft = t
	}
}

func (t *tmTrapezoid) setUpperLeft(o *tmTrapezoid) {
	t.upperLeft = o
	if o != nil {
		o.upperRight = t
	}
}

func (t *tmTrapezoid) setUpperRight(o *tmTrapezoid) {
	t.upperRight = o
	if o != nil {
		o.upperLeft = t
	}
}

// --- construction ---

// tmNeighbor records, for a triangle edge, the adjacent (unmasked) triangle and
// the matching local edge index within it.
type tmNeighbor struct {
	tri  int
	edge int
}

func (f *TrapezoidMapTriFinder) initialize() error {
	t := f.tri
	npoints := len(t.X)
	if npoints == 0 || len(t.Triangles) == 0 {
		return fmt.Errorf("empty triangulation")
	}

	f.points = make([]tmPoint, npoints+4)
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for i := 0; i < npoints; i++ {
		x, y := t.X[i], t.Y[i]
		if x == 0 {
			x = 0 // normalise -0.0
		}
		if y == 0 {
			y = 0
		}
		f.points[i] = tmPoint{x: x, y: y, tri: -1}
		minX, maxX = math.Min(minX, x), math.Max(maxX, x)
		minY, maxY = math.Min(minY, y), math.Max(maxY, y)
	}

	// Enclosing rectangle, expanded slightly so real points sit strictly inside.
	if minX > maxX {
		minX, minY, maxX, maxY = 0, 0, 1, 1
	} else {
		const small = 0.1
		dx := (maxX - minX) * small
		dy := (maxY - minY) * small
		minX, maxX = minX-dx, maxX+dx
		minY, maxY = minY-dy, maxY+dy
	}
	sw := npoints
	se := npoints + 1
	nw := npoints + 2
	ne := npoints + 3
	f.points[sw] = tmPoint{x: minX, y: minY, tri: -1}
	f.points[se] = tmPoint{x: maxX, y: minY, tri: -1}
	f.points[nw] = tmPoint{x: minX, y: maxY, tri: -1}
	f.points[ne] = tmPoint{x: maxX, y: maxY, tri: -1}

	neighbors := f.computeNeighbors()

	// Bottom and top edges of the enclosing rectangle.
	f.edges = f.edges[:0]
	f.edges = append(
		f.edges,
		&tmEdge{left: &f.points[sw], right: &f.points[se], triangleBelow: -1, triangleAbove: -1},
		&tmEdge{left: &f.points[nw], right: &f.points[ne], triangleBelow: -1, triangleAbove: -1},
	)

	for triIdx, triangle := range t.Triangles {
		if t.Masked(triIdx) {
			continue
		}
		for e := 0; e < 3; e++ {
			startIdx := triangle[e]
			endIdx := triangle[(e+1)%3]
			otherIdx := triangle[(e+2)%3]
			start := &f.points[startIdx]
			end := &f.points[endIdx]
			other := &f.points[otherIdx]
			nb := neighbors[triIdx][e]

			if end.rightOf(start) {
				var pointBelow *tmPoint
				if nb.tri != -1 {
					pointBelow = &f.points[t.Triangles[nb.tri][(nb.edge+2)%3]]
				}
				f.edges = append(f.edges, &tmEdge{
					left: start, right: end,
					triangleBelow: nb.tri, triangleAbove: triIdx,
					pointBelow: pointBelow, pointAbove: other,
				})
			} else if nb.tri == -1 {
				f.edges = append(f.edges, &tmEdge{
					left: end, right: start,
					triangleBelow: triIdx, triangleAbove: -1,
					pointBelow: other, pointAbove: nil,
				})
			}

			if start.tri == -1 {
				start.tri = triIdx
			}
		}
	}

	// Initial trapezoid spans the enclosing rectangle.
	f.tree = newTrapNode(newTrapezoid(&f.points[sw], &f.points[se], f.edges[0], f.edges[1]))

	// Shuffle all but the first two edges deterministically. Query results are
	// independent of insertion order; the shuffle only balances the tree.
	rng := rand.New(rand.NewSource(1234))
	rng.Shuffle(len(f.edges)-2, func(i, j int) {
		f.edges[i+2], f.edges[j+2] = f.edges[j+2], f.edges[i+2]
	})

	for i := 2; i < len(f.edges); i++ {
		if err := f.addEdgeToTree(f.edges[i]); err != nil {
			return err
		}
	}
	return nil
}

// computeNeighbors derives, for every triangle edge (start->next), the adjacent
// unmasked triangle and its matching local edge. Masked neighbours are treated
// as absent (-1), matching the brute-force finder. Assumes anticlockwise
// winding so the directed edge (u,v) is shared as (v,u) by the neighbour.
func (f *TrapezoidMapTriFinder) computeNeighbors() [][3]tmNeighbor {
	t := f.tri
	directed := make(map[[2]int]tmNeighbor, len(t.Triangles)*3)
	for triIdx, triangle := range t.Triangles {
		if t.Masked(triIdx) {
			continue
		}
		for e := 0; e < 3; e++ {
			directed[[2]int{triangle[e], triangle[(e+1)%3]}] = tmNeighbor{tri: triIdx, edge: e}
		}
	}
	out := make([][3]tmNeighbor, len(t.Triangles))
	for triIdx, triangle := range t.Triangles {
		out[triIdx] = [3]tmNeighbor{{-1, -1}, {-1, -1}, {-1, -1}}
		if t.Masked(triIdx) {
			continue
		}
		for e := 0; e < 3; e++ {
			u, v := triangle[e], triangle[(e+1)%3]
			if l, ok := directed[[2]int{v, u}]; ok {
				out[triIdx][e] = l
			}
		}
	}
	return out
}

func (f *TrapezoidMapTriFinder) addEdgeToTree(edge *tmEdge) error {
	trapezoids, err := f.findTrapezoidsIntersectingEdge(edge)
	if err != nil {
		return err
	}
	if len(trapezoids) == 0 {
		return fmt.Errorf("no trapezoids intersect edge")
	}

	p := edge.left
	q := edge.right
	var leftOld, leftBelow, leftAbove *tmTrapezoid

	ntraps := len(trapezoids)
	for i := 0; i < ntraps; i++ {
		old := trapezoids[i]
		startTrap := i == 0
		endTrap := i == ntraps-1
		haveLeft := startTrap && edge.left != old.left
		haveRight := endTrap && edge.right != old.right

		var left, below, above, right *tmTrapezoid

		switch {
		case startTrap && endTrap:
			if haveLeft {
				left = newTrapezoid(old.left, p, old.below, old.above)
			}
			below = newTrapezoid(p, q, old.below, edge)
			above = newTrapezoid(p, q, edge, old.above)
			if haveRight {
				right = newTrapezoid(q, old.right, old.below, old.above)
			}

			if haveLeft {
				left.setLowerLeft(old.lowerLeft)
				left.setUpperLeft(old.upperLeft)
				left.setLowerRight(below)
				left.setUpperRight(above)
			} else {
				below.setLowerLeft(old.lowerLeft)
				above.setUpperLeft(old.upperLeft)
			}

			if haveRight {
				right.setLowerRight(old.lowerRight)
				right.setUpperRight(old.upperRight)
				below.setLowerRight(right)
				above.setUpperRight(right)
			} else {
				below.setLowerRight(old.lowerRight)
				above.setUpperRight(old.upperRight)
			}
		case startTrap:
			if haveLeft {
				left = newTrapezoid(old.left, p, old.below, old.above)
			}
			below = newTrapezoid(p, old.right, old.below, edge)
			above = newTrapezoid(p, old.right, edge, old.above)

			if haveLeft {
				left.setLowerLeft(old.lowerLeft)
				left.setUpperLeft(old.upperLeft)
				left.setLowerRight(below)
				left.setUpperRight(above)
			} else {
				below.setLowerLeft(old.lowerLeft)
				above.setUpperLeft(old.upperLeft)
			}

			below.setLowerRight(old.lowerRight)
			above.setUpperRight(old.upperRight)
		case endTrap:
			if leftBelow.below == old.below {
				below = leftBelow
				below.right = q
			} else {
				below = newTrapezoid(old.left, q, old.below, edge)
			}

			if leftAbove.above == old.above {
				above = leftAbove
				above.right = q
			} else {
				above = newTrapezoid(old.left, q, edge, old.above)
			}

			if haveRight {
				right = newTrapezoid(q, old.right, old.below, old.above)
			}

			if haveRight {
				right.setLowerRight(old.lowerRight)
				right.setUpperRight(old.upperRight)
				below.setLowerRight(right)
				above.setUpperRight(right)
			} else {
				below.setLowerRight(old.lowerRight)
				above.setUpperRight(old.upperRight)
			}

			if below != leftBelow {
				below.setUpperLeft(leftBelow)
				if old.lowerLeft == leftOld {
					below.setLowerLeft(leftBelow)
				} else {
					below.setLowerLeft(old.lowerLeft)
				}
			}

			if above != leftAbove {
				above.setLowerLeft(leftAbove)
				if old.upperLeft == leftOld {
					above.setUpperLeft(leftAbove)
				} else {
					above.setUpperLeft(old.upperLeft)
				}
			}
		default: // middle trapezoid
			if leftBelow.below == old.below {
				below = leftBelow
				below.right = old.right
			} else {
				below = newTrapezoid(old.left, old.right, old.below, edge)
			}

			if leftAbove.above == old.above {
				above = leftAbove
				above.right = old.right
			} else {
				above = newTrapezoid(old.left, old.right, edge, old.above)
			}

			if below != leftBelow {
				below.setUpperLeft(leftBelow)
				if old.lowerLeft == leftOld {
					below.setLowerLeft(leftBelow)
				} else {
					below.setLowerLeft(old.lowerLeft)
				}
			}

			if above != leftAbove {
				above.setLowerLeft(leftAbove)
				if old.upperLeft == leftOld {
					above.setUpperLeft(leftAbove)
				} else {
					above.setUpperLeft(old.upperLeft)
				}
			}

			below.setLowerRight(old.lowerRight)
			above.setUpperRight(old.upperRight)
		}

		// Build replacement subtree. Reuse owning trapezoid nodes where below or
		// above are carried over from the previous (left) trapezoid.
		var belowNode, aboveNode *tmNode
		if below == leftBelow {
			belowNode = below.node
		} else {
			belowNode = newTrapNode(below)
		}
		if above == leftAbove {
			aboveNode = above.node
		} else {
			aboveNode = newTrapNode(above)
		}
		newTop := newYNode(edge, belowNode, aboveNode)
		if haveRight {
			newTop = newXNode(q, newTop, newTrapNode(right))
		}
		if haveLeft {
			newTop = newXNode(p, newTrapNode(left), newTop)
		}

		oldNode := old.node
		if oldNode == f.tree {
			f.tree = newTop
		} else {
			oldNode.replaceWith(newTop)
		}

		if !endTrap {
			leftOld = old
			leftAbove = above
			leftBelow = below
		}
	}
	return nil
}

func (f *TrapezoidMapTriFinder) findTrapezoidsIntersectingEdge(edge *tmEdge) ([]*tmTrapezoid, error) {
	trapezoid, err := f.tree.searchEdge(edge)
	if err != nil {
		return nil, err
	}
	if trapezoid == nil {
		return nil, fmt.Errorf("searchEdge returned nil trapezoid")
	}

	trapezoids := []*tmTrapezoid{trapezoid}
	for edge.right.rightOf(trapezoid.right) {
		orient := edge.getPointOrientation(trapezoid.right.x, trapezoid.right.y)
		if orient == 0 {
			switch {
			case edge.pointBelow == trapezoid.right:
				orient = +1
			case edge.pointAbove == trapezoid.right:
				orient = -1
			default:
				return nil, fmt.Errorf("unable to deal with point on edge")
			}
		}

		switch orient {
		case -1:
			trapezoid = trapezoid.lowerRight
		case +1:
			trapezoid = trapezoid.upperRight
		}

		if trapezoid == nil {
			return nil, fmt.Errorf("expected trapezoid neighbour")
		}
		trapezoids = append(trapezoids, trapezoid)
	}
	return trapezoids, nil
}
