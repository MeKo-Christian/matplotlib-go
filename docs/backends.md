# Backend System

Matplotlib-Go uses a pluggable backend architecture that allows different rendering engines to be used interchangeably.

## Available Backends

### GoBasic (Default)
- **Type**: Pure Go renderer using `golang.org/x/image/vector`
- **Status**: ✅ Fully implemented
- **Capabilities**: Anti-aliasing, Path clipping, Vector output
- **Dependencies**: None (pure Go)
- **Use cases**: Development, testing, basic plotting

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
go run ./examples/lines/basic-backend/main.go --backend=gobasic
```

### Programmatic
```go
import "matplotlib-go/backends"
import _ "matplotlib-go/backends/gobasic" // Register backend

// Auto-select best backend
backend, err := backends.GetRecommendedBackend("publication")

// Create renderer
config := backends.Config{
    Width: 800, Height: 600,
    Background: render.Color{R: 1, G: 1, B: 1, A: 1},
}
renderer, err := backends.Create(backend, config)

// Use with figures
err = core.SavePNG(fig, renderer, "output.png")
```

## Backend Capabilities

| Backend | Anti-aliasing | GPU Accel | Text Shaping | Vector Output |
|---------|---------------|-----------|--------------|---------------|
| GoBasic | ✅            | ❌        | ❌           | ✅            |
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
- `go build ./...` - GoBasic only (default)

## Testing

The backend system includes a comprehensive test suite:
```bash
go test ./backends/...        # Test backend system
make backend-info             # Show available backends
just backend-info             # Alternative using justfile
```