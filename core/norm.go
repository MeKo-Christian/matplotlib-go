package core

import (
	"fmt"
	"math"
	"sort"
)

// ScalarNormalizer maps data values into the scalar-mappable colormap domain.
type ScalarNormalizer interface {
	Map(value float64) float64
	Inverse(value float64) (float64, bool)
	Autoscale(values []float64) ScalarNormalizer
	Range() (float64, float64)
	Validate() error
	NormName() string
}

// Normalize linearly maps values from [VMin, VMax] into [0, 1].
type Normalize struct {
	VMin float64
	VMax float64
	Clip bool
}

func (n Normalize) Map(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return math.NaN()
	}
	vmin, vmax := n.VMin, n.VMax
	if !isFinite(vmin) || !isFinite(vmax) {
		return math.NaN()
	}
	if vmin == vmax {
		return 0
	}
	out := (value - vmin) / (vmax - vmin)
	if n.Clip {
		return clamp01(out)
	}
	return out
}

func (n Normalize) Inverse(value float64) (float64, bool) {
	if !isFinite(n.VMin) || !isFinite(n.VMax) {
		return 0, false
	}
	return n.VMin + value*(n.VMax-n.VMin), true
}

func (n Normalize) Autoscale(values []float64) ScalarNormalizer {
	vmin, vmax := n.VMin, n.VMax
	dataMin, dataMax := finiteRange(values)
	if !isFinite(vmin) {
		vmin = dataMin
	}
	if !isFinite(vmax) {
		vmax = dataMax
	}
	n.VMin, n.VMax = vmin, vmax
	return n
}

func (n Normalize) Range() (float64, float64) { return n.VMin, n.VMax }

func (n Normalize) Validate() error {
	if isFinite(n.VMin) && isFinite(n.VMax) && n.VMin > n.VMax {
		return fmt.Errorf("minvalue must be less than or equal to maxvalue")
	}
	return nil
}

func (n Normalize) NormName() string { return "linear" }

// NoNorm passes values through unchanged, matching Matplotlib's index-style norm.
type NoNorm struct{}

func (n NoNorm) Map(value float64) float64 { return value }

func (n NoNorm) Inverse(value float64) (float64, bool) { return value, true }

func (n NoNorm) Autoscale([]float64) ScalarNormalizer { return n }

func (n NoNorm) Range() (float64, float64) { return 0, 1 }

func (n NoNorm) Validate() error { return nil }

func (n NoNorm) NormName() string { return "none" }

// LogNorm maps positive values logarithmically into [0, 1].
type LogNorm struct {
	VMin float64
	VMax float64
	Clip bool
}

func (n LogNorm) Map(value float64) float64 {
	if value <= 0 || !isFinite(value) || n.VMin <= 0 || n.VMax <= 0 || !isFinite(n.VMin) || !isFinite(n.VMax) {
		return math.NaN()
	}
	if n.VMin == n.VMax {
		return 0
	}
	out := (math.Log(value) - math.Log(n.VMin)) / (math.Log(n.VMax) - math.Log(n.VMin))
	if n.Clip {
		return clamp01(out)
	}
	return out
}

func (n LogNorm) Inverse(value float64) (float64, bool) {
	if n.VMin <= 0 || n.VMax <= 0 || !isFinite(n.VMin) || !isFinite(n.VMax) {
		return 0, false
	}
	return math.Exp(math.Log(n.VMin) + value*(math.Log(n.VMax)-math.Log(n.VMin))), true
}

func (n LogNorm) Autoscale(values []float64) ScalarNormalizer {
	vmin, vmax := n.VMin, n.VMax
	dataMin := math.Inf(1)
	dataMax := math.Inf(-1)
	for _, value := range values {
		if !isFinite(value) || value <= 0 {
			continue
		}
		dataMin = minF(dataMin, value)
		dataMax = maxF(dataMax, value)
	}
	if math.IsInf(dataMin, 1) || math.IsInf(dataMax, -1) {
		dataMin, dataMax = 1, 10
	}
	if !isFinite(vmin) {
		vmin = dataMin
	}
	if !isFinite(vmax) {
		vmax = dataMax
	}
	n.VMin, n.VMax = vmin, vmax
	return n
}

func (n LogNorm) Range() (float64, float64) { return n.VMin, n.VMax }

