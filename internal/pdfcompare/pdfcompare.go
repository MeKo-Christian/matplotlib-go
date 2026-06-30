// Package pdfcompare provides structural parsing, normalization, and diffing
// for PDF documents produced by the matplotlib-go PDF backend.
//
// It is intentionally narrow: it compares indirect objects by object number,
// ignores xref byte offsets, normalizes insignificant token whitespace, and
// decodes FlateDecode streams so content-stream diffs are readable.
package pdfcompare

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Document is a parsed, normalized PDF document.
type Document struct {
	Objects []Object
}

// Object is one indirect PDF object with normalized body text.
type Object struct {
	ID         int
	Generation int
	Body       string
}

// Parse returns a normalized document parsed from raw PDF bytes.
func Parse(src []byte) (*Document, error) {
	if !bytes.HasPrefix(src, []byte("%PDF-")) {
		return nil, fmt.Errorf("pdfcompare: parse: missing PDF header")
	}

	var objects []Object
	pos := 0
	for {
		headerStart, id, gen, bodyStart, ok := nextObjectHeader(src, pos)
		if !ok {
			break
		}
		endRel := bytes.Index(src[bodyStart:], []byte("endobj"))
		if endRel < 0 {
			return nil, fmt.Errorf("pdfcompare: parse: object %d %d has no endobj", id, gen)
		}
		bodyEnd := bodyStart + endRel
		body, err := normalizeObjectBody(src[bodyStart:bodyEnd])
		if err != nil {
			return nil, fmt.Errorf("pdfcompare: parse object %d %d: %w", id, gen, err)
		}
		objects = append(objects, Object{ID: id, Generation: gen, Body: body})
		pos = bodyEnd + len("endobj")
		if pos <= headerStart {
			return nil, fmt.Errorf("pdfcompare: parse: parser made no progress")
		}
	}
	if len(objects) == 0 {
		return nil, fmt.Errorf("pdfcompare: parse: no indirect objects")
	}

	sort.SliceStable(objects, func(i, j int) bool {
		if objects[i].ID != objects[j].ID {
			return objects[i].ID < objects[j].ID
		}
		return objects[i].Generation < objects[j].Generation
	})
	return &Document{Objects: objects}, nil
}

func nextObjectHeader(src []byte, start int) (headerStart, id, gen, bodyStart int, ok bool) {
	for i := start; i < len(src); {
		for i < len(src) && !isDigit(src[i]) {
			i++
		}
		if i >= len(src) {
			return 0, 0, 0, 0, false
		}
		candidate := i
		first, next, ok := readInt(src, i)
		if !ok {
			i++
			continue
		}
		next = skipPDFSpace(src, next)
		second, next, ok := readInt(src, next)
		if !ok {
			i = candidate + 1
			continue
		}
		next = skipPDFSpace(src, next)
		if next+3 <= len(src) && string(src[next:next+3]) == "obj" && (next+3 == len(src) || isPDFSpace(src[next+3])) {
			return candidate, first, second, next + 3, true
		}
		i = candidate + 1
	}
	return 0, 0, 0, 0, false
}

