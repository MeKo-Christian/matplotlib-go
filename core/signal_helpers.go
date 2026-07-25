package core

import (
	"math"
	"math/cmplx"

	algofft "github.com/cwbudde/algo-fft"
	"github.com/cwbudde/matplotlib-go/internal/optarg"
	"github.com/cwbudde/matplotlib-go/ticker"
)

// SpecgramOptions configures Axes.Specgram.
type SpecgramOptions struct {
	Fs       float64
	NFFT     int
	NOverlap int
	PadTo    int
	Window   string
	DB       *bool
	Colormap *string
	VMin     *float64
	VMax     *float64
	Alpha    *float64
	Label    string
}

// SignalSpectrumOptions configures PSD/CSD/coherence helpers.
type SignalSpectrumOptions struct {
	Fs       float64
	Fc       float64
	NFFT     int
	NOverlap int
	PadTo    int
	Window   string
	Detrend  SignalDetrend
	Sides    SignalSpectrumSides
	Scale    SignalSpectrumScale
	PlotOptions
}

// SignalDetrend configures per-segment detrending before FFT calculation.
type SignalDetrend string

const (
	SignalDetrendNone   SignalDetrend = "none"
	SignalDetrendMean   SignalDetrend = "mean"
	SignalDetrendLinear SignalDetrend = "linear"
)

// SignalSpectrumSides configures whether FFT helpers return one or both sides
// of the spectrum.
type SignalSpectrumSides string

const (
	SignalSpectrumSidesDefault  SignalSpectrumSides = ""
	SignalSpectrumSidesOneSided SignalSpectrumSides = "onesided"
	SignalSpectrumSidesTwoSided SignalSpectrumSides = "twosided"
)

// SignalSpectrumScale configures magnitude-spectrum value scaling.
type SignalSpectrumScale string

const (
	SignalSpectrumScaleDefault SignalSpectrumScale = ""
	SignalSpectrumScaleLinear  SignalSpectrumScale = "linear"
	SignalSpectrumScaleDB      SignalSpectrumScale = "dB"
)

// CorrelationOptions configures Axes.XCorr and Axes.ACorr.
type CorrelationOptions struct {
	MaxLags   int
	Normalize *bool
	PlotOptions
}

// SpecgramResult stores the rendered spectrogram and computed bins.
type SpecgramResult struct {
	Image       *Image2D
	Spectrum    [][]float64
	Frequencies []float64
	Times       []float64
}

// SpectrumResult stores the line artist and computed frequency-domain data.
type SpectrumResult struct {
	Line        *Line2D
	Frequencies []float64
	Values      []float64
}

// CorrelationResult stores the line artist and computed lag-domain data.
type CorrelationResult struct {
	Line   *Line2D
	Lags   []float64
	Values []float64
}

// Specgram computes a simple spectrogram and renders it as an image.
func (a *Axes) Specgram(samples []float64, opts ...SpecgramOptions) *SpecgramResult {
	cfg := optarg.One("specgram", opts)

	samples = finiteSeries(samples)
	fs, nfft, noverlap, padTo, ok := resolveSignalParams(len(samples), cfg.Fs, cfg.NFFT, cfg.NOverlap, cfg.PadTo)
	if !ok {
		return nil
	}
	freqs, times, spectrum := computeSpectrogram(samples, signalFFTConfig{
		Fs:       fs,
		NFFT:     nfft,
		NOverlap: noverlap,
		PadTo:    padTo,
		Window:   cfg.Window,
	})
	if len(freqs) == 0 || len(times) == 0 || len(spectrum) == 0 {
		return nil
	}

	useDB := true
	if cfg.DB != nil {
		useDB = *cfg.DB
	}
	if useDB {
		spectrum = scaleSpectrumDB(spectrum)
	}

	xMin, xMax := centersExtent(times, float64(nfft-noverlap)/(2*fs))
	yMin, yMax := freqs[0], freqs[len(freqs)-1]
	img := a.Image(spectrum, ImageOptions{
		Colormap: cfg.Colormap,
		VMin:     cfg.VMin,
		VMax:     cfg.VMax,
		Alpha:    cfg.Alpha,
		XMin:     &xMin,
		XMax:     &xMax,
		YMin:     &yMin,
		YMax:     &yMax,
		Origin:   ImageOriginLower,
		Label:    cfg.Label,
	})
	if img == nil {
		return nil
	}
	a.SetXLim(xMin, xMax)
	a.SetYLim(yMin, yMax)
	return &SpecgramResult{
		Image:       img,
		Spectrum:    spectrum,
		Frequencies: freqs,
		Times:       times,
	}
}

