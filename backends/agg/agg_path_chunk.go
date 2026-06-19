package agg

import "github.com/cwbudde/matplotlib-go/geom"

const defaultPathChunkVertices = 32768

func chunkStrokePath(path geom.Path, maxVertices int) []geom.Path {
	if maxVertices <= 0 {
		maxVertices = defaultPathChunkVertices
	}
	if len(path.V) <= maxVertices || pathHasCurvesOrClose(path) {
		return []geom.Path{path}
	}

	chunks := make([]geom.Path, 0, len(path.V)/maxVertices+1)
	vi := 0
	var current geom.Path
	currentVertices := 0
	haveCurrent := false
	var last geom.Pt

	flush := func() {
		if len(current.C) > 1 {
			chunks = append(chunks, current)
		}
		current = geom.Path{}
		currentVertices = 0
		haveCurrent = false
	}

	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				flush()
				return chunks
			}
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
			last = to
			haveCurrent = true
		}
	}
	flush()
	if len(chunks) == 0 {
		return []geom.Path{path}
	}
	return chunks
}
