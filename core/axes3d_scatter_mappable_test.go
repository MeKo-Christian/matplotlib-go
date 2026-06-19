package core

import (
	"sort"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestAxes3DScatterDefaultColorUsesShapeCycle(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	palette := fig.RC.Palette()

	line := ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	scatter := ax.Scatter3D([]float64{0.5}, []float64{0.2}, []float64{0.1})
	nextLine := ax.Plot3D([]float64{0, 1}, []float64{1, 0}, []float64{0, 1})

	if got, want := line.Col, palette[0]; got != want {
		t.Fatalf("first 3D line color = %+v, want %+v", got, want)
	}
	if got, want := scatter.Color, palette[0]; got != want {
		t.Fatalf("3D scatter color = %+v, want independent shape cycle first color %+v", got, want)
	}
	if got, want := nextLine.Col, palette[1]; got != want {
		t.Fatalf("second 3D line color = %+v, want line cycle second color %+v", got, want)
	}
}

func TestAxes3DScatterUsesMatplotlibDefaultSize(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	scatter := ax.Scatter3D([]float64{0.5}, []float64{0.2}, []float64{0.1})
	if scatter == nil {
		t.Fatal("Scatter3D returned nil")
	}
	if got, want := scatter.Size, 20.0; got != want {
		t.Fatalf("3D scatter default size = %v, want Matplotlib Axes3D.scatter s=%v", got, want)
	}

	explicit := 42.0
	scatter = ax.Scatter3D(
		[]float64{0.5},
		[]float64{0.2},
		[]float64{0.1},
		ScatterOptions{Size: &explicit},
	)
	if scatter == nil {
		t.Fatal("Scatter3D with explicit size returned nil")
	}
	if got := scatter.Size; got != explicit {
		t.Fatalf("3D scatter explicit size = %v, want %v", got, explicit)
	}
}

func TestAxes3DScatterDepthShadesAndSortsMarkersLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	scatter := ax.Scatter3D(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{0, 1},
	)
	if scatter == nil {
		t.Fatal("Scatter3D returned nil")
	}
	if got, want := len(scatter.Colors), 2; got != want {
		t.Fatalf("3D scatter per-marker colors = %d, want %d depth-shaded colors", got, want)
	}
	if !approx(scatter.Colors[0].A, 0.3, 1e-12) || !approx(scatter.Colors[1].A, 1.0, 1e-12) {
		t.Fatalf("3D scatter depth-shaded alphas = %.12g, %.12g; want Matplotlib z-sorted alpha range 0.3..1.0", scatter.Colors[0].A, scatter.Colors[1].A)
	}
	if got, want := len(scatter.EdgeColors), 2; got != want {
		t.Fatalf("3D scatter per-marker edge colors = %d, want %d depth-shaded edge colors", got, want)
	}
	if !approx(scatter.EdgeColors[0].A, 0.3, 1e-12) || !approx(scatter.EdgeColors[1].A, 1.0, 1e-12) {
		t.Fatalf("3D scatter depth-shaded edge alphas = %.12g, %.12g; want Matplotlib z-sorted alpha range 0.3..1.0", scatter.EdgeColors[0].A, scatter.EdgeColors[1].A)
	}
}

