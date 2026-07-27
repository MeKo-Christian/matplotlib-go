package color

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/render"
)

// ErrUnknownColormap is returned by LookupColormapStrict for a name that is not
// registered. The lenient LookupColormap warns and falls back to the default
// colormap instead of returning this error.
var ErrUnknownColormap = errors.New("unknown colormap")

// ColorStop defines a point in a piecewise-linear colormap.
// Colors are interpolated component-wise in render.Color space.
type ColorStop struct {
	Pos   float64
	Color render.Color
}

// Colormap maps normalized values in [0,1] to colors.
type Colormap struct {
	name     string
	stops    []ColorStop
	listed   []render.Color
	bad      render.Color
	under    render.Color
	over     render.Color
	hasBad   bool
	hasUnder bool
	hasOver  bool
}

// Name returns the user-visible colormap name.
func (c Colormap) Name() string {
	return c.name
}

// Copy returns an independent copy of the colormap.
func (c Colormap) Copy(name string) Colormap {
	out := c
	out.name = normalizeColormapName(name)
	if out.name == "" {
		out.name = c.name
	}
	out.stops = append([]ColorStop(nil), c.stops...)
	out.listed = append([]render.Color(nil), c.listed...)
	return out
}

// Reversed returns an independent copy with the colormap direction reversed.
func (c Colormap) Reversed(name string) Colormap {
	out := c.Copy(name)
	if out.name == c.name {
		out.name = c.name + "_r"
	}
	if len(out.listed) > 0 {
		for i, j := 0, len(out.listed)-1; i < j; i, j = i+1, j-1 {
			out.listed[i], out.listed[j] = out.listed[j], out.listed[i]
		}
	}
	if len(out.stops) > 0 {
		rev := make([]ColorStop, len(out.stops))
		for i, stop := range out.stops {
			rev[len(out.stops)-1-i] = ColorStop{
				Pos:   1 - stop.Pos,
				Color: stop.Color,
			}
		}
		out.stops = rev
	}
	return out
}

// Resampled returns a listed colormap with n samples from this colormap.
func (c Colormap) Resampled(n int) Colormap {
	out := c.Copy(c.name)
	if n <= 0 {
		n = 1
	}
	out.stops = nil
	out.listed = make([]render.Color, n)
	if n == 1 {
		out.listed[0] = c.At(0)
		return out
	}
	for i := range out.listed {
		out.listed[i] = c.At(float64(i) / float64(n-1))
	}
	return out
}

// At returns a color for normalized input t in [0,1].
// Interpolation is component-wise in normalized render.Color space.
func (c Colormap) At(t float64) render.Color {
	if len(c.listed) > 0 {
		v := clamp01(t)
		idx := int(v * float64(len(c.listed)))
		if idx >= len(c.listed) {
			idx = len(c.listed) - 1
		}
		return c.listed[idx]
	}

	if len(c.stops) == 0 {
		return render.Color{R: 0, G: 0, B: 0, A: 1}
	}

	v := clamp01(t)

	if v <= c.stops[0].Pos {
		return c.stops[0].Color
	}
	if v >= c.stops[len(c.stops)-1].Pos {
		return c.stops[len(c.stops)-1].Color
	}

	for i := 1; i < len(c.stops); i++ {
		left := c.stops[i-1]
		right := c.stops[i]
		if v < right.Pos {
			if right.Pos-left.Pos <= 0 {
				return right.Color
			}
			local := (v - left.Pos) / (right.Pos - left.Pos)
			return mixColor(left.Color, right.Color, local)
		}
	}

	return c.stops[len(c.stops)-1].Color
}

