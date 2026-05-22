package pdf

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/cwbudde/matplotlib-go/render"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// buildDocument assembles the PDF bytes for one page given the encoded
// content stream.
func buildDocument(width, height int, contentStream []byte, images []pdfImage, hatches []pdfHatchPattern, fillPatterns []pdfFillPattern, shadings []pdfShading, forms []pdfFormXObject, alphaStates []pdfAlphaState, fonts []pdfEmbeddedFont, opts render.PDFOptions) ([]byte, error) {
	imageObjects := assignImageObjects(images, 6)
	hatchObjects := assignHatchObjects(hatches, nextImageObjectID(imageObjects, 6))
	fillPatternObjects := assignFillPatternObjects(fillPatterns, nextHatchObjectID(hatchObjects, nextImageObjectID(imageObjects, 6)))
	shadingObjects := assignShadingObjects(shadings, nextFillPatternObjectID(fillPatternObjects, nextHatchObjectID(hatchObjects, nextImageObjectID(imageObjects, 6))))
	formObjects := assignFormObjects(forms, nextShadingObjectID(shadingObjects, nextFillPatternObjectID(fillPatternObjects, nextHatchObjectID(hatchObjects, nextImageObjectID(imageObjects, 6)))))
	fontObjects := assignFontObjects(fonts, nextFormObjectID(formObjects, nextShadingObjectID(shadingObjects, nextFillPatternObjectID(fillPatternObjects, nextHatchObjectID(hatchObjects, nextImageObjectID(imageObjects, 6))))))
	// We emit five fixed indirect objects, followed by image XObjects:
	//   1: /Catalog
	//   2: /Pages
	//   3: /Page
	//   4: content stream
	//   5: /Info (always present so the trailer can reference it)
	w := newPDFWriter()
	w.header()

	w.beginObject(1)
	w.writeString("<< /Type /Catalog /Pages 2 0 R >>")
	w.endObject()

	w.beginObject(2)
	w.writeString("<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	w.endObject()

	w.beginObject(3)
	fmt.Fprintf(&w.buf, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %d %d] /Contents 4 0 R /Resources %s >>",
		width, height, pageResources(imageObjects, hatchObjects, fillPatternObjects, shadingObjects, formObjects, alphaStates, fontObjects))
	w.endObject()

	// Compress the content stream with FlateDecode for determinism and size.
	encoded, err := flateEncode(contentStream)
	if err != nil {
		return nil, fmt.Errorf("pdf: flate encode content stream: %w", err)
	}
	w.beginObject(4)
	fmt.Fprintf(&w.buf, "<< /Length %d /Filter /FlateDecode >>\nstream\n", len(encoded))
	w.buf.Write(encoded)
	w.writeString("\nendstream")
	w.endObject()

	w.beginObject(5)
	w.writeInfo(opts)
	w.endObject()

	for _, img := range imageObjects {
		if err := w.writeImageObject(img); err != nil {
			return nil, err
		}
		if img.smaskID != 0 {
			if err := w.writeSoftMaskObject(img); err != nil {
				return nil, err
			}
		}
	}
	for _, hatch := range hatchObjects {
		if err := w.writeHatchPatternObject(hatch); err != nil {
			return nil, err
		}
	}
	for _, pattern := range fillPatternObjects {
		if err := w.writeFillPatternObject(pattern); err != nil {
			return nil, err
		}
	}
	for _, shading := range shadingObjects {
		w.writeShadingObject(shading)
	}
	for _, form := range formObjects {
		if err := w.writeFormXObject(form, alphaStates); err != nil {
			return nil, err
		}
	}
	for _, font := range fontObjects {
		if err := w.writeEmbeddedFontObjects(font); err != nil {
			return nil, err
		}
	}

	xrefOffset := w.buf.Len()
	w.writeXRef()
	w.writeTrailer(xrefOffset)

	return w.buf.Bytes(), nil
}

// pdfWriter helps assemble a PDF document with deterministic xref offsets.
type pdfWriter struct {
	buf     bytes.Buffer
	offsets []int // offsets[i] is the byte offset of object i (1-indexed; offsets[0] is unused)
}

func newPDFWriter() *pdfWriter {
	return &pdfWriter{offsets: []int{0}}
}

