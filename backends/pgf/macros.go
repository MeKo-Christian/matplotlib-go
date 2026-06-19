package pgf

import (
	"fmt"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// DrawMarkers renders one marker path at many display-space offsets using a
// reusable PGF macro for identical marker geometry and paint.
func (r *Renderer) DrawMarkers(batch render.MarkerBatch) bool {
	if rr := r.activeRaster(); rr != nil {
		if markers, ok := rr.(render.MarkerDrawer); ok {
			return markers.DrawMarkers(batch)
		}
		return false
	}
	if !r.began || len(batch.Marker.C) == 0 || len(batch.Items) == 0 || !batch.Marker.Validate() {
		return false
	}
	emitted := false
	for i := range batch.Items {
		item := batch.Items[i]
		if !paintVisible(&item.Paint) {
			continue
		}
		name := r.registerPathMacro("M", batch.Marker, &item.Paint)
		if name == "" {
			continue
		}
		t := normalizedAffine(item.Transform)
		t.E += item.Offset.X
		t.F += item.Offset.Y
		r.content.WriteString("\\pgfscope\n")
		writeTransform(&r.content, t)
		fmt.Fprintf(&r.content, "\\csname %s\\endcsname\n", name)
		r.content.WriteString("\\endpgfscope\n")
		emitted = true
	}
	return emitted
}

// DrawPathCollection renders display-space paths through reusable PGF macros
// keyed by path geometry and paint.
func (r *Renderer) DrawPathCollection(batch render.PathCollectionBatch) bool {
	if rr := r.activeRaster(); rr != nil {
		if paths, ok := rr.(render.PathCollectionDrawer); ok {
			return paths.DrawPathCollection(batch)
		}
		return false
	}
	if !r.began || len(batch.Items) == 0 {
		return false
	}
	emitted := false
	for i := range batch.Items {
		item := batch.Items[i]
		if !item.Path.Validate() || len(item.Path.C) == 0 {
			continue
		}
		paint := item.Paint
		if item.Hatch != "" {
			paint.Hatch = item.Hatch
			paint.HatchColor = item.HatchColor
			paint.HatchLineWidth = item.HatchWidth
			paint.HatchSpacing = item.HatchSpacing
		}
		if !paintVisible(&paint) {
			continue
		}
		name := r.registerPathMacro("P", item.Path, &paint)
		if name == "" {
			continue
		}
		fmt.Fprintf(&r.content, "\\csname %s\\endcsname\n", name)
		emitted = true
	}
	return emitted
}

func (r *Renderer) registerPathMacro(prefix string, path geom.Path, paint *render.Paint) string {
	key := pathMacroKey(prefix, path, paint)
	if name, ok := r.pathIDs[key]; ok {
		return name
	}
	var body strings.Builder
	if !r.writePathPaintOps(&body, path, paint) {
		return ""
	}
	name := fmt.Sprintf("mplgpgf%s%d", prefix, len(r.pathIDs)+1)
	r.pathIDs[key] = name
	fmt.Fprintf(&r.content, "\\expandafter\\def\\csname %s\\endcsname{%%\n%s}\n", name, body.String())
	return name
}

func pathMacroKey(prefix string, path geom.Path, paint *render.Paint) string {
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteByte('|')
	for _, cmd := range path.C {
		fmt.Fprintf(&b, "%d;", cmd)
	}
	b.WriteByte('|')
	for _, pt := range path.V {
		b.WriteString(shortFloat(pt.X))
		b.WriteByte(',')
		b.WriteString(shortFloat(pt.Y))
		b.WriteByte(';')
	}
	b.WriteByte('|')
	writePaintKey(&b, paint)
	return b.String()
}

func writePaintKey(b *strings.Builder, paint *render.Paint) {
	if paint == nil {
		return
	}
	writeColorKey(b, paint.Fill)
	b.WriteByte('|')
	writeColorKey(b, paint.Stroke)
	b.WriteByte('|')
	b.WriteString(shortFloat(paint.LineWidth))
	b.WriteByte('|')
	b.WriteString(paint.Hatch)
	b.WriteByte('|')
	writeColorKey(b, paint.HatchColor)
	b.WriteByte('|')
	b.WriteString(shortFloat(paint.HatchLineWidth))
	b.WriteByte('|')
	b.WriteString(shortFloat(paint.HatchSpacing))
}

func writeColorKey(b *strings.Builder, c render.Color) {
	b.WriteString(shortFloat(c.R))
	b.WriteByte(',')
	b.WriteString(shortFloat(c.G))
	b.WriteByte(',')
	b.WriteString(shortFloat(c.B))
	b.WriteByte(',')
	b.WriteString(shortFloat(c.A))
}