// PSD computes a Welch power spectral density estimate and plots it.
func (a *Axes) PSD(samples []float64, opts ...SignalSpectrumOptions) *SpectrumResult {
	cfg := optarg.One("psd", opts)
	samples = finiteSeries(samples)
	freqs, psd := computePSD(samples, cfg)
	offsetFrequencies(freqs, cfg.Fc)
	a.SetXLabel("Frequency")
	a.SetYLabel("Power Spectral Density (dB/Hz)")
	a.AddXGrid()
	a.AddYGrid()
	return plotPowerSpectrumResult(a, freqs, psd, cfg.PlotOptions)
}

// MagnitudeSpectrum computes a one-sided FFT magnitude spectrum and plots it.
func (a *Axes) MagnitudeSpectrum(samples []float64, opts ...SignalSpectrumOptions) *SpectrumResult {
	cfg := optarg.One("magnitude spectrum", opts)
	samples = finiteSeries(samples)
	freqs, values := computeMagnitudeSpectrum(samples, cfg)
	a.SetXLabel("Frequency")
	plotValues := values
	if cfg.Scale == SignalSpectrumScaleDB {
		a.SetYLabel("Magnitude (dB)")
		plotValues = magnitudeSpectrumDBValues(values)
	} else {
		a.SetYLabel("Magnitude (energy)")
	}
	result := plotSpectrumResult(a, freqs, plotValues, cfg.PlotOptions)
	if result == nil {
		return nil
	}
	result.Values = append([]float64(nil), values...)
	return result
}

// AngleSpectrum computes a one-sided FFT phase angle spectrum in radians.
func (a *Axes) AngleSpectrum(samples []float64, opts ...SignalSpectrumOptions) *SpectrumResult {
	cfg := optarg.One("angle spectrum", opts)
	samples = finiteSeries(samples)
	freqs, values := computeAngleSpectrum(samples, cfg)
	a.SetXLabel("Frequency")
	a.SetYLabel("Angle (radians)")
	return plotSpectrumResult(a, freqs, values, cfg.PlotOptions)
}

// PhaseSpectrum computes a one-sided unwrapped FFT phase spectrum in radians.
func (a *Axes) PhaseSpectrum(samples []float64, opts ...SignalSpectrumOptions) *SpectrumResult {
	cfg := optarg.One("phase spectrum", opts)
	samples = finiteSeries(samples)
	freqs, values := computePhaseSpectrum(samples, cfg)
	a.SetXLabel("Frequency")
	a.SetYLabel("Phase (radians)")
	return plotSpectrumResult(a, freqs, values, cfg.PlotOptions)
}

// CSD computes the magnitude of the cross spectral density estimate and plots it.
func (a *Axes) CSD(x, y []float64, opts ...SignalSpectrumOptions) *SpectrumResult {
	cfg := optarg.One("csd", opts)
	x, y = finitePairs(x, y)
	freqs, values := computeCSDMagnitude(x, y, cfg)
	offsetFrequencies(freqs, cfg.Fc)
	a.SetXLabel("Frequency")
	a.SetYLabel("Cross Spectrum Magnitude (dB)")
	a.AddXGrid()
	a.AddYGrid()
	return plotPowerSpectrumResult(a, freqs, values, cfg.PlotOptions)
}

