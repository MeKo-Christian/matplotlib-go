package main_test

import (
	"strings"
	"testing"
)

func TestLargeFileAuditPlumbingIsDocumented(t *testing.T) {
	justfile := readTextFile(t, "Justfile")
	for _, want := range []string{
		"large-file-audit:",
		"git ls-files --cached --others --exclude-standard '*.go'",
		"Large tracked or untracked Go files (>= 1000 lines)",
		"Large tracked or untracked non-Go artifacts (>= 256 KiB)",
	} {
		if !strings.Contains(justfile, want) {
			t.Fatalf("Justfile missing %q", want)
		}
	}

	doc := readTextFile(t, "docs/large-file-decomposition.md")
	for _, want := range []string{
		"# Large File Decomposition",
		"`just large-file-audit`",
		"## Baseline Inventory",
		"plot3d/wire_surface_test.go",
		"core/contour.go",
		"docs/matplotlib-parity-status.md",
		"testdata/svg_golden/mathtext_basic.svg",
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("large-file decomposition doc missing %q", want)
		}
	}
}

func TestGeneratedDataStrategyIsDocumented(t *testing.T) {
	doc := readTextFile(t, "docs/large-file-decomposition.md")
	for _, want := range []string{
		"## Generated and Fixture Data Strategy",
		"`internal/examplecatalog/public_surface_parity.go`",
		"Keep-large curated catalog",
		"`test/testdata/parity_surface/upstream_public_surface.json`",
		"`internal/examplecatalog/extract_public_surface.py`",
		"`color/named_colors_data.go`",
		"Keep-large generated table",
		"`third_party/matplotlib/lib/matplotlib/_color_data.py`",
		"`TestNamedColorInventoryMatchesMatplotlibTables`",
		"Golden/reference PNG, SVG, PDF, and JSON fixtures",
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("large-file decomposition doc missing L7 decision detail %q", want)
		}
	}

	publicSurface := readTextFile(t, "internal/examplecatalog/public_surface_parity.go")
	if !strings.Contains(publicSurface, "curated parity catalog") {
		t.Fatal("public_surface_parity.go should identify itself as curated catalog data")
	}

	namedColors := readTextFile(t, "color/named_colors_data.go")
	if !strings.HasPrefix(namedColors, "// Code generated from third_party/matplotlib/lib/matplotlib/_color_data.py; DO NOT EDIT.") {
		t.Fatal("named_colors_data.go should retain its generated-source header")
	}
}

func TestContourAPISplitIsTracked(t *testing.T) {
	api := readTextFile(t, "core/contour_api.go")
	for _, want := range []string{
		"type ContourOptions struct",
		"type ClabelOptions struct",
		"type ContourLabel struct",
		"type ContourSet struct",
		"func (a *Axes) Contour(data [][]float64, opt ContourOptions) *ContourSet",
		"func (a *Axes) Contourf(data [][]float64, opt ContourOptions) *ContourSet",
		"func (a *Axes) TriContour(tri Triangulation, values []float64, opt ContourOptions) *ContourSet",
		"func (a *Axes) TriContourf(tri Triangulation, values []float64, opt ContourOptions) *ContourSet",
		"func (a *Axes) buildContourSet(tri Triangulation, values []float64, filled bool, opt ContourOptions) *ContourSet",
	} {
		if !strings.Contains(api, want) {
			t.Fatalf("contour_api.go missing %q", want)
		}
	}

	contour := readTextFile(t, "core/contour.go")
	for _, moved := range []string{
		"type ContourOptions struct",
		"type ContourSet struct",
		"func (a *Axes) buildContourSet(",
	} {
		if strings.Contains(contour, moved) {
			t.Fatalf("core/contour.go still contains moved API/construction item %q", moved)
		}
	}
}

