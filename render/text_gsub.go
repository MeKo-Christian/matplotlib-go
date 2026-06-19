package render

import (
	"encoding/binary"
	"sort"
	"strings"
	"sync"

	"golang.org/x/image/font/sfnt"
)

type shapingGlyph struct {
	Rune       rune
	Cluster    int
	GlyphIndex sfnt.GlyphIndex
}

type gsubLigature struct {
	Glyph      sfnt.GlyphIndex
	Component  []sfnt.GlyphIndex
	Rune       rune
	Components int
}

type gsubLigatureTable struct {
	ByFirst map[sfnt.GlyphIndex][]gsubLigature
}

type gsubSingleSubstitutionTable struct {
	ByGlyph map[sfnt.GlyphIndex]sfnt.GlyphIndex
}

var (
	gsubLigatureCacheMu sync.RWMutex
	gsubLigatureCache   = map[string]gsubLigatureTable{}

	gsubSingleSubstitutionCacheMu sync.RWMutex
	gsubSingleSubstitutionCache   = map[string]gsubSingleSubstitutionTable{}
)

func applyGSUBLigatures(face FontFace, glyphs []shapingGlyph, opts TextShapingOptions) []shapingGlyph {
	if len(glyphs) < 2 {
		return glyphs
	}
	tags := enabledLigatureFeatureTags(opts)
	if len(tags) == 0 {
		return glyphs
	}
	table, ok := gsubLigatureTableForFace(face, tags)
	if !ok || len(table.ByFirst) == 0 {
		return glyphs
	}

	out := make([]shapingGlyph, 0, len(glyphs))
	for i := 0; i < len(glyphs); {
		ligatures := table.ByFirst[glyphs[i].GlyphIndex]
		match := gsubLigature{}
		matchLen := 0
		for _, ligature := range ligatures {
			need := len(ligature.Component) + 1
			if need <= matchLen || i+need > len(glyphs) {
				continue
			}
			matched := true
			for j, component := range ligature.Component {
				if glyphs[i+1+j].GlyphIndex != component {
					matched = false
					break
				}
			}
			if matched {
				match = ligature
				matchLen = need
			}
		}
		if matchLen > 0 {
			r := match.Rune
			if compat, ok := compatibilityLigatureRune(glyphs[i : i+matchLen]); ok {
				r = compat
			} else if r == 0 {
				r = glyphs[i].Rune
			}
			out = append(out, shapingGlyph{
				Rune:       r,
				Cluster:    glyphs[i].Cluster,
				GlyphIndex: match.Glyph,
			})
			i += matchLen
			continue
		}
		out = append(out, glyphs[i])
		i++
	}
	return out
}

func compatibilityLigatureRune(glyphs []shapingGlyph) (rune, bool) {
	var b strings.Builder
	for _, glyph := range glyphs {
		b.WriteRune(glyph.Rune)
	}
	switch b.String() {
	case "ff":
		return '\ufb00', true
	case "fi":
		return '\ufb01', true
	case "fl":
		return '\ufb02', true
	case "ffi":
		return '\ufb03', true
	case "ffl":
		return '\ufb04', true
	case "st":
		return '\ufb06', true
	default:
		return 0, false
	}
}

func gsubSingleSubstitutionForFace(face FontFace, tag string, glyph sfnt.GlyphIndex) (sfnt.GlyphIndex, bool) {
	table, ok := gsubSingleSubstitutionTableForFace(face, []string{tag})
	if !ok {
		return 0, false
	}
	substitution, ok := table.ByGlyph[glyph]
	return substitution, ok && substitution != 0
}

func enabledLigatureFeatureTags(opts TextShapingOptions) []string {
	var tags []string
	for _, tag := range []string{"liga", "clig"} {
		if openTypeFeatureEnabled(opts, tag, true) {
			tags = append(tags, tag)
		}
	}
	return tags
}

func openTypeFeatureEnabled(opts TextShapingOptions, tag string, defaultValue bool) bool {
	enabled := defaultValue
	for _, feature := range opts.Features {
		if normalizeOpenTypeTag(feature.Tag) == tag {
			enabled = feature.Value != 0
		}
	}
	return enabled
}

func normalizeOpenTypeTag(tag string) string {
	tag = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(tag, "+"), "-"))
	if len(tag) != 4 {
		return ""
	}
	return tag
}

