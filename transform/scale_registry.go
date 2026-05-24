package transform

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

// NonPositiveMode controls how log-like scales treat values outside their valid domain.
type NonPositiveMode string

const (
	NonPositiveMask NonPositiveMode = "mask"
	NonPositiveClip NonPositiveMode = "clip"
)

// ScaleOptions configures named scale construction.
type ScaleOptions struct {
	DomainMin   float64
	DomainMax   float64
	Base        float64
	Subs        []float64
	LinThresh   float64
	LinearScale float64
	LinearWidth float64
	ClipEpsilon float64
	NonPositive NonPositiveMode
	Forward     func(float64) float64
	Inverse     func(float64) (float64, bool)
}

// ScaleOption mutates ScaleOptions for named scale creation.
type ScaleOption func(*ScaleOptions)

// DefaultScaleOptions returns the default options used by the scale registry.
func DefaultScaleOptions() ScaleOptions {
	return ScaleOptions{
		DomainMin:   0,
		DomainMax:   1,
		Base:        10,
		LinThresh:   1,
		LinearScale: 1,
		LinearWidth: 1,
		ClipEpsilon: 1e-6,
		NonPositive: NonPositiveMask,
	}
}

// ResolveScaleOptions applies options onto the registry defaults.
func ResolveScaleOptions(opts ...ScaleOption) ScaleOptions {
	cfg := DefaultScaleOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func WithScaleDomain(minVal, maxVal float64) ScaleOption {
	return func(cfg *ScaleOptions) {
		cfg.DomainMin = minVal
		cfg.DomainMax = maxVal
	}
}

func WithScaleBase(base float64) ScaleOption {
	return func(cfg *ScaleOptions) {
		cfg.Base = base
	}
}

func WithScaleSubs(subs ...float64) ScaleOption {
	return func(cfg *ScaleOptions) {
		cfg.Subs = append([]float64(nil), subs...)
	}
}

func WithScaleLinThresh(v float64) ScaleOption {
	return func(cfg *ScaleOptions) {
		cfg.LinThresh = v
	}
}

func WithScaleLinearScale(v float64) ScaleOption {
	return func(cfg *ScaleOptions) {
		cfg.LinearScale = v
	}
}

func WithScaleLinearWidth(v float64) ScaleOption {
	return func(cfg *ScaleOptions) {
		cfg.LinearWidth = v
	}
}

func WithScaleClipEpsilon(v float64) ScaleOption {
	return func(cfg *ScaleOptions) {
		cfg.ClipEpsilon = v
	}
}

func WithScaleNonPositive(mode NonPositiveMode) ScaleOption {
	return func(cfg *ScaleOptions) {
		cfg.NonPositive = mode
	}
}

func WithScaleFunctions(forward func(float64) float64, inverse func(float64) (float64, bool)) ScaleOption {
	return func(cfg *ScaleOptions) {
		cfg.Forward = forward
		cfg.Inverse = inverse
	}
}

// ScaleFactory builds a scale from resolved options.
type ScaleFactory func(opts ScaleOptions) (Scale, error)

// ScaleRegistry manages named scale factories.
type ScaleRegistry struct {
	mu        sync.RWMutex
	factories map[string]ScaleFactory
}

func NewScaleRegistry() *ScaleRegistry {
	return &ScaleRegistry{
		factories: make(map[string]ScaleFactory),
	}
}

func (r *ScaleRegistry) Register(name string, factory ScaleFactory) error {
	if factory == nil {
		return errors.New("scale factory cannot be nil")
	}
	key := normalizeScaleName(name)
	if key == "" {
		return errors.New("scale name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.factories[key]; exists {
		return fmt.Errorf("scale %q already registered", key)
	}
	r.factories[key] = factory
	return nil
}

func (r *ScaleRegistry) MustRegister(name string, factory ScaleFactory) {
	if err := r.Register(name, factory); err != nil {
		panic(err)
	}
}

func (r *ScaleRegistry) New(name string, opts ...ScaleOption) (Scale, error) {
	return r.NewWithOptions(name, ResolveScaleOptions(opts...))
}

func (r *ScaleRegistry) NewWithOptions(name string, opts ScaleOptions) (Scale, error) {
	key := normalizeScaleName(name)

	r.mu.RLock()
	factory, ok := r.factories[key]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown scale %q", key)
	}
	return factory(opts)
}

func (r *ScaleRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

var DefaultScaleRegistry = func() *ScaleRegistry {
	r := NewScaleRegistry()
	registerBuiltInScales(r)
	return r
}()

func RegisterScale(name string, factory ScaleFactory) error {
	return DefaultScaleRegistry.Register(name, factory)
}

func MustRegisterScale(name string, factory ScaleFactory) {
	DefaultScaleRegistry.MustRegister(name, factory)
}

func NewScale(name string, opts ...ScaleOption) (Scale, error) {
	return DefaultScaleRegistry.New(name, opts...)
}

func NewScaleWithOptions(name string, opts ScaleOptions) (Scale, error) {
	return DefaultScaleRegistry.NewWithOptions(name, opts)
}

func ScaleNames() []string {
	return DefaultScaleRegistry.Names()
}

func normalizeScaleName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func registerBuiltInScales(r *ScaleRegistry) {
	r.MustRegister("linear", func(opts ScaleOptions) (Scale, error) {
		return NewLinear(opts.DomainMin, opts.DomainMax), nil
	})
	r.MustRegister("log", func(opts ScaleOptions) (Scale, error) {
		if opts.Base <= 1 {
			return nil, fmt.Errorf("log scale base must be > 1")
		}
		minVal, maxVal := normalizeLogDomain(opts.DomainMin, opts.DomainMax, opts.Base)
		return Log{
			Min:         minVal,
			Max:         maxVal,
			Base:        opts.Base,
			NonPositive: normalizeNonPositive(opts.NonPositive),
		}, nil
	})
	r.MustRegister("symlog", func(opts ScaleOptions) (Scale, error) {
		if opts.Base <= 1 {
			return nil, fmt.Errorf("symlog scale base must be > 1")
		}
		return NewSymLog(opts.DomainMin, opts.DomainMax, opts.Base, opts.LinThresh, opts.LinearScale), nil
	})
	r.MustRegister("asinh", func(opts ScaleOptions) (Scale, error) {
		return NewAsinh(opts.DomainMin, opts.DomainMax, opts.LinearWidth), nil
	})
	r.MustRegister("logit", func(opts ScaleOptions) (Scale, error) {
		minVal, maxVal := normalizeLogitDomain(opts.DomainMin, opts.DomainMax, opts.ClipEpsilon)
		return NewLogit(minVal, maxVal, normalizeNonPositive(opts.NonPositive), opts.ClipEpsilon), nil
	})
	functionFactory := func(opts ScaleOptions) (Scale, error) {
		if opts.Forward == nil || opts.Inverse == nil {
			return nil, errors.New("function scale requires both forward and inverse functions")
		}
		return NewFuncScale(opts.DomainMin, opts.DomainMax, opts.Forward, opts.Inverse), nil
	}
	r.MustRegister("function", functionFactory)
	r.MustRegister("func", functionFactory)
	r.MustRegister("functionlog", func(opts ScaleOptions) (Scale, error) {
		if opts.Forward == nil || opts.Inverse == nil {
			return nil, errors.New("functionlog scale requires both forward and inverse functions")
		}
		if opts.Base <= 1 {
			return nil, fmt.Errorf("functionlog scale base must be > 1")
		}
		return NewFuncLogScale(opts.DomainMin, opts.DomainMax, opts.Base, opts.Forward, opts.Inverse), nil
	})
}

// SymLog applies a signed logarithmic transform with a linear region around zero.
type SymLog struct {
	Min, Max    float64
	Base        float64
	LinThresh   float64
	LinearScale float64
}

func NewSymLog(minVal, maxVal, base, linThresh, linearScale float64) SymLog {
	if linThresh <= 0 {
		linThresh = 1
	}
	if linearScale <= 0 {
		linearScale = 1
	}
	return SymLog{
		Min:         minVal,
		Max:         maxVal,
		Base:        base,
		LinThresh:   linThresh,
		LinearScale: linearScale,
	}
}

func (s SymLog) Domain() (float64, float64) { return s.Min, s.Max }

func (s SymLog) WithDomain(minVal, maxVal float64) Scale {
	s.Min = minVal
	s.Max = maxVal
	return s
}

func (s SymLog) valid() bool {
	return s.Base > 1 && s.Min != s.Max && s.LinThresh > 0 && s.LinearScale > 0
}

func (s SymLog) Fwd(x float64) float64 {
	if !s.valid() {
		return 0
	}
	return normalizedMappedForward(s.Min, s.Max, s.transform, x)
}

func (s SymLog) Inv(u float64) (float64, bool) {
	if !s.valid() {
		return s.Min, false
	}
	return normalizedMappedInverse(s.Min, s.Max, s.transform, s.inverse, u)
}

func (s SymLog) transform(x float64) (float64, bool) {
	sign := 1.0
	if x < 0 {
		sign = -1
		x = -x
	}
	if x <= s.LinThresh {
		return sign * s.LinearScale * x / s.LinThresh, true
	}
	return sign * (s.LinearScale + math.Log(x/s.LinThresh)/math.Log(s.Base)), true
}

func (s SymLog) inverse(y float64) (float64, bool) {
	sign := 1.0
	if y < 0 {
		sign = -1
		y = -y
	}
	if y <= s.LinearScale {
		return sign * s.LinThresh * y / s.LinearScale, true
	}
	return sign * s.LinThresh * math.Pow(s.Base, y-s.LinearScale), true
}

// Asinh applies an inverse-hyperbolic-sine transform with a configurable linear width.
type Asinh struct {
	Min, Max    float64
	LinearWidth float64
}

func NewAsinh(minVal, maxVal, linearWidth float64) Asinh {
	if linearWidth <= 0 {
		linearWidth = 1
	}
	return Asinh{Min: minVal, Max: maxVal, LinearWidth: linearWidth}
}

func (s Asinh) Domain() (float64, float64) { return s.Min, s.Max }

func (s Asinh) WithDomain(minVal, maxVal float64) Scale {
	s.Min = minVal
	s.Max = maxVal
	return s
}

func (s Asinh) valid() bool {
	return s.Min != s.Max && s.LinearWidth > 0
}

func (s Asinh) Fwd(x float64) float64 {
	if !s.valid() {
		return 0
	}
	return normalizedMappedForward(s.Min, s.Max, s.transform, x)
}

func (s Asinh) Inv(u float64) (float64, bool) {
	if !s.valid() {
		return s.Min, false
	}
	return normalizedMappedInverse(s.Min, s.Max, s.transform, s.inverse, u)
}

func (s Asinh) transform(x float64) (float64, bool) {
	return math.Asinh(x / s.LinearWidth), true
}

func (s Asinh) inverse(y float64) (float64, bool) {
	return math.Sinh(y) * s.LinearWidth, true
}

// Logit applies log(x / (1 - x)) on probabilities in (0,1).
type Logit struct {
	Min, Max    float64
	NonPositive NonPositiveMode
	ClipEpsilon float64
}

func NewLogit(minVal, maxVal float64, nonPositive NonPositiveMode, clipEpsilon float64) Logit {
	if clipEpsilon <= 0 || clipEpsilon >= 0.5 {
		clipEpsilon = 1e-6
	}
	return Logit{
		Min:         minVal,
		Max:         maxVal,
		NonPositive: normalizeNonPositive(nonPositive),
		ClipEpsilon: clipEpsilon,
	}
}

func (s Logit) Domain() (float64, float64) { return s.Min, s.Max }

func (s Logit) WithDomain(min, max float64) Scale {
	s.Min, s.Max = normalizeLogitDomain(min, max, s.ClipEpsilon)
	return s
}

func (s Logit) valid() bool {
	return s.Min != s.Max && s.ClipEpsilon > 0 && s.ClipEpsilon < 0.5
}

func (s Logit) Fwd(x float64) float64 {
	if !s.valid() {
		return 0
	}
	return normalizedMappedForward(s.Min, s.Max, s.transform, x)
}

func (s Logit) Inv(u float64) (float64, bool) {
	if !s.valid() {
		return s.Min, false
	}
	return normalizedMappedInverse(s.Min, s.Max, s.transform, s.inverse, u)
}

func (s Logit) transform(x float64) (float64, bool) {
	x, ok := s.normalize(x)
	if !ok {
		return 0, false
	}
	return math.Log(x / (1 - x)), true
}

func (s Logit) inverse(y float64) (float64, bool) {
	x := 1 / (1 + math.Exp(-y))
	return s.normalize(x)
}

func (s Logit) normalize(x float64) (float64, bool) {
	if x > 0 && x < 1 {
		return x, true
	}
	if s.NonPositive != NonPositiveClip {
		return 0, false
	}
	if x <= 0 {
		return s.ClipEpsilon, true
	}
	return 1 - s.ClipEpsilon, true
}

// FuncScale applies caller-provided forward/inverse mapping functions.
type FuncScale struct {
	Min, Max float64
	Forward  func(float64) float64
	Inverse  func(float64) (float64, bool)
}

func NewFuncScale(minVal, maxVal float64, forward func(float64) float64, inverse func(float64) (float64, bool)) FuncScale {
	return FuncScale{Min: minVal, Max: maxVal, Forward: forward, Inverse: inverse}
}

func (s FuncScale) Domain() (float64, float64) { return s.Min, s.Max }

func (s FuncScale) WithDomain(min, max float64) Scale {
	s.Min = min
	s.Max = max
	return s
}

func (s FuncScale) valid() bool {
	return s.Min != s.Max && s.Forward != nil && s.Inverse != nil
}

func (s FuncScale) Fwd(x float64) float64 {
	if !s.valid() {
		return 0
	}
	return normalizedMappedForward(s.Min, s.Max, s.transform, x)
}

func (s FuncScale) Inv(u float64) (float64, bool) {
	if !s.valid() {
		return s.Min, false
	}
	return normalizedMappedInverse(s.Min, s.Max, s.transform, s.Inverse, u)
}

func (s FuncScale) transform(x float64) (float64, bool) {
	y := s.Forward(x)
	if !isFinite(y) {
		return 0, false
	}
	return y, true
}

// FuncLogScale applies caller-provided forward/inverse mappings on a log axis.
type FuncLogScale struct {
	Min, Max float64
	Base     float64
	Forward  func(float64) float64
	Inverse  func(float64) (float64, bool)
}

func NewFuncLogScale(minVal, maxVal, base float64, forward func(float64) float64, inverse func(float64) (float64, bool)) FuncLogScale {
	if base <= 1 {
		base = 10
	}
	return FuncLogScale{Min: minVal, Max: maxVal, Base: base, Forward: forward, Inverse: inverse}
}

func (s FuncLogScale) Domain() (float64, float64) { return s.Min, s.Max }

func (s FuncLogScale) WithDomain(min, max float64) Scale {
	s.Min = min
	s.Max = max
	return s
}

func (s FuncLogScale) valid() bool {
	return s.Min != s.Max && s.Base > 1 && s.Forward != nil && s.Inverse != nil
}

func (s FuncLogScale) Fwd(x float64) float64 {
	if !s.valid() {
		return 0
	}
	return normalizedMappedForward(s.Min, s.Max, s.transform, x)
}

func (s FuncLogScale) Inv(u float64) (float64, bool) {
	if !s.valid() {
		return s.Min, false
	}
	return normalizedMappedInverse(s.Min, s.Max, s.transform, s.inverse, u)
}

func (s FuncLogScale) transform(x float64) (float64, bool) {
	y := s.Forward(x)
	if !isFinite(y) || y <= 0 {
		return 0, false
	}
	return math.Log(y) / math.Log(s.Base), true
}

func (s FuncLogScale) inverse(y float64) (float64, bool) {
	x := math.Pow(s.Base, y)
	return s.Inverse(x)
}

func normalizedMappedForward(minVal, maxVal float64, transform func(float64) (float64, bool), x float64) float64 {
	lo, ok := transform(minVal)
	if !ok {
		return 0
	}
	hi, ok := transform(maxVal)
	if !ok {
		return 0
	}
	vx, ok := transform(x)
	if !ok {
		return math.NaN()
	}
	den := hi - lo
	if den == 0 || !isFinite(den) {
		return 0
	}
	return (vx - lo) / den
}

func normalizedMappedInverse(minVal, maxVal float64, transform, inverse func(float64) (float64, bool), u float64) (float64, bool) {
	lo, ok := transform(minVal)
	if !ok {
		return minVal, false
	}
	hi, ok := transform(maxVal)
	if !ok {
		return minVal, false
	}
	den := hi - lo
	if den == 0 || !isFinite(den) {
		return minVal, false
	}
	return inverse(lo + u*den)
}

func normalizeNonPositive(mode NonPositiveMode) NonPositiveMode {
	switch mode {
	case NonPositiveClip:
		return NonPositiveClip
	default:
		return NonPositiveMask
	}
}

func normalizeLogDomain(minVal, maxVal, base float64) (float64, float64) {
	if base <= 1 {
		base = 10
	}

	reversed := minVal > maxVal
	if reversed {
		minVal, maxVal = maxVal, minVal
	}

	switch {
	case maxVal <= 0:
		minVal, maxVal = 1, base
	case minVal <= 0:
		minVal = logClipFloor(minVal, maxVal, base)
	}

	if minVal == maxVal {
		if minVal <= 0 {
			minVal, maxVal = 1, base
		} else {
			minVal, maxVal = minVal/base, minVal*base
		}
	}

	if reversed {
		return maxVal, minVal
	}
	return minVal, maxVal
}

func logClipFloor(minVal, maxVal, base float64) float64 {
	if base <= 1 {
		base = 10
	}
	floor := 1 / (base * base)
	if minVal > 0 {
		floor = minVal
	} else if maxVal > 0 {
		floor = maxVal / (base * base)
	}
	if floor <= 0 || !isFinite(floor) {
		floor = 1 / (base * base)
	}
	return floor
}

func normalizeLogitDomain(minVal, maxVal, eps float64) (float64, float64) {
	if eps <= 0 || eps >= 0.5 {
		eps = 1e-6
	}

	reversed := minVal > maxVal
	if reversed {
		minVal, maxVal = maxVal, minVal
	}

	if minVal <= 0 {
		minVal = eps
	}
	if maxVal >= 1 {
		maxVal = 1 - eps
	}
	if minVal == maxVal {
		minVal, maxVal = eps, 1-eps
	}

	if reversed {
		return maxVal, minVal
	}
	return minVal, maxVal
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
