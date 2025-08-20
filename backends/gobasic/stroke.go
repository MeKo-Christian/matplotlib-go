package gobasic

import (
	"math"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// strokeToPath converts a stroked path to a filled path that represents the stroke.
// This allows us to use the vector rasterizer's fill capabilities for complex strokes.
func strokeToPath(p geom.Path, paint *render.Paint) geom.Path {
	if len(p.C) == 0 || paint.LineWidth <= 0 {
		return geom.Path{}
	}

	// Quantize the path for deterministic stroke calculation
	p = quantizePath(p)

	// Handle dashes first if present
	if len(paint.Dashes) > 0 {
		p = applyDashes(p, paint.Dashes)
	}

	// Convert each subpath to stroke polygons
	var result geom.Path
	subpaths := splitIntoSubpaths(p)

	for _, subpath := range subpaths {
		strokePath := strokeSubpath(subpath, paint)
		result = appendPath(result, strokePath)
	}

	// Quantize the final stroke path
	return quantizePath(result)
}

// splitIntoSubpaths breaks a path into individual subpaths separated by MoveTo commands.
func splitIntoSubpaths(p geom.Path) []geom.Path {
	var subpaths []geom.Path
	var currentSubpath geom.Path

	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			// Start new subpath
			if len(currentSubpath.C) > 0 {
				subpaths = append(subpaths, currentSubpath)
			}
			currentSubpath = geom.Path{
				C: []geom.Cmd{geom.MoveTo},
				V: []geom.Pt{p.V[vi]},
			}
			vi++
		case geom.LineTo:
			currentSubpath.C = append(currentSubpath.C, geom.LineTo)
			currentSubpath.V = append(currentSubpath.V, p.V[vi])
			vi++
		case geom.QuadTo:
			currentSubpath.C = append(currentSubpath.C, geom.QuadTo)
			currentSubpath.V = append(currentSubpath.V, p.V[vi], p.V[vi+1])
			vi += 2
		case geom.CubicTo:
			currentSubpath.C = append(currentSubpath.C, geom.CubicTo)
			currentSubpath.V = append(currentSubpath.V, p.V[vi], p.V[vi+1], p.V[vi+2])
			vi += 3
		case geom.ClosePath:
			currentSubpath.C = append(currentSubpath.C, geom.ClosePath)
		}
	}

	if len(currentSubpath.C) > 0 {
		subpaths = append(subpaths, currentSubpath)
	}

	return subpaths
}

// strokeSubpath generates the stroke geometry for a single subpath.
func strokeSubpath(p geom.Path, paint *render.Paint) geom.Path {
	if len(p.C) == 0 || paint.LineWidth <= 0 {
		return geom.Path{}
	}

	// Convert to line segments (flatten curves for now)
	segments := pathToSegments(p)
	if len(segments) == 0 {
		return geom.Path{}
	}

	halfWidth := quantize(paint.LineWidth / 2.0)
	isClosed := len(p.C) > 0 && p.C[len(p.C)-1] == geom.ClosePath

	// Generate offset curves on both sides
	leftOffsets := make([]geom.Pt, len(segments)+1)
	rightOffsets := make([]geom.Pt, len(segments)+1)

	for i, seg := range segments {
		normal := segmentNormal(seg, halfWidth)
		leftOffsets[i] = quantizePt(geom.Pt{X: seg.Start.X + normal.X, Y: seg.Start.Y + normal.Y})
		rightOffsets[i] = quantizePt(geom.Pt{X: seg.Start.X - normal.X, Y: seg.Start.Y - normal.Y})

		// Handle the end point of the last segment
		if i == len(segments)-1 {
			leftOffsets[i+1] = quantizePt(geom.Pt{X: seg.End.X + normal.X, Y: seg.End.Y + normal.Y})
			rightOffsets[i+1] = quantizePt(geom.Pt{X: seg.End.X - normal.X, Y: seg.End.Y - normal.Y})
		}
	}

	// Apply line joins at interior vertices
	for i := 1; i < len(segments); i++ {
		prev := segments[i-1]
		curr := segments[i]

		leftJoin, rightJoin := calculateJoin(prev, curr, halfWidth, paint.LineJoin, paint.MiterLimit)
		leftOffsets[i] = leftJoin
		rightOffsets[i] = rightJoin
	}

	// Create the stroke polygon
	var result geom.Path

	if !isClosed {
		// Add start cap
		startCap := calculateCap(segments[0], true, halfWidth, paint.LineCap)
		result = appendPath(result, startCap)
	}

	// Build the main stroke body
	if len(leftOffsets) > 0 {
		// Left side
		result.C = append(result.C, geom.MoveTo)
		result.V = append(result.V, leftOffsets[0])
		for i := 1; i < len(leftOffsets); i++ {
			result.C = append(result.C, geom.LineTo)
			result.V = append(result.V, leftOffsets[i])
		}

		// Right side (in reverse)
		for i := len(rightOffsets) - 1; i >= 0; i-- {
			result.C = append(result.C, geom.LineTo)
			result.V = append(result.V, rightOffsets[i])
		}

		result.C = append(result.C, geom.ClosePath)
	}

	if !isClosed {
		// Add end cap
		endCap := calculateCap(segments[len(segments)-1], false, halfWidth, paint.LineCap)
		result = appendPath(result, endCap)
	}

	return result
}