func TestContourLevelsSplitIsTracked(t *testing.T) {
	levels := readTextFile(t, "core/contour_levels.go")
	for _, want := range []string{
		"func contourGridCoordsValues(data [][]float64, opt ContourOptions) ([]float64, []float64, []float64, bool)",
		"func triangleFinite(values []float64, tri [3]int) bool",
		"func resolvedContourCoords(size int, coords, edges []float64) []float64",
		"func contourLevels(values, explicit []float64, levelCount int, filled bool) []float64",
		"func contourLocatorLevels(minValue, maxValue float64, levelCount int, filled bool) []float64",
		"func dedupeFloat64(values []float64) []float64",
	} {
		if !strings.Contains(levels, want) {
			t.Fatalf("contour_levels.go missing %q", want)
		}
	}

	contour := readTextFile(t, "core/contour.go")
	for _, moved := range []string{
		"func contourGridCoordsValues(",
		"func contourLevels(",
		"func contourLocatorLevels(",
		"func dedupeFloat64(",
	} {
		if strings.Contains(contour, moved) {
			t.Fatalf("core/contour.go still contains moved level helper %q", moved)
		}
	}
}

func TestContourLinesSplitIsTracked(t *testing.T) {
	lines := readTextFile(t, "core/contour_lines.go")
	for _, want := range []string{
		"func contourPolylines(tri Triangulation, values, levels []float64) ([][]geom.Pt, []float64)",
		"func contourGridPolylines(x, y []float64, data [][]float64, levels []float64) ([][]geom.Pt, []float64)",
		"type contourBoundarySide uint8",
		"func orientStructuredOpenBoundaryPolyline(polyline []geom.Pt, x, y []float64) []geom.Pt",
		"func contourCellSegmentsForLevel(points [4]geom.Pt, values [4]float64, level float64) [][]geom.Pt",
		"func triangleContourSegment(points [3]geom.Pt, values [3]float64, level float64) ([]geom.Pt, bool)",
		"func stitchContourSegments(segments [][]geom.Pt) [][]geom.Pt",
		"func rotateClosedContourPolylineToMatplotlibStart(polyline []geom.Pt) []geom.Pt",
		"func reversePoints(points []geom.Pt) []geom.Pt",
	} {
		if !strings.Contains(lines, want) {
			t.Fatalf("contour_lines.go missing %q", want)
		}
	}

	contour := readTextFile(t, "core/contour.go")
	for _, moved := range []string{
		"func contourPolylines(",
		"func contourGridPolylines(",
		"type contourBoundarySide",
		"func contourCellSegmentsForLevel(",
		"func contourSegmentsForLevel(",
		"func triangleContourSegment(",
		"func stitchContourSegments(",
		"func rotateClosedContourPolylineToMatplotlibStart(",
		"func reversePoints(",
	} {
		if strings.Contains(contour, moved) {
			t.Fatalf("core/contour.go still contains moved line helper %q", moved)
		}
	}
}

func TestContourFilledSplitIsTracked(t *testing.T) {
	filled := readTextFile(t, "core/contour_filled.go")
	for _, want := range []string{
		"func contourBandPolygons(tri Triangulation, values, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64) ([][]geom.Pt, []render.Color, []string)",
		"func contourGridBandPolygons(x, y []float64, data [][]float64, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64) ([][]geom.Pt, []render.Color, []string)",
		"func contourCellBandPolygons(points [4]geom.Pt, values [4]float64, low, high float64) [][]geom.Pt",
		"func contourCellBandPolygon(points [4]geom.Pt, values [4]float64, low, high float64) []geom.Pt",
		"func contourSaddleBandPolygons(points [4]geom.Pt, values [4]float64, low, high float64) [][]geom.Pt",
		"func contourBandOutsideSameSide(a, b, low, high float64) bool",
		"func contourBandBoundaryIntersection(insidePoint geom.Pt, insideValue float64, outsidePoint geom.Pt, outsideValue float64, low, high float64) (geom.Pt, bool)",
		"func triangleBandPolygon(points [3]geom.Pt, values [3]float64, low, high float64) []geom.Pt",
		"type contourVertex struct",
		"func rotateContourPolygonToMatplotlibStart(points []geom.Pt) []geom.Pt",
		"func contourPolygonHasConsecutiveDuplicate(points []geom.Pt) bool",
		"func contourPolygonClosed(points []geom.Pt) bool",
		"func clipContourPolygonMin(polygon []contourVertex, threshold float64) []contourVertex",
		"func clipContourPolygonMax(polygon []contourVertex, threshold float64) []contourVertex",
		"func clipContourPolygon(polygon []contourVertex, inside func(float64) bool, threshold float64) []contourVertex",
		"func contourBandColor(low, high float64, idx int, opt ContourOptions, mapping ScalarMapInfo, alpha float64) render.Color",
	} {
		if !strings.Contains(filled, want) {
			t.Fatalf("contour_filled.go missing %q", want)
		}
	}

	contour := readTextFile(t, "core/contour.go")
	for _, moved := range []string{
		"func contourBandPolygons(",
		"func contourGridBandPolygons(",
		"func contourCellBandPolygons(",
		"func contourSaddleBandPolygons(",
		"func contourBandBoundaryIntersection(",
		"type contourVertex",
		"func rotateContourPolygonToMatplotlibStart(",
		"func contourPolygonClosed(",
		"func clipContourPolygon(",
		"func contourBandColor(",
	} {
		if strings.Contains(contour, moved) {
			t.Fatalf("core/contour.go still contains moved filled helper %q", moved)
		}
	}
}

