package examplecatalog

// PublicSurfaceRow records one upstream Matplotlib public API or registry item.
type PublicSurfaceRow struct {
	ID     string `json:"id"`
	Module string `json:"module"`
	Kind   string `json:"kind"`
	Name   string `json:"name"`
}

// PublicSurfaceParityStatus records how a public upstream surface maps to this
// Go port.
type PublicSurfaceParityStatus string

const (
	PublicSurfaceDirectEquivalent    PublicSurfaceParityStatus = "direct-equivalent"
	PublicSurfaceIdiomaticEquivalent PublicSurfaceParityStatus = "idiomatic-equivalent"
	PublicSurfacePartial             PublicSurfaceParityStatus = "partial"
	PublicSurfaceNotStarted          PublicSurfaceParityStatus = "not-started"
	PublicSurfaceIntentionalOmission PublicSurfaceParityStatus = "intentional-omission"
)

// PublicSurfaceParity maps a single upstream API or registry row to local
// implementation and coverage context.
type PublicSurfaceParity struct {
	ID                string
	UpstreamID        string
	FeatureCoverageID string
	Status            PublicSurfaceParityStatus
	GoFiles           []string
	CatalogIDs        []string
	ExampleIDs        []string
	Note              string
}

var publicSurfaceParityRows = []PublicSurfaceParity{
	{
		ID:                "artist-class",
		UpstreamID:        "artist.py:class:Artist",
		FeatureCoverageID: "artist",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/artist.go", "core/lifecycle.go", "core/rasterization.go"},
		CatalogIDs:        []string{"basic_line", "patch_showcase", "mixed_raster_vector"},
		Note:              "Go has an Artist interface and lifecycle/rendering metadata, but not Matplotlib's full dynamic property, stale, callback, and setter surface.",
	},
	{
		ID:                "line2d-class",
		UpstreamID:        "lines.py:class:Line2D",
		FeatureCoverageID: "lines",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/line.go", "core/plot.go", "core/scatter.go"},
		CatalogIDs:        []string{"basic_line", "dashes", "joins_caps", "scatter_marker_types"},
		ExampleIDs:        []string{"basic_line", "dashes"},
		Note:              "Line strokes, dashes, joins, and caps exist; integrated Line2D marker and mutable data semantics remain Phase 9C work.",
	},
	{
		ID:                "star-marker",
		UpstreamID:        "markers.py:registry:marker:*",
		FeatureCoverageID: "lines",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"core/marker.go", "core/scatter.go"},
		CatalogIDs:        []string{"scatter_marker_types"},
		Note:              "Built-in marker registry parity includes the star marker and is covered by the marker grid fixture.",
	},
	{
		ID:                "lanczos-interpolation",
		UpstreamID:        "image.py:registry:interpolation:lanczos",
		FeatureCoverageID: "image",
		Status:            PublicSurfaceNotStarted,
		GoFiles:           []string{"core/image.go", "backends/agg/interpolation.go"},
		CatalogIDs:        []string{"image_heatmap", "imshow_bilinear", "imshow_bicubic"},
		Note:              "Nearest, bilinear, bicubic, hamming, and hanning are represented; Lanczos and the remaining interpolation filters are Phase 9C work.",
	},
	{
		ID:                "pyplot-plot",
		UpstreamID:        "pyplot.py:function:plot",
		FeatureCoverageID: "pyplot-state",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"pyplot/pyplot.go", "core/plot.go"},
		CatalogIDs:        []string{"basic_line"},
		ExampleIDs:        []string{"basic_line"},
		Note:              "A stateful Go pyplot package exists for common migration paths, but it intentionally does not mirror the full Python overload surface.",
	},
	{
		ID:                "button-widget",
		UpstreamID:        "widgets.py:class:Button",
		FeatureCoverageID: "widgets-events-animation",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/widgets.go", "canvas/dispatcher.go"},
		Note:              "Static widget artists and event infrastructure exist; interactive widget workflows remain Phase 9C work.",
	},
	{
		ID:                "func-animation",
		UpstreamID:        "animation.py:class:FuncAnimation",
		FeatureCoverageID: "widgets-events-animation",
		Status:            PublicSurfaceNotStarted,
		GoFiles:           []string{"canvas/scheduler.go"},
		Note:              "Canvas scheduling exists, but Matplotlib-style FuncAnimation parity has not been implemented or demoed.",
	},
}

// PublicSurfaceParityRows returns Phase 9B public-surface parity
// classifications.
func PublicSurfaceParityRows() []PublicSurfaceParity {
	out := make([]PublicSurfaceParity, len(publicSurfaceParityRows))
	copy(out, publicSurfaceParityRows)
	return out
}

// LookupPublicSurfaceParityByUpstreamID finds a classification by upstream
// public-surface row ID.
func LookupPublicSurfaceParityByUpstreamID(upstreamID string) (PublicSurfaceParity, bool) {
	for _, row := range PublicSurfaceParityRows() {
		if row.UpstreamID == upstreamID {
			return row, true
		}
	}
	return PublicSurfaceParity{}, false
}
