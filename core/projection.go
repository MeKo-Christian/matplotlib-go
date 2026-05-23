package core

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/transform"
)

// ProjectionFactory constructs a fresh projection instance.
type ProjectionFactory func() Projection

// Projection customizes how an axes maps data into axes coordinates.
type Projection interface {
	Name() string
	ConfigureAxes(ax *Axes)
	DataToAxes(ax *Axes) transform.T
}

// ProjectionCloner preserves projection-local state when axes create derived
// contexts or overlay axes.
type ProjectionCloner interface {
	CloneProjection() Projection
}

var (
	projectionRegistryMu sync.RWMutex
	projectionRegistry   = map[string]ProjectionFactory{}
)

func init() {
	mustRegisterProjection("rectilinear", func() Projection { return rectilinearProjection{} })
	mustRegisterProjection("polar", func() Projection { return newPolarProjection() })
	mustRegisterProjection("radar", func() Projection { return newRadarProjection() })
	mustRegisterProjection("mollweide", func() Projection { return newMollweideProjection() })
	mustRegisterProjection("hammer", func() Projection { return newHammerProjection() })
	mustRegisterProjection("aitoff", func() Projection { return newAitoffProjection() })
	mustRegisterProjection("lambert", func() Projection { return newLambertProjection() })
	mustRegisterProjection("skewx", func() Projection { return newSkewXProjection() })
	mustRegisterProjection("3d", func() Projection { return newAxes3DProjection() })
	mustRegisterProjection("axes3d", func() Projection { return newAxes3DProjection() })
}

// RegisterProjection installs a named axes projection.
func RegisterProjection(name string, factory ProjectionFactory) error {
	key := normalizeProjectionName(name)
	if key == "" {
		return fmt.Errorf("projection name must not be empty")
	}
	if factory == nil {
		return fmt.Errorf("projection %q has nil factory", key)
	}

	projectionRegistryMu.Lock()
	defer projectionRegistryMu.Unlock()

	if _, exists := projectionRegistry[key]; exists {
		return fmt.Errorf("projection %q already registered", key)
	}
	projectionRegistry[key] = factory
	return nil
}

func mustRegisterProjection(name string, factory ProjectionFactory) {
	if err := RegisterProjection(name, factory); err != nil {
		panic(err)
	}
}

func normalizeProjectionName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func lookupProjection(name string) (Projection, error) {
	key := normalizeProjectionName(name)
	if key == "" {
		key = "rectilinear"
	}

	projectionRegistryMu.RLock()
	factory := projectionRegistry[key]
	projectionRegistryMu.RUnlock()

	if factory == nil {
		return nil, fmt.Errorf("unknown projection %q", name)
	}
	return factory(), nil
}

func cloneProjection(proj Projection) Projection {
	if proj == nil {
		clone, _ := lookupProjection("rectilinear")
		return clone
	}
	if cloner, ok := proj.(ProjectionCloner); ok {
		return cloner.CloneProjection()
	}
	clone, err := lookupProjection(proj.Name())
	if err != nil {
		return proj
	}
	return clone
}

func isPolarProjection(proj Projection) bool {
	_, ok := polarProjectionFor(proj)
	return ok
}

type projectionFrameProvider interface {
	FramePath(clip geom.Rect) geom.Path
	ContainsDisplayPoint(clip geom.Rect, p geom.Pt) bool
}

func projectionFramePath(proj Projection, clip geom.Rect) (geom.Path, bool) {
	if provider, ok := proj.(projectionFrameProvider); ok {
		return provider.FramePath(clip), true
	}
	if isPolarProjection(proj) {
		return polarProjectionFramePath(proj, clip), true
	}
	return geom.Path{}, false
}

func projectionContainsDisplayPoint(proj Projection, clip geom.Rect, p geom.Pt) bool {
	if provider, ok := proj.(projectionFrameProvider); ok {
		return provider.ContainsDisplayPoint(clip, p)
	}
	if isPolarProjection(proj) {
		return polarProjectionContainsDisplayPoint(proj, clip, p)
	}
	return clip.Contains(p)
}

type rectilinearProjection struct{}

func (rectilinearProjection) Name() string { return "rectilinear" }

func (rectilinearProjection) ConfigureAxes(*Axes) {}

func (rectilinearProjection) DataToAxes(ax *Axes) transform.T {
	if ax == nil {
		return nil
	}
	return transform.NewScaleTransform(ax.effectiveXScale(), ax.effectiveYScale())
}

