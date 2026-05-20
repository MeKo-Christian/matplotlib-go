package mathtext

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// Metrics is the renderer-neutral text measurement subset needed for MathText
// layout.
type Metrics struct{ W, H, Ascent, Descent float64 }

// Measurer measures text for one font key and size.
type Measurer interface {
	MeasureText(text string, size float64, fontKey string) Metrics
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

const mathGlyphDelimiterScale = 0.45

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
		children = append(children, mathLayoutNode{kind: mathLayoutSpace, widthEm: 0.222})
		children = appendMathText(children, text)
		children = append(children, mathLayoutNode{kind: mathLayoutSpace, widthEm: 0.222})
	} else {
		children = appendMathText(children, text)
	}
	p.pos++
	return children
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
		default:
			appendText(string(r))
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
		default:
			appendText(string(r))
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
		return mathLayoutBox{Width: size * n.widthEm}
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
	case "[", "]", "⌊", "⌋", "⌈", "⌉":
		return layoutMathBracketDelimiter(delim, targetAscent, targetDescent, size)
	default:
		delimiterSize := maxFloat64(size*1.1, (targetAscent+targetDescent)*mathGlyphDelimiterScale)
		return centerMathDelimiterBox(layoutMathTextRun(r, delim, delimiterSize, mathDelimiterFontKey()), targetAscent, targetDescent)
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

	delimiterSize := maxFloat64(size, bodyHeight*mathGlyphDelimiterScale)
	delimiterFontKey := mathDelimiterFontKey()
	left := layoutMathTextRun(r, n.left, delimiterSize, delimiterFontKey)
	right := layoutMathTextRun(r, n.right, delimiterSize, delimiterFontKey)
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

	base := layoutMathNode(r, pointerNode(n.base), size, fontKey, opts)
	scriptSize := size * 0.7
	x := base.Width
	var out mathLayoutBox
	out.appendTranslated(base, 0, 0)
	out.Width = base.Width
	out.Ascent = base.Ascent
	out.Descent = base.Descent

	scriptWidth := 0.0
	if n.super != nil {
		super := layoutMathNode(r, *n.super, scriptSize, fontKey, opts)
		y := -maxFloat64(base.Ascent*0.55, scriptSize*0.35)
		out.appendTranslated(super, x, y)
		scriptWidth = maxFloat64(scriptWidth, super.Width)
		out.Ascent = maxFloat64(out.Ascent, -y+super.Ascent)
		out.Descent = maxFloat64(out.Descent, y+super.Descent)
	}
	if n.sub != nil {
		sub := layoutMathNode(r, *n.sub, scriptSize, fontKey, opts)
		y := maxFloat64(base.Descent*0.70, scriptSize*0.25)
		out.appendTranslated(sub, x, y)
		scriptWidth = maxFloat64(scriptWidth, sub.Width)
		out.Ascent = maxFloat64(out.Ascent, -y+sub.Ascent)
		out.Descent = maxFloat64(out.Descent, y+sub.Descent)
	}
	out.Width = base.Width + scriptWidth
	return out
}

func layoutMathLimits(r Measurer, n mathLayoutNode, size float64, fontKey string, opts Options) mathLayoutBox {
	baseSize := size
	if baseText := nodePlainText(pointerNode(n.base)); isMathLargeGlyph(baseText) {
		baseSize *= 1.2
	}
	base := layoutMathNode(r, pointerNode(n.base), baseSize, fontKey, opts)
	scriptSize := size * 0.7

	var super, sub mathLayoutBox
	if n.super != nil {
		super = layoutMathNode(r, *n.super, scriptSize, fontKey, opts)
	}
	if n.sub != nil {
		sub = layoutMathNode(r, *n.sub, scriptSize, fontKey, opts)
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
	gap := size * 0.14

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

func isMathLargeGlyph(text string) bool {
	switch text {
	case "∑", "∏":
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
	childSize := size * 0.70
	if display {
		childSize = size
	}
	numBox := layoutMathNode(r, num, childSize, fontKey, opts)
	denBox := layoutMathNode(r, den, childSize, fontKey, opts)
	padding := size * 0.18
	gap := size * 0.14
	ruleThickness := maxFloat64(size*0.045, 0.5)
	if !rule {
		ruleThickness = 0
	}
	width := maxFloat64(numBox.Width, denBox.Width) + 2*padding
	numX := (width - numBox.Width) / 2
	denX := (width - denBox.Width) / 2
	numY := -(gap + ruleThickness/2 + numBox.Descent)
	denY := gap + ruleThickness/2 + denBox.Ascent

	out := mathLayoutBox{
		Width:   width,
		Ascent:  -numY + numBox.Ascent,
		Descent: denY + denBox.Descent,
	}
	if rule {
		out.rules = append(out.rules, MathTextLayoutRule{
			Rect: geom.Rect{
				Min: geom.Pt{X: 0, Y: -ruleThickness / 2},
				Max: geom.Pt{X: width, Y: ruleThickness / 2},
			},
		})
	}
	out.appendTranslated(numBox, numX, numY)
	out.appendTranslated(denBox, denX, denY)
	if leftDelim == "" && rightDelim == "" {
		return out
	}

	left := layoutMathDelimiter(r, leftDelim, out.Ascent, out.Descent, size, fontKey)
	right := layoutMathDelimiter(r, rightDelim, out.Ascent, out.Descent, size, fontKey)
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

func layoutMathSqrt(r Measurer, radicand mathLayoutNode, index *mathLayoutNode, size float64, fontKey string, opts Options) mathLayoutBox {
	root := layoutMathTextRun(r, "√", size, fontKey)
	radicandBox := layoutMathNode(r, radicand, size, fontKey, opts)
	padding := size * 0.08
	ruleThickness := maxFloat64(size*0.04, 0.5)
	ruleY := -radicandBox.Ascent - padding

	var out mathLayoutBox
	out.appendTranslated(root, 0, 0)
	out.appendTranslated(radicandBox, root.Width, 0)
	out.Width = root.Width + radicandBox.Width
	out.Ascent = maxFloat64(root.Ascent, -ruleY+ruleThickness)
	out.Descent = maxFloat64(root.Descent, radicandBox.Descent)
	out.rules = append(out.rules, MathTextLayoutRule{
		Rect: geom.Rect{
			Min: geom.Pt{X: root.Width, Y: ruleY},
			Max: geom.Pt{X: out.Width, Y: ruleY + ruleThickness},
		},
	})

	if index != nil {
		indexBox := layoutMathNode(r, *index, size*0.55, fontKey, opts)
		x := -indexBox.Width * 0.65
		y := -root.Ascent * 0.55
		out.appendTranslated(indexBox, x, y)
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
