package agg

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
)

const defaultPathChunkVertices = 32768

// pathChunk is one stroke-sized piece of a very large path, together with the
// dash phase that piece starts at.
//
// Chunking splits a polyline across several separate Stroke calls, and AGG's
// dash generator restarts at every subpath of every stroke — so without a phase
// a dashed line longer than the chunk size would snap back to the start of the
// pattern at each boundary. DashPhase is the arc length from the last genuine
// MoveTo to the point the chunk resumes at, which is exactly what has to be
// added to the paint's dash offset for the pattern to continue.
type pathChunk struct {
	Path geom.Path
	// DashPhase is arc length in the path's own units. It is 0 whenever the
	// chunk opens on a MoveTo that was present in the source path — a genuine
	// subpath start, where the pattern is meant to restart.
	DashPhase float64
}

func chunkStrokePath(path geom.Path, maxVertices int) []pathChunk {
	if maxVertices <= 0 {
		maxVertices = defaultPathChunkVertices
	}
	if len(path.V) <= maxVertices || pathHasCurvesOrClose(path) {
		return []pathChunk{{Path: path}}
	}

	chunks := make([]pathChunk, 0, len(path.V)/maxVertices+1)
	vi := 0
	var current geom.Path
	currentVertices := 0
	haveCurrent := false
	var last geom.Pt

	// phase is the arc length walked since the last genuine MoveTo; chunkPhase
	// freezes that value at the moment the chunk under construction opened.
	// Only the chunk's first subpath needs it — any later subpath in the same
	// chunk begins at a genuine MoveTo and restarts the pattern on its own.
	phase := 0.0
	chunkPhase := 0.0

	flush := func() {
		if len(current.C) > 1 {
			chunks = append(chunks, pathChunk{Path: current, DashPhase: chunkPhase})
		}
		current = geom.Path{}
		currentVertices = 0
		haveCurrent = false
		chunkPhase = phase
	}

	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				flush()
				return chunks
			}
			phase = 0
			if currentVertices >= maxVertices {
				flush()
			}
			last = path.V[vi]
			vi++
			current.MoveTo(last)
			currentVertices++
			haveCurrent = true
		case geom.LineTo:
			if vi >= len(path.V) {
				flush()
				return chunks
			}
			to := path.V[vi]
			vi++
			// flush() clears haveCurrent, so the segment test has to be read
			// before the chunk boundary is taken.
			drawsSegment := haveCurrent
			if !haveCurrent {
				current.MoveTo(to)
				currentVertices++
			} else if currentVertices >= maxVertices {
				flush()
				current.MoveTo(last)
				currentVertices++
			}
			current.LineTo(to)
			currentVertices++
			if drawsSegment {
				phase += math.Hypot(to.X-last.X, to.Y-last.Y)
			}
			last = to
			haveCurrent = true
		}
	}
	flush()
	if len(chunks) == 0 {
		return []pathChunk{{Path: path}}
	}
	return chunks
}
