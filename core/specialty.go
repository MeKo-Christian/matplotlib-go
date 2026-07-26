package core

// Shared helpers used by the specialty plot types (eventplot, hexbin, pie,
// violin, table, sankey). Each plot type lives in its own file.

import "github.com/cwbudde/matplotlib-go/optional"

func floatAt(values []float64, i int, fallback float64) float64 {
	if i < len(values) && isFinite(values[i]) {
		return values[i]
	}
	return fallback
}

func specialtyBool(value optional.Value[bool], fallback bool) bool {
	return value.Or(fallback)
}
