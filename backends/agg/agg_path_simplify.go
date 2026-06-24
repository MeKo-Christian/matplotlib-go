package agg

import "github.com/cwbudde/matplotlib-go/geom"

// simplifyLinePath reduces vertices on line-only paths using Matplotlib's
// single-pass running-segment algorithm (PathSimplifier in path_converters.h).
// Paths containing curves or ClosePath are returned unchanged.
//
// The algorithm tracks the furthest point in the forward (parallel) and
// backward (anti-parallel) directions from the current reference segment.
// When perpendicular deviation exceeds the threshold the accumulated extrema
// are emitted—preserving local maxima in oscillating data—and a new segment
// begins. This is O(n) and matches Matplotlib's pixel-exact behaviour on
// dense lines, including zigzag/wave paths where Douglas–Peucker fails.
func simplifyLinePath(path geom.Path, threshold float64) geom.Path {
	if threshold <= 0 || pathHasCurvesOrClose(path) {
		return path
	}

	threshold2 := threshold * threshold
	out := geom.Path{}
	vi := 0

	// Running-segment state; all fields are reset on each MoveTo.
	var (
		lastx, lasty                 float64
		currVecStartX, currVecStartY float64
		origdx, origdy               float64
		origdNorm2                   float64
		nextX, nextY                 float64
		dnorm2ForwardMax             float64
		lastForwardMax               bool
		nextBackwardX, nextBackwardY float64
		dnorm2BackwardMax            float64
		lastBackwardMax              bool
		lastEmitX, lastEmitY         float64
		hasOrig                      bool
		inSubpath                    bool
	)

	emitLine := func(x, y float64) {
		out.LineTo(geom.Pt{X: x, Y: y})
		lastEmitX, lastEmitY = x, y
	}

	// flushSubpath emits the accumulated forward/backward maxima and—only when
	// different from the last emitted vertex—the actual last input point.
	// Matches Matplotlib's PathSimplifier end-of-path handling.
	flushSubpath := func() {
		if !hasOrig {
			return
		}
		emitLine(nextX, nextY)
		if dnorm2BackwardMax > 0 {
			emitLine(nextBackwardX, nextBackwardY)
		}
		if lastx != lastEmitX || lasty != lastEmitY {
			emitLine(lastx, lasty)
		}
		hasOrig = false
	}

	// pushSegment is called when perpendicular distance to the reference line
	// exceeds the threshold. It emits accumulated extrema in the order that
	// avoids artifacts (mirrors PathSimplifier::_push), then resets state for
	// the new segment running from lastx,lasty to x,y.
	pushSegment := func(x, y float64) {
		if dnorm2BackwardMax > 0 {
			if lastForwardMax {
				emitLine(nextBackwardX, nextBackwardY)
				emitLine(nextX, nextY)
			} else {
				emitLine(nextX, nextY)
				emitLine(nextBackwardX, nextBackwardY)
			}
		} else {
			emitLine(nextX, nextY)
		}
		// If the last processed point was not one of the maxima, emit it so the
		// rendered line returns to the true endpoint of the preceding group.
		if !lastForwardMax && !lastBackwardMax {
			emitLine(lastx, lasty)
		}
		// New direction: from last input point to the threshold-breaking point.
		origdx = x - lastx
		origdy = y - lasty
		origdNorm2 = origdx*origdx + origdy*origdy
		dnorm2ForwardMax = origdNorm2
		lastForwardMax = true
		lastBackwardMax = false
		// currVecStart is the last canvas point (last emitted), not lastx.
		currVecStartX = lastEmitX
		currVecStartY = lastEmitY
		lastx = x
		lasty = y
		nextX = x
		nextY = y
		dnorm2BackwardMax = 0
	}

	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if inSubpath {
				flushSubpath()
			}
			pt := path.V[vi]
			vi++
			out.MoveTo(pt)
			lastEmitX, lastEmitY = pt.X, pt.Y
			lastx, lasty = pt.X, pt.Y
			origdNorm2 = 0
			dnorm2BackwardMax = 0
			lastForwardMax = false
			lastBackwardMax = false
			hasOrig = false
			inSubpath = true

		case geom.LineTo:
			x := path.V[vi].X
			y := path.V[vi].Y
			vi++

			if !inSubpath {
				// LineTo with no preceding MoveTo: treat first vertex as origin.
				out.MoveTo(geom.Pt{X: x, Y: y})
				lastEmitX, lastEmitY = x, y
				lastx, lasty = x, y
				origdNorm2 = 0
				dnorm2BackwardMax = 0
				hasOrig = false
				inSubpath = true
				continue
			}

			if origdNorm2 == 0 {
				// First LineTo after MoveTo (or after a zero-length push):
				// establish the reference direction for this segment.
				origdx = x - lastx
				origdy = y - lasty
				origdNorm2 = origdx*origdx + origdy*origdy
				if origdNorm2 == 0 {
					// Duplicate vertex — advance without changing state.
					lastx = x
					lasty = y
					continue
				}
				dnorm2ForwardMax = origdNorm2
				dnorm2BackwardMax = 0
				lastForwardMax = true
				lastBackwardMax = false
				currVecStartX = lastx
				currVecStartY = lasty
				lastx = x
				lasty = y
				nextX = x
				nextY = y
				hasOrig = true
				continue
			}

			// Perpendicular distance from the reference line.
			// Let o = (origdx, origdy) and v = vector from currVecStart to (x,y).
			// perp = v − (o·v / o·o)·o
			totdx := x - currVecStartX
			totdy := y - currVecStartY
			totdot := origdx*totdx + origdy*totdy
			paradx := totdot * origdx / origdNorm2
			parady := totdot * origdy / origdNorm2
			perpdx := totdx - paradx
			perpdy := totdy - parady
			perpdNorm2 := perpdx*perpdx + perpdy*perpdy

			if perpdNorm2 < threshold2 {
				// Within threshold: track whichever direction extends furthest.
				paradNorm2 := paradx*paradx + parady*parady
				lastForwardMax = false
				lastBackwardMax = false
				if totdot > 0 {
					if paradNorm2 > dnorm2ForwardMax {
						lastForwardMax = true
						dnorm2ForwardMax = paradNorm2
						nextX = x
						nextY = y
					}
				} else {
					if paradNorm2 > dnorm2BackwardMax {
						lastBackwardMax = true
						dnorm2BackwardMax = paradNorm2
						nextBackwardX = x
						nextBackwardY = y
					}
				}
				lastx = x
				lasty = y
				continue
			}

			// Perpendicular deviation too large: flush the accumulated segment
			// and start a new one.
			pushSegment(x, y)
		}
	}

	flushSubpath()
	return out
}

func pathHasCurvesOrClose(path geom.Path) bool {
	for _, cmd := range path.C {
		if cmd == geom.QuadTo || cmd == geom.CubicTo || cmd == geom.ClosePath {
			return true
		}
	}
	return false
}
