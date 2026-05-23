package examplecatalog

// ReferenceConsistencyClassification explains why a fixture-only catalog case
// does not currently have its own user-facing showcase.
type ReferenceConsistencyClassification struct {
	CaseID         string
	Classification string
	Rationale      string
}

var publicFixtureOnlyTopics = map[string]bool{
	"color":      true,
	"colorbar":   true,
	"colormap":   true,
	"image":      true,
	"mathtext":   true,
	"mesh":       true,
	"mplot3d":    true,
	"raster":     true,
	"signal":     true,
	"statistics": true,
}

var referenceConsistencyClassifications = []ReferenceConsistencyClassification{
	{
		CaseID:         "large_scatter",
		Classification: "backend-stress",
		Rationale:      "AGG-native path-collection batching stress fixture; public scatter breadth is tracked by the advanced scatter demo gap.",
	},
	{
		CaseID:         "mixed_collection",
		Classification: "backend-stress",
		Rationale:      "Backend-native mixed path collection fixture; not a standalone public plot family.",
	},
	{
		CaseID:         "quad_mesh",
		Classification: "backend-stress",
		Rationale:      "AGG-native quad mesh batching stress fixture; public mesh breadth is tracked by mesh and image demo gaps.",
	},
	{
		CaseID:         "gouraud_triangles",
		Classification: "backend-stress",
		Rationale:      "AGG-native Gouraud triangle batching stress fixture; public shading breadth is tracked by image and mesh demo gaps.",
	},
	{
		CaseID:         "clip_path_batch",
		Classification: "backend-stress",
		Rationale:      "AGG-native clip-path batching stress fixture; public clipping/mixed-output breadth is tracked elsewhere.",
	},
	{
		CaseID:         "pcolormesh_nearest",
		Classification: "mesh-shading-variant",
		Rationale:      "PColorMesh nearest shading is a fixture-only mesh variant covered by the image and triangulation demo-breadth gaps.",
	},
	{
		CaseID:         "pcolormesh_gouraud",
		Classification: "mesh-shading-variant",
		Rationale:      "PColorMesh Gouraud shading is a fixture-only mesh variant covered by the image and triangulation demo-breadth gaps.",
	},
	{
		CaseID:         "spectrum_variants",
		Classification: "signal-demo-gap",
		Rationale:      "Signal helper variants are public API, but a dedicated signal/spectrum showcase has not been added yet.",
	},
	{
		CaseID:         "specialty_depth",
		Classification: "compound-statistics-fixture",
		Rationale:      "Compound statistical and specialty residual fixture; public statistics coverage is represented by stat_variants and specialty_artists.",
	},
}

// ReferenceConsistencyClassifications returns fixture-only cases that have an
// explicit 9A.5 classification instead of a direct demo-breadth row.
func ReferenceConsistencyClassifications() []ReferenceConsistencyClassification {
	out := make([]ReferenceConsistencyClassification, len(referenceConsistencyClassifications))
	copy(out, referenceConsistencyClassifications)
	return out
}

func referenceConsistencyFixtureOnlyClassification(caseID string) string {
	for _, item := range ReferenceConsistencyClassifications() {
		if item.CaseID == caseID {
			return item.Classification
		}
	}
	return ""
}

func isPublicFixtureOnlyTopic(topic string) bool {
	return publicFixtureOnlyTopics[topic]
}
