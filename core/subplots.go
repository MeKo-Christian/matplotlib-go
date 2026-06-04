package core

// ShareMode controls how subplot axes reuse scale and axis state.
type ShareMode uint8

const (
	ShareNone ShareMode = iota
	ShareAll
	ShareRow
	ShareCol
)

// SubplotOptions controls automatic subplot placement and axis sharing.
type SubplotOptions struct {
	// Figure-normalized margins.
	Left, Right, Bottom, Top float64

	// Normalized inter-subplot spacing.
	WSpace float64
	HSpace float64

	// Share modes determine how x/y scales and axis controls are reused.
	ShareXMode ShareMode
	ShareYMode ShareMode
}

// SubplotOption configures the behavior of Subplots.
type SubplotOption func(*SubplotOptions)

// WithSubplotPadding overrides the subplot figure margins.
func WithSubplotPadding(left, right, bottom, top float64) SubplotOption {
	return func(cfg *SubplotOptions) {
		cfg.Left = left
		cfg.Right = right
		cfg.Bottom = bottom
		cfg.Top = top
	}
}

// WithSubplotSpacing overrides the spacing between subplot cells.
func WithSubplotSpacing(wspace, hspace float64) SubplotOption {
	return func(cfg *SubplotOptions) {
		cfg.WSpace = wspace
		cfg.HSpace = hspace
	}
}

// WithSubplotShareX shares x scales and x-axis settings across all subplots.
func WithSubplotShareX() SubplotOption {
	return func(cfg *SubplotOptions) {
		cfg.ShareXMode = ShareAll
	}
}

// WithSubplotShareY shares y scales and y-axis settings across all subplots.
func WithSubplotShareY() SubplotOption {
	return func(cfg *SubplotOptions) {
		cfg.ShareYMode = ShareAll
	}
}

// WithSubplotShareBoth shares x and y scales and axis settings across all subplots.
func WithSubplotShareBoth() SubplotOption {
	return func(cfg *SubplotOptions) {
		cfg.ShareXMode = ShareAll
		cfg.ShareYMode = ShareAll
	}
}

// WithSubplotShareXMode configures how x scales are shared within a subplot grid.
func WithSubplotShareXMode(mode ShareMode) SubplotOption {
	return func(cfg *SubplotOptions) {
		cfg.ShareXMode = mode
	}
}

// WithSubplotShareYMode configures how y scales are shared within a subplot grid.
func WithSubplotShareYMode(mode ShareMode) SubplotOption {
	return func(cfg *SubplotOptions) {
		cfg.ShareYMode = mode
	}
}

func defaultSubplotOptions() SubplotOptions {
	return SubplotOptions{
		Left:   matplotlibSubplotLeft,
		Right:  matplotlibSubplotRight,
		Bottom: matplotlibSubplotBottom,
		Top:    matplotlibSubplotTop,
		WSpace: 0.05,
		HSpace: 0.06,
	}
}

// Subplots creates an nRows x nCols grid of axes with automatic layout.
func (f *Figure) Subplots(nRows, nCols int, opts ...SubplotOption) [][]*Axes {
	cfg := defaultSubplotOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	gs := f.GridSpec(
		nRows,
		nCols,
		subplotGridSpecOptions(cfg.Left, cfg.Right, cfg.Bottom, cfg.Top, cfg.WSpace, cfg.HSpace)...,
	)
	if gs == nil {
		return nil
	}

	grid := make([][]*Axes, nRows)
	for row := 0; row < nRows; row++ {
		rowAxes := make([]*Axes, nCols)
		for col := 0; col < nCols; col++ {
			rowAxes[col] = gs.Cell(row, col).AddAxes()
		}
		grid[row] = rowAxes
	}

	applyGridShareMode(grid, cfg.ShareXMode, true)
	applyGridShareMode(grid, cfg.ShareYMode, false)

	return grid
}

func subplotGridSpecOptions(left, right, bottom, top, wspace, hspace float64) []GridSpecOption {
	gridW := 0.0
	gridH := 0.0
	if width := right - left; width > 0 {
		gridW = wspace / width
	}
	if height := top - bottom; height > 0 {
		gridH = hspace / height
	}
	return []GridSpecOption{
		WithGridSpecPadding(left, right, bottom, top),
		WithGridSpecSpacing(gridW, gridH),
	}
}

func applyGridShareMode(grid [][]*Axes, mode ShareMode, isX bool) {
	if len(grid) == 0 || len(grid[0]) == 0 || mode == ShareNone {
		return
	}

	for row := range grid {
		for col := range grid[row] {
			ax := grid[row][col]
			if ax == nil {
				continue
			}

			root := sharedAxesRoot(grid, row, col, mode, isX)
			if root == nil || root == ax {
				continue
			}

			if isX {
				ax.shareX = root.xScaleRoot()
				ax.XAxis = root.xScaleRoot().XAxis
			} else {
				ax.shareY = root.yScaleRoot()
				ax.YAxis = root.yScaleRoot().YAxis
			}
		}
	}
}

func sharedAxesRoot(grid [][]*Axes, row, col int, mode ShareMode, isX bool) *Axes {
	switch mode {
	case ShareAll:
		return grid[0][0]
	case ShareRow:
		if col == 0 {
			return nil
		}
		return grid[row][0]
	case ShareCol:
		if row == 0 {
			return nil
		}
		return grid[0][col]
	default:
		return nil
	}
}
