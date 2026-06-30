package render

import (
	"errors"
	"math"
	"os"
	"sync"
	"unicode/utf8"

	"github.com/cwbudde/matplotlib-go/geom"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

var (
	textPathFontCacheMu        sync.RWMutex
	textPathFontCache          = map[string]*sfnt.Font{}
	fontFaceRuneSupportCacheMu sync.RWMutex
	fontFaceRuneSupportCache   = map[fontFaceRuneSupportKey]bool{}
	glyphReverseCmapMu         sync.RWMutex
	glyphReverseCmapCache      = map[string]map[uint32]rune{}
)

type fontFaceRuneSupportKey struct {
	faceKey string
	r       rune
}

// TextPath converts text to filled glyph outline paths at a baseline origin.
//
// Outlines are emitted in y-up display space: a glyph's ascenders sit at larger
// Y than the baseline, matching the rest of the display coordinate system. The
// sfnt loader reports glyph vectors y-down (ascenders negative), so sfntPoint
// negates the font-space Y. Backends that own a device y-inversion (AGG, SVG,
// gobasic) hand the result straight to Path; natively y-up backends (PDF, PS,
// PGF) emit it verbatim. No backend needs to reflect glyphs about the baseline.
func TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if text == "" || size <= 0 {
		return geom.Path{}, false
	}

	manager := DefaultFontManager()
	runs, ok := manager.ResolveTextRuns(text, fontKey)
	if !ok {
		return geom.Path{}, false
	}

	path, ok := textRunsPath(runs, origin, size)
	if !ok || !path.Validate() || len(path.C) == 0 {
		return geom.Path{}, false
	}
	return path, true
}

// GlyphIDToRune resolves a font glyph index (the value carried in
// render.Glyph.ID) back to the Unicode scalar that maps to it under the font's
// cmap, for the font selected by fontKey. It returns false when the font cannot
// be loaded or no rune maps to the glyph (e.g. a ligature or a .notdef index).
//
// render.Glyph.ID is a glyph index, not a code point: backends that can only
// draw strings (SVG/PDF/PS/PGF and the AGG legacy fallback) must reverse the
// cmap to recover a drawable rune rather than casting the index directly with
// rune(id), which yields a wrong character for any non-trivial font. The reverse
// map is built once per font (over the Basic Multilingual Plane plus the
// supplementary range matplotlib's bundled fonts use) and cached.
func GlyphIDToRune(fontKey string, gid uint32) (rune, bool) {
	if gid == 0 {
		return 0, false
	}
	face, ok := DefaultFontManager().FindFont(ParseFontProperties(fontKey))
	if !ok {
		return 0, false
	}
	key := fontFaceCacheKey(face)
	if key == "" {
		return 0, false
	}

	glyphReverseCmapMu.RLock()
	table, ok := glyphReverseCmapCache[key]
	glyphReverseCmapMu.RUnlock()
	if !ok {
		table = buildGlyphReverseCmap(face, key)
	}
	r, ok := table[gid]
	return r, ok
}

// buildGlyphReverseCmap parses the font and inverts its cmap (rune -> glyph
// index) into a glyph index -> rune table, caching the result by face key. The
// first rune mapping to a given glyph wins, which keeps the lowest code point
// for glyphs shared by several runes.
func buildGlyphReverseCmap(face FontFace, key string) map[uint32]rune {
	table := map[uint32]rune{}
	if fontData, err := loadTextPathFontFaceByKey(face, key); err == nil {
		var buf sfnt.Buffer
		// 0x0000–0xFFFF covers the BMP; 0x10000–0x1FFFF covers the supplementary
		// math/symbol planes the bundled DejaVu/STIX fonts populate.
		for r := rune(0x20); r <= 0x1FFFF; r++ {
			idx, err := fontData.GlyphIndex(&buf, r)
			if err != nil || idx == 0 {
				continue
			}
			if _, seen := table[uint32(idx)]; !seen {
				table[uint32(idx)] = r
			}
		}
	}
	glyphReverseCmapMu.Lock()
	glyphReverseCmapCache[key] = table
	glyphReverseCmapMu.Unlock()
	return table
}

func loadTextPathFontFace(face FontFace) (*sfnt.Font, error) {
	key := fontFaceCacheKey(face)
	if key == "" {
		return nil, errors.New("render: font face has no path or embedded data")
	}
	return loadTextPathFontFaceByKey(face, key)
}

