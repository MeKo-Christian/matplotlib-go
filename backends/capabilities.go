package backends

import "fmt"

// CapabilityMatrix returns a formatted table of backend capabilities.
func CapabilityMatrix() string {
	available := Available()
	if len(available) == 0 {
		return "No backends available"
	}

	capabilities := []Capability{
		AntiAliasing, SubPixel, GradientFill, PathClip,
		GPUAccel, Threading, VectorOutput, TextShaping, FontHinting,
	}

	result := fmt.Sprintf("%-12s", "Backend")
	for _, cap := range capabilities {
		result += fmt.Sprintf("%-12s", string(cap))
	}
	result += "\n"

	// Add separator line
	result += fmt.Sprintf("%-12s", "--------")
	for range capabilities {
		result += fmt.Sprintf("%-12s", "--------")
	}
	result += "\n"

	// Add backend rows
	for _, backend := range available {
		result += fmt.Sprintf("%-12s", string(backend))
		for _, cap := range capabilities {
			if HasCapability(backend, cap) {
				result += fmt.Sprintf("%-12s", "✓")
			} else {
				result += fmt.Sprintf("%-12s", "✗")
			}
		}
		result += "\n"
	}

	return result
}

// RequiredCapabilities defines capability sets for common use cases.
var RequiredCapabilities = map[string][]Capability{
	"basic": {
		AntiAliasing,
	},
	"publication": {
		AntiAliasing,
		VectorOutput,
		TextShaping,
	},
	"interactive": {
		AntiAliasing,
		GPUAccel,
		Threading,
	},
	"scientific": {
		AntiAliasing,
		SubPixel,
		VectorOutput,
		TextShaping,
		FontHinting,
	},
}

// GetRecommendedBackend returns the best backend for a specific use case.
func GetRecommendedBackend(useCase string) (Backend, error) {
	required, ok := RequiredCapabilities[useCase]
	if !ok {
		return "", fmt.Errorf("unknown use case: %s", useCase)
	}

	return GetBestBackend(required)
}