func TestContourLabelsSplitIsTracked(t *testing.T) {
	labels := readTextFile(t, "core/contour_labels.go")
	for _, want := range []string{
		"func contourLabels(polylines [][]geom.Pt, levels []float64, colors []render.Color, formatter ticker.Formatter, rightSideUp bool) []contourLabel",
		"func (c *ContourSet) clabelLineIndices(levels []float64) ([]int, bool)",
		"func (c *ContourSet) clabelPlaceAutomatic(indices []int, opt ClabelOptions) []contourLabel",
		"func (c *ContourSet) clabelPlaceManual(indices []int, opt ClabelOptions) []contourLabel",
		"func (c *ContourSet) nearestContourLabelPoint(indices []int, point geom.Pt) (int, geom.Pt, float64, bool)",
		"func (c *ContourSet) clabelColor(segmentIndex int, level float64, labelIndex int, opt ClabelOptions) render.Color",
		"func publicContourLabels(labels []contourLabel) []ContourLabel",
		"func uniqueLevelsForIndices(levels []float64, indices []int) []float64",
		"func contourInlineLabelSegmentsForLevels(lines *LineCollection, levels, selectedLevels []float64, formatter ticker.Formatter, fontSize, inlineSpacing float64, rightSideUp bool, r render.Renderer, ctx *DrawContext) ([][]geom.Pt, []render.Color, []float64, [][]float64, []contourLabel)",
		"func contourLabelWidth(text string, fontSize float64, r render.Renderer, ctx *DrawContext) float64",
		"func contourLocateLabel(line []geom.Pt, labelWidth float64, placed []geom.Pt) (geom.Pt, int)",
		"func splitContourPolylineForLabel(data, screen []geom.Pt, labelIdx int, labelWidth, spacing float64, rightSideUp bool) (float64, [][]geom.Pt)",
		"func splitClosedContourPolylineForLabel(data, screen []geom.Pt, cpls []float64, labelIdx int, labelWidth, spacing float64, rightSideUp bool) (float64, [][]geom.Pt)",
		"func contourRotatedTextAnchor(center geom.Pt, layout singleLineTextLayout, angle float64) geom.Pt",
		"func contourFormatter(formatter ticker.Formatter) ticker.Formatter",
		"func polylineLabelPlacement(polyline []geom.Pt) (geom.Pt, float64)",
		"func normalizeLabelAngle(angle float64) float64",
	} {
		if !strings.Contains(labels, want) {
			t.Fatalf("contour_labels.go missing %q", want)
		}
	}

	contour := readTextFile(t, "core/contour.go")
	for _, moved := range []string{
		"func contourLabels(",
		"func (c *ContourSet) clabelLineIndices(",
		"func (c *ContourSet) clabelPlaceAutomatic(",
		"func (c *ContourSet) clabelPlaceManual(",
		"func contourInlineLabelSegmentsForLevels(",
		"func contourLabelWidth(",
		"func contourLocateLabel(",
		"func splitContourPolylineForLabel(",
		"func contourRotatedTextAnchor(",
		"func contourFormatter(",
		"func polylineLabelPlacement(",
		"func normalizeLabelAngle(",
	} {
		if strings.Contains(contour, moved) {
			t.Fatalf("core/contour.go still contains moved label helper %q", moved)
		}
	}
}

