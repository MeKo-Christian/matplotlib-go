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
		Status:          BrowserDemoPlanned,
		Action:          "wire into an annotation browser demo or fold into a future annotation/legend gallery",
		Rationale:       "Annotation behavior is user-facing and already has a catalog showcase; the Python web reference should not remain orphaned.",
		ReferenceModule: "annotations",
		CatalogIDs:      []string{"annotation_composition"},
	},
	{
		ID:              "webref-bars",
		Title:           "Bars Web Reference Module",
		Status:          BrowserDemoPlanned,
		Action:          "wire into a bar variants browser demo",
		Rationale:       "Bar variants are fixture-rich and currently thin in the browser gallery.",
		ReferenceModule: "bars",
		CatalogIDs:      []string{"bar_basic", "bar_horizontal", "bar_grouped"},
	},
	{
		ID:              "webref-errorbars",
		Title:           "Errorbars Web Reference Module",
		Status:          BrowserDemoPlanned,
		Action:          "wire into an errorbar browser demo or a statistics/uncertainty group",
		Rationale:       "Errorbar caps, asymmetric errors, and marker styling are parity-sensitive and should be browser-visible.",
		ReferenceModule: "errorbars",
		CatalogIDs:      []string{"errorbar_basic"},
	},
	{
		ID:              "webref-fills",
		Title:           "Fills Web Reference Module",
		Status:          BrowserDemoPlanned,
		Action:          "wire into a fill variants browser demo",
		Rationale:       "fill_between and stacked fill are important examples that are currently fixture-heavy.",
		ReferenceModule: "fills",
		CatalogIDs:      []string{"fill_basic", "fill_between", "fill_stacked"},
	},
	{
		ID:              "webref-heatmap",
		Title:           "Heatmap Web Reference Module",
		Status:          BrowserDemoPlanned,
		Action:          "wire into the matrix/image browser family",
		Rationale:       "The heatmap showcase exists as a CLI example and should share browser coverage with matrix/image helpers.",
		ReferenceModule: "heatmap",
		CatalogIDs:      []string{"image_heatmap", "arrays_showcase"},
	},
	{
		ID:              "webref-histogram",
		Title:           "Histogram Web Reference Module",
		Status:          BrowserDemoPlanned,
		Action:          "wire into a histogram variants browser demo",
		Rationale:       "Histogram density, binning, and weighted 2D cases need browser inspection beyond the basic counts showcase.",
		ReferenceModule: "histogram",
		CatalogIDs:      []string{"hist_basic", "hist_density", "hist_strategies", "hist2d_weighted_density"},
	},
	{
		ID:              "webref-lines",
		Title:           "Lines Web Reference Module",
		Status:          BrowserDemoPlanned,
		Action:          "wire into a line style and marker browser demo",
		Rationale:       "Line dash, cap, join, and marker parity is foundational and should be browser-visible.",
		ReferenceModule: "lines",
		CatalogIDs:      []string{"basic_line", "dashes", "joins_caps", "scatter_marker_types"},
	},
	{
		ID:              "webref-patches",
		Title:           "Patches Web Reference Module",
		Status:          BrowserDemoPlanned,
		Action:          "wire into a patch and hatch browser demo",
		Rationale:       "Patch and hatch behavior has parity fixtures but no active browser entry.",
		ReferenceModule: "patches",
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
		Status:          BrowserDemoPlanned,
		Action:          "wire into an advanced scatter browser demo",
		Rationale:       "Scatter has rich parity fixtures but only a basic user-facing showcase.",
		ReferenceModule: "scatter",
		CatalogIDs:      []string{"scatter_basic", "scatter_marker_types", "scatter_advanced", "large_scatter"},
	},
	{
		ID:              "webref-subplots",
		Title:           "Subplots Web Reference Module",
		Status:          BrowserDemoPlanned,
		Action:          "wire into the composition browser family or a layout/subplots browser demo",
		Rationale:       "Subplot and figure-label behavior is core gallery material and should not remain an unused reference module.",
		ReferenceModule: "subplots",
		CatalogIDs:      []string{"gridspec_composition", "figure_labels_composition"},
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
		ID:         "showcase-scatter_basic",
		Title:      "Basic Scatter Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in an advanced scatter browser demo",
		Rationale:  "The basic scatter showcase should be the baseline panel for richer scatter browser coverage.",
		CatalogIDs: []string{"scatter_basic"},
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
		ID:         "showcase-fill_basic",
		Title:      "Fill Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in a fill variants browser demo",
		Rationale:  "Fill-to-baseline should anchor a browser gallery for fill_between and stacked fills.",
		CatalogIDs: []string{"fill_basic"},
	},
	{
		ID:         "showcase-errorbar_basic",
		Title:      "Errorbar Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in an errorbar browser demo",
		Rationale:  "Errorbar marker/cap behavior is visually sensitive and currently CLI-only.",
		CatalogIDs: []string{"errorbar_basic"},
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
		ID:         "showcase-boxplot_basic",
		Title:      "Boxplot Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the statistics browser demo or a statistics-depth browser group",
		Rationale:  "Boxplot is user-facing but not currently mapped to an active browser demo.",
		CatalogIDs: []string{"boxplot_basic"},
	},
	{
		ID:         "showcase-image_heatmap",
		Title:      "Heatmap Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the matrix/image browser family",
		Rationale:  "The heatmap showcase should share browser coverage with matrix helpers and colorbars.",
		CatalogIDs: []string{"image_heatmap"},
	},
	{
		ID:         "showcase-figure_labels_composition",
		Title:      "Figure Labels Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the composition browser demo",
		Rationale:  "Figure-level labels and figure legends are core composition behavior.",
		CatalogIDs: []string{"figure_labels_composition"},
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
		ID:         "showcase-unstructured_showcase",
		Title:      "Unstructured Triangulation Browser Coverage",
		Status:     BrowserDemoPlanned,
		Action:     "include in the mesh browser demo or a triangulation browser group",
		Rationale:  "Triangulation examples are CLI-only despite being central to contour parity.",
		CatalogIDs: []string{"unstructured_showcase"},
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
