package pyplot

import (
	"fmt"
	"strings"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/optarg"
	"github.com/cwbudde/matplotlib-go/ticker"
	"github.com/cwbudde/matplotlib-go/transform"
)

// TickLabelFormatOptions configures ScalarFormatter behavior on the current axes.
//
// Axis accepts "", "both", "x", or "y". Style accepts "", "sci",
// "scientific", or "plain". SciLimits and UseMathText only apply when non-nil.
type TickLabelFormatOptions struct {
	Axis        string
	Style       string
	SciLimits   *[2]int
	UseMathText *bool
}

// XLim sets the current axes x-axis limits.
func XLim(minVal, maxVal float64) {
	GCA().SetXLim(minVal, maxVal)
}

// YLim sets the current axes y-axis limits.
func YLim(minVal, maxVal float64) {
	GCA().SetYLim(minVal, maxVal)
}

// AutoScale updates current axes limits from artist bounds.
func AutoScale(margin float64) {
	GCA().AutoScale(margin)
}

// XScale sets the current axes x-axis scale.
func XScale(name string, opts ...transform.ScaleOption) error {
	return GCA().SetXScale(name, opts...)
}

// YScale sets the current axes y-axis scale.
func YScale(name string, opts ...transform.ScaleOption) error {
	return GCA().SetYScale(name, opts...)
}

// Axis applies a supported Matplotlib-style axis mode to the current axes.
func Axis(mode string) error {
	ax := GCA()
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "off":
		show := false
		ax.ShowFrame = false
		return ax.TickParams(core.TickParams{
			Axis:       "both",
			ShowTicks:  &show,
			ShowLabels: &show,
		})
	case "on":
		show := true
		ax.ShowFrame = true
		return ax.TickParams(core.TickParams{
			Axis:       "both",
			ShowTicks:  &show,
			ShowLabels: &show,
		})
	case "equal":
		return ax.SetAspect("equal")
	case "", "auto":
		return ax.SetAspect("auto")
	default:
		return fmt.Errorf("pyplot: unsupported axis mode %q", mode)
	}
}

// Grid shows or hides grid lines on the current axes, creating grid artists as needed.
func Grid(visible bool, params ...core.TickParams) ([]*core.Grid, error) {
	ax := GCA()
	tickParams := core.TickParams{
		Axis:  "both",
		Which: "major",
	}
	supplied, err := optarg.Only("grid", params)
	if err != nil {
		return nil, err
	}
	if len(params) == 1 {
		tickParams = supplied
		if tickParams.Axis == "" {
			tickParams.Axis = "both"
		}
		if tickParams.Which == "" {
			tickParams.Which = "major"
		}
	}
	tickParams.GridVisible = &visible

	grids, err := ensureGridArtists(ax, tickParams.Axis)
	if err != nil {
		return nil, err
	}
	if err := ax.TickParams(tickParams); err != nil {
		return nil, err
	}
	return grids, nil
}

// TickParams applies tick visibility and styling options to the current axes.
func TickParams(params core.TickParams) error {
	return GCA().TickParams(params)
}

// LocatorParams applies tick locator density options to the current axes.
func LocatorParams(params core.LocatorParams) error {
	return GCA().LocatorParams(params)
}

// MinorTicksOn enables minor ticks on the selected current-axes side.
func MinorTicksOn(axis string) error {
	return GCA().MinorticksOn(axis)
}

// MinorTicksOff disables minor ticks on the selected current-axes side.
func MinorTicksOff(axis string) error {
	return GCA().MinorticksOff(axis)
}

// XTicks sets fixed x-axis tick locations and optional labels on the current axes.
func XTicks(ticks []float64, labels ...[]string) error {
	return setFixedTicks(GCA().XAxis, "x", ticks, labels...)
}

// YTicks sets fixed y-axis tick locations and optional labels on the current axes.
func YTicks(ticks []float64, labels ...[]string) error {
	return setFixedTicks(GCA().YAxis, "y", ticks, labels...)
}

// TickLabelFormat applies ScalarFormatter options to current axes tick labels.
func TickLabelFormat(opts TickLabelFormatOptions) error {
	axes, err := tickLabelFormatAxes(GCA(), opts.Axis)
	if err != nil {
		return err
	}
	formatters := make([]ticker.ScalarFormatter, len(axes))
	for i, target := range axes {
		formatter, ok := target.axis.Formatter.(ticker.ScalarFormatter)
		if !ok {
			return fmt.Errorf("pyplot: %s-axis formatter is %T, want ticker.ScalarFormatter", target.name, target.axis.Formatter)
		}
		formatter, err = applyTickLabelFormat(formatter, opts)
		if err != nil {
			return err
		}
		formatters[i] = formatter
	}
	for i, target := range axes {
		target.axis.Formatter = formatters[i]
	}
	return nil
}

// Title sets the current axes title.
func Title(label string) {
	GCA().SetTitle(label)
}

// XLabel sets the current axes x-axis label.
func XLabel(label string) {
	GCA().SetXLabel(label)
}

// YLabel sets the current axes y-axis label.
func YLabel(label string) {
	GCA().SetYLabel(label)
}

// FigText adds text in figure coordinates to the current figure.
func FigText(x, y float64, text string, opts ...core.TextOptions) *core.Text {
	return GCF().Text(x, y, text, opts...)
}

// Suptitle sets the current figure-level title.
func Suptitle(label string) {
	GCF().SetSuptitle(label)
}

// SupXLabel sets the current figure-level x label.
func SupXLabel(label string) {
	GCF().SetSupXLabel(label)
}

// SupYLabel sets the current figure-level y label.
func SupYLabel(label string) {
	GCF().SetSupYLabel(label)
}