func TestAxes3DScatterScalarValuesKeepMappedColorsThroughDepthShade(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	values := []float64{0, 1}
	scatter := ax.Scatter3D(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{0, 1},
		ScatterOptions{ScalarValues: values, Colormap: "viridis"},
	)
	if scatter == nil {
		t.Fatal("Scatter3D returned nil")
	}
	if got, want := len(scatter.ScalarValues), len(values); got != want {
		t.Fatalf("3D scatter scalar array len = %d, want %d visible values", got, want)
	}
	for _, want := range values {
		if !containsFloat64Approx(scatter.ScalarValues, want, 1e-12) {
			t.Fatalf("3D scatter scalar array = %v, want values %v preserved through z sort", scatter.ScalarValues, values)
		}
	}
	if got, want := len(scatter.Colors), len(values); got != want {
		t.Fatalf("3D scatter mapped colors = %d, want %d depth-shaded mapped colors", got, want)
	}
	if len(scatter.Colors) == 2 && scatter.Colors[0].R == scatter.Colors[1].R && scatter.Colors[0].G == scatter.Colors[1].G && scatter.Colors[0].B == scatter.Colors[1].B {
		t.Fatalf("3D scatter mapped colors = %+v, want distinct scalar-mapped RGB values", scatter.Colors)
	}
	if !approx(scatter.Colors[0].A, 0.3, 1e-12) || !approx(scatter.Colors[1].A, 1.0, 1e-12) {
		t.Fatalf("3D scatter scalar depth-shaded alphas = %.12g, %.12g; want Matplotlib z-sorted alpha range 0.3..1.0", scatter.Colors[0].A, scatter.Colors[1].A)
	}
	mapping := scatter.ScalarMap()
	if mapping.Colormap != "viridis" || mapping.VMin != 0 || mapping.VMax != 1 {
		t.Fatalf("3D scatter scalar map = %+v, want viridis range 0..1 for colorbar", mapping)
	}
	if got, want := scatter.GetArray(), scatter.ScalarValues; len(got) != len(want) {
		t.Fatalf("3D scatter GetArray = %v, want %v", got, want)
	}
	for _, want := range scatter.ScalarValues {
		if !containsFloat64Approx(scatter.GetArray(), want, 1e-12) {
			t.Fatalf("3D scatter GetArray = %v, want values %v", scatter.GetArray(), scatter.ScalarValues)
		}
	}
	cbAx := fig.AddColorbar(ax.Axes, scatter)
	if cbAx == nil || len(cbAx.Artists) == 0 {
		t.Fatal("AddColorbar returned no colorbar axes for 3D scatter scalar mappable")
	}
	cb, ok := cbAx.Artists[0].(*Colorbar)
	if !ok {
		t.Fatalf("3D scatter colorbar artist = %T, want *Colorbar", cbAx.Artists[0])
	}
	if cb.Mapping.Colormap != "viridis" || cb.Mapping.VMin != 0 || cb.Mapping.VMax != 1 {
		t.Fatalf("3D scatter colorbar mapping = %+v, want viridis range 0..1", cb.Mapping)
	}
	pc := scatter.toPathCollection(&render.NullRenderer{}, createTestDrawContext())
	if got, want := pc.GetArray(), scatter.ScalarValues; len(got) != len(want) {
		t.Fatalf("3D scatter path collection scalar array = %v, want %v", got, want)
	}
	for i, want := range scatter.Colors {
		if got := pc.FaceColors[i]; got != want {
			t.Fatalf("3D scatter path face color %d = %+v, want depth-shaded mapped color %+v", i, got, want)
		}
		if got := pc.EdgeColors[i]; got != want {
			t.Fatalf("3D scatter default edge color %d = %+v, want face color %+v", i, got, want)
		}
	}
}

func TestAxes3DScatterScalarValuesFollowAxLimClip(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1)

	x := []float64{0.25, 0.75, 2}
	y := []float64{0, 0.5, 0}
	z := []float64{0, 1, 0}
	values := []float64{2, 8, 9}
	vmin := 0.0
	vmax := 10.0
	scatter := ax.Scatter3D(
		x,
		y,
		z,
		ScatterOptions{ScalarValues: values, Colormap: "viridis", VMin: &vmin, VMax: &vmax, AxLimClip: true},
	)
	if scatter == nil {
		t.Fatal("Scatter3D returned nil")
	}
	if got, want := len(scatter.XY), 2; got != want {
		t.Fatalf("3D scatter clipped points = %d, want %d", got, want)
	}
	projected := ax.projectedScatterData(x, y, z, true)
	sort.SliceStable(projected, func(i, j int) bool {
		return projected[i].depth > projected[j].depth
	})
	wantValues := make([]float64, len(projected))
	for i, point := range projected {
		wantValues[i] = values[point.index]
	}
	if len(scatter.ScalarValues) != len(wantValues) {
		t.Fatalf("3D scatter clipped scalar array = %v, want visible sorted values %v", scatter.ScalarValues, wantValues)
	}
	for i, want := range wantValues {
		if !approx(scatter.ScalarValues[i], want, 1e-12) {
			t.Fatalf("3D scatter clipped scalar array = %v, want visible sorted values %v", scatter.ScalarValues, wantValues)
		}
	}
	if got, want := scatter.GetArray(), scatter.ScalarValues; len(got) != len(want) {
		t.Fatalf("3D scatter clipped GetArray = %v, want %v", got, want)
	}
	cbAx := fig.AddColorbar(ax.Axes, scatter)
	if cbAx == nil || len(cbAx.Artists) == 0 {
		t.Fatal("AddColorbar returned no colorbar axes for clipped 3D scatter")
	}
	cb, ok := cbAx.Artists[0].(*Colorbar)
	if !ok {
		t.Fatalf("clipped 3D scatter colorbar artist = %T, want *Colorbar", cbAx.Artists[0])
	}
	if cb.Mappable != scatter {
		t.Fatalf("clipped 3D scatter colorbar mappable = %p, want scatter %p", cb.Mappable, scatter)
	}
	if cb.Mapping.Colormap != "viridis" || cb.Mapping.VMin != vmin || cb.Mapping.VMax != vmax {
		t.Fatalf("clipped 3D scatter colorbar mapping = %+v, want viridis range %.1f..%.1f", cb.Mapping, vmin, vmax)
	}
}

func TestAxes3DScalarMappableContractAudit(t *testing.T) {
	type scalarArrayMappable interface {
		ScalarMappable
		GetArray() []float64
	}

	cmap := "viridis"
	vmin := 0.0
	vmax := 10.0
	gridX := []float64{0, 1}
	gridY := []float64{0, 1}
	gridZ := [][]float64{{0, 2}, {4, 6}}
	tri := Triangulation{
		X:         []float64{0, 1, 0, 1},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {1, 3, 2}},
	}

	tests := []struct {
		name    string
		make    func(*Axes3D) scalarArrayMappable
		wantLen int
	}{
		{
			name: "Surface",
			make: func(ax *Axes3D) scalarArrayMappable {
				return ax.Surface(gridX, gridY, gridZ, PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax})
			},
			wantLen: 1,
		},
		{
			name: "Trisurf",
			make: func(ax *Axes3D) scalarArrayMappable {
				return ax.Trisurf(tri, []float64{0, 2, 4, 6}, PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax})
			},
			wantLen: 2,
		},
		{
			name: "Contour",
			make: func(ax *Axes3D) scalarArrayMappable {
				return ax.Contour(gridX, gridY, gridZ, PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax, Levels: []float64{2, 4}})
			},
			wantLen: 2,
		},
		{
			name: "Contourf",
			make: func(ax *Axes3D) scalarArrayMappable {
				return ax.Contourf(gridX, gridY, gridZ, PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax, Levels: []float64{0, 2, 4, 6}})
			},
			wantLen: 3,
		},
		{
			name: "TriContour",
			make: func(ax *Axes3D) scalarArrayMappable {
				return ax.TriContour(tri, []float64{0, 2, 4, 6}, PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax, Levels: []float64{2, 4}})
			},
			wantLen: 2,
		},
		{
			name: "TriContourf",
			make: func(ax *Axes3D) scalarArrayMappable {
				return ax.TriContourf(tri, []float64{0, 2, 4, 6}, PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax, Levels: []float64{0, 2, 4, 6}})
			},
			wantLen: 3,
		},
		{
			name: "Scatter3D",
			make: func(ax *Axes3D) scalarArrayMappable {
				return ax.Scatter3D(
					[]float64{0, 1},
					[]float64{0, 1},
					[]float64{0, 1},
					ScatterOptions{ScalarValues: []float64{2, 8}, Colormap: cmap, VMin: &vmin, VMax: &vmax},
				)
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fig := NewFigure(640, 480)
			ax, err := fig.AddAxes3D(unitRect())
			if err != nil {
				t.Fatalf("AddAxes3D: %v", err)
			}
			mappable := tt.make(ax)
			if mappable == nil {
				t.Fatalf("%s returned nil", tt.name)
			}
			mapping := mappable.ScalarMap()
			if mapping.Colormap != cmap || mapping.VMin != vmin || mapping.VMax != vmax {
				t.Fatalf("%s scalar map = %+v, want cmap=%q range %.1f..%.1f", tt.name, mapping, cmap, vmin, vmax)
			}
			array := mappable.GetArray()
			if got := len(array); got != tt.wantLen {
				t.Fatalf("%s scalar array = %v, len %d, want len %d", tt.name, array, got, tt.wantLen)
			}
		})
	}
}