func gsubLigatureTableForFace(face FontFace, tags []string) (gsubLigatureTable, bool) {
	key := fontFaceCacheKey(face) + "|" + strings.Join(tags, ",")
	if key == "|" {
		return gsubLigatureTable{}, false
	}
	gsubLigatureCacheMu.RLock()
	if cached, ok := gsubLigatureCache[key]; ok {
		gsubLigatureCacheMu.RUnlock()
		return cached, len(cached.ByFirst) > 0
	}
	gsubLigatureCacheMu.RUnlock()

	data, err := loadFontFaceData(face)
	if err != nil {
		return gsubLigatureTable{}, false
	}
	table := parseGSUBLigatureTable(data, tags)

	gsubLigatureCacheMu.Lock()
	gsubLigatureCache[key] = table
	gsubLigatureCacheMu.Unlock()
	return table, len(table.ByFirst) > 0
}

func gsubSingleSubstitutionTableForFace(face FontFace, tags []string) (gsubSingleSubstitutionTable, bool) {
	key := fontFaceCacheKey(face) + "|" + strings.Join(tags, ",")
	if key == "|" {
		return gsubSingleSubstitutionTable{}, false
	}
	gsubSingleSubstitutionCacheMu.RLock()
	if cached, ok := gsubSingleSubstitutionCache[key]; ok {
		gsubSingleSubstitutionCacheMu.RUnlock()
		return cached, len(cached.ByGlyph) > 0
	}
	gsubSingleSubstitutionCacheMu.RUnlock()

	data, err := loadFontFaceData(face)
	if err != nil {
		return gsubSingleSubstitutionTable{}, false
	}
	table := parseGSUBSingleSubstitutionTable(data, tags)

	gsubSingleSubstitutionCacheMu.Lock()
	gsubSingleSubstitutionCache[key] = table
	gsubSingleSubstitutionCacheMu.Unlock()
	return table, len(table.ByGlyph) > 0
}

func parseGSUBLigatureTable(fontData []byte, tags []string) gsubLigatureTable {
	gsub, ok := sfntTable(fontData, "GSUB")
	if !ok || len(gsub) < 10 {
		return gsubLigatureTable{}
	}
	featureListOffset := int(be16(gsub, 6))
	lookupListOffset := int(be16(gsub, 8))
	if featureListOffset <= 0 || lookupListOffset <= 0 || featureListOffset >= len(gsub) || lookupListOffset >= len(gsub) {
		return gsubLigatureTable{}
	}

	wanted := map[string]bool{}
	for _, tag := range tags {
		wanted[tag] = true
	}
	lookupIndices := gsubFeatureLookupIndices(gsub[featureListOffset:], wanted)
	if len(lookupIndices) == 0 {
		return gsubLigatureTable{}
	}

	table := gsubLigatureTable{ByFirst: map[sfnt.GlyphIndex][]gsubLigature{}}
	parseGSUBLigatureLookups(gsub, lookupListOffset, lookupIndices, table.ByFirst)
	for first := range table.ByFirst {
		sort.SliceStable(table.ByFirst[first], func(i, j int) bool {
			return table.ByFirst[first][i].Components > table.ByFirst[first][j].Components
		})
	}
	return table
}

func parseGSUBSingleSubstitutionTable(fontData []byte, tags []string) gsubSingleSubstitutionTable {
	gsub, ok := sfntTable(fontData, "GSUB")
	if !ok || len(gsub) < 10 {
		return gsubSingleSubstitutionTable{}
	}
	featureListOffset := int(be16(gsub, 6))
	lookupListOffset := int(be16(gsub, 8))
	if featureListOffset <= 0 || lookupListOffset <= 0 || featureListOffset >= len(gsub) || lookupListOffset >= len(gsub) {
		return gsubSingleSubstitutionTable{}
	}

	wanted := map[string]bool{}
	for _, tag := range tags {
		wanted[normalizeOpenTypeTag(tag)] = true
	}
	lookupIndices := gsubFeatureLookupIndices(gsub[featureListOffset:], wanted)
	if len(lookupIndices) == 0 {
		return gsubSingleSubstitutionTable{}
	}

	table := gsubSingleSubstitutionTable{ByGlyph: map[sfnt.GlyphIndex]sfnt.GlyphIndex{}}
	parseGSUBSingleSubstitutionLookups(gsub, lookupListOffset, lookupIndices, table.ByGlyph)
	return table
}

