package color

import (
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestNamedColorCatalogSizesMatchMatplotlib(t *testing.T) {
	if got, want := len(BaseColors()), 8; got != want {
		t.Fatalf("BaseColors length = %d, want %d", got, want)
	}
	if got, want := len(TableauColors()), 10; got != want {
		t.Fatalf("TableauColors length = %d, want %d", got, want)
	}
	if got, want := len(CSS4Colors()), 148; got != want {
		t.Fatalf("CSS4Colors length = %d, want %d", got, want)
	}
	if got, want := len(XKCDColors()), 949; got != want {
		t.Fatalf("XKCDColors length = %d, want %d", got, want)
	}
}

func TestNamedColorInventoryMatchesMatplotlibTables(t *testing.T) {
	upstream := loadUpstreamNamedColorTables(t)

	assertColorMapEqual(t, "base", BaseColors(), upstream.base)
	assertColorMapEqual(t, "tableau", TableauColors(), upstream.tableau)
	assertColorMapEqual(t, "css4", CSS4Colors(), upstream.css4)
	assertColorMapEqual(t, "xkcd", XKCDColors(), upstream.xkcd)
	assertColorMapEqual(t, "full", NamedColors(), upstream.full)
}

func TestToRGBAResolvesMatplotlibNamedColors(t *testing.T) {
	tests := []struct {
		name string
		want render.Color
	}{
		{"b", render.Color{R: 0, G: 0, B: 1, A: 1}},
		{"rebeccapurple", render.Color{R: 0x66 / 255.0, G: 0x33 / 255.0, B: 0x99 / 255.0, A: 1}},
		{"tab:orange", Tab10[1]},
		{"tab:grey", Tab10[7]},
		{"xkcd:cloudy blue", render.Color{R: 0xac / 255.0, G: 0xc2 / 255.0, B: 0xd9 / 255.0, A: 1}},
		{"xkcd:warm gray", render.Color{R: 0x97 / 255.0, G: 0x8a / 255.0, B: 0x84 / 255.0, A: 1}},
	}
	for _, tc := range tests {
		got, err := ToRGBA(tc.name)
		if err != nil {
			t.Fatalf("ToRGBA(%q) error = %v", tc.name, err)
		}
		if !sameColor(got, tc.want) {
			t.Fatalf("ToRGBA(%q) = %+v, want %+v", tc.name, got, tc.want)
		}
	}
}

func TestToRGBAParsesHexGrayCycleAndTuples(t *testing.T) {
	palette := Palette{
		{R: 0.1, G: 0.2, B: 0.3, A: 1},
		{R: 0.4, G: 0.5, B: 0.6, A: 0.7},
	}
	tests := []struct {
		name string
		spec any
		opts []ToRGBAOption
		want render.Color
	}{
		{"long hex", "#336699", nil, render.Color{R: 0x33 / 255.0, G: 0x66 / 255.0, B: 0x99 / 255.0, A: 1}},
		{"short hex alpha", "#369c", nil, render.Color{R: 0x33 / 255.0, G: 0x66 / 255.0, B: 0x99 / 255.0, A: 0xcc / 255.0}},
		{"bare hex opt in", "336699", []ToRGBAOption{WithBareHex()}, render.Color{R: 0x33 / 255.0, G: 0x66 / 255.0, B: 0x99 / 255.0, A: 1}},
		{"gray string", "0.25", nil, render.Color{R: 0.25, G: 0.25, B: 0.25, A: 1}},
		{"cycle", "C3", []ToRGBAOption{WithColorCycle(palette)}, palette[1]},
		{"tuple string", "(0.2, 0.3, 0.4, 0.5)", nil, render.Color{R: 0.2, G: 0.3, B: 0.4, A: 0.5}},
		{"float slice", []float64{0.2, 0.3, 0.4}, []ToRGBAOption{WithAlpha(0.6)}, render.Color{R: 0.2, G: 0.3, B: 0.4, A: 0.6}},
		{"float array", [4]float64{0.2, 0.3, 0.4, 0.5}, nil, render.Color{R: 0.2, G: 0.3, B: 0.4, A: 0.5}},
	}
	for _, tc := range tests {
		got, err := ToRGBA(tc.spec, tc.opts...)
		if err != nil {
			t.Fatalf("%s: ToRGBA(%v) error = %v", tc.name, tc.spec, err)
		}
		if !sameColor(got, tc.want) {
			t.Fatalf("%s: ToRGBA(%v) = %+v, want %+v", tc.name, tc.spec, got, tc.want)
		}
	}
}

func TestToRGBARejectsInvalidSpecs(t *testing.T) {
	for _, spec := range []any{"B", "#12", "1.5", []float64{1, 0, 0, 0, 1}, []float64{1.2, 0, 0}} {
		if got, err := ToRGBA(spec); err == nil {
			t.Fatalf("ToRGBA(%v) = %+v, want error", spec, got)
		}
	}
	if got, err := ToRGBA("red", WithAlpha(2)); err == nil {
		t.Fatalf("ToRGBA(red, WithAlpha(2)) = %+v, want error", got)
	}
}

func sameColor(a, b render.Color) bool {
	const eps = 1e-12
	return math.Abs(a.R-b.R) < eps &&
		math.Abs(a.G-b.G) < eps &&
		math.Abs(a.B-b.B) < eps &&
		math.Abs(a.A-b.A) < eps
}

type upstreamNamedColorTables struct {
	base    map[string]render.Color
	tableau map[string]render.Color
	css4    map[string]render.Color
	xkcd    map[string]render.Color
	full    map[string]render.Color
}

func loadUpstreamNamedColorTables(t *testing.T) upstreamNamedColorTables {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "_color_data.py"))
	if err != nil {
		t.Fatalf("read upstream color data: %v", err)
	}
	src := string(data)
	tables := upstreamNamedColorTables{
		base:    parseUpstreamColorDict(t, src, "BASE_COLORS", false),
		tableau: parseUpstreamColorDict(t, src, "TABLEAU_COLORS", false),
		css4:    parseUpstreamColorDict(t, src, "CSS4_COLORS", false),
		xkcd:    parseUpstreamColorDict(t, src, "XKCD_COLORS", true),
	}
	tables.full = buildUpstreamFullNamedColors(tables)
	return tables
}