func (n LogNorm) Validate() error {
	if isFinite(n.VMin) && n.VMin <= 0 {
		return fmt.Errorf("log norm vmin must be positive")
	}
	if isFinite(n.VMax) && n.VMax <= 0 {
		return fmt.Errorf("log norm vmax must be positive")
	}
	if isFinite(n.VMin) && isFinite(n.VMax) && n.VMin > n.VMax {
		return fmt.Errorf("minvalue must be less than or equal to maxvalue")
	}
	return nil
}

func (n LogNorm) NormName() string { return "log" }

// SymLogNorm applies a logarithmic transform outside a linear threshold.
type SymLogNorm struct {
	VMin      float64
	VMax      float64
	LinThresh float64
	LinScale  float64
	Base      float64
	Clip      bool
}

func (n SymLogNorm) Map(value float64) float64 {
	if !isFinite(value) || !isFinite(n.VMin) || !isFinite(n.VMax) {
		return math.NaN()
	}
	if n.VMin == n.VMax {
		return 0
	}
	out := Normalize{VMin: symlogTransform(n.VMin, n), VMax: symlogTransform(n.VMax, n), Clip: n.Clip}.Map(symlogTransform(value, n))
	if math.IsInf(out, 0) {
		return math.NaN()
	}
	return out
}

func (n SymLogNorm) Inverse(value float64) (float64, bool) {
	if !isFinite(n.VMin) || !isFinite(n.VMax) {
		return 0, false
	}
	tmin, tmax := symlogTransform(n.VMin, n), symlogTransform(n.VMax, n)
	return inverseSymlogTransform(tmin+value*(tmax-tmin), n), true
}

func (n SymLogNorm) Autoscale(values []float64) ScalarNormalizer {
	linear := Normalize{VMin: n.VMin, VMax: n.VMax}.Autoscale(values).(Normalize)
	n.VMin, n.VMax = linear.VMin, linear.VMax
	return n
}

func (n SymLogNorm) Range() (float64, float64) { return n.VMin, n.VMax }

func (n SymLogNorm) Validate() error {
	if n.LinThresh < 0 {
		return fmt.Errorf("symlog linthresh must be non-negative")
	}
	if isFinite(n.VMin) && isFinite(n.VMax) && n.VMin > n.VMax {
		return fmt.Errorf("minvalue must be less than or equal to maxvalue")
	}
	return nil
}

func (n SymLogNorm) NormName() string { return "symlog" }

func symlogTransform(value float64, n SymLogNorm) float64 {
	linthresh := n.LinThresh
	if linthresh <= 0 {
		linthresh = 1
	}
	linscale := n.LinScale
	if linscale == 0 {
		linscale = 1
	}
	base := n.Base
	if base <= 0 || base == 1 {
		base = 10
	}
	adj := linscale / (1 - 1/base)
	absValue := math.Abs(value)
	if absValue <= linthresh {
		return value * adj
	}
	return math.Copysign(linthresh*(adj+math.Log(absValue/linthresh)/math.Log(base)), value)
}

func inverseSymlogTransform(value float64, n SymLogNorm) float64 {
	linthresh := n.LinThresh
	if linthresh <= 0 {
		linthresh = 1
	}
	linscale := n.LinScale
	if linscale == 0 {
		linscale = 1
	}
	base := n.Base
	if base <= 0 || base == 1 {
		base = 10
	}
	adj := linscale / (1 - 1/base)
	absValue := math.Abs(value)
	if absValue <= linthresh*adj {
		return value / adj
	}
	return math.Copysign(linthresh*math.Pow(base, absValue/linthresh-adj), value)
}

// AsinhNorm applies a smooth inverse-hyperbolic-sine transform before normalization.
type AsinhNorm struct {
	LinearWidth float64
	VMin        float64
	VMax        float64
	Clip        bool
}

func (n AsinhNorm) Map(value float64) float64 {
	if !isFinite(value) || !isFinite(n.VMin) || !isFinite(n.VMax) {
		return math.NaN()
	}
	if n.VMin == n.VMax {
		return 0
	}
	width := asinhNormLinearWidth(n.LinearWidth)
	out := Normalize{
		VMin: asinhNormTransform(n.VMin, width),
		VMax: asinhNormTransform(n.VMax, width),
		Clip: n.Clip,
	}.Map(asinhNormTransform(value, width))
	if math.IsInf(out, 0) {
		return math.NaN()
	}
	return out
}

func (n AsinhNorm) Inverse(value float64) (float64, bool) {
	if !isFinite(n.VMin) || !isFinite(n.VMax) {
		return 0, false
	}
	width := asinhNormLinearWidth(n.LinearWidth)
	tmin := asinhNormTransform(n.VMin, width)
	tmax := asinhNormTransform(n.VMax, width)
	return asinhNormInverse(tmin+value*(tmax-tmin), width), true
}

