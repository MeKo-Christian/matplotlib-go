package svg

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func (r *Renderer) DrawMarkers(batch render.MarkerBatch) bool {
	if rr := r.activeRaster(); rr != nil {
		if markers, ok := rr.(render.MarkerDrawer); ok {
			return markers.DrawMarkers(batch)
		}
		return false
	}
	if len(batch.Marker.C) == 0 || len(batch.Items) == 0 {
		return false
	}
	if !batch.Marker.Validate() {
		return false
	}
	var b strings.Builder
	emitted := 0
	for i := range batch.Items {
		item := &batch.Items[i]
		paint := item.Paint
		hasFill := paint.Fill.A > 0
		hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
		if !hasFill && !hasStroke {
			continue
		}

		// Bake the per-item transform into the geometry. Applying it on <use>
		// would also scale the SVG stroke, although Paint.LineWidth is already
		// expressed in display space. This mirrors the PDF and raster backends.
		marker := affinePath(batch.Marker, item.Transform)
		if !marker.Validate() || len(marker.C) == 0 {
			continue
		}
		d := buildPathData(marker)
		if d == "" {
			continue
		}
		markerID := r.registerMarker(d)
		t := geom.Affine{A: 1, D: 1, E: item.Offset.X, F: item.Offset.Y}

		b.WriteString(`<use href="#`)
		b.WriteString(markerID)
		b.WriteString(`" xlink:href="#`)
		b.WriteString(markerID)
		b.WriteString(`"`)
		writeAttr(&b, "transform", matrixTransform(r.deviceFlip().Mul(t)))
		writeForcedOpacity(&b, paint)

		if hasFill {
			writeColorAttrs(&b, "fill", paint.Fill, forcedOpacity(paint))
		} else {
			writeAttr(&b, "fill", "none")
		}

		writeStrokeAttrs(&b, paint)

		b.WriteString(" />")
		emitted++
	}

	if emitted == 0 {
		return true
	}

	r.nodes = append(r.nodes, r.newNode(b.String()))
	return true
}

func (r *Renderer) registerMarker(d string) string {
	if id, ok := r.markerDefs[d]; ok {
		return id
	}

	r.markerIDCounter++
	id := r.defID("marker", d, r.markerIDCounter)
	r.markerDefs[d] = id
	r.markerOrder = append(r.markerOrder, markerDef{id: id, data: d})
	return id
}

func (r *Renderer) DrawPathCollection(batch render.PathCollectionBatch) bool {
	if rr := r.activeRaster(); rr != nil {
		if paths, ok := rr.(render.PathCollectionDrawer); ok {
			return paths.DrawPathCollection(batch)
		}
		return false
	}
	if len(batch.Items) == 0 {
		return false
	}

	var b strings.Builder
	emitted := 0
	for i := range batch.Items {
		item := &batch.Items[i]
		if !item.Path.Validate() {
			continue
		}
		d := buildPathData(affinePath(item.Path, r.deviceFlip()))
		if d == "" {
			continue
		}

		paint := item.Paint
		if item.Hatch != "" {
			paint.Hatch = item.Hatch
			paint.HatchColor = item.HatchColor
			paint.HatchLineWidth = item.HatchWidth
			paint.HatchSpacing = item.HatchSpacing
		}
		hasFill := paint.Fill.A > 0
		hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
		hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
		if !hasFill && !hasHatch && !hasStroke {
			continue
		}

		pathID := r.registerCollectionPath(d)
		b.WriteString(`<use href="#`)
		b.WriteString(pathID)
		b.WriteString(`" xlink:href="#`)
		b.WriteString(pathID)
		b.WriteString(`"`)
		writeForcedOpacity(&b, paint)
		if hasHatch {
			writeAttr(&b, "fill", "url(#"+r.registerHatch(paint)+")")
		} else if hasFill {
			writeColorAttrs(&b, "fill", paint.Fill, forcedOpacity(paint))
		} else {
			writeAttr(&b, "fill", "none")
		}
		writeStrokeAttrs(&b, paint)
		b.WriteString(" />")
		emitted++
	}

	if emitted == 0 {
		return true
	}

	r.nodes = append(r.nodes, r.newNode(b.String()))
	return true
}

func (r *Renderer) registerCollectionPath(d string) string {
	if id, ok := r.pathDefs[d]; ok {
		return id
	}

	r.pathIDCounter++
	id := r.defID("pathcoll", d, r.pathIDCounter)
	r.pathDefs[d] = id
	r.pathOrder = append(r.pathOrder, pathCollectionDef{id: id, data: d})
	return id
}