func (w *pdfWriter) header() {
	// PDF-1.7 plus a 4-byte binary marker to satisfy PDF readers that look for
	// non-ASCII bytes in the header comment.
	w.buf.WriteString("%PDF-1.7\n")
	w.buf.WriteString("%\xE2\xE3\xCF\xD3\n")
}

func (w *pdfWriter) beginObject(id int) {
	for len(w.offsets) <= id {
		w.offsets = append(w.offsets, 0)
	}
	w.offsets[id] = w.buf.Len()
	fmt.Fprintf(&w.buf, "%d 0 obj\n", id)
}

func (w *pdfWriter) endObject() {
	w.buf.WriteString("\nendobj\n")
}

func (w *pdfWriter) writeString(s string) {
	w.buf.WriteString(s)
}

func (w *pdfWriter) writeInfo(opts render.PDFOptions) {
	w.buf.WriteString("<< /Producer (matplotlib-go)")
	if len(opts.Metadata) > 0 {
		// Sort keys for deterministic order.
		keys := sortedKeys(opts.Metadata)
		for _, k := range keys {
			fmt.Fprintf(&w.buf, " /%s %s", escapeName(k), pdfLiteralString(opts.Metadata[k]))
		}
	}
	if date := resolveCreationDate(opts.CreationDate); !date.IsZero() {
		fmt.Fprintf(&w.buf, " /CreationDate %s", pdfDateString(date))
	}
	w.buf.WriteString(" >>")
}

func (w *pdfWriter) writeImageObject(img pdfImageObject) error {
	filter := img.filter
	if filter == "" {
		filter = "FlateDecode"
	}
	encoded := img.rgb
	if filter == "FlateDecode" {
		var err error
		encoded, err = flateEncode(pngPredictorRows(img.rgb, img.width, img.height, imageColorCount(img.pdfImage)))
		if err != nil {
			return fmt.Errorf("pdf: flate encode image %s: %w", img.name, err)
		}
	}
	w.beginObject(img.objectID)
	fmt.Fprintf(&w.buf,
		"<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8",
		img.width, img.height,
	)
	if img.smaskID != 0 {
		fmt.Fprintf(&w.buf, " /SMask %d 0 R", img.smaskID)
	}
	fmt.Fprintf(&w.buf, " /Length %d /Filter /%s", len(encoded), escapeName(filter))
	if filter == "FlateDecode" {
		fmt.Fprintf(&w.buf, " /DecodeParms << /Predictor 10 /Colors %d /Columns %d /BitsPerComponent 8 >>",
			imageColorCount(img.pdfImage), img.width)
	}
	w.buf.WriteString(" >>\nstream\n")
	w.buf.Write(encoded)
	w.writeString("\nendstream")
	w.endObject()
	return nil
}

func (w *pdfWriter) writeSoftMaskObject(img pdfImageObject) error {
	encoded, err := flateEncode(pngPredictorRows(img.alpha, img.width, img.height, 1))
	if err != nil {
		return fmt.Errorf("pdf: flate encode image soft mask %s: %w", img.name, err)
	}
	w.beginObject(img.smaskID)
	fmt.Fprintf(&w.buf,
		"<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceGray /BitsPerComponent 8 /Length %d /Filter /FlateDecode /DecodeParms << /Predictor 10 /Colors 1 /Columns %d /BitsPerComponent 8 >> >>\nstream\n",
		img.width, img.height, len(encoded),
		img.width,
	)
	w.buf.Write(encoded)
	w.writeString("\nendstream")
	w.endObject()
	return nil
}

func (w *pdfWriter) writeHatchPatternObject(hatch pdfHatchPatternObject) error {
	stream := hatchPatternStream(hatch.pdfHatchPattern)
	encoded, err := flateEncode(stream)
	if err != nil {
		return fmt.Errorf("pdf: flate encode hatch pattern %s: %w", hatch.name, err)
	}
	w.beginObject(hatch.objectID)
	fmt.Fprintf(&w.buf,
		"<< /Type /Pattern /PatternType 1 /PaintType 1 /TilingType 1 /BBox [0 0 72 72] /XStep 72 /YStep 72 /Resources << >> /Length %d /Filter /FlateDecode >>\nstream\n",
		len(encoded),
	)
	w.buf.Write(encoded)
	w.writeString("\nendstream")
	w.endObject()
	return nil
}

