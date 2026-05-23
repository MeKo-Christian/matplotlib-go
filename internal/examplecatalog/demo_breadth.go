package examplecatalog

// DemoBreadthPriority ranks how important a missing or thin demo is for the
// examples-driven parity roadmap.
type DemoBreadthPriority string

const (
	DemoBreadthHigh   DemoBreadthPriority = "high"
	DemoBreadthMedium DemoBreadthPriority = "medium"
	DemoBreadthLow    DemoBreadthPriority = "low"
)

// DemoBreadthGap records one Phase 9A.3 gap where parity fixtures exist, but
// the user-facing examples or gallery breadth do not yet show the feature well.
type DemoBreadthGap struct {
	ID              string
	Topic           string
	Title           string
	Priority        DemoBreadthPriority
	CatalogIDs      []string
	ShowcaseIDs     []string
	WebDemoIDs      []string
	TargetFeatures  []string
	CurrentCoverage string
	Need            string
	RecommendedDemo string
}

var demoBreadthGaps = []DemoBreadthGap{
	{
		ID:         "marker-grid",
		Topic:      "lines",
		Title:      "Line and Marker Style Grid",
		Priority:   DemoBreadthHigh,
		CatalogIDs: []string{"basic_line", "dashes", "joins_caps", "scatter_marker_types"},
		ShowcaseIDs: []string{
			"basic_line",
			"dashes",
		},
		TargetFeatures:  []string{"line joins", "line caps", "dash arrays", "built-in marker grid", "marker edge/fill styling"},
		CurrentCoverage: "Basic lines and dashes are showcased; joins/caps and full marker grids are parity-only.",
		Need:            "Users need a visual gallery for Line2D stroke and marker semantics rather than discovering them through fixture names.",
		RecommendedDemo: "Promote a line-style and marker-grid showcase that combines dashes, caps, joins, marker shapes, and legend samples.",
	},
	{
		ID:         "advanced-scatter",
		Topic:      "scatter",
		Title:      "Advanced Scatter Styles",
		Priority:   DemoBreadthHigh,
		CatalogIDs: []string{"scatter_basic", "scatter_marker_types", "scatter_advanced", "large_scatter"},
		ShowcaseIDs: []string{
			"scatter_basic",
		},
		TargetFeatures:  []string{"colormapped scatter", "variable size", "alpha", "marker families", "large path collections"},
		CurrentCoverage: "Only a compact scatter showcase is user-facing; advanced marker/color/size cases are fixtures.",
		Need:            "Scatter is a core Matplotlib migration path and should show marker, scalar-mappable, and batching behavior.",
		RecommendedDemo: "Add an advanced scatter showcase with marker families, size/color arrays, alpha, colorbar, and dense collection stress.",
	},
	{
		ID:         "bar-variants",
		Topic:      "bar",
		Title:      "Bar Variants",
		Priority:   DemoBreadthHigh,
		CatalogIDs: []string{"bar_basic", "bar_horizontal", "bar_grouped", "bar_basic_tick_labels", "plot_variants"},
		ShowcaseIDs: []string{
			"bar_basic",
			"plot_variants",
		},
		WebDemoIDs:      []string{"variants"},
		TargetFeatures:  []string{"horizontal bars", "grouped bars", "stacked bars", "tick labels", "bar labels"},
		CurrentCoverage: "Basic bars and stacked bars inside plot variants are showcased; horizontal/grouped variants remain fixture-heavy.",
		Need:            "Bar chart users need one example that shows the normal categorical variants and label behavior.",
		RecommendedDemo: "Add a bar variants showcase covering vertical, horizontal, grouped, stacked, tick-label, and bar-label usage.",
	},
	{
		ID:         "fill-variants",
		Topic:      "fill",
		Title:      "Fill Between and Stacked Areas",
		Priority:   DemoBreadthHigh,
		CatalogIDs: []string{"fill_basic", "fill_between", "fill_stacked", "plot_variants"},
		ShowcaseIDs: []string{
			"fill_basic",
			"plot_variants",
		},
		WebDemoIDs:      []string{"variants"},
		TargetFeatures:  []string{"fill_between", "fill_betweenx", "stacked fill", "alpha blending", "edge ordering"},
		CurrentCoverage: "Fill-to-baseline is showcased; fill_between and stacked fill are parity cases or embedded in broader variants.",
		Need:            "The fill family needs a direct user-facing example that makes area construction and alpha semantics visible.",
		RecommendedDemo: "Add a fill variants showcase for fill_between, fill_betweenx, stacked fill, baseline fills, and translucent overlaps.",
	},
	{
		ID:         "histogram-variants",
		Topic:      "histogram",
		Title:      "Histogram Density and Binning Strategies",
		Priority:   DemoBreadthHigh,
		CatalogIDs: []string{"hist_basic", "hist_density", "hist_strategies", "hist2d_weighted_density"},
		ShowcaseIDs: []string{
			"hist_basic",
		},
		TargetFeatures:  []string{"density", "cumulative", "bin strategies", "weighted hist2d", "multiple histograms"},
		CurrentCoverage: "Histogram counts are showcased; density, strategies, and weighted 2D behavior are fixture-only.",
		Need:            "Histogram examples should show the statistical options users commonly migrate from Matplotlib.",
		RecommendedDemo: "Add a histogram variants showcase covering density, cumulative, multiple datasets, bin strategies, and weighted hist2d.",
	},
	{
		ID:              "named-colors",
		Topic:           "color",
		Title:           "Named Colors",
		Priority:        DemoBreadthMedium,
		CatalogIDs:      []string{"named_colors"},
		TargetFeatures:  []string{"CSS4 names", "Tableau tab names", "xkcd names", "single-letter colors"},
		CurrentCoverage: "Named colors have parity fixtures but no user-facing gallery.",
		Need:            "Color-name compatibility is hard to inspect without a swatch-style example.",
		RecommendedDemo: "Add a named-color swatch showcase that displays CSS4, Tableau, xkcd, grayscale, and shorthand color forms.",
	},
	{
		ID:              "colormap-families",
		Topic:           "colormap",
		Title:           "Colormap Families",
		Priority:        DemoBreadthHigh,
		CatalogIDs:      []string{"colormap_diverging", "colormap_qualitative", "colormap_cyclic", "image_heatmap"},
		ShowcaseIDs:     []string{"image_heatmap"},
		TargetFeatures:  []string{"sequential", "diverging", "qualitative", "cyclic", "reversed maps", "resampled maps"},
		CurrentCoverage: "Colormap family parity is fixture-only apart from a single heatmap showcase.",
		Need:            "Colormap breadth is an enumerable public surface and should be visually browsable.",
		RecommendedDemo: "Add a colormap family showcase with representative sequential, diverging, qualitative, cyclic, reversed, and resampled maps.",
	},
	{
		ID:          "image-variants",
		Topic:       "image",
		Title:       "Image Interpolation, Alpha, Matshow, and Spy",
		Priority:    DemoBreadthHigh,
		CatalogIDs:  []string{"image_heatmap", "imshow_clipped", "imshow_transformed", "imshow_bilinear", "imshow_bicubic", "image_alpha", "matshow_basic", "spy_marker", "spy_image", "arrays_showcase"},
		ShowcaseIDs: []string{"image_heatmap", "arrays_showcase"},
		WebDemoIDs:  []string{"matrix"},
		TargetFeatures: []string{
			"nearest interpolation",
			"bilinear interpolation",
			"bicubic interpolation",
			"alpha masks",
			"matshow",
			"spy marker mode",
			"spy image mode",
		},
		CurrentCoverage: "Heatmap and matrix helpers are showcased; interpolation, alpha, transformed images, and spy variants are mostly fixtures.",
		Need:            "Image behavior is visually sensitive, so users need demos that expose sampling, alpha, and matrix helper choices.",
		RecommendedDemo: "Add an image variants gallery with side-by-side interpolation modes, alpha, transformed/clipped images, matshow, and spy modes.",
	},
	{
		ID:          "colorbar-norms-extensions",
		Topic:       "colorbar",
		Title:       "Colorbar Norms and Extensions",
		Priority:    DemoBreadthHigh,
		CatalogIDs:  []string{"colorbar_composition", "boundarynorm_pcolormesh", "lognorm_imshow", "twoslope_norm_image", "colorbar_extensions"},
		ShowcaseIDs: []string{"colorbar_composition"},
		WebDemoIDs:  []string{"composition", "matrix"},
		TargetFeatures: []string{
			"BoundaryNorm",
			"LogNorm",
			"TwoSlopeNorm",
			"extended colorbars",
			"shared colorbars",
		},
		CurrentCoverage: "A composed colorbar example exists; norm-specific and extension behavior is mostly fixture-only.",
		Need:            "Norm and colorbar layout behavior are common parity pain points and should be inspectable in one gallery.",
		RecommendedDemo: "Add a colorbar variants showcase for BoundaryNorm, LogNorm, TwoSlopeNorm, extensions, shared axes, and labels.",
	},
	{
		ID:              "mathtext-gallery",
		Topic:           "mathtext",
		Title:           "MathText Gallery",
		Priority:        DemoBreadthHigh,
		CatalogIDs:      []string{"mathtext_basic", "mathtext_fractions", "mathtext_integrals", "mathtext_matrices", "mathtext_inline_labels"},
		TargetFeatures:  []string{"fractions", "roots", "integrals", "large operators", "matrices", "inline labels"},
		CurrentCoverage: "MathText has parity fixtures but no user-facing showcase.",
		Need:            "MathText is important enough to need a gallery that users can visually inspect without running fixture tests.",
		RecommendedDemo: "Promote a MathText showcase with inline labels plus panels for fractions, operators, roots, fences, and matrices.",
	},
	{
		ID:          "ticks-scales-formatters",
		Topic:       "axes",
		Title:       "Ticks, Scales, and Formatters",
		Priority:    DemoBreadthHigh,
		CatalogIDs:  []string{"axes_control_surface", "units_dates", "units_categories", "units_custom_converter", "lognorm_imshow", "skewt_basic"},
		ShowcaseIDs: []string{"axes_control_surface", "units_overview", "skewt_basic"},
		WebDemoIDs:  []string{"axes", "units"},
		TargetFeatures: []string{
			"major locators",
			"minor locators",
			"date formatters",
			"category formatters",
			"log scale",
			"custom converters",
		},
		CurrentCoverage: "Axes controls and units are showcased, but ticker/formatter breadth is not isolated as its own gallery.",
		Need:            "Tick and scale parity should be visible without reading broader axes examples.",
		RecommendedDemo: "Add a ticks/scales gallery covering locator and formatter variants, date/category units, log/norm cases, and custom converters.",
	},
	{
		ID:              "text-layout-gallery",
		Topic:           "text",
		Title:           "Text Layout, Rotation, and Wrapping",
		Priority:        DemoBreadthHigh,
		CatalogIDs:      []string{"text_labels_strict", "title_strict", "annotation_composition", "figure_labels_composition"},
		ShowcaseIDs:     []string{"annotation_composition", "figure_labels_composition"},
		TargetFeatures:  []string{"alignment", "rotation", "multiline text", "bbox text", "figure labels", "tick label layout"},
		CurrentCoverage: "Strict text cases are fixtures; figure labels and annotations show some text behavior but not the breadth of text layout.",
		Need:            "Text metrics are a major parity driver and need a dedicated visual gallery.",
		RecommendedDemo: "Add a text layout showcase with alignment, rotation, multiline text, bbox padding, figure labels, and tick-label variants.",
	},
	{
		ID:          "annotation-legend-offsetbox",
		Topic:       "annotation",
		Title:       "Annotations, Legends, and Offset Boxes",
		Priority:    DemoBreadthHigh,
		CatalogIDs:  []string{"annotation_composition", "figure_labels_composition", "multi_series_basic", "multi_series_color_cycle"},
		ShowcaseIDs: []string{"annotation_composition", "figure_labels_composition", "multi_series_basic"},
		WebDemoIDs:  []string{"composition"},
		TargetFeatures: []string{
			"annotation arrows",
			"offset coordinates",
			"legend placement",
			"figure legends",
			"anchored boxes",
			"proxy-like legend entries",
		},
		CurrentCoverage: "Annotations and figure labels are showcased, but legend/offsetbox breadth is mixed into broader examples.",
		Need:            "Advanced annotation and legend behavior should have a single gallery for parity inspection.",
		RecommendedDemo: "Add an annotation/legend showcase for arrow styles, coordinate modes, anchored boxes, figure legends, and legend handles.",
	},
	{
		ID:          "mplot3d-gallery",
		Topic:       "mplot3d",
		Title:       "mplot3d Gallery",
		Priority:    DemoBreadthHigh,
		CatalogIDs:  []string{"mplot3d_basic", "mplot3d_terrain", "mplot3d_plot3d", "mplot3d_scatter3d", "mplot3d_surface3d", "mplot3d_wire3d", "mplot3d_trisurf3d", "mplot3d_bar3d", "mplot3d_voxels", "mplot3d_quiver3d", "mplot3d_stem3d", "mplot3d_fill_between3d"},
		ShowcaseIDs: []string{"mplot3d_terrain"},
		TargetFeatures: []string{
			"3D line",
			"3D scatter",
			"surface",
			"wireframe",
			"trisurf",
			"bar3d",
			"voxels",
			"quiver3d",
		},
		CurrentCoverage: "mplot3d has many parity fixtures and one terrain showcase.",
		Need:            "3D parity is broad enough that a single terrain example does not expose the toolkit surface.",
		RecommendedDemo: "Add a 3D gallery with line, scatter, surface, wireframe, trisurf, bar, voxels, quiver, stem, and fill-between panels.",
	},
	{
		ID:          "projection-toolkit-gallery",
		Topic:       "projections",
		Title:       "Projection and Toolkit Gallery",
		Priority:    DemoBreadthMedium,
		CatalogIDs:  []string{"polar_axes", "geo_mollweide_axes", "geo_aitoff_axes", "geo_hammer_axes", "geo_lambert_axes", "radar_basic", "skewt_basic", "axisartist_showcase", "axes_grid1_showcase"},
		ShowcaseIDs: []string{"polar_axes", "geo_mollweide_axes", "geo_aitoff_axes", "radar_basic", "skewt_basic", "axisartist_showcase", "axes_grid1_showcase"},
		WebDemoIDs:  []string{"polar", "projections"},
		TargetFeatures: []string{
			"polar",
			"Mollweide",
			"Aitoff",
			"Hammer",
			"Lambert",
			"radar",
			"skewx",
			"axisartist",
			"axes_grid1",
		},
		CurrentCoverage: "Projection/toolkit cases have CLI showcases, but browser/gallery breadth is uneven.",
		Need:            "Advanced projection and toolkit behavior should be discoverable as a grouped gallery.",
		RecommendedDemo: "Expand the projections gallery to include geographic projections, radar, skewx, axisartist, and axes_grid1 helpers.",
	},
	{
		ID:          "unstructured-triangulation",
		Topic:       "unstructured",
		Title:       "Unstructured Triangulation and Contours",
		Priority:    DemoBreadthMedium,
		CatalogIDs:  []string{"unstructured_showcase", "mesh_contour_tri", "pcolor_flat", "pcolormesh_masked"},
		ShowcaseIDs: []string{"unstructured_showcase", "mesh_contour_tri"},
		WebDemoIDs:  []string{"mesh"},
		TargetFeatures: []string{
			"triplot",
			"tripcolor",
			"tricontour",
			"tricontourf",
			"masked meshes",
		},
		CurrentCoverage: "Unstructured and mesh examples exist, but irregular triangulation is not isolated in the browser gallery.",
		Need:            "Triangulation parity has distinct topology and labeling risks that should be visually grouped.",
		RecommendedDemo: "Add or expand a triangulation gallery covering triplot, tripcolor, tricontour, tricontourf, and masked meshes.",
	},
	{
		ID:              "mixed-raster-vector-output",
		Topic:           "raster",
		Title:           "Mixed Raster / Vector Output",
		Priority:        DemoBreadthMedium,
		CatalogIDs:      []string{"mixed_raster_vector", "large_scatter", "mixed_collection", "quad_mesh", "gouraud_triangles", "clip_path_batch"},
		TargetFeatures:  []string{"rasterized artist", "vector text preservation", "clip paths", "dense scatter", "quad mesh", "Gouraud triangles"},
		CurrentCoverage: "Mixed raster/vector behavior is covered by backend stress fixtures but not by public examples.",
		Need:            "Backend users need a visual example that explains when dense artists rasterize while labels and axes remain vector.",
		RecommendedDemo: "Add a mixed-output showcase that saves SVG/PDF examples with rasterized dense artists and vector surrounding content.",
	},
}

// DemoBreadthGaps returns the Phase 9A.3 demo-breadth audit rows.
func DemoBreadthGaps() []DemoBreadthGap {
	out := make([]DemoBreadthGap, len(demoBreadthGaps))
	copy(out, demoBreadthGaps)
	for i := range out {
		out[i].CatalogIDs = append([]string(nil), out[i].CatalogIDs...)
		out[i].ShowcaseIDs = append([]string(nil), out[i].ShowcaseIDs...)
		out[i].WebDemoIDs = append([]string(nil), out[i].WebDemoIDs...)
		out[i].TargetFeatures = append([]string(nil), out[i].TargetFeatures...)
	}
	return out
}

// LookupDemoBreadthGap finds a Phase 9A.3 demo-breadth gap by stable ID.
func LookupDemoBreadthGap(id string) (DemoBreadthGap, bool) {
	for _, gap := range DemoBreadthGaps() {
		if gap.ID == id {
			return gap, true
		}
	}
	return DemoBreadthGap{}, false
}
