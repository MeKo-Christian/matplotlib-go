package gobasic

import (
	"math"
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestStrokeToPath_SimpleLineJoins(t *testing.T) {
	// Test different line join styles
	testCases := []struct {
		name      string
		joinStyle render.LineJoin
	}{
		{"miter", render.JoinMiter},
		{"round", render.JoinRound},
		{"bevel", render.JoinBevel},
	}

	// Simple L-shaped path
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paint := render.Paint{
				LineWidth:  2.0,
				LineJoin:   tc.joinStyle,
				LineCap:    render.CapButt,
				MiterLimit: 10.0,
				Stroke:     render.Color{R: 1, G: 0, B: 0, A: 1},
			}

			strokePath := strokeToPath(path, &paint)
			if len(strokePath.C) == 0 {
				t.Error("Expected stroke path to have commands")
			}
			if len(strokePath.V) == 0 {
				t.Error("Expected stroke path to have vertices")
			}
		})
	}
}

func TestStrokeToPath_LineCaps(t *testing.T) {
	// Test different line cap styles
	testCases := []struct {
		name     string
		capStyle render.LineCap
	}{
		{"butt", render.CapButt},
		{"round", render.CapRound},
		{"square", render.CapSquare},
	}

	// Simple line
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 10, Y: 0}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paint := render.Paint{
				LineWidth:  2.0,
				LineJoin:   render.JoinMiter,
				LineCap:    tc.capStyle,
				MiterLimit: 10.0,
				Stroke:     render.Color{R: 1, G: 0, B: 0, A: 1},
			}

			strokePath := strokeToPath(path, &paint)
			if len(strokePath.C) == 0 {
				t.Error("Expected stroke path to have commands")
			}
			if len(strokePath.V) == 0 {
				t.Error("Expected stroke path to have vertices")
			}
		})
	}
}

func TestStrokeToPath_Dashes(t *testing.T) {
	// Simple line
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 20, Y: 0}},
	}

	paint := render.Paint{
		LineWidth:  2.0,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
		MiterLimit: 10.0,
		Stroke:     render.Color{R: 1, G: 0, B: 0, A: 1},
		Dashes:     []float64{5, 2}, // 5 units on, 2 units off
	}

	strokePath := strokeToPath(path, &paint)

	// Should have multiple subpaths for dashed line
	if len(strokePath.C) == 0 {
		t.Error("Expected stroke path to have commands")
	}
	if len(strokePath.V) == 0 {
		t.Error("Expected stroke path to have vertices")
	}
}

func TestSegmentNormal(t *testing.T) {
	testCases := []struct {
		name     string
		seg      segment
		expected geom.Pt
	}{
		{
			name:     "horizontal_line",
			seg:      segment{Start: geom.Pt{X: 0, Y: 0}, End: geom.Pt{X: 10, Y: 0}},
			expected: geom.Pt{X: 0, Y: 1}, // perpendicular pointing up
		},
		{
			name:     "vertical_line",
			seg:      segment{Start: geom.Pt{X: 0, Y: 0}, End: geom.Pt{X: 0, Y: 10}},
			expected: geom.Pt{X: -1, Y: 0}, // perpendicular pointing left
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normal := segmentNormal(tc.seg, 1.0)

			// Check magnitude is approximately 1
			mag := math.Sqrt(normal.X*normal.X + normal.Y*normal.Y)
			if math.Abs(mag-1.0) > 1e-10 {
				t.Errorf("Expected normal magnitude ~1.0, got %v", mag)
			}

			// Check direction (within tolerance)
			if math.Abs(normal.X-tc.expected.X) > 1e-10 || math.Abs(normal.Y-tc.expected.Y) > 1e-10 {
				t.Errorf("Expected normal %v, got %v", tc.expected, normal)
			}
		})
	}
}

func TestPathToSegments(t *testing.T) {
	// Test converting a simple path to segments
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.ClosePath},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}},
	}

	segments := pathToSegments(path)

	if len(segments) != 3 {
		t.Errorf("Expected 3 segments, got %d", len(segments))
	}

	// Check first segment
	if segments[0].Start != (geom.Pt{X: 0, Y: 0}) || segments[0].End != (geom.Pt{X: 10, Y: 0}) {
		t.Errorf("First segment incorrect: %v", segments[0])
	}

	// Check last segment (close path)
	if segments[2].Start != (geom.Pt{X: 10, Y: 10}) || segments[2].End != (geom.Pt{X: 0, Y: 0}) {
		t.Errorf("Last segment incorrect: %v", segments[2])
	}
}