func (w *pdfWriter) writeFillPatternObject(pattern pdfFillPatternObject) error {
	stream := fillPatternStream(pattern.pattern)
	encoded, err := flateEncode(stream)
	if err != nil {
		return fmt.Errorf("pdf: flate encode fill pattern %s: %w", pattern.name, err)
	}
	cell := normalizedPatternCell(pattern.pattern.Cell)
	w.beginObject(pattern.objectID)
	fmt.Fprintf(&w.buf,
		"<< /Type /Pattern /PatternType 1 /PaintType 1 /TilingType 1 /BBox [%s %s %s %s] /XStep %s /YStep %s /Resources << >>",
		shortFloat(cell.Min.X),
		shortFloat(cell.Min.Y),
		shortFloat(cell.Max.X),
		shortFloat(cell.Max.Y),
		shortFloat(cell.W()),
		shortFloat(cell.H()),
	)
	if pattern.pattern.HasTransform {
		fmt.Fprintf(&w.buf, " /Matrix [%s %s %s %s %s %s]",
			shortFloat(pattern.pattern.Transform.A),
			shortFloat(pattern.pattern.Transform.B),
			shortFloat(pattern.pattern.Transform.C),
			shortFloat(pattern.pattern.Transform.D),
			shortFloat(pattern.pattern.Transform.E),
			shortFloat(pattern.pattern.Transform.F),
		)
	}
	fmt.Fprintf(&w.buf, " /Length %d /Filter /FlateDecode >>\nstream\n", len(encoded))
	w.buf.Write(encoded)
	w.writeString("\nendstream")
	w.endObject()
	return nil
}

func (w *pdfWriter) writeShadingObject(shading pdfShadingObject) {
	w.beginObject(shading.objectID)
	w.writeString(shadingDictionary(shading.gradient))
	w.endObject()
}

func (w *pdfWriter) writeFormXObject(form pdfFormXObjectObject, alphaStates []pdfAlphaState) error {
	var stream bytes.Buffer
	if form.hasContent {
		stream.Write(form.content)
	} else {
		formPaint := render.Paint{
			LineJoin: form.lineJoin,
			LineCap:  form.lineCap,
		}
		writeLineState(&stream, &formPaint)
		if !writePathOps(&stream, form.path) {
			return nil
		}
		stream.WriteString(form.paintOp)
		stream.WriteByte('\n')
	}
	encoded, err := flateEncode(stream.Bytes())
	if err != nil {
		return fmt.Errorf("pdf: flate encode form %s: %w", form.name, err)
	}
	w.beginObject(form.objectID)
	resources := "<< >>"
	if form.hasContent && len(alphaStates) > 0 {
		resources = pageResources(nil, nil, nil, nil, nil, alphaStates, nil)
	}
	fmt.Fprintf(&w.buf,
		"<< /Type /XObject /Subtype /Form /BBox [%s %s %s %s] /Resources %s",
		shortFloat(form.bbox.Min.X),
		shortFloat(form.bbox.Min.Y),
		shortFloat(form.bbox.Max.X),
		shortFloat(form.bbox.Max.Y),
		resources,
	)
	if form.transparencyGroup {
		w.writeString(" /Group << /S /Transparency /CS /DeviceRGB >>")
	}
	fmt.Fprintf(&w.buf, " /Length %d /Filter /FlateDecode >>\nstream\n", len(encoded))
	w.buf.Write(encoded)
	w.writeString("\nendstream")
	w.endObject()
	return nil
}