func normalizeObjectBody(body []byte) (string, error) {
	streamIdx := bytes.Index(body, []byte("stream"))
	if streamIdx < 0 {
		return normalizeTokens(body), nil
	}
	dict := body[:streamIdx]
	streamStart := streamIdx + len("stream")
	if streamStart < len(body) && body[streamStart] == '\r' {
		streamStart++
		if streamStart < len(body) && body[streamStart] == '\n' {
			streamStart++
		}
	} else if streamStart < len(body) && body[streamStart] == '\n' {
		streamStart++
	}

	// Prefer the declared /Length to bound the stream: binary stream data (e.g.
	// a FlateDecode-compressed image) can contain the literal bytes
	// "endstream", and scanning for that token would truncate the stream early
	// and corrupt the decode. Fall back to the token scan only when /Length is
	// absent or an indirect reference we can't resolve here.
	var streamData []byte
	if n, ok := streamDeclaredLength(dict); ok && streamStart+n <= len(body) {
		streamData = body[streamStart : streamStart+n]
	} else {
		endStreamIdx := bytes.Index(body[streamStart:], []byte("endstream"))
		if endStreamIdx < 0 {
			return "", fmt.Errorf("stream has no endstream")
		}
		streamData = bytes.TrimRight(body[streamStart:streamStart+endStreamIdx], "\r\n")
	}
	decoded := streamData
	if hasFlateDecode(dict) {
		var err error
		decoded, err = decodeFlate(streamData)
		if err != nil {
			return "", err
		}
	}
	if isImageStream(dict) {
		return normalizeStreamDict(dict) + "\nstream\n" + binaryStreamDigest(decoded) + "\nendstream", nil
	}
	return normalizeStreamDict(dict) + "\nstream\n" + normalizeTokens(decoded) + "\nendstream", nil
}

// streamDeclaredLength returns the direct integer value of the /Length entry in
// a stream dictionary. It deliberately ignores indirect references
// (/Length n g R), returning ok=false so the caller falls back to scanning for
// the endstream token.
func streamDeclaredLength(dict []byte) (int, bool) {
	tokens := pdfTokens(dict)
	for i := 0; i < len(tokens); i++ {
		if tokens[i] != "/Length" || i+1 >= len(tokens) {
			continue
		}
		n, err := strconv.Atoi(tokens[i+1])
		if err != nil {
			return 0, false
		}
		// Indirect reference "n g R" — cannot resolve from the dict alone.
		if i+3 < len(tokens) && tokens[i+3] == "R" {
			if _, err := strconv.Atoi(tokens[i+2]); err == nil {
				return 0, false
			}
		}
		if n < 0 {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

func hasFlateDecode(dict []byte) bool {
	normalized := normalizeTokens(dict)
	return strings.Contains(normalized, "/Filter /FlateDecode") ||
		strings.Contains(normalized, "/Filter [ /FlateDecode ]") ||
		strings.Contains(normalized, "/FlateDecode")
}

func isImageStream(dict []byte) bool {
	normalized := normalizeTokens(dict)
	return strings.Contains(normalized, "/Subtype /Image")
}

func binaryStreamDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("stream-bytes len=%d sha256=%x", len(data), sum)
}

func decodeFlate(data []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode FlateDecode stream: %w", err)
	}
	defer zr.Close()
	out, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("read FlateDecode stream: %w", err)
	}
	return out, nil
}

func normalizeTokens(src []byte) string {
	tokens := pdfTokens(src)
	return strings.Join(tokens, " ")
}

func normalizeStreamDict(src []byte) string {
	tokens := pdfTokens(src)
	out := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "/Length" && i+1 < len(tokens) {
			i++
			continue
		}
		out = append(out, tokens[i])
	}
	return strings.Join(out, " ")
}

func pdfTokens(src []byte) []string {
	tokens := make([]string, 0)
	for i := 0; i < len(src); {
		if isPDFSpace(src[i]) {
			i++
			continue
		}
		if src[i] == '%' {
			for i < len(src) && src[i] != '\n' && src[i] != '\r' {
				i++
			}
			continue
		}
		switch src[i] {
		case '<':
			if i+1 < len(src) && src[i+1] == '<' {
				tokens = append(tokens, "<<")
				i += 2
				continue
			}
			end := i + 1
			for end < len(src) && src[end] != '>' {
				end++
			}
			if end < len(src) {
				tokens = append(tokens, "<"+compactPDFSpace(string(src[i+1:end]))+">")
				i = end + 1
			} else {
				tokens = append(tokens, string(src[i:]))
				i = len(src)
			}
		case '>':
			if i+1 < len(src) && src[i+1] == '>' {
				tokens = append(tokens, ">>")
				i += 2
			} else {
				tokens = append(tokens, ">")
				i++
			}
		case '[', ']', '{', '}':
			tokens = append(tokens, string(src[i]))
			i++
		case '(':
			token, next := readLiteralString(src, i)
			tokens = append(tokens, token)
			i = next
		case '/':
			end := i + 1
			for end < len(src) && !isPDFSpace(src[end]) && !isPDFDelimiter(src[end]) {
				end++
			}
			tokens = append(tokens, string(src[i:end]))
			i = end
		default:
			end := i
			for end < len(src) && !isPDFSpace(src[end]) && !isPDFDelimiter(src[end]) {
				end++
			}
			tokens = append(tokens, string(src[i:end]))
			i = end
		}
	}
	return tokens
}