func sfntTable(data []byte, tag string) ([]byte, bool) {
	if len(data) < 12 || len(tag) != 4 {
		return nil, false
	}
	numTables := int(be16(data, 4))
	for i := 0; i < numTables; i++ {
		entry := 12 + i*16
		if entry+16 > len(data) {
			return nil, false
		}
		if string(data[entry:entry+4]) != tag {
			continue
		}
		offset := int(be32(data, entry+8))
		length := int(be32(data, entry+12))
		if offset < 0 || length < 0 || offset+length > len(data) {
			return nil, false
		}
		return data[offset : offset+length], true
	}
	return nil, false
}

func gsubFeatureLookupIndices(featureList []byte, wanted map[string]bool) []uint16 {
	if len(featureList) < 2 {
		return nil
	}
	count := int(be16(featureList, 0))
	seen := map[uint16]bool{}
	var out []uint16
	for i := 0; i < count; i++ {
		record := 2 + i*6
		if record+6 > len(featureList) {
			return out
		}
		if !wanted[string(featureList[record:record+4])] {
			continue
		}
		featureOffset := int(be16(featureList, record+4))
		if featureOffset+4 > len(featureList) {
			continue
		}
		lookupCount := int(be16(featureList, featureOffset+2))
		for j := 0; j < lookupCount; j++ {
			idxOffset := featureOffset + 4 + j*2
			if idxOffset+2 > len(featureList) {
				break
			}
			idx := be16(featureList, idxOffset)
			if !seen[idx] {
				seen[idx] = true
				out = append(out, idx)
			}
		}
	}
	return out
}

func parseGSUBLigatureLookups(gsub []byte, lookupListOffset int, lookupIndices []uint16, byFirst map[sfnt.GlyphIndex][]gsubLigature) {
	if lookupListOffset+2 > len(gsub) {
		return
	}
	lookupCount := int(be16(gsub, lookupListOffset))
	for _, lookupIndex := range lookupIndices {
		if int(lookupIndex) >= lookupCount {
			continue
		}
		lookupOffsetOffset := lookupListOffset + 2 + int(lookupIndex)*2
		if lookupOffsetOffset+2 > len(gsub) {
			continue
		}
		lookupOffset := lookupListOffset + int(be16(gsub, lookupOffsetOffset))
		if lookupOffset+6 > len(gsub) || be16(gsub, lookupOffset) != 4 {
			continue
		}
		subtableCount := int(be16(gsub, lookupOffset+4))
		for i := 0; i < subtableCount; i++ {
			subOffsetOffset := lookupOffset + 6 + i*2
			if subOffsetOffset+2 > len(gsub) {
				break
			}
			subtableOffset := lookupOffset + int(be16(gsub, subOffsetOffset))
			parseGSUBLigatureSubtable(gsub, subtableOffset, byFirst)
		}
	}
}

func parseGSUBSingleSubstitutionLookups(gsub []byte, lookupListOffset int, lookupIndices []uint16, byGlyph map[sfnt.GlyphIndex]sfnt.GlyphIndex) {
	if lookupListOffset+2 > len(gsub) {
		return
	}
	lookupCount := int(be16(gsub, lookupListOffset))
	for _, lookupIndex := range lookupIndices {
		if int(lookupIndex) >= lookupCount {
			continue
		}
		lookupOffsetOffset := lookupListOffset + 2 + int(lookupIndex)*2
		if lookupOffsetOffset+2 > len(gsub) {
			continue
		}
		lookupOffset := lookupListOffset + int(be16(gsub, lookupOffsetOffset))
		if lookupOffset+6 > len(gsub) {
			continue
		}
		lookupType := be16(gsub, lookupOffset)
		subtableCount := int(be16(gsub, lookupOffset+4))
		for i := 0; i < subtableCount; i++ {
			subOffsetOffset := lookupOffset + 6 + i*2
			if subOffsetOffset+2 > len(gsub) {
				break
			}
			subtableOffset := lookupOffset + int(be16(gsub, subOffsetOffset))
			switch lookupType {
			case 1:
				parseGSUBSingleSubstitutionSubtable(gsub, subtableOffset, byGlyph)
			case 7:
				parseGSUBExtensionSubstitutionSubtable(gsub, subtableOffset, 1, byGlyph)
			}
		}
	}
}

func parseGSUBExtensionSubstitutionSubtable(gsub []byte, offset int, wantedType uint16, byGlyph map[sfnt.GlyphIndex]sfnt.GlyphIndex) {
	if offset+8 > len(gsub) || be16(gsub, offset) != 1 || be16(gsub, offset+2) != wantedType {
		return
	}
	extensionOffset := offset + int(be32(gsub, offset+4))
	if extensionOffset <= offset || extensionOffset >= len(gsub) {
		return
	}
	parseGSUBSingleSubstitutionSubtable(gsub, extensionOffset, byGlyph)
}