func TestAxisTypesSplitIsTracked(t *testing.T) {
	types := readTextFile(t, "core/axis_types.go")
	for _, want := range []string{
		"type AxisSide uint8",
		"const (",
		"AxisBottom AxisSide = iota",
		"type TickLabelStyle struct",
		"type TickLevel struct",
		"type TickDirection uint8",
		"type AxisSpinePositionMode uint8",
		"type Axis struct",
		"func NewXAxis() *Axis",
		"func NewYAxis() *Axis",
	} {
		if !strings.Contains(types, want) {
			t.Fatalf("axis_types.go missing %q", want)
		}
	}

	axis := readTextFile(t, "core/axis.go")
	for _, moved := range []string{
		"type AxisSide uint8",
		"type TickLabelStyle struct",
		"type TickLevel struct",
		"type TickDirection uint8",
		"type AxisSpinePositionMode uint8",
		"type Axis struct",
		"func NewXAxis() *Axis",
		"func NewYAxis() *Axis",
	} {
		if strings.Contains(axis, moved) {
			t.Fatalf("core/axis.go still contains moved type/constructor item %q", moved)
		}
	}
}

func TestAxisSpineSplitIsTracked(t *testing.T) {
	spine := readTextFile(t, "core/axis_spine.go")
	for _, want := range []string{
		"func (a *Axis) drawSpine(r render.Renderer, ctx *DrawContext)",
		"func spinePixelEndpoints(side AxisSide, px geom.Rect, contexts ...*DrawContext) (geom.Pt, geom.Pt)",
		"func snapDisplayX(x float64) float64",
		"func snapDisplayY(y float64, heights ...float64) float64",
		"func figureSnapHeight(ctx *DrawContext) float64",
		"func DrawFrame(r render.Renderer, ctx *DrawContext, ref *Axis, drawTop, drawRight bool)",
		"func getSpinePosition(axis *Axis, ctx *DrawContext) float64",
		"func axisSpinePixelEndpoints(axis *Axis, ctx *DrawContext, px geom.Rect) (geom.Pt, geom.Pt)",
		"func (a *Axis) SetLineStyle(lineCap render.LineCap, join render.LineJoin, dashes ...float64)",
		"func (a *Axis) SetSpinePositionData(value float64)",
		"func (a *Axis) ResetSpinePosition()",
	} {
		if !strings.Contains(spine, want) {
			t.Fatalf("axis_spine.go missing %q", want)
		}
	}

	axis := readTextFile(t, "core/axis.go")
	for _, moved := range []string{
		"func (a *Axis) drawSpine(",
		"func spinePixelEndpoints(",
		"func snapDisplayX(",
		"func snapDisplayY(",
		"func figureSnapHeight(",
		"func DrawFrame(",
		"func getSpinePosition(",
		"func axisSpinePixelEndpoints(",
		"func (a *Axis) SetLineStyle(",
		"func (a *Axis) SetSpinePositionData(",
		"func (a *Axis) ResetSpinePosition(",
	} {
		if strings.Contains(axis, moved) {
			t.Fatalf("core/axis.go still contains moved spine helper %q", moved)
		}
	}
}

