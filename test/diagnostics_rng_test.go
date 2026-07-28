package test

import (
	"encoding/json"
	"math"
	"math/rand/v2"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
)

// === from rng_parity_test.go ==========================================

type rngDebugPayload struct {
	NormalData    map[string][]float64 `json:"normal_data"`
	UniformData   map[string][]float64 `json:"uniform_data"`
	HistogramData map[string]struct {
		Edges   []float64 `json:"edges"`
		Heights []float64 `json:"heights"`
	} `json:"histogram_data"`
}

func TestHistogramBarHeightParity(t *testing.T) {
	payload := loadMatplotlibRNGDebugPayload(t)

	basic := normalData(42, 0, 500, 5.0, 1.5)
	density := basic
	strategies1 := normalData(42, 0, 300, 4.0, 1.0)
	strategies2 := normalData(7, 0, 300, 7.0, 1.2)

	wantBasic, ok := payload.HistogramData["hist_basic"]
	if !ok {
		t.Fatalf("missing payload hist_basic")
	}
	wantDensity, ok := payload.HistogramData["hist_density"]
	if !ok {
		t.Fatalf("missing payload hist_density")
	}
	wantStrategies1, ok := payload.HistogramData["hist_strategies_data1"]
	if !ok {
		t.Fatalf("missing payload hist_strategies_data1")
	}
	wantStrategies2, ok := payload.HistogramData["hist_strategies_data2"]
	if !ok {
		t.Fatalf("missing payload hist_strategies_data2")
	}

	gotEdges, gotHeights := goHistogramBinCounts(basic, 0, core.HistNormCount, core.BinStrategySturges)
	compareHistogramHeights(t, "hist_basic", gotEdges, gotHeights, wantBasic.Edges, wantBasic.Heights, 1e-12)

	gotEdges, gotHeights = goHistogramBinCounts(density, 20, core.HistNormDensity, core.BinStrategyAuto)
	compareHistogramHeights(t, "hist_density", gotEdges, gotHeights, wantDensity.Edges, wantDensity.Heights, 1e-12)

	gotEdges, gotHeights = goHistogramBinCounts(strategies1, 15, core.HistNormProbability, core.BinStrategyAuto)
	compareHistogramHeights(t, "hist_strategies_data1", gotEdges, gotHeights, wantStrategies1.Edges, wantStrategies1.Heights, 1e-12)

	gotEdges, gotHeights = goHistogramBinCounts(strategies2, 15, core.HistNormProbability, core.BinStrategyAuto)
	compareHistogramHeights(t, "hist_strategies_data2", gotEdges, gotHeights, wantStrategies2.Edges, wantStrategies2.Heights, 1e-12)
}

func TestHistogramRNGParity(t *testing.T) {
	payload := loadMatplotlibRNGDebugPayload(t)

	compareFloatSlices(t, payload.UniformData["pcg_42_0_1000"], pcgFloat64Samples(42, 0, 1000), "pcg_42_0_1000", 1e-15)
	compareFloatSlices(t, payload.UniformData["pcg_7_0_600"], pcgFloat64Samples(7, 0, 600), "pcg_7_0_600", 1e-15)

	wantBasic := normalData(42, 0, 500, 5.0, 1.5)
	wantDensity := normalData(42, 0, 500, 5.0, 1.5)
	wantStrategies1 := normalData(42, 0, 300, 4.0, 1.0)
	wantStrategies2 := normalData(7, 0, 300, 7.0, 1.2)

	compareFloatSlices(t, payload.NormalData["hist_basic"], wantBasic, "hist_basic", 1e-12)
	compareFloatSlices(t, payload.NormalData["hist_density"], wantDensity, "hist_density", 1e-12)
	compareFloatSlices(t, payload.NormalData["hist_strategies_data1"], wantStrategies1, "hist_strategies_data1", 1e-12)
	compareFloatSlices(t, payload.NormalData["hist_strategies_data2"], wantStrategies2, "hist_strategies_data2", 1e-12)

	// Verify RNG values are consumed in pair order for normal transformation:
	// first sample uses u1/u2 at indices 0/1, second sample uses 2/3, etc.
	checkHistogramRNGOrder(t, payload.UniformData["pcg_42_0_1000"], wantBasic, 5.0, 1.5, 3)
	checkHistogramRNGOrder(t, payload.UniformData["pcg_7_0_600"], wantStrategies2, 7.0, 1.2, 3)
}