func TestAxes3DScalarMappableHelpersApplyAlphaToMappedColors(t *testing.T) {
	cmap := "viridis"
	alpha := 0.4
	gridX := []float64{0, 1}
	gridY := []float64{0, 1}
	gridZ := [][]float64{{0, 2}, {4, 6}}
	tri := Triangulation{
		X:         []float64{0, 1, 0, 1},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {1, 3, 2}},
	}

	tests := []struct {
		name   string
		colors func(*Axes3D) []render.Color
	}{
		{
			name: "Surface",
			colors: func(ax *Axes3D) []render.Color {
				return ax.Surface(gridX, gridY, gridZ, PlotOptions{Colormap: &cmap, Alpha: &alpha}).FaceColors
			},
		},
		{
			name: "Trisurf",
			colors: func(ax *Axes3D) []render.Color {
				return ax.Trisurf(tri, []float64{0, 2, 4, 6}, PlotOptions{Colormap: &cmap, Alpha: &alpha}).FaceColors
			},
		},
		{
			name: "Contour",
			colors: func(ax *Axes3D) []render.Color {
				return ax.Contour(gridX, gridY, gridZ, PlotOptions{Colormap: &cmap, Alpha: &alpha, Levels: []float64{2, 4}}).Colors
			},
		},
		{
			name: "Contourf",
			colors: func(ax *Axes3D) []render.Color {
				return ax.Contourf(gridX, gridY, gridZ, PlotOptions{Colormap: &cmap, Alpha: &alpha, Levels: []float64{0, 2, 4, 6}}).FaceColors
			},
		},
		{
			name: "TriContour",
			colors: func(ax *Axes3D) []render.Color {
				return ax.TriContour(tri, []float64{0, 2, 4, 6}, PlotOptions{Colormap: &cmap, Alpha: &alpha, Levels: []float64{2, 4}}).Colors
			},
		},
		{
			name: "TriContourf",
			colors: func(ax *Axes3D) []render.Color {
				return ax.TriContourf(tri, []float64{0, 2, 4, 6}, PlotOptions{Colormap: &cmap, Alpha: &alpha, Levels: []float64{0, 2, 4, 6}}).FaceColors
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fig := NewFigure(640, 480)
			ax, err := fig.AddAxes3D(unitRect())
			if err != nil {
				t.Fatalf("AddAxes3D: %v", err)
			}
			colors := tt.colors(ax)
			if len(colors) == 0 {
				t.Fatalf("%s produced no mapped colors", tt.name)
			}
			for i, color := range colors {
				if !approx(color.A, alpha, 1e-12) {
					t.Fatalf("%s mapped color %d alpha = %.12g, want %.12g", tt.name, i, color.A, alpha)
				}
			}
		})
	}

	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	scatter := ax.Scatter3D(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{0, 1},
		ScatterOptions{ScalarValues: []float64{0, 1}, Colormap: cmap, Alpha: &alpha},
	)
	if scatter == nil {
		t.Fatal("Scatter3D returned nil")
	}
	if got, want := len(scatter.Colors), 2; got != want {
		t.Fatalf("scatter mapped colors = %d, want %d", got, want)
	}
	if !approx(scatter.Colors[0].A, alpha*0.3, 1e-12) || !approx(scatter.Colors[1].A, alpha, 1e-12) {
		t.Fatalf("scatter mapped depth-shaded alphas = %.12g, %.12g; want %.12g..%.12g", scatter.Colors[0].A, scatter.Colors[1].A, alpha*0.3, alpha)
	}
}