// Cohere computes magnitude-squared coherence and plots it.
func (a *Axes) Cohere(x, y []float64, opts ...SignalSpectrumOptions) *SpectrumResult {
	cfg := optarg.One("cohere", opts)
	x, y = finitePairs(x, y)
	freqs, values := computeCoherence(x, y, cfg)
	offsetFrequencies(freqs, cfg.Fc)
	a.SetXLabel("Frequency")
	a.SetYLabel("Coherence")
	a.AddXGrid()
	a.AddYGrid()
	return plotSpectrumResult(a, freqs, values, cfg.PlotOptions)
}

// XCorr computes the cross-correlation sequence and plots it.
func (a *Axes) XCorr(x, y []float64, opts ...CorrelationOptions) *CorrelationResult {
	cfg := optarg.One("xcorr", opts)
	x, y = finitePairs(x, y)
	lags, values := computeCorrelation(x, y, cfg)
	return plotCorrelationResult(a, lags, values, cfg.PlotOptions)
}

// ACorr computes the auto-correlation sequence and plots it.
func (a *Axes) ACorr(x []float64, opts ...CorrelationOptions) *CorrelationResult {
	cfg := optarg.One("acorr", opts)
	x = finiteSeries(x)
	lags, values := computeCorrelation(x, x, cfg)
	return plotCorrelationResult(a, lags, values, cfg.PlotOptions)
}

type signalFFTConfig struct {
	Fs       float64
	NFFT     int
	NOverlap int
	PadTo    int
	Window   string
}

func resolveSignalParams(length int, fs float64, nfft, noverlap, padTo int) (float64, int, int, int, bool) {
	if length < 2 {
		return 0, 0, 0, 0, false
	}
	if fs <= 0 {
		fs = 1
	}
	if nfft <= 0 {
		nfft = min(256, length)
	}
	if nfft < 2 {
		return 0, 0, 0, 0, false
	}
	if nfft > length {
		nfft = length
	}
	if noverlap < 0 {
		noverlap = 0
	}
	if noverlap >= nfft {
		noverlap = nfft / 2
		if noverlap >= nfft {
			noverlap = nfft - 1
		}
	}
	if noverlap == 0 {
		noverlap = nfft / 2
	}
	if padTo < nfft {
		padTo = nfft
	}
	return fs, nfft, noverlap, padTo, true
}

func computeSpectrogram(samples []float64, cfg signalFFTConfig) ([]float64, []float64, [][]float64) {
	segments, starts := windowedSegments(samples, cfg.NFFT, cfg.NOverlap)
	if len(segments) == 0 {
		return nil, nil, nil
	}
	window := signalWindow(cfg.Window, cfg.NFFT)
	freqs := fftFrequencies(cfg.Fs, cfg.PadTo)
	spectrum := make([][]float64, len(freqs))
	for i := range spectrum {
		spectrum[i] = make([]float64, len(segments))
	}
	times := make([]float64, len(segments))
	scale := windowPower(window) * cfg.Fs
	if scale == 0 {
		scale = 1
	}

	for col, segment := range segments {
		bins := oneSidedDFTPower(applyWindow(segment, window), cfg.PadTo)
		scaleOneSidedPSD(bins, cfg.NFFT)
		for row := range bins {
			spectrum[row][col] = bins[row] / scale
		}
		times[col] = (float64(starts[col]) + float64(cfg.NFFT)*0.5) / cfg.Fs
	}
	return freqs, times, spectrum
}

func computePSD(samples []float64, opts SignalSpectrumOptions) ([]float64, []float64) {
	fs, nfft, noverlap, padTo, ok := resolveSignalParams(len(samples), opts.Fs, opts.NFFT, opts.NOverlap, opts.PadTo)
	if !ok {
		return nil, nil
	}
	cfg := signalFFTConfig{Fs: fs, NFFT: nfft, NOverlap: noverlap, PadTo: padTo, Window: opts.Window}
	segments, _ := windowedSegments(samples, cfg.NFFT, cfg.NOverlap)
	if len(segments) == 0 {
		return nil, nil
	}
	window := signalWindow(cfg.Window, cfg.NFFT)
	scale := windowPower(window) * cfg.Fs
	if scale == 0 {
		scale = 1
	}
	out := make([]float64, len(fftFrequencies(cfg.Fs, cfg.PadTo)))
	for _, segment := range segments {
		segment = detrendSeries(segment, opts.Detrend)
		power := oneSidedDFTPower(applyWindow(segment, window), cfg.PadTo)
		scaleOneSidedPSD(power, cfg.NFFT)
		for i := range power {
			out[i] += power[i] / scale
		}
	}
	for i := range out {
		out[i] /= float64(len(segments))
	}
	return fftFrequencies(cfg.Fs, cfg.PadTo), out
}