func (w *pdfWriter) writeEmbeddedFontObjects(fontObj pdfEmbeddedFontObject) error {
	fontData, err := sfnt.Parse(fontObj.data)
	if err != nil {
		return fmt.Errorf("pdf: parse embedded font %s: %w", fontObj.name, err)
	}
	subsetName := subsetFontName(fontObj.pdfEmbeddedFont)

	w.beginObject(fontObj.type0ID)
	fmt.Fprintf(&w.buf,
		"<< /Type /Font /Subtype /Type0 /BaseFont /%s /Encoding /Identity-H /DescendantFonts [%d 0 R] /ToUnicode %d 0 R >>",
		escapeName(subsetName),
		fontObj.cidFontID,
		fontObj.toUnicodeID,
	)
	w.endObject()

	w.beginObject(fontObj.cidFontID)
	fmt.Fprintf(&w.buf,
		"<< /Type /Font /Subtype /CIDFontType2 /BaseFont /%s /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> /FontDescriptor %d 0 R /W %d 0 R /CIDToGIDMap %d 0 R >>",
		escapeName(subsetName),
		fontObj.descriptorID,
		fontObj.widthsID,
		fontObj.cidToGIDID,
	)
	w.endObject()

	descriptor := pdfFontDescriptor(fontObj, fontData, subsetName)
	w.beginObject(fontObj.descriptorID)
	w.writeString(descriptor)
	w.endObject()

	encodedFont, err := flateEncode(fontObj.data)
	if err != nil {
		return fmt.Errorf("pdf: flate encode font %s: %w", fontObj.name, err)
	}
	w.beginObject(fontObj.fontFileID)
	fmt.Fprintf(&w.buf, "<< /Length %d /Length1 %d /Filter /FlateDecode >>\nstream\n", len(encodedFont), len(fontObj.data))
	w.buf.Write(encodedFont)
	w.writeString("\nendstream")
	w.endObject()

	cidMap := pdfCIDToGIDMap(fontObj.gidByCID)
	encodedCIDMap, err := flateEncode(cidMap)
	if err != nil {
		return fmt.Errorf("pdf: flate encode CIDToGIDMap %s: %w", fontObj.name, err)
	}
	w.beginObject(fontObj.cidToGIDID)
	fmt.Fprintf(&w.buf, "<< /Length %d /Filter /FlateDecode >>\nstream\n", len(encodedCIDMap))
	w.buf.Write(encodedCIDMap)
	w.writeString("\nendstream")
	w.endObject()

	w.beginObject(fontObj.widthsID)
	w.writeString(pdfFontWidths(fontObj.gidByCID, fontData))
	w.endObject()

	toUnicode, err := flateEncode(pdfToUnicodeCMap(fontObj.runeByCID))
	if err != nil {
		return fmt.Errorf("pdf: flate encode ToUnicode %s: %w", fontObj.name, err)
	}
	w.beginObject(fontObj.toUnicodeID)
	fmt.Fprintf(&w.buf, "<< /Length %d /Filter /FlateDecode >>\nstream\n", len(toUnicode))
	w.buf.Write(toUnicode)
	w.writeString("\nendstream")
	w.endObject()
	return nil
}

func (w *pdfWriter) writeXRef() {
	w.buf.WriteString("xref\n")
	fmt.Fprintf(&w.buf, "0 %d\n", len(w.offsets))
	// Object 0 is the head of the free list.
	w.buf.WriteString("0000000000 65535 f \n")
	for i := 1; i < len(w.offsets); i++ {
		fmt.Fprintf(&w.buf, "%010d 00000 n \n", w.offsets[i])
	}
}

func (w *pdfWriter) writeTrailer(xrefOffset int) {
	fmt.Fprintf(&w.buf,
		"trailer\n<< /Size %d /Root 1 0 R /Info 5 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(w.offsets), xrefOffset,
	)
}

func flateEncode(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw, err := zlib.NewWriterLevel(&buf, zlib.BestSpeed)
	if err != nil {
		return nil, err
	}
	if _, err := zw.Write(data); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// stdlib sort is in the import list of registry but not here; use the
	// minimal in-place insertion sort because the key set is tiny.
	for i := 1; i < len(keys); i++ {
		j := i
		for j > 0 && keys[j-1] > keys[j] {
			keys[j-1], keys[j] = keys[j], keys[j-1]
			j--
		}
	}
	return keys
}

func pdfFontDescriptor(fontObj pdfEmbeddedFontObject, fontData *sfnt.Font, subsetName string) string {
	metrics, bounds := pdfFontMetrics(fontData)
	return fmt.Sprintf("<< /Type /FontDescriptor /FontName /%s /Flags 32 /FontBBox [%d %d %d %d] /Ascent %d /Descent %d /CapHeight %d /ItalicAngle 0 /StemV 0 /FontFile2 %d 0 R >>",
		escapeName(subsetName),
		bounds[0], bounds[1], bounds[2], bounds[3],
		metrics[0], metrics[1], metrics[2],
		fontObj.fontFileID,
	)
}