func TestAxes3DCollectionMappablesCreateColorbars(t *testing.T) {
	cmap := "plasma"
	vmin := 0.0
	vmax := 10.0
	gridX := []float64{0, 1}
	gridY := []float64{0, 1}
	gridZ := [][]float64{{0, 2}, {4, 6}}
	tri := Triangulation{
		X:         []float64{0, 1, 0, 1},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {1, 3, 2}},
	}

	tests := []struct {
		name string
		make func(*Axes3D) ScalarMappable
	}{
		{
			name: "Surface",
			make: func(ax *Axes3D) ScalarMappable {
				return ax.Surface(gridX, gridY, gridZ, PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax})
			},
		},
		{
			name: "Trisurf",
			make: func(ax *Axes3D) ScalarMappable {
				return ax.Trisurf(tri, []float64{0, 2, 4, 6}, PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax})
			},
		},
		{
			name: "Contour",
			make: func(ax *Axes3D) ScalarMappable {
				return ax.Contour(gridX, gridY, gridZ, PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax, Levels: []float64{2, 4}})
			},
		},
		{
			name: "Contourf",
			make: func(ax *Axes3D) ScalarMappable {
				return ax.Contourf(gridX, gridY, gridZ, PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax, Levels: []float64{0, 2, 4, 6}})
			},
		},
		{
			name: "TriContour",
			make: func(ax *Axes3D) ScalarMappable {
				return ax.TriContour(tri, []float64{0, 2, 4, 6}, PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax, Levels: []float64{2, 4}})
			},
		},
		{
			name: "TriContourf",
			make: func(ax *Axes3D) ScalarMappable {
				return ax.TriContourf(tri, []float64{0, 2, 4, 6}, PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax, Levels: []float64{0, 2, 4, 6}})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fig := NewFigure(640, 480)
			ax, err := fig.AddAxes3D(unitRect())
			if err != nil {
				t.Fatalf("AddAxes3D: %v", err)
			}
			mappable := tt.make(ax)
			if mappable == nil {
				t.Fatalf("%s returned nil", tt.name)
			}
			cbAx := fig.AddColorbar(ax.Axes, mappable)
			if cbAx == nil || len(cbAx.Artists) == 0 {
				t.Fatalf("AddColorbar returned no colorbar axes for %s", tt.name)
			}
			cb, ok := cbAx.Artists[0].(*Colorbar)
			if !ok {
				t.Fatalf("%s colorbar artist = %T, want *Colorbar", tt.name, cbAx.Artists[0])
			}
			if cb.Mappable != mappable {
				t.Fatalf("%s colorbar mappable = %p, want returned collection %p", tt.name, cb.Mappable, mappable)
			}
			if cb.Mapping.Colormap != cmap || cb.Mapping.VMin != vmin || cb.Mapping.VMax != vmax {
				t.Fatalf("%s colorbar mapping = %+v, want cmap=%q range %.1f..%.1f", tt.name, cb.Mapping, cmap, vmin, vmax)
			}
			yMin, yMax := cbAx.YScale.Domain()
			if !approx(yMin, vmin, 1e-12) || !approx(yMax, vmax, 1e-12) {
				t.Fatalf("%s colorbar y domain = %.12g..%.12g, want %.12g..%.12g", tt.name, yMin, yMax, vmin, vmax)
			}
		})
	}
}

