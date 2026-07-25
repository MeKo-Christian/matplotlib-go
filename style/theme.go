package style

import (
	"sort"
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/render"
)

// Theme is a named stylesheet preset backed by a fully resolved RC.
type Theme struct {
	Name string
	RC   RC
}

var (
	// ThemeDefault preserves the library defaults.
	ThemeDefault = NewTheme("default")

	// ThemeGGPlot mirrors the broad look of Matplotlib's ggplot style.
	ThemeGGPlot = NewTheme(
		"ggplot",
		WithFont("DejaVu Sans", 10),
		WithBackground(1, 1, 1, 1),
		WithAxesBackground(render.Color{R: 0.898, G: 0.898, B: 0.898, A: 1}),
		WithAxesEdgeColor(render.Color{R: 1, G: 1, B: 1, A: 1}),
		WithGridColors(
			render.Color{R: 1, G: 1, B: 1, A: 1},
			render.Color{R: 1, G: 1, B: 1, A: 0.55},
		),
		WithGridLineWidths(1.4, 0.8),
		WithColorCycle(color.Palette{
			{R: 0.886, G: 0.290, B: 0.200, A: 1},
			{R: 0.204, G: 0.541, B: 0.741, A: 1},
			{R: 0.596, G: 0.557, B: 0.835, A: 1},
			{R: 0.467, G: 0.467, B: 0.467, A: 1},
			{R: 0.984, G: 0.757, B: 0.369, A: 1},
			{R: 0.557, G: 0.729, B: 0.259, A: 1},
			{R: 1.000, G: 0.710, B: 0.722, A: 1},
		}),
		WithLegendColors(
			render.Color{R: 1, G: 1, B: 1, A: 0.85},
			render.Color{R: 1, G: 1, B: 1, A: 0},
			render.Color{R: 0.18, G: 0.18, B: 0.18, A: 1},
		),
	)

	// ThemeDarkBackground is a dark presentation theme.
	ThemeDarkBackground = NewTheme(
		"dark_background",
		WithTextColor(1, 1, 1, 1),
		WithLineColor(1, 1, 1, 1),
		WithBackground(0, 0, 0, 1),
		WithAxesBackground(render.Color{R: 0, G: 0, B: 0, A: 1}),
		WithAxesEdgeColor(render.Color{R: 1, G: 1, B: 1, A: 1}),
		WithGridColors(
			render.Color{R: 1, G: 1, B: 1, A: 0.35},
			render.Color{R: 1, G: 1, B: 1, A: 0.18},
		),
		WithColorCycle(color.Palette{
			{R: 0.553, G: 0.827, B: 0.780, A: 1},
			{R: 0.996, G: 1.000, B: 0.702, A: 1},
			{R: 0.749, G: 0.733, B: 0.851, A: 1},
			{R: 0.980, G: 0.506, B: 0.455, A: 1},
			{R: 0.506, G: 0.694, B: 0.824, A: 1},
			{R: 0.992, G: 0.706, B: 0.384, A: 1},
		}),
		WithLegendColors(
			render.Color{R: 0.05, G: 0.05, B: 0.05, A: 0.85},
			render.Color{R: 1, G: 1, B: 1, A: 0.2},
			render.Color{R: 1, G: 1, B: 1, A: 1},
		),
	)

	// ThemePublication is a restrained light theme for papers and reports.
	ThemePublication = NewTheme(
		"publication",
		WithDPI(144),
		WithFont("DejaVu Sans", 11),
		WithTextColor(0.12, 0.12, 0.12, 1),
		WithLineColor(0.12, 0.12, 0.12, 1),
		WithBackground(1, 1, 1, 1),
		WithAxesBackground(render.Color{R: 1, G: 1, B: 1, A: 1}),
		WithAxesEdgeColor(render.Color{R: 0.16, G: 0.16, B: 0.16, A: 1}),
		WithAxisLineWidth(1.0),
		WithGridColors(
			render.Color{R: 0.86, G: 0.86, B: 0.86, A: 1},
			render.Color{R: 0.92, G: 0.92, B: 0.92, A: 1},
		),
		WithGridLineWidths(0.8, 0.5),
		WithColorCycle(color.Palette{
			{R: 0.247, G: 0.565, B: 0.855, A: 1},
			{R: 0.922, G: 0.557, B: 0.118, A: 1},
			{R: 0.463, G: 0.667, B: 0.192, A: 1},
			{R: 0.781, G: 0.282, B: 0.314, A: 1},
			{R: 0.580, G: 0.353, B: 0.706, A: 1},
			{R: 0.167, G: 0.631, B: 0.596, A: 1},
		}),
		WithLegendColors(
			render.Color{R: 1, G: 1, B: 1, A: 0.96},
			render.Color{R: 0.25, G: 0.25, B: 0.25, A: 0.2},
			render.Color{R: 0.12, G: 0.12, B: 0.12, A: 1},
		),
	)
)

var themeRegistry = map[string]Theme{
	ThemeDefault.Name:        ThemeDefault,
	ThemeGGPlot.Name:         ThemeGGPlot,
	ThemeDarkBackground.Name: ThemeDarkBackground,
	ThemePublication.Name:    ThemePublication,
}

var themeRegistryMu sync.RWMutex

// NewTheme creates a named theme from the default RC plus overrides.
func NewTheme(name string, opts ...Option) Theme {
	normalized := normalizeThemeName(name)
	if normalized == "" {
		normalized = "custom"
	}
	return Theme{
		Name: normalized,
		RC:   Apply(Default, opts...),
	}
}

// RegisterTheme adds or replaces a named theme in the runtime registry.
func RegisterTheme(theme Theme) {
	normalized := normalizeThemeName(theme.Name)
	if normalized == "" {
		return
	}
	theme.Name = normalized
	theme.RC = Apply(theme.RC)
	themeRegistryMu.Lock()
	defer themeRegistryMu.Unlock()
	themeRegistry[normalized] = theme
}

// LookupTheme returns a named theme and whether it was found.
func LookupTheme(name string) (Theme, bool) {
	themeRegistryMu.RLock()
	defer themeRegistryMu.RUnlock()
	theme, ok := themeRegistry[normalizeThemeName(name)]
	if !ok {
		return Theme{}, false
	}
	return Theme{Name: theme.Name, RC: Apply(theme.RC)}, true
}

// MustTheme returns the named theme or the default theme when it is missing.
func MustTheme(name string) Theme {
	theme, ok := LookupTheme(name)
	if !ok {
		return ThemeDefault
	}
	return theme
}

// AvailableThemes returns the registered theme names in stable order.
func AvailableThemes() []string {
	themeRegistryMu.RLock()
	defer themeRegistryMu.RUnlock()
	names := make([]string, 0, len(themeRegistry))
	for name := range themeRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func normalizeThemeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