func pdfFontMetrics(fontData *sfnt.Font) ([3]int, [4]int) {
	if fontData == nil {
		return [3]int{800, -200, 700}, [4]int{-1000, -300, 2000, 1000}
	}
	ppem := fixed.I(1000)
	var buf sfnt.Buffer
	metrics, err := fontData.Metrics(&buf, ppem, xfont.HintingNone)
	if err != nil {
		return [3]int{800, -200, 700}, [4]int{-1000, -300, 2000, 1000}
	}
	bounds, err := fontData.Bounds(&buf, ppem, xfont.HintingNone)
	if err != nil {
		return [3]int{
			roundFixed(metrics.Ascent),
			roundFixed(-metrics.Descent),
			roundFixed(metrics.Ascent),
		}, [4]int{-1000, -300, 2000, 1000}
	}
	return [3]int{
			roundFixed(metrics.Ascent),
			roundFixed(-metrics.Descent),
			roundFixed(metrics.Ascent),
		}, [4]int{
			roundFixed(bounds.Min.X),
			roundFixed(bounds.Min.Y),
			roundFixed(bounds.Max.X),
			roundFixed(bounds.Max.Y),
		}
}

func pdfCIDToGIDMap(gidByCID map[uint16]sfnt.GlyphIndex) []byte {
	maxCID := uint16(0)
	for cid := range gidByCID {
		if cid > maxCID {
			maxCID = cid
		}
	}
	out := make([]byte, (int(maxCID)+1)*2)
	for cid, gid := range gidByCID {
		binary.BigEndian.PutUint16(out[int(cid)*2:int(cid)*2+2], uint16(gid))
	}
	return out
}

func pdfFontWidths(gidByCID map[uint16]sfnt.GlyphIndex, fontData *sfnt.Font) string {
	cids := sortedFontCIDs(gidByCID)
	if len(cids) == 0 || fontData == nil {
		return "[]"
	}
	var b strings.Builder
	var buf sfnt.Buffer
	ppem := fixed.I(1000)
	b.WriteString("[")
	i := 0
	for i < len(cids) {
		start := cids[i]
		b.WriteByte(' ')
		b.WriteString(strconv.Itoa(int(start)))
		b.WriteString(" [")
		prev := start
		for i < len(cids) {
			cid := cids[i]
			if cid != prev && cid != prev+1 {
				break
			}
			width := 0
			if advance, err := fontData.GlyphAdvance(&buf, gidByCID[cid], ppem, xfont.HintingNone); err == nil {
				width = roundFixed(advance)
			}
			if cid != start {
				b.WriteByte(' ')
			}
			b.WriteString(strconv.Itoa(width))
			prev = cid
			i++
		}
		b.WriteByte(']')
	}
	b.WriteString(" ]")
	return b.String()
}

func pdfToUnicodeCMap(runeByCID map[uint16]rune) []byte {
	cids := sortedRuneCIDs(runeByCID)
	var b strings.Builder
	b.WriteString("/CIDInit /ProcSet findresource begin\n")
	b.WriteString("12 dict begin\nbegincmap\n")
	b.WriteString("/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n")
	b.WriteString("/CMapName /Adobe-Identity-UCS def\n/CMapType 2 def\n")
	b.WriteString("1 begincodespacerange\n<0000> <FFFF>\nendcodespacerange\n")
	fmt.Fprintf(&b, "%d beginbfchar\n", len(cids))
	for _, cid := range cids {
		fmt.Fprintf(&b, "<%04X> <%s>\n", cid, utf16BEHex(runeByCID[cid]))
	}
	b.WriteString("endbfchar\nendcmap\nCMapName currentdict /CMap defineresource pop\nend\nend\n")
	return []byte(b.String())
}

func utf16BEHex(r rune) string {
	encoded := utf16.Encode([]rune{r})
	var b strings.Builder
	for _, v := range encoded {
		fmt.Fprintf(&b, "%04X", v)
	}
	return b.String()
}

func sortedFontCIDs(gidByCID map[uint16]sfnt.GlyphIndex) []uint16 {
	cids := make([]uint16, 0, len(gidByCID))
	for cid := range gidByCID {
		cids = append(cids, cid)
	}
	sortUint16s(cids)
	return cids
}