func (n AsinhNorm) Autoscale(values []float64) ScalarNormalizer {
	linear := Normalize{VMin: n.VMin, VMax: n.VMax}.Autoscale(values).(Normalize)
	n.VMin, n.VMax = linear.VMin, linear.VMax
	return n
}

func (n AsinhNorm) Range() (float64, float64) { return n.VMin, n.VMax }

func (n AsinhNorm) Validate() error {
	if n.LinearWidth < 0 || math.IsInf(n.LinearWidth, 0) || math.IsNaN(n.LinearWidth) {
		return fmt.Errorf("asinh norm linear_width must be positive")
	}
	return Normalize{VMin: n.VMin, VMax: n.VMax}.Validate()
}

func (n AsinhNorm) NormName() string { return "asinh" }

func asinhNormLinearWidth(width float64) float64 {
	if width <= 0 || math.IsNaN(width) || math.IsInf(width, 0) {
		return 1
	}
	return width
}

func asinhNormTransform(value, width float64) float64 {
	return width * math.Asinh(value/width)
}

func asinhNormInverse(value, width float64) float64 {
	return width * math.Sinh(value/width)
}

// PowerNorm applies a power-law transform after linear normalization.
type PowerNorm struct {
	Gamma float64
	VMin  float64
	VMax  float64
	Clip  bool
}

func (n PowerNorm) Map(value float64) float64 {
	if !isFinite(value) {
		return math.NaN()
	}
	base := Normalize{VMin: n.VMin, VMax: n.VMax, Clip: n.Clip}.Map(value)
	if base <= 0 {
		return base
	}
	gamma := n.Gamma
	if gamma == 0 {
		gamma = 1
	}
	return math.Pow(base, gamma)
}

func (n PowerNorm) Inverse(value float64) (float64, bool) {
	gamma := n.Gamma
	if gamma == 0 {
		gamma = 1
	}
	if value > 0 {
		value = math.Pow(value, 1/gamma)
	}
	return Normalize{VMin: n.VMin, VMax: n.VMax}.Inverse(value)
}

func (n PowerNorm) Autoscale(values []float64) ScalarNormalizer {
	linear := Normalize{VMin: n.VMin, VMax: n.VMax}.Autoscale(values).(Normalize)
	n.VMin, n.VMax = linear.VMin, linear.VMax
	return n
}

func (n PowerNorm) Range() (float64, float64) { return n.VMin, n.VMax }

func (n PowerNorm) Validate() error {
	return Normalize{VMin: n.VMin, VMax: n.VMax}.Validate()
}

func (n PowerNorm) NormName() string { return "power" }

// TwoSlopeNorm maps VCenter to 0.5 with independent linear slopes on each side.
type TwoSlopeNorm struct {
	VMin    float64
	VCenter float64
	VMax    float64
}

func (n TwoSlopeNorm) Map(value float64) float64 {
	if !isFinite(value) || !isFinite(n.VMin) || !isFinite(n.VCenter) || !isFinite(n.VMax) {
		return math.NaN()
	}
	if value <= n.VCenter {
		return 0.5 * (value - n.VMin) / (n.VCenter - n.VMin)
	}
	return 0.5 + 0.5*(value-n.VCenter)/(n.VMax-n.VCenter)
}

func (n TwoSlopeNorm) Inverse(value float64) (float64, bool) {
	if value <= 0.5 {
		return n.VMin + (value/0.5)*(n.VCenter-n.VMin), true
	}
	return n.VCenter + ((value-0.5)/0.5)*(n.VMax-n.VCenter), true
}

func (n TwoSlopeNorm) Autoscale(values []float64) ScalarNormalizer {
	dataMin, dataMax := finiteRange(values)
	if !isFinite(n.VMin) {
		n.VMin = dataMin
	}
	if !isFinite(n.VMax) {
		n.VMax = dataMax
	}
	if n.VMin >= n.VCenter {
		n.VMin = n.VCenter - (n.VMax - n.VCenter)
	}
	if n.VMax <= n.VCenter {
		n.VMax = n.VCenter + (n.VCenter - n.VMin)
	}
	return n
}

func (n TwoSlopeNorm) Range() (float64, float64) { return n.VMin, n.VMax }

func (n TwoSlopeNorm) Validate() error {
	if isFinite(n.VMin) && isFinite(n.VCenter) && isFinite(n.VMax) && !(n.VMin <= n.VCenter && n.VCenter <= n.VMax) {
		return fmt.Errorf("vmin, vcenter, and vmax must be in ascending order")
	}
	return nil
}

