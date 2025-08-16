package style

// RC holds global rendering defaults (rc-like configuration).
// Fields are simple value types to keep configuration immutable-ish by copy.
type RC struct {
	DPI        float64
	FontKey    string
	FontSize   float64
	LineWidth  float64
	TextColor  [4]float64
	LineColor  [4]float64
	Background [4]float64
	TickCountX int
	TickCountY int
}

// Default contains the library defaults. Copy and apply options to customize.
var Default = RC{
	DPI:        96,
	FontKey:    "DejaVuSans",
	FontSize:   12,
	LineWidth:  1.25,
	TextColor:  [4]float64{0, 0, 0, 1},
	LineColor:  [4]float64{0, 0, 0, 1},
	Background: [4]float64{1, 1, 1, 1},
	TickCountX: 5,
	TickCountY: 5,
}

// Option mutates an RC. Options should be applied on a copy derived from Default.
type Option func(*RC)

// Apply copies base and applies the given options in order, returning the result.
func Apply(base RC, opts ...Option) RC {
	rc := base
	for _, opt := range opts {
		if opt != nil {
			opt(&rc)
		}
	}
	return rc
}

// WithDPI sets the DPI.
func WithDPI(d float64) Option { return func(rc *RC) { rc.DPI = d } }

// WithFont sets the font key and size.
func WithFont(key string, size float64) Option {
	return func(rc *RC) { rc.FontKey, rc.FontSize = key, size }
}

// WithLineWidth sets the default line width.
func WithLineWidth(w float64) Option { return func(rc *RC) { rc.LineWidth = w } }

// WithTextColor sets the default text color RGBA (0..1).
func WithTextColor(r, g, b, a float64) Option {
	return func(rc *RC) { rc.TextColor = [4]float64{r, g, b, a} }
}

// WithLineColor sets the default stroke color RGBA (0..1).
func WithLineColor(r, g, b, a float64) Option {
	return func(rc *RC) { rc.LineColor = [4]float64{r, g, b, a} }
}

// WithBackground sets the default background color RGBA (0..1).
func WithBackground(r, g, b, a float64) Option {
	return func(rc *RC) { rc.Background = [4]float64{r, g, b, a} }
}

// WithTickCounts sets the target tick counts for X and Y.
func WithTickCounts(nx, ny int) Option { return func(rc *RC) { rc.TickCountX, rc.TickCountY = nx, ny } }