func loadMatplotlibRNGDebugPayload(t *testing.T) rngDebugPayload {
	t.Helper()

	script := filepath.Join("matplotlib_ref", "generate.py")
	cmd := selectPythonCommand(t, "uv", script, "--emit-rng-debug")
	if cmd == nil {
		cmd = selectPythonCommand(t, "python3", script, "--emit-rng-debug")
	}
	if cmd == nil {
		t.Skip("matplotlib RNG parity check skipped: no uv/python3 with numpy and matplotlib")
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run matplotlib RNG debug generator: %v\n%s", err, strings.TrimSpace(string(out)))
	}
	outStr := strings.TrimSpace(string(out))
	jsonStart := strings.Index(outStr, "{")
	if jsonStart == -1 {
		t.Fatalf("failed to parse RNG debug JSON: no JSON object in output\n%s", outStr)
	}

	var payload rngDebugPayload
	if err := json.Unmarshal([]byte(outStr[jsonStart:]), &payload); err != nil {
		t.Fatalf("failed to parse RNG debug JSON: %v\n%s", err, outStr[jsonStart:])
	}

	return payload
}

// generatorImportProbe covers the top-level imports test/matplotlib_ref/generate.py
// needs before it can emit anything.
const generatorImportProbe = "import numpy, matplotlib"

func selectPythonCommand(t *testing.T, name, script string, args ...string) *exec.Cmd {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		return nil
	}

	var cmd, probe *exec.Cmd
	if name == "uv" {
		allArgs := append([]string{"run", script}, args...)
		cmd = exec.Command(path, allArgs...)
		probe = exec.Command(path, "run", "python", "-c", generatorImportProbe)
	} else {
		allArgs := append([]string{script}, args...)
		cmd = exec.Command(path, allArgs...)
		probe = exec.Command(path, "-c", generatorImportProbe)
	}

	// Being on PATH is not enough. Bare CI runners ship an interpreter without
	// the generator's dependencies, and running the script anyway turns a missing
	// optional toolchain into an ImportError traceback instead of a skip.
	if probe.Run() != nil {
		return nil
	}
	return cmd
}

func compareFloatSlices(t *testing.T, got, want []float64, label string, eps float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length mismatch: got %d, want %d", label, len(got), len(want))
	}
	for i := range got {
		if math.Abs(got[i]-want[i]) > eps {
			t.Fatalf("%s mismatch at index %d: got %.17g, want %.17g", label, i, got[i], want[i])
		}
	}
}

func compareHistogramHeights(
	t *testing.T,
	label string,
	gotEdges, gotHeights, wantEdges, wantHeights []float64,
	eps float64,
) {
	t.Helper()
	compareFloatSlices(t, gotEdges, wantEdges, label+" edges", eps)
	compareFloatSlices(t, gotHeights, wantHeights, label+" heights", eps)
}

func pcgFloat64Samples(seed1, seed2 uint64, n int) []float64 {
	rng := rand.New(rand.NewPCG(seed1, seed2))
	out := make([]float64, n)
	for i := range out {
		out[i] = rng.Float64()
	}
	return out
}

func checkHistogramRNGOrder(t *testing.T, uniforms, normals []float64, mean, std float64, maxSamples int) {
	t.Helper()
	if maxSamples <= 0 {
		return
	}
	if len(normals) < maxSamples {
		t.Fatalf("normal sample count %d is less than maxSamples %d", len(normals), maxSamples)
	}
	if len(uniforms) < maxSamples*2 {
		t.Fatalf("uniform sample count %d is less than needed %d", len(uniforms), maxSamples*2)
	}

	for i := 0; i < maxSamples; i++ {
		want := normalFromUniformPair(uniforms[i*2], uniforms[i*2+1], mean, std)
		if math.Abs(normals[i]-want) > 1e-15 {
			t.Fatalf("pair-order mismatch at sample %d: got %.17g, want %.17g", i, normals[i], want)
		}
	}
}

func normalFromUniformPair(u1, u2, mean, std float64) float64 {
	return math.Sqrt(-2*math.Log(u1))*math.Cos(2*math.Pi*u2)*std + mean
}

func goHistogramBinCounts(data []float64, bins int, norm core.HistNorm, strategy core.BinStrategy) ([]float64, []float64) {
	h := &core.Hist2D{
		Data:     data,
		Bins:     bins,
		Norm:     norm,
		BinStrat: strategy,
	}
	return h.BinCounts()
}