// AtValue returns a color for a possibly out-of-range normalized value.
// NaN uses the bad color, values below zero use under (this includes -inf), and
// values above one use over (this includes +inf). This mirrors Matplotlib's
// Colormap.__call__, where a norm emits -inf/+inf for out-of-range inputs and
// only masked (NaN) values map to the bad color. When under/over/bad are not
// configured, the colormap endpoints (bad → transparent) are used, matching
// Matplotlib's default behavior.
func (c Colormap) AtValue(t float64) render.Color {
	if math.IsNaN(t) {
		if c.hasBad {
			return c.bad
		}
		return render.Color{}
	}
	if t < 0 {
		if c.hasUnder {
			return c.under
		}
		return c.At(0)
	}
	if t > 1 {
		if c.hasOver {
			return c.over
		}
		return c.At(1)
	}
	return c.At(t)
}

// WithBad returns a copy of the colormap with a configured bad-value color.
func (c Colormap) WithBad(color render.Color) Colormap {
	c.bad = color
	c.hasBad = true
	return c
}

// SetBad configures the bad-value color in place.
func (c *Colormap) SetBad(color render.Color) {
	if c == nil {
		return
	}
	c.bad = color
	c.hasBad = true
}

// WithUnder returns a copy of the colormap with a configured under-range color.
func (c Colormap) WithUnder(color render.Color) Colormap {
	c.under = color
	c.hasUnder = true
	return c
}

// SetUnder configures the under-range color in place.
func (c *Colormap) SetUnder(color render.Color) {
	if c == nil {
		return
	}
	c.under = color
	c.hasUnder = true
}

// WithOver returns a copy of the colormap with a configured over-range color.
func (c Colormap) WithOver(color render.Color) Colormap {
	c.over = color
	c.hasOver = true
	return c
}

// SetOver configures the over-range color in place.
func (c *Colormap) SetOver(color render.Color) {
	if c == nil {
		return
	}
	c.over = color
	c.hasOver = true
}

// NewColormap creates a new linear colormap from color stops.
// Stops are sorted by Pos and clamped to [0,1].
func NewColormap(name string, stops []ColorStop) Colormap {
	norm := normalizeColormapName(name)
	if len(stops) == 0 {
		return Colormap{name: norm, stops: []ColorStop{
			{Pos: 0, Color: render.Color{R: 0, G: 0, B: 0, A: 1}},
			{Pos: 1, Color: render.Color{R: 1, G: 1, B: 1, A: 1}},
		}}
	}

	normalized := make([]ColorStop, len(stops))
	for i, stop := range stops {
		normalized[i] = ColorStop{
			Pos:   clamp01(stop.Pos),
			Color: stop.Color,
		}
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		return normalized[i].Pos < normalized[j].Pos
	})

	return Colormap{name: norm, stops: normalized}
}

// NewLinearSegmentedColormap creates a piecewise-linear colormap from color stops.
func NewLinearSegmentedColormap(name string, stops []ColorStop) Colormap {
	return NewColormap(name, stops)
}

// NewListedColormap creates a colormap from a discrete color lookup table.
func NewListedColormap(name string, colors []render.Color) Colormap {
	norm := normalizeColormapName(name)
	listed := append([]render.Color(nil), colors...)
	if len(listed) == 0 {
		listed = []render.Color{{R: 0, G: 0, B: 0, A: 1}}
	}
	return Colormap{name: norm, listed: listed}
}

var defaultColormapName = "viridis"

// colormapMu guards colormaps. Every read or write of the map must hold it:
// draws resolve colormaps concurrently with user RegisterColormap calls.
var colormapMu sync.RWMutex