func sortedRuneCIDs(runeByCID map[uint16]rune) []uint16 {
	cids := make([]uint16, 0, len(runeByCID))
	for cid := range runeByCID {
		cids = append(cids, cid)
	}
	sortUint16s(cids)
	return cids
}

func sortUint16s(values []uint16) {
	for i := 1; i < len(values); i++ {
		j := i
		for j > 0 && values[j-1] > values[j] {
			values[j-1], values[j] = values[j], values[j-1]
			j--
		}
	}
}

func roundFixed(v fixed.Int26_6) int {
	f := float64(v) / 64
	if f < 0 {
		return int(math.Ceil(f - 0.5))
	}
	return int(math.Floor(f + 0.5))
}

func assignImageObjects(images []pdfImage, firstID int) []pdfImageObject {
	if len(images) == 0 {
		return nil
	}
	out := make([]pdfImageObject, len(images))
	nextID := firstID
	for i, img := range images {
		out[i] = pdfImageObject{
			pdfImage: img,
			objectID: nextID,
		}
		nextID++
		if img.hasAlpha {
			out[i].smaskID = nextID
			nextID++
		}
	}
	return out
}

func nextImageObjectID(images []pdfImageObject, firstID int) int {
	nextID := firstID
	for _, img := range images {
		if img.objectID >= nextID {
			nextID = img.objectID + 1
		}
		if img.smaskID >= nextID {
			nextID = img.smaskID + 1
		}
	}
	return nextID
}

func assignHatchObjects(hatches []pdfHatchPattern, firstID int) []pdfHatchPatternObject {
	if len(hatches) == 0 {
		return nil
	}
	out := make([]pdfHatchPatternObject, len(hatches))
	for i, hatch := range hatches {
		out[i] = pdfHatchPatternObject{
			pdfHatchPattern: hatch,
			objectID:        firstID + i,
		}
	}
	return out
}

func nextHatchObjectID(hatches []pdfHatchPatternObject, firstID int) int {
	nextID := firstID
	for _, hatch := range hatches {
		if hatch.objectID >= nextID {
			nextID = hatch.objectID + 1
		}
	}
	return nextID
}

func assignFillPatternObjects(patterns []pdfFillPattern, firstID int) []pdfFillPatternObject {
	if len(patterns) == 0 {
		return nil
	}
	out := make([]pdfFillPatternObject, len(patterns))
	for i, pattern := range patterns {
		out[i] = pdfFillPatternObject{
			pdfFillPattern: pattern,
			objectID:       firstID + i,
		}
	}
	return out
}

func nextFillPatternObjectID(patterns []pdfFillPatternObject, firstID int) int {
	nextID := firstID
	for _, pattern := range patterns {
		if pattern.objectID >= nextID {
			nextID = pattern.objectID + 1
		}
	}
	return nextID
}

func assignShadingObjects(shadings []pdfShading, firstID int) []pdfShadingObject {
	if len(shadings) == 0 {
		return nil
	}
	out := make([]pdfShadingObject, len(shadings))
	for i, shading := range shadings {
		out[i] = pdfShadingObject{
			pdfShading: shading,
			objectID:   firstID + i,
		}
	}
	return out
}

func nextShadingObjectID(shadings []pdfShadingObject, firstID int) int {
	nextID := firstID
	for _, shading := range shadings {
		if shading.objectID >= nextID {
			nextID = shading.objectID + 1
		}
	}
	return nextID
}

func assignFormObjects(forms []pdfFormXObject, firstID int) []pdfFormXObjectObject {
	if len(forms) == 0 {
		return nil
	}
	out := make([]pdfFormXObjectObject, len(forms))
	for i, form := range forms {
		out[i] = pdfFormXObjectObject{
			pdfFormXObject: form,
			objectID:       firstID + i,
		}
	}
	return out
}

func nextFormObjectID(forms []pdfFormXObjectObject, firstID int) int {
	nextID := firstID
	for _, form := range forms {
		if form.objectID >= nextID {
			nextID = form.objectID + 1
		}
	}
	return nextID
}

