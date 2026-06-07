package examplecatalog

// CatalogAuditDecision records whether an enumerable Matplotlib catalog was
// implemented directly/idiomatically or deliberately omitted for this Go API.
type CatalogAuditDecision string

const (
	CatalogDecisionImplemented CatalogAuditDecision = "implemented"
	CatalogDecisionOmitted     CatalogAuditDecision = "intentional-omission"
)

// CatalogAudit records the evidence used to close one Phase 9 enumerable
// catalog or miscellaneous coverage-audit decision.
type CatalogAudit struct {
	ID              string
	Title           string
	Decision        CatalogAuditDecision
	UpstreamSources []string
	GoSources       []string
	GuardTests      []string
	CatalogIDs      []string
	Notes           string
}

// EnumerableCatalogAudits returns the machine-readable Phase 9 closure
// ledger. The guard tests named here are the CI checks that keep each catalog
// from silently drifting away from the local Matplotlib 3.10.9 snapshot.
func EnumerableCatalogAudits() []CatalogAudit {
	rows := []CatalogAudit{
		{
			ID:       "colormap-registry",
			Title:    "Matplotlib colormap registry",
			Decision: CatalogDecisionImplemented,
			UpstreamSources: []string{
				"third_party/matplotlib/lib/matplotlib/_cm.py",
				"third_party/matplotlib/lib/matplotlib/_cm_listed.py",
				"third_party/matplotlib/lib/matplotlib/colors.py",
			},
			GoSources: []string{
				"color/colormap.go",
				"color/listed_colormaps.go",
			},
			GuardTests: []string{
				"color.TestMatplotlibPublicColormapCatalogRegistered",
				"color.TestGetColormap_ReversedVariantGeneratedFromBase",
				"color.TestColormapResampledCreatesListedLookup",
				"color.TestListedAndLinearSegmentedConstructors",
			},
			CatalogIDs: []string{"colormap_diverging", "colormap_qualitative", "colormap_cyclic"},
			Notes:      "Base, reversed, listed, linear-segmented, resampled, and representative visual cases are catalog-visible.",
		},
		{
			ID:       "marker-registry",
			Title:    "Matplotlib marker and fillstyle registry",
			Decision: CatalogDecisionImplemented,
			UpstreamSources: []string{
				"third_party/matplotlib/lib/matplotlib/markers.py",
				"third_party/matplotlib/lib/matplotlib/lines.py",
			},
			GoSources: []string{"core/scatter.go", "core/line.go"},
			GuardTests: []string{
				"core.TestMarkerTypeFromStringCoversMatplotlibAliases",
				"core.TestAllBuiltInMarkersHaveNonEmptyPaths",
				"core.TestScatterHalfFilledMarkerDrawsSplitFillAndWholeEdge",
				"core.TestLine2DHalfFilledMarkerDrawsSplitFillAndWholeEdge",
			},
			CatalogIDs: []string{"scatter_marker_types", "line2d_markers"},
			Notes:      "Built-ins, fillstyle, tuple markers, custom paths, mathtext markers, and Line2D/scatter routing are covered.",
		},
		{
			ID:       "named-colors",
			Title:    "Named color tables and to_rgba-equivalent parsing",
			Decision: CatalogDecisionImplemented,
			UpstreamSources: []string{
				"third_party/matplotlib/lib/matplotlib/_color_data.py",
				"third_party/matplotlib/lib/matplotlib/colors.py",
			},
			GoSources: []string{"color/named_colors.go", "color/named_colors_data.go"},
			GuardTests: []string{
				"color.TestNamedColorCatalogSizesMatchMatplotlib",
				"color.TestNamedColorInventoryMatchesMatplotlibTables",
				"color.TestToRGBAResolvesMatplotlibNamedColors",
			},
			CatalogIDs: []string{"named_colors"},
			Notes:      "CSS4/X11, Tableau, xkcd, single-letter, hex, grayscale, tuple, and C-cycle parsing are covered.",
		},
		{
			ID:              "image-interpolation-registry",
			Title:           "imshow interpolation names and adaptive policy",
			Decision:        CatalogDecisionImplemented,
			UpstreamSources: []string{"third_party/matplotlib/lib/matplotlib/image.py"},
			GoSources:       []string{"backends/agg/interpolation.go", "core/image.go", "core/image_api.go"},
			GuardTests: []string{
				"agg.TestParseInterpolationName",
				"agg.TestResolveInterpolationName",
				"agg.TestAggImage_AllMatplotlibInterpolationNamesRender",
				"core.TestTransformedImageInterpolationKernelAlignmentIsDocumented",
			},
			CatalogIDs: []string{"imshow_interpolation_matrix"},
			Notes:      "AGG consumes the full registry; GoBasic/vector fallbacks are documented.",
		},
		{
			ID:       "patch-style-registries",
			Title:    "Patch, BoxStyle, ArrowStyle, and ConnectionStyle registries",
			Decision: CatalogDecisionImplemented,
			UpstreamSources: []string{
				"third_party/matplotlib/lib/matplotlib/patches.py",
			},
			GoSources: []string{"core/patch.go", "core/patch_extra.go", "core/arrow_patch.go"},
			GuardTests: []string{
				"core.TestFancyBboxPatchAdditionalBoxStyles",
				"core.TestFancyBboxPatchToothStyles",
				"core.TestArrowAndConnectionStyleRegistriesParseMatplotlibNames",
				"internal/examplecatalog.TestPatchDebugHelperRowsAreExplicitlyOmitted",
			},
			CatalogIDs: []string{"patch_showcase", "patch_style_matrix"},
			Notes:      "Static registry entries are implemented; Python debug helpers and broad mutable property grammar are explicit divergences.",
		},
		{
			ID:              "hatch-registry",
			Title:           "Hatch character set and repeat density",
			Decision:        CatalogDecisionImplemented,
			UpstreamSources: []string{"third_party/matplotlib/lib/matplotlib/hatch.py"},
			GoSources:       []string{"render/hatch.go", "backends/agg/agg_paths.go", "backends/internal/vectorhatch/vectorhatch.go"},
			GuardTests: []string{
				"render.TestDrawHatchFallbackRepeatedPatternTightensSpacing",
				"render.TestDrawHatchFallbackSupportsShapePatterns",
				"render.TestDrawHatchFallbackShapeSizesFollowMatplotlibRatios",
				"agg.TestNativeDiagonalHatchDensityMatchesMatplotlibReference",
				"internal/examplecatalog.TestHatchImplementationRowsAreSplitFromRendererHatchSurface",
			},
			CatalogIDs: []string{"patch_showcase", "patch_style_matrix"},
			Notes:      "Renderer-neutral and native/vector hatches cover the upstream character set; Python implementation classes are omitted.",
		},
		{
			ID:       "sketch-xkcd",
			Title:    "Sketch-style / pyplot.xkcd mode",
			Decision: CatalogDecisionOmitted,
			UpstreamSources: []string{
				"third_party/matplotlib/lib/matplotlib/pyplot.py",
				"third_party/matplotlib/lib/matplotlib/artist.py",
			},
			GoSources: []string{"render/render.go", "render/graphics_context.go", "internal/examplecatalog/public_surface_parity.go"},
			GuardTests: []string{
				"render.TestGraphicsContextToPaintCarriesOptionalState",
				"internal/examplecatalog.TestSketchAndFigureImageRowsHaveExplicitDecisions",
			},
			Notes: "Global xkcd rcParams mutation and Artist.set_sketch_params are intentionally omitted until a fixture needs sketch jitter parity.",
		},
		{
			ID:       "figimage",
			Title:    "FigureImage / pyplot.figimage",
			Decision: CatalogDecisionOmitted,
			UpstreamSources: []string{
				"third_party/matplotlib/lib/matplotlib/image.py",
				"third_party/matplotlib/lib/matplotlib/pyplot.py",
			},
			GoSources: []string{"core/image_api.go", "internal/examplecatalog/public_surface_parity.go"},
			GuardTests: []string{
				"internal/examplecatalog.TestSketchAndFigureImageRowsHaveExplicitDecisions",
			},
			Notes: "Pixel-anchored figure overlays are intentionally omitted; use a full-figure frameless Axes with Image2D.",
		},
		{
			ID:       "rcparams",
			Title:    "rcParams key universe",
			Decision: CatalogDecisionOmitted,
			UpstreamSources: []string{
				"third_party/matplotlib/lib/matplotlib/rcsetup.py",
				"third_party/matplotlib/lib/matplotlib/mpl-data/matplotlibrc",
			},
			GoSources: []string{"style/mplstyle.go", "style/runtime.go", "docs/matplotlib-migration-notes.md"},
			GuardTests: []string{
				"style.TestSupportedMPLStyleKeysAreAuditedAgainstUpstream",
				"style.TestMPLStyleParamsApplySupportedKeysAndReportUnsupported",
			},
			Notes: "Go supports a documented rcParams subset; unsupported keys are reported and ignored as typed-API divergences.",
		},
	}
	out := make([]CatalogAudit, len(rows))
	copy(out, rows)
	return out
}

