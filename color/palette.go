package color

import "matplotlib-go/render"

// Palette defines a set of colors for automatic cycling.
type Palette []render.Color

// Tab10 is the default matplotlib tab10 color palette.
var Tab10 = Palette{
	{R: 0.12, G: 0.47, B: 0.71, A: 1}, // blue
	{R: 1.00, G: 0.50, B: 0.05, A: 1}, // orange
	{R: 0.17, G: 0.63, B: 0.17, A: 1}, // green
	{R: 0.84, G: 0.15, B: 0.16, A: 1}, // red
	{R: 0.58, G: 0.40, B: 0.74, A: 1}, // purple
	{R: 0.55, G: 0.34, B: 0.29, A: 1}, // brown
	{R: 0.89, G: 0.47, B: 0.76, A: 1}, // pink
	{R: 0.50, G: 0.50, B: 0.50, A: 1}, // gray
	{R: 0.74, G: 0.74, B: 0.13, A: 1}, // olive
	{R: 0.09, G: 0.75, B: 0.81, A: 1}, // cyan
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