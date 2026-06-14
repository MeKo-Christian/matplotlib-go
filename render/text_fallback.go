package render

import "unicode/utf8"

// FontRun is a contiguous text run resolved to one font face.
type FontRun struct {
	Text string
	Face FontFace
}

// ResolveTextRuns resolves text into contiguous font runs. The first resolved
// face is always the requested face; generic families are used only when that
// face cannot represent a rune.
func (m *FontManager) ResolveTextRuns(text, fontKey string) ([]FontRun, bool) {
	if m == nil || text == "" {
		return nil, false
	}

	primary, ok := m.FindFont(ParseFontProperties(fontKey))
	if !ok && fontKey == "" {
		primary, ok = m.FindFont(ParseFontProperties("DejaVu Sans"))
	}
	if !ok {
		return nil, false
	}
	primaryKey := fontFaceCacheKey(primary)
	if primaryKey == "" {
		return nil, false
	}

	fallbacks := m.fallbackFaces(primary)
	var runs []FontRun
	var current FontFace
	var currentKey string
	var currentText []rune

	flush := func() {
		if len(currentText) == 0 {
			return
		}
		runs = append(runs, FontRun{
			Text: string(currentText),
			Face: current,
		})
		currentText = currentText[:0]
	}

	for text != "" {
		r, n := utf8.DecodeRuneInString(text)
		text = text[n:]

		face := primary
		faceKey := primaryKey
		if !fontFaceSupportsRuneWithKey(primary, primaryKey, r) {
			fallbackFace, fallbackKey := firstFaceSupportingRune(fallbacks, r)
			if fallbackKey == "" {
				face = primary
				faceKey = primaryKey
			} else {
				face = fallbackFace
				faceKey = fallbackKey
			}
		}

		if currentKey != "" && faceKey != currentKey {
			flush()
		}
		current = face
		currentKey = faceKey
		currentText = append(currentText, r)
	}
	flush()

	return runs, len(runs) > 0
}

func (m *FontManager) fallbackFaces(primary FontFace) []FontFace {
	var faces []FontFace
	seen := map[string]struct{}{}
	if key := fontFaceCacheKey(primary); key != "" {
		seen[key] = struct{}{}
	}
	for _, family := range []string{fontFamilySansSerif, fontFamilySerif, fontFamilyMonospace} {
		face, ok := m.FindFont(FontProperties{Families: []string{family}})
		if !ok {
			continue
		}
		key := fontFaceCacheKey(face)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		faces = append(faces, face)
	}
	return faces
}

func firstFaceSupportingRune(faces []FontFace, r rune) (FontFace, string) {
	for _, face := range faces {
		key := fontFaceCacheKey(face)
		if fontFaceSupportsRuneWithKey(face, key, r) {
			return face, key
		}
	}
	return FontFace{}, ""
}
