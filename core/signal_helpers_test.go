package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/ticker"
)

func TestAxesSpecgramFindsDominantFrequency(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())

	samples := sineWave(8, 64, 128, 0)
	result := ax.Specgram(samples, SpecgramOptions{
		Fs:       64,
		NFFT:     32,
		NOverlap: 16,
	})
	if result == nil {
		t.Fatal("Specgram() returned nil")
	}
	if result.Image == nil {
		t.Fatal("Specgram() should create an image")
	}
	if got, want := len(result.Frequencies), 17; got != want {
		t.Fatalf("frequency bin count = %d, want %d", got, want)
	}
	if len(result.Times) < 2 {
		t.Fatalf("time bin count = %d, want >= 2", len(result.Times))
	}

	peak := dominantFrequency(result.Frequencies, result.Spectrum)
	if math.Abs(peak-8) > 2 {
		t.Fatalf("dominant spectrogram frequency = %v, want about 8", peak)
	}
}

func TestAxesSignalAnalysisHelpers(t *testing.T) {
	x := sineWave(5, 64, 128, 0)
	y := sineWave(5, 64, 128, math.Pi/4)
	opts := SignalSpectrumOptions{
		Fs:       64,
		NFFT:     64,
		NOverlap: 32,
	}

	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	psd := ax.PSD(x, opts)
	if psd == nil || psd.Line == nil {
		t.Fatal("PSD() returned nil")
	}
	if got := dominantLineFrequency(psd); math.Abs(got-5) > 1 {
		t.Fatalf("PSD dominant frequency = %v, want about 5", got)
	}

	fig = NewFigure(400, 300)
	ax = fig.AddAxes(unitRect())
	csd := ax.CSD(x, y, opts)
	if csd == nil || csd.Line == nil {
		t.Fatal("CSD() returned nil")
	}
	if got := dominantLineFrequency(csd); math.Abs(got-5) > 1 {
		t.Fatalf("CSD dominant frequency = %v, want about 5", got)
	}

	fig = NewFigure(400, 300)
	ax = fig.AddAxes(unitRect())
	cohere := ax.Cohere(x, y, opts)
	if cohere == nil || cohere.Line == nil {
		t.Fatal("Cohere() returned nil")
	}
	peakIndex := argmax(cohere.Values)
	if peakIndex < 0 || cohere.Values[peakIndex] < 0.9 {
		t.Fatalf("coherence peak = %v, want >= 0.9", cohere.Values[peakIndex])
	}

	fig = NewFigure(400, 300)
	ax = fig.AddAxes(unitRect())
	acorr := ax.ACorr(x, CorrelationOptions{MaxLags: 8})
	if acorr == nil || acorr.Line == nil {
		t.Fatal("ACorr() returned nil")
	}
	if got, want := len(acorr.Lags), 17; got != want {
		t.Fatalf("lag count = %d, want %d", got, want)
	}
	zeroLag := indexOf(acorr.Lags, 0)
	if zeroLag < 0 {
		t.Fatal("zero lag not found in ACorr output")
	}
	if math.Abs(acorr.Values[zeroLag]-1) > 0.05 {
		t.Fatalf("ACorr zero-lag value = %v, want about 1", acorr.Values[zeroLag])
	}
}

func TestAxesPowerSpectrumDisplayMatchesMatplotlib(t *testing.T) {
	x := make([]float64, 256)
	y := make([]float64, 256)
	for i := range x {
		time := float64(i) / 64
		x[i] = math.Sin(2*math.Pi*8*time) +
			0.35*math.Sin(2*math.Pi*15*time+0.2) +
			0.1*math.Cos(2*math.Pi*3*time)
		y[i] = 0.8*math.Sin(2*math.Pi*8*time+0.45) +
			0.25*math.Sin(2*math.Pi*15*time-0.3) +
			0.12*math.Cos(2*math.Pi*20*time)
	}
	opts := SignalSpectrumOptions{
		Fs: 64, Fc: 2, NFFT: 64, NOverlap: 32, PadTo: 128,
		Window: "hann", Detrend: SignalDetrendMean,
	}

	for name, testCase := range map[string]struct {
		run      func(*Axes) *SpectrumResult
		expected [3]float64
	}{
		"psd": {
			run:      func(ax *Axes) *SpectrumResult { return ax.PSD(x, opts) },
			expected: [3]float64{2.6620585620737396e-08, 0.32810728602295575, 0.04020410895017642},
		},
		"csd": {
			run:      func(ax *Axes) *SpectrumResult { return ax.CSD(x, y, opts) },
			expected: [3]float64{2.2555878280023126e-09, 0.26248334333136814, 0.028715938883848956},
		},
	} {
		t.Run(name, func(t *testing.T) {
			ax := NewFigure(400, 300).AddAxes(unitRect())
			result := testCase.run(ax)
			if result == nil || result.Line == nil || len(result.Values) == 0 {
				t.Fatalf("%s returned no spectrum", name)
			}
			for expectedIndex, bin := range []int{0, 16, 30} {
				if got, want := result.Values[bin], testCase.expected[expectedIndex]; math.Abs(got-want) > 1e-12 {
					t.Fatalf("%s value[%d] = %.17g, want Matplotlib %.17g", name, bin, got, want)
				}
			}
			if got := result.Frequencies[0]; math.Abs(got-2) > 1e-12 {
				t.Fatalf("%s first frequency = %v, want Fc offset 2", name, got)
			}
			if got, want := result.Line.XY[0].Y, 10*math.Log10(result.Values[0]); math.Abs(got-want) > 1e-12 {
				t.Fatalf("%s plotted value = %v, want 10*log10(raw) = %v", name, got, want)
			}
			if ax.XLabel() != "Frequency" {
				t.Fatalf("%s x label = %q, want Frequency", name, ax.XLabel())
			}
			if _, ok := ax.YAxis.Locator.(ticker.FixedLocator); !ok {
				t.Fatalf("%s y locator = %T, want Matplotlib power-spectrum FixedLocator", name, ax.YAxis.Locator)
			}
		})
	}
}

