package svg

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cwbudde/matplotlib-go/render"
)

func (r *Renderer) SaveSVG(path string) error {
	return r.SaveSVGWithOptions(path, r.options)
}

func (r *Renderer) SaveSVGWithOptions(path string, opts render.SVGOptions) error {
	if path == "" {
		return errors.New("svg: path is required")
	}
	r.SetSVGOptions(opts)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(r.renderSVG())
	return err
}

func (r *Renderer) renderSVG() string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("\n")
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"`+"\n"+`width="%spt" height="%spt" viewBox="0 0 %s %s"`+"\n"+`preserveAspectRatio="xMidYMid meet">`+"\n",
		formatFloat(r.width),
		formatFloat(r.height),
		formatFloat(r.width),
		formatFloat(r.height))
	writeMetadata(&b, r.options)

	if len(r.clipOrder) > 0 || len(r.markerOrder) > 0 || len(r.pathOrder) > 0 || len(r.hatchOrder) > 0 || len(r.gradientOrder) > 0 || len(r.patternFillOrder) > 0 || len(r.fontFaceOrder) > 0 || len(r.filterOrder) > 0 {
		b.WriteString("  <defs>\n")
		if len(r.fontFaceOrder) > 0 {
			b.WriteString("    <style type=\"text/css\"><![CDATA[\n")
			for _, face := range r.fontFaceOrder {
				b.WriteString("      @font-face { font-family: \"")
				b.WriteString(face.family)
				b.WriteString("\"; src: url(\"data:")
				b.WriteString(face.mime)
				b.WriteString(";base64,")
				b.WriteString(face.data)
				b.WriteString("\") format(\"")
				b.WriteString(face.format)
				b.WriteString("\"); }\n")
			}
			b.WriteString("    ]]></style>\n")
		}
		for _, filter := range r.filterOrder {
			writeFilterDef(&b, filter)
		}
		for _, clip := range r.clipOrder {
			b.WriteString("    <clipPath id=\"" + clip.id + "\" clipPathUnits=\"userSpaceOnUse\">")
			if clip.rect != nil {
				w := clip.rect.W()
				h := clip.rect.H()
				b.WriteString("<rect x=\"")
				b.WriteString(formatFloat(clip.rect.Min.X))
				b.WriteString(`" y="`)
				b.WriteString(formatFloat(r.flipY(clip.rect.Max.Y)))
				b.WriteString(`" width="`)
				b.WriteString(formatFloat(w))
				b.WriteString(`" height="`)
				b.WriteString(formatFloat(h))
				b.WriteString(`" />`)
			} else {
				b.WriteString(`<path d="`)
				b.WriteString(clip.path)
				if clip.transform != "" {
					b.WriteString(`" transform="`)
					b.WriteString(clip.transform)
				}
				b.WriteString(`" />`)
			}
			b.WriteString("</clipPath>\n")
		}
		for _, hatch := range r.hatchOrder {
			writeHatchDef(&b, hatch)
		}
		for i := range r.gradientOrder {
			writeGradientDef(&b, &r.gradientOrder[i])
		}
		for i := range r.patternFillOrder {
			writePatternFillDef(&b, &r.patternFillOrder[i])
		}
		for _, m := range r.markerOrder {
			b.WriteString(`    <path id="`)
			b.WriteString(m.id)
			b.WriteString(`" d="`)
			b.WriteString(m.data)
			b.WriteString(`" />` + "\n")
		}
		for _, p := range r.pathOrder {
			b.WriteString(`    <path id="`)
			b.WriteString(p.id)
			b.WriteString(`" d="`)
			b.WriteString(p.data)
			b.WriteString(`" />` + "\n")
		}
		b.WriteString("  </defs>\n")
	}

	bgColor, bgAlpha := colorToStyle(r.background)
	b.WriteString("  <rect x=\"0\" y=\"0\" width=\"100%\" height=\"100%\" ")
	if bgAlpha <= 0 {
		b.WriteString(`fill="none" />`)
		b.WriteString("\n")
	} else {
		b.WriteString(`fill="`)
		b.WriteString(bgColor)
		b.WriteString(`"`)
		if bgAlpha < 1 {
			b.WriteString(` fill-opacity="`)
			b.WriteString(formatFloat(bgAlpha))
			b.WriteString(`"`)
		}
		b.WriteString(" />\n")
	}

	for _, node := range r.nodes {
		b.WriteString("  ")
		// url wraps outermost as a hyperlink; gid wraps the element in an
		// identified group; clip/filter groups wrap closest to the content.
		if node.url != "" {
			b.WriteString(`<a xlink:href="`)
			b.WriteString(escapeText(node.url))
			b.WriteString(`">`)
		}
		if node.gid != "" {
			b.WriteString(`<g id="`)
			b.WriteString(escapeText(node.gid))
			b.WriteString(`">`)
		}
		for _, id := range node.clipIDs {
			b.WriteString("<g clip-path=\"url(#")
			b.WriteString(id)
			b.WriteString(")\">")
		}
		for _, id := range node.filterIDs {
			b.WriteString("<g filter=\"url(#")
			b.WriteString(id)
			b.WriteString(")\">")
		}
		b.WriteString(node.content)
		for range node.filterIDs {
			b.WriteString("</g>")
		}
		for range node.clipIDs {
			b.WriteString("</g>")
		}
		if node.gid != "" {
			b.WriteString("</g>")
		}
		if node.url != "" {
			b.WriteString("</a>")
		}
		b.WriteString("\n")
	}

	b.WriteString("</svg>\n")
	return b.String()
}

func writeMetadata(b *strings.Builder, opts render.SVGOptions) {
	metadata := normalizeSVGOptions(opts).Metadata
	if epoch := strings.TrimSpace(os.Getenv("SOURCE_DATE_EPOCH")); epoch != "" {
		if _, ok := metadata["Date"]; !ok {
			if ts, err := strconv.ParseInt(epoch, 10, 64); err == nil {
				if metadata == nil {
					metadata = map[string]string{}
				}
				metadata["Date"] = time.Unix(ts, 0).UTC().Format(time.RFC3339)
			}
		}
	}
	if len(metadata) == 0 {
		b.WriteString("  <metadata></metadata>\n")
		return
	}
	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	b.WriteString("  <metadata>\n")
	for _, key := range keys {
		b.WriteString("    <meta")
		writeAttr(b, "name", key)
		writeAttr(b, "content", metadata[key])
		b.WriteString(" />\n")
	}
	b.WriteString("  </metadata>\n")
}
