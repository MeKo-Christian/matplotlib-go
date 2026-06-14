package render

import (
	"math"
	"unicode"
	"unicode/utf8"

	"github.com/cwbudde/matplotlib-go/geom"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"golang.org/x/text/unicode/bidi"
	"golang.org/x/text/unicode/norm"
)

// TextDirection describes the intended logical flow of shaped text.
type TextDirection string

const (
	TextDirectionLTR TextDirection = "ltr"
	TextDirectionRTL TextDirection = "rtl"
)

// TextFeature describes one OpenType feature toggle requested for shaping.
// The current fallback shaper supports selected GSUB and positioning features.
type TextFeature struct {
	Tag   string
	Value int
}

// TextShapingOptions carries font and script context for text shaping.
type TextShapingOptions struct {
	FontKey   string
	Direction TextDirection
	Script    string
	Language  string
	Features  []TextFeature
}

// ShapedGlyph is one glyph positioned by the shared text shaping layer.
type ShapedGlyph struct {
	Rune       rune
	Cluster    int
	Face       FontFace
	GlyphIndex sfnt.GlyphIndex
	Origin     geom.Pt
	Offset     geom.Pt
	Advance    geom.Pt
	Kern       float64
}

// ShapedRun is a contiguous shaped glyph run using one resolved font face.
type ShapedRun struct {
	Face      FontFace
	Direction TextDirection
	Script    string
	Language  string
	Features  []TextFeature
	Glyphs    []ShapedGlyph
	Advance   geom.Pt
	Bounds    TextBounds
}

// ShapedText is a renderer-independent shaped single-line text layout. Size is
// interpreted as pixels-per-em by this low-level helper.
type ShapedText struct {
	Runs    []ShapedRun
	Glyphs  []ShapedGlyph
	Advance geom.Pt
	Bounds  TextBounds
}

// ShapeText resolves text to font runs and shapes it into positioned glyphs.
func ShapeText(text string, origin geom.Pt, size float64, opts TextShapingOptions) (ShapedText, bool) {
	if text == "" || size <= 0 {
		return ShapedText{}, false
	}
	opts = mergeFontKeyShapingOptions(opts)
	runs, ok := DefaultFontManager().ResolveTextRuns(text, opts.FontKey)
	if !ok {
		return ShapedText{}, false
	}
	return ShapeTextRuns(runs, origin, size, opts)
}

func mergeFontKeyShapingOptions(opts TextShapingOptions) TextShapingOptions {
	props := ParseFontProperties(opts.FontKey)
	if opts.Language == "" && props.Language != "" {
		opts.Language = props.Language
	}
	if len(props.Features) > 0 {
		merged := make([]TextFeature, 0, len(props.Features)+len(opts.Features))
		merged = append(merged, props.Features...)
		merged = append(merged, opts.Features...)
		opts.Features = merged
	}
	return opts
}