func TestAxesSpecgramUsesMatplotlibImageExtent(t *testing.T) {
	ax := NewFigure(400, 300).AddAxes(unitRect())
	result := ax.Specgram(sineWave(8, 64, 384, 0), SpecgramOptions{
		Fs: 64, NFFT: 64, NOverlap: 48, PadTo: 128,
	})
	if result == nil || result.Image == nil {
		t.Fatal("Specgram returned no image")
	}
	for name, values := range map[string]struct {
		got  float64
		want float64
	}{
		"xmin": {result.Image.XMin, 0.375},
		"xmax": {result.Image.XMax, 5.625},
		"ymin": {result.Image.YMin, 0},
		"ymax": {result.Image.YMax, 32},
	} {
		if math.Abs(values.got-values.want) > 1e-12 {
			t.Fatalf("Specgram %s = %v, want %v", name, values.got, values.want)
		}
	}
}

func TestAxesSpectrumVariants(t *testing.T) {
	x := sineWave(5, 64, 128, math.Pi/6)
	opts := SignalSpectrumOptions{
		Fs:   64,
		NFFT: 64,
	}

	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	mag := ax.MagnitudeSpectrum(x, opts)
	if mag == nil || mag.Line == nil {
		t.Fatal("MagnitudeSpectrum() returned nil")
	}
	if got := dominantLineFrequency(mag); math.Abs(got-5) > 1 {
		t.Fatalf("MagnitudeSpectrum dominant frequency = %v, want about 5", got)
	}

	fig = NewFigure(400, 300)
	ax = fig.AddAxes(unitRect())
	angle := ax.AngleSpectrum(x, opts)
	if angle == nil || angle.Line == nil {
		t.Fatal("AngleSpectrum() returned nil")
	}
	if len(angle.Frequencies) != 33 || len(angle.Values) != 33 {
		t.Fatalf("AngleSpectrum bins = (%d, %d), want 33", len(angle.Frequencies), len(angle.Values))
	}
	for _, value := range angle.Values {
		if value < -math.Pi || value > math.Pi {
			t.Fatalf("AngleSpectrum value = %v, want within [-pi, pi]", value)
		}
	}

	fig = NewFigure(400, 300)
	ax = fig.AddAxes(unitRect())
	phase := ax.PhaseSpectrum(x, opts)
	if phase == nil || phase.Line == nil {
		t.Fatal("PhaseSpectrum() returned nil")
	}
	if len(phase.Frequencies) != 33 || len(phase.Values) != 33 {
		t.Fatalf("PhaseSpectrum bins = (%d, %d), want 33", len(phase.Frequencies), len(phase.Values))
	}
}

func TestMagnitudeSpectrumSidesScaleAndFrequencyOffset(t *testing.T) {
	samples := make([]float64, 8)
	for i := range samples {
		samples[i] = math.Cos(2 * math.Pi * float64(i) / float64(len(samples)))
	}

	freqs, values := computeMagnitudeSpectrum(samples, SignalSpectrumOptions{
		Fs:     8,
		Window: "none",
		Sides:  SignalSpectrumSidesTwoSided,
		Scale:  SignalSpectrumScaleDB,
		Fc:     10,
	})

	wantFreqs := []float64{6, 7, 8, 9, 10, 11, 12, 13}
	if len(freqs) != len(wantFreqs) {
		t.Fatalf("frequency count = %d, want %d", len(freqs), len(wantFreqs))
	}
	for i, want := range wantFreqs {
		if math.Abs(freqs[i]-want) > 1e-12 {
			t.Fatalf("freq[%d] = %v, want %v", i, freqs[i], want)
		}
	}

	wantPeak := 0.5
	if math.Abs(values[3]-wantPeak) > 1e-10 {
		t.Fatalf("negative peak magnitude = %v, want Matplotlib-returned linear spectrum %v", values[3], wantPeak)
	}
	if math.Abs(values[5]-wantPeak) > 1e-10 {
		t.Fatalf("positive peak magnitude = %v, want Matplotlib-returned linear spectrum %v", values[5], wantPeak)
	}
}