func computeCSDMagnitude(x, y []float64, opts SignalSpectrumOptions) ([]float64, []float64) {
	fs, nfft, noverlap, padTo, ok := resolveSignalParams(min(len(x), len(y)), opts.Fs, opts.NFFT, opts.NOverlap, opts.PadTo)
	if !ok {
		return nil, nil
	}
	cfg := signalFFTConfig{Fs: fs, NFFT: nfft, NOverlap: noverlap, PadTo: padTo, Window: opts.Window}
	segmentsX, _ := windowedSegments(x, cfg.NFFT, cfg.NOverlap)
	segmentsY, _ := windowedSegments(y, cfg.NFFT, cfg.NOverlap)
	if len(segmentsX) == 0 || len(segmentsY) == 0 {
		return nil, nil
	}
	window := signalWindow(cfg.Window, cfg.NFFT)
	scale := windowPower(window) * cfg.Fs
	if scale == 0 {
		scale = 1
	}

	crossSum := make([]complex128, len(fftFrequencies(cfg.Fs, cfg.PadTo)))
	for i := range segmentsX {
		xSegment := detrendSeries(segmentsX[i], opts.Detrend)
		ySegment := detrendSeries(segmentsY[i], opts.Detrend)
		cross := oneSidedDFTCross(applyWindow(xSegment, window), applyWindow(ySegment, window), cfg.PadTo)
		scaleOneSidedCross(cross, cfg.NFFT)
		for k := range cross {
			crossSum[k] += cross[k] / complex(scale, 0)
		}
	}
	out := make([]float64, len(crossSum))
	for i := range out {
		out[i] = cmplx.Abs(crossSum[i] / complex(float64(len(segmentsX)), 0))
	}
	return fftFrequencies(cfg.Fs, cfg.PadTo), out
}

func computeMagnitudeSpectrum(samples []float64, opts SignalSpectrumOptions) ([]float64, []float64) {
	freqs, coeffs := computeSingleSpectrum(samples, opts)
	if len(coeffs) == 0 {
		return nil, nil
	}
	window := signalWindow(opts.Window, resolvedSingleSpectrumNFFT(len(samples), opts.NFFT))
	scale := windowSum(window)
	out := make([]float64, len(coeffs))
	for i, coeff := range coeffs {
		out[i] = cmplx.Abs(coeff) / scale
	}
	return freqs, out
}

func magnitudeSpectrumDBValues(values []float64) []float64 {
	out := make([]float64, len(values))
	for i, value := range values {
		if value <= 0 {
			out[i] = -120
			continue
		}
		out[i] = 20 * math.Log10(value)
	}
	return out
}

func computeAngleSpectrum(samples []float64, opts SignalSpectrumOptions) ([]float64, []float64) {
	freqs, coeffs := computeSingleSpectrum(samples, opts)
	if len(coeffs) == 0 {
		return nil, nil
	}
	out := make([]float64, len(coeffs))
	for i, coeff := range coeffs {
		out[i] = cmplx.Phase(coeff)
	}
	return freqs, out
}

func computePhaseSpectrum(samples []float64, opts SignalSpectrumOptions) ([]float64, []float64) {
	freqs, angles := computeAngleSpectrum(samples, opts)
	if len(angles) == 0 {
		return nil, nil
	}
	unwrapPhaseAngles(angles)
	return freqs, angles
}

func unwrapPhaseAngles(angles []float64) {
	if len(angles) < 2 {
		return
	}
	threshold := math.Pi
	offset := 0.0
	previousRaw := angles[0]
	for i := 1; i < len(angles); i++ {
		raw := angles[i]
		delta := raw - previousRaw
		if delta > threshold {
			offset -= 2 * math.Pi
		} else if delta < -threshold {
			offset += 2 * math.Pi
		}
		angles[i] = raw + offset
		previousRaw = raw
	}
}

