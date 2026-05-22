package color

import (
	"fmt"
	stdcolor "image/color"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/cwbudde/matplotlib-go/render"
)

// ToRGBAOption configures ToRGBA parsing.
type ToRGBAOption func(*toRGBAConfig)

type toRGBAConfig struct {
	alpha        *float64
	palette      Palette
	allowBareHex bool
}

// WithAlpha forces the returned alpha channel. The special color "none"
// remains fully transparent, matching Matplotlib.
func WithAlpha(alpha float64) ToRGBAOption {
	return func(cfg *toRGBAConfig) {
		cfg.alpha = &alpha
	}
}

// WithColorCycle sets the palette used to resolve Matplotlib C0, C1, ...
// color-cycle references. Empty palettes fall back to Tab10.
func WithColorCycle(palette Palette) ToRGBAOption {
	return func(cfg *toRGBAConfig) {
		cfg.palette = append(Palette(nil), palette...)
	}
}

// WithBareHex accepts rrggbb / rgb style strings without a leading '#'. This
// is a compatibility extension for older .mplstyle files; Matplotlib's public
// to_rgba requires the leading '#'.
func WithBareHex() ToRGBAOption {
	return func(cfg *toRGBAConfig) {
		cfg.allowBareHex = true
	}
}

// ToRGBA converts a Matplotlib-style color specification to a normalized sRGBA
// tuple. Supported inputs include named colors, #rgb/#rgba/#rrggbb/#rrggbbaa
// hex strings, grayscale strings in [0,1], Cn color-cycle references,
// render.Color values, Go color.Color values, and numeric RGB/RGBA slices or
// arrays.
func ToRGBA(c any, opts ...ToRGBAOption) (render.Color, error) {
	cfg := toRGBAConfig{palette: Tab10}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.alpha != nil && (*cfg.alpha < 0 || *cfg.alpha > 1 || math.IsNaN(*cfg.alpha)) {
		return render.Color{}, fmt.Errorf("alpha must be between 0 and 1, got %g", *cfg.alpha)
	}

	switch v := c.(type) {
	case nil:
		return render.Color{}, fmt.Errorf("invalid RGBA argument: <nil>")
	case string:
		return parseColorString(v, cfg)
	case render.Color:
		return applyAlpha(validateColor(v), cfg)
	case stdcolor.Color:
		r, g, b, a := v.RGBA()
		return applyAlpha(render.Color{
			R: float64(r) / 65535.0,
			G: float64(g) / 65535.0,
			B: float64(b) / 65535.0,
			A: float64(a) / 65535.0,
		}, cfg)
	default:
		if col, ok, err := numericColorSequence(c, cfg); ok || err != nil {
			return col, err
		}
		return render.Color{}, fmt.Errorf("invalid RGBA argument: %T", c)
	}
}

// NamedColors returns Matplotlib's full named-color mapping, including xkcd:,
// CSS4/X11, tab:, and base single-letter colors.
func NamedColors() map[string]render.Color {
	return cloneColorMap(namedColors)
}

// BaseColors returns Matplotlib's base single-letter color mapping.
func BaseColors() map[string]render.Color {
	return cloneColorMap(baseColors)
}

// TableauColors returns Matplotlib's tab: color mapping.
func TableauColors() map[string]render.Color {
	return cloneColorMap(tableauColors)
}

// CSS4Colors returns Matplotlib's CSS4/X11 named-color mapping.
func CSS4Colors() map[string]render.Color {
	return cloneColorMap(css4Colors)
}

// XKCDColors returns Matplotlib's xkcd: color survey mapping.
func XKCDColors() map[string]render.Color {
	return cloneColorMap(xkcdColors)
}

func parseColorString(value string, cfg toRGBAConfig) (render.Color, error) {
	orig := value
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "none") {
		return render.Color{A: 0}, nil
	}
	if value == "" {
		return render.Color{}, fmt.Errorf("empty color")
	}
	if isNthColor(value) {
		palette := cfg.palette
		if len(palette) == 0 {
			palette = Tab10
		}
		idx, _ := strconv.Atoi(value[1:])
		return applyAlpha(validateColor(palette[idx%len(palette)]), cfg)
	}
	if strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")") {
		return parseTupleString(value, cfg)
	}
	if col, ok := lookupNamedColor(value); ok {
		return applyAlpha(col, cfg)
	}
	if strings.HasPrefix(value, "#") || cfg.allowBareHex && looksLikeBareHex(value) {
		return parseHexColor(value, cfg)
	}
	if grayscale, err := strconv.ParseFloat(value, 64); err == nil {
		if grayscale < 0 || grayscale > 1 || math.IsNaN(grayscale) {
			return render.Color{}, fmt.Errorf("invalid string grayscale value %q: value must be within 0-1 range", orig)
		}
		return applyAlpha(render.Color{R: grayscale, G: grayscale, B: grayscale, A: 1}, cfg)
	}
	return render.Color{}, fmt.Errorf("invalid RGBA argument: %q", orig)
}

