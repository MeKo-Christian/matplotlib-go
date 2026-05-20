package examplecatalog

// Case describes a plotting example that participates in parity testing.
//
// The catalog is the shared source of truth for the relationship between
// user-facing showcase examples, Go golden fixtures, Matplotlib references,
// and the curated web demo subset.
//
// Layout rules:
//   - Showcase: true        → body lives at examples/<id>/example.go (also
//     the GoPath for source display); test/parity
//     wrapper imports it.
//   - FixtureOnly: true     → body lives at test/parity/<id>/plot.go; no
//     examples/<id>/ entry. These are stress
//     fixtures, not user examples.
//   - neither               → body lives at test/parity/<id>/plot.go; not
//     surfaced in examples/. They participate in
//     parity testing but are not curated showcases.
type Case struct {
	ID          string
	Topic       string
	Title       string
	Description string
	Width       int
	Height      int
	DPI         int
	GoPath      string
	PythonPath  string
	Optional    bool
	FixtureOnly bool
	Showcase    bool
	WebDemoID   string

	// NativeBackend marks fixtures that intentionally exercise backend-native
	// renderer capabilities rather than renderer-neutral fallback paths.
	// NativeCapabilities names the backend capabilities the fixture covers.
	NativeBackend      string
	NativeCapabilities []string
	SVGGoldenFamily    string

	// MinPSNR / MaxMeanAbs / MaxRMSE override the matplotlib reference-compare
	// tolerance defaults for this case. Zero means "use the package default".
	// Only consulted by tests in test/; harmless metadata otherwise.
	MinPSNR    float64
	MaxMeanAbs float64
	MaxRMSE    float64
}

const (
	DefaultWidth  = 640
	DefaultHeight = 360
	DefaultDPI    = 100
)

