package mathtext

import (
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// Metrics is the renderer-neutral text measurement subset needed for MathText
// layout. BoundsY/BoundsH describe the signed ink bounds relative to the
// baseline when the renderer can provide them.
type Metrics struct{ W, H, Ascent, Descent, BoundsY, BoundsH float64 }

// GlyphInfo carries matplotlib `_get_info` metrics for one glyph (device pixels,
// baseline-relative, y-up). Advance is the UNHINTED linearHoriAdvance, Iceberg =
// horiBearingY/64 (TeX height), Height the full ink height, Xmin/Xmax/Ymin/Ymax
// the ink bbox, KernToPrev the kerning to the previous glyph in the run.
type GlyphInfo struct {
	Advance    float64
	Iceberg    float64
	Height     float64
	Xmin       float64
	Xmax       float64
	Ymin       float64
	Ymax       float64
	KernToPrev float64
}

// Measurer measures text for one font key and size.
type Measurer interface {
	MeasureText(text string, size float64, fontKey string) Metrics
}

// DPIMeasurer is an optional Measurer capability exposing the render DPI, so the
// layout can compute matplotlib's exact `fontsize*dpi/72` device pixel size for
// the (font-independent) underline thickness instead of reconstructing it from
// the x-height.
type DPIMeasurer interface {
	DPI() float64
}

// GlyphMeasurer is an optional Measurer capability returning matplotlib's exact
// per-glyph `_get_info` metrics. When a Measurer implements it and GlyphRun
// returns ok, the layout positions glyphs pixel-exactly; otherwise it falls back
// to whole-run MeasureText. (Unavailable on purego/WASM, which lacks FreeType.)
type GlyphMeasurer interface {
	GlyphRun(text string, size float64, fontKey string) ([]GlyphInfo, bool)
}

// FontStyle describes the font posture requested by MathText style commands.
type FontStyle string

const (
	FontStyleNormal FontStyle = "normal"
	FontStyleItalic FontStyle = "italic"
)

// FontRequest describes a MathText font override relative to the current font.
type FontRequest struct {
	Families []string
	Style    FontStyle
	Weight   int
}

// FontResolver resolves MathText font requests into renderer font keys.
type FontResolver interface {
	ResolveMathFontKey(base string, request FontRequest) string
}

// Options configures MathText layout.
type Options struct {
	FontResolver   FontResolver
	Cache          *Cache
	MeasurementKey string
}

// MathTextLayoutRun is one text draw in a laid-out MathText expression.
type MathTextLayoutRun struct {
	Text     string
	Offset   geom.Pt
	FontSize float64
	FontKey  string
}

// MathTextLayoutRule is a filled rule, such as a fraction bar or root vinculum.
type MathTextLayoutRule struct {
	Rect geom.Rect
}

// MathTextLayout is a lightweight layout tree flattened into draw runs and
// rules. Offsets and rectangles are relative to the expression baseline.
type MathTextLayout struct {
	Runs    []MathTextLayoutRun
	Rules   []MathTextLayoutRule
	Width   float64
	Ascent  float64
	Descent float64
	Height  float64
}

type mathLayoutKind uint8

const (
	mathLayoutList mathLayoutKind = iota
	mathLayoutText
	mathLayoutScript
	mathLayoutFrac
	mathLayoutSqrt
	mathLayoutSpace
	mathLayoutStyled
	mathLayoutFence
	mathLayoutMatrix
)

const (
	mathGlyphDelimiterScale = 0.45
	mathRuleDelimiterScale  = 0.76
	mathSpaceScale          = 1.60
	mathCMXHeightScale      = 0.43
)

type mathDelimiterGlyph struct {
	text    string
	fontKey string
}

type mathLayoutNode struct {
	kind     mathLayoutKind
	text     string
	widthEm  float64
	children []mathLayoutNode
	segments []mathLayoutNode
	base     *mathLayoutNode
	super    *mathLayoutNode
	sub      *mathLayoutNode
	num      *mathLayoutNode
	den      *mathLayoutNode
	radicand *mathLayoutNode
	index    *mathLayoutNode
	child    *mathLayoutNode
	left     string
	middles  []string
	right    string
	fracRule bool
	fracDisp bool
	rows     [][]mathLayoutNode
	families []string
	style    FontStyle
	weight   int
}

type mathLayoutParser struct {
	input          []rune
	pos            int
	implicitItalic bool
}

// LayoutMathText parses and lays out one MathText expression without requiring
// dollar delimiters. It supports the same fallback command set as displayed
// text normalization plus baseline-shifted scripts, stacked fractions, and
// square-root vincula.
func LayoutMathText(m Measurer, expr string, size float64, fontKey string, opts Options) (MathTextLayout, bool) {
	if m == nil || strings.TrimSpace(expr) == "" || size <= 0 {
		return MathTextLayout{}, false
	}
	expr = strings.ReplaceAll(expr, `\\`, `\`)
	cacheKey, useLayoutCache := opts.layoutCacheKey("math", expr, size, fontKey)
	if useLayoutCache {
		if layout, ok := opts.Cache.layout(cacheKey); ok {
			return layout, true
		}
	}
	node := parseMathLayoutNode(expr, opts.Cache)
	box := layoutMathNode(m, node, size, fontKey, opts)
	if box.Width <= 0 && len(box.runs) == 0 && len(box.rules) == 0 {
		return MathTextLayout{}, false
	}
	layout := MathTextLayout{
		Runs:    box.runs,
		Rules:   box.rules,
		Width:   box.Width,
		Ascent:  box.Ascent,
		Descent: box.Descent,
		Height:  box.Ascent + box.Descent,
	}
	if useLayoutCache {
		opts.Cache.storeLayout(cacheKey, layout)
	}
	return cloneLayout(layout), true
}

// LayoutDisplay lays out display text with either a single full $...$
// expression or mixed plain text and inline $...$ MathText segments.
func LayoutDisplay(m Measurer, text string, size float64, fontKey string, opts Options) (MathTextLayout, bool) {
	if expr, ok := FullExpression(text); ok {
		return LayoutMathText(m, expr, size, fontKey, opts)
	}

	cacheKey, useLayoutCache := opts.layoutCacheKey("display", text, size, fontKey)
	if useLayoutCache {
		if layout, ok := opts.Cache.layout(cacheKey); ok {
			return layout, true
		}
	}

	segments, hasMath, ok := SplitDisplaySegments(text)
	if !ok || !hasMath {
		return MathTextLayout{}, false
	}

	var out mathLayoutBox
	x := 0.0
	for _, segment := range segments {
		var child mathLayoutBox
		if segment.IsMath {
			layout, ok := LayoutMathText(m, segment.Text, size, fontKey, opts)
			if !ok {
				return MathTextLayout{}, false
			}
			child = mathLayoutBox{
				runs:    append([]MathTextLayoutRun(nil), layout.Runs...),
				rules:   append([]MathTextLayoutRule(nil), layout.Rules...),
				Width:   layout.Width,
				Ascent:  layout.Ascent,
				Descent: layout.Descent,
			}
		} else {
			child = layoutMathTextRun(m, displayTextCommandReplacer.Replace(segment.Text), size, fontKey)
		}
		if child.Width <= 0 && len(child.runs) == 0 && len(child.rules) == 0 {
			continue
		}
		out.appendTranslated(child, x, 0)
		x += child.Width
		out.Ascent = maxFloat64(out.Ascent, child.Ascent)
		out.Descent = maxFloat64(out.Descent, child.Descent)
	}

	if x <= 0 && len(out.runs) == 0 && len(out.rules) == 0 {
		return MathTextLayout{}, false
	}

	layout := MathTextLayout{
		Runs:    out.runs,
		Rules:   out.rules,
		Width:   x,
		Ascent:  out.Ascent,
		Descent: out.Descent,
		Height:  out.Ascent + out.Descent,
	}
	if useLayoutCache {
		opts.Cache.storeLayout(cacheKey, layout)
	}
	return cloneLayout(layout), true
}

func (o Options) layoutCacheKey(kind, text string, size float64, fontKey string) (layoutCacheKey, bool) {
	if o.Cache == nil || strings.TrimSpace(o.MeasurementKey) == "" {
		return layoutCacheKey{}, false
	}
	return layoutCacheKey{
		kind:           kind,
		text:           text,
		size:           size,
		fontKey:        fontKey,
		measurementKey: o.MeasurementKey,
	}, true
}

func parseMathLayoutNode(expr string, cache *Cache) mathLayoutNode {
	if cache != nil {
		if node, ok := cache.parsedNode(expr); ok {
			return node
		}
	}
	parser := mathLayoutParser{input: []rune(expr), implicitItalic: true}
	node := parser.parseUntil(0)
	if cache != nil {
		cache.storeParsedNode(expr, node)
	}
	return node
}

func (p *mathLayoutParser) parseUntil(stop rune) mathLayoutNode {
	var children []mathLayoutNode
	appendText := func(text string) {
		children = appendMathText(children, text)
	}

	for p.pos < len(p.input) {
		r := p.input[p.pos]
		if stop != 0 && r == stop {
			break
		}
		switch r {
		case '{':
			p.pos++
			children = append(children, p.parseUntil('}').children...)
			if p.pos < len(p.input) && p.input[p.pos] == '}' {
				p.pos++
			}
		case '}':
			if stop == 0 {
				p.pos++
				continue
			}
			return mathLayoutNode{kind: mathLayoutList, children: children}
		case '^', '_':
			p.pos++
			children = attachMathScript(children, r, p.parseArgumentNode())
		case '\\':
			node := p.parseCommandNode()
			if node.kind == mathLayoutText {
				appendText(node.text)
			} else if !node.isEmpty() {
				children = append(children, node)
			}
		case '~':
			appendText(" ")
			p.pos++
		case '+', '-', '=':
			children = p.appendMathOperator(children, r, stop)
		case ',', ';', '.', '!':
			children = p.appendMathPunctuation(children, r, stop)
		case ' ', '\t', '\n', '\r':
			if !p.implicitItalic {
				appendText(" ")
			}
			p.pos++
		default:
			children = appendMathAtom(children, r, p.implicitItalic)
			p.pos++
		}
	}
	return mathLayoutNode{kind: mathLayoutList, children: children}
}

func (p *mathLayoutParser) parseArgumentNode() mathLayoutNode {
	p.skipSpace()
	if p.pos >= len(p.input) {
		return mathLayoutNode{kind: mathLayoutText}
	}
	switch p.input[p.pos] {
	case '{':
		p.pos++
		node := p.parseUntil('}')
		if p.pos < len(p.input) && p.input[p.pos] == '}' {
			p.pos++
		}
		return node
	case '\\':
		return p.parseCommandNode()
	default:
		r := p.input[p.pos]
		p.pos++
		return mathAtomNode(r, p.implicitItalic)
	}
}

func (p *mathLayoutParser) parseCommandNode() mathLayoutNode {
	p.pos++
	if p.pos >= len(p.input) {
		return mathLayoutNode{kind: mathLayoutText, text: `\`}
	}
	r := p.input[p.pos]
	if !unicode.IsLetter(r) {
		p.pos++
		switch r {
		case ',':
			return mathLayoutNode{kind: mathLayoutSpace, widthEm: 0.166}
		case ':':
			return mathLayoutNode{kind: mathLayoutSpace, widthEm: 0.222}
		case ';':
			return mathLayoutNode{kind: mathLayoutSpace, widthEm: 0.278}
		case ' ':
			return mathLayoutNode{kind: mathLayoutSpace, widthEm: 0.333}
		case '!':
			return mathLayoutNode{kind: mathLayoutSpace, widthEm: -0.166}
		default:
			return mathLayoutNode{kind: mathLayoutText, text: string(r)}
		}
	}

	start := p.pos
	for p.pos < len(p.input) && unicode.IsLetter(p.input[p.pos]) {
		p.pos++
	}
	name := string(p.input[start:p.pos])

	if mapped, ok := mathTextCommandMap[name]; ok {
		return mathLayoutNode{kind: mathLayoutText, text: mapped}
	}
	if width, ok := mathTextSpacingCommandWidths[name]; ok {
		return mathLayoutNode{kind: mathLayoutSpace, widthEm: width}
	}
	if delim, ok := mathTextDelimiterCommands[name]; ok {
		return mathLayoutNode{kind: mathLayoutText, text: delim}
	}
	if op, ok := mathTextOperatorMap[name]; ok {
		return mathLayoutNode{kind: mathLayoutText, text: op}
	}
	if _, ok := mathTextPassthroughCommands[name]; ok {
		return p.parseStyledArgumentNode(name)
	}
	if mark, ok := mathTextAccentMarks[name]; ok {
		return mathLayoutNode{kind: mathLayoutText, text: applyMathAccent(nodePlainText(p.parseArgumentNode()), mark)}
	}
	if name == "begin" {
		return p.parseEnvironmentNode()
	}
	if name == "left" {
		return p.parseFenceNode()
	}
	if _, ok := mathTextEmptyCommands[name]; ok {
		p.skipSpace()
		return mathLayoutNode{kind: mathLayoutText}
	}

	switch name {
	case "frac":
		num := p.parseArgumentNode()
		den := p.parseArgumentNode()
		return mathLayoutNode{kind: mathLayoutFrac, num: &num, den: &den, fracRule: true}
	case "dfrac":
		num := p.parseArgumentNode()
		den := p.parseArgumentNode()
		return mathLayoutNode{kind: mathLayoutFrac, num: &num, den: &den, fracRule: true, fracDisp: true}
	case "binom":
		num := p.parseArgumentNode()
		den := p.parseArgumentNode()
		return mathLayoutNode{kind: mathLayoutFrac, num: &num, den: &den, left: "(", right: ")"}
	case "genfrac":
		left := parseMathDelimiterSpec(p.parseBraceText())
		right := parseMathDelimiterSpec(p.parseBraceText())
		rule := parseMathSpaceDimension(p.parseBraceText()) > 0
		display := strings.TrimSpace(p.parseBraceText()) == "0"
		num := p.parseArgumentNode()
		den := p.parseArgumentNode()
		return mathLayoutNode{kind: mathLayoutFrac, num: &num, den: &den, left: left, right: right, fracRule: rule, fracDisp: display}
	case "hspace", "kern":
		return mathLayoutNode{kind: mathLayoutSpace, widthEm: parseMathSpaceDimension(p.parseBraceText())}
	case "sqrt":
		var index *mathLayoutNode
		if p.pos < len(p.input) && p.input[p.pos] == '[' {
			p.pos++
			idx := p.parseUntil(']')
			if p.pos < len(p.input) && p.input[p.pos] == ']' {
				p.pos++
			}
			index = &idx
		}
		radicand := p.parseArgumentNode()
		return mathLayoutNode{kind: mathLayoutSqrt, radicand: &radicand, index: index}
	default:
		return mathLayoutNode{kind: mathLayoutText, text: `\` + name}
	}
}

func (p *mathLayoutParser) parseEnvironmentNode() mathLayoutNode {
	name := p.parseBraceText()
	left, right, ok := matrixEnvironmentDelimiters(name)
	if !ok {
		return mathLayoutNode{kind: mathLayoutText}
	}
	if name == "array" && p.pos < len(p.input) && p.input[p.pos] == '{' {
		_ = p.parseBraceText()
	}
	rows := p.parseMatrixRows(name)
	return mathLayoutNode{kind: mathLayoutMatrix, rows: rows, left: left, right: right}
}

func (p *mathLayoutParser) parseStyledArgumentNode(name string) mathLayoutNode {
	implicitItalic := p.implicitItalic
	p.implicitItalic = false
	arg := p.parseArgumentNode()
	p.implicitItalic = implicitItalic
	node := mathLayoutNode{kind: mathLayoutStyled, child: &arg, style: FontStyleNormal, weight: 400}
	switch name {
	case "mathsf":
		node.families = []string{"DejaVu Sans", "sans-serif"}
	case "mathtt":
		node.families = []string{"DejaVu Sans Mono", "monospace"}
	case "mathrm", "text":
		node.families = []string{"DejaVu Serif", "serif"}
	case "mathit":
		node.style = FontStyleItalic
	case "mathbf":
		node.weight = 700
	case "operatorname":
		// Preserve current face selection but normalize posture/weight.
	default:
		return arg
	}
	return node
}

func appendMathText(children []mathLayoutNode, text string) []mathLayoutNode {
	if text == "" {
		return children
	}
	n := len(children)
	if n > 0 && children[n-1].kind == mathLayoutText {
		children[n-1].text += text
		return children
	}
	return append(children, mathLayoutNode{kind: mathLayoutText, text: text})
}

func appendMathAtom(children []mathLayoutNode, r rune, implicitItalic bool) []mathLayoutNode {
	node := mathAtomNode(r, implicitItalic)
	if node.kind == mathLayoutText {
		return appendMathText(children, node.text)
	}
	return append(children, node)
}

func mathAtomNode(r rune, implicitItalic bool) mathLayoutNode {
	if implicitItalic && isMathItalicLatin(r) {
		text := mathLayoutNode{kind: mathLayoutText, text: string(r)}
		return mathLayoutNode{kind: mathLayoutStyled, child: &text, style: FontStyleItalic}
	}
	return mathLayoutNode{kind: mathLayoutText, text: string(r)}
}

func isMathItalicLatin(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

func mathDelimiterFontKey() string {
	return "DejaVu Serif"
}

func (p *mathLayoutParser) appendMathOperator(children []mathLayoutNode, op rune, stop rune) []mathLayoutNode {
	text := string(op)
	if op == '-' {
		text = "−"
	}
	if p.implicitItalic && p.hasPreviousMathOperand(children) && p.hasNextMathOperand(stop) {
		children = append(children, mathLayoutNode{kind: mathLayoutSpace, widthEm: 0.2})
		children = appendMathText(children, text)
		children = append(children, mathLayoutNode{kind: mathLayoutSpace, widthEm: 0.2})
	} else {
		children = appendMathText(children, text)
	}
	p.pos++
	return children
}

func (p *mathLayoutParser) appendMathPunctuation(children []mathLayoutNode, punct rune, stop rune) []mathLayoutNode {
	text := string(punct)
	if punct == '.' && p.previousNonSpaceIsDigit() && p.nextNonSpaceIsDigit(stop) {
		children = appendMathText(children, text)
	} else {
		children = appendMathText(children, text)
		children = append(children, mathLayoutNode{kind: mathLayoutSpace, widthEm: 0.2})
	}
	p.pos++
	return children
}

func (p *mathLayoutParser) previousNonSpaceIsDigit() bool {
	for i := p.pos - 1; i >= 0; i-- {
		r := p.input[i]
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			continue
		}
		return unicode.IsDigit(r)
	}
	return false
}

func (p *mathLayoutParser) nextNonSpaceIsDigit(stop rune) bool {
	for i := p.pos + 1; i < len(p.input); i++ {
		r := p.input[i]
		if stop != 0 && r == stop {
			return false
		}
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			continue
		}
		return unicode.IsDigit(r)
	}
	return false
}

func (p *mathLayoutParser) hasPreviousMathOperand(children []mathLayoutNode) bool {
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		switch child.kind {
		case mathLayoutSpace:
			continue
		case mathLayoutText:
			return child.text != "" && child.text != "+" && child.text != "−" && child.text != "="
		default:
			return true
		}
	}
	return false
}

func (p *mathLayoutParser) hasNextMathOperand(stop rune) bool {
	for i := p.pos + 1; i < len(p.input); i++ {
		r := p.input[i]
		if stop != 0 && r == stop {
			return false
		}
		switch r {
		case ' ', '\t', '\n', '\r':
			continue
		case '+', '-', '=', '^', '_', '}':
			return false
		default:
			return true
		}
	}
	return false
}

func (p *mathLayoutParser) parseFenceNode() mathLayoutNode {
	left := p.parseDelimiterToken()
	segments := []mathLayoutNode{p.parseUntilFenceBoundary()}
	middles := []string{}
	for p.consumeNamedCommand("middle") {
		middles = append(middles, p.parseDelimiterToken())
		segments = append(segments, p.parseUntilFenceBoundary())
	}
	right := ""
	if p.consumeNamedCommand("right") {
		right = p.parseDelimiterToken()
	}
	node := mathLayoutNode{kind: mathLayoutFence, left: left, middles: middles, right: right, segments: segments}
	if len(segments) == 1 {
		node.child = &segments[0]
	}
	return node
}

func (p *mathLayoutParser) parseMatrixRows(envName string) [][]mathLayoutNode {
	rows := [][]mathLayoutNode{}
	for {
		if p.startsEnvironmentEnd(envName) {
			p.consumeEnvironmentEnd(envName)
			break
		}
		row := []mathLayoutNode{}
		for {
			cell := p.parseMatrixCell(envName)
			row = append(row, cell)
			if p.startsEnvironmentEnd(envName) {
				p.consumeEnvironmentEnd(envName)
				rows = append(rows, row)
				return rows
			}
			if p.consumeMatrixRowSeparator() {
				rows = append(rows, row)
				break
			}
			if p.pos < len(p.input) && p.input[p.pos] == '&' {
				p.pos++
				continue
			}
			rows = append(rows, row)
			return rows
		}
	}
	return rows
}

func (p *mathLayoutParser) parseMatrixCell(envName string) mathLayoutNode {
	var children []mathLayoutNode
	appendText := func(text string) {
		if text == "" {
			return
		}
		n := len(children)
		if n > 0 && children[n-1].kind == mathLayoutText {
			children[n-1].text += text
			return
		}
		children = append(children, mathLayoutNode{kind: mathLayoutText, text: text})
	}

	for p.pos < len(p.input) {
		if p.startsEnvironmentEnd(envName) || p.pos < len(p.input) && p.input[p.pos] == '&' || p.startsMatrixRowSeparator() {
			break
		}
		r := p.input[p.pos]
		switch r {
		case '{':
			p.pos++
			children = append(children, p.parseUntil('}').children...)
			if p.pos < len(p.input) && p.input[p.pos] == '}' {
				p.pos++
			}
		case '^', '_':
			p.pos++
			children = attachMathScript(children, r, p.parseArgumentNode())
		case '\\':
			node := p.parseCommandNode()
			if node.kind == mathLayoutText {
				appendText(node.text)
			} else if !node.isEmpty() {
				children = append(children, node)
			}
		case '~':
			children = append(children, mathLayoutNode{kind: mathLayoutSpace, widthEm: 0.333})
			p.pos++
		case ' ', '\t', '\n', '\r':
			if !p.implicitItalic {
				appendText(" ")
			}
			p.pos++
		default:
			children = appendMathAtom(children, r, p.implicitItalic)
			p.pos++
		}
	}

	return mathLayoutNode{kind: mathLayoutList, children: children}
}

func (p *mathLayoutParser) parseUntilFenceBoundary() mathLayoutNode {
	var children []mathLayoutNode
	appendText := func(text string) {
		if text == "" {
			return
		}
		n := len(children)
		if n > 0 && children[n-1].kind == mathLayoutText {
			children[n-1].text += text
			return
		}
		children = append(children, mathLayoutNode{kind: mathLayoutText, text: text})
	}

	for p.pos < len(p.input) {
		if p.startsNamedCommand("middle") || p.startsNamedCommand("right") {
			break
		}
		r := p.input[p.pos]
		switch r {
		case '{':
			p.pos++
			children = append(children, p.parseUntil('}').children...)
			if p.pos < len(p.input) && p.input[p.pos] == '}' {
				p.pos++
			}
		case '^', '_':
			p.pos++
			children = attachMathScript(children, r, p.parseArgumentNode())
		case '\\':
			node := p.parseCommandNode()
			if node.kind == mathLayoutText {
				appendText(node.text)
			} else if !node.isEmpty() {
				children = append(children, node)
			}
		case '~':
			children = append(children, mathLayoutNode{kind: mathLayoutSpace, widthEm: 0.333})
			p.pos++
		case ' ', '\t', '\n', '\r':
			if !p.implicitItalic {
				appendText(" ")
			}
			p.pos++
		default:
			children = appendMathAtom(children, r, p.implicitItalic)
			p.pos++
		}
	}
	return mathLayoutNode{kind: mathLayoutList, children: children}
}

func (p *mathLayoutParser) parseDelimiterToken() string {
	p.skipSpace()
	if p.pos >= len(p.input) {
		return ""
	}
	if p.input[p.pos] == '.' {
		p.pos++
		return ""
	}
	if p.input[p.pos] == '\\' {
		p.pos++
		if p.pos >= len(p.input) {
			return ""
		}
		r := p.input[p.pos]
		if !unicode.IsLetter(r) {
			p.pos++
			switch r {
			case '{', '}':
				return string(r)
			case '|':
				return "|"
			default:
				return string(r)
			}
		}
		start := p.pos
		for p.pos < len(p.input) && unicode.IsLetter(p.input[p.pos]) {
			p.pos++
		}
		name := string(p.input[start:p.pos])
		if delim, ok := mathTextDelimiterCommands[name]; ok {
			return delim
		}
		if mapped, ok := mathTextCommandMap[name]; ok {
			return mapped
		}
		return ""
	}
	r := p.input[p.pos]
	p.pos++
	return string(r)
}

func parseMathDelimiterSpec(text string) string {
	parser := mathLayoutParser{input: []rune(strings.TrimSpace(text))}
	return parser.parseDelimiterToken()
}

func (p *mathLayoutParser) parseBraceText() string {
	p.skipSpace()
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return ""
	}
	p.pos++
	start := p.pos
	depth := 1
	for p.pos < len(p.input) {
		switch p.input[p.pos] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				text := string(p.input[start:p.pos])
				p.pos++
				return text
			}
		}
		p.pos++
	}
	return string(p.input[start:])
}

func (p *mathLayoutParser) startsNamedCommand(name string) bool {
	if p.pos >= len(p.input)-len(name) || p.input[p.pos] != '\\' {
		return false
	}
	i := p.pos + 1
	for _, want := range name {
		if i >= len(p.input) || p.input[i] != want {
			return false
		}
		i++
	}
	return i >= len(p.input) || !unicode.IsLetter(p.input[i])
}

func (p *mathLayoutParser) consumeNamedCommand(name string) bool {
	if !p.startsNamedCommand(name) {
		return false
	}
	p.pos += 1 + len([]rune(name))
	return true
}

func (p *mathLayoutParser) startsEnvironmentEnd(name string) bool {
	save := p.pos
	if !p.consumeNamedCommand("end") {
		p.pos = save
		return false
	}
	text := p.parseBraceText()
	p.pos = save
	return text == name
}

func (p *mathLayoutParser) consumeEnvironmentEnd(name string) bool {
	save := p.pos
	if !p.consumeNamedCommand("end") {
		p.pos = save
		return false
	}
	if p.parseBraceText() != name {
		p.pos = save
		return false
	}
	return true
}

func (p *mathLayoutParser) startsMatrixRowSeparator() bool {
	if p.pos+1 >= len(p.input) || p.input[p.pos] != '\\' {
		return false
	}
	if p.input[p.pos+1] == '\\' {
		return true
	}
	if !unicode.IsLetter(p.input[p.pos+1]) {
		return false
	}
	i := p.pos + 1
	for i < len(p.input) && unicode.IsLetter(p.input[i]) {
		i++
	}
	return string(p.input[p.pos+1:i]) == "cr"
}

func (p *mathLayoutParser) consumeMatrixRowSeparator() bool {
	if !p.startsMatrixRowSeparator() {
		return false
	}
	if p.input[p.pos+1] == '\\' {
		p.pos += 2
	} else {
		p.pos++
		for p.pos < len(p.input) && unicode.IsLetter(p.input[p.pos]) {
			p.pos++
		}
	}
	return true
}

func (p *mathLayoutParser) skipSpace() {
	for p.pos < len(p.input) && unicode.IsSpace(p.input[p.pos]) {
		p.pos++
	}
}

func attachMathScript(children []mathLayoutNode, marker rune, script mathLayoutNode) []mathLayoutNode {
	if len(children) == 0 {
		return append(children, mathLayoutNode{kind: mathLayoutText, text: string(marker) + nodePlainText(script)})
	}
	last := children[len(children)-1]
	if last.kind != mathLayoutScript {
		base := last
		last = mathLayoutNode{kind: mathLayoutScript, base: &base}
	}
	if marker == '^' {
		last.super = &script
	} else {
		last.sub = &script
	}
	children[len(children)-1] = last
	return children
}

func (n mathLayoutNode) isEmpty() bool {
	return n.kind == mathLayoutText && n.text == ""
}

func nodePlainText(n mathLayoutNode) string {
	switch n.kind {
	case mathLayoutText:
		return n.text
	case mathLayoutList:
		var out strings.Builder
		for _, child := range n.children {
			out.WriteString(nodePlainText(child))
		}
		return out.String()
	case mathLayoutScript:
		base := nodePlainText(pointerNode(n.base))
		if n.sub != nil {
			base += "_" + nodePlainText(*n.sub)
		}
		if n.super != nil {
			base += "^" + nodePlainText(*n.super)
		}
		return base
	case mathLayoutFrac:
		return formatMathFraction(nodePlainText(pointerNode(n.num)), nodePlainText(pointerNode(n.den)))
	case mathLayoutSqrt:
		return "√" + groupMathAtom(nodePlainText(pointerNode(n.radicand)))
	case mathLayoutStyled:
		return nodePlainText(pointerNode(n.child))
	case mathLayoutFence:
		var out strings.Builder
		if n.left != "" {
			out.WriteString(n.left)
		}
		for i, segment := range fenceSegments(n) {
			if i > 0 && i-1 < len(n.middles) {
				out.WriteString(n.middles[i-1])
			}
			out.WriteString(nodePlainText(segment))
		}
		if n.right != "" {
			out.WriteString(n.right)
		}
		return out.String()
	case mathLayoutMatrix:
		var out strings.Builder
		if n.left != "" {
			out.WriteString(n.left)
		}
		for i, row := range n.rows {
			if i > 0 {
				out.WriteString("; ")
			}
			for j, cell := range row {
				if j > 0 {
					out.WriteByte(' ')
				}
				out.WriteString(nodePlainText(cell))
			}
		}
		if n.right != "" {
			out.WriteString(n.right)
		}
		return out.String()
	default:
		return ""
	}
}

func pointerNode(n *mathLayoutNode) mathLayoutNode {
	if n == nil {
		return mathLayoutNode{kind: mathLayoutText}
	}
	return *n
}

func fenceSegments(n mathLayoutNode) []mathLayoutNode {
	if len(n.segments) > 0 {
		return n.segments
	}
	if n.child != nil {
		return []mathLayoutNode{*n.child}
	}
	return nil
}

type mathLayoutBox struct {
	runs    []MathTextLayoutRun
	rules   []MathTextLayoutRule
	Width   float64
	Ascent  float64
	Descent float64
}

func layoutMathNode(r Measurer, n mathLayoutNode, size float64, fontKey string, opts Options) mathLayoutBox {
	switch n.kind {
	case mathLayoutText:
		return layoutMathTextRun(r, n.text, size, fontKey)
	case mathLayoutList:
		return layoutMathList(r, n.children, size, fontKey, opts)
	case mathLayoutScript:
		return layoutMathScript(r, n, size, fontKey, opts)
	case mathLayoutFrac:
		return layoutMathFrac(r, pointerNode(n.num), pointerNode(n.den), size, fontKey, opts, n.fracRule, n.fracDisp, n.left, n.right)
	case mathLayoutSqrt:
		return layoutMathSqrt(r, pointerNode(n.radicand), n.index, size, fontKey, opts)
	case mathLayoutSpace:
		return layoutMathSpace(r, n, size, fontKey)
	case mathLayoutStyled:
		return layoutMathStyled(r, n, size, fontKey, opts)
	case mathLayoutFence:
		return layoutMathFence(r, n, size, fontKey, opts)
	case mathLayoutMatrix:
		return layoutMathMatrix(r, n, size, fontKey, opts)
	default:
		return mathLayoutBox{}
	}
}

func layoutMathTextRun(r Measurer, text string, size float64, fontKey string) mathLayoutBox {
	if text == "" {
		return mathLayoutBox{}
	}
	fontKey = mathDisplayOperatorFontKey(text, fontKey)

	// Pixel-exact path: position every glyph individually from matplotlib
	// `_get_info` metrics (mirrors ship() Char/Kern packing). Falls back to the
	// whole-run path when the renderer lacks the FreeType capability (purego).
	if gm, ok := r.(GlyphMeasurer); ok {
		if infos, ok := gm.GlyphRun(text, size, fontKey); ok {
			if box, ok := layoutMathGlyphRun(text, infos, size, fontKey); ok {
				return box
			}
		}
	}

	metrics := r.MeasureText(text, size, fontKey)
	if metrics.W <= 0 {
		metrics.W = float64(len([]rune(text))) * size * 0.5
	}
	if metrics.Ascent <= 0 && metrics.Descent <= 0 {
		metrics.Ascent = size * 0.8
		metrics.Descent = size * 0.2
	}
	return mathLayoutBox{
		runs:    []MathTextLayoutRun{{Text: text, FontSize: size, FontKey: fontKey}},
		Width:   metrics.W,
		Ascent:  metrics.Ascent,
		Descent: metrics.Descent,
	}
}

// layoutMathGlyphRun packs each glyph of text at its exact matplotlib position:
// glyph i renders at cumulative Σ(advance)+Σ(kern), box width = total advance +
// kerns, ascent = max(iceberg), depth = max(height - iceberg) (Char.depth). One
// MathTextLayoutRun is emitted per glyph (baseline-aligned, layout y-down).
func layoutMathGlyphRun(text string, infos []GlyphInfo, size float64, fontKey string) (mathLayoutBox, bool) {
	runes := []rune(text)
	if len(infos) != len(runes) || len(runes) == 0 {
		return mathLayoutBox{}, false
	}
	var box mathLayoutBox
	box.runs = make([]MathTextLayoutRun, 0, len(runes))
	x := 0.0
	for i, info := range infos {
		x += info.KernToPrev
		box.runs = append(box.runs, MathTextLayoutRun{
			Text:     string(runes[i]),
			Offset:   geom.Pt{X: x, Y: 0},
			FontSize: size,
			FontKey:  fontKey,
		})
		if info.Iceberg > box.Ascent {
			box.Ascent = info.Iceberg
		}
		if depth := info.Height - info.Iceberg; depth > box.Descent {
			box.Descent = depth
		}
		x += info.Advance
	}
	box.Width = x
	return box, true
}

func mathDisplayOperatorFontKey(text, fontKey string) string {
	if !isMathDisplayOperatorGlyph(text) {
		return fontKey
	}
	switch normalizeMathFontKey(fontKey) {
	case "dejavuserif", "serif":
		return "DejaVu Serif Display"
	default:
		return "DejaVu Sans Display"
	}
}

func normalizeMathFontKey(fontKey string) string {
	fontKey = strings.ToLower(strings.TrimSpace(strings.Trim(fontKey, `"'`)))
	fontKey = strings.ReplaceAll(fontKey, " ", "")
	fontKey = strings.ReplaceAll(fontKey, "_", "")
	return strings.ReplaceAll(fontKey, "-", "")
}

func isMathDisplayOperatorGlyph(text string) bool {
	switch text {
	case "∑", "∏", "∫", "∮":
		return true
	default:
		return false
	}
}

func layoutMathSpace(r Measurer, n mathLayoutNode, size float64, fontKey string) mathLayoutBox {
	return mathLayoutBox{Width: mathQuadWidth(r, size, fontKey) * n.widthEm}
}

func mathQuadWidth(r Measurer, size float64, fontKey string) float64 {
	quad := 0.0
	if r != nil {
		quad = r.MeasureText("m", size, fontKey).W
	}
	if quad <= 0 {
		quad = size * mathSpaceScale
	}
	return quad
}

func layoutMathStyled(r Measurer, n mathLayoutNode, size float64, fontKey string, opts Options) mathLayoutBox {
	childFontKey := resolveMathFontKey(fontKey, n, opts)
	return layoutMathNode(r, pointerNode(n.child), size, childFontKey, opts)
}

func layoutMathFence(r Measurer, n mathLayoutNode, size float64, fontKey string, opts Options) mathLayoutBox {
	segments := fenceSegments(n)
	if len(segments) == 0 {
		return mathLayoutBox{}
	}

	bodyBoxes := make([]mathLayoutBox, len(segments))
	bodyAscent := 0.0
	bodyDescent := 0.0
	for i, segment := range segments {
		bodyBoxes[i] = layoutMathNode(r, segment, size, fontKey, opts)
		if bodyBoxes[i].Ascent > bodyAscent {
			bodyAscent = bodyBoxes[i].Ascent
		}
		if bodyBoxes[i].Descent > bodyDescent {
			bodyDescent = bodyBoxes[i].Descent
		}
	}

	left := layoutMathDelimiter(r, n.left, bodyAscent, bodyDescent, size, fontKey)
	middles := make([]mathLayoutBox, len(n.middles))
	for i, middle := range n.middles {
		middles[i] = layoutMathDelimiter(r, middle, bodyAscent, bodyDescent, size, fontKey)
	}
	right := layoutMathDelimiter(r, n.right, bodyAscent, bodyDescent, size, fontKey)

	var out mathLayoutBox
	x := 0.0
	out.appendTranslated(left, x, 0)
	x += left.Width
	for i, body := range bodyBoxes {
		out.appendTranslated(body, x, 0)
		x += body.Width
		if i < len(middles) {
			out.appendTranslated(middles[i], x, 0)
			x += middles[i].Width
		}
	}
	out.appendTranslated(right, x, 0)
	x += right.Width
	out.Width = x
	out.Ascent = maxFloat64(maxFloat64(left.Ascent, bodyAscent), right.Ascent)
	out.Descent = maxFloat64(maxFloat64(left.Descent, bodyDescent), right.Descent)
	for _, middle := range middles {
		out.Ascent = maxFloat64(out.Ascent, middle.Ascent)
		out.Descent = maxFloat64(out.Descent, middle.Descent)
	}
	return out
}

func layoutMathDelimiter(r Measurer, delim string, targetAscent, targetDescent, size float64, fontKey string) mathLayoutBox {
	if delim == "" {
		return mathLayoutBox{}
	}
	if targetAscent <= 0 && targetDescent <= 0 {
		targetAscent = size * 0.8
		targetDescent = size * 0.2
	}
	switch delim {
	case "|", "‖":
		return layoutMathVerticalRuleDelimiter(delim, targetAscent, targetDescent, size)
	default:
		if candidates := mathSizedDelimiterGlyphs(delim); len(candidates) > 0 {
			return layoutMathSizedDelimiter(r, candidates, targetAscent, targetDescent, size)
		}
		delimiterSize := maxFloat64(size*1.1, (targetAscent+targetDescent)*mathGlyphDelimiterScale)
		return centerMathDelimiterBox(layoutMathTextRun(r, delim, delimiterSize, mathDelimiterFontKey()), targetAscent, targetDescent)
	}
}

func layoutMathSizedDelimiter(r Measurer, candidates []mathDelimiterGlyph, targetAscent, targetDescent, size float64) mathLayoutBox {
	return autoHeightChar(r, candidates, targetAscent, targetDescent, size)
}

// mathAutoHeightBaseFontKey is the size-0 (unscaled) variant font in the DejaVu
// Sans fontset's sized-alternatives list. matplotlib never scales the size-0
// glyph when larger variants exist.
const mathAutoHeightBaseFontKey = "DejaVu Sans"

// autoHeightChar ports matplotlib _mathtext.AutoHeightChar for the DejaVu Sans
// fontset (used for auto-sized delimiters and the sqrt radical). candidates must
// be ordered smallest→largest with candidates[0] the size-0 DejaVu Sans glyph.
// It selects the first variant whose ink height+depth reaches targetTotal minus
// the 0.2*xHeight slack (else the largest), then — unless the size-0 glyph was
// chosen and alternatives exist — scales it by factor = targetTotal/(h+d) and
// shifts it down by shift_amount = targetDescent - char.depth.
func autoHeightChar(r Measurer, candidates []mathDelimiterGlyph, targetAscent, targetDescent, size float64) mathLayoutBox {
	if len(candidates) == 0 {
		return mathLayoutBox{}
	}
	targetTotal := targetAscent + targetDescent
	if targetTotal <= 0 {
		targetTotal = size
	}
	threshold := targetTotal - 0.2*mathXHeight(r, size, mathAutoHeightBaseFontKey)

	selected := candidates[len(candidates)-1]
	selectedTotal := 0.0
	for _, candidate := range candidates {
		box := layoutMathTextRun(r, candidate.text, size, candidate.fontKey)
		total := box.Ascent + box.Descent
		if total <= 0 {
			continue
		}
		selected = candidate
		selectedTotal = total
		if total >= threshold {
			break
		}
	}

	// size-0 (DejaVu Sans) is rendered unscaled, baseline-aligned (shift 0).
	if selected.fontKey == mathAutoHeightBaseFontKey && len(candidates) > 1 {
		return layoutMathTextRun(r, selected.text, size, selected.fontKey)
	}
	if selectedTotal <= 0 {
		selectedTotal = size
	}
	fontsize := size * targetTotal / selectedTotal
	if fontsize <= 0 {
		fontsize = size
	}
	box := layoutMathTextRun(r, selected.text, fontsize, selected.fontKey)
	return shiftMathBoxDown(box, targetDescent-box.Descent)
}

// shrinkMathBox ports matplotlib Node.shrink(): it LINEARLY scales a laid-out
// box's geometry (run offsets, rule rects, width, ascent, depth) by factor and
// multiplies each run's render FontSize by factor. This deliberately differs
// from re-measuring at the smaller size — matplotlib's box model positions
// shrunk content with metrics scaled from the base measurement, while the glyph
// bitmaps still rasterize at the shrunk fontsize.
func shrinkMathBox(box mathLayoutBox, factor float64) mathLayoutBox {
	out := mathLayoutBox{
		Width:   box.Width * factor,
		Ascent:  box.Ascent * factor,
		Descent: box.Descent * factor,
	}
	for _, run := range box.runs {
		run.Offset.X *= factor
		run.Offset.Y *= factor
		run.FontSize *= factor
		out.runs = append(out.runs, run)
	}
	for _, rule := range box.rules {
		rule.Rect.Min.X *= factor
		rule.Rect.Min.Y *= factor
		rule.Rect.Max.X *= factor
		rule.Rect.Max.Y *= factor
		out.rules = append(out.rules, rule)
	}
	return out
}

// shiftMathBoxRight shifts a box's contents right by dx, preserving metrics.
func shiftMathBoxRight(box mathLayoutBox, dx float64) mathLayoutBox {
	var out mathLayoutBox
	out.appendTranslated(box, dx, 0)
	out.Width = box.Width
	out.Ascent = box.Ascent
	out.Descent = box.Descent
	return out
}

// shiftMathBoxDown shifts a box's contents down by dy (layout y-down), updating
// ascent/descent (matplotlib Char.shift_amount).
func shiftMathBoxDown(box mathLayoutBox, dy float64) mathLayoutBox {
	var out mathLayoutBox
	out.appendTranslated(box, 0, dy)
	out.Width = box.Width
	out.Ascent = box.Ascent - dy
	out.Descent = box.Descent + dy
	return out
}

func mathSizedDelimiterGlyphs(delim string) []mathDelimiterGlyph {
	switch delim {
	case "(", ")", "[", "]", "{", "}", "⌊", "⌋", "⌈", "⌉", "⟨", "⟩":
		return []mathDelimiterGlyph{
			{text: delim, fontKey: mathAutoHeightBaseFontKey},
			stixSizeGlyph(1, delim),
			stixSizeGlyph(2, delim),
			stixSizeGlyph(3, delim),
			stixSizeGlyph(4, delim),
		}
	default:
		return nil
	}
}

func stixSizeGlyph(size int, text string) mathDelimiterGlyph {
	switch size {
	case 1:
		return mathDelimiterGlyph{text: text, fontKey: "STIXSizeOneSym"}
	case 2:
		return mathDelimiterGlyph{text: text, fontKey: "STIXSizeTwoSym"}
	case 3:
		return mathDelimiterGlyph{text: text, fontKey: "STIXSizeThreeSym"}
	case 4:
		return mathDelimiterGlyph{text: text, fontKey: "STIXSizeFourSym"}
	default:
		return mathDelimiterGlyph{text: text, fontKey: "STIXSizeFiveSym"}
	}
}

func centerMathDelimiterBox(box mathLayoutBox, targetAscent, targetDescent float64) mathLayoutBox {
	if box.Width <= 0 && len(box.runs) == 0 && len(box.rules) == 0 {
		return mathLayoutBox{}
	}
	targetCenter := (targetDescent - targetAscent) / 2
	boxCenter := (box.Descent - box.Ascent) / 2
	dy := targetCenter - boxCenter
	var out mathLayoutBox
	out.appendTranslated(box, 0, dy)
	out.Width = box.Width
	out.Ascent = maxFloat64(targetAscent, -dy+box.Ascent)
	out.Descent = maxFloat64(targetDescent, dy+box.Descent)
	return out
}

func layoutMathVerticalRuleDelimiter(delim string, targetAscent, targetDescent, size float64) mathLayoutBox {
	thickness := maxFloat64(size*0.065, 1.0)
	pad := size * 0.04
	targetAscent *= mathRuleDelimiterScale
	targetDescent *= mathRuleDelimiterScale
	top := -targetAscent - pad
	bottom := targetDescent + pad
	width := maxFloat64(size*0.18, thickness*2)
	centers := []float64{width / 2}
	if delim == "‖" {
		width = maxFloat64(size*0.30, thickness*4)
		gap := maxFloat64(thickness*1.4, size*0.06)
		centers = []float64{width/2 - gap/2, width/2 + gap/2}
	}
	rules := make([]MathTextLayoutRule, 0, len(centers))
	for _, center := range centers {
		rules = append(rules, MathTextLayoutRule{
			Rect: geom.Rect{
				Min: geom.Pt{X: center - thickness/2, Y: top},
				Max: geom.Pt{X: center + thickness/2, Y: bottom},
			},
		})
	}
	return mathLayoutBox{
		rules:   rules,
		Width:   width,
		Ascent:  -top,
		Descent: bottom,
	}
}

func layoutMathBracketDelimiter(delim string, targetAscent, targetDescent, size float64) mathLayoutBox {
	thickness := maxFloat64(size*0.065, 1.0)
	pad := size * 0.04
	targetAscent *= mathRuleDelimiterScale
	targetDescent *= mathRuleDelimiterScale
	top := -targetAscent - pad
	bottom := targetDescent + pad
	width := maxFloat64(size*0.28, thickness*4)
	left := delim == "[" || delim == "⌊" || delim == "⌈"
	topCap := delim == "[" || delim == "]" || delim == "⌈" || delim == "⌉"
	bottomCap := delim == "[" || delim == "]" || delim == "⌊" || delim == "⌋"
	x0 := 0.0
	x1 := thickness
	if !left {
		x0 = width - thickness
		x1 = width
	}
	rules := []MathTextLayoutRule{{
		Rect: geom.Rect{
			Min: geom.Pt{X: x0, Y: top},
			Max: geom.Pt{X: x1, Y: bottom},
		},
	}}
	if topCap {
		rules = append(rules, MathTextLayoutRule{
			Rect: geom.Rect{
				Min: geom.Pt{X: 0, Y: top},
				Max: geom.Pt{X: width, Y: top + thickness},
			},
		})
	}
	if bottomCap {
		rules = append(rules, MathTextLayoutRule{
			Rect: geom.Rect{
				Min: geom.Pt{X: 0, Y: bottom - thickness},
				Max: geom.Pt{X: width, Y: bottom},
			},
		})
	}
	return mathLayoutBox{
		rules:   rules,
		Width:   width,
		Ascent:  -top,
		Descent: bottom,
	}
}

func layoutMathMatrix(r Measurer, n mathLayoutNode, size float64, fontKey string, opts Options) mathLayoutBox {
	if len(n.rows) == 0 {
		return mathLayoutBox{}
	}

	cellBoxes := make([][]mathLayoutBox, len(n.rows))
	numCols := 0
	for i, row := range n.rows {
		cellBoxes[i] = make([]mathLayoutBox, len(row))
		if len(row) > numCols {
			numCols = len(row)
		}
		for j, cell := range row {
			cellBoxes[i][j] = layoutMathNode(r, cell, size, fontKey, opts)
		}
	}
	if numCols == 0 {
		return mathLayoutBox{}
	}

	colWidths := make([]float64, numCols)
	rowAscents := make([]float64, len(n.rows))
	rowDescents := make([]float64, len(n.rows))
	for i, row := range cellBoxes {
		for j, cell := range row {
			if cell.Width > colWidths[j] {
				colWidths[j] = cell.Width
			}
			if cell.Ascent > rowAscents[i] {
				rowAscents[i] = cell.Ascent
			}
			if cell.Descent > rowDescents[i] {
				rowDescents[i] = cell.Descent
			}
		}
		if rowAscents[i] == 0 && rowDescents[i] == 0 {
			rowAscents[i] = size * 0.5
			rowDescents[i] = size * 0.3
		}
	}

	colGap := size * 0.6
	rowGap := size * 0.4
	bodyWidth := 0.0
	for i, width := range colWidths {
		bodyWidth += width
		if i > 0 {
			bodyWidth += colGap
		}
	}
	bodyHeight := 0.0
	for i := range n.rows {
		bodyHeight += rowAscents[i] + rowDescents[i]
		if i > 0 {
			bodyHeight += rowGap
		}
	}

	left := layoutMathDelimiter(r, n.left, bodyHeight/2, bodyHeight/2, size, fontKey)
	right := layoutMathDelimiter(r, n.right, bodyHeight/2, bodyHeight/2, size, fontKey)
	leftGap := 0.0
	rightGap := 0.0
	if left.Width > 0 {
		leftGap = size * 0.18
	}
	if right.Width > 0 {
		rightGap = size * 0.18
	}

	var out mathLayoutBox
	x := 0.0
	out.appendTranslated(left, x, 0)
	x += left.Width
	if left.Width > 0 {
		x += leftGap
	}

	top := -bodyHeight / 2
	for i, row := range cellBoxes {
		baselineY := top + rowAscents[i]
		cellX := x
		for j := 0; j < numCols; j++ {
			var cell mathLayoutBox
			if j < len(row) {
				cell = row[j]
			}
			cellOffsetX := cellX + (colWidths[j]-cell.Width)/2
			out.appendTranslated(cell, cellOffsetX, baselineY)
			cellX += colWidths[j] + colGap
		}
		top += rowAscents[i] + rowDescents[i] + rowGap
	}
	x += bodyWidth
	if right.Width > 0 {
		x += rightGap
	}
	out.appendTranslated(right, x, 0)
	out.Width = left.Width + leftGap + bodyWidth + rightGap + right.Width
	out.Ascent = bodyHeight / 2
	out.Descent = bodyHeight / 2
	if left.Ascent > out.Ascent {
		out.Ascent = left.Ascent
	}
	if right.Ascent > out.Ascent {
		out.Ascent = right.Ascent
	}
	if left.Descent > out.Descent {
		out.Descent = left.Descent
	}
	if right.Descent > out.Descent {
		out.Descent = right.Descent
	}
	return out
}

func layoutMathList(r Measurer, children []mathLayoutNode, size float64, fontKey string, opts Options) mathLayoutBox {
	var out mathLayoutBox
	x := 0.0
	for _, child := range children {
		box := layoutMathNode(r, child, size, fontKey, opts)
		out.appendTranslated(box, x, 0)
		x += box.Width
		if box.Ascent > out.Ascent {
			out.Ascent = box.Ascent
		}
		if box.Descent > out.Descent {
			out.Descent = box.Descent
		}
	}
	out.Width = x
	return out
}

func layoutMathScript(r Measurer, n mathLayoutNode, size float64, fontKey string, opts Options) mathLayoutBox {
	if isMathLimitOperator(pointerNode(n.base)) {
		return layoutMathLimits(r, n, size, fontKey, opts)
	}

	// Faithful port of matplotlib 3.8.4 _mathtext.Parser.subsuper (regular
	// sub/superscript branch). Vertical shifts come from the DejaVu Sans font
	// constants scaled by x-height; the script box is shrunk once. Layout space
	// is y-down (negative Y above baseline); matplotlib height=ascent, depth=descent.
	base := layoutMathNode(r, pointerNode(n.base), size, fontKey, opts)
	// matplotlib wraps each script as Hlist([Kern, script]).shrink(), which
	// LINEARLY scales the base-size box metrics by SHRINK_FACTOR (it does not
	// re-measure at the smaller size — see shrinkMathBox / layoutMathFrac). So
	// lay each script out at the base size and shrink the box.
	layoutScript := func(node mathLayoutNode) mathLayoutBox {
		return shrinkMathBox(layoutMathNode(r, node, size, fontKey, opts), mathFracShrink)
	}
	xHeight := mathXHeight(r, size, fontKey)
	ruleThickness := mathUnderlineThickness(r, size, fontKey)

	baseText := nodePlainText(pointerNode(n.base))
	dropSub := isMathDropSubGlyph(baseText)
	// matplotlib is_slanted: italic variable faces are slanted, and so are the
	// drop-sub integral operators (∫∮), which render from a slanted math face.
	// The slant comes from the node's font style (FontStyleItalic), not the
	// resolved fontKey, which is empty at layout time.
	slanted := nodeIsSlanted(pointerNode(n.base)) || dropSub
	lcHeight := base.Ascent
	lcBaseline := 0.0
	if dropSub {
		lcBaseline = base.Descent
	}

	// Horizontal kerning of the scripts relative to the nucleus advance.
	superKern := mathScriptDelta * xHeight
	subKern := mathScriptDelta * xHeight
	if slanted {
		superKern += mathScriptDelta * xHeight
		superKern += mathScriptDeltaSlanted * (lcHeight - xHeight*2.0/3.0)
		if dropSub {
			subKern = (3*mathScriptDelta - mathScriptDeltaIntegral) * lcHeight
			superKern = (3*mathScriptDelta + mathScriptDeltaIntegral) * lcHeight
		} else {
			subKern = 0
		}
	}

	// matplotlib builds each script as Hlist([Kern(kern), script]) and calls
	// x.shrink() once, which scales the kern by SHRINK_FACTOR (the script box is
	// also laid out at scriptSize). The vertical shift_amount and the trailing
	// script_space are applied AFTER shrink, so they are not scaled.
	superKern *= mathFracShrink
	subKern *= mathFracShrink

	var out mathLayoutBox
	out.appendTranslated(base, 0, 0)
	out.Width = base.Width
	out.Ascent = base.Ascent
	out.Descent = base.Descent
	scriptMaxX := base.Width

	switch {
	case n.super == nil && n.sub != nil:
		// node757: subscript without superscript.
		sub := layoutScript(*n.sub)
		shiftDown := mathScriptSub1 * xHeight
		if dropSub {
			shiftDown = lcBaseline + mathScriptSubdrop*xHeight
		}
		scriptX := base.Width + subKern
		out.appendTranslated(sub, scriptX, shiftDown)
		scriptMaxX = maxFloat64(scriptMaxX, scriptX+sub.Width)
		out.Descent = maxFloat64(out.Descent, shiftDown+sub.Descent)
		out.Ascent = maxFloat64(out.Ascent, sub.Ascent-shiftDown)
	case n.super != nil:
		super := layoutScript(*n.super)
		shiftUp := mathScriptSup1 * xHeight
		if dropSub {
			shiftUp = lcHeight - mathScriptSubdrop*xHeight
		}
		superX := base.Width + superKern
		if n.sub == nil {
			out.appendTranslated(super, superX, -shiftUp)
			scriptMaxX = maxFloat64(scriptMaxX, superX+super.Width)
			out.Ascent = maxFloat64(out.Ascent, shiftUp+super.Ascent)
			out.Descent = maxFloat64(out.Descent, super.Descent-shiftUp)
		} else {
			// node759: both sub and superscript; if they would collide, raise super.
			sub := layoutScript(*n.sub)
			shiftDown := mathScriptSub2 * xHeight
			if dropSub {
				shiftDown = lcBaseline + mathScriptSubdrop*xHeight
			}
			clr := 2.0*ruleThickness - ((shiftUp - super.Descent) - (sub.Ascent - shiftDown))
			if clr > 0 {
				shiftUp += clr
			}
			subX := base.Width + subKern
			out.appendTranslated(super, superX, -shiftUp)
			out.appendTranslated(sub, subX, shiftDown)
			scriptMaxX = maxFloat64(scriptMaxX, maxFloat64(superX+super.Width, subX+sub.Width))
			out.Ascent = maxFloat64(out.Ascent, shiftUp+super.Ascent)
			out.Descent = maxFloat64(out.Descent, shiftDown+sub.Descent)
		}
	}

	// script_space trailing kern, except after drop-sub operators.
	if !dropSub {
		scriptMaxX += mathScriptSpace * xHeight
	}
	out.Width = scriptMaxX
	return out
}

// nodeIsSlanted reports whether the nucleus renders with a slanted (italic)
// face, matching matplotlib's Char.is_slanted() for the sub/superscript kerning
// adjustments. The slant is carried by the layout node's FontStyleItalic (math
// variables are implicitly italic); the resolved fontKey is empty at layout time
// so it cannot be inspected. For a multi-element nucleus the last element's slant
// is used (matplotlib keys off last_char).
func nodeIsSlanted(n mathLayoutNode) bool {
	switch n.kind {
	case mathLayoutStyled:
		if n.style == FontStyleItalic {
			return true
		}
		if n.style == FontStyleNormal && n.child != nil {
			return nodeIsSlanted(*n.child)
		}
		return false
	case mathLayoutList:
		if len(n.children) == 0 {
			return false
		}
		return nodeIsSlanted(n.children[len(n.children)-1])
	default:
		return false
	}
}

func layoutMathLimits(r Measurer, n mathLayoutNode, size float64, fontKey string, opts Options) mathLayoutBox {
	base := layoutMathNode(r, pointerNode(n.base), size, fontKey, opts)
	// As in layoutMathScript, the limits are shrunk via linear box scaling
	// (matplotlib Node.shrink()), not re-measured at the smaller font size.
	layoutScript := func(node mathLayoutNode) mathLayoutBox {
		return shrinkMathBox(layoutMathNode(r, node, size, fontKey, opts), mathFracShrink)
	}

	var super, sub mathLayoutBox
	if n.super != nil {
		super = layoutScript(*n.super)
	}
	if n.sub != nil {
		sub = layoutScript(*n.sub)
	}

	width := base.Width
	if super.Width > width {
		width = super.Width
	}
	if sub.Width > width {
		width = sub.Width
	}

	baseX := (width - base.Width) / 2
	superX := (width - super.Width) / 2
	subX := (width - sub.Width) / 2
	// matplotlib 3.8.4 subsuper over/under stack: vgap = rule_thickness * 3.
	gap := mathUnderlineThickness(r, size, fontKey) * 3.0

	var out mathLayoutBox
	out.Width = width
	out.appendTranslated(base, baseX, 0)
	out.Ascent = base.Ascent
	out.Descent = base.Descent

	if n.super != nil {
		y := -(base.Ascent + gap + super.Descent)
		out.appendTranslated(super, superX, y)
		out.Ascent = maxFloat64(out.Ascent, -y+super.Ascent)
		out.Descent = maxFloat64(out.Descent, y+super.Descent)
	}
	if n.sub != nil {
		y := base.Descent + gap + sub.Ascent
		out.appendTranslated(sub, subX, y)
		out.Ascent = maxFloat64(out.Ascent, -y+sub.Ascent)
		out.Descent = maxFloat64(out.Descent, y+sub.Descent)
	}
	return out
}

func isMathLimitOperator(n mathLayoutNode) bool {
	return n.kind == mathLayoutText && isMathLimitText(n.text)
}

// FontConstantsBase defaults from matplotlib's lib/matplotlib/_mathtext.py.
// These sub/superscript constants are unchanged between matplotlib 3.8.4 and
// 3.10.9 (verified against the vendored 3.10.9 source), and DejaVuSansFontConstants
// is literally `pass` in both versions, so DejaVu Sans uses these base values.
// They are multiples of the scaled x-height (sup/sub shifts) or the pixel font
// size (underline). The reference images (now generated under 3.10.9) match.
const (
	// matplotlib FontConstantsBase sub/superscript constants (DejaVu Sans = base).
	mathScriptDelta         = 0.025 // delta
	mathScriptDeltaSlanted  = 0.2   // delta_slanted
	mathScriptDeltaIntegral = 0.1   // delta_integral
	mathScriptSpace         = 0.05  // script_space
	mathScriptSubdrop       = 0.4   // subdrop
	mathScriptSup1          = 0.7   // sup1
	mathScriptSub1          = 0.3   // sub1
	mathScriptSub2          = 0.5   // sub2

	// Design-em ratio for DejaVu Sans' x-height (xHeight 1120 / unitsPerEm 2048),
	// used only to recover the pixel font size for underline thickness. NOTE:
	// matplotlib's get_xheight does NOT use this ratio for DejaVu (which has no
	// PCLT table) — it returns the iceberg (top ink extent) of glyph 'x'. See
	// mathXHeight/mathFontSizePixels for how this cancels to the iceberg value.
	mathDejaVuSansXHeight = 1120.0 / 2048.0
	mathUnderlineRatio    = 0.75 / 12.0 // get_underline_thickness (hardcoded)

	// SHRINK_FACTOR: TeX style step shrinking numerator/denominator and scripts.
	mathFracShrink = 0.70
)

// mathIceberg returns matplotlib's "iceberg" for a glyph: the top ink extent
// above the baseline (FreeType horiBearingY), in device pixels. Layout bounds
// are y-down with negative Y above the baseline, so the iceberg is -BoundsY.
// Returns 0 when the renderer cannot supply ink bounds.
func mathIceberg(r Measurer, text string, size float64, fontKey string) float64 {
	if r == nil {
		return 0
	}
	// Pixel-exact: iceberg = max(horiBearingY/64) over the glyphs (matplotlib
	// `_get_info` iceberg / get_xheight). Falls back to hinted ink bounds.
	if gm, ok := r.(GlyphMeasurer); ok {
		if infos, ok := gm.GlyphRun(text, size, fontKey); ok && len(infos) > 0 {
			ice := 0.0
			for _, info := range infos {
				if info.Iceberg > ice {
					ice = info.Iceberg
				}
			}
			if ice > 0 {
				return ice
			}
		}
	}
	m := r.MeasureText(text, size, fontKey)
	if m.BoundsH <= 0 {
		return 0
	}
	if ice := -m.BoundsY; ice > 0 {
		return ice
	}
	return 0
}

// mathXHeight is matplotlib TruetypeFonts.get_xheight for DejaVu Sans. DejaVu
// has no PCLT table, so matplotlib uses the "poor man's xHeight" = the iceberg
// (top ink extent above the baseline) of glyph 'x'. We return that directly.
func mathXHeight(r Measurer, size float64, fontKey string) float64 {
	if xh := mathIceberg(r, "x", size, fontKey); xh > 0 {
		return xh
	}
	return mathFontSizePixels(r, size, fontKey) * mathDejaVuSansXHeight
}

// mathFontSizePixels recovers matplotlib's box-model unit — the font size in
// device pixels (fontsize*dpi/72) — which is needed for the (font-independent)
// underline thickness. The Measurer interface exposes only the point size, so
// we recover the pixel size from the measured x-height of 'x' (its iceberg)
// divided by DejaVu Sans' design-em x-height ratio (1120/2048). For DejaVu 'x'
// (which sits on the baseline) this iceberg equals the full ink height.
func mathFontSizePixels(r Measurer, size float64, fontKey string) float64 {
	// Exact when the renderer exposes DPI (matplotlib fontsize*dpi/72). The
	// x-height reconstruction below is an approximation used only as a fallback
	// (purego/mock renderers) — its ratio drifts with the autohinter per size,
	// which is why fraction-bar/thickness positions were off without exact DPI.
	if dp, ok := r.(DPIMeasurer); ok {
		if dpi := dp.DPI(); dpi > 0 {
			return size * dpi / 72.0
		}
	}
	if xh := mathIceberg(r, "x", size, fontKey); xh > 0 {
		return xh / mathDejaVuSansXHeight
	}
	return mathQuadWidth(r, size, fontKey)
}

// mathUnderlineThickness is matplotlib get_underline_thickness: hardcoded to
// (0.75/12)*fontsize*dpi/72 because upstream found font underline metrics too
// unreliable. The base unit for fraction bars, radical rules, and script gaps.
func mathUnderlineThickness(r Measurer, size float64, fontKey string) float64 {
	return mathFontSizePixels(r, size, fontKey) * mathUnderlineRatio
}

func isMathDropSubGlyph(text string) bool {
	switch text {
	case "∫", "∮":
		return true
	default:
		return false
	}
}

func isMathLimitText(text string) bool {
	switch text {
	case "∑", "∏", "lim", "lim inf", "lim sup", "max", "min", "sup", "inf":
		return true
	default:
		return false
	}
}

func layoutMathFrac(r Measurer, num, den mathLayoutNode, size float64, fontKey string, opts Options, rule, display bool, leftDelim, rightDelim string) mathLayoutBox {
	// matplotlib shrinks the numerator/denominator with Node.shrink(), which
	// LINEARLY scales the base-size box metrics by SHRINK_FACTOR (it does NOT
	// re-measure at the smaller size — the hinted iceberg at the shrunk size
	// differs, e.g. '1' iceberg is 10.0 at fs10 → 7.0 scaled, but 8.0 if
	// re-measured at fs7). The glyph bitmap still renders at the shrunk size.
	// So measure at the base size and shrink the box (scaling render FontSize).
	numBox := layoutMathNode(r, num, size, fontKey, opts)
	denBox := layoutMathNode(r, den, size, fontKey, opts)
	if !display { // TEXTSTYLE shrinks once; DISPLAYSTYLE not at all
		numBox = shrinkMathBox(numBox, mathFracShrink)
		denBox = shrinkMathBox(denBox, mathFracShrink)
	}
	thickness := mathUnderlineThickness(r, size, fontKey)
	ruleThickness := thickness
	if !rule {
		ruleThickness = 0
	}
	contentWidth := maxFloat64(numBox.Width, denBox.Width)
	width := contentWidth + 2*thickness // matplotlib trailing Hbox(thickness*2)
	ruleWidth := contentWidth
	// matplotlib centres num/den via HCentered = Hlist([Glue('ss'), x, Glue('ss')]).
	// hlist_out ROUNDS the stretched glue (round(glue_set*cur_glue)), so the left
	// pad is round((contentWidth-boxWidth)/2), not the exact half — a 0.29px
	// centering offset rounds to 0 (e.g. u and v in a 2-row matrix share an ox).
	numX := math.RoundToEven((contentWidth - numBox.Width) / 2)
	denX := math.RoundToEven((contentWidth - denBox.Width) / 2)

	// Faithful port of matplotlib _mathtext.Parser._genfrac. This algorithm is
	// unchanged between 3.8.4 and 3.10.9 (the vendored 3.10.9 source has no
	// axis_height-based variant — _genfrac is still "="-centered). The
	// numerator/rule/denominator stack is Vlist[cnum, Vbox(0,2t), Hrule,
	// Vbox(0,2t), cden] (Hrule height=depth=t/2, so a rule-less fraction keeps a
	// 4t gap), shifted so the rule sits in the middle of "=": shift =
	// cden.height - ("=" center - 3t). Layout space is y-down (negative Y above
	// baseline); matplotlib height=ascent, depth=descent.
	space := thickness * 2.0
	// matplotlib centres the fraction line on "=": shift uses (ymax+ymin)/2 of the
	// "=" glyph. Use exact _get_info bbox when available (GlyphRun), else hinted.
	eqCenter := 0.0
	if gm, ok := r.(GlyphMeasurer); ok {
		if infos, ok := gm.GlyphRun("=", size, fontKey); ok && len(infos) > 0 {
			eqCenter = (infos[0].Ymax + infos[0].Ymin) / 2
		}
	}
	if eqCenter == 0 {
		eq := r.MeasureText("=", size, fontKey)
		eqCenter = (eq.Ascent + eq.Descent) / 2
		if eq.BoundsH > 0 {
			eqCenter = -(eq.BoundsY + eq.BoundsH/2)
		}
	}
	shift := denBox.Ascent - (eqCenter - thickness*3.0)
	denY := shift
	// num→den gap = num.depth + 2*space + ruleAdvance + den.height, where the
	// Hrule advances cur_v by its full height (ruleThickness) BEFORE drawing in
	// vlist_out (a rule-less binom has ruleThickness 0 → 4t gap, a \frac → 5t).
	numY := shift - denBox.Ascent - 2*space - ruleThickness - numBox.Descent
	// The Hrule's pre-advance places its rect top one space below the den top:
	// ruleTop = (denY - den.height) - space (NOT -3t — the earlier −3t put the
	// bar one thickness too high; see vlist_out Box branch).
	ruleTop := shift - denBox.Ascent - space
	vlistHeight := numBox.Ascent + numBox.Descent + 2*space + ruleThickness + denBox.Ascent
	ascent := vlistHeight - shift
	descent := denBox.Descent + shift

	out := mathLayoutBox{
		Width:   width,
		Ascent:  ascent,
		Descent: descent,
	}
	if rule {
		out.rules = append(out.rules, MathTextLayoutRule{
			Rect: geom.Rect{
				Min: geom.Pt{X: 0, Y: ruleTop},
				Max: geom.Pt{X: ruleWidth, Y: ruleTop + ruleThickness},
			},
		})
	}
	out.appendTranslated(numBox, numX, numY)
	out.appendTranslated(denBox, denX, denY)
	if leftDelim == "" && rightDelim == "" {
		return out
	}

	delimAscent := out.Ascent
	delimDescent := out.Descent
	left := layoutMathDelimiter(r, leftDelim, delimAscent, delimDescent, size, fontKey)
	right := layoutMathDelimiter(r, rightDelim, delimAscent, delimDescent, size, fontKey)
	var delimited mathLayoutBox
	x := 0.0
	delimited.appendTranslated(left, x, 0)
	x += left.Width
	delimited.appendTranslated(out, x, 0)
	x += out.Width
	delimited.appendTranslated(right, x, 0)
	x += right.Width
	delimited.Width = x
	delimited.Ascent = maxFloat64(maxFloat64(left.Ascent, out.Ascent), right.Ascent)
	delimited.Descent = maxFloat64(maxFloat64(left.Descent, out.Descent), right.Descent)
	return delimited
}

// mathSqrtRadicalGlyphs lists the auto-height variants of the radical sign for
// the DejaVu Sans fontset: size-0 DejaVu Sans, then STIX size variants 1-3
// (matplotlib drops the largest STIX radical: alternatives[:-1]).
func mathSqrtRadicalGlyphs() []mathDelimiterGlyph {
	return []mathDelimiterGlyph{
		{text: "√", fontKey: mathAutoHeightBaseFontKey},
		{text: "√", fontKey: "STIXSizeOneSym"},
		{text: "√", fontKey: "STIXSizeTwoSym"},
		{text: "√", fontKey: "STIXSizeThreeSym"},
	}
}

func layoutMathSqrt(r Measurer, radicand mathLayoutNode, index *mathLayoutNode, size float64, fontKey string, opts Options) mathLayoutBox {
	thickness := mathUnderlineThickness(r, size, fontKey)
	radicandBox := layoutMathNode(r, radicand, size, fontKey, opts)
	// matplotlib sqrt(): the radical (check) is an AutoHeightChar sized to the
	// body height + 5*thickness (extra so it doesn't look cramped), body depth.
	targetHeight := radicandBox.Ascent + 5*thickness
	targetDepth := radicandBox.Descent
	root := autoHeightChar(r, mathSqrtRadicalGlyphs(), targetHeight, targetDepth, size)

	// matplotlib re-derives height/depth from the (possibly scaled) radical:
	//   height = check.height - check.shift_amount  (= root.Ascent)
	//   depth  = check.depth  + check.shift_amount  (= root.Descent)
	// then builds rightside = Vlist[Hrule, Glue('fill'), padded_body] and
	// vpacks it to EXACTLY height + extra, where
	//   extra = (fontsize*dpi)/(100*12) = mathFontSizePixels * 72/1200.
	bodyHeight := root.Ascent
	extra := mathFontSizePixels(r, size, fontKey) * (72.0 / 1200.0)
	rightsideHeight := bodyHeight + extra
	padding := 2 * thickness // Hbox(2*thickness) on each side of the body

	// vlist_out stacks (Hrule total = thickness) + (Glue('fill')) + (body).
	// The natural height above the body baseline is thickness + body.height; the
	// glue stretches by `stretch` to fill rightsideHeight, and vlist_out ROUNDS
	// the stretched glue — so the body baseline lands round(stretch)-stretch
	// below the sqrt baseline (a sub-pixel shift that the int-blit needs exact).
	stretch := rightsideHeight - thickness - radicandBox.Ascent
	bodyDy := math.Round(stretch) - stretch

	// Vinculum (Hrule): vlist_out advances cur_v by the Hrule's full height
	// (thickness) from the top edge BEFORE drawing, so the rule's top sits one
	// thickness below the box top, i.e. at -(rightsideHeight) + thickness.
	ruleTop := -rightsideHeight + thickness

	var out mathLayoutBox
	out.appendTranslated(root, 0, 0)
	bodyX := root.Width + padding
	out.appendTranslated(shiftMathBoxDown(radicandBox, bodyDy), bodyX, 0)
	out.Width = root.Width + radicandBox.Width + 2*padding
	// sqrt hlist hpack: height = rightside.height (= rightsideHeight), depth =
	// max(check.depth+shift, rightside.depth) = root.Descent (= radicand depth).
	out.Ascent = rightsideHeight
	out.Descent = maxFloat64(root.Descent, radicandBox.Descent)
	out.rules = append(out.rules, MathTextLayoutRule{
		Rect: geom.Rect{
			Min: geom.Pt{X: root.Width, Y: ruleTop},
			Max: geom.Pt{X: out.Width, Y: ruleTop + thickness},
		},
	})

	if index != nil {
		// matplotlib sqrt: the root index is shrunk twice (SHRINK_FACTOR^2) and
		// laid out as Hlist([root_vlist, Kern(-check.width*0.5), check, rightside])
		// with the index (root_vlist) pinned at x=0. The radical therefore sits at
		// x = index.Width - check.Width*0.5 (the box origin is the index's left
		// edge). When that overhang is positive, shift the radical+body+rule right
		// by it and keep the index at 0; otherwise the radical stays at 0 and the
		// index sits root.Width*0.5 - index.Width to its left (origin at radical).
		// The index is shifted up by height*0.6 (matplotlib's hard-coded 0.6 hack).
		indexBox := layoutMathNode(r, *index, size*mathFracShrink*mathFracShrink, fontKey, opts)
		over := indexBox.Width - root.Width*0.5
		indexX := 0.0
		if over > 0 {
			out = shiftMathBoxRight(out, over)
			out.Width += over
		} else {
			indexX = -over // = root.Width*0.5 - indexBox.Width
		}
		y := -root.Ascent * 0.6
		out.appendTranslated(indexBox, indexX, y)
		out.Ascent = maxFloat64(out.Ascent, -y+indexBox.Ascent)
	}
	return out
}

func (b *mathLayoutBox) appendTranslated(child mathLayoutBox, dx, dy float64) {
	for _, run := range child.runs {
		run.Offset.X += dx
		run.Offset.Y += dy
		b.runs = append(b.runs, run)
	}
	for _, rule := range child.rules {
		rule.Rect.Min.X += dx
		rule.Rect.Max.X += dx
		rule.Rect.Min.Y += dy
		rule.Rect.Max.Y += dy
		b.rules = append(b.rules, rule)
	}
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

var mathTextSpacingCommandWidths = map[string]float64{
	",":             0.166,
	":":             0.222,
	";":             0.278,
	"thinspace":     0.166,
	"medspace":      0.222,
	"thickspace":    0.278,
	"negthinspace":  -0.166,
	"negmedspace":   -0.222,
	"negthickspace": -0.278,
	"enspace":       0.5,
	"enskip":        0.5,
	"quad":          1.0,
	"qquad":         2.0,
}

func parseMathSpaceDimension(text string) float64 {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	i := 0
	for i < len(text) {
		c := text[i]
		if c != '+' && c != '-' && c != '.' && (c < '0' || c > '9') {
			break
		}
		i++
	}
	if i == 0 || text[:i] == "+" || text[:i] == "-" || text[:i] == "." {
		return 0
	}
	value, err := strconv.ParseFloat(text[:i], 64)
	if err != nil {
		return 0
	}
	unit := strings.TrimSpace(text[i:])
	switch unit {
	case "", "em":
		return value
	case "ex":
		return value * 0.5
	case "mu":
		return value / 18
	case "pt":
		return value / 10
	default:
		return value
	}
}

func resolveMathFontKey(base string, n mathLayoutNode, opts Options) string {
	request := FontRequest{
		Families: append([]string(nil), n.families...),
		Style:    n.style,
		Weight:   n.weight,
	}
	if opts.FontResolver != nil {
		if resolved := strings.TrimSpace(opts.FontResolver.ResolveMathFontKey(base, request)); resolved != "" {
			return resolved
		}
	}
	if len(request.Families) > 0 {
		return strings.Join(request.Families, ", ")
	}
	return base
}

func matrixEnvironmentDelimiters(name string) (left, right string, ok bool) {
	switch name {
	case "matrix", "array":
		return "", "", true
	case "pmatrix":
		return "(", ")", true
	case "bmatrix":
		return "[", "]", true
	case "Bmatrix":
		return "{", "}", true
	case "vmatrix":
		return "|", "|", true
	case "Vmatrix":
		return "‖", "‖", true
	default:
		return "", "", false
	}
}