func (n TwoSlopeNorm) NormName() string { return "two-slope" }

// CenteredNorm maps values symmetrically around VCenter.
type CenteredNorm struct {
	VCenter   float64
	HalfRange float64
	Clip      bool
}

func (n CenteredNorm) Map(value float64) float64 {
	vmin, vmax := n.Range()
	return Normalize{VMin: vmin, VMax: vmax, Clip: n.Clip}.Map(value)
}

func (n CenteredNorm) Inverse(value float64) (float64, bool) {
	vmin, vmax := n.Range()
	return Normalize{VMin: vmin, VMax: vmax}.Inverse(value)
}

func (n CenteredNorm) Autoscale(values []float64) ScalarNormalizer {
	if n.HalfRange > 0 {
		return n
	}
	halfRange := 0.0
	for _, value := range values {
		if !isFinite(value) {
			continue
		}
		halfRange = maxF(halfRange, math.Abs(value-n.VCenter))
	}
	if halfRange == 0 {
		halfRange = 1
	}
	n.HalfRange = halfRange
	return n
}

func (n CenteredNorm) Range() (float64, float64) {
	halfRange := math.Abs(n.HalfRange)
	if halfRange == 0 || math.IsNaN(halfRange) {
		halfRange = 1
	}
	return n.VCenter - halfRange, n.VCenter + halfRange
}

func (n CenteredNorm) Validate() error { return nil }

func (n CenteredNorm) NormName() string { return "centered" }

// BoundaryNorm maps intervals to discrete colormap indexes.
type BoundaryNorm struct {
	Boundaries []float64
	NColors    int
	Clip       bool
	Extend     string
}

func (n BoundaryNorm) Map(value float64) float64 {
	idx := n.Index(value)
	if idx < 0 {
		return -1
	}
	denom := maxInt(n.NColors-1, 1)
	return float64(idx) / float64(denom)
}

func (n BoundaryNorm) Index(value float64) int {
	if len(n.Boundaries) < 2 || n.NColors <= 0 || !isFinite(value) {
		return -1
	}
	vmin, vmax := n.Boundaries[0], n.Boundaries[len(n.Boundaries)-1]
	if n.Clip {
		if value < vmin {
			return 0
		}
		if value >= vmax {
			return n.NColors - 1
		}
	} else {
		if value < vmin {
			return -1
		}
		if value >= vmax {
			return n.NColors
		}
	}

	idx := sort.Search(len(n.Boundaries), func(i int) bool {
		return value < n.Boundaries[i]
	}) - 1 + boundaryNormOffset(n.Extend)
	regions := boundaryNormRegionCount(n)
	if n.NColors > regions {
		if regions == 1 {
			idx = (n.NColors - 1) / 2
		} else {
			idx = int(float64(n.NColors-1) / float64(regions-1) * float64(idx))
		}
	}
	return idx
}

func (n BoundaryNorm) Inverse(float64) (float64, bool) { return 0, false }

func (n BoundaryNorm) Autoscale([]float64) ScalarNormalizer { return n }

func (n BoundaryNorm) Range() (float64, float64) {
	if len(n.Boundaries) < 2 {
		return math.NaN(), math.NaN()
	}
	return n.Boundaries[0], n.Boundaries[len(n.Boundaries)-1]
}

func (n BoundaryNorm) Validate() error {
	if n.Clip && n.Extend != "" && n.Extend != "neither" {
		return fmt.Errorf("clip=true is not compatible with extend")
	}
	if len(n.Boundaries) < 2 {
		return fmt.Errorf("boundary norm requires at least 2 boundaries")
	}
	for i := 1; i < len(n.Boundaries); i++ {
		if n.Boundaries[i] <= n.Boundaries[i-1] {
			return fmt.Errorf("boundary norm boundaries must be strictly increasing")
		}
	}
	if regions := boundaryNormRegionCount(n); n.NColors < regions {
		return fmt.Errorf("ncolors must equal or exceed the number of bins")
	}
	return nil
}

func (n BoundaryNorm) NormName() string { return "boundary" }

func boundaryNormRegionCount(n BoundaryNorm) int {
	regions := len(n.Boundaries) - 1
	switch n.Extend {
	case "min":
		regions++
	case "max":
		regions++
	case "both":
		regions += 2
	}
	return regions
}

func boundaryNormOffset(extend string) int {
	switch extend {
	case "min", "both":
		return 1
	default:
		return 0
	}
}