func computeSingleSpectrum(samples []float64, opts SignalSpectrumOptions) ([]float64, []complex128) {
	fs, nfft, padTo, ok := resolveSingleSpectrumParams(len(samples), opts.Fs, opts.NFFT, opts.PadTo)
	if !ok {
		return nil, nil
	}
	segment := singleSegment(samples, nfft)
	segment = detrendSeries(segment, opts.Detrend)
	segment = applyWindow(segment, signalWindow(opts.Window, nfft))
	coeffs := fullDFT(segment, padTo)
	freqs, coeffs := selectSpectrumSides(coeffs, fs, opts.Fc, opts.Sides)
	return freqs, coeffs
}

func resolveSingleSpectrumParams(length int, fs float64, nfft, padTo int) (float64, int, int, bool) {
	if length < 2 {
		return 0, 0, 0, false
	}
	if fs <= 0 {
		fs = 2
	}
	if nfft <= 0 {
		nfft = length
	}
	if nfft < 2 {
		return 0, 0, 0, false
	}
	if padTo < nfft {
		padTo = nfft
	}
	return fs, nfft, padTo, true
}

func resolvedSingleSpectrumNFFT(length, nfft int) int {
	if nfft > 0 {
		return nfft
	}
	return length
}

func singleSegment(samples []float64, nfft int) []float64 {
	segment := make([]float64, nfft)
	copy(segment, samples)
	return segment
}

func detrendSeries(values []float64, mode SignalDetrend) []float64 {
	out := append([]float64(nil), values...)
	switch mode {
	case SignalDetrendMean:
		mean := 0.0
		for _, value := range out {
			mean += value
		}
		mean /= float64(len(out))
		for i := range out {
			out[i] -= mean
		}
	case SignalDetrendLinear:
		n := float64(len(out))
		if n <= 1 {
			for i := range out {
				out[i] = 0
			}
			return out
		}
		xMean := (n - 1) * 0.5
		yMean := 0.0
		for _, value := range out {
			yMean += value
		}
		yMean /= n
		variance := 0.0
		covariance := 0.0
		for i, value := range out {
			dx := float64(i) - xMean
			variance += dx * dx
			covariance += dx * (value - yMean)
		}
		if variance == 0 {
			return out
		}
		slope := covariance / variance
		intercept := yMean - slope*xMean
		for i := range out {
			out[i] -= slope*float64(i) + intercept
		}
	}
	return out
}

func computeCoherence(x, y []float64, opts SignalSpectrumOptions) ([]float64, []float64) {
	fs, nfft, noverlap, padTo, ok := resolveSignalParams(min(len(x), len(y)), opts.Fs, opts.NFFT, opts.NOverlap, opts.PadTo)
	if !ok {
		return nil, nil
	}
	cfg := signalFFTConfig{Fs: fs, NFFT: nfft, NOverlap: noverlap, PadTo: padTo, Window: opts.Window}
	segmentsX, _ := windowedSegments(x, cfg.NFFT, cfg.NOverlap)
	segmentsY, _ := windowedSegments(y, cfg.NFFT, cfg.NOverlap)
	if len(segmentsX) == 0 || len(segmentsY) == 0 {
		return nil, nil
	}

	window := signalWindow(cfg.Window, cfg.NFFT)
	pxx := make([]float64, len(fftFrequencies(cfg.Fs, cfg.PadTo)))
	pyy := make([]float64, len(pxx))
	pxy := make([]complex128, len(pxx))
	for i := range segmentsX {
		wx := applyWindow(detrendSeries(segmentsX[i], opts.Detrend), window)
		wy := applyWindow(detrendSeries(segmentsY[i], opts.Detrend), window)
		powerX := oneSidedDFTPower(wx, cfg.PadTo)
		powerY := oneSidedDFTPower(wy, cfg.PadTo)
		cross := oneSidedDFTCross(wx, wy, cfg.PadTo)
		scaleOneSidedPSD(powerX, cfg.NFFT)
		scaleOneSidedPSD(powerY, cfg.NFFT)
		scaleOneSidedCross(cross, cfg.NFFT)
		for k := range pxx {
			pxx[k] += powerX[k]
			pyy[k] += powerY[k]
			pxy[k] += cross[k]
		}
	}
	out := make([]float64, len(pxx))
	count := float64(len(segmentsX))
	for k := range out {
		pxx[k] /= count
		pyy[k] /= count
		pxy[k] /= complex(count, 0)
		denom := pxx[k] * pyy[k]
		if denom > 0 {
			out[k] = math.Min(1, math.Max(0, cmplx.Abs(pxy[k])*cmplx.Abs(pxy[k])/denom))
		}
	}
	return fftFrequencies(cfg.Fs, cfg.PadTo), out
}