func TestAxisTicksSplitIsTracked(t *testing.T) {
	ticks := readTextFile(t, "core/axis_ticks.go")
	for _, want := range []string{
		"func (a *Axis) DrawTicks(r render.Renderer, ctx *DrawContext)",
		"func (a *Axis) drawTicks(r render.Renderer, ctx *DrawContext, ticks []float64, isXAxis bool)",
		"func (a *Axis) drawMinorTicks(r render.Renderer, ctx *DrawContext, ticks []float64, isXAxis bool)",
		"func (a *Axis) drawSingleTick(r render.Renderer, ctx *DrawContext, tickValue, tickSize, lineWidth float64, stroke render.Color, isXAxis bool)",
		"func (a *Axis) majorTickTargetCount() int",
		"func (a *Axis) minorTickTargetCount() int",
		"func (a *Axis) majorTickTargetCountForContext(ctx *DrawContext, isXAxis bool) int",
		"func (a *Axis) minorTickTargetCountForContext(ctx *DrawContext, isXAxis bool) int",
		"func visibleTicks(ticks []float64, minVal, maxVal float64) []float64",
		"func axisTickDisplayPoint(a *Axis, ctx *DrawContext, tickValue float64, isXAxis bool, spineValue float64) geom.Pt",
		"func axisTickSegment(axis *Axis, spine geom.Pt, tickSize float64, isXAxis bool) (geom.Pt, geom.Pt)",
		"func ParseTickDirection(direction string) (TickDirection, error)",
		"func axisStrokePaint(a *Axis, ctx *DrawContext, forTicks bool) render.Paint",
		"func (a *Axis) tickColor() render.Color",
		"func (a *Axis) minorTickColor() render.Color",
		"func (a *Axis) tickLineWidth() float64",
		"func (a *Axis) minorTickLineWidth() float64",
		"func tickLevelSize(level TickLevel, fallback float64) float64",
		"func (a *Axis) minorTickSize() float64",
		"func (a *Axis) AddTickLevel(level TickLevel)",
		"func (a *Axis) ClearTickLevels()",
	} {
		if !strings.Contains(ticks, want) {
			t.Fatalf("axis_ticks.go missing %q", want)
		}
	}

	axis := readTextFile(t, "core/axis.go")
	for _, moved := range []string{
		"func (a *Axis) DrawTicks(",
		"func (a *Axis) drawTicks(",
		"func (a *Axis) drawMinorTicks(",
		"func (a *Axis) drawSingleTick(",
		"func (a *Axis) majorTickTargetCount(",
		"func (a *Axis) minorTickTargetCount(",
		"func (a *Axis) majorTickTargetCountForContext(",
		"func (a *Axis) minorTickTargetCountForContext(",
		"func visibleTicks(",
		"func axisTickDisplayPoint(",
		"func axisTickSegment(",
		"func ParseTickDirection(",
		"func axisStrokePaint(",
		"func (a *Axis) tickColor(",
		"func (a *Axis) minorTickColor(",
		"func (a *Axis) tickLineWidth(",
		"func (a *Axis) minorTickLineWidth(",
		"func tickLevelSize(",
		"func (a *Axis) minorTickSize(",
		"func (a *Axis) AddTickLevel(",
		"func (a *Axis) ClearTickLevels(",
	} {
		if strings.Contains(axis, moved) {
			t.Fatalf("core/axis.go still contains moved tick helper %q", moved)
		}
	}
}

func TestAxisTickLabelsSplitIsTracked(t *testing.T) {
	labels := readTextFile(t, "core/axis_ticklabels.go")
	for _, want := range []string{
		"func (a *Axis) DrawTickLabels(r render.Renderer, ctx *DrawContext)",
		"func (a *Axis) drawTickLabels(r render.Renderer, ctx *DrawContext, ticks []float64, formatter ticker.Formatter, style TickLabelStyle, tickSize float64, labelColor render.Color, isXAxis bool)",
		"func (a *Axis) drawTickOffsetText(r render.Renderer, ctx *DrawContext, ticks []float64, formatter ticker.Formatter, style TickLabelStyle, tickSize float64, labelColor render.Color, isXAxis bool)",
		"func tickLabelFontSize(a *Axis, ctx *DrawContext) float64",
		"func tickLabelPadForAxisSize(a *Axis, tickSize float64, style TickLabelStyle, ctx *DrawContext) float64",
		"func tickLabelOrigin(a *Axis, ctx *DrawContext, tickValue float64, layout singleLineTextLayout, labelPadPx float64, style TickLabelStyle, isXAxis bool) (geom.Pt, bool)",
		"func textInkRect(origin geom.Pt, layout singleLineTextLayout) (geom.Rect, bool)",
		"func axisTickLabelBounds(a *Axis, r render.Renderer, ctx *DrawContext) (geom.Rect, bool)",
		"func tickLabelBoundsForLevel(a *Axis, r render.Renderer, ctx *DrawContext, ticks []float64, formatter ticker.Formatter, style TickLabelStyle, tickSize float64, isXAxis bool) (geom.Rect, bool)",
		"func tickLabelDisplayRect(side AxisSide, style TickLabelStyle, isXAxis bool, origin geom.Pt, layout singleLineTextLayout, lineHeight float64) (geom.Rect, bool)",
		"func alignedTextLayoutRect(anchor geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign, lineHeight float64) (geom.Rect, bool)",
		"func tickLabelDrawOriginFromP(p geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign, angle float64, anchorMode bool) geom.Pt",
		"func rotatedTextBackendAnchorFromP(p geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign, angle float64, anchorMode bool) geom.Pt",
		"func resolvedTickLabelAlignments(side AxisSide, style TickLabelStyle, isXAxis bool) (TextAlign, TextVerticalAlign)",
		"func defaultTickLabelStyle() TickLabelStyle",
		"func normalizeTickLabelStyle(style TickLabelStyle) TickLabelStyle",
		"func styleOrCurrentRC(ctx *DrawContext) style.RC",
	} {
		if !strings.Contains(labels, want) {
			t.Fatalf("axis_ticklabels.go missing %q", want)
		}
	}

	axis := readTextFile(t, "core/axis.go")
	for _, moved := range []string{
		"func (a *Axis) DrawTickLabels(",
		"func (a *Axis) drawTickLabels(",
		"func (a *Axis) drawTickOffsetText(",
		"func tickLabelFontSize(",
		"func tickLabelPadForAxisSize(",
		"func tickLabelOrigin(",
		"func textInkRect(",
		"func axisTickLabelBounds(",
		"func tickLabelBoundsForLevel(",
		"func tickLabelDisplayRect(",
		"func alignedTextLayoutRect(",
		"func tickLabelDrawOriginFromP(",
		"func rotatedTextBackendAnchorFromP(",
		"func resolvedTickLabelAlignments(",
		"func defaultTickLabelStyle(",
		"func normalizeTickLabelStyle(",
		"func styleOrCurrentRC(",
	} {
		if strings.Contains(axis, moved) {
			t.Fatalf("core/axis.go still contains moved tick-label helper %q", moved)
		}
	}
}