func parseUpstreamColorDict(t *testing.T, src, name string, prefixXKCD bool) map[string]render.Color {
	t.Helper()
	block := extractUpstreamDictBlock(t, src, name)
	out := map[string]render.Color{}
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = stripUpstreamLineComment(line)
		line = strings.TrimSuffix(line, ",")
		if line == "" {
			continue
		}
		colorName, value := splitUpstreamDictRow(t, name, line)
		if prefixXKCD {
			colorName = "xkcd:" + colorName
		}
		out[colorName] = parseUpstreamColorValue(t, name, colorName, value)
	}
	return out
}

func stripUpstreamLineComment(line string) string {
	var quote byte
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '\'', '"':
			if i == 0 || line[i-1] != '\\' {
				if quote == 0 {
					quote = line[i]
				} else if quote == line[i] {
					quote = 0
				}
			}
		case '#':
			if quote == 0 {
				return strings.TrimSpace(line[:i])
			}
		}
	}
	return line
}

func splitUpstreamDictRow(t *testing.T, table, line string) (string, string) {
	t.Helper()
	line = strings.TrimSpace(line)
	if line == "" || line[0] != '\'' && line[0] != '"' {
		t.Fatalf("%s: malformed upstream row %q", table, line)
	}
	quote := line[0]
	end := -1
	for i := 1; i < len(line); i++ {
		if line[i] == quote && line[i-1] != '\\' {
			end = i
			break
		}
	}
	if end < 0 {
		t.Fatalf("%s: unterminated upstream key in %q", table, line)
	}
	rest := strings.TrimSpace(line[end+1:])
	if !strings.HasPrefix(rest, ":") {
		t.Fatalf("%s: missing ':' after upstream key in %q", table, line)
	}
	return line[1:end], strings.TrimSpace(rest[1:])
}

func extractUpstreamDictBlock(t *testing.T, src, name string) string {
	t.Helper()
	startMarker := name + " = {"
	start := strings.Index(src, startMarker)
	if start < 0 {
		t.Fatalf("upstream color data missing %s", name)
	}
	bodyStart := start + len(startMarker)
	depth := 1
	var quote byte
	for i := bodyStart; i < len(src); i++ {
		switch src[i] {
		case '\'', '"':
			if i == 0 || src[i-1] != '\\' {
				if quote == 0 {
					quote = src[i]
				} else if quote == src[i] {
					quote = 0
				}
			}
		case '{':
			if quote == 0 {
				depth++
			}
		case '}':
			if quote == 0 {
				depth--
				if depth == 0 {
					return src[bodyStart:i]
				}
			}
		}
	}
	t.Fatalf("upstream color data has unterminated %s", name)
	return ""
}

func parseUpstreamColorValue(t *testing.T, table, name, value string) render.Color {
	t.Helper()
	if strings.HasPrefix(value, "'") || strings.HasPrefix(value, `"`) {
		hex := strings.Trim(value, `'"`)
		col, err := parseHexColor(hex, toRGBAConfig{})
		if err != nil {
			t.Fatalf("%s[%q]: parse %q: %v", table, name, hex, err)
		}
		return col
	}
	if !strings.HasPrefix(value, "(") || !strings.HasSuffix(value, ")") {
		t.Fatalf("%s[%q]: unsupported upstream color value %q", table, name, value)
	}
	parts := strings.Split(strings.Trim(value, "()"), ",")
	if len(parts) != 3 {
		t.Fatalf("%s[%q]: tuple length = %d, want 3", table, name, len(parts))
	}
	var rgb [3]float64
	for i, part := range parts {
		parsed, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			t.Fatalf("%s[%q]: parse tuple component %q: %v", table, name, part, err)
		}
		rgb[i] = parsed
	}
	return render.Color{R: rgb[0], G: rgb[1], B: rgb[2], A: 1}
}

func buildUpstreamFullNamedColors(tables upstreamNamedColorTables) map[string]render.Color {
	out := map[string]render.Color{}
	for name, col := range tables.xkcd {
		out[name] = col
		if strings.Contains(name, "grey") {
			out[strings.ReplaceAll(name, "grey", "gray")] = col
		}
	}
	for name, col := range tables.css4 {
		out[name] = col
	}
	for name, col := range tables.tableau {
		out[name] = col
		if strings.Contains(name, "gray") {
			out[strings.ReplaceAll(name, "gray", "grey")] = col
		}
	}
	for name, col := range tables.base {
		out[name] = col
	}
	return out
}

func assertColorMapEqual(t *testing.T, label string, got, want map[string]render.Color) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s color table length = %d, want %d; missing=%v extra=%v",
			label, len(got), len(want), missingKeys(want, got), missingKeys(got, want))
	}
	for name, wantColor := range want {
		gotColor, ok := got[name]
		if !ok {
			t.Fatalf("%s color table missing %q", label, name)
		}
		if !sameColor(gotColor, wantColor) {
			t.Fatalf("%s[%q] = %+v, want %+v", label, name, gotColor, wantColor)
		}
	}
}

func missingKeys(want, got map[string]render.Color) []string {
	var missing []string
	for key := range want {
		if _, ok := got[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 12 {
		missing = append(missing[:12], "...")
	}
	return missing
}
