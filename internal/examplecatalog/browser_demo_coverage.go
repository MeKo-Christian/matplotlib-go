package examplecatalog

// BrowserDemoCoverageStatus records how a browser demo coverage row is handled.
type BrowserDemoCoverageStatus string

const (
	BrowserDemoActive        BrowserDemoCoverageStatus = "active"
	BrowserDemoPlanned       BrowserDemoCoverageStatus = "planned"
	BrowserDemoReferenceOnly BrowserDemoCoverageStatus = "reference-only"
)

// BrowserDemoCoverage records Phase 9A.4 reconciliation between the curated
// browser demo catalog, Python web reference modules, and CLI-only showcases.
type BrowserDemoCoverage struct {
	ID              string
	Title           string
	Status          BrowserDemoCoverageStatus
	Action          string
	Rationale       string
	ReferenceModule string
	ActiveWebDemoID string
	CatalogIDs      []string
}

var browserDemoCoverageRows = []BrowserDemoCoverage{
	{
		ID:              "webref-annotations",
		Title:           "Annotations Web Reference Module",
		Status:          BrowserDemoActive,
		Action:          "expose through the annotations web demo backed by the annotation/legend/offsetbox gallery",
		Rationale:       "Annotation behavior is user-facing and already has catalog showcases; the browser entry now uses the broad catalog-backed annotation gallery.",
		ReferenceModule: "annotations",
		ActiveWebDemoID: "annotations",
		CatalogIDs:      []string{"annotation_legend_offsetbox_gallery", "annotation_composition"},
	},
	{
		ID:              "webref-bars",
		Title:           "Bars Web Reference Module",
		Status:          BrowserDemoActive,
		Action:          "expose through the bars web demo backed by the bar variants gallery",
		Rationale:       "Bar variants are fixture-rich and now share a catalog-backed browser inspection path.",
		ReferenceModule: "bars",
		ActiveWebDemoID: "bars",
		CatalogIDs:      []string{"bar_variants", "bar_basic", "bar_horizontal", "bar_grouped"},
	},
	{
		ID:              "webref-errorbars",
		Title:           "Errorbars Web Reference Module",
		Status:          BrowserDemoActive,
		Action:          "expose through the errorbars web demo backed by the errorbar showcase",
		Rationale:       "Errorbar caps, asymmetric errors, and marker styling are parity-sensitive and now browser-visible.",
		ReferenceModule: "errorbars",
		ActiveWebDemoID: "errorbars",
		CatalogIDs:      []string{"errorbar_basic"},
	},
	{
		ID:              "webref-fills",
		Title:           "Fills Web Reference Module",
		Status:          BrowserDemoActive,
		Action:          "expose through the fills web demo backed by the fill variants gallery",
		Rationale:       "fill_between and stacked fill are important examples and now share a catalog-backed browser entry.",
		ReferenceModule: "fills",
		ActiveWebDemoID: "fills",
		CatalogIDs:      []string{"fill_variants", "fill_basic", "fill_between", "fill_stacked"},
	},
	{
		ID:              "webref-heatmap",
		Title:           "Heatmap Web Reference Module",
		Status:          BrowserDemoActive,
		Action:          "expose through the heatmap web demo backed by the heatmap image showcase",
		Rationale:       "The heatmap showcase exists as a catalog case and now has a direct browser inspection path alongside matrix helpers.",
		ReferenceModule: "heatmap",
		ActiveWebDemoID: "heatmap",
		CatalogIDs:      []string{"image_heatmap", "arrays_showcase"},
	},
	{
		ID:              "webref-histogram",
		Title:           "Histogram Web Reference Module",
		Status:          BrowserDemoActive,
		Action:          "expose through the histogram web demo backed by the histogram variants gallery",
		Rationale:       "Histogram density, binning, and weighted 2D cases now have browser inspection beyond the basic counts showcase.",
		ReferenceModule: "histogram",
		ActiveWebDemoID: "histogram",
		CatalogIDs:      []string{"histogram_variants", "hist_basic", "hist_density", "hist_strategies", "hist2d_weighted_density"},
	},
	{
		ID:              "webref-lines",
		Title:           "Lines Web Reference Module",
		Status:          BrowserDemoActive,
		Action:          "expose through the lines web demo backed by the line and marker style gallery",
		Rationale:       "Line dash, cap, join, and marker parity is foundational and now browser-visible.",
		ReferenceModule: "lines",
		ActiveWebDemoID: "lines",
		CatalogIDs:      []string{"lines_markers_gallery", "basic_line", "dashes", "joins_caps", "scatter_marker_types"},
	},
	{
		ID:              "webref-patches",
		Title:           "Patches Web Reference Module",
		Status:          BrowserDemoActive,
		Action:          "expose through the patches web demo backed by the patch showcase",
		Rationale:       "Patch and hatch behavior has parity fixtures and now has an active browser entry.",
		ReferenceModule: "patches",
		ActiveWebDemoID: "patches",
		CatalogIDs:      []string{"patch_showcase"},
	},
	{
		ID:              "showcase-patch_showcase",
		Title:           "Patch Showcase Browser Coverage",
		Status:          BrowserDemoActive,
		Action:          "expose through the patches web demo",
		Rationale:       "Patch primitives, hatches, fancy arrows, path patches, and fancy boxes are user-facing and now browser-visible.",
		ActiveWebDemoID: "patches",
		CatalogIDs:      []string{"patch_showcase"},
	},
	{
		ID:              "webref-radialforce",
		Title:           "Radial Force Web Reference Module",
		Status:          BrowserDemoReferenceOnly,
		Action:          "keep as reference-only until it is promoted to a catalog case",
		Rationale:       "This module has no current catalog case, so wiring it directly would break the catalog-driven source-of-truth rule.",
		ReferenceModule: "radialforce",
	},
	{
		ID:              "webref-scatter",
		Title:           "Scatter Web Reference Module",
		Status:          BrowserDemoActive,
		Action:          "expose through the scatter web demo backed by the advanced scatter gallery",
		Rationale:       "Scatter has rich parity fixtures and now uses the broad catalog-backed scatter gallery in the browser.",
		ReferenceModule: "scatter",
		ActiveWebDemoID: "scatter",
		CatalogIDs:      []string{"scatter_gallery", "scatter_basic", "scatter_marker_types", "scatter_advanced", "large_scatter"},
	},
	{
		ID:              "webref-subplots",
		Title:           "Subplots Web Reference Module",
		Status:          BrowserDemoActive,
		Action:          "expose through the subplots web demo backed by the figure-label composition showcase",
		Rationale:       "Subplot and figure-label behavior is core gallery material and now has a catalog-backed browser entry.",
		ReferenceModule: "subplots",
		ActiveWebDemoID: "subplots",
		CatalogIDs:      []string{"figure_labels_composition", "gridspec_composition"},
	},
	{
		ID:        "showcase-basic_line",
		Title:     "Basic Line Browser Coverage",
		Status:    BrowserDemoPlanned,
		Action:    "include in a line styles browser demo",
		Rationale: "The canonical line showcase is CLI-only while a lines web reference module already exists.",
		CatalogIDs: []string{
			"basic_line",
		},
	},
	{
		ID:        "showcase-basic_line_labels",
		Title:     "Basic Line Labels Browser Coverage",
		Status:    BrowserDemoPlanned,
		Action:    "include in a line styles or text layout browser demo",
		Rationale: "The labeled line showcase is CLI-only and isolates axis-label layout behavior for browser inspection.",
		CatalogIDs: []string{
			"basic_line_labels",
		},
	},
	{
		ID:        "showcase-dashes",
		Title:     "Dash Patterns Browser Coverage",
		Status:    BrowserDemoPlanned,
		Action:    "include in a line styles browser demo",
		Rationale: "Dash arrays, cap styles, and legend samples are useful browser-inspection targets.",
		CatalogIDs: []string{
			"dashes",
		},
	},
	{
		ID:              "showcase-lines_markers_gallery",
		Title:           "Line and Marker Style Gallery Browser Coverage",
		Status:          BrowserDemoActive,
		Action:          "expose through the lines web demo",
		Rationale:       "The combined gallery (dashes, joins/caps, marker grid, multi-series legend) is the natural browser panel for the Line2D stroke and marker surface.",
		ActiveWebDemoID: "lines",
		CatalogIDs:      []string{"lines_markers_gallery"},
	},
	{
		ID:         "showcase-scatter_basic",
		Title:      "Basic Scatter Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in an advanced scatter browser demo",
		Rationale:  "The basic scatter showcase should be the baseline panel for richer scatter browser coverage.",
		CatalogIDs: []string{"scatter_basic"},
	},
	{
		ID:              "showcase-scatter_gallery",
		Title:           "Advanced Scatter Gallery Browser Coverage",
		Status:          BrowserDemoActive,
		Action:          "expose through the scatter web demo",
		Rationale:       "The combined gallery (colormapped, variable size, alpha blending, marker families) is the natural browser panel for advanced scatter behavior.",
		ActiveWebDemoID: "scatter",
		CatalogIDs:      []string{"scatter_gallery"},
	},
	{
		ID:         "showcase-bar_basic",
		Title:      "Basic Bar Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a bar variants browser demo",
		Rationale:  "The basic bar showcase is CLI-only while bar reference modules already exist.",
		CatalogIDs: []string{"bar_basic"},
	},
	{
		ID:              "showcase-bar_variants",
		Title:           "Bar Variants Gallery Browser Coverage",
		Status:          BrowserDemoActive,
		Action:          "expose through the bars web demo",
		Rationale:       "The combined gallery (vertical+labels, horizontal, grouped, stacked+labels) is the natural browser panel for bar variants.",
		ActiveWebDemoID: "bars",
		CatalogIDs:      []string{"bar_variants"},
	},
	{
		ID:         "showcase-fill_basic",
		Title:      "Fill Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a fill variants browser demo",
		Rationale:  "Fill-to-baseline should anchor a browser gallery for fill_between and stacked fills.",
		CatalogIDs: []string{"fill_basic"},
	},
	{
		ID:              "showcase-fill_variants",
		Title:           "Fill Variants Gallery Browser Coverage",
		Status:          BrowserDemoActive,
		Action:          "expose through the fills web demo",
		Rationale:       "The combined gallery (fill_between, fill_betweenx, stacked, alpha overlap) is the natural browser panel for the fill surface.",
		ActiveWebDemoID: "fills",
		CatalogIDs:      []string{"fill_variants"},
	},
	{
		ID:              "showcase-errorbar_basic",
		Title:           "Errorbar Browser Coverage",
		Status:          BrowserDemoActive,
		Action:          "expose through the errorbars web demo",
		Rationale:       "Errorbar marker/cap behavior is visually sensitive and now browser-visible.",
		ActiveWebDemoID: "errorbars",
		CatalogIDs:      []string{"errorbar_basic"},
	},
	{
		ID:         "showcase-multi_series_basic",
		Title:      "Multi-Series Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a lines or annotation/legend browser demo",
		Rationale:  "Multi-series color cycling and legends should be inspectable in the browser.",
		CatalogIDs: []string{"multi_series_basic"},
	},
	{
		ID:         "showcase-hist_basic",
		Title:      "Histogram Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a histogram variants browser demo",
		Rationale:  "The basic histogram showcase should become the baseline panel for histogram variants.",
		CatalogIDs: []string{"hist_basic"},
	},
	{
		ID:              "showcase-histogram_variants",
		Title:           "Histogram Variants Gallery Browser Coverage",
		Status:          BrowserDemoActive,
		Action:          "expose through the histogram web demo",
		Rationale:       "The combined gallery (counts, density, cumulative, multiple) is the natural browser panel for histogram variants.",
		ActiveWebDemoID: "histogram",
		CatalogIDs:      []string{"histogram_variants"},
	},
	{
		ID:         "showcase-boxplot_basic",
		Title:      "Boxplot Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the statistics browser demo or a statistics-depth browser group",
		Rationale:  "Boxplot is user-facing but not currently mapped to an active browser demo.",
		CatalogIDs: []string{"boxplot_basic"},
	},
	{
		ID:              "showcase-image_heatmap",
		Title:           "Heatmap Browser Coverage",
		Status:          BrowserDemoActive,
		Action:          "expose through the heatmap web demo",
		Rationale:       "The heatmap showcase should share browser coverage with matrix helpers and colorbars.",
		ActiveWebDemoID: "heatmap",
		CatalogIDs:      []string{"image_heatmap"},
	},
	{
		ID:         "showcase-image_variants_gallery",
		Title:      "Image Variants Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the matrix browser demo or a dedicated image sampling gallery",
		Rationale:  "Interpolation, alpha, MatShow, and spy behavior is visually sensitive and should be browser-inspectable after the CLI showcase.",
		CatalogIDs: []string{"image_variants_gallery"},
	},
	{
		ID:         "showcase-mathtext_gallery",
		Title:      "MathText Gallery Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a text and annotation browser demo",
		Rationale:  "MathText layout is visually sensitive and now has a CLI showcase that should be browser-inspectable.",
		CatalogIDs: []string{"mathtext_gallery"},
	},
	{
		ID:         "showcase-text_layout_gallery",
		Title:      "Text Layout Gallery Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a text and annotation browser demo",
		Rationale:  "Alignment, rotation, wrapping, and bbox text behavior are common browser-inspection targets.",
		CatalogIDs: []string{"text_layout_gallery"},
	},
	{
		ID:         "showcase-ticks_scales_formatters_gallery",
		Title:      "Ticks, Scales, and Formatters Gallery Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the axes or units browser demo",
		Rationale:  "Locator, formatter, scale, date, category, and custom-unit behavior should be browser-inspectable after the CLI showcase.",
		CatalogIDs: []string{"ticks_scales_formatters_gallery"},
	},
	{
		ID:              "showcase-annotation_legend_offsetbox_gallery",
		Title:           "Annotation, Legend, and Offset Box Gallery Browser Coverage",
		Status:          BrowserDemoActive,
		Action:          "expose through the annotations web demo",
		Rationale:       "Annotation arrows, legend handlers, and anchored offset boxes are user-facing and visually sensitive.",
		ActiveWebDemoID: "annotations",
		CatalogIDs:      []string{"annotation_legend_offsetbox_gallery"},
	},
	{
		ID:         "showcase-named_colors_gallery",
		Title:      "Named Colors Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a color browser demo",
		Rationale:  "Named color swatches are user-facing and useful as a browser-visible compatibility reference.",
		CatalogIDs: []string{"named_colors_gallery"},
	},
	{
		ID:         "showcase-colormap_families_gallery",
		Title:      "Colormap Families Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a color browser demo",
		Rationale:  "Colormap family strips are a browsable compatibility reference and should not remain CLI-only long term.",
		CatalogIDs: []string{"colormap_families_gallery"},
	},
	{
		ID:              "showcase-figure_labels_composition",
		Title:           "Figure Labels Browser Coverage",
		Status:          BrowserDemoActive,
		Action:          "expose through the subplots web demo",
		Rationale:       "Figure-level labels and figure legends are core composition behavior.",
		ActiveWebDemoID: "subplots",
		CatalogIDs:      []string{"figure_labels_composition"},
	},
	{
		ID:         "showcase-colorbar_composition",
		Title:      "Colorbar Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the matrix or composition browser demo",
		Rationale:  "The colorbar showcase should be browser-visible alongside norm and extension variants.",
		CatalogIDs: []string{"colorbar_composition"},
	},
	{
		ID:         "showcase-colorbar_variants_gallery",
		Title:      "Colorbar Variants Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the matrix or color browser demo",
		Rationale:  "Norm and extension colorbars benefit from browser-side visual inspection after the CLI showcase.",
		CatalogIDs: []string{"colorbar_variants_gallery"},
	},
	{
		ID:         "showcase-mixed_raster_vector",
		Title:      "Mixed Raster/Vector Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a backend output browser demo or export gallery",
		Rationale:  "Mixed raster/vector output is user-facing and has SVG/PDF artifact coverage, but browser inspection still needs an export-focused grouping.",
		CatalogIDs: []string{"mixed_raster_vector"},
	},
	{
		ID:         "showcase-annotation_composition",
		Title:      "Annotation Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in an annotation browser demo",
		Rationale:  "Annotation arrows, anchored labels, and mixed coordinate placement are currently CLI-only.",
		CatalogIDs: []string{"annotation_composition"},
	},
	{
		ID:         "showcase-geo_aitoff_axes",
		Title:      "Aitoff Projection Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the projections browser demo",
		Rationale:  "The projections browser currently represents only part of the geographic projection family.",
		CatalogIDs: []string{"geo_aitoff_axes"},
	},
	{
		ID:         "showcase-radar_basic",
		Title:      "Radar Projection Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the projections browser demo",
		Rationale:  "Radar projection behavior is user-facing but not active in the browser gallery.",
		CatalogIDs: []string{"radar_basic"},
	},
	{
		ID:         "showcase-skewt_basic",
		Title:      "Skew-T Projection Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the projections browser demo or a specialized projection group",
		Rationale:  "Skew-T projection behavior is user-facing and currently CLI-only.",
		CatalogIDs: []string{"skewt_basic"},
	},
	{
		ID:         "showcase-mplot3d_terrain",
		Title:      "mplot3d Terrain Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a 3D browser demo",
		Rationale:  "The 3D toolkit has many fixtures and one terrain showcase, but no active browser demo.",
		CatalogIDs: []string{"mplot3d_terrain"},
	},
	{
		ID:         "showcase-mplot3d_gallery",
		Title:      "mplot3d Gallery Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a 3D browser demo",
		Rationale:  "The broad 3D gallery is user-facing but remains CLI-only until the browser gallery gets a dedicated 3D grouping.",
		CatalogIDs: []string{"mplot3d_gallery"},
	},
	{
		ID:         "showcase-projection_toolkit_gallery",
		Title:      "Projection Toolkit Gallery Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "promote into the projections browser demo",
		Rationale:  "The grouped projection/toolkit gallery is the natural browser panel for polar, geographic, radar, skew-T, axisartist, and axes_grid1 behavior.",
		CatalogIDs: []string{"projection_toolkit_gallery"},
	},
	{
		ID:         "showcase-unstructured_showcase",
		Title:      "Unstructured Triangulation Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the mesh browser demo or a triangulation browser group",
		Rationale:  "Triangulation examples are CLI-only despite being central to contour parity.",
		CatalogIDs: []string{"unstructured_showcase"},
	},
	{
		ID:         "showcase-triangulation_gallery",
		Title:      "Triangulation Gallery Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the mesh browser demo or a dedicated triangulation browser group",
		Rationale:  "The focused triangulation gallery groups triplot, tripcolor, tricontour, tricontourf, and masked mesh behavior for browser inspection.",
		CatalogIDs: []string{"triangulation_gallery"},
	},
	{
		ID:         "showcase-axisartist_showcase",
		Title:      "AxisArtist Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a toolkit browser demo",
		Rationale:  "AxisArtist behavior is advanced but user-facing and should be inspectable.",
		CatalogIDs: []string{"axisartist_showcase"},
	},
	{
		ID:         "showcase-axes_grid1_showcase",
		Title:      "Axes Grid1 Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a toolkit browser demo",
		Rationale:  "Axes grid helpers are layout-heavy and benefit from browser inspection.",
		CatalogIDs: []string{"axes_grid1_showcase"},
	},
	{
		ID:         "showcase-widgets_gallery",
		Title:      "Widgets Gallery Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a widgets or interactive-controls browser demo",
		Rationale:  "The widgets gallery is catalog-visible for static widget/selector coverage, but browser event-loop and interaction coverage should be promoted separately.",
		CatalogIDs: []string{"widgets_gallery"},
	},
	{
		ID:              "showcase-animation_gallery",
		Title:           "Animation Gallery Browser Coverage",
		Status:          BrowserDemoActive,
		Action:          "expose as the animation web demo while timer-driven browser playback remains covered by the event-loop and writer tests",
		Rationale:       "The animation gallery is catalog-visible for setup, deterministic stepping, and static preview, and the WASM demo catalog now exposes that preview directly.",
		ActiveWebDemoID: "animation",
		CatalogIDs:      []string{"animation_gallery"},
	},
}

// BrowserDemoCoverageRows returns the Phase 9A.4 browser coverage reconciliation rows.
func BrowserDemoCoverageRows() []BrowserDemoCoverage {
	out := make([]BrowserDemoCoverage, len(browserDemoCoverageRows))
	copy(out, browserDemoCoverageRows)
	for i := range out {
		out[i].CatalogIDs = append([]string(nil), out[i].CatalogIDs...)
	}
	return out
}

// LookupBrowserDemoCoverage finds a Phase 9A.4 browser coverage row by stable ID.
func LookupBrowserDemoCoverage(id string) (BrowserDemoCoverage, bool) {
	for _, row := range BrowserDemoCoverageRows() {
		if row.ID == id {
			return row, true
		}
	}
	return BrowserDemoCoverage{}, false
}