func readLiteralString(src []byte, start int) (string, int) {
	depth := 0
	for i := start; i < len(src); i++ {
		switch src[i] {
		case '\\':
			i++
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return string(src[start : i+1]), i + 1
			}
		}
	}
	return string(src[start:]), len(src)
}

func compactPDFSpace(s string) string {
	return strings.Join(strings.Fields(s), "")
}

// Equal reports whether two parsed PDF documents are structurally equal.
func Equal(a, b *Document) bool {
	return Diff(a, b) == ""
}

// Diff returns a human-readable description of the first structural
// difference between expected and actual, or the empty string if they match.
func Diff(expected, actual *Document) string {
	if expected == nil && actual == nil {
		return ""
	}
	if expected == nil {
		return "unexpected document"
	}
	if actual == nil {
		return "missing document"
	}
	if len(expected.Objects) != len(actual.Objects) {
		return fmt.Sprintf("object count differs: want %d, got %d", len(expected.Objects), len(actual.Objects))
	}
	for i := range expected.Objects {
		want := expected.Objects[i]
		got := actual.Objects[i]
		if want.ID != got.ID || want.Generation != got.Generation {
			return fmt.Sprintf("object differs at index %d: want %d %d obj, got %d %d obj",
				i, want.ID, want.Generation, got.ID, got.Generation)
		}
		if want.Body != got.Body {
			label := fmt.Sprintf("%d %d obj", want.ID, want.Generation)
			if strings.Contains(want.Body, "\nstream\n") || strings.Contains(got.Body, "\nstream\n") {
				return fmt.Sprintf("%s: stream differs:\n  want: %s\n  got:  %s",
					label, previewQuoted(want.Body), previewQuoted(got.Body))
			}
			return fmt.Sprintf("%s: body differs:\n  want: %s\n  got:  %s",
				label, previewQuoted(want.Body), previewQuoted(got.Body))
		}
	}
	return ""
}

func previewQuoted(s string) string {
	const max = 1200
	if len(s) <= max {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%q... (truncated, %d bytes total)", s[:max], len(s))
}

// ParseAndDiff parses both inputs and returns the structural diff, or the
// empty string when they match.
func ParseAndDiff(expected, actual []byte) (string, error) {
	want, err := Parse(expected)
	if err != nil {
		return "", fmt.Errorf("parse expected: %w", err)
	}
	got, err := Parse(actual)
	if err != nil {
		return "", fmt.Errorf("parse actual: %w", err)
	}
	return Diff(want, got), nil
}

func readInt(src []byte, start int) (int, int, bool) {
	if start >= len(src) || !isDigit(src[start]) {
		return 0, start, false
	}
	end := start
	for end < len(src) && isDigit(src[end]) {
		end++
	}
	v, err := strconv.Atoi(string(src[start:end]))
	if err != nil {
		return 0, start, false
	}
	return v, end, true
}

func skipPDFSpace(src []byte, i int) int {
	for i < len(src) && isPDFSpace(src[i]) {
		i++
	}
	return i
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isPDFSpace(c byte) bool {
	switch c {
	case 0, '\t', '\n', '\f', '\r', ' ':
		return true
	}
	return false
}

func isPDFDelimiter(c byte) bool {
	switch c {
	case '(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	}
	return false
}