func TestAxes3DCollectionColorbarSyncsMutableMapping(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	cmap := "viridis"
	surface := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{Colormap: &cmap},
	)
	if surface == nil {
		t.Fatal("Surface returned nil")
	}
	cbAx := fig.AddColorbar(ax.Axes, surface)
	if cbAx == nil || len(cbAx.Artists) == 0 {
		t.Fatal("AddColorbar returned no colorbar axes for mutable 3D surface")
	}

	if err := surface.SetCLim(-1, 2); err != nil {
		t.Fatalf("SetCLim: %v", err)
	}
	surface.SetColormap("plasma")
	DrawFigure(fig, &colorbarRecordingRenderer{})

	yMin, yMax := cbAx.YScale.Domain()
	if yMin != -1 || yMax != 2 {
		t.Fatalf("synced 3D surface colorbar limits = %v..%v, want -1..2", yMin, yMax)
	}
	cb, ok := cbAx.Artists[0].(*Colorbar)
	if !ok {
		t.Fatalf("3D surface colorbar artist = %T, want *Colorbar", cbAx.Artists[0])
	}
	if cb.Mapping.VMin != -1 || cb.Mapping.VMax != 2 {
		t.Fatalf("synced 3D surface colorbar mapping = %+v, want -1..2", cb.Mapping)
	}
	if cb.Mapping.Colormap != "plasma" {
		t.Fatalf("synced 3D surface colorbar colormap = %q, want plasma", cb.Mapping.Colormap)
	}
}