// Box toggles the current axes frame.
func Box(on bool) {
	GCA().ShowFrame = on
}

// TwinX creates a current overlay axes sharing the current axes x-scale.
func TwinX() *core.Axes {
	return setCurrentAxes(GCA().TwinX())
}

// TwinY creates a current overlay axes sharing the current axes y-scale.
func TwinY() *core.Axes {
	return setCurrentAxes(GCA().TwinY())
}

// Legend adds a legend to the current axes.
func Legend() *core.Legend {
	return GCA().AddLegend()
}

// FigLegend adds a figure-level legend to the current figure.
func FigLegend() *core.Legend {
	return GCF().AddLegend()
}

// Colorbar adds a figure-level colorbar for the current axes.
func Colorbar(mappable core.ScalarMappable, opts ...core.ColorbarOptions) *core.Axes {
	ax := GCA()
	fig := GCF()
	opt := optarg.One("colorbar", opts)
	if cb := fig.AddColorbar(ax, mappable, opt); cb != nil {
		return cb
	}
	makeColorbarRoom(ax, opt)
	return fig.AddColorbar(ax, mappable, opt)
}

func ensureGridArtists(ax *core.Axes, axisSpec string) ([]*core.Grid, error) {
	sides, err := gridAxisSides(axisSpec)
	if err != nil {
		return nil, err
	}

	grids := make([]*core.Grid, 0, len(sides))
	for _, side := range sides {
		if grid := findGridArtist(ax, side); grid != nil {
			grids = append(grids, grid)
			continue
		}
		grids = append(grids, ax.AddGrid(side))
	}
	return grids, nil
}

func findGridArtist(ax *core.Axes, side core.AxisSide) *core.Grid {
	if ax == nil {
		return nil
	}
	for _, artist := range ax.Artists {
		grid, ok := artist.(*core.Grid)
		if ok && grid.Axis == side {
			return grid
		}
	}
	return nil
}

func gridAxisSides(axisSpec string) ([]core.AxisSide, error) {
	switch strings.ToLower(strings.TrimSpace(axisSpec)) {
	case "", "both":
		return []core.AxisSide{core.AxisBottom, core.AxisLeft}, nil
	case "x", "bottom":
		return []core.AxisSide{core.AxisBottom}, nil
	case "top":
		return []core.AxisSide{core.AxisTop}, nil
	case "y", "left":
		return []core.AxisSide{core.AxisLeft}, nil
	case "right":
		return []core.AxisSide{core.AxisRight}, nil
	default:
		return nil, fmt.Errorf("pyplot: unsupported grid axis %q", axisSpec)
	}
}

type tickLabelFormatTarget struct {
	name string
	axis *core.Axis
}

func tickLabelFormatAxes(ax *core.Axes, axisSpec string) ([]tickLabelFormatTarget, error) {
	switch strings.ToLower(strings.TrimSpace(axisSpec)) {
	case "", "both":
		return []tickLabelFormatTarget{
			{name: "x", axis: ax.XAxis},
			{name: "y", axis: ax.YAxis},
		}, nil
	case "x":
		return []tickLabelFormatTarget{{name: "x", axis: ax.XAxis}}, nil
	case "y":
		return []tickLabelFormatTarget{{name: "y", axis: ax.YAxis}}, nil
	default:
		return nil, fmt.Errorf("pyplot: unsupported ticklabel_format axis %q", axisSpec)
	}
}

func applyTickLabelFormat(formatter ticker.ScalarFormatter, opts TickLabelFormatOptions) (ticker.ScalarFormatter, error) {
	switch strings.ToLower(strings.TrimSpace(opts.Style)) {
	case "":
	case "sci", "scientific":
		formatter.DisableScientific = false
	case "plain":
		formatter.DisableScientific = true
	default:
		return formatter, fmt.Errorf("pyplot: unsupported ticklabel_format style %q", opts.Style)
	}
	if opts.SciLimits != nil {
		formatter.UsePowerLimits = true
		formatter.PowerLimits = *opts.SciLimits
	}
	if opts.UseMathText != nil {
		formatter.UseMathText = *opts.UseMathText
	}
	return formatter, nil
}

func setFixedTicks(axis *core.Axis, name string, ticks []float64, labels ...[]string) error {
	if axis == nil {
		return fmt.Errorf("pyplot: nil %s axis", name)
	}
	if len(labels) > 1 {
		return fmt.Errorf("pyplot: %sticks accepts at most one label set", name)
	}
	copiedTicks := append([]float64(nil), ticks...)
	axis.Locator = ticker.FixedLocator{TicksList: copiedTicks}
	if len(labels) == 0 {
		return nil
	}
	if len(labels[0]) != len(ticks) {
		return fmt.Errorf("pyplot: %sticks label count %d does not match tick count %d", name, len(labels[0]), len(ticks))
	}
	axis.Formatter = ticker.FixedFormatter{Labels: append([]string(nil), labels[0]...)}
	return nil
}

//nolint:gocritic // ColorbarOptions is read-only here; only Width and Padding are consulted.
func makeColorbarRoom(ax *core.Axes, opt core.ColorbarOptions) {
	if ax == nil {
		return
	}

	width := 0.035
	padding := 0.02
	if opt.Width > 0 {
		width = opt.Width
	}
	if opt.Padding > 0 {
		padding = opt.Padding
	}

	maxRight := 1 - padding - width
	if ax.RectFraction.Max.X <= maxRight {
		return
	}
	if maxRight <= ax.RectFraction.Min.X {
		return
	}
	ax.RectFraction.Max.X = maxRight
}