var cases = []Case{
	{ID: "basic_line", Topic: "lines", Title: "Basic Line", Showcase: true},
	{ID: "joins_caps", Topic: "lines", Title: "Line Joins and Caps"},
	{ID: "dashes", Topic: "lines", Title: "Dash Patterns", Showcase: true},
	{ID: "scatter_basic", Topic: "scatter", Title: "Basic Scatter", Showcase: true},
	{ID: "scatter_marker_types", Topic: "scatter", Title: "Scatter Marker Types"},
	{ID: "scatter_advanced", Topic: "scatter", Title: "Advanced Scatter"},
	{ID: "bar_basic_frame", Topic: "bar", Title: "Bar Frame"},
	{ID: "bar_basic_ticks", Topic: "bar", Title: "Bar Ticks"},
	{ID: "bar_basic_tick_labels", Topic: "bar", Title: "Bar Tick Labels"},
	{ID: "bar_basic_title", Topic: "bar", Title: "Bar Title"},
	{ID: "bar_basic", Topic: "bar", Title: "Basic Bars", Showcase: true, SVGGoldenFamily: "bar"},
	{ID: "bar_horizontal", Topic: "bar", Title: "Horizontal Bars"},
	{ID: "bar_grouped", Topic: "bar", Title: "Grouped Bars"},
	{ID: "fill_basic", Topic: "fill", Title: "Fill to Baseline", Showcase: true, MinPSNR: 45.0, MaxMeanAbs: 6.0},
	{ID: "fill_between", Topic: "fill", Title: "Fill Between Curves"},
	{ID: "fill_stacked", Topic: "fill", Title: "Stacked Fill"},
	{ID: "errorbar_basic", Topic: "errorbar", Title: "Error Bars", Showcase: true, SVGGoldenFamily: "errorbar"},
	{ID: "multi_series_basic", Topic: "multi", Title: "Multiple Series", Showcase: true},
	{ID: "multi_series_color_cycle", Topic: "multi", Title: "Color Cycle"},
	{ID: "hist_basic", Topic: "histogram", Title: "Histogram Counts", Showcase: true, SVGGoldenFamily: "hist"},
	{ID: "hist_density", Topic: "histogram", Title: "Histogram Density"},
	{ID: "hist_strategies", Topic: "histogram", Title: "Histogram Strategies"},
	{ID: "boxplot_basic", Topic: "boxplot", Title: "Box Plot", Optional: true, Showcase: true, MinPSNR: 44.0, MaxMeanAbs: 2.0},
	{ID: "text_labels_strict", Topic: "text", Title: "Strict Text Labels", Optional: true, SVGGoldenFamily: "text_layout"},
	{ID: "title_strict", Topic: "text", Title: "Strict Title"},
	{ID: "mathtext_basic", Topic: "mathtext", Title: "MathText Basic", FixtureOnly: true, SVGGoldenFamily: "mathtext"},
	{ID: "mathtext_fractions", Topic: "mathtext", Title: "MathText Fractions", FixtureOnly: true},
	{ID: "mathtext_integrals", Topic: "mathtext", Title: "MathText Operators", FixtureOnly: true},
	{ID: "mathtext_matrices", Topic: "mathtext", Title: "MathText Matrices", FixtureOnly: true},
	{ID: "mathtext_inline_labels", Topic: "mathtext", Title: "MathText Inline Labels", FixtureOnly: true},
	{ID: "image_heatmap", Topic: "image", Title: "Heatmap Image", Showcase: true, SVGGoldenFamily: "image"},
	{ID: "imshow_clipped", Topic: "image", Title: "Clipped Imshow", FixtureOnly: true, MinPSNR: 30.0, MaxMeanAbs: 10.0},
	{ID: "imshow_transformed", Topic: "image", Title: "Transformed Imshow", FixtureOnly: true, Width: 420, Height: 420, MinPSNR: 24.0, MaxMeanAbs: 18.0, MaxRMSE: 30.0},
	{ID: "imshow_bilinear", Topic: "image", Title: "Bilinear Imshow", FixtureOnly: true, Width: 256, Height: 256, MinPSNR: 30.0, MaxMeanAbs: 16.0},
	{ID: "imshow_bicubic", Topic: "image", Title: "Bicubic Imshow", FixtureOnly: true, Width: 256, Height: 256, MinPSNR: 30.0, MaxMeanAbs: 16.0},
	{ID: "image_alpha", Topic: "image", Title: "Image Alpha", FixtureOnly: true, MinPSNR: 30.0, MaxMeanAbs: 16.0, MaxRMSE: 10.0},
	{ID: "matshow_basic", Topic: "image", Title: "Matshow", FixtureOnly: true, MinPSNR: 30.0, MaxMeanAbs: 10.0, MaxRMSE: 10.0},
	{ID: "spy_marker", Topic: "image", Title: "Spy Marker Mode", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 12.0},
	{ID: "spy_image", Topic: "image", Title: "Spy Image Mode", FixtureOnly: true, MinPSNR: 27.0, MaxMeanAbs: 22.0, MaxRMSE: 30.0},
	{ID: "axes_top_right_inverted", Topic: "axes", Title: "Top/Right Inverted Axes", Optional: true},
	{ID: "axes_control_surface", Topic: "axes", Title: "Axes, Scales, and Twins", Optional: true, WebDemoID: "axes", Description: "Minor ticks, top/right axes, aspect controls, log scale, twin axes, and secondary axes.", Showcase: true, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "transform_coordinates", Topic: "axes", Title: "Transform Coordinates", Optional: true, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "gridspec_composition", Topic: "composition", Title: "Figure Composition", WebDemoID: "composition", Description: "GridSpec spans, figure-level labels, figure legends, anchored text, and colorbars.", Showcase: true, MinPSNR: 35.0, MaxMeanAbs: 8.0},
	{ID: "figure_labels_composition", Topic: "composition", Title: "Figure Labels", Showcase: true, MinPSNR: 32.0, MaxMeanAbs: 9.0},
	{ID: "colorbar_composition", Topic: "colorbar", Title: "Colorbar Composition", Showcase: true, MinPSNR: 32.0, MaxMeanAbs: 16.0},
	{ID: "annotation_composition", Topic: "annotation", Title: "Annotations", Showcase: true, MinPSNR: 35.0, MaxMeanAbs: 7.0},
	{ID: "patch_showcase", Topic: "patches", Title: "Patch Showcase", Optional: true, SVGGoldenFamily: "hatch_bars", MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "mesh_contour_tri", Topic: "mesh", Title: "Meshes and Contours", Optional: true, WebDemoID: "mesh", Description: "PColorMesh, contour/contourf, Hist2D, triplot, tripcolor, and tricontour.", Showcase: true, MinPSNR: 37.5, MaxMeanAbs: 7.5},
	{ID: "plot_variants", Topic: "variants", Title: "Plot Variants", Optional: true, WebDemoID: "variants", Description: "Step, stairs, reference lines, spans, broken bars, and stacked bars.", Showcase: true, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "spectrum_variants", Topic: "signal", Title: "Spectrum Variants", FixtureOnly: true, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "stat_variants", Topic: "statistics", Title: "Statistical Views", Optional: true, WebDemoID: "statistics", Description: "Box plots, violin plots, empirical CDFs, and stack plots.", Showcase: true, MinPSNR: 32.0, MaxMeanAbs: 9.0},
	{ID: "specialty_depth", Topic: "statistics", Title: "Specialty Depth", FixtureOnly: true, MinPSNR: 22.0, MaxMeanAbs: 20.0, MaxRMSE: 35.0},
	{ID: "stem_plot", Topic: "specialty", Title: "Stem Plot", Optional: true},
	{ID: "specialty_artists", Topic: "specialty", Title: "Specialty Artists", Optional: true, WebDemoID: "specialty", Description: "Event plots, hexbin, pie charts, stem plots, tables, and Sankey-style flows.", Showcase: true},
	{ID: "units_overview", Topic: "units", Title: "Dates and Categories", Optional: true, WebDemoID: "units", Description: "Time-aware axes, categorical bars, and horizontal categorical bars.", Showcase: true, MinPSNR: 43.5},
	{ID: "units_dates", Topic: "units", Title: "Date Units", Optional: true, MinPSNR: 45.0, MaxMeanAbs: 1.6},
	{ID: "units_categories", Topic: "units", Title: "Category Units", Optional: true, MinPSNR: 41.0, MaxMeanAbs: 3.2},
	{ID: "units_custom_converter", Topic: "units", Title: "Custom Unit Converter", Optional: true, MinPSNR: 40.0, MaxMeanAbs: 3.5},
	{ID: "vector_fields", Topic: "vectors", Title: "Vector Fields", Optional: true, WebDemoID: "vectors", Description: "Quiver, quiver keys, barbs, streamplots, and grid-based vector input.", Showcase: true, MinPSNR: 41.5, MaxMeanAbs: 3.0},
	{ID: "polar_axes", Topic: "polar", Title: "Polar Wave", WebDemoID: "polar", Description: "A filled polar curve with custom radial and angular grid styling.", Showcase: true, SVGGoldenFamily: "clipped_polar", MinPSNR: 32.0, MaxMeanAbs: 9.0},
	{ID: "geo_mollweide_axes", Topic: "geo", Title: "Projections and Insets", WebDemoID: "projections", Description: "Mollweide geo projection plus a zoomed inset axes.", Showcase: true, MinPSNR: 30.0, MaxMeanAbs: 12.0},
	{ID: "geo_aitoff_axes", Topic: "geo", Title: "Aitoff Projection", Optional: true, Showcase: true, MinPSNR: 30.0, MaxMeanAbs: 12.0},
	{ID: "geo_hammer_axes", Topic: "geo", Title: "Hammer Projection", Optional: true, MinPSNR: 30.0, MaxMeanAbs: 12.0},
	{ID: "geo_lambert_axes", Topic: "geo", Title: "Lambert Projection", Optional: true, MinPSNR: 30.0, MaxMeanAbs: 12.0},
	{ID: "radar_basic", Topic: "radar", Title: "Radar Projection", Optional: true, Showcase: true, MinPSNR: 45.0, MaxMeanAbs: 2.0},
	{ID: "skewt_basic", Topic: "skewt", Title: "Skew-T Projection", Optional: true, Showcase: true, MinPSNR: 24.0, MaxMeanAbs: 18.0},
	{ID: "mplot3d_basic", Topic: "mplot3d", Title: "3D Toolkit Scaffold", Optional: true, MinPSNR: 39.0, MaxMeanAbs: 5.0, MaxRMSE: 18.0},
	{ID: "mplot3d_terrain", Topic: "mplot3d", Title: "3D Terrain", Optional: true, Width: 900, Height: 640, Showcase: true, MinPSNR: 38.0, MaxMeanAbs: 5.0, MaxRMSE: 18.0},
	{ID: "mplot3d_plot3d", Topic: "mplot3d", Title: "3D Plot", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 38.0, MaxMeanAbs: 8.0, MaxRMSE: 18.0},
	{ID: "mplot3d_scatter3d", Topic: "mplot3d", Title: "3D Scatter", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 35.0, MaxMeanAbs: 8.0, MaxRMSE: 18.0},
	{ID: "mplot3d_surface3d", Topic: "mplot3d", Title: "3D Surface", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 35.0, MaxMeanAbs: 10.0, MaxRMSE: 18.0},
	{ID: "mplot3d_wire3d", Topic: "mplot3d", Title: "3D Wireframe", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 30.0, MaxMeanAbs: 10.0, MaxRMSE: 18.0},
	{ID: "mplot3d_trisurf3d", Topic: "mplot3d", Title: "3D Triangulated Surface", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 30.0, MaxMeanAbs: 12.0, MaxRMSE: 18.0},
	{ID: "mplot3d_bar3d", Topic: "mplot3d", Title: "3D Bars", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 30.0, MaxMeanAbs: 8.0, MaxRMSE: 18.0},
	{ID: "mplot3d_voxels", Topic: "mplot3d", Title: "3D Voxels", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 30.0, MaxMeanAbs: 12.0, MaxRMSE: 18.0},
	{ID: "mplot3d_quiver3d", Topic: "mplot3d", Title: "3D Quiver", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 30.0, MaxMeanAbs: 10.0, MaxRMSE: 18.0},
	{ID: "mplot3d_stem3d", Topic: "mplot3d", Title: "3D Stem", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 30.0, MaxMeanAbs: 10.0, MaxRMSE: 18.0},
	{ID: "mplot3d_fill_between3d", Topic: "mplot3d", Title: "3D Fill Between", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 35.0, MaxMeanAbs: 10.0, MaxRMSE: 18.0},
	{ID: "unstructured_showcase", Topic: "unstructured", Title: "Unstructured Showcase", Optional: true, Showcase: true, MinPSNR: 30.0, MaxMeanAbs: 10.0},
	{ID: "arrays_showcase", Topic: "arrays", Title: "Matrix Helpers", Optional: true, WebDemoID: "matrix", Description: "MatShow, sparsity spy plots, annotated heatmaps, and colorbars.", Width: 1240, Height: 620, Showcase: true, MinPSNR: 30.0, MaxMeanAbs: 10.0},
	{ID: "axisartist_showcase", Topic: "axisartist", Title: "AxisArtist Showcase", Optional: true, Showcase: true, MinPSNR: 28.0, MaxMeanAbs: 12.0},
	{ID: "axes_grid1_showcase", Topic: "axes_grid1", Title: "Axes Grid1 Showcase", Optional: true, Showcase: true, MinPSNR: 28.0, MaxMeanAbs: 12.0},
	{ID: "pcolor_flat", Topic: "mesh", Title: "PColor Flat", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 15.0},
	{ID: "pcolormesh_nearest", Topic: "mesh", Title: "PColorMesh Nearest", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 15.0},
	{ID: "pcolormesh_gouraud", Topic: "mesh", Title: "PColorMesh Gouraud", FixtureOnly: true, MinPSNR: 20.0, MaxMeanAbs: 22.0, MaxRMSE: 30.0},
	{ID: "pcolormesh_masked", Topic: "mesh", Title: "PColorMesh Masked", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 15.0},
	{ID: "hist2d_weighted_density", Topic: "mesh", Title: "Hist2D Weighted Density", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0, MaxRMSE: 30.0},
	{ID: "boundarynorm_pcolormesh", Topic: "colorbar", Title: "BoundaryNorm PColorMesh", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0},
	{ID: "lognorm_imshow", Topic: "colorbar", Title: "LogNorm Imshow", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0},
	{ID: "twoslope_norm_image", Topic: "colorbar", Title: "TwoSlopeNorm Image", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0},
	{ID: "colorbar_extensions", Topic: "colorbar", Title: "Colorbar Extensions", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0},
	{ID: "large_scatter", Topic: "raster", Title: "Large Scatter Batch", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"pathcollectionbatch"}, MinPSNR: 55.0, MaxMeanAbs: 0.5, MaxRMSE: 4.0},
	{ID: "mixed_collection", Topic: "raster", Title: "Mixed Path Collection", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"pathcollectionbatch"}, SVGGoldenFamily: "collection", MinPSNR: 60.0, MaxMeanAbs: 0.5, MaxRMSE: 2.0},
	{ID: "quad_mesh", Topic: "raster", Title: "Quad Mesh Batch", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"quadmeshbatch"}, MinPSNR: 48.0, MaxMeanAbs: 1.0, MaxRMSE: 4.0},
	{ID: "gouraud_triangles", Topic: "raster", Title: "Gouraud Triangles", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"gouraudtrianglebatch"}, MinPSNR: 25.0, MaxMeanAbs: 18.0},
	{ID: "clip_path_batch", Topic: "raster", Title: "Clip Path Batch", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"pathclip", "quadmeshbatch"}, MinPSNR: 45.0, MaxMeanAbs: 1.0, MaxRMSE: 6.0},
}

