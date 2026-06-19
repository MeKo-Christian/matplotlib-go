package core

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// BinStrategy specifies how to automatically determine histogram bin count.
type BinStrategy uint8

const (
	BinStrategyAuto    BinStrategy = iota // Sturges for n<1000, Scott otherwise
	BinStrategySturges                    // ceil(log2(n)) + 1
	BinStrategyScott                      // 3.49 * std * n^(-1/3)
	BinStrategyFD                         // 2 * IQR * n^(-1/3) (Freedman-Diaconis)
	BinStrategySqrt                       // ceil(sqrt(n))
)

// HistNorm specifies how to normalize histogram bin heights.
type HistNorm uint8

const (
	HistNormCount       HistNorm = iota // raw counts (default)
	HistNormProbability                 // count/total — each bar is fraction of total
	HistNormDensity                     // count/(total*width) — area integrates to 1
)

// HistRange defines an explicit histogram domain. Values outside the closed
// range are excluded from the bin counts.
type HistRange struct {
	Min float64
	Max float64
}

// HistType controls how histogram bins are presented.
type HistType uint8

const (
	HistTypeBar HistType = iota
	HistTypeStep
	HistTypeStepFilled
)

// Hist2D renders histogram plots computed from raw data.
// Bars span from edge[i] to edge[i+1] with no gap between adjacent bins.
type Hist2D struct {
	Data              []float64    // raw data values
	Weights           []float64    // per-sample weights; nil means each sample has weight 1
	Bins              int          // number of bins (0 = auto)
	BinEdges          []float64    // explicit bin edges; overrides Bins when len > 1
	Range             *HistRange   // explicit histogram range; ignored when BinEdges is set
	BinStrat          BinStrategy  // automatic binning strategy (used when Bins==0 and BinEdges==nil)
	Norm              HistNorm     // normalization mode
	Cumulative        bool         // accumulate bin heights from left to right
	ReverseCumulative bool         // accumulate bin heights from right to left
	HistType          HistType     // bar, step, or filled step presentation
	Baselines         []float64    // optional per-bin baselines for stacked histograms
	Color             render.Color // bar fill color
	EdgeColor         render.Color // bar outline color
	EdgeWidth         float64      // bar outline width in pixels (0 = no outline)
	Alpha             float64      // alpha transparency (0-1, 0 means 1.0)
	Label             string       // series label for legend
	z                 float64      // z-order

	// Computed lazily on first Draw/Bounds call.
	computed bool
	counts   []float64 // normalized heights per bin
	edges    []float64 // bin edge values (len = len(counts)+1)
}

// compute calculates bin edges and counts from Data.
func (h *Hist2D) compute() {
	if h.computed {
		return
	}
	h.computed = true

	if len(h.Data) == 0 {
		return
	}
	if len(h.Weights) > 0 && len(h.Weights) != len(h.Data) {
		return
	}

	// Determine bin edges.
	if len(h.BinEdges) > 1 {
		h.edges = h.BinEdges
	} else {
		edges, ok := computeHistBinEdges(h.Data, h.Bins, h.BinStrat, h.Range)
		if !ok {
			return
		}
		h.edges = edges
	}

	nBins := len(h.edges) - 1
	if nBins <= 0 {
		return
	}

	// Count data in each bin. The last bin includes the right edge.
	raw := make([]float64, nBins)
	total := 0.0
	for i, v := range h.Data {
		if !isFinite(v) {
			continue
		}
		idx := findBin(v, h.edges)
		if idx >= 0 && idx < nBins {
			weight := 1.0
			if len(h.Weights) > 0 {
				weight = h.Weights[i]
				if !isFinite(weight) {
					continue
				}
			}
			raw[idx] += weight
			total += weight
		}
	}
	if h.Cumulative {
		if h.ReverseCumulative {
			running := 0.0
			for i := len(raw) - 1; i >= 0; i-- {
				running += raw[i]
				raw[i] = running
			}
		} else {
			running := 0.0
			for i, c := range raw {
				running += c
				raw[i] = running
			}
		}
	}

	// Apply normalization.
	h.counts = make([]float64, nBins)
	for i, c := range raw {
		switch h.Norm {
		case HistNormProbability:
			if total != 0 {
				h.counts[i] = c / total
			}
		case HistNormDensity:
			if total == 0 {
				continue
			}
			if h.Cumulative {
				h.counts[i] = c / total
				continue
			}
			width := h.edges[i+1] - h.edges[i]
			if width > 0 {
				h.counts[i] = c / (total * width)
			}
		default: // HistNormCount
			h.counts[i] = c
		}
	}
}

