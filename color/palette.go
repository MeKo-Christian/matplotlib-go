package color

import "github.com/cwbudde/matplotlib-go/render"

// Palette defines a set of colors for automatic cycling.
type Palette []render.Color

// Tab10 is the default matplotlib tab10 color palette.
var Tab10 = Palette{
	{R: 31.0 / 255.0, G: 119.0 / 255.0, B: 180.0 / 255.0, A: 1},  // blue
	{R: 255.0 / 255.0, G: 127.0 / 255.0, B: 14.0 / 255.0, A: 1},  // orange
	{R: 44.0 / 255.0, G: 160.0 / 255.0, B: 44.0 / 255.0, A: 1},   // green
	{R: 214.0 / 255.0, G: 39.0 / 255.0, B: 40.0 / 255.0, A: 1},   // red
	{R: 148.0 / 255.0, G: 103.0 / 255.0, B: 189.0 / 255.0, A: 1}, // purple
	{R: 140.0 / 255.0, G: 86.0 / 255.0, B: 75.0 / 255.0, A: 1},   // brown
	{R: 227.0 / 255.0, G: 119.0 / 255.0, B: 194.0 / 255.0, A: 1}, // pink
	{R: 127.0 / 255.0, G: 127.0 / 255.0, B: 127.0 / 255.0, A: 1}, // gray
	{R: 188.0 / 255.0, G: 189.0 / 255.0, B: 34.0 / 255.0, A: 1},  // olive
	{R: 23.0 / 255.0, G: 190.0 / 255.0, B: 207.0 / 255.0, A: 1},  // cyan
}

// ColorCycle manages automatic color cycling for plot series.
type ColorCycle struct {
	palette Palette
	index   int
}

// NewColorCycle creates a new color cycle with the given palette.
func NewColorCycle(palette Palette) *ColorCycle {
	if len(palette) == 0 {
		palette = Tab10 // fallback to default
	}
	return &ColorCycle{
		palette: palette,
		index:   0,
	}
}

// NewDefaultColorCycle creates a new color cycle with the default Tab10 palette.
func NewDefaultColorCycle() *ColorCycle {
	return NewColorCycle(Tab10)
}

// Next returns the next color in the cycle and advances the index.
func (c *ColorCycle) Next() render.Color {
	if len(c.palette) == 0 {
		return render.Color{R: 0, G: 0, B: 0, A: 1} // black fallback
	}

	color := c.palette[c.index]
	c.index = (c.index + 1) % len(c.palette)
	return color
}

// Peek returns the current color without advancing the index.
func (c *ColorCycle) Peek() render.Color {
	if len(c.palette) == 0 {
		return render.Color{R: 0, G: 0, B: 0, A: 1} // black fallback
	}

	return c.palette[c.index]
}

// Reset resets the color cycle to the first color.
func (c *ColorCycle) Reset() {
	c.index = 0
}

// Index returns the current index in the color cycle.
func (c *ColorCycle) Index() int {
	return c.index
}

// Length returns the number of colors in the palette.
func (c *ColorCycle) Length() int {
	return len(c.palette)
}

// At returns the color at the given index (modulo palette length).
func (c *ColorCycle) At(index int) render.Color {
	if len(c.palette) == 0 {
		return render.Color{R: 0, G: 0, B: 0, A: 1} // black fallback
	}

	idx := index % len(c.palette)
	if idx < 0 {
		idx += len(c.palette)
	}
	return c.palette[idx]
}
