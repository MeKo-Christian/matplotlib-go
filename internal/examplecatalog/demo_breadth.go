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
		CatalogIDs: []string{"basic_line", "dashes", "joins_caps", "scatter_marker_types", "lines_markers_gallery"},
		ShowcaseIDs: []string{
			"basic_line",
			"dashes",
			"lines_markers_gallery",
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
		CatalogIDs: []string{"scatter_basic", "scatter_marker_types", "scatter_advanced", "large_scatter", "scatter_gallery"},
		ShowcaseIDs: []string{
			"scatter_basic",
			"scatter_gallery",
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
		CatalogIDs: []string{"bar_basic", "bar_horizontal", "bar_grouped", "bar_basic_tick_labels", "plot_variants", "bar_variants"},
		ShowcaseIDs: []string{
			"bar_basic",
			"plot_variants",
			"bar_variants",
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
		CatalogIDs: []string{"fill_basic", "fill_between", "fill_stacked", "plot_variants", "fill_variants"},
		ShowcaseIDs: []string{
			"fill_basic",
			"plot_variants",
			"fill_variants",
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
		CatalogIDs: []string{"hist_basic", "hist_density", "hist_strategies", "hist2d_weighted_density", "histogram_variants"},
		ShowcaseIDs: []string{
			"hist_basic",
			"histogram_variants",
		},
		TargetFeatures:  []string{"density", "cumulative", "bin strategies", "weighted hist2d", "multiple histograms"},
		CurrentCoverage: "Histogram counts are showcased; density, strategies, and weighted 2D behavior are fixture-only.",
		Need:            "Histogram examples should show the statistical options users commonly migrate from Matplotlib.",
		RecommendedDemo: "Add a histogram variants showcase covering density, cumulative, multiple datasets, bin strategies, and weighted hist2d.",
	},
	{
		ID:         "named-colors",
		Topic:      "color",
		Title:      "Named Colors",
		Priority:   DemoBreadthMedium,
		CatalogIDs: []string{"named_colors", "named_colors_gallery"},
		ShowcaseIDs: []string{
			"named_colors_gallery",
		},
		TargetFeatures:  []string{"CSS4 names", "Tableau tab names", "xkcd names", "single-letter colors"},
		CurrentCoverage: "Named colors now have a swatch showcase for common name families; the compact parity fixture remains for resolution checks.",
		Need:            "Color-name compatibility is hard to inspect without a swatch-style example.",
		RecommendedDemo: "Add a named-color swatch showcase that displays CSS4, Tableau, xkcd, grayscale, and shorthand color forms.",
	},
	{
		ID:         "colormap-families",
		Topic:      "colormap",
		Title:      "Colormap Families",
		Priority:   DemoBreadthHigh,
		CatalogIDs: []string{"colormap_diverging", "colormap_qualitative", "colormap_cyclic", "image_heatmap", "colormap_families_gallery"},
		ShowcaseIDs: []string{
			"image_heatmap",
			"colormap_families_gallery",
		},
		TargetFeatures:  []string{"sequential", "diverging", "qualitative", "cyclic", "reversed maps", "resampled maps"},
		CurrentCoverage: "Colormap families now have a strip-gallery showcase; resampled-map behavior remains fixture/API-level follow-up.",
		Need:            "Colormap breadth is an enumerable public surface and should be visually browsable.",
		RecommendedDemo: "Add a colormap family showcase with representative sequential, diverging, qualitative, cyclic, reversed, and resampled maps.",
	},
	{
		ID:          "image-variants",
		Topic:       "image",
		Title:       "Image Interpolation, Alpha, Matshow, and Spy",
		Priority:    DemoBreadthHigh,
		CatalogIDs:  []string{"image_heatmap", "imshow_clipped", "imshow_transformed", "imshow_bilinear", "imshow_bicubic", "imshow_interpolation_matrix", "image_alpha", "matshow_basic", "spy_marker", "spy_image", "arrays_showcase", "image_variants_gallery"},
		ShowcaseIDs: []string{"image_heatmap", "arrays_showcase", "image_variants_gallery"},
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
		CurrentCoverage: "Heatmap, matrix helpers, interpolation modes, alpha overlays, MatShow, and spy modes are showcased; transformed-image edge cases remain fixture-only.",
		Need:            "Image behavior is visually sensitive, so users need demos that expose sampling, alpha, and matrix helper choices.",
		RecommendedDemo: "Add an image variants gallery with side-by-side interpolation modes, alpha, transformed/clipped images, matshow, and spy modes.",
	},
	{
		ID:          "colorbar-norms-extensions",
		Topic:       "colorbar",
		Title:       "Colorbar Norms and Extensions",
		Priority:    DemoBreadthHigh,
		CatalogIDs:  []string{"colorbar_composition", "asinh_norm_image", "boundarynorm_pcolormesh", "collection_mutable_scalarmap", "colorbar_boundary_values", "colorbar_horizontal_ticks", "lognorm_imshow", "twoslope_norm_image", "colorbar_extensions", "colorbar_variants_gallery"},
		ShowcaseIDs: []string{"colorbar_composition", "colorbar_variants_gallery"},
		WebDemoIDs:  []string{"composition", "matrix"},
		TargetFeatures: []string{
			"BoundaryNorm",
			"AsinhNorm",
			"LogNorm",
			"TwoSlopeNorm",
			"mutable scalar-mapped collections",
			"explicit boundary values and drawedges",
			"horizontal colorbars with explicit ticks",
			"extended colorbars",
			"shared colorbars",
		},
		CurrentCoverage: "Composition plus LogNorm, TwoSlopeNorm, BoundaryNorm, drawedges, and extension colorbars are showcased; AsinhNorm and mutable scalar maps remain fixture-heavy.",
		Need:            "Norm and colorbar layout behavior are common parity pain points and should be inspectable in one gallery.",
		RecommendedDemo: "Add a colorbar variants showcase for BoundaryNorm, LogNorm, TwoSlopeNorm, extensions, shared axes, and labels.",
	},
	{
		ID:              "mathtext-gallery",
		Topic:           "mathtext",
		Title:           "MathText Gallery",
		Priority:        DemoBreadthHigh,
		CatalogIDs:      []string{"mathtext_basic", "mathtext_fractions", "mathtext_integrals", "mathtext_matrices", "mathtext_inline_labels", "mathtext_gallery"},
		ShowcaseIDs:     []string{"mathtext_gallery"},
		TargetFeatures:  []string{"fractions", "roots", "integrals", "large operators", "matrices", "inline labels"},
		CurrentCoverage: "MathText now has a user-facing gallery for fractions, roots, operators, matrices, and inline labels; the focused parity fixtures remain for residual checks.",
		Need:            "MathText is important enough to need a gallery that users can visually inspect without running fixture tests.",
		RecommendedDemo: "Promote a MathText showcase with inline labels plus panels for fractions, operators, roots, fences, and matrices.",
	},
	{
		ID:          "ticks-scales-formatters",
		Topic:       "axes",
		Title:       "Ticks, Scales, and Formatters",
		Priority:    DemoBreadthHigh,
		CatalogIDs:  []string{"axes_control_surface", "ticks_scales_formatters_gallery", "units_dates", "units_categories", "units_custom_converter", "lognorm_imshow", "skewt_basic"},
		ShowcaseIDs: []string{"axes_control_surface", "ticks_scales_formatters_gallery", "units_overview", "skewt_basic"},
		WebDemoIDs:  []string{"axes", "units"},
		TargetFeatures: []string{
			"major locators",
			"minor locators",
			"date formatters",
			"category formatters",
			"log scale",
			"custom converters",
		},
		CurrentCoverage: "Ticks, scales, formatters, date labels, category labels, and custom units now have a dedicated gallery; focused fixtures remain for locator/formatter edge cases.",
		Need:            "Tick and scale parity should be visible without reading broader axes examples.",
		RecommendedDemo: "Keep ticks_scales_formatters_gallery as the user-facing breadth example and split residual locator/formatter API edge cases into Phase 9C.",
	},
	{
		ID:              "text-layout-gallery",
		Topic:           "text",
		Title:           "Text Layout, Rotation, and Wrapping",
		Priority:        DemoBreadthHigh,
		CatalogIDs:      []string{"text_labels_strict", "title_strict", "text_layout_gallery", "annotation_composition", "figure_labels_composition"},
		ShowcaseIDs:     []string{"text_layout_gallery", "annotation_composition", "figure_labels_composition"},
		TargetFeatures:  []string{"alignment", "rotation", "multiline text", "bbox text", "figure labels", "tick label layout"},
		CurrentCoverage: "Text layout now has a dedicated alignment, rotation, multiline, wrapping, and bbox gallery alongside figure-label and annotation showcases.",
		Need:            "Text metrics are a major parity driver and need a dedicated visual gallery.",
		RecommendedDemo: "Add a text layout showcase with alignment, rotation, multiline text, bbox padding, figure labels, and tick-label variants.",
	},
	{
		ID:          "annotation-legend-offsetbox",
		Topic:       "annotation",
		Title:       "Annotations, Legends, and Offset Boxes",
		Priority:    DemoBreadthHigh,
		CatalogIDs:  []string{"annotation_composition", "annotation_legend_offsetbox_gallery", "figure_labels_composition", "multi_series_basic", "multi_series_color_cycle"},
		ShowcaseIDs: []string{"annotation_composition", "annotation_legend_offsetbox_gallery", "figure_labels_composition", "multi_series_basic"},
		WebDemoIDs:  []string{"composition"},
		TargetFeatures: []string{
			"annotation arrows",
			"offset coordinates",
			"legend placement",
			"figure legends",
			"anchored boxes",
			"proxy-like legend entries",
		},
		CurrentCoverage: "Annotations, bbox annotations, legends, proxy handles, and anchored offset boxes now have a combined user-facing gallery.",
		Need:            "Advanced annotation and legend behavior should have a single gallery for parity inspection.",
		RecommendedDemo: "Add an annotation/legend showcase for arrow styles, coordinate modes, anchored boxes, figure legends, and legend handles.",
	},
	{
		ID:          "mplot3d-gallery",
		Topic:       "mplot3d",
		Title:       "mplot3d Gallery",
		Priority:    DemoBreadthHigh,
		CatalogIDs:  []string{"mplot3d_basic", "mplot3d_terrain", "mplot3d_gallery", "mplot3d_plot3d", "mplot3d_scatter3d", "mplot3d_surface3d", "mplot3d_wire3d", "mplot3d_trisurf3d", "mplot3d_bar3d", "mplot3d_voxels", "mplot3d_quiver3d", "mplot3d_errorbar3d", "mplot3d_stem3d", "mplot3d_fill_between3d", "mplot3d_contour3d", "mplot3d_contourf3d", "mplot3d_tricontour3d", "mplot3d_tricontourf3d", "mplot3d_bar2d_zdir", "mplot3d_text3d"},
		ShowcaseIDs: []string{"mplot3d_terrain", "mplot3d_gallery"},
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
		CurrentCoverage: "mplot3d now has a broad multi-panel gallery plus the focused terrain showcase; specialized contour, text, and errorbar cases remain as parity fixtures.",
		Need:            "3D parity is broad enough that users need a visual gallery in addition to focused fixture names.",
		RecommendedDemo: "Add a 3D gallery with line, scatter, surface, wireframe, trisurf, bar, voxels, quiver, stem, and fill-between panels.",
	},
	{
		ID:          "projection-toolkit-gallery",
		Topic:       "projections",
		Title:       "Projection and Toolkit Gallery",
		Priority:    DemoBreadthMedium,
		CatalogIDs:  []string{"polar_axes", "geo_mollweide_axes", "geo_aitoff_axes", "geo_hammer_axes", "geo_lambert_axes", "radar_basic", "skewt_basic", "projection_toolkit_gallery", "axisartist_showcase", "axes_grid1_showcase"},
		ShowcaseIDs: []string{"polar_axes", "geo_mollweide_axes", "geo_aitoff_axes", "radar_basic", "skewt_basic", "projection_toolkit_gallery", "axisartist_showcase", "axes_grid1_showcase"},
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
		CurrentCoverage: "Projection/toolkit cases now have a grouped CLI showcase spanning polar, geo, radar, skew-T, axisartist-style, and axes_grid1-style panels; browser promotion remains uneven.",
		Need:            "Advanced projection and toolkit behavior should be discoverable as a grouped gallery.",
		RecommendedDemo: "Keep the projection toolkit gallery as the grouped CLI example and promote it into the browser projections family.",
	},
	{
		ID:          "unstructured-triangulation",
		Topic:       "unstructured",
		Title:       "Unstructured Triangulation and Contours",
		Priority:    DemoBreadthMedium,
		CatalogIDs:  []string{"unstructured_showcase", "triangulation_gallery", "mesh_contour_tri", "pcolor_flat", "pcolormesh_masked"},
		ShowcaseIDs: []string{"unstructured_showcase", "triangulation_gallery", "mesh_contour_tri"},
		WebDemoIDs:  []string{"mesh"},
		TargetFeatures: []string{
			"triplot",
			"tripcolor",
			"tricontour",
			"tricontourf",
			"masked meshes",
		},
		CurrentCoverage: "Unstructured triangulation now has a focused gallery covering triplot, tripcolor, tricontour, tricontourf, and masked pcolormesh; browser promotion remains planned.",
		Need:            "Triangulation parity has distinct topology and labeling risks that should be visually grouped.",
		RecommendedDemo: "Promote triangulation_gallery into the mesh browser demo or a dedicated triangulation browser group.",
	},
	{
		ID:          "widgets-animation-gallery",
		Topic:       "widgets",
		Title:       "Widgets and Animation Gallery",
		Priority:    DemoBreadthMedium,
		CatalogIDs:  []string{"widgets_gallery", "animation_gallery"},
		ShowcaseIDs: []string{"widgets_gallery", "animation_gallery"},
		TargetFeatures: []string{
			"buttons",
			"sliders",
			"selectors",
			"cursor",
			"multi-cursor",
			"FuncAnimation",
			"ArtistAnimation",
			"blit playback",
		},
		CurrentCoverage: "widgets_gallery covers static widgets, selectors, cursor, and multi-cursor in the catalog; animation_gallery covers FuncAnimation and ArtistAnimation setup with deterministic stepping.",
		Need:            "Browser-facing coverage still needs promotion once the demo host can exercise timers deterministically.",
		RecommendedDemo: "Promote the widgets and animation galleries into an interactive browser demo with timer-driven line, artist-list, and blit-capable examples.",
	},
	{
		ID:              "mixed-raster-vector-output",
		Topic:           "raster",
		Title:           "Mixed Raster / Vector Output",
		Priority:        DemoBreadthMedium,
		CatalogIDs:      []string{"mixed_raster_vector", "large_scatter", "mixed_collection", "quad_mesh", "gouraud_triangles", "clip_path_batch"},
		ShowcaseIDs:     []string{"mixed_raster_vector"},
		TargetFeatures:  []string{"rasterized artist", "vector text preservation", "clip paths", "dense scatter", "quad mesh", "Gouraud triangles"},
		CurrentCoverage: "Mixed raster/vector behavior now has a public SVG/PDF-oriented showcase for rasterized dense scatter with vector text, axes, line, and legend; lower-level batch capabilities remain stress fixtures.",
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
