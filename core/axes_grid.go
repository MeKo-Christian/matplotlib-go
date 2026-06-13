package core

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/style"
)

func (a *Axes) AddGrid(axis AxisSide) *Grid {
	grid := NewGrid(axis)
	rc := a.resolvedRC()
	grid.Color = rc.GridColor
	grid.LineWidth = rc.GridLineWidth
	grid.MinorColor = rc.MinorGridColor
	grid.MinorLineWidth = rc.MinorGridLineWidth
	grid.Dashes = styleCloneDashes(rc.GridDashes)
	grid.MinorDashes = styleCloneDashes(rc.MinorGridDashes)
	if a.axisBelowSet {
		grid.z = a.axisBelowZ
	}
	a.Add(grid)
	return grid
}

func (a *Axes) AddXGrid() *Grid {
	return a.AddGrid(AxisBottom)
}

func (a *Axes) AddYGrid() *Grid {
	return a.AddGrid(AxisLeft)
}

func (a *Axes) addDefaultGrids(rc style.RC) {
	if a == nil || !rc.GridVisible {
		return
	}

	addGrid := func(axis AxisSide) {
		grid := a.AddGrid(axis)
		grid.Dashes = styleCloneDashes(rc.GridDashes)
		grid.MinorDashes = styleCloneDashes(rc.MinorGridDashes)
		switch strings.ToLower(strings.TrimSpace(rc.GridWhich)) {
		case "minor":
			grid.Major = false
			grid.Minor = true
		case "both":
			grid.Major = true
			grid.Minor = true
		default:
			grid.Major = true
			grid.Minor = false
		}
	}

	switch strings.ToLower(strings.TrimSpace(rc.GridAxis)) {
	case "x":
		addGrid(AxisBottom)
	case "y":
		addGrid(AxisLeft)
	default:
		addGrid(AxisBottom)
		addGrid(AxisLeft)
	}
}

// SetAxisBelow mirrors Matplotlib's Axes.set_axisbelow for the boolean modes.
// true places gridlines below default patches (z=0.5); false places them above
// default lines (z=2.5). The default grid z-order remains Matplotlib's
// axisbelow="line" behavior (z=1.5).
func (a *Axes) SetAxisBelow(below bool) {
	if a == nil {
		return
	}
	z := 2.5
	if below {
		z = 0.5
	}
	a.axisBelowSet = true
	a.axisBelowZ = z
	for _, art := range a.Artists {
		if grid, ok := art.(*Grid); ok {
			grid.z = z
		}
	}
	a.zsorted = false
}
