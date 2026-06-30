package ps

import (
	"fmt"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// DrawMarkers renders one marker path at many display-space offsets using a
// reusable PostScript procedure for identical marker geometry and paint.
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
		marker := affinePath(batch.Marker, normalizedAffine(item.Transform))
		if !marker.Validate() || len(marker.C) == 0 {
			continue
		}
		if !paintVisible(&item.Paint) {
			warnGradientCollectionDrop(&item.Paint)
			continue
		}
		name := r.registerMarkerProcedure(marker, &item.Paint)
		if name == "" {
			continue
		}
		fmt.Fprintf(
			&r.content, "gsave\n%s %s translate\n%s\ngrestore\n",
			shortFloat(item.Offset.X),
			shortFloat(item.Offset.Y),
			name,
		)
		emitted = true
	}
	return emitted
}

// DrawPathCollection renders display-space paths through reusable PostScript
// procedures keyed by path geometry and paint.
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
			warnGradientCollectionDrop(&paint)
			continue
		}
		name := r.registerPathProcedure("P", item.Path, &paint)
		if name == "" {
			continue
		}
		fmt.Fprintf(&r.content, "%s\n", name)
		emitted = true
	}
	return emitted
}

func (r *Renderer) registerMarkerProcedure(marker geom.Path, paint *render.Paint) string {
	return r.registerPathProcedure("M", marker, paint)
}

func (r *Renderer) registerPathProcedure(prefix string, path geom.Path, paint *render.Paint) string {
	key := pathProcedureKey(prefix, path, paint)
	if name, ok := r.markerIDs[key]; ok {
		return name
	}
	var body strings.Builder
	if !writePathOps(&body, path) || !writePathPaintOps(&body, path, paint) {
		return ""
	}
	name := fmt.Sprintf("%s%d", prefix, len(r.markerIDs)+1)
	r.markerIDs[key] = name
	fmt.Fprintf(&r.content, "/%s {\n%s} bind def\n", name, body.String())
	return name
}

func pathProcedureKey(prefix string, path geom.Path, paint *render.Paint) string {
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