// findBin returns the bin index for value v using the given edges.
// The last bin is closed on both sides [edges[n-1], edges[n]].
func findBin(v float64, edges []float64) int {
	n := len(edges) - 1
	if v < edges[0] || v > edges[n] {
		return -1
	}
	if v == edges[n] {
		return n - 1 // include right edge in last bin
	}
	// Binary search.
	lo, hi := 0, n-1
	for lo < hi {
		mid := (lo + hi) / 2
		if v < edges[mid+1] {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

// computeBinEdges computes evenly-spaced bin edges from data.
func computeBinEdges(data []float64, nBins int, start BinStrategy) []float64 {
	edges, _ := computeHistBinEdges(data, nBins, start, nil)
	return edges
}

func computeHistBinEdges(data []float64, nBins int, start BinStrategy, histRange *HistRange) ([]float64, bool) {
	if histRange != nil {
		if !isFinite(histRange.Min) || !isFinite(histRange.Max) || histRange.Min >= histRange.Max {
			return nil, false
		}
		if nBins <= 0 {
			filtered := finiteValuesInRange(data, histRange.Min, histRange.Max)
			if len(filtered) == 0 {
				nBins = 1
			} else {
				nBins = autoBinCount(filtered, start)
			}
		}
		if nBins < 1 {
			nBins = 1
		}
		return evenBinEdges(histRange.Min, histRange.Max, nBins), true
	}

	data = finiteValues(data)
	if len(data) == 0 {
		return nil, false
	}

	// Find data range.
	minV, maxV := data[0], data[0]
	for _, v := range data[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}

	// If all values identical, create a single bin around that value.
	if minV == maxV {
		half := math.Max(math.Abs(minV)*0.5, 0.5)
		minV -= half
		maxV += half
		nBins = 1
	}

	if nBins <= 0 {
		nBins = autoBinCount(data, start)
	}
	if nBins < 1 {
		nBins = 1
	}

	return evenBinEdges(minV, maxV, nBins), true
}

func evenBinEdges(minV, maxV float64, nBins int) []float64 {
	edges := make([]float64, nBins+1)
	width := (maxV - minV) / float64(nBins)
	for i := range edges {
		edges[i] = minV + float64(i)*width
	}
	// Ensure last edge is exactly maxV to avoid floating-point drift.
	edges[nBins] = maxV
	return edges
}

func finiteValuesInRange(data []float64, minV, maxV float64) []float64 {
	out := make([]float64, 0, len(data))
	for _, v := range data {
		if isFinite(v) && v >= minV && v <= maxV {
			out = append(out, v)
		}
	}
	return out
}

// autoBinCount chooses bin count based on strategy.
func autoBinCount(data []float64, start BinStrategy) int {
	n := len(data)
	if n == 0 {
		return 1
	}

	switch start {
	case BinStrategySturges:
		return sturgesBins(n)
	case BinStrategyScott:
		return scottBins(data)
	case BinStrategyFD:
		return fdBins(data)
	case BinStrategySqrt:
		return int(math.Ceil(math.Sqrt(float64(n))))
	default: // BinStrategyAuto
		if n < 1000 {
			return sturgesBins(n)
		}
		return scottBins(data)
	}
}

func sturgesBins(n int) int {
	if n <= 1 {
		return 1
	}
	return int(math.Ceil(math.Log2(float64(n)))) + 1
}

func scottBins(data []float64) int {
	n := len(data)
	sigma := stddev(data)
	if sigma == 0 {
		return sturgesBins(n)
	}
	minV, maxV := data[0], data[0]
	for _, v := range data[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	h := 3.49 * sigma * math.Pow(float64(n), -1.0/3.0)
	if h <= 0 {
		return sturgesBins(n)
	}
	k := int(math.Ceil((maxV - minV) / h))
	if k < 1 {
		return 1
	}
	return k
}

func fdBins(data []float64) int {
	n := len(data)
	iqr := computeIQR(data)
	if iqr == 0 {
		return sturgesBins(n)
	}
	minV, maxV := data[0], data[0]
	for _, v := range data[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	h := 2.0 * iqr * math.Pow(float64(n), -1.0/3.0)
	if h <= 0 {
		return sturgesBins(n)
	}
	k := int(math.Ceil((maxV - minV) / h))
	if k < 1 {
		return 1
	}
	return k
}

func stddev(data []float64) float64 {
	if len(data) < 2 {
		return 0
	}
	mean := 0.0
	for _, v := range data {
		mean += v
	}
	mean /= float64(len(data))
	variance := 0.0
	for _, v := range data {
		d := v - mean
		variance += d * d
	}
	variance /= float64(len(data) - 1)
	return math.Sqrt(variance)
}

func computeIQR(data []float64) float64 {
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)
	n := len(sorted)
	q1 := sorted[n/4]
	q3 := sorted[3*n/4]
	return q3 - q1
}

// Draw renders the histogram bars.
func (h *Hist2D) Draw(r render.Renderer, ctx *DrawContext) {
	h.compute()
	if len(h.counts) == 0 {
		return
	}

	alpha := h.Alpha
	if alpha <= 0 {
		alpha = 1.0
	}
	if alpha > 1 {
		alpha = 1.0
	}

	fillColor := h.Color
	if h.Alpha > 0 && h.Alpha <= 1 {
		fillColor.A = alpha
	}

	edgeColor := h.EdgeColor
	if h.Alpha > 0 && h.Alpha <= 1 {
		edgeColor.A = alpha
	}
	if edgeColor.A == 0 && h.HistType != HistTypeBar {
		edgeColor = fillColor
	}

	if h.HistType == HistTypeStep || h.HistType == HistTypeStepFilled {
		h.drawStepHistogram(r, ctx, fillColor, edgeColor)
		return
	}

	for i, count := range h.counts {
		left := h.edges[i]
		right := h.edges[i+1]
		baseline := h.baselineAt(i)

		px0 := ctx.DataToPixel.Apply(geom.Pt{X: left, Y: baseline})
		px1 := ctx.DataToPixel.Apply(geom.Pt{X: right, Y: baseline + count})
		rect, ok := rectFromPoints(px0, px1)
		if !ok {
			continue
		}

		path := pixelRectPath(rect)
		if len(path.C) == 0 {
			continue
		}
		paint := render.Paint{
			Snap: render.SnapAuto,
		}
		if h.EdgeWidth > 0 && edgeColor.A > 0 {
			paint.Stroke = edgeColor
			paint.LineWidth = h.EdgeWidth
			paint.LineJoin = render.JoinMiter
			paint.LineCap = render.CapButt
		}
		if fillColor.A > 0 {
			paint.Fill = fillColor
		}
		if paint.Fill.A == 0 && paint.Stroke.A == 0 {
			continue
		}
		r.Path(path, &paint)
	}
}

// Z returns the z-order for sorting.
func (h *Hist2D) Z() float64 {
	return zOrDefault(h.z, defaultPatchZ)
}

// Bounds returns the bounding box of the histogram in data coordinates.
func (h *Hist2D) Bounds(*DrawContext) geom.Rect {
	h.compute()
	if len(h.counts) == 0 || len(h.edges) < 2 {
		return geom.Rect{}
	}

	minY := h.baselineAt(0)
	maxY := minY + h.counts[0]
	if maxY < minY {
		minY, maxY = maxY, minY
	}
	for i, c := range h.counts {
		baseline := h.baselineAt(i)
		top := baseline + c
		if baseline < minY {
			minY = baseline
		}
		if baseline > maxY {
			maxY = baseline
		}
		if top < minY {
			minY = top
		}
		if top > maxY {
			maxY = top
		}
	}

	return geom.Rect{
		Min: geom.Pt{X: h.edges[0], Y: minY},
		Max: geom.Pt{X: h.edges[len(h.edges)-1], Y: maxY},
	}
}

// BinCounts returns the computed bin edges and counts.
// Useful for inspecting histogram results without drawing.
func (h *Hist2D) BinCounts() (edges, counts []float64) {
	h.compute()
	return h.edges, h.counts
}

func (h *Hist2D) baselineAt(i int) float64 {
	if i >= 0 && i < len(h.Baselines) {
		return h.Baselines[i]
	}
	return 0
}

func (h *Hist2D) drawStepHistogram(r render.Renderer, ctx *DrawContext, fillColor, edgeColor render.Color) {
	n := len(h.counts)
	if n == 0 {
		return
	}

	path := h.stepHistogramPath(ctx)
	if len(path.C) == 0 {
		return
	}

	paint := render.Paint{}
	paint.Snap = render.SnapAuto
	if h.HistType == HistTypeStepFilled && fillColor.A > 0 {
		paint.Fill = fillColor
	}
	if h.EdgeWidth > 0 && edgeColor.A > 0 {
		paint.Stroke = edgeColor
		paint.LineWidth = h.EdgeWidth
		paint.LineJoin = render.JoinMiter
		paint.LineCap = render.CapButt
	}
	if paint.Fill.A == 0 && paint.Stroke.A == 0 {
		return
	}
	r.Path(path, &paint)
}

func (h *Hist2D) stepHistogramPath(ctx *DrawContext) geom.Path {
	n := len(h.counts)
	if n == 0 || len(h.edges) < n+1 {
		return geom.Path{}
	}

	path := geom.Path{}
	firstBase := h.baselineAt(0)
	path.MoveTo(ctx.DataToPixel.Apply(geom.Pt{X: h.edges[0], Y: firstBase}))
	path.LineTo(ctx.DataToPixel.Apply(geom.Pt{X: h.edges[0], Y: firstBase + h.counts[0]}))
	for i := 0; i < n; i++ {
		top := h.baselineAt(i) + h.counts[i]
		path.LineTo(ctx.DataToPixel.Apply(geom.Pt{X: h.edges[i+1], Y: top}))
		if i+1 < n {
			nextTop := h.baselineAt(i+1) + h.counts[i+1]
			path.LineTo(ctx.DataToPixel.Apply(geom.Pt{X: h.edges[i+1], Y: nextTop}))
		}
	}
	if h.HistType != HistTypeStepFilled {
		path.LineTo(ctx.DataToPixel.Apply(geom.Pt{X: h.edges[n], Y: h.baselineAt(n - 1)}))
		return path
	}

	for i := n - 1; i >= 0; i-- {
		base := h.baselineAt(i)
		path.LineTo(ctx.DataToPixel.Apply(geom.Pt{X: h.edges[i+1], Y: base}))
		path.LineTo(ctx.DataToPixel.Apply(geom.Pt{X: h.edges[i], Y: base}))
	}
	path.Close()
	return path
}