func assignFontObjects(fonts []pdfEmbeddedFont, firstID int) []pdfEmbeddedFontObject {
	if len(fonts) == 0 {
		return nil
	}
	out := make([]pdfEmbeddedFontObject, len(fonts))
	nextID := firstID
	for i, font := range fonts {
		out[i] = pdfEmbeddedFontObject{
			pdfEmbeddedFont: font,
			type0ID:         nextID,
			cidFontID:       nextID + 1,
			descriptorID:    nextID + 2,
			fontFileID:      nextID + 3,
			cidToGIDID:      nextID + 4,
			widthsID:        nextID + 5,
			toUnicodeID:     nextID + 6,
		}
		nextID += 7
	}
	return out
}

func pageResources(images []pdfImageObject, hatches []pdfHatchPatternObject, fillPatterns []pdfFillPatternObject, shadings []pdfShadingObject, forms []pdfFormXObjectObject, alphaStates []pdfAlphaState, fonts []pdfEmbeddedFontObject) string {
	if len(images) == 0 && len(hatches) == 0 && len(fillPatterns) == 0 && len(shadings) == 0 && len(forms) == 0 && len(alphaStates) == 0 && len(fonts) == 0 {
		return "<< >>"
	}
	var b strings.Builder
	b.WriteString("<<")
	if len(fonts) > 0 {
		b.WriteString(" /Font <<")
		for _, font := range fonts {
			fmt.Fprintf(&b, " /%s %d 0 R", escapeName(font.name), font.type0ID)
		}
		b.WriteString(" >>")
	}
	if len(images) > 0 || len(forms) > 0 {
		b.WriteString(" /XObject <<")
		for _, img := range images {
			fmt.Fprintf(&b, " /%s %d 0 R", escapeName(img.name), img.objectID)
		}
		for _, form := range forms {
			fmt.Fprintf(&b, " /%s %d 0 R", escapeName(form.name), form.objectID)
		}
		b.WriteString(" >>")
	}
	if len(hatches) > 0 || len(fillPatterns) > 0 {
		b.WriteString(" /Pattern <<")
		for _, hatch := range hatches {
			fmt.Fprintf(&b, " /%s %d 0 R", escapeName(hatch.name), hatch.objectID)
		}
		for _, pattern := range fillPatterns {
			fmt.Fprintf(&b, " /%s %d 0 R", escapeName(pattern.name), pattern.objectID)
		}
		b.WriteString(" >>")
	}
	if len(shadings) > 0 {
		b.WriteString(" /Shading <<")
		for _, shading := range shadings {
			fmt.Fprintf(&b, " /%s %d 0 R", escapeName(shading.name), shading.objectID)
		}
		b.WriteString(" >>")
	}
	if len(alphaStates) > 0 {
		b.WriteString(" /ExtGState <<")
		for _, state := range alphaStates {
			fmt.Fprintf(&b, " /%s << /Type /ExtGState /CA %s /ca %s >>",
				escapeName(state.name),
				shortFloat(state.strokeAlpha),
				shortFloat(state.fillAlpha),
			)
		}
		b.WriteString(" >>")
	}
	b.WriteString(" >>")
	return b.String()
}

// pdfLiteralString encodes s as a PDF literal string. It escapes parentheses
// and backslashes per ISO 32000-1 §7.3.4.2.
func pdfLiteralString(s string) string {
	var b strings.Builder
	b.WriteByte('(')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '(', ')', '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte(')')
	return b.String()
}

// escapeName escapes a PDF name token per ISO 32000-1 §7.3.5. Only safe-ASCII
// alphanumeric characters and a handful of punctuation are emitted verbatim;
// everything else is hex-escaped.
func escapeName(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= '0' && c <= '9') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			c == '.' || c == '-' || c == '_' {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "#%02X", c)
		}
	}
	return b.String()
}

// resolveCreationDate returns the explicit override when set, otherwise the
// SOURCE_DATE_EPOCH environment value, otherwise a zero time (which suppresses
// the /CreationDate entry for full reproducibility).
func resolveCreationDate(explicit time.Time) time.Time {
	if !explicit.IsZero() {
		return explicit
	}
	if v := os.Getenv("SOURCE_DATE_EPOCH"); v != "" {
		if secs, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Unix(secs, 0).UTC()
		}
	}
	return time.Time{}
}

// pdfDateString formats t per ISO 32000-1 §7.9.4 as `(D:YYYYMMDDHHmmSSZ)`.
func pdfDateString(t time.Time) string {
	t = t.UTC()
	return fmt.Sprintf("(D:%04d%02d%02d%02d%02d%02dZ)",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(),
	)
}