func parseGSUBSingleSubstitutionSubtable(gsub []byte, offset int, byGlyph map[sfnt.GlyphIndex]sfnt.GlyphIndex) {
	if offset+6 > len(gsub) {
		return
	}
	coverage := parseCoverageGlyphs(gsub, offset+int(be16(gsub, offset+2)))
	if len(coverage) == 0 {
		return
	}
	switch be16(gsub, offset) {
	case 1:
		delta := int(beInt16(gsub, offset+4))
		for _, glyph := range coverage {
			byGlyph[glyph] = sfnt.GlyphIndex(uint16(int(glyph) + delta))
		}
	case 2:
		count := int(be16(gsub, offset+4))
		if offset+6+count*2 > len(gsub) {
			return
		}
		for i := 0; i < count && i < len(coverage); i++ {
			byGlyph[coverage[i]] = sfnt.GlyphIndex(be16(gsub, offset+6+i*2))
		}
	}
}

func parseGSUBLigatureSubtable(gsub []byte, offset int, byFirst map[sfnt.GlyphIndex][]gsubLigature) {
	if offset+6 > len(gsub) || be16(gsub, offset) != 1 {
		return
	}
	coverageOffset := offset + int(be16(gsub, offset+2))
	ligatureSetCount := int(be16(gsub, offset+4))
	coverage := parseCoverageGlyphs(gsub, coverageOffset)
	if len(coverage) == 0 {
		return
	}
	for i := 0; i < ligatureSetCount && i < len(coverage); i++ {
		setOffsetOffset := offset + 6 + i*2
		if setOffsetOffset+2 > len(gsub) {
			break
		}
		first := coverage[i]
		setOffset := offset + int(be16(gsub, setOffsetOffset))
		parseGSUBLigatureSet(gsub, setOffset, first, byFirst)
	}
}

func parseGSUBLigatureSet(gsub []byte, offset int, first sfnt.GlyphIndex, byFirst map[sfnt.GlyphIndex][]gsubLigature) {
	if offset+2 > len(gsub) {
		return
	}
	count := int(be16(gsub, offset))
	for i := 0; i < count; i++ {
		ligOffsetOffset := offset + 2 + i*2
		if ligOffsetOffset+2 > len(gsub) {
			break
		}
		ligOffset := offset + int(be16(gsub, ligOffsetOffset))
		if ligOffset+4 > len(gsub) {
			continue
		}
		componentCount := int(be16(gsub, ligOffset+2))
		if componentCount < 2 || ligOffset+4+(componentCount-1)*2 > len(gsub) {
			continue
		}
		ligature := gsubLigature{
			Glyph:      sfnt.GlyphIndex(be16(gsub, ligOffset)),
			Components: componentCount,
		}
		for j := 0; j < componentCount-1; j++ {
			ligature.Component = append(ligature.Component, sfnt.GlyphIndex(be16(gsub, ligOffset+4+j*2)))
		}
		byFirst[first] = append(byFirst[first], ligature)
	}
}

func parseCoverageGlyphs(table []byte, offset int) []sfnt.GlyphIndex {
	if offset+4 > len(table) {
		return nil
	}
	switch be16(table, offset) {
	case 1:
		count := int(be16(table, offset+2))
		if offset+4+count*2 > len(table) {
			return nil
		}
		out := make([]sfnt.GlyphIndex, 0, count)
		for i := 0; i < count; i++ {
			out = append(out, sfnt.GlyphIndex(be16(table, offset+4+i*2)))
		}
		return out
	case 2:
		count := int(be16(table, offset+2))
		out := []sfnt.GlyphIndex{}
		for i := 0; i < count; i++ {
			rangeOffset := offset + 4 + i*6
			if rangeOffset+6 > len(table) {
				return out
			}
			start := be16(table, rangeOffset)
			end := be16(table, rangeOffset+2)
			for glyph := start; glyph <= end; glyph++ {
				out = append(out, sfnt.GlyphIndex(glyph))
				if glyph == 0xffff {
					break
				}
			}
		}
		return out
	default:
		return nil
	}
}

func be16(data []byte, offset int) uint16 {
	if offset < 0 || offset+2 > len(data) {
		return 0
	}
	return binary.BigEndian.Uint16(data[offset : offset+2])
}

func be32(data []byte, offset int) uint32 {
	if offset < 0 || offset+4 > len(data) {
		return 0
	}
	return binary.BigEndian.Uint32(data[offset : offset+4])
}
