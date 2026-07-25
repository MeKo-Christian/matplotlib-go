package color

import (
	"sync"

	"github.com/cwbudde/matplotlib-go/render"
)

// Petroff10 is Matplotlib's "petroff10" color sequence: ten colorblind-friendly
// hues from Petroff (2021), "Accessible Color Sequences for Data
// Visualization". The values mirror matplotlib._cm._petroff10_data exactly.
var Petroff10 = Palette{
	{R: 0x3f / 255.0, G: 0x90 / 255.0, B: 0xda / 255.0, A: 1}, // 3f90da
	{R: 0xff / 255.0, G: 0xa9 / 255.0, B: 0x0e / 255.0, A: 1}, // ffa90e
	{R: 0xbd / 255.0, G: 0x1f / 255.0, B: 0x01 / 255.0, A: 1}, // bd1f01
	{R: 0x94 / 255.0, G: 0xa4 / 255.0, B: 0xa2 / 255.0, A: 1}, // 94a4a2
	{R: 0x83 / 255.0, G: 0x2d / 255.0, B: 0xb6 / 255.0, A: 1}, // 832db6
	{R: 0xa9 / 255.0, G: 0x6b / 255.0, B: 0x59 / 255.0, A: 1}, // a96b59
	{R: 0xe7 / 255.0, G: 0x63 / 255.0, B: 0x00 / 255.0, A: 1}, // e76300
	{R: 0xb9 / 255.0, G: 0xac / 255.0, B: 0x70 / 255.0, A: 1}, // b9ac70
	{R: 0x71 / 255.0, G: 0x75 / 255.0, B: 0x81 / 255.0, A: 1}, // 717581
	{R: 0x92 / 255.0, G: 0xda / 255.0, B: 0xdd / 255.0, A: 1}, // 92dadd
}

// colorSequences mirrors Matplotlib's color_sequences registry, which is
// distinct from the colormap registry: a named, ordered list of discrete colors
// suitable for an axes property cycle. Matplotlib registers petroff10 (and the
// qualitative tab/Set families) here without making them gradient colormaps.
var colorSequenceRegistry = struct {
	sync.RWMutex
	sequences map[string]Palette
}{
	sequences: map[string]Palette{
		"petroff10": Petroff10,
		"tab10":     Tab10,
	},
}

// ColorSequence returns the named color sequence and whether it is registered.
// Names match Matplotlib's color_sequences registry (e.g. "petroff10",
// "tab10"). The returned palette is a copy and is safe to mutate.
func ColorSequence(name string) (Palette, bool) {
	colorSequenceRegistry.RLock()
	defer colorSequenceRegistry.RUnlock()
	seq, ok := colorSequenceRegistry.sequences[name]
	if !ok {
		return nil, false
	}
	return append(Palette(nil), seq...), true
}

// RegisterColorSequence adds or replaces a named color sequence, mirroring
// Matplotlib's color_sequences.register. It is safe for concurrent use with
// ColorSequence and ColorSequenceNames.
func RegisterColorSequence(name string, palette Palette) {
	seq := append(Palette(nil), palette...)
	colorSequenceRegistry.Lock()
	colorSequenceRegistry.sequences[name] = seq
	colorSequenceRegistry.Unlock()
}

// ColorSequenceNames returns the registered color-sequence names in no
// particular order.
func ColorSequenceNames() []string {
	colorSequenceRegistry.RLock()
	defer colorSequenceRegistry.RUnlock()
	names := make([]string, 0, len(colorSequenceRegistry.sequences))
	for name := range colorSequenceRegistry.sequences {
		names = append(names, name)
	}
	return names
}

func init() {
	// Also expose petroff10 through the colormap registry as a ListedColormap so
	// LookupColormap("petroff10") and colorbar/swatch paths can resolve it. This is
	// additive and intentionally kept out of matplotlibListedColormapNames, which
	// mirrors only the upstream gradient/colormap catalog (petroff10 is a color
	// sequence upstream, not a registered colormap).
	colors := make([]render.Color, len(Petroff10))
	copy(colors, Petroff10)
	RegisterColormap("petroff10", NewListedColormap("petroff10", colors))
}