var colormaps = map[string]Colormap{
	"viridis": {
		name: "viridis",
		stops: []ColorStop{
			{0.0000, render.Color{R: 0.267004, G: 0.004874, B: 0.329415, A: 1}},
			{0.0625, render.Color{R: 0.282327, G: 0.094955, B: 0.417331, A: 1}},
			{0.1250, render.Color{R: 0.278826, G: 0.175490, B: 0.483397, A: 1}},
			{0.1875, render.Color{R: 0.258965, G: 0.251537, B: 0.524736, A: 1}},
			{0.2500, render.Color{R: 0.229739, G: 0.322361, B: 0.545706, A: 1}},
			{0.3125, render.Color{R: 0.199430, G: 0.387607, B: 0.554642, A: 1}},
			{0.3750, render.Color{R: 0.172719, G: 0.448791, B: 0.557885, A: 1}},
			{0.4375, render.Color{R: 0.149039, G: 0.508051, B: 0.557250, A: 1}},
			{0.5000, render.Color{R: 0.127568, G: 0.566949, B: 0.550556, A: 1}},
			{0.5625, render.Color{R: 0.120638, G: 0.625828, B: 0.533488, A: 1}},
			{0.6250, render.Color{R: 0.157851, G: 0.683765, B: 0.501686, A: 1}},
			{0.6875, render.Color{R: 0.246070, G: 0.738910, B: 0.452024, A: 1}},
			{0.7500, render.Color{R: 0.369214, G: 0.788888, B: 0.382914, A: 1}},
			{0.8125, render.Color{R: 0.515992, G: 0.831158, B: 0.294279, A: 1}},
			{0.8750, render.Color{R: 0.678489, G: 0.863742, B: 0.189503, A: 1}},
			{0.9375, render.Color{R: 0.845561, G: 0.887322, B: 0.099702, A: 1}},
			{1.0000, render.Color{R: 0.993248, G: 0.906157, B: 0.143936, A: 1}},
		},
	},
	"gray": {
		name: "gray",
		stops: []ColorStop{
			{0, render.Color{R: 0, G: 0, B: 0, A: 1}},
			{1, render.Color{R: 1, G: 1, B: 1, A: 1}},
		},
	},
	"binary": {
		name: "binary",
		stops: []ColorStop{
			{0, render.Color{R: 1, G: 1, B: 1, A: 1}},
			{1, render.Color{R: 0, G: 0, B: 0, A: 1}},
		},
	},
	"red channel": {
		name: "red channel",
		stops: []ColorStop{
			{0, render.Color{R: 0.18, G: 0.02, B: 0.02, A: 1}},
			{1, render.Color{R: 1.00, G: 0.18, B: 0.12, A: 1}},
		},
	},
	"green channel": {
		name: "green channel",
		stops: []ColorStop{
			{0, render.Color{R: 0.02, G: 0.14, B: 0.05, A: 1}},
			{1, render.Color{R: 0.20, G: 0.90, B: 0.28, A: 1}},
		},
	},
	"blue channel": {
		name: "blue channel",
		stops: []ColorStop{
			{0, render.Color{R: 0.02, G: 0.05, B: 0.18, A: 1}},
			{1, render.Color{R: 0.18, G: 0.45, B: 1.00, A: 1}},
		},
	},
	"plasma": {
		name: "plasma",
		stops: []ColorStop{
			{0.0000, render.Color{R: 0.050383, G: 0.029803, B: 0.527975, A: 1}},
			{0.0625, render.Color{R: 0.193374, G: 0.018354, B: 0.590330, A: 1}},
			{0.1250, render.Color{R: 0.299855, G: 0.009561, B: 0.631624, A: 1}},
			{0.1875, render.Color{R: 0.399411, G: 0.000859, B: 0.656133, A: 1}},
			{0.2500, render.Color{R: 0.494877, G: 0.011990, B: 0.657865, A: 1}},
			{0.3125, render.Color{R: 0.584391, G: 0.068579, B: 0.632812, A: 1}},
			{0.3750, render.Color{R: 0.665129, G: 0.138566, B: 0.585582, A: 1}},
			{0.4375, render.Color{R: 0.736019, G: 0.209439, B: 0.527908, A: 1}},
			{0.5000, render.Color{R: 0.798216, G: 0.280197, B: 0.469538, A: 1}},
			{0.5625, render.Color{R: 0.853319, G: 0.351553, B: 0.413734, A: 1}},
			{0.6250, render.Color{R: 0.901807, G: 0.425087, B: 0.359688, A: 1}},
			{0.6875, render.Color{R: 0.942598, G: 0.502639, B: 0.305816, A: 1}},
			{0.7500, render.Color{R: 0.973416, G: 0.585761, B: 0.251540, A: 1}},
			{0.8125, render.Color{R: 0.991365, G: 0.675355, B: 0.198453, A: 1}},
			{0.8750, render.Color{R: 0.993033, G: 0.771720, B: 0.154808, A: 1}},
			{0.9375, render.Color{R: 0.974443, G: 0.874622, B: 0.144061, A: 1}},
			{1.0000, render.Color{R: 0.940015, G: 0.975158, B: 0.131326, A: 1}},
		},
	},
	"inferno": {
		name: "inferno",
		stops: []ColorStop{
			{0.000000, render.Color{R: 0.001462, G: 0.000466, B: 0.013866, A: 1}},
			{0.032258, render.Color{R: 0.013995, G: 0.011225, B: 0.071862, A: 1}},
			{0.064516, render.Color{R: 0.042253, G: 0.028139, B: 0.141141, A: 1}},
			{0.096774, render.Color{R: 0.081962, G: 0.043328, B: 0.215289, A: 1}},
			{0.129032, render.Color{R: 0.135778, G: 0.046856, B: 0.299776, A: 1}},
			{0.161290, render.Color{R: 0.190367, G: 0.039309, B: 0.361447, A: 1}},
			{0.193548, render.Color{R: 0.244967, G: 0.037055, B: 0.400007, A: 1}},
			{0.225806, render.Color{R: 0.297178, G: 0.047470, B: 0.420491, A: 1}},
			{0.258065, render.Color{R: 0.354032, G: 0.066925, B: 0.430906, A: 1}},
			{0.290323, render.Color{R: 0.403894, G: 0.085580, B: 0.433179, A: 1}},
			{0.322581, render.Color{R: 0.453651, G: 0.103848, B: 0.430498, A: 1}},
			{0.354839, render.Color{R: 0.503493, G: 0.121575, B: 0.423356, A: 1}},
			{0.387097, render.Color{R: 0.559624, G: 0.141346, B: 0.410078, A: 1}},
			{0.419355, render.Color{R: 0.609330, G: 0.159474, B: 0.393589, A: 1}},
			{0.451613, render.Color{R: 0.658463, G: 0.178962, B: 0.372748, A: 1}},
			{0.483871, render.Color{R: 0.706500, G: 0.200728, B: 0.347777, A: 1}},
			{0.516129, render.Color{R: 0.758422, G: 0.229097, B: 0.315266, A: 1}},
			{0.548387, render.Color{R: 0.801871, G: 0.258674, B: 0.283099, A: 1}},
			{0.580645, render.Color{R: 0.841969, G: 0.292933, B: 0.248564, A: 1}},
			{0.612903, render.Color{R: 0.878001, G: 0.332060, B: 0.212268, A: 1}},
			{0.645161, render.Color{R: 0.912966, G: 0.381636, B: 0.169755, A: 1}},
			{0.677419, render.Color{R: 0.938675, G: 0.430091, B: 0.130438, A: 1}},
			{0.709677, render.Color{R: 0.959114, G: 0.482014, B: 0.089499, A: 1}},
			{0.741935, render.Color{R: 0.974176, G: 0.536780, B: 0.048392, A: 1}},
			{0.774194, render.Color{R: 0.984591, G: 0.601122, B: 0.023606, A: 1}},
			{0.806452, render.Color{R: 0.987926, G: 0.660250, B: 0.051750, A: 1}},
			{0.838710, render.Color{R: 0.985566, G: 0.720782, B: 0.112229, A: 1}},
			{0.870968, render.Color{R: 0.977497, G: 0.782258, B: 0.185923, A: 1}},
			{0.903226, render.Color{R: 0.962517, G: 0.851476, B: 0.285546, A: 1}},
			{0.935484, render.Color{R: 0.948683, G: 0.910473, B: 0.395289, A: 1}},
			{0.967742, render.Color{R: 0.951740, G: 0.960587, B: 0.524203, A: 1}},
			{1.000000, render.Color{R: 0.988362, G: 0.998364, B: 0.644924, A: 1}},
		},
	},
	"blues": {
		name: "blues",
		stops: []ColorStop{
			{0.000, render.Color{R: 0.968627, G: 0.984314, B: 1.000000, A: 1}},
			{0.125, render.Color{R: 0.870588, G: 0.921569, B: 0.968627, A: 1}},
			{0.250, render.Color{R: 0.776471, G: 0.858824, B: 0.937255, A: 1}},
			{0.375, render.Color{R: 0.619608, G: 0.792157, B: 0.882353, A: 1}},
			{0.500, render.Color{R: 0.419608, G: 0.682353, B: 0.839216, A: 1}},
			{0.625, render.Color{R: 0.258824, G: 0.572549, B: 0.776471, A: 1}},
			{0.750, render.Color{R: 0.129412, G: 0.443137, B: 0.709804, A: 1}},
			{0.875, render.Color{R: 0.031373, G: 0.317647, B: 0.611765, A: 1}},
			{1.000, render.Color{R: 0.031373, G: 0.188235, B: 0.419608, A: 1}},
		},
	},
}