// segment represents a line segment.
type segment struct {
	Start, End geom.Pt
}

// pathToSegments converts a path to a series of line segments, flattening curves.
func pathToSegments(p geom.Path) []segment {
	var segments []segment
	var currentPt geom.Pt
	var startPt geom.Pt

	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			currentPt = p.V[vi]
			startPt = currentPt
			vi++
		case geom.LineTo:
			nextPt := p.V[vi]
			segments = append(segments, segment{Start: currentPt, End: nextPt})
			currentPt = nextPt
			vi++
		case geom.QuadTo:
			// Flatten quadratic curve to line segments
			ctrl := p.V[vi]
			end := p.V[vi+1]
			flatSegments := flattenQuad(currentPt, ctrl, end)
			segments = append(segments, flatSegments...)
			currentPt = end
			vi += 2
		case geom.CubicTo:
			// Flatten cubic curve to line segments
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			end := p.V[vi+2]
			flatSegments := flattenCubic(currentPt, c1, c2, end)
			segments = append(segments, flatSegments...)
			currentPt = end
			vi += 3
		case geom.ClosePath:
			if currentPt != startPt {
				segments = append(segments, segment{Start: currentPt, End: startPt})
			}
		}
	}

	return segments
}

// segmentNormal calculates the normal vector for a segment with the given width.
func segmentNormal(seg segment, halfWidth float64) geom.Pt {
	dx := quantize(seg.End.X - seg.Start.X)
	dy := quantize(seg.End.Y - seg.Start.Y)
	length := quantize(math.Sqrt(dx*dx + dy*dy))

	if length == 0 {
		return geom.Pt{X: halfWidth, Y: 0}
	}

	// Perpendicular vector (rotated 90 degrees)
	return quantizePt(geom.Pt{
		X: quantize(-dy / length * halfWidth),
		Y: quantize(dx / length * halfWidth),
	})
}