func computeCorrelation(x, y []float64, opts CorrelationOptions) ([]float64, []float64) {
	n := min(len(x), len(y))
	if n == 0 {
		return nil, nil
	}
	x = meanCentered(x[:n])
	y = meanCentered(y[:n])
	maxLags := opts.MaxLags
	if maxLags <= 0 || maxLags >= n {
		maxLags = n - 1
	}
	normalize := true
	if opts.Normalize != nil {
		normalize = *opts.Normalize
	}
	lags := make([]float64, 0, 2*maxLags+1)
	values := make([]float64, 0, 2*maxLags+1)
	denom := 1.0
	if normalize {
		denom = math.Sqrt(signalEnergy(x) * signalEnergy(y))
		if denom == 0 {
			denom = 1
		}
	}
	for lag := -maxLags; lag <= maxLags; lag++ {
		sum := 0.0
		for i := range x {
			j := i + lag
			if j < 0 || j >= len(y) {
				continue
			}
			sum += x[i] * y[j]
		}
		if normalize {
			sum /= denom
		}
		lags = append(lags, float64(lag))
		values = append(values, sum)
	}
	return lags, values
}

func plotSpectrumResult(a *Axes, x, y []float64, opts PlotOptions) *SpectrumResult {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}
	line := a.plot(x, y, opts)
	if line == nil {
		return nil
	}
	return &SpectrumResult{
		Line:        line,
		Frequencies: append([]float64(nil), x...),
		Values:      append([]float64(nil), y...),
	}
}

func plotPowerSpectrumResult(a *Axes, x, values []float64, opts PlotOptions) *SpectrumResult {
	plotValues := make([]float64, len(values))
	for i, value := range values {
		if value <= 0 {
			plotValues[i] = -120
			continue
		}
		plotValues[i] = 10 * math.Log10(value)
	}
	result := plotSpectrumResult(a, x, plotValues, opts)
	if result == nil {
		return nil
	}
	setPowerSpectrumTicks(a)
	result.Values = append([]float64(nil), values...)
	return result
}

func setPowerSpectrumTicks(a *Axes) {
	if a == nil || a.YAxis == nil || a.YScale == nil {
		return
	}
	minVal, maxVal := a.YScale.Domain()
	span := maxVal - minVal
	if span <= 0 || !isFinite(span) {
		return
	}
	step := math.Max(10*float64(int(math.Log10(span))), 1)
	start := math.Floor(minVal)
	stop := math.Ceil(maxVal) + 1
	ticks := make([]float64, 0, int((stop-start)/step)+1)
	for tick := start; tick <= stop; tick += step {
		ticks = append(ticks, tick)
	}
	a.YAxis.Locator = ticker.FixedLocator{TicksList: ticks}
	// Matplotlib's set_yticks expands the view interval just enough to keep all
	// explicit ticks visible. Preserve the autoscaled side when it already
	// encloses the outer tick.
	if len(ticks) > 0 {
		a.SetYLim(math.Min(minVal, ticks[0]), math.Max(maxVal, ticks[len(ticks)-1]))
	}
}

func offsetFrequencies(frequencies []float64, offset float64) {
	for i := range frequencies {
		frequencies[i] += offset
	}
}

