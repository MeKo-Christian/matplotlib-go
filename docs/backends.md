# Backend System

Matplotlib-Go uses a pluggable backend architecture that allows different rendering engines to be used interchangeably.

## Available Backends

### AGG (Advanced optional backend)

- **Type**: AGG-backed raster renderer
- **Status**: ✅ Implemented
- **Capabilities**: Anti-aliasing, sub-pixel positioning, hinted text, transformed images
- **Dependencies**: Optional `agg_go`
- **Use cases**: Publication-quality output and advanced text fidelity

### GoBasic (Default)

- **Type**: Pure Go renderer using `golang.org/x/image/vector`
- **Status**: ✅ Implemented
- **Capabilities**: Anti-aliasing, basic text/image support
- **Dependencies**: None (pure Go)
- **Use cases**: Default backend and pure-Go compatibility

### Skia (Future)

- **Type**: High-quality renderer with GPU acceleration
- **Status**: 🚧 Scaffold implemented, awaiting Skia bindings
- **Capabilities**: Anti-aliasing, GPU acceleration, Advanced text shaping
- **Dependencies**: Skia library, CGO
- **Use cases**: High-quality output, interactive applications

## Usage

### Command Line

```bash
# List available backends
go run ./examples/backends/demo/main.go --list

# Show capability matrix
go run ./examples/backends/demo/main.go --capabilities

# Use specific backend
go run ./examples/lines/basic-backend/main.go --backend=agg
```

### Programmatic

```go
import (
    "matplotlib-go/backends"
    _ "matplotlib-go/backends/agg"     // Optional advanced backend
    _ "matplotlib-go/backends/gobasic" // Register default backend
)

// Auto-select best available backend (falls back to GoBasic)
backend, err := backends.GetBestBackend(nil)

// Create renderer
config := backends.Config{
    Width: 800, Height: 600,
    Background: render.Color{R: 1, G: 1, B: 1, A: 1},
}
renderer, err := backends.Create(backend, config)

// Use with figures
err = core.SavePNG(fig, renderer, "output.png")

// Or let SaveFig pick the exporter from the file extension:
err = core.SaveFig(fig, renderer, "plot.png") // dispatches to SavePNG
err = core.SaveFig(fig, renderer, "plot.svg") // dispatches to SaveSVG
err = core.SaveFig(fig, renderer, "plot.pdf",
    render.WithPDFMetadata(map[string]string{"Title": "Example"}),
)
```

Format-specific save options share the `render.SaveOption` surface. SVG, PDF,
PostScript, and PGF options can be passed through `core.SaveFig`,
`pyplot.Savefig`, the backend registry save dispatch, and headless manager save
tools. Options are validated against the selected extension, so unsupported
combinations fail explicitly instead of being ignored.

## Backend Capabilities

| Backend | Anti-aliasing | GPU Accel | Text Shaping | Vector Output |
| ------- | ------------- | --------- | ------------ | ------------- |
| AGG     | ✅            | ❌        | ✅           | ❌            |
| GoBasic | ✅            | ❌        | ❌           | ❌            |
| Skia    | ✅            | ✅        | ✅           | ✅            |

## Adding New Backends

1. Create package in `backends/newbackend/`
2. Implement `render.Renderer` interface
3. Register in `init()` function:
   ```go
   func init() {
       backends.Register(backends.NewBackend, &backends.BackendInfo{
           Name: "New Backend",
           Capabilities: []backends.Capability{...},
           Factory: func(config backends.Config) (render.Renderer, error) {
               return New(config)
           },
           Available: checkAvailability(),
       })
   }
   ```

## Build Tags

Use build tags for optional backends:

- `go build -tags skia ./...` - Include Skia backend
- `go build ./...` - Register and use GoBasic by default

## Testing

The backend system includes a comprehensive test suite:

```bash
go test ./backends/...        # Test backend system
just backend-info             # Show available backends
```