// ShapeTextRuns shapes already-resolved font runs.
func ShapeTextRuns(runs []FontRun, origin geom.Pt, size float64, opts TextShapingOptions) (ShapedText, bool) {
	if len(runs) == 0 || size <= 0 {
		return ShapedText{}, false
	}
	if opts.Direction == "" {
		opts.Direction = TextDirectionLTR
	}

	ppem := fixed.Int26_6(math.Round(size * 64))
	if ppem <= 0 {
		return ShapedText{}, false
	}

	var (
		shaped        ShapedText
		penX          = origin.X
		kernEnabled   = openTypeFeatureEnabled(opts, "kern", true)
		haveBounds    bool
		minX, minY    float64
		maxX, maxY    float64
		previous      sfnt.GlyphIndex
		previousFace  string
		havePrevious  bool
		clusterOffset int
		laidOutGlyphs bool
		buf           sfnt.Buffer
	)

	for _, inputRun := range runs {
		faceKey := fontFaceCacheKey(inputRun.Face)
		if inputRun.Text == "" || faceKey == "" {
			clusterOffset += len(inputRun.Text)
			continue
		}
		fontData, err := loadTextPathFontFace(inputRun.Face)
		if err != nil {
			return ShapedText{}, false
		}

		run := ShapedRun{
			Face:      inputRun.Face,
			Direction: opts.Direction,
			Script:    opts.Script,
			Language:  opts.Language,
			Features:  append([]TextFeature(nil), opts.Features...),
		}
		runStartX := penX
		runHaveBounds := false
		var runMinX, runMinY, runMaxX, runMaxY float64

		inputGlyphs, ok := shapeRunInputGlyphs(fontData, &buf, inputRun.Text, clusterOffset)
		if !ok {
			return ShapedText{}, false
		}
		inputGlyphs, ok = applyArabicFormSubstitutions(fontData, &buf, inputRun.Face, inputGlyphs, opts)
		if !ok {
			return ShapedText{}, false
		}
		inputGlyphs = applyGSUBLigatures(inputRun.Face, inputGlyphs, opts)
		reorderDirectionalGlyphs(inputGlyphs, opts.Direction)
		markTable, haveMarkTable := gposMarkToBaseTableForFace(inputRun.Face, opts)
		scale := size / float64(fontData.UnitsPerEm())

		attachCenterX := 0.0
		haveAttachCenter := false
		var previousOrigin geom.Pt
		for _, inputGlyph := range inputGlyphs {
			isMark := isCombiningMark(inputGlyph.Rune)
			kern := 0.0
			if !isMark && kernEnabled && havePrevious && previousFace == faceKey {
				if k, err := fontData.Kern(&buf, previous, inputGlyph.GlyphIndex, ppem, font.HintingNone); err == nil {
					kern = fixedToFloat(k)
					penX += kern
				}
			}

			segments, err := fontData.LoadGlyph(&buf, inputGlyph.GlyphIndex, ppem, nil)
			if err != nil {
				return ShapedText{}, false
			}
			originPt := geom.Pt{X: penX, Y: origin.Y}
			offsetPt := geom.Pt{}
			if isMark {
				if haveMarkTable && havePrevious {
					if offset, ok := markTable.offset(previous, inputGlyph.GlyphIndex); ok {
						offsetPt = geom.Pt{
							X: float64(offset.x) * scale,
							Y: float64(offset.y) * scale,
						}
						originPt = geom.Pt{
							X: previousOrigin.X + offsetPt.X,
							Y: previousOrigin.Y + offsetPt.Y,
						}
					}
				}
				if offsetPt == (geom.Pt{}) && haveAttachCenter && len(segments) > 0 {
					segBounds := segments.Bounds()
					markCenter := fixedToFloat(segBounds.Min.X+segBounds.Max.X) / 2
					originPt.X = attachCenterX - markCenter
					offsetPt = geom.Pt{X: originPt.X - previousOrigin.X, Y: originPt.Y - previousOrigin.Y}
				}
			}
			if len(segments) > 0 {
				bounds := segments.Bounds()
				x0 := originPt.X + fixedToFloat(bounds.Min.X)
				y0 := originPt.Y + fixedToFloat(bounds.Min.Y)
				x1 := originPt.X + fixedToFloat(bounds.Max.X)
				y1 := originPt.Y + fixedToFloat(bounds.Max.Y)
				if !haveBounds {
					minX, minY, maxX, maxY = x0, y0, x1, y1
					haveBounds = true
				} else {
					minX = math.Min(minX, x0)
					minY = math.Min(minY, y0)
					maxX = math.Max(maxX, x1)
					maxY = math.Max(maxY, y1)
				}
				if !runHaveBounds {
					runMinX, runMinY, runMaxX, runMaxY = x0, y0, x1, y1
					runHaveBounds = true
				} else {
					runMinX = math.Min(runMinX, x0)
					runMinY = math.Min(runMinY, y0)
					runMaxX = math.Max(runMaxX, x1)
					runMaxY = math.Max(runMaxY, y1)
				}
			}

			advance, err := fontData.GlyphAdvance(&buf, inputGlyph.GlyphIndex, ppem, font.HintingNone)
			if err != nil {
				return ShapedText{}, false
			}
			advancePt := geom.Pt{X: fixedToFloat(advance)}
			if isMark {
				advancePt = geom.Pt{}
			}
			glyph := ShapedGlyph{
				Rune:       inputGlyph.Rune,
				Cluster:    inputGlyph.Cluster,
				Face:       inputRun.Face,
				GlyphIndex: inputGlyph.GlyphIndex,
				Origin:     originPt,
				Offset:     offsetPt,
				Advance:    advancePt,
				Kern:       kern,
			}
			run.Glyphs = append(run.Glyphs, glyph)
			shaped.Glyphs = append(shaped.Glyphs, glyph)
			if !isMark {
				penX += advancePt.X
				previous = inputGlyph.GlyphIndex
				previousFace = faceKey
				previousOrigin = originPt
				havePrevious = true
				if len(segments) > 0 {
					bounds := segments.Bounds()
					attachCenterX = originPt.X + fixedToFloat(bounds.Min.X+bounds.Max.X)/2
				} else {
					attachCenterX = originPt.X + advancePt.X/2
				}
				haveAttachCenter = true
			}
			laidOutGlyphs = true
		}

		run.Advance = geom.Pt{X: penX - runStartX}
		if runHaveBounds {
			run.Bounds = TextBounds{X: runMinX - origin.X, Y: runMinY - origin.Y, W: runMaxX - runMinX, H: runMaxY - runMinY}
		}
		if len(run.Glyphs) > 0 {
			shaped.Runs = append(shaped.Runs, run)
		}
		clusterOffset += len(inputRun.Text)
	}

	shaped.Advance = geom.Pt{X: penX - origin.X}
	if haveBounds {
		shaped.Bounds = TextBounds{X: minX - origin.X, Y: minY - origin.Y, W: maxX - minX, H: maxY - minY}
	}
	return shaped, laidOutGlyphs || shaped.Advance.X > 0 || shaped.Advance.Y > 0
}

