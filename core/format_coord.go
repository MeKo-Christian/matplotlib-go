package core

import (
	"fmt"
	"strconv"

	"github.com/cwbudde/matplotlib-go/geom"
)

// CoordFormatter renders a (data x, data y) pair as a status-bar string.
// Implementations may format using axis-specific units or precision.
type CoordFormatter func(p geom.Pt) string

// FormatCoord turns a figure-pixel position into the human-readable string
// Matplotlib's status bar shows. Returns the empty string when the point lies
// outside the axes or the axes has no usable transform.
//
// If a custom [CoordFormatter] is registered via [Axes.SetFormatCoord] it is
// used in place of the default x={x:g} y={y:g} rendering.
func (a *Axes) FormatCoord(p geom.Pt) string {
	if a == nil {
		return ""
	}
	data, ok := a.PixelToData(p)
	if !ok {
		return ""
	}
	if a.coordFormatter != nil {
		return a.coordFormatter(data)
	}
	return defaultCoordString(data)
}

// SetFormatCoord installs a custom coordinate formatter; pass nil to restore
// the default.
func (a *Axes) SetFormatCoord(f CoordFormatter) {
	if a == nil {
		return
	}
	a.coordFormatter = f
}

func defaultCoordString(p geom.Pt) string {
	return fmt.Sprintf("x=%s y=%s", formatCoordScalar(p.X), formatCoordScalar(p.Y))
}

func formatCoordScalar(v float64) string {
	// Matplotlib uses '%1.4g'; keep parity within reason.
	return strconv.FormatFloat(v, 'g', 4, 64)
}