// CatalogExitStatus records whether a Phase 9 exit criterion is satisfied.
type CatalogExitStatus string

const (
	CatalogExitSatisfied CatalogExitStatus = "satisfied"
)

// CatalogExitCriterion records evidence for one Phase 9 exit criterion.
type CatalogExitCriterion struct {
	ID        string
	Criterion string
	Status    CatalogExitStatus
	Evidence  string
}

// CatalogExitCriteria returns the machine-checked closure ledger for PLAN.md
// Phase 9's exit criteria.
func CatalogExitCriteria() []CatalogExitCriterion {
	rows := []CatalogExitCriterion{
		{
			ID:        "catalog-decisions",
			Criterion: "Every enumerable catalog is implemented or explicitly justified as an intentional divergence.",
			Status:    CatalogExitSatisfied,
			Evidence:  "EnumerableCatalogAudits plus PublicSurfaceParityRowsForSurface classify each tracked catalog and omission.",
		},
		{
			ID:        "catalog-visibility",
			Criterion: "Newly implemented catalog features have at least one TestGolden/TestReferenceCompare-visible catalog case.",
			Status:    CatalogExitSatisfied,
			Evidence:  "Implemented EnumerableCatalogAudits rows require CatalogIDs that resolve through examplecatalog.Lookup.",
		},
		{
			ID:        "sentinel-user-calls",
			Criterion: "Representative user calls for colormap, marker, named color, and interpolation resolve correctly or have documented errors.",
			Status:    CatalogExitSatisfied,
			Evidence:  "Colormap, marker, named-color, and AGG interpolation guard tests cover coolwarm, star markers, xkcd colors, and lanczos-style interpolation routing.",
		},
	}
	out := make([]CatalogExitCriterion, len(rows))
	copy(out, rows)
	return out
}
