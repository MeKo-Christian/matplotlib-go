package core

// Shared helpers used by the specialty plot types (eventplot, hexbin, pie,
// violin, table, sankey). Each plot type lives in its own file.

func floatAt(values []float64, i int, fallback float64) float64 {
	if i < len(values) && isFinite(values[i]) {
		return values[i]
	}
	return fallback
}

func specialtyBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