type polarProjection struct {
	name             string
	thetaOffset      float64
	thetaDirection   float64
	radialLabelAngle float64
	radarVariables   int
	radarLabels      []string
}

func newPolarProjection() *polarProjection {
	return &polarProjection{
		name:             "polar",
		thetaDirection:   1,
		radialLabelAngle: defaultPolarRadialLabelAngle,
	}
}

func newRadarProjection() *polarProjection {
	return &polarProjection{
		name:             "radar",
		thetaOffset:      math.Pi / 2,
		thetaDirection:   1,
		radialLabelAngle: math.Pi/2 + defaultPolarRadialLabelAngle,
		radarVariables:   defaultRadarVariables,
	}
}

func (p *polarProjection) Name() string {
	if p == nil || p.name == "" {
		return "polar"
	}
	return p.name
}

func (p *polarProjection) CloneProjection() Projection {
	if p == nil {
		return newPolarProjection()
	}
	clone := *p
	clone.radarLabels = append([]string(nil), p.radarLabels...)
	return &clone
}

func (p *polarProjection) ConfigureAxes(ax *Axes) {
	if ax == nil {
		return
	}

	ax.XScale = transform.NewLinear(0, 2*math.Pi)
	ax.YScale = transform.NewLinear(0, 1)
	_ = ax.SetBoxAspect(1)
	ax.XAxis = NewXAxis()
	ax.YAxis = NewYAxis()
	ax.XAxisTop = nil
	ax.YAxisRight = nil
	ax.ShowFrame = false

	ax.XAxis.Locator = MultipleLocator{Base: math.Pi / 4}
	ax.XAxis.MinorLocator = nil
	ax.XAxis.Formatter = FuncFormatter(formatPolarThetaLabel)
	ax.XAxis.ShowSpine = true
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = true

	ax.YAxis.Locator = LinearLocator{}
	ax.YAxis.MinorLocator = nil
	ax.YAxis.Formatter = ScalarFormatter{Prec: 3}
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = true

	if p.isRadar() {
		configureRadarThetaAxis(ax, p)
		ax.YAxis.ShowTicks = false
	}
}

func (p *polarProjection) DataToAxes(ax *Axes) transform.T {
	if ax == nil {
		return nil
	}
	return polarDataTransform{
		theta:          ax.effectiveXScale(),
		r:              ax.effectiveYScale(),
		thetaOffset:    p.thetaOffset,
		thetaDirection: p.thetaDirection,
	}
}

func formatPolarThetaLabel(theta float64) string {
	deg := math.Mod(theta*180/math.Pi, 360)
	if deg < 0 {
		deg += 360
	}
	if approx(deg, math.Round(deg), 1e-9) {
		return fmt.Sprintf("%.0f°", math.Round(deg))
	}
	return fmt.Sprintf("%.1f°", deg)
}

const defaultRadarVariables = 5

func (p *polarProjection) isRadar() bool {
	return p != nil && normalizeProjectionName(p.Name()) == "radar"
}

func (p *polarProjection) radarVariableCount() int {
	if p == nil {
		return defaultRadarVariables
	}
	if len(p.radarLabels) >= 3 {
		return len(p.radarLabels)
	}
	if p.radarVariables >= 3 {
		return p.radarVariables
	}
	return defaultRadarVariables
}

func configureRadarThetaAxis(ax *Axes, p *polarProjection) {
	if ax == nil || p == nil {
		return
	}
	count := p.radarVariableCount()
	ax.XAxis.Locator = FixedLocator{TicksList: RadarAngles(count)}
	ax.XAxis.MinorLocator = nil
	ax.XAxis.Formatter = radarThetaFormatter(p.radarLabels, count)
}

func radarThetaFormatter(labels []string, count int) Formatter {
	copied := append([]string(nil), labels...)
	if count < 3 {
		count = len(copied)
	}
	if count < 3 {
		count = defaultRadarVariables
	}
	return FuncFormatter(func(theta float64) string {
		fraction := normalizePolarFraction(theta / (2 * math.Pi))
		idx := int(math.Round(fraction*float64(count))) % count
		if idx >= 0 && idx < len(copied) && copied[idx] != "" {
			return copied[idx]
		}
		return fmt.Sprintf("%d", idx+1)
	})
}

// RadarAngles returns evenly spaced theta coordinates for a radar projection.
func RadarAngles(n int) []float64 {
	if n <= 0 {
		return nil
	}
	angles := make([]float64, n)
	for i := range angles {
		angles[i] = 2 * math.Pi * float64(i) / float64(n)
	}
	return angles
}