func plotCorrelationResult(a *Axes, lags, values []float64, opts PlotOptions) *CorrelationResult {
	if len(lags) == 0 || len(values) == 0 {
		return nil
	}
	line := a.plot(lags, values, opts)
	if line == nil {
		return nil
	}
	setLineView(a, lags, values)
	return &CorrelationResult{
		Line:   line,
		Lags:   append([]float64(nil), lags...),
		Values: append([]float64(nil), values...),
	}
}

func setLineView(a *Axes, x, y []float64) {
	if a == nil {
		return
	}
	xMin, xMax := finiteRange(x)
	yMin, yMax := finiteRange(y)
	if xMin == xMax {
		pad := math.Max(1, math.Abs(xMin)*0.05)
		xMin -= pad
		xMax += pad
	}
	a.SetXLim(xMin, xMax)
	if yMin == yMax {
		pad := math.Max(1, math.Abs(yMin)*0.05)
		yMin -= pad
		yMax += pad
	} else {
		pad := (yMax - yMin) * 0.05
		yMin -= pad
		yMax += pad
	}
	a.SetYLim(yMin, yMax)
}

func finiteSeries(values []float64) []float64 {
	out := make([]float64, 0, len(values))
	for _, value := range values {
		if isFinite(value) {
			out = append(out, value)
		}
	}
	return out
}

func finitePairs(x, y []float64) ([]float64, []float64) {
	n := min(len(x), len(y))
	outX := make([]float64, 0, n)
	outY := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		if !isFinite(x[i]) || !isFinite(y[i]) {
			continue
		}
		outX = append(outX, x[i])
		outY = append(outY, y[i])
	}
	return outX, outY
}

func windowedSegments(samples []float64, nfft, noverlap int) ([][]float64, []int) {
	if len(samples) < nfft || nfft <= 0 {
		if len(samples) == 0 {
			return nil, nil
		}
		segment := make([]float64, nfft)
		copy(segment, samples)
		return [][]float64{segment}, []int{0}
	}
	step := nfft - noverlap
	if step <= 0 {
		step = 1
	}
	segments := make([][]float64, 0, 1+(len(samples)-nfft)/step)
	starts := make([]int, 0, cap(segments))
	for start := 0; start+nfft <= len(samples); start += step {
		segment := append([]float64(nil), samples[start:start+nfft]...)
		segments = append(segments, segment)
		starts = append(starts, start)
	}
	if len(segments) == 0 {
		segment := make([]float64, nfft)
		copy(segment, samples)
		return [][]float64{segment}, []int{0}
	}
	return segments, starts
}

func signalWindow(name string, n int) []float64 {
	if n <= 0 {
		return nil
	}
	window := make([]float64, n)
	switch name {
	case "", "hann", "hanning":
		if n == 1 {
			window[0] = 1
			return window
		}
		for i := range window {
			window[i] = 0.5 - 0.5*math.Cos(2*math.Pi*float64(i)/float64(n-1))
		}
	case "hamming":
		if n == 1 {
			window[0] = 1
			return window
		}
		for i := range window {
			window[i] = 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(n-1))
		}
	default:
		for i := range window {
			window[i] = 1
		}
	}
	return window
}

func applyWindow(segment, window []float64) []float64 {
	out := make([]float64, len(segment))
	copy(out, segment)
	for i := range out {
		if i < len(window) {
			out[i] *= window[i]
		}
	}
	return out
}

func windowPower(window []float64) float64 {
	sum := 0.0
	for _, value := range window {
		sum += value * value
	}
	if sum == 0 {
		return 1
	}
	return sum
}

func windowSum(window []float64) float64 {
	sum := 0.0
	for _, value := range window {
		sum += value
	}
	if sum == 0 {
		return 1
	}
	return sum
}

func fftFrequencies(fs float64, n int) []float64 {
	count := n/2 + 1
	freqs := make([]float64, count)
	for i := range freqs {
		freqs[i] = float64(i) * fs / float64(n)
	}
	return freqs
}

func oneSidedDFTPower(samples []float64, n int) []float64 {
	coeffs := oneSidedDFT(samples, n)
	power := make([]float64, len(coeffs))
	for i, value := range coeffs {
		power[i] = real(value)*real(value) + imag(value)*imag(value)
	}
	return power
}