func TestSplitIntoSubpaths(t *testing.T) {
	// Path with two subpaths
	path := geom.Path{
		C: []geom.Cmd{
			geom.MoveTo, geom.LineTo, // First subpath
			geom.MoveTo, geom.LineTo, geom.LineTo, // Second subpath
		},
		V: []geom.Pt{
			{X: 0, Y: 0}, {X: 10, Y: 0}, // First subpath
			{X: 20, Y: 0}, {X: 30, Y: 0}, {X: 30, Y: 10}, // Second subpath
		},
	}

	subpaths := splitIntoSubpaths(path)

	if len(subpaths) != 2 {
		t.Errorf("Expected 2 subpaths, got %d", len(subpaths))
	}

	// Check first subpath
	if len(subpaths[0].C) != 2 || subpaths[0].C[0] != geom.MoveTo || subpaths[0].C[1] != geom.LineTo {
		t.Errorf("First subpath commands incorrect: %v", subpaths[0].C)
	}

	// Check second subpath
	if len(subpaths[1].C) != 3 {
		t.Errorf("Second subpath should have 3 commands, got %d", len(subpaths[1].C))
	}
}

func TestDistance(t *testing.T) {
	a := geom.Pt{X: 0, Y: 0}
	b := geom.Pt{X: 3, Y: 4}

	dist := distance(a, b)
	expected := 5.0 // 3-4-5 triangle

	if math.Abs(dist-expected) > 1e-10 {
		t.Errorf("Expected distance %v, got %v", expected, dist)
	}
}

func TestInterpolate(t *testing.T) {
	a := geom.Pt{X: 0, Y: 0}
	b := geom.Pt{X: 10, Y: 10}

	mid := interpolate(a, b, 0.5)
	expected := geom.Pt{X: 5, Y: 5}

	if mid != expected {
		t.Errorf("Expected midpoint %v, got %v", expected, mid)
	}
}

func TestApplyDashes_SimpleLine(t *testing.T) {
	// Line longer than dash pattern
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 20, Y: 0}},
	}

	dashes := []float64{5, 2} // 5 on, 2 off

	dashedPath := applyDashes(path, dashes)

	// Should have multiple MoveTo commands for separate dash segments
	if len(dashedPath.C) == 0 {
		t.Error("Expected dashed path to have commands")
	}

	// Count MoveTo commands (each dash segment starts with MoveTo)
	moveCount := 0
	for _, cmd := range dashedPath.C {
		if cmd == geom.MoveTo {
			moveCount++
		}
	}

	if moveCount < 2 {
		t.Errorf("Expected at least 2 dash segments, got %d MoveTo commands", moveCount)
	}
}

func TestApplyDashes_InvalidPattern(t *testing.T) {
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 10, Y: 0}},
	}

	// Invalid dash pattern (odd number of elements)
	dashes := []float64{5, 2, 3}

	dashedPath := applyDashes(path, dashes)

	// Should return original path unchanged
	if len(dashedPath.C) != len(path.C) {
		t.Errorf("Expected original path to be returned for invalid dash pattern")
	}
}

func TestAdaptiveCurveFlattening(t *testing.T) {
	// Test that adaptive flattening produces more segments for high-curvature curves
	start := geom.Pt{X: 0, Y: 0}
	ctrl := geom.Pt{X: 5, Y: 10} // High curvature
	end := geom.Pt{X: 10, Y: 0}

	segments := flattenQuad(start, ctrl, end)

	// Should produce more than the old fixed 4 segments for high curvature
	if len(segments) < 4 {
		t.Errorf("Expected at least 4 segments for high-curvature quad, got %d", len(segments))
	}

	// Test low curvature case
	ctrlLow := geom.Pt{X: 5, Y: 0.1} // Very low curvature
	segmentsLow := flattenQuad(start, ctrlLow, end)

	// Should produce fewer segments for low curvature
	if len(segmentsLow) > len(segments) {
		t.Errorf("Expected fewer segments for low-curvature quad: %d vs %d", len(segmentsLow), len(segments))
	}
}

func TestMiterJoinEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		prev     segment
		curr     segment
		expected string // "miter", "bevel", or description
	}{
		{
			name:     "nearly_parallel",
			prev:     segment{Start: geom.Pt{X: 0, Y: 0}, End: geom.Pt{X: 10, Y: 0}},
			curr:     segment{Start: geom.Pt{X: 10, Y: 0}, End: geom.Pt{X: 20, Y: 0.1}}, // 0.57 degrees
			expected: "bevel", // Should fall back to bevel for nearly parallel lines
		},
		{
			name:     "acute_angle",
			prev:     segment{Start: geom.Pt{X: 0, Y: 0}, End: geom.Pt{X: 10, Y: 0}},
			curr:     segment{Start: geom.Pt{X: 10, Y: 0}, End: geom.Pt{X: 15, Y: 10}}, // Sharp angle
			expected: "depends on miter limit",
		},
		{
			name:     "right_angle",
			prev:     segment{Start: geom.Pt{X: 0, Y: 0}, End: geom.Pt{X: 10, Y: 0}},
			curr:     segment{Start: geom.Pt{X: 10, Y: 0}, End: geom.Pt{X: 10, Y: 10}},
			expected: "miter", // Should always work for right angles
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			halfWidth := 1.0
			miterLimit := 10.0

			left, right := calculateJoin(tc.prev, tc.curr, halfWidth, render.JoinMiter, miterLimit)

			// Verify that join points are reasonable (not NaN, not extremely far)
			if math.IsNaN(left.X) || math.IsNaN(left.Y) || math.IsNaN(right.X) || math.IsNaN(right.Y) {
				t.Error("Join calculation produced NaN values")
			}

			joinPt := tc.prev.End
			leftDist := distance(joinPt, left)
			rightDist := distance(joinPt, right)

			// Miter points shouldn't be extremely far from the join point
			maxReasonableDistance := miterLimit * halfWidth * 2
			if leftDist > maxReasonableDistance || rightDist > maxReasonableDistance {
				t.Errorf("Join points too far from join: left=%.2f, right=%.2f, max=%.2f",
					leftDist, rightDist, maxReasonableDistance)
			}
		})
	}
}

func TestRoundCapAdaptiveSegments(t *testing.T) {
	// Test that round caps adapt segment count based on radius
	testCases := []struct {
		name      string
		halfWidth float64
		minSegs   int
		maxSegs   int
	}{
		{"small_radius", 0.5, 8, 10},
		{"medium_radius", 5.0, 10, 20},
		{"large_radius", 20.0, 30, 32},
	}

	seg := segment{Start: geom.Pt{X: 0, Y: 0}, End: geom.Pt{X: 10, Y: 0}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			capPath := calculateCap(seg, true, tc.halfWidth, render.CapRound)

			// Count the number of segments in the cap
			segCount := 0
			for _, cmd := range capPath.C {
				if cmd == geom.LineTo {
					segCount++
				}
			}

			if segCount < tc.minSegs || segCount > tc.maxSegs {
				t.Errorf("Expected %d-%d segments for radius %.1f, got %d",
					tc.minSegs, tc.maxSegs, tc.halfWidth, segCount)
			}
		})
	}
}

func TestDashPatternPrecision(t *testing.T) {
	// Test that dash patterns maintain precision and don't accumulate errors
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 100, Y: 0}},
	}

	dashes := []float64{1.0, 1.0} // 1 on, 1 off
	dashedPath := applyDashesToSubpath(path, dashes)

	// Count the number of dash segments
	dashCount := 0
	for _, cmd := range dashedPath.C {
		if cmd == geom.MoveTo {
			dashCount++
		}
	}

	// For a 100-unit line with 1-unit dashes, expect ~50 dash segments
	expectedMin, expectedMax := 45, 55
	if dashCount < expectedMin || dashCount > expectedMax {
		t.Errorf("Expected %d-%d dash segments for 100-unit line, got %d",
			expectedMin, expectedMax, dashCount)
	}
}

func TestZeroLengthSegmentHandling(t *testing.T) {
	// Test that zero-length segments don't cause crashes or infinite loops
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 0, Y: 0}, {X: 10, Y: 0}}, // Zero-length first segment
	}

	paint := render.Paint{
		LineWidth:  2.0,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
		MiterLimit: 10.0,
		Stroke:     render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	// Should not crash or hang
	strokePath := strokeToPath(path, &paint)

	// Should produce some valid geometry
	if len(strokePath.C) == 0 {
		t.Error("Expected stroke path to have commands even with zero-length segments")
	}
}
