package ps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SavePS writes the finalized document to path.
func (r *Renderer) SavePS(path string) error {
	if len(r.document) == 0 {
		return errors.New("ps: SavePS called before End")
	}
	ext := strings.ToLower(filepath.Ext(path))
	data := r.document
	if ext == ".eps" {
		data = buildDocument(r.width, r.height, r.content.String(), true)
	}
	return os.WriteFile(path, data, 0o644)
}

func buildDocument(width, height int, content string, eps bool) []byte {
	var b strings.Builder
	if eps {
		b.WriteString("%!PS-Adobe-3.0 EPSF-3.0\n")
	} else {
		b.WriteString("%!PS-Adobe-3.0\n")
	}
	b.WriteString("%%Creator: matplotlib-go\n")
	fmt.Fprintf(&b, "%%%%BoundingBox: 0 0 %d %d\n", width, height)
	b.WriteString("%%LanguageLevel: 2\n")
	if !eps {
		b.WriteString("%%Pages: 1\n")
	}
	b.WriteString("%%EndComments\n")
	if !eps {
		b.WriteString("%%Page: 1 1\n")
	}
	b.WriteString(content)
	if !eps {
		b.WriteString("showpage\n")
	}
	b.WriteString("%%EOF\n")
	return []byte(b.String())
}