func shapeRunInputGlyphs(fontData *sfnt.Font, buf *sfnt.Buffer, text string, clusterOffset int) ([]shapingGlyph, bool) {
	var (
		out      []shapingGlyph
		runes    []rune
		clusters []int
	)
	for cluster, r := range text {
		runes = append(runes, r)
		clusters = append(clusters, clusterOffset+cluster)
	}

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		cluster := clusters[i]
		if !isCombiningMark(r) {
			j := i + 1
			for j < len(runes) && isCombiningMark(runes[j]) {
				j++
			}
			if j > i+1 {
				if glyph, ok := composedGlyph(fontData, buf, runes[i:j], cluster); ok {
					out = append(out, glyph)
					i = j - 1
					continue
				}
			}
		}

		glyphIndex, err := fontData.GlyphIndex(buf, r)
		if err != nil {
			return nil, false
		}
		if glyphIndex == 0 {
			continue
		}
		out = append(out, shapingGlyph{
			Rune:       r,
			Cluster:    cluster,
			GlyphIndex: glyphIndex,
		})
	}
	return out, true
}

func composedGlyph(fontData *sfnt.Font, buf *sfnt.Buffer, runes []rune, cluster int) (shapingGlyph, bool) {
	normalized := norm.NFC.String(string(runes))
	if utf8.RuneCountInString(normalized) != 1 {
		return shapingGlyph{}, false
	}
	r, _ := utf8.DecodeRuneInString(normalized)
	glyphIndex, err := fontData.GlyphIndex(buf, r)
	if err != nil || glyphIndex == 0 {
		return shapingGlyph{}, false
	}
	return shapingGlyph{
		Rune:       r,
		Cluster:    cluster,
		GlyphIndex: glyphIndex,
	}, true
}

func isCombiningMark(r rune) bool {
	return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || unicode.Is(unicode.Me, r)
}

type arabicForms struct {
	isolated rune
	final    rune
	initial  rune
	medial   rune
}

var arabicPresentationForms = map[rune]arabicForms{
	'\u0627': {isolated: '\ufe8d', final: '\ufe8e'},
	'\u0628': {isolated: '\ufe8f', final: '\ufe90', initial: '\ufe91', medial: '\ufe92'},
	'\u0629': {isolated: '\ufe93', final: '\ufe94'},
	'\u062a': {isolated: '\ufe95', final: '\ufe96', initial: '\ufe97', medial: '\ufe98'},
	'\u062b': {isolated: '\ufe99', final: '\ufe9a', initial: '\ufe9b', medial: '\ufe9c'},
	'\u062c': {isolated: '\ufe9d', final: '\ufe9e', initial: '\ufe9f', medial: '\ufea0'},
	'\u062d': {isolated: '\ufea1', final: '\ufea2', initial: '\ufea3', medial: '\ufea4'},
	'\u062e': {isolated: '\ufea5', final: '\ufea6', initial: '\ufea7', medial: '\ufea8'},
	'\u062f': {isolated: '\ufea9', final: '\ufeaa'},
	'\u0630': {isolated: '\ufeab', final: '\ufeac'},
	'\u0631': {isolated: '\ufead', final: '\ufeae'},
	'\u0632': {isolated: '\ufeaf', final: '\ufeb0'},
	'\u0633': {isolated: '\ufeb1', final: '\ufeb2', initial: '\ufeb3', medial: '\ufeb4'},
	'\u0634': {isolated: '\ufeb5', final: '\ufeb6', initial: '\ufeb7', medial: '\ufeb8'},
	'\u0635': {isolated: '\ufeb9', final: '\ufeba', initial: '\ufebb', medial: '\ufebc'},
	'\u0636': {isolated: '\ufebd', final: '\ufebe', initial: '\ufebf', medial: '\ufec0'},
	'\u0637': {isolated: '\ufec1', final: '\ufec2', initial: '\ufec3', medial: '\ufec4'},
	'\u0638': {isolated: '\ufec5', final: '\ufec6', initial: '\ufec7', medial: '\ufec8'},
	'\u0639': {isolated: '\ufec9', final: '\ufeca', initial: '\ufecb', medial: '\ufecc'},
	'\u063a': {isolated: '\ufecd', final: '\ufece', initial: '\ufecf', medial: '\ufed0'},
	'\u0641': {isolated: '\ufed1', final: '\ufed2', initial: '\ufed3', medial: '\ufed4'},
	'\u0642': {isolated: '\ufed5', final: '\ufed6', initial: '\ufed7', medial: '\ufed8'},
	'\u0643': {isolated: '\ufed9', final: '\ufeda', initial: '\ufedb', medial: '\ufedc'},
	'\u0644': {isolated: '\ufedd', final: '\ufede', initial: '\ufedf', medial: '\ufee0'},
	'\u0645': {isolated: '\ufee1', final: '\ufee2', initial: '\ufee3', medial: '\ufee4'},
	'\u0646': {isolated: '\ufee5', final: '\ufee6', initial: '\ufee7', medial: '\ufee8'},
	'\u0647': {isolated: '\ufee9', final: '\ufeea', initial: '\ufeeb', medial: '\ufeec'},
	'\u0648': {isolated: '\ufeed', final: '\ufeee'},
	'\u0649': {isolated: '\ufeef', final: '\ufef0'},
	'\u064a': {isolated: '\ufef1', final: '\ufef2', initial: '\ufef3', medial: '\ufef4'},
}