func TestAxisPolarSplitIsTracked(t *testing.T) {
	polar := readTextFile(t, "core/axis_polar.go")
	for _, want := range []string{
		"func (a *Axis) drawPolarSpine(r render.Renderer, ctx *DrawContext)",
		"func (a *Axis) drawPolarTicks(r render.Renderer, ctx *DrawContext)",
		"func (a *Axis) drawPolarThetaTicks(r render.Renderer, ctx *DrawContext, ticks []float64, tickSize, lineWidth float64)",
		"func (a *Axis) drawPolarRadialTicks(r render.Renderer, ctx *DrawContext, ticks []float64, tickSize, lineWidth float64)",
		"func (a *Axis) drawPolarTickLabels(r render.Renderer, ctx *DrawContext)",
		"func (a *Axis) drawPolarThetaTickLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, ticks []float64, formatter ticker.Formatter, style TickLabelStyle, tickSize float64)",
		"func polarThetaTickLabelPadPx(a *Axis, tickSize float64, style TickLabelStyle, ctx *DrawContext) float64",
		"func (a *Axis) drawPolarRadialTickLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, ticks []float64, formatter ticker.Formatter, style TickLabelStyle, tickSize float64)",
		"func (a *Axis) polarTickLabelBounds(r render.Renderer, ctx *DrawContext) (geom.Rect, bool)",
		"func (a *Axis) polarTickLabelBoundsForLevel(r render.Renderer, ctx *DrawContext, ticks []float64, formatter ticker.Formatter, style TickLabelStyle, tickSize float64) (geom.Rect, bool)",
	} {
		if !strings.Contains(polar, want) {
			t.Fatalf("axis_polar.go missing %q", want)
		}
	}

	axis := readTextFile(t, "core/axis.go")
	for _, moved := range []string{
		"func (a *Axis) drawPolarSpine(",
		"func (a *Axis) drawPolarTicks(",
		"func (a *Axis) drawPolarThetaTicks(",
		"func (a *Axis) drawPolarRadialTicks(",
		"func (a *Axis) drawPolarTickLabels(",
		"func (a *Axis) drawPolarThetaTickLabels(",
		"func polarThetaTickLabelPadPx(",
		"func (a *Axis) drawPolarRadialTickLabels(",
		"func (a *Axis) polarTickLabelBounds(",
		"func (a *Axis) polarTickLabelBoundsForLevel(",
	} {
		if strings.Contains(axis, moved) {
			t.Fatalf("core/axis.go still contains moved polar helper %q", moved)
		}
	}
}
