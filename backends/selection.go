package backends

import (
	"fmt"
	"os"
	"strings"

	"github.com/cwbudde/matplotlib-go/render"
)

const AutoBackend Backend = "auto"

// TextCapabilities are the capabilities required for Matplotlib-style text
// rendering: TrueType glyph shaping plus font hinting support.
var TextCapabilities = []Capability{TextShaping, FontHinting}

func hasAllCapabilities(backend Backend, required []Capability) bool {
	for _, capability := range required {
		if !HasCapability(backend, capability) {
			return false
		}
	}
	return true
}

// ResolveBackend selects a backend from a user-supplied choice.
// Empty, "auto", and "default" all fall back to the best available backend
// for the requested capabilities.
func ResolveBackend(choice string, required []Capability) (Backend, error) {
	normalized := strings.ToLower(strings.TrimSpace(choice))
	switch normalized {
	case "", string(AutoBackend), "default":
		return BestBackend(required)
	}

	backend := Backend(normalized)
	info, ok := DefaultRegistry.Get(backend)
	if !ok {
		return "", fmt.Errorf("unknown backend: %s", choice)
	}
	if !info.Available {
		return "", fmt.Errorf("backend %s is not available (missing dependencies?)", backend)
	}
	if len(required) > 0 && !hasAllCapabilities(backend, required) {
		return "", fmt.Errorf("backend %s does not satisfy required capabilities", backend)
	}
	return backend, nil
}

// NewRenderer constructs a renderer for the chosen backend.
func NewRenderer(choice string, config Config, required []Capability) (render.Renderer, Backend, error) {
	backend, err := ResolveBackend(choice, required)
	if err != nil {
		return nil, "", err
	}

	renderer, err := Create(backend, config)
	if err != nil {
		return nil, "", err
	}

	return renderer, backend, nil
}

// NewRendererFromEnv constructs a renderer using MATPLOTLIB_BACKEND.
func NewRendererFromEnv(config Config, required []Capability) (render.Renderer, Backend, error) {
	return NewRenderer(os.Getenv("MATPLOTLIB_BACKEND"), config, required)
}