func applyArabicFormSubstitutions(fontData *sfnt.Font, buf *sfnt.Buffer, face FontFace, glyphs []shapingGlyph, opts TextShapingOptions) ([]shapingGlyph, bool) {
	out := append([]shapingGlyph(nil), glyphs...)
	runes := make([]rune, len(out))
	for i, glyph := range out {
		runes[i] = glyph.Rune
	}
	for i, glyph := range out {
		_, ok := arabicPresentationForms[glyph.Rune]
		if !ok {
			continue
		}
		presentation, formTag, ok := arabicPresentationFormForContext(runes, i, opts)
		if !ok {
			continue
		}
		if substitution, ok := gsubSingleSubstitutionForFace(face, formTag, glyph.GlyphIndex); ok {
			out[i].GlyphIndex = substitution
			continue
		}
		glyphIndex, err := fontData.GlyphIndex(buf, presentation)
		if err != nil {
			return nil, false
		}
		if glyphIndex != 0 {
			out[i].Rune = presentation
			out[i].GlyphIndex = glyphIndex
		}
	}
	return out, true
}

func arabicPresentationFormForContext(runes []rune, i int, opts TextShapingOptions) (rune, string, bool) {
	r := runes[i]
	forms, ok := arabicPresentationForms[r]
	if !ok {
		return 0, "", false
	}
	joinPrev := arabicJoinsPrevious(runes, i)
	joinNext := arabicJoinsNext(runes, i)
	candidate := forms.isolated
	formTag := "isol"
	switch {
	case joinPrev && joinNext && forms.medial != 0:
		candidate = forms.medial
		formTag = "medi"
	case joinPrev && forms.final != 0:
		candidate = forms.final
		formTag = "fina"
	case joinNext && forms.initial != 0:
		candidate = forms.initial
		formTag = "init"
	}
	if !enabledArabicFormFeature(formTag, opts) {
		return 0, formTag, false
	}
	if candidate == 0 {
		return 0, formTag, false
	}
	return candidate, formTag, true
}

func arabicJoinsPrevious(runes []rune, i int) bool {
	forms, ok := arabicPresentationForms[runes[i]]
	if !ok || forms.final == 0 {
		return false
	}
	for j := i - 1; j >= 0; j-- {
		if isCombiningMark(runes[j]) {
			continue
		}
		prev, ok := arabicPresentationForms[runes[j]]
		return ok && prev.initial != 0
	}
	return false
}

func arabicJoinsNext(runes []rune, i int) bool {
	forms, ok := arabicPresentationForms[runes[i]]
	if !ok || forms.initial == 0 {
		return false
	}
	for j := i + 1; j < len(runes); j++ {
		if isCombiningMark(runes[j]) {
			continue
		}
		next, ok := arabicPresentationForms[runes[j]]
		return ok && next.final != 0
	}
	return false
}

func reorderDirectionalGlyphs(glyphs []shapingGlyph, direction TextDirection) {
	if direction == TextDirectionRTL {
		reverseShapingGlyphs(glyphs)
		return
	}
	for start := 0; start < len(glyphs); {
		if !isRTLGlyph(glyphs[start]) {
			start++
			continue
		}
		end := start + 1
		for end < len(glyphs) && isRTLGlyph(glyphs[end]) {
			end++
		}
		reverseShapingGlyphs(glyphs[start:end])
		start = end
	}
}

func isRTLGlyph(glyph shapingGlyph) bool {
	props, _ := bidi.LookupRune(glyph.Rune)
	switch props.Class() {
	case bidi.R, bidi.AL:
		return true
	default:
		return false
	}
}

func reverseShapingGlyphs(glyphs []shapingGlyph) {
	for i, j := 0, len(glyphs)-1; i < j; i, j = i+1, j-1 {
		glyphs[i], glyphs[j] = glyphs[j], glyphs[i]
	}
}