func oneSidedDFTCross(x, y []float64, n int) []complex128 {
	fx := oneSidedDFT(x, n)
	fy := oneSidedDFT(y, n)
	out := make([]complex128, len(fx))
	for i := range out {
		out[i] = fx[i] * cmplx.Conj(fy[i])
	}
	return out
}

func scaleOneSidedPSD(values []float64, nfft int) {
	end := len(values)
	if nfft%2 == 0 && end > 0 {
		end--
	}
	for i := 1; i < end; i++ {
		values[i] *= 2
	}
}

func scaleOneSidedCross(values []complex128, nfft int) {
	end := len(values)
	if nfft%2 == 0 && end > 0 {
		end--
	}
	for i := 1; i < end; i++ {
		values[i] *= 2
	}
}

func fullDFT(samples []float64, n int) []complex128 {
	input := make([]complex128, n)
	for i, value := range samples {
		if i >= n {
			break
		}
		input[i] = complex(value, 0)
	}
	out := make([]complex128, n)
	plan, err := algofft.NewPlan64(n)
	if err != nil {
		return nil
	}
	if err := plan.Forward(out, input); err != nil {
		return nil
	}
	return out
}

func selectSpectrumSides(coeffs []complex128, fs, fc float64, sides SignalSpectrumSides) ([]float64, []complex128) {
	n := len(coeffs)
	if n == 0 {
		return nil, nil
	}
	if sides == SignalSpectrumSidesTwoSided {
		return twoSidedFrequencies(fs, fc, n), twoSidedCoefficients(coeffs)
	}
	count := n/2 + 1
	freqs := make([]float64, count)
	out := make([]complex128, count)
	for i := 0; i < count; i++ {
		freqs[i] = fc + float64(i)*fs/float64(n)
		out[i] = coeffs[i]
	}
	return freqs, out
}

func twoSidedFrequencies(fs, fc float64, n int) []float64 {
	freqs := make([]float64, n)
	start := -n / 2
	if n%2 != 0 {
		start = -(n - 1) / 2
	}
	for i := range freqs {
		freqs[i] = fc + float64(start+i)*fs/float64(n)
	}
	return freqs
}

func twoSidedCoefficients(coeffs []complex128) []complex128 {
	n := len(coeffs)
	center := n / 2
	if n%2 != 0 {
		center = (n-1)/2 + 1
	}
	out := make([]complex128, 0, n)
	out = append(out, coeffs[center:]...)
	out = append(out, coeffs[:center]...)
	return out
}

func oneSidedDFT(samples []float64, n int) []complex128 {
	input := make([]float64, n)
	copy(input, samples)
	count := n/2 + 1
	out := make([]complex128, count)
	for k := 0; k < count; k++ {
		var sum complex128
		for i, value := range input {
			angle := -2 * math.Pi * float64(k*i) / float64(n)
			sum += complex(value, 0) * cmplx.Exp(complex(0, angle))
		}
		out[k] = sum
	}
	return out
}

func scaleSpectrumDB(data [][]float64) [][]float64 {
	out := make([][]float64, len(data))
	for row := range data {
		out[row] = make([]float64, len(data[row]))
		for col, value := range data[row] {
			if value <= 0 {
				out[row][col] = -120
				continue
			}
			out[row][col] = 10 * math.Log10(value)
		}
	}
	return out
}

func centersExtent(centers []float64, fallbackHalfWidth float64) (float64, float64) {
	if len(centers) == 0 {
		return 0, 1
	}
	if len(centers) == 1 {
		return centers[0] - fallbackHalfWidth, centers[0] + fallbackHalfWidth
	}
	half := (centers[1] - centers[0]) * 0.5
	return centers[0] - half, centers[len(centers)-1] + half
}

func meanCentered(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))
	out := append([]float64(nil), values...)
	for i := range out {
		out[i] -= mean
	}
	return out
}

func signalEnergy(values []float64) float64 {
	sum := 0.0
	for _, value := range values {
		sum += value * value
	}
	return sum
}