// calculateJoin computes the join point between two segments.
func calculateJoin(prev, curr segment, halfWidth float64, joinStyle render.LineJoin, miterLimit float64) (left, right geom.Pt) {
	// Get normals for both segments
	prevNormal := segmentNormal(prev, halfWidth)
	currNormal := segmentNormal(curr, halfWidth)

	// Join point (where segments meet)
	joinPt := prev.End // Should be same as curr.Start

	// Default to bevel join
	left = geom.Pt{X: joinPt.X + currNormal.X, Y: joinPt.Y + currNormal.Y}
	right = geom.Pt{X: joinPt.X - currNormal.X, Y: joinPt.Y - currNormal.Y}

	switch joinStyle {
	case render.JoinMiter:
		// Calculate direction vectors
		prevDir := geom.Pt{X: prev.End.X - prev.Start.X, Y: prev.End.Y - prev.Start.Y}
		currDir := geom.Pt{X: curr.End.X - curr.Start.X, Y: curr.End.Y - curr.Start.Y}
		
		// Normalize direction vectors
		prevLen := math.Sqrt(prevDir.X*prevDir.X + prevDir.Y*prevDir.Y)
		currLen := math.Sqrt(currDir.X*currDir.X + currDir.Y*currDir.Y)
		
		if prevLen > 0 && currLen > 0 {
			prevDir.X /= prevLen
			prevDir.Y /= prevLen  
			currDir.X /= currLen
			currDir.Y /= currLen
			
			// Calculate the angle between the segments
			dot := prevDir.X*currDir.X + prevDir.Y*currDir.Y
			
			// Avoid miter for nearly parallel or opposite directions
			if math.Abs(dot) > 0.999 { // Nearly parallel (< 2.5 degrees)
				// Fall back to bevel for nearly parallel lines
				left = geom.Pt{X: joinPt.X + currNormal.X, Y: joinPt.Y + currNormal.Y}
				right = geom.Pt{X: joinPt.X - currNormal.X, Y: joinPt.Y - currNormal.Y}
			} else {
				// Calculate miter length using sin of half-angle
				halfAngleSin := math.Sqrt((1 - dot) / 2)
				if halfAngleSin > 0 {
					miterLength := halfWidth / halfAngleSin
					
					// Check miter limit
					if miterLength <= miterLimit * halfWidth {
						// Calculate miter intersection points
						leftMiter := intersectLines(
							geom.Pt{X: prev.Start.X + prevNormal.X, Y: prev.Start.Y + prevNormal.Y},
							geom.Pt{X: prev.End.X + prevNormal.X, Y: prev.End.Y + prevNormal.Y},
							geom.Pt{X: curr.Start.X + currNormal.X, Y: curr.Start.Y + currNormal.Y},
							geom.Pt{X: curr.End.X + currNormal.X, Y: curr.End.Y + currNormal.Y},
						)
						rightMiter := intersectLines(
							geom.Pt{X: prev.Start.X - prevNormal.X, Y: prev.Start.Y - prevNormal.Y},
							geom.Pt{X: prev.End.X - prevNormal.X, Y: prev.End.Y - prevNormal.Y},
							geom.Pt{X: curr.Start.X - currNormal.X, Y: curr.Start.Y - currNormal.Y},
							geom.Pt{X: curr.End.X - currNormal.X, Y: curr.End.Y - currNormal.Y},
						)
						
						left = quantizePt(leftMiter)
						right = quantizePt(rightMiter)
					} else {
						// Miter limit exceeded, fall back to bevel
						left = geom.Pt{X: joinPt.X + currNormal.X, Y: joinPt.Y + currNormal.Y}
						right = geom.Pt{X: joinPt.X - currNormal.X, Y: joinPt.Y - currNormal.Y}
					}
				}
			}
		}

	case render.JoinRound:
		// Calculate the angle between segments to determine arc
		prevDir := geom.Pt{X: prev.End.X - prev.Start.X, Y: prev.End.Y - prev.Start.Y}
		currDir := geom.Pt{X: curr.End.X - curr.Start.X, Y: curr.End.Y - curr.Start.Y}
		
		// Normalize directions
		prevLen := math.Sqrt(prevDir.X*prevDir.X + prevDir.Y*prevDir.Y)
		currLen := math.Sqrt(currDir.X*currDir.X + currDir.Y*currDir.Y)
		
		if prevLen > 0 && currLen > 0 {
			prevDir.X /= prevLen
			prevDir.Y /= prevLen
			currDir.X /= currLen
			currDir.Y /= currLen
			
			// Calculate angle between directions using dot product and cross product
			dot := prevDir.X*currDir.X + prevDir.Y*currDir.Y
			cross := prevDir.X*currDir.Y - prevDir.Y*currDir.X
			angle := math.Atan2(cross, dot)
			
			if math.Abs(angle) > 0.01 { // Only create round join if there's significant angle
				// Use miter intersection as approximation for round join center
				leftMiter := intersectLines(
					geom.Pt{X: prev.Start.X + prevNormal.X, Y: prev.Start.Y + prevNormal.Y},
					geom.Pt{X: prev.End.X + prevNormal.X, Y: prev.End.Y + prevNormal.Y},
					geom.Pt{X: curr.Start.X + currNormal.X, Y: curr.Start.Y + currNormal.Y},
					geom.Pt{X: curr.End.X + currNormal.X, Y: curr.End.Y + currNormal.Y},
				)
				rightMiter := intersectLines(
					geom.Pt{X: prev.Start.X - prevNormal.X, Y: prev.Start.Y - prevNormal.Y},
					geom.Pt{X: prev.End.X - prevNormal.X, Y: prev.End.Y - prevNormal.Y},
					geom.Pt{X: curr.Start.X - currNormal.X, Y: curr.Start.Y - currNormal.Y},
					geom.Pt{X: curr.End.X - currNormal.X, Y: curr.End.Y - currNormal.Y},
				)
				
				// Use the miter points if they're reasonable, otherwise default to bevel
				if distance(joinPt, leftMiter) < miterLimit*halfWidth {
					left = leftMiter
				}
				if distance(joinPt, rightMiter) < miterLimit*halfWidth {
					right = rightMiter
				}
			}
		}

	case render.JoinBevel:
		// Already set to bevel above
	}

	return left, right
}