func TestSingleSpectrumOneSidedEvenNyquistFrequencyMatchesMatplotlib(t *testing.T) {
	freqs, values := computeMagnitudeSpectrum([]float64{1, 0, -1, 0}, SignalSpectrumOptions{
		Fs:     8,
		Fc:     10,
		Window: "none",
		Sides:  SignalSpectrumSidesOneSided,
	})

	wantFreqs := []float64{10, 12, 14}
	if len(freqs) != len(wantFreqs) || len(values) != len(wantFreqs) {
		t.Fatalf("spectrum size = (%d, %d), want %d", len(freqs), len(values), len(wantFreqs))
	}
	for i, want := range wantFreqs {
		if math.Abs(freqs[i]-want) > 1e-12 {
			t.Fatalf("freq[%d] = %v, want Matplotlib fftfreq bin %v", i, freqs[i], want)
		}
	}
}

func TestAxesMagnitudeSpectrumPlotsDBButReturnsLinearValues(t *testing.T) {
	samples := make([]float64, 8)
	for i := range samples {
		samples[i] = math.Cos(2 * math.Pi * float64(i) / float64(len(samples)))
	}
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(unitRect())

	result := ax.MagnitudeSpectrum(samples, SignalSpectrumOptions{
		Fs:     8,
		Window: "none",
		Scale:  SignalSpectrumScaleDB,
	})

	if result == nil || result.Line == nil {
		t.Fatal("MagnitudeSpectrum returned nil result")
	}
	if got, want := result.Values[1], 0.5; math.Abs(got-want) > 1e-10 {
		t.Fatalf("returned magnitude value = %v, want Matplotlib linear spec %v", got, want)
	}
	if got, want := result.Line.XY[1].Y, 20*math.Log10(0.5); math.Abs(got-want) > 1e-10 {
		t.Fatalf("plotted magnitude value = %v, want dB plot value %v", got, want)
	}
}

func TestSpectrumDetrendMeanRemovesConstantComponent(t *testing.T) {
	freqs, values := computeMagnitudeSpectrum([]float64{3, 3, 3, 3}, SignalSpectrumOptions{
		Fs:      4,
		Window:  "none",
		Detrend: SignalDetrendMean,
	})
	if len(freqs) != 3 || len(values) != 3 {
		t.Fatalf("spectrum size = (%d, %d), want 3", len(freqs), len(values))
	}
	for i, value := range values {
		if math.Abs(value) > 1e-12 {
			t.Fatalf("detrended magnitude[%d] = %v, want 0", i, value)
		}
	}
}

func TestUnwrapPhaseAngles(t *testing.T) {
	angles := []float64{0, 3.0, -3.0, -2.8, 2.9}

	unwrapPhaseAngles(angles)

	want := []float64{0, 3.0, -3.0 + 2*math.Pi, -2.8 + 2*math.Pi, 2.9}
	for i := range want {
		if math.Abs(angles[i]-want[i]) > 1e-12 {
			t.Fatalf("unwrapped angle[%d] = %v, want %v", i, angles[i], want[i])
		}
	}
}

func TestUnwrapPhaseAnglesLeavesNearPiPositiveJumpWrapped(t *testing.T) {
	angles := []float64{0, 0.2, 0.4, 0.6, 0.8, 0.9, 0, math.Pi - 0.1}

	unwrapPhaseAngles(angles)

	want := math.Pi - 0.1
	if math.Abs(angles[7]-want) > 1e-12 {
		t.Fatalf("unwrapped angle = %v, want NumPy unwrap value %v", angles[7], want)
	}
}

func dominantFrequency(freqs []float64, spectrum [][]float64) float64 {
	bestIndex := -1
	bestValue := math.Inf(-1)
	for row := range spectrum {
		sum := 0.0
		for _, value := range spectrum[row] {
			sum += value
		}
		if sum > bestValue {
			bestValue = sum
			bestIndex = row
		}
	}
	if bestIndex < 0 {
		return 0
	}
	return freqs[bestIndex]
}

func dominantLineFrequency(result *SpectrumResult) float64 {
	if result == nil {
		return 0
	}
	index := argmax(result.Values)
	if index < 0 {
		return 0
	}
	return result.Frequencies[index]
}

func argmax(values []float64) int {
	best := -1
	bestValue := math.Inf(-1)
	for i, value := range values {
		if value > bestValue {
			bestValue = value
			best = i
		}
	}
	return best
}

func indexOf(values []float64, want float64) int {
	for i, value := range values {
		if value == want {
			return i
		}
	}
	return -1
}

func sineWave(freq, fs float64, count int, phase float64) []float64 {
	out := make([]float64, count)
	for i := range out {
		t := float64(i) / fs
		out[i] = math.Sin(2*math.Pi*freq*t + phase)
	}
	return out
}