// RegisterColormap adds a named colormap to the runtime registry. It is safe
// for concurrent use with the lookup functions.
func RegisterColormap(name string, cmap Colormap) {
	key := normalizeColormapName(name)
	if key == "" {
		return
	}
	cmap.name = key
	colormapMu.Lock()
	colormaps[key] = cmap
	colormapMu.Unlock()
}

// LookupColormap returns a colormap by name. Unknown names warn (via internal/diag)
// and fall back to the default `viridis` colormap, so a typo renders a
// plausible-but-wrong plot rather than failing. Callers that want Matplotlib's
// strict behavior (a hard error on an unknown name) should use LookupColormapStrict.
func LookupColormap(name string) Colormap {
	if cmap, err := LookupColormapStrict(name); err == nil {
		return cmap
	}
	diag.Warnf("unknown colormap %q; falling back to %q", name, defaultColormapName)
	colormapMu.RLock()
	cmap := colormaps[defaultColormapName]
	colormapMu.RUnlock()
	return cmap
}

// LookupColormapStrict returns the colormap registered under name, or a wrapped
// ErrUnknownColormap if none is registered. Unlike LookupColormap it never falls
// back to a default, mirroring Matplotlib's `matplotlib.colormaps[name]`
// KeyError.
//
// Name matching is currently case-insensitive because the registry stores
// normalized (lower-cased) keys; Matplotlib is case-sensitive ("Blues" !=
// "blues"). Honoring case sensitivity requires re-registering the builtin
// colormaps under their canonical mixed-case names, which has not been done.
func LookupColormapStrict(name string) (Colormap, error) {
	key := normalizeColormapName(name)
	colormapMu.RLock()
	defer colormapMu.RUnlock()
	if cmap, ok := colormaps[key]; ok {
		return cmap, nil
	}
	if baseKey, ok := strings.CutSuffix(key, "_r"); ok {
		if cmap, ok := colormaps[baseKey]; ok {
			return cmap.Reversed(key), nil
		}
	}
	return Colormap{}, fmt.Errorf("%w: %q", ErrUnknownColormap, name)
}

// DefaultColormap returns the configured default colormap.
func DefaultColormap() Colormap {
	return LookupColormap(defaultColormapName)
}

// ColormapNames returns the registered base colormap names in sorted order.
func ColormapNames() []string {
	colormapMu.RLock()
	names := make([]string, 0, len(colormaps))
	for name := range colormaps {
		names = append(names, name)
	}
	colormapMu.RUnlock()
	sort.Strings(names)
	return names
}

func mixColor(a, b render.Color, t float64) render.Color {
	return render.Color{
		R: a.R + (b.R-a.R)*t,
		G: a.G + (b.G-a.G)*t,
		B: a.B + (b.B-a.B)*t,
		A: a.A + (b.A-a.A)*t,
	}
}

func normalizeColormapName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	if math.IsNaN(v) {
		return 0
	}
	return v
}