// calculateCap generates the cap geometry for the start or end of a path.
func calculateCap(seg segment, isStart bool, halfWidth float64, capStyle render.LineCap) geom.Path {
	var result geom.Path

	normal := segmentNormal(seg, halfWidth)
	var capPt geom.Pt

	if isStart {
		capPt = seg.Start
	} else {
		capPt = seg.End
		// Flip normal direction for end cap
		normal.X = -normal.X
		normal.Y = -normal.Y
	}

	switch capStyle {
	case render.CapButt:
		// No additional geometry needed for butt cap
		return geom.Path{}

	case render.CapSquare:
		// Extend by half line width
		dx := seg.End.X - seg.Start.X
		dy := seg.End.Y - seg.Start.Y
		length := math.Sqrt(dx*dx + dy*dy)

		var extendVec geom.Pt
		if length > 0 {
			extendVec = geom.Pt{X: dx / length * halfWidth, Y: dy / length * halfWidth}
		}

		if isStart {
			extendVec.X = -extendVec.X
			extendVec.Y = -extendVec.Y
		}

		// Create square cap rectangle
		p1 := geom.Pt{X: capPt.X + normal.X + extendVec.X, Y: capPt.Y + normal.Y + extendVec.Y}
		p2 := geom.Pt{X: capPt.X + normal.X, Y: capPt.Y + normal.Y}
		p3 := geom.Pt{X: capPt.X - normal.X, Y: capPt.Y - normal.Y}
		p4 := geom.Pt{X: capPt.X - normal.X + extendVec.X, Y: capPt.Y - normal.Y + extendVec.Y}

		result.C = []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.LineTo, geom.ClosePath}
		result.V = []geom.Pt{p1, p2, p3, p4}

	case render.CapRound:
		// Create semicircle with adaptive subdivision based on radius
		radius := halfWidth
		numSegments := int(math.Max(8, math.Min(32, radius*2))) // Adaptive based on size
		
		result.C = append(result.C, geom.MoveTo)
		result.V = append(result.V, geom.Pt{X: capPt.X + normal.X, Y: capPt.Y + normal.Y})

		for i := 1; i <= numSegments; i++ {
			angle := math.Pi * float64(i) / float64(numSegments)
			if !isStart {
				angle = -angle
			}

			cos := math.Cos(angle)
			sin := math.Sin(angle)

			// Rotate normal vector to create points on the semicircle
			x := normal.X*cos - normal.Y*sin
			y := normal.X*sin + normal.Y*cos

			result.C = append(result.C, geom.LineTo)
			result.V = append(result.V, quantizePt(geom.Pt{X: capPt.X + x, Y: capPt.Y + y}))
		}

		result.C = append(result.C, geom.ClosePath)
	}

	return result
}