func loadTextPathFontFaceByKey(face FontFace, key string) (*sfnt.Font, error) {
	textPathFontCacheMu.RLock()
	if cached, ok := textPathFontCache[key]; ok {
		textPathFontCacheMu.RUnlock()
		return cached, nil
	}
	textPathFontCacheMu.RUnlock()

	data, err := loadFontFaceData(face)
	if err != nil {
		return nil, err
	}
	parsed, err := sfnt.Parse(data)
	if err != nil {
		return nil, err
	}

	textPathFontCacheMu.Lock()
	textPathFontCache[key] = parsed
	textPathFontCacheMu.Unlock()
	return parsed, nil
}

func loadFontFaceData(face FontFace) ([]byte, error) {
	if len(face.Data) > 0 {
		return face.Data, nil
	}
	if face.Path != "" {
		return os.ReadFile(face.Path)
	}
	return nil, errors.New("render: font face has no path or embedded data")
}

func textRunsPath(runs []FontRun, origin geom.Pt, size float64) (geom.Path, bool) {
	var out geom.Path
	ok := false
	shaped, haveShape := ShapeTextRuns(runs, origin, size, TextShapingOptions{})
	if !haveShape {
		return geom.Path{}, false
	}
	ppem := fixed.Int26_6(math.Round(size * 64))
	var buf sfnt.Buffer
	for _, glyph := range shaped.Glyphs {
		fontData, err := loadTextPathFontFace(glyph.Face)
		if err != nil {
			return geom.Path{}, false
		}
		segments, err := fontData.LoadGlyph(&buf, glyph.GlyphIndex, ppem, nil)
		if err != nil {
			return geom.Path{}, false
		}
		before := len(out.C)
		appendGlyphSegments(&out, segments, glyph.Origin)
		ok = ok || len(out.C) > before
	}
	return out, ok
}

func fontFaceSupportsRune(face FontFace, r rune) bool {
	key := fontFaceCacheKey(face)
	if key == "" || r == utf8.RuneError {
		return false
	}
	return fontFaceSupportsRuneWithKey(face, key, r)
}

func fontFaceSupportsRuneWithKey(face FontFace, faceKey string, r rune) bool {
	if faceKey == "" || r == utf8.RuneError {
		return false
	}
	cacheKey := fontFaceRuneSupportKey{faceKey: faceKey, r: r}
	fontFaceRuneSupportCacheMu.RLock()
	if supported, ok := fontFaceRuneSupportCache[cacheKey]; ok {
		fontFaceRuneSupportCacheMu.RUnlock()
		return supported
	}
	fontFaceRuneSupportCacheMu.RUnlock()

	fontData, err := loadTextPathFontFaceByKey(face, faceKey)
	if err != nil {
		return false
	}
	var buf sfnt.Buffer
	glyphIndex, err := fontData.GlyphIndex(&buf, r)
	supported := err == nil && glyphIndex != 0

	fontFaceRuneSupportCacheMu.Lock()
	fontFaceRuneSupportCache[cacheKey] = supported
	fontFaceRuneSupportCacheMu.Unlock()
	return supported
}

func appendGlyphSegments(path *geom.Path, segments sfnt.Segments, origin geom.Pt) {
	for _, segment := range segments {
		switch segment.Op {
		case sfnt.SegmentOpMoveTo:
			path.MoveTo(sfntPoint(segment.Args[0], origin))
		case sfnt.SegmentOpLineTo:
			path.LineTo(sfntPoint(segment.Args[0], origin))
		case sfnt.SegmentOpQuadTo:
			path.QuadTo(
				sfntPoint(segment.Args[0], origin),
				sfntPoint(segment.Args[1], origin),
			)
		case sfnt.SegmentOpCubeTo:
			path.CubicTo(
				sfntPoint(segment.Args[0], origin),
				sfntPoint(segment.Args[1], origin),
				sfntPoint(segment.Args[2], origin),
			)
		}
	}
}

func sfntPoint(p fixed.Point26_6, origin geom.Pt) geom.Pt {
	// sfnt reports glyph vectors y-down (it negates the font's native y-up
	// units). Negate again so the outline is y-up display space: ascenders sit
	// above the baseline at larger Y.
	return geom.Pt{
		X: origin.X + fixedToFloat(p.X),
		Y: origin.Y - fixedToFloat(p.Y),
	}
}

func fixedToFloat(v fixed.Int26_6) float64 {
	return float64(v) / 64
}
