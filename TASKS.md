# Matplotlib-Go Development Plan

This plan prioritizes getting useful plotting functionality working quickly with your existing GoBasic renderer, rather than diving deep into rendering backends. We'll build core plot types first, then enhance with axes features, and defer complex backend work until later.

---

# ✅ Foundation (COMPLETED)

**What we have working:**

- ✅ Artist hierarchy (Figure→Axes→Artists) with proper traversal
- ✅ Transform system (Linear/Log scales, data→pixel transforms)
- ✅ GoBasic renderer using `golang.org/x/image/vector`
- ✅ Line2D artist with stroke support (joins, caps, dashes)
- ✅ Golden image testing infrastructure
- ✅ Working example: `examples/lines/basic.go` produces clean line plots

**Current capabilities:**

```go
// Basic line plots work
line := &core.Line2D{
    XY: []geom.Pt{{0,0}, {1,0.2}, {3,0.9}},
    W: 2.0,
    Col: render.Color{R: 0, G: 0, B: 0, A: 1},
}
ax.Add(line)
```

---

# Phase 1: Core Plot Types (IMMEDIATE PRIORITY)

**Goal:** Get the most commonly used plot types working

### 1.1 Scatter Plots

- [x] `Scatter2D` artist with point/marker rendering
- [x] Basic marker shapes: circle, square, triangle, diamond, plus, cross
- [x] Variable marker sizes and colors per point
- [x] Edge colors and stroke support for marker outlines
- [x] Alpha transparency support
- [x] Proper bounds calculation
- [x] Comprehensive unit tests and golden tests
- [x] Example: `examples/scatter/basic.go`

### 1.2 Bar Charts

- [x] `Bar2D` artist using rectangle patches
- [x] Vertical and horizontal bars
- [x] Grouped bars (multiple series)
- [x] Comprehensive unit tests and golden tests
- [x] Edge colors and transparency support
- [x] Variable bar widths and colors per bar
- [x] Proper bounds calculation
- [x] Example: `examples/bar/basic.go`

### 1.3 Fill Operations

- [x] `Fill2D` artist for area plots and fill_between
- [x] Alpha transparency support
- [x] Edge colors and stroke support for fill outlines
- [x] Multiple fill regions on same axes
- [x] Proper bounds calculation
- [x] Comprehensive unit tests and golden tests
- [x] Performance optimization for large datasets
- [x] Example: `examples/fill/basic.go`

### 1.4 Multiple Series Support

- [x] Plot multiple lines/scatter/bars on same axes
- [x] Automatic color cycling for series
- [x] Series labels for legend preparation
- [x] Example: `examples/multi/basic.go`

**Exit Criteria:**

- Basic scatter, bar, and fill plots render correctly
- Multiple plot types can coexist on same axes
- All examples work with `go run main.go`

---

# Phase 2: Axes Features (HIGH PRIORITY)

**Goal:** Make plots look professional with proper axes

### 2.1 Axis Rendering

- [ ] Draw actual axis lines (spines) using existing line drawing
- [ ] Tick marks (major/minor) positioned correctly
- [ ] Use existing LinearLocator/LogLocator for tick placement
- [ ] Example: `examples/axes/spines.go`

### 2.2 Grid Lines

- [ ] Major and minor grid lines
- [ ] Grid styling (color, alpha, line style)
- [ ] Grid behind/in front of data
- [ ] Example: `examples/axes/grid.go`

### 2.3 Axis Limits and Scaling

- [ ] `SetXLim(min, max)` and `SetYLim(min, max)` methods
- [ ] Auto-scaling based on data bounds
- [ ] Margin handling around data
- [ ] Example: `examples/axes/limits.go`

### 2.4 Basic Text Labels

- [ ] Simple text rendering using basic ASCII fonts
- [ ] Title, xlabel, ylabel placement
- [ ] Tick labels using existing formatters
- [ ] Text alignment and rotation basics
- [ ] Example: `examples/axes/labels.go`

**Exit Criteria:**

- Plots have proper axis lines, ticks, and labels
- Grid lines work and look good
- Axis limits can be set manually or auto-computed

---

# Phase 3: Additional Plot Types (MEDIUM PRIORITY)

**Goal:** Expand the plotting vocabulary

### 3.1 Histograms