func isNthColor(value string) bool {
	if len(value) < 2 || value[0] != 'C' {
		return false
	}
	for _, r := range value[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func lookupNamedColor(value string) (render.Color, bool) {
	if col, ok := namedColors[value]; ok {
		return col, true
	}
	if len(value) != 1 {
		col, ok := namedColors[strings.ToLower(value)]
		return col, ok
	}
	return render.Color{}, false
}

func parseTupleString(value string, cfg toRGBAConfig) (render.Color, error) {
	parts := strings.Split(strings.TrimSpace(value[1:len(value)-1]), ",")
	if len(parts) != 3 && len(parts) != 4 {
		return render.Color{}, fmt.Errorf("RGBA sequence should have length 3 or 4")
	}
	values := make([]float64, len(parts))
	for i, part := range parts {
		parsed, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return render.Color{}, fmt.Errorf("invalid tuple component %q", part)
		}
		values[i] = parsed
	}
	return rgbaFromFloats(values, cfg)
}

func parseHexColor(value string, cfg toRGBAConfig) (render.Color, error) {
	normalized := strings.TrimPrefix(value, "#")
	switch len(normalized) {
	case 3:
		normalized = strings.Repeat(string(normalized[0]), 2) +
			strings.Repeat(string(normalized[1]), 2) +
			strings.Repeat(string(normalized[2]), 2)
	case 4:
		normalized = strings.Repeat(string(normalized[0]), 2) +
			strings.Repeat(string(normalized[1]), 2) +
			strings.Repeat(string(normalized[2]), 2) +
			strings.Repeat(string(normalized[3]), 2)
	case 6, 8:
	default:
		return render.Color{}, fmt.Errorf("invalid hex color specifier: %q", value)
	}
	for _, r := range normalized {
		if !isHexDigit(r) {
			return render.Color{}, fmt.Errorf("invalid hex color specifier: %q", value)
		}
	}
	parseByte := func(part string) float64 {
		n, _ := strconv.ParseUint(part, 16, 8)
		return float64(n) / 255.0
	}
	col := render.Color{
		R: parseByte(normalized[0:2]),
		G: parseByte(normalized[2:4]),
		B: parseByte(normalized[4:6]),
		A: 1,
	}
	if len(normalized) == 8 {
		col.A = parseByte(normalized[6:8])
	}
	return applyAlpha(col, cfg)
}

func looksLikeBareHex(value string) bool {
	switch len(value) {
	case 3, 4, 6, 8:
	default:
		return false
	}
	for _, r := range value {
		if !isHexDigit(r) {
			return false
		}
	}
	return true
}

func isHexDigit(r rune) bool {
	return r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F'
}

func numericColorSequence(c any, cfg toRGBAConfig) (render.Color, bool, error) {
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return render.Color{}, false, nil
	}
	if v.Len() != 3 && v.Len() != 4 {
		return render.Color{}, true, fmt.Errorf("RGBA sequence should have length 3 or 4")
	}
	values := make([]float64, v.Len())
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}
		switch elem.Kind() {
		case reflect.Float32, reflect.Float64:
			values[i] = elem.Convert(reflect.TypeOf(float64(0))).Float()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			values[i] = float64(elem.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			values[i] = float64(elem.Uint())
		default:
			return render.Color{}, true, fmt.Errorf("invalid RGBA argument: %T", c)
		}
	}
	col, err := rgbaFromFloats(values, cfg)
	return col, true, err
}

func rgbaFromFloats(values []float64, cfg toRGBAConfig) (render.Color, error) {
	col := render.Color{R: values[0], G: values[1], B: values[2], A: 1}
	if len(values) == 4 {
		col.A = values[3]
	}
	return applyAlpha(validateColor(col), cfg)
}

func validateColor(col render.Color) render.Color {
	return col
}

func applyAlpha(col render.Color, cfg toRGBAConfig) (render.Color, error) {
	if cfg.alpha != nil {
		col.A = *cfg.alpha
	}
	if col.R < 0 || col.R > 1 || col.G < 0 || col.G > 1 || col.B < 0 || col.B > 1 || col.A < 0 || col.A > 1 ||
		math.IsNaN(col.R) || math.IsNaN(col.G) || math.IsNaN(col.B) || math.IsNaN(col.A) {
		return render.Color{}, fmt.Errorf("RGBA values should be within 0-1 range")
	}
	return col, nil
}

func cloneColorMap(src map[string]render.Color) map[string]render.Color {
	out := make(map[string]render.Color, len(src))
	for name, col := range src {
		out[name] = col
	}
	return out
}

var (
	baseColors    = parseNamedColorData(baseColorHexData)
	tableauColors = parseNamedColorData(tableauColorHexData)
	css4Colors    = parseNamedColorData(css4ColorHexData)
	xkcdColors    = parseNamedColorData(xkcdColorHexData)
	namedColors   = buildNamedColors()
)

func buildNamedColors() map[string]render.Color {
	out := map[string]render.Color{}
	for name, col := range xkcdColors {
		out[name] = col
		if strings.Contains(name, "grey") {
			out[strings.ReplaceAll(name, "grey", "gray")] = col
		}
	}
	for name, col := range css4Colors {
		out[name] = col
	}
	for name, col := range tableauColors {
		out[name] = col
		if strings.Contains(name, "gray") {
			out[strings.ReplaceAll(name, "gray", "grey")] = col
		}
	}
	for name, col := range baseColors {
		out[name] = col
	}
	return out
}

func parseNamedColorData(data string) map[string]render.Color {
	out := map[string]render.Color{}
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, hex, ok := strings.Cut(line, "\t")
		if !ok {
			panic("invalid named color table row: " + line)
		}
		col, err := parseHexColor(hex, toRGBAConfig{})
		if err != nil {
			panic(err)
		}
		out[name] = col
	}
	return out
}
