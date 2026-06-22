package sketch

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
)

// segCmd mirrors the AGG path commands the segmentator emits.
type segCmd uint8

const (
	segStop segCmd = iota
	segMove
	segLine
)

// segmentator is a faithful port of AGG's vpgen_segmentator with an
// approximation scale of 1.0: it resamples each line segment into pieces of
// roughly one pixel, which gives the sketch sine wave enough sample points.
//
// Usage mirrors AGG's conv_adaptor_vpgen: call moveTo/lineTo to feed a segment,
// then drain vertex() until it returns segStop.
type segmentator struct {
	x1, y1 float64
	dx, dy float64
	dl     float64 // current parameter along the segment, 0..1
	ddl    float64 // step per emitted vertex
	cmd    segCmd
}

func newSegmentator() *segmentator {
	return &segmentator{cmd: segStop}
}

func (s *segmentator) moveTo(p geom.Pt) {
	s.x1 = p.X
	s.y1 = p.Y
	s.dx = 0
	s.dy = 0
	s.dl = 2 // start beyond range so the first vertex() emits the move point
	s.ddl = 2
	s.cmd = segMove
}

func (s *segmentator) lineTo(p geom.Pt) {
	s.x1 += s.dx
	s.y1 += s.dy
	s.dx = p.X - s.x1
	s.dy = p.Y - s.y1

	length := math.Sqrt(s.dx*s.dx+s.dy*s.dy) * 1.0
	if length < 1e-30 {
		length = 1e-30
	}
	s.ddl = 1.0 / length

	if s.cmd == segMove {
		s.dl = 0
	} else {
		s.dl = s.ddl
	}
	if s.cmd == segStop {
		s.cmd = segLine
	}
}

func (s *segmentator) vertex() (x, y float64, cmd segCmd) {
	if s.cmd == segStop {
		return 0, 0, segStop
	}
	cmd = s.cmd
	s.cmd = segLine

	if s.dl >= 1.0-s.ddl {
		s.dl = 1.0
		s.cmd = segStop
		return s.x1 + s.dx, s.y1 + s.dy, cmd
	}
	x = s.x1 + s.dx*s.dl
	y = s.y1 + s.dy*s.dl
	s.dl += s.ddl
	return x, y, cmd
}