- [ ] `Histogram` artist built on bar charts
- [ ] Automatic binning with various bin strategies
- [ ] Density normalization options
- [ ] Example: `examples/histogram/basic.go`

### 3.2 Box Plots

- [ ] `BoxPlot` artist for statistical visualization
- [ ] Quartiles, whiskers, outliers
- [ ] Multiple box plots per axes
- [ ] Example: `examples/boxplot/basic.go`

### 3.3 Error Bars

- [ ] `ErrorBar` artist for scientific plots
- [ ] X and Y error bars with caps
- [ ] Combine with scatter/line plots
- [ ] Example: `examples/errorbar/basic.go`

### 3.4 Images and Heatmaps

- [ ] `Image2D` artist for imshow functionality
- [ ] Basic nearest-neighbor image scaling
- [ ] Simple colormaps (grayscale, basic colors)
- [ ] Example: `examples/image/basic.go`

**Exit Criteria:**

- Common scientific plot types work
- Examples demonstrate real-world use cases

---

# Phase 4: Layout & Annotation (MEDIUM PRIORITY)

**Goal:** Polish and professional presentation

### 4.1 Subplots

- [ ] `Subplot` functionality for multiple axes grids
- [ ] Automatic spacing and layout
- [ ] Shared axes between subplots
- [ ] Example: `examples/subplots/basic.go`

### 4.2 Legends

- [ ] `Legend` artist with automatic entries
- [ ] Legend placement and styling
- [ ] Line/marker/patch legend entries
- [ ] Example: `examples/legend/basic.go`

### 4.3 Text Annotations

- [ ] `Text` artist for arbitrary text placement
- [ ] Arrow annotations pointing to data
- [ ] Math symbols and Greek letters (basic)
- [ ] Example: `examples/annotation/basic.go`

### 4.4 Colorbars

- [ ] `Colorbar` artist for heatmaps
- [ ] Automatic scaling and labels
- [ ] Various colormap support
- [ ] Example: `examples/colorbar/basic.go`

**Exit Criteria:**

- Multi-panel figures work well
- Plots are publication-ready with legends and annotations

---

# Phase 5: Export & Polish (LOW PRIORITY)

**Goal:** Multiple output formats and refinements

### 5.1 SVG Export

- [ ] SVG backend using path recording
- [ ] Vector output for publications
- [ ] Text as actual text (not paths)
- [ ] Example: `examples/export/svg.go`

### 5.2 Interactive Features

- [ ] Basic pan/zoom using mouse
- [ ] Simple event handling
- [ ] Real-time plot updates
- [ ] Example: `examples/interactive/basic.go`

### 5.3 Styling and Themes

- [ ] Style sheets and themes
- [ ] Color palettes and defaults
- [ ] Publication-ready themes
- [ ] Example: `examples/styling/themes.go`

**Exit Criteria:**

- Multiple export formats work
- Library feels polished and complete

---

# Phase 6: Advanced Backends (FUTURE)

**Goal:** High-performance and specialized rendering (deferred)

### 6.1 Performance Optimization

- [ ] Optimize GoBasic renderer performance
- [ ] Path simplification and culling
- [ ] Large dataset handling

### 6.2 Alternative Backends (Future Consideration)

- [ ] AGG backend for anti-aliasing (if needed)
- [ ] Skia backend for GPU acceleration (if needed)
- [ ] PDF export for publications (if needed)

### 6.3 Advanced Text

- [ ] Font loading and management
- [ ] Complex text shaping
- [ ] LaTeX-style math rendering

**Exit Criteria:**

- Only implement if performance or quality demands it
- GoBasic should handle most use cases well

---

# Development Guidelines

## Testing Strategy

- Golden image tests for all plot types
- Property-based tests for data ranges
- Visual regression testing
- `go test ./...` runs all tests

## API Design Principles

- Follow matplotlib conventions where sensible
- Use functional options for configuration
- Keep simple cases simple
- Provide escape hatches for complex cases

## Performance Goals

- Handle datasets up to 100k points smoothly
- Sub-second rendering for typical plots
- Memory efficient for long-running applications

## Examples-Driven Development

- Every feature gets a working example
- Examples serve as integration tests
- README showcases example gallery
- Examples demonstrate real-world usage

---

This plan gets you to a fully functional plotting library quickly while keeping the foundation solid for future enhancements.