// applyDashes decomposes a path into dashed segments.
func applyDashes(p geom.Path, dashes []float64) geom.Path {
	if len(dashes) == 0 || len(dashes)%2 != 0 {
		return p // Invalid dash pattern
	}

	var result geom.Path
	subpaths := splitIntoSubpaths(p)

	for _, subpath := range subpaths {
		dashedSubpath := applyDashesToSubpath(subpath, dashes)
		result = appendPath(result, dashedSubpath)
	}

	return result
}

// applyDashesToSubpath applies dash pattern to a single subpath with improved precision.
func applyDashesToSubpath(p geom.Path, dashes []float64) geom.Path {
	segments := pathToSegments(p)
	if len(segments) == 0 {
		return geom.Path{}
	}

	// Quantize dash pattern for consistency
	quantizedDashes := make([]float64, len(dashes))
	for i, dash := range dashes {
		quantizedDashes[i] = quantize(dash)
	}

	var result geom.Path
	dashIndex := 0
	dashRemaining := quantizedDashes[0]
	isDrawing := true // First dash is always "on"
	const epsilon = 1e-10

	for _, seg := range segments {
		segLength := quantize(distance(seg.Start, seg.End))
		segConsumed := 0.0

		for segConsumed < segLength-epsilon {
			available := segLength - segConsumed
			consume := math.Min(available, dashRemaining)
			
			// Quantize consume to avoid precision issues
			consume = quantize(consume)

			if isDrawing && consume > epsilon {
				// Add this segment to the result
				t1 := segConsumed / segLength
				t2 := (segConsumed + consume) / segLength
				
				// Clamp t values to [0,1]
				t1 = math.Max(0, math.Min(1, t1))
				t2 = math.Max(0, math.Min(1, t2))

				start := quantizePt(interpolate(seg.Start, seg.End, t1))
				end := quantizePt(interpolate(seg.Start, seg.End, t2))

				// Only add if start and end are different
				if distance(start, end) > epsilon {
					result.C = append(result.C, geom.MoveTo, geom.LineTo)
					result.V = append(result.V, start, end)
				}
			}

			segConsumed += consume
			dashRemaining -= consume

			if dashRemaining <= epsilon {
				// Move to next dash
				dashIndex = (dashIndex + 1) % len(quantizedDashes)
				dashRemaining = quantizedDashes[dashIndex]
				isDrawing = !isDrawing
			}
		}
	}

	return result
}

// Helper functions

func appendPath(dest, src geom.Path) geom.Path {
	dest.C = append(dest.C, src.C...)
	dest.V = append(dest.V, src.V...)
	return dest
}

func intersectLines(p1, p2, p3, p4 geom.Pt) geom.Pt {
	// Line intersection using parametric form
	x1, y1 := p1.X, p1.Y
	x2, y2 := p2.X, p2.Y
	x3, y3 := p3.X, p3.Y
	x4, y4 := p4.X, p4.Y

	denom := (x1-x2)*(y3-y4) - (y1-y2)*(x3-x4)
	if math.Abs(denom) < 1e-10 {
		// Lines are parallel, return midpoint
		return geom.Pt{X: (p2.X + p3.X) / 2, Y: (p2.Y + p3.Y) / 2}
	}

	t := ((x1-x3)*(y3-y4) - (y1-y3)*(x3-x4)) / denom

	return geom.Pt{
		X: x1 + t*(x2-x1),
		Y: y1 + t*(y2-y1),
	}
}

func distance(a, b geom.Pt) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func interpolate(a, b geom.Pt, t float64) geom.Pt {
	return geom.Pt{
		X: a.X + t*(b.X-a.X),
		Y: a.Y + t*(b.Y-a.Y),
	}
}

// flattenQuad approximates a quadratic bezier curve with line segments using adaptive subdivision.
func flattenQuad(start, ctrl, end geom.Pt) []segment {
	const tolerance = 0.5 // Maximum allowed deviation from true curve
	var segments []segment
	flattenQuadRecursive(start, ctrl, end, 0, 1, tolerance, &segments)
	return segments
}