func TestAxes3DUnsupportedColorbarHelpersExposeNoScalarData(t *testing.T) {
	type scalarArrayMappable interface {
		ScalarMappable
		GetArray() []float64
	}

	tests := []struct {
		name string
		make func(*Axes3D) scalarArrayMappable
	}{
		{
			name: "Wireframe",
			make: func(ax *Axes3D) scalarArrayMappable {
				return ax.Wireframe([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}})
			},
		},
		{
			name: "Quiver3D",
			make: func(ax *Axes3D) scalarArrayMappable {
				return ax.Quiver3D([]float64{0}, []float64{0}, []float64{0}, []float64{1}, []float64{0}, []float64{0})
			},
		},
		{
			name: "ErrorBar3D",
			make: func(ax *Axes3D) scalarArrayMappable {
				return ax.ErrorBar3D([]float64{0}, []float64{0}, []float64{0}, nil, nil, []float64{0.1})
			},
		},
		{
			name: "Stem3D lines",
			make: func(ax *Axes3D) scalarArrayMappable {
				container := ax.Stem3D([]float64{0, 1}, []float64{0, 1}, []float64{1, 2})
				if container == nil {
					return nil
				}
				return container.StemLines
			},
		},
		{
			name: "Stem3D markers",
			make: func(ax *Axes3D) scalarArrayMappable {
				container := ax.Stem3D([]float64{0, 1}, []float64{0, 1}, []float64{1, 2})
				if container == nil {
					return nil
				}
				return container.MarkerCollection
			},
		},
		{
			name: "FillBetween3D",
			make: func(ax *Axes3D) scalarArrayMappable {
				return ax.FillBetween3D(
					[]float64{0, 1, 2},
					[]float64{0, 0, 0},
					[]float64{1, 1, 1},
					[]float64{0, 1, 2},
					[]float64{1, 1, 1},
					[]float64{0, 0, 0},
					FillBetween3DOptions{Mode: FillBetween3DModeQuad},
				)
			},
		},
		{
			name: "Bar3D edges",
			make: func(ax *Axes3D) scalarArrayMappable {
				return ax.Bar3D([]float64{0}, []float64{0}, []float64{0}, []float64{1}, []float64{1}, []float64{1})
			},
		},
		{
			name: "Bar3D faces",
			make: func(ax *Axes3D) scalarArrayMappable {
				if ax.Bar3D([]float64{0}, []float64{0}, []float64{0}, []float64{1}, []float64{1}, []float64{1}) == nil {
					return nil
				}
				return latestBar3DFaceCollection(t, ax, 6)
			},
		},
		{
			name: "Voxels",
			make: func(ax *Axes3D) scalarArrayMappable {
				voxels := ax.Voxels([][][]bool{{{true}}})
				return voxels[[3]int{0, 0, 0}]
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fig := NewFigure(640, 480)
			ax, err := fig.AddAxes3D(unitRect())
			if err != nil {
				t.Fatalf("AddAxes3D: %v", err)
			}
			mappable := tt.make(ax)
			if mappable == nil {
				t.Fatalf("%s returned nil", tt.name)
			}
			if array := mappable.GetArray(); len(array) != 0 {
				t.Fatalf("%s scalar array = %v, want no data-backed colorbar array", tt.name, array)
			}
			mapping := mappable.ScalarMap()
			if mapping.Colormap != "" || mapping.Norm != nil || mapping.VMin != 0 || mapping.VMax != 0 {
				t.Fatalf("%s scalar map = %+v, want no scalar-map metadata", tt.name, mapping)
			}
			cbAx := fig.AddColorbar(ax.Axes, mappable)
			if cbAx == nil || len(cbAx.Artists) == 0 {
				t.Fatalf("AddColorbar returned no axes for empty %s mappable", tt.name)
			}
			cb, ok := cbAx.Artists[0].(*Colorbar)
			if !ok {
				t.Fatalf("%s colorbar artist = %T, want *Colorbar", tt.name, cbAx.Artists[0])
			}
			if cb.Mapping.Colormap != "viridis" || cb.Mapping.VMin != 0 || cb.Mapping.VMax != 1 {
				t.Fatalf("%s forced colorbar mapping = %+v, want generic default mapping", tt.name, cb.Mapping)
			}
		})
	}

	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	line := ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	if line == nil {
		t.Fatal("Plot3D returned nil")
	}
	if _, ok := any(line).(ScalarMappable); ok {
		t.Fatal("Plot3D unexpectedly implements ScalarMappable")
	}
}

func TestAxes3DScatterAxLimClipDropsOutsideMarkers(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1)

	scatter := ax.Scatter3D(
		[]float64{0.25, 2},
		[]float64{0, 0},
		[]float64{0, 0},
		ScatterOptions{AxLimClip: true},
	)
	if scatter == nil {
		t.Fatal("Scatter3D returned nil")
	}
	if got, want := len(scatter.XY), 1; got != want {
		t.Fatalf("Scatter3D clipped markers = %d, want %d", got, want)
	}
	want := ax.ProjectPoint(0.25, 0, 0)
	if got := scatter.XY[0]; !approx(got.X, want.X, 1e-12) || !approx(got.Y, want.Y, 1e-12) {
		t.Fatalf("Scatter3D clipped marker = %+v, want %+v", got, want)
	}
}

func TestAxes3DCollectionsUseComputedDepthZOrder(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	line := ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	low := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 0}, {0, 0}},
	)
	high := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{1, 1}, {1, 1}},
	)
	if line == nil || low == nil || high == nil {
		t.Fatalf("expected line and surface artists, got line=%v low=%v high=%v", line, low, high)
	}
	if !(low.Z() > line.Z() && high.Z() > line.Z()) {
		t.Fatalf("3D surface zorders = low %.6g high %.6g, want both above 3D line %.6g like Matplotlib computed_zorder", low.Z(), high.Z(), line.Z())
	}
	if !(high.Z() > low.Z()) {
		t.Fatalf("3D surface zorders = low %.6g high %.6g, want higher projected plane drawn after lower plane", low.Z(), high.Z())
	}
}