// Cases returns every cataloged parity example/fixture.
func Cases() []Case {
	out := make([]Case, len(cases))
	copy(out, cases)
	for i := range out {
		out[i].NativeCapabilities = append([]string(nil), out[i].NativeCapabilities...)
		applyDefaults(&out[i])
	}
	return out
}

// Lookup finds a parity case by canonical case ID.
func Lookup(id string) (Case, bool) {
	all := Cases()
	for i := range all {
		if all[i].ID == id {
			return all[i], true
		}
	}
	return Case{}, false
}

// WebDemos returns the curated web-demo subset in display order.
func WebDemos() []Case {
	all := Cases()
	var out []Case
	for i := range all {
		if all[i].WebDemoID != "" {
			out = append(out, all[i])
		}
	}
	return out
}

// LookupWebDemo finds a catalog case by browser demo ID.
func LookupWebDemo(id string) (Case, bool) {
	demos := WebDemos()
	for i := range demos {
		if demos[i].WebDemoID == id {
			return demos[i], true
		}
	}
	return Case{}, false
}

func applyDefaults(c *Case) {
	if c.ID != "" {
		if c.GoPath == "" {
			if c.Showcase {
				c.GoPath = "examples/" + c.ID + "/example.go"
			} else {
				// Non-showcase cases (parity-only and fixture-only alike) live
				// exclusively under test/parity/ and are not surfaced in examples/.
				c.GoPath = "test/parity/" + c.ID + "/plot.go"
			}
		}
		c.PythonPath = "test/parity/" + c.ID + "/plot.py"
	}
	if c.Width == 0 {
		c.Width = DefaultWidth
	}
	if c.Height == 0 {
		c.Height = DefaultHeight
	}
	if c.DPI == 0 {
		c.DPI = DefaultDPI
	}
	if c.Description == "" {
		c.Description = c.Title
	}
}