// flattenQuadRecursive adaptively subdivides a quadratic curve segment.
func flattenQuadRecursive(start, ctrl, end geom.Pt, t0, t1, tolerance float64, segments *[]segment) {
	// Calculate midpoint on curve and midpoint of chord
	tmid := (t0 + t1) * 0.5
	curveMid := evaluateQuad(start, ctrl, end, tmid)
	chordMid := geom.Pt{
		X: (evaluateQuad(start, ctrl, end, t0).X + evaluateQuad(start, ctrl, end, t1).X) * 0.5,
		Y: (evaluateQuad(start, ctrl, end, t0).Y + evaluateQuad(start, ctrl, end, t1).Y) * 0.5,
	}
	
	// Calculate error (distance from curve to chord)
	curveError := distance(curveMid, chordMid)
	
	// If error is small enough or segment is very short, create a line segment
	if curveError <= tolerance || (t1-t0) < 0.01 {
		p1 := evaluateQuad(start, ctrl, end, t0)
		p2 := evaluateQuad(start, ctrl, end, t1)
		*segments = append(*segments, segment{Start: p1, End: p2})
	} else {
		// Subdivide further
		flattenQuadRecursive(start, ctrl, end, t0, tmid, tolerance, segments)
		flattenQuadRecursive(start, ctrl, end, tmid, t1, tolerance, segments)
	}
}

// flattenCubic approximates a cubic bezier curve with line segments using adaptive subdivision.
func flattenCubic(start, c1, c2, end geom.Pt) []segment {
	const tolerance = 0.5 // Maximum allowed deviation from true curve
	var segments []segment
	flattenCubicRecursive(start, c1, c2, end, 0, 1, tolerance, &segments)
	return segments
}

// flattenCubicRecursive adaptively subdivides a cubic curve segment.
func flattenCubicRecursive(start, c1, c2, end geom.Pt, t0, t1, tolerance float64, segments *[]segment) {
	// Calculate midpoint on curve and midpoint of chord
	tmid := (t0 + t1) * 0.5
	curveMid := evaluateCubic(start, c1, c2, end, tmid)
	chordMid := geom.Pt{
		X: (evaluateCubic(start, c1, c2, end, t0).X + evaluateCubic(start, c1, c2, end, t1).X) * 0.5,
		Y: (evaluateCubic(start, c1, c2, end, t0).Y + evaluateCubic(start, c1, c2, end, t1).Y) * 0.5,
	}
	
	// Calculate error (distance from curve to chord)
	curveError := distance(curveMid, chordMid)
	
	// If error is small enough or segment is very short, create a line segment
	if curveError <= tolerance || (t1-t0) < 0.01 {
		p1 := evaluateCubic(start, c1, c2, end, t0)
		p2 := evaluateCubic(start, c1, c2, end, t1)
		*segments = append(*segments, segment{Start: p1, End: p2})
	} else {
		// Subdivide further
		flattenCubicRecursive(start, c1, c2, end, t0, tmid, tolerance, segments)
		flattenCubicRecursive(start, c1, c2, end, tmid, t1, tolerance, segments)
	}
}

func evaluateQuad(start, ctrl, end geom.Pt, t float64) geom.Pt {
	// B(t) = (1-t)²P₀ + 2(1-t)tP₁ + t²P₂
	t1 := 1 - t
	return geom.Pt{
		X: t1*t1*start.X + 2*t1*t*ctrl.X + t*t*end.X,
		Y: t1*t1*start.Y + 2*t1*t*ctrl.Y + t*t*end.Y,
	}
}

func evaluateCubic(start, c1, c2, end geom.Pt, t float64) geom.Pt {
	// B(t) = (1-t)³P₀ + 3(1-t)²tP₁ + 3(1-t)t²P₂ + t³P₃
	t1 := 1 - t
	return geom.Pt{
		X: t1*t1*t1*start.X + 3*t1*t1*t*c1.X + 3*t1*t*t*c2.X + t*t*t*end.X,
		Y: t1*t1*t1*start.Y + 3*t1*t1*t*c1.Y + 3*t1*t*t*c2.Y + t*t*t*end.Y,
	}
}
