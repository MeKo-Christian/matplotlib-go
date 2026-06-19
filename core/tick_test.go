package core

import (
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/transform"
)

func strictlyIncreasing(xs []float64) bool {
	for i := 1; i < len(xs); i++ {
		if !(xs[i] > xs[i-1]) {
			return false
		}
	}
	return true
}

func TestLinearLocator_BasicRanges(t *testing.T) {
	cases := [][2]float64{{-1, 1}, {0, 1e-9}, {1, 1e6}, {-1e6, -1}, {2, 2}}
	targets := []int{3, 5, 7}
	for _, c := range cases {
		for _, n := range targets {
			ticks := (LinearLocator{}).Ticks(c[0], c[1], n)
			if len(ticks) == 0 {
				t.Fatalf("no ticks for range %+v", c)
			}
			if !strictlyIncreasing(ticks) {
				t.Fatalf("ticks not strictly increasing: %+v", ticks)
			}
			minVal, maxVal := c[0], c[1]
			if minVal > maxVal {
				minVal, maxVal = maxVal, minVal
			}
			if ticks[0] > minVal+1e-12 {
				t.Fatalf("first tick %v > min %v", ticks[0], minVal)
			}
			if ticks[len(ticks)-1] < maxVal-1e-12 {
				t.Fatalf("last tick %v < max %v", ticks[len(ticks)-1], maxVal)
			}
			// Do not assert exact count band here; coverage and monotonicity suffice.
		}
	}
}

func TestLinearLocator_Property(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 200; i++ {
		a := r.Float64()*2e6 - 1e6
		b := a + (r.Float64()*2e6 + 1e-9)
		n := 2 + int(r.Float64()*8)
		ticks := (LinearLocator{}).Ticks(a, b, n)
		if !strictlyIncreasing(ticks) {
			t.Fatalf("not increasing: %+v", ticks)
		}
		// Coverage
		minVal, maxVal := a, b
		if minVal > maxVal {
			minVal, maxVal = maxVal, minVal
		}
		if ticks[0] > minVal+1e-9 {
			t.Fatalf("first > min: %v > %v", ticks[0], minVal)
		}
		if ticks[len(ticks)-1] < maxVal-1e-9 {
			t.Fatalf("last < max: %v < %v", ticks[len(ticks)-1], maxVal)
		}
	}
}

func TestLinearLocator_HistogramStyleRange(t *testing.T) {
	ticks := (LinearLocator{}).Ticks(0, 0.196, 6)
	want := []float64{0, 0.05, 0.1, 0.15, 0.2}
	if len(ticks) != len(want) {
		t.Fatalf("tick count mismatch: got %v want %v", ticks, want)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-12 {
			t.Fatalf("tick %d mismatch: got %.17g want %.17g", i, ticks[i], want[i])
		}
	}
}

func TestLinearLocator_TargetIsMaximumDensity(t *testing.T) {
	ticks := (LinearLocator{}).Ticks(0, 80, 6)
	want := []float64{0, 20, 40, 60, 80}
	if len(ticks) != len(want) {
		t.Fatalf("tick count mismatch: got %v want %v", ticks, want)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-12 {
			t.Fatalf("tick %d mismatch: got %.17g want %.17g", i, ticks[i], want[i])
		}
	}
}

func TestLinearLocator_NumTicksUsesMatplotlibEvenSpacing(t *testing.T) {
	ticks := (LinearLocator{NumTicks: 5}).Ticks(2, 10, 0)
	want := []float64{2, 4, 6, 8, 10}
	if len(ticks) != len(want) {
		t.Fatalf("tick count mismatch: got %v want %v", ticks, want)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-12 {
			t.Fatalf("tick %d mismatch: got %.17g want %.17g", i, ticks[i], want[i])
		}
	}
}

func TestLinearLocator_PresetOverridesNumTicks(t *testing.T) {
	locator := LinearLocator{
		NumTicks: 4,
		Presets: map[[2]float64][]float64{
			{0, 1}: {0.2, 0.4, 0.8},
		},
	}
	ticks := locator.Ticks(0, 1, 0)
	want := []float64{0.2, 0.4, 0.8}
	if len(ticks) != len(want) {
		t.Fatalf("tick count mismatch: got %v want %v", ticks, want)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-12 {
			t.Fatalf("tick %d mismatch: got %.17g want %.17g", i, ticks[i], want[i])
		}
	}
	ticks[0] = 99
	if locator.Presets[[2]float64{0, 1}][0] == 99 {
		t.Fatal("LinearLocator preset ticks should be cloned")
	}
}

func TestMaxNLocatorOptionsMatchMatplotlibSemantics(t *testing.T) {
	symmetric := (MaxNLocator{N: 4, Symmetric: true}).Ticks(1, 3, 0)
	if len(symmetric) == 0 || symmetric[0] != -3 || symmetric[len(symmetric)-1] != 3 {
		t.Fatalf("symmetric ticks = %v, want to cover [-3, 3]", symmetric)
	}

	integer := (MaxNLocator{N: 4, Integer: true}).Ticks(0.2, 3.8, 0)
	for _, tick := range integer {
		if tick != math.Round(tick) {
			t.Fatalf("integer locator tick %v is not integral: %v", tick, integer)
		}
	}

	relaxed := (MaxNLocator{N: 4, Integer: true, MinTicks: 3}).Ticks(0.2, 0.8, 0)
	if len(relaxed) == 0 {
		t.Fatal("integer locator should relax integer constraint when not enough integer ticks exist")
	}
	allInteger := true
	for _, tick := range relaxed {
		allInteger = allInteger && tick == math.Round(tick)
	}
	if allInteger {
		t.Fatalf("integer locator did not relax integer constraint: %v", relaxed)
	}

	pruned := (MaxNLocator{N: 5, Prune: "both"}).Ticks(0, 10, 0)
	if len(pruned) == 0 || pruned[0] <= 0 || pruned[len(pruned)-1] >= 10 {
		t.Fatalf("pruned ticks = %v, want lower and upper ticks removed", pruned)
	}

	customSteps := (MaxNLocator{N: 4, Steps: []float64{2, 4}}).Ticks(0, 8, 0)
	if len(customSteps) < 2 || customSteps[1]-customSteps[0] != 2 {
		t.Fatalf("custom-step ticks = %v, want step 2", customSteps)
	}
}

func TestMaxNLocatorKeepsTinyAndOffsetTicksDistinct(t *testing.T) {
	cases := []struct {
		name   string
		minVal float64
		maxVal float64
	}{
		{name: "large-offset", minVal: 1e12 + 0.1, maxVal: 1e12 + 0.9},
		{name: "tiny-positive", minVal: 1e-12, maxVal: 5e-12},
		{name: "negative-offset", minVal: -1e12 - 0.9, maxVal: -1e12 - 0.1},
	}
	for _, tc := range cases {
		ticks := (MaxNLocator{N: 4}).Ticks(tc.minVal, tc.maxVal, 0)
		if len(ticks) < 3 {
			t.Fatalf("%s ticks = %v, want at least three distinct ticks", tc.name, ticks)
		}
		if !strictlyIncreasing(ticks) {
			t.Fatalf("%s ticks not strictly increasing: %v", tc.name, ticks)
		}
		if ticks[0] > tc.minVal || ticks[len(ticks)-1] < tc.maxVal {
			t.Fatalf("%s ticks = %.17g, do not cover %.17g..%.17g", tc.name, ticks, tc.minVal, tc.maxVal)
		}
	}
}

func TestIndexLocator_Basic(t *testing.T) {
	ticks := (IndexLocator{Base: 3, Offset: 1}).Ticks(0, 10, 0)
	want := []float64{1, 4, 7, 10}
	if len(ticks) != len(want) {
		t.Fatalf("IndexLocator tick count = %d, want %d (%v)", len(ticks), len(want), ticks)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-12 {
			t.Fatalf("IndexLocator tick %d = %v, want %v", i, ticks[i], want[i])
		}
	}
}

func TestLogLocator_MajorsMonotone(t *testing.T) {
	bases := []float64{2, 10}
	for _, b := range bases {
		l := LogLocator{Base: b}
		ticks := l.Ticks(1, 1e6, 0)
		if len(ticks) == 0 {
			t.Fatalf("no log ticks for base %v", b)
		}
		if !strictlyIncreasing(ticks) {
			t.Fatalf("log ticks not increasing: %+v", ticks)
		}
	}
}

func TestLogLocator_MinorsBetweenMajors(t *testing.T) {
	l := LogLocator{Base: 10, Minor: true}
	ticks := l.Ticks(1, 1e3, 0)
	if !strictlyIncreasing(ticks) {
		t.Fatalf("log ticks not increasing: %+v", ticks)
	}
	// Must contain the canonical set within [1,1000]
	want := []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000}
	// Build a map for quick lookup with tolerance
	has := func(v float64) bool {
		for _, t := range ticks {
			if math.Abs(t-v) <= 1e-12 {
				return true
			}
		}
		return false
	}
	for _, v := range want {
		if !has(v) {
			t.Fatalf("missing expected tick %v in %+v", v, ticks)
		}
	}
}

func TestLogLocatorSubsAutoAndAllModes(t *testing.T) {
	auto := (LogLocator{Base: 10, SubsMode: "auto"}).Ticks(1, 100, 0)
	for _, major := range []float64{0.1, 1, 10, 100, 1000} {
		for _, tick := range auto {
			if tick == major {
				t.Fatalf("auto subs should omit integer powers, got %v", auto)
			}
		}
	}
	for _, want := range []float64{0.2, 0.9, 2, 3, 9, 20, 90, 200, 9000} {
		if !hasTick(auto, want) {
			t.Fatalf("auto subs missing %v in %v", want, auto)
		}
	}

	all := (LogLocator{Base: 10, SubsMode: "all"}).Ticks(1, 100, 0)
	for _, want := range []float64{0.1, 0.2, 0.9, 1, 2, 9, 10, 20, 90, 100, 1000, 9000} {
		if !hasTick(all, want) {
			t.Fatalf("all subs missing %v in %v", want, all)
		}
	}

	denseMinor := (LogLocator{Base: 10, Minor: true, SubsMode: "auto"}).Ticks(1, 1e12, 0)
	if len(denseMinor) != 0 {
		t.Fatalf("dense auto minor locator = %v, want no ticks", denseMinor)
	}
}

func TestLogLocatorRespectsTickBudgetAndSuppressesDenseMinors(t *testing.T) {
	majors := (LogLocator{Base: 10, NumTicks: 4}).Ticks(1, 1e12, 0)
	if len(majors) != 7 {
		t.Fatalf("budgeted log major ticks = %v, want Matplotlib's one-stride-padded thinned set", majors)
	}
	if !strictlyIncreasing(majors) {
		t.Fatalf("budgeted log major ticks not increasing: %v", majors)
	}
	for _, tick := range majors {
		if !isPowerOfTen(tick) {
			t.Fatalf("budgeted log major tick %v is not a power of ten: %v", tick, majors)
		}
	}

	minors := (LogLocator{Base: 10, Subs: []float64{2, 5}, NumTicks: 4}).Ticks(1, 1e12, 0)
	if len(minors) != 0 {
		t.Fatalf("dense explicit minor ticks = %v, want suppressed when major stride is > 1", minors)
	}
	all := (LogLocator{Base: 10, SubsMode: "all", NumTicks: 4}).Ticks(1, 1e12, 0)
	if len(all) == 0 || !strictlyIncreasing(all) {
		t.Fatalf("dense all-sub mode should fall back to thinned major ticks, got %v", all)
	}
}

func TestLogLocatorStrideAndInvalidDomains(t *testing.T) {
	cases := []struct {
		name   string
		loc    LogLocator
		minVal float64
		maxVal float64
		want   []float64
	}{
		{
			name:   "base2-budgeted",
			loc:    LogLocator{Base: 2, NumTicks: 4},
			minVal: 1,
			maxVal: 1024,
			want:   []float64{0.125, 1, 8, 64, 512, 4096, 32768},
		},
		{
			name:   "base10-budgeted",
			loc:    LogLocator{Base: 10, NumTicks: 4},
			minVal: 1,
			maxVal: 1e12,
			want:   []float64{1e-3, 1, 1e3, 1e6, 1e9, 1e12, 1e15},
		},
		{
			name:   "inverted-domain",
			loc:    LogLocator{Base: 10},
			minVal: 100,
			maxVal: 1,
			want:   []float64{0.1, 1, 10, 100, 1000},
		},
	}
	for _, tc := range cases {
		ticks := tc.loc.Ticks(tc.minVal, tc.maxVal, 0)
		if len(ticks) != len(tc.want) {
			t.Fatalf("%s ticks = %v, want %v", tc.name, ticks, tc.want)
		}
		for i, want := range tc.want {
			if math.Abs(ticks[i]-want) > 1e-12*math.Max(1, math.Abs(want)) {
				t.Fatalf("%s tick %d = %.17g, want %.17g (%v)", tc.name, i, ticks[i], want, ticks)
			}
		}
	}

	if got := (LogLocator{Base: 10}).Ticks(-1, 100, 0); len(got) != 0 {
		t.Fatalf("non-positive log domain ticks = %v, want none", got)
	}
	if got := (LogLocator{Base: 1}).Ticks(1, 100, 0); len(got) != 0 {
		t.Fatalf("invalid-base log ticks = %v, want none", got)
	}
}

func hasTick(ticks []float64, want float64) bool {
	for _, tick := range ticks {
		if math.Abs(tick-want) <= 1e-12*math.Max(1, math.Abs(want)) {
			return true
		}
	}
	return false
}

func TestSymLogLocator_LinearRange(t *testing.T) {
	ticks := (SymLogLocator{Base: 10, LinThresh: 1}).Ticks(-0.5, 0.5, 0)
	want := []float64{-0.5, 0, 0.5}
	if len(ticks) != len(want) {
		t.Fatalf("SymLogLocator tick count = %d, want %d (%v)", len(ticks), len(want), ticks)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-12 {
			t.Fatalf("SymLogLocator tick %d = %v, want %v", i, ticks[i], want[i])
		}
	}
}

func TestSymLogLocator_LogRanges(t *testing.T) {
	ticks := (SymLogLocator{Base: 10, LinThresh: 1}).Ticks(-1000, 1000, 0)
	want := []float64{-1000, -100, -10, -1, 0, 1, 10, 100, 1000}
	if len(ticks) != len(want) {
		t.Fatalf("SymLogLocator tick count = %d, want %d (%v)", len(ticks), len(want), ticks)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-12 {
			t.Fatalf("SymLogLocator tick %d = %v, want %v", i, ticks[i], want[i])
		}
	}
}

func TestAsinhLocator_RoundsSymmetricTicks(t *testing.T) {
	ticks := (AsinhLocator{LinearWidth: 1, Base: 10}).Ticks(-100, 100, 0)
	want := []float64{-100, -10, -1, 0, 1, 10, 100}
	if len(ticks) != len(want) {
		t.Fatalf("AsinhLocator tick count = %d, want %d (%v)", len(ticks), len(want), ticks)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-12 {
			t.Fatalf("AsinhLocator tick %d = %v, want %v", i, ticks[i], want[i])
		}
	}
}

func TestAsinhLocator_MinorSubs(t *testing.T) {
	ticks := (AsinhLocator{LinearWidth: 1, Base: 10, Subs: []float64{2, 5}}).Ticks(1, 100, 0)
	has := func(v float64) bool {
		for _, tick := range ticks {
			if math.Abs(tick-v) <= 1e-12 {
				return true
			}
		}
		return false
	}
	for _, want := range []float64{2, 5, 20, 50} {
		if !has(want) {
			t.Fatalf("AsinhLocator missing tick %v in %v", want, ticks)
		}
	}
}

func TestLogitLocator_MajorIdealTicks(t *testing.T) {
	ticks := (LogitLocator{}).Ticks(0.001, 0.999, 0)
	want := []float64{0.001, 0.01, 0.1, 0.5, 0.9, 0.99, 0.999}
	if len(ticks) != len(want) {
		t.Fatalf("LogitLocator tick count = %d, want %d (%v)", len(ticks), len(want), ticks)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-12 {
			t.Fatalf("LogitLocator tick %d = %v, want %v", i, ticks[i], want[i])
		}
	}
}

func TestLogitLocator_MinorTicks(t *testing.T) {
	ticks := (LogitLocator{Minor: true}).Ticks(0.01, 0.99, 0)
	has := func(v float64) bool {
		for _, tick := range ticks {
			if math.Abs(tick-v) <= 1e-12 {
				return true
			}
		}
		return false
	}
	for _, want := range []float64{0.02, 0.2, 0.8, 0.98} {
		if !has(want) {
			t.Fatalf("LogitLocator missing minor tick %v in %v", want, ticks)
		}
	}
}

func TestMinorLinearLocator_SubdividesMajors(t *testing.T) {
	minors := (MinorLinearLocator{N: 5}).Ticks(0, 10, 0)
	if len(minors) == 0 {
		t.Fatal("expected minor ticks")
	}
	// Minor ticks should not coincide with major ticks
	majors := (LinearLocator{}).Ticks(0, 10, 6)
	majorSet := map[float64]bool{}
	for _, m := range majors {
		majorSet[m] = true
	}
	for _, v := range minors {
		for mj := range majorSet {
			if math.Abs(v-mj) < 1e-10 {
				t.Errorf("minor tick %v coincides with major tick %v", v, mj)
			}
		}
	}
	// Should be strictly increasing
	if !strictlyIncreasing(minors) {
		t.Errorf("minor ticks not strictly increasing: %v", minors)
	}
}

func TestMinorLinearLocator_DefaultN(t *testing.T) {
	// N=0 should default to 5
	m0 := (MinorLinearLocator{N: 0}).Ticks(0, 10, 0)
	m5 := (MinorLinearLocator{N: 5}).Ticks(0, 10, 0)
	if len(m0) != len(m5) {
		t.Errorf("N=0 should default to N=5: got %d vs %d ticks", len(m0), len(m5))
	}
}

func TestFixedLocator_SortsAndDedupes(t *testing.T) {
	ticks := (FixedLocator{TicksList: []float64{3, 1, 2, 2}}).Ticks(0, 10, 0)
	want := []float64{1, 2, 3}
	if len(ticks) != len(want) {
		t.Fatalf("FixedLocator tick count = %d, want %d (%v)", len(ticks), len(want), ticks)
	}
	for i := range want {
		if ticks[i] != want[i] {
			t.Fatalf("FixedLocator tick %d = %v, want %v", i, ticks[i], want[i])
		}
	}
}

func TestFixedLocator_NbinsSubsamplesIncludingSmallestAbs(t *testing.T) {
	ticks := (FixedLocator{
		TicksList: []float64{-4, -2, 0, 2, 4, 6},
		Nbins:     2,
	}).Ticks(0, 10, 0)
	want := []float64{0, 6}
	if len(ticks) != len(want) {
		t.Fatalf("FixedLocator subsampled tick count = %d, want %d (%v)", len(ticks), len(want), ticks)
	}
	for i := range want {
		if ticks[i] != want[i] {
			t.Fatalf("FixedLocator subsampled tick %d = %v, want %v", i, ticks[i], want[i])
		}
	}
}

func TestMultipleLocator_Basic(t *testing.T) {
	ticks := (MultipleLocator{Base: 2.5}).Ticks(0.5, 8.5, 0)
	want := []float64{0, 2.5, 5, 7.5, 10}
	if len(ticks) != len(want) {
		t.Fatalf("MultipleLocator tick count = %d, want %d (%v)", len(ticks), len(want), ticks)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-12 {
			t.Fatalf("MultipleLocator tick %d = %v, want %v", i, ticks[i], want[i])
		}
	}
}

func TestAutoMinorLocator_UsesMajorLocatorStepLikeMatplotlib(t *testing.T) {
	ticks := (AutoMinorLocator{N: 3, Major: MultipleLocator{Base: 1.5}}).Ticks(0, 6, 0)
	want := []float64{0.5, 1, 2, 2.5, 3.5, 4, 5, 5.5}
	if len(ticks) != len(want) {
		t.Fatalf("AutoMinorLocator ticks = %v, want %v", ticks, want)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-12 {
			t.Fatalf("AutoMinorLocator tick %d = %v, want %v", i, ticks[i], want[i])
		}
	}
}

func TestMaxNLocator_RespectsIntervalBudget(t *testing.T) {
	ticks := (MaxNLocator{N: 4}).Ticks(0.3, 9.6, 0)
	if len(ticks) == 0 {
		t.Fatal("expected ticks from MaxNLocator")
	}
	if len(ticks) > 5 {
		t.Fatalf("MaxNLocator produced %d ticks, want <= 5: %v", len(ticks), ticks)
	}
	if ticks[0] > 0.3 || ticks[len(ticks)-1] < 9.6 {
		t.Fatalf("MaxNLocator does not cover range: %v", ticks)
	}
}

func TestAutoMinorLocator_SubdividesAutoMajors(t *testing.T) {
	ticks := (AutoMinorLocator{N: 4}).Ticks(0, 10, 5)
	if len(ticks) == 0 {
		t.Fatal("expected minor ticks from AutoMinorLocator")
	}
	majors := (AutoLocator{}).Ticks(0, 10, 5)
	for _, tick := range ticks {
		for _, major := range majors {
			if math.Abs(tick-major) < 1e-9 {
				t.Fatalf("AutoMinorLocator tick %v should not coincide with major tick %v", tick, major)
			}
		}
	}
}

func TestScalarFormatter_TrimAndScientific(t *testing.T) {
	f := ScalarFormatter{Prec: 6}
	if got := f.Format(1.0); got != "1" {
		t.Fatalf("Format(1.0)=%q", got)
	}
	if got := f.Format(1.230000); got != "1.23" {
		t.Fatalf("trim zeros: %q", got)
	}
	if got := f.Format(1234567); !strings.Contains(got, "e") {
		t.Fatalf("expected scientific for large: %q", got)
	}
	if got := f.Format(0.0000123); !strings.Contains(got, "e") {
		t.Fatalf("expected scientific for small: %q", got)
	}
}

func TestScalarFormatterPowerLimitsAndMathText(t *testing.T) {
	f := ScalarFormatter{Prec: 2, UsePowerLimits: true, PowerLimits: [2]int{-3, 4}}
	if got := f.Format(9999); got != "9999" {
		t.Fatalf("below upper power limit = %q, want %q", got, "9999")
	}
	if got := f.Format(10000); got != "1e+4" {
		t.Fatalf("at upper power limit = %q, want %q", got, "1e+4")
	}
	if got := f.Format(0.001); got != "1e−3" {
		t.Fatalf("at lower power limit = %q, want %q", got, "1e−3")
	}
	if got := (ScalarFormatter{Prec: 1, UsePowerLimits: true, PowerLimits: [2]int{0, 0}, UseMathText: true}).Format(-1200); got != `$\mathdefault{−1.2\times10^{3}}$` {
		t.Fatalf("mathtext scientific = %q", got)
	}
	if got := (ScalarFormatter{Prec: 2, DisableScientific: true}).Format(1e7); got != "10000000" {
		t.Fatalf("disabled scientific = %q, want %q", got, "10000000")
	}
}

func TestScalarFormatterLocaleIndependentPolicy(t *testing.T) {
	t.Setenv("LC_NUMERIC", "de_DE.UTF-8")

	f := ScalarFormatter{Prec: 2, DisableScientific: true}
	if got := f.Format(-1234.5); got != "−1234.5" {
		t.Fatalf("locale-independent scalar label = %q, want %q", got, "−1234.5")
	}
}

func TestFormatScalarTickLabel_UsesStepPrecision(t *testing.T) {
	f := ScalarFormatter{Prec: 3}
	cases := []struct {
		value float64
		step  float64
		want  string
	}{
		{value: 0, step: 0.025, want: "0.000"},
		{value: 0.05, step: 0.025, want: "0.050"},
		{value: 2, step: 2, want: "2"},
		{value: 0.2, step: 0.2, want: "0.2"},
	}

	for _, tc := range cases {
		if got := formatScalarTickLabel(f, tc.value, tc.step); got != tc.want {
			t.Fatalf("formatScalarTickLabel(%v, %v)=%q want %q", tc.value, tc.step, got, tc.want)
		}
	}
}

func TestFormatScalarTickLabel_PreservesMathTextForStepLabels(t *testing.T) {
	f := ScalarFormatter{UseMathText: true}
	if got, want := formatScalarTickLabel(f, 0, 1200), `$\mathdefault{0}$`; got != want {
		t.Fatalf("zero mathtext step label = %q, want %q", got, want)
	}
	if got, want := formatScalarTickLabel(f, 0.5, 0.5), `$\mathdefault{0.5}$`; got != want {
		t.Fatalf("decimal mathtext step label = %q, want %q", got, want)
	}
}

func TestFixedFormatter_UsesTickIndex(t *testing.T) {
	formatter := FixedFormatter{Labels: []string{"low", "mid", "high"}}
	ticks := []float64{1, 2, 3}
	if got := formatTickLabel(formatter, 2, 1, ticks); got != "mid" {
		t.Fatalf("FixedFormatter label = %q, want %q", got, "mid")
	}
}

func TestExtraFormatters(t *testing.T) {
	if got := (NullFormatter{}).Format(12); got != "" {
		t.Fatalf("NullFormatter = %q, want empty", got)
	}
	if got := FuncFormatter(func(v float64) string { return strings.ToUpper((ScalarFormatter{Prec: 0}).Format(v)) }).Format(12); got != "12" {
		t.Fatalf("FuncFormatter = %q, want %q", got, "12")
	}
	if got := (FormatStrFormatter{Pattern: "%.1f m"}).Format(2.25); got != "2.2 m" {
		t.Fatalf("FormatStrFormatter = %q, want %q", got, "2.2 m")
	}
	if got := (StrMethodFormatter{Template: "{x:.2f} s"}).Format(1.234); got != "1.23 s" {
		t.Fatalf("StrMethodFormatter = %q, want %q", got, "1.23 s")
	}
	if got := (EngFormatter{Unit: "Hz", Places: 1, SepSet: true}).Format(1200); got != "1.2kHz" {
		t.Fatalf("EngFormatter = %q, want %q", got, "1.2kHz")
	}
	if got := (PercentFormatter{XMax: 1, Decimals: 0}).Format(0.375); got != "38%" {
		t.Fatalf("PercentFormatter = %q, want %q", got, "38%")
	}
}

func TestEngFormatterMatplotlibStyleDefaults(t *testing.T) {
	if got := (EngFormatter{}).Format(1200); got != "1.2 k" {
		t.Fatalf("EngFormatter default separator/places = %q, want %q", got, "1.2 k")
	}
	if got := (EngFormatter{}).Format(-1e-6); got != "\u22121 \u00b5" {
		t.Fatalf("EngFormatter default micro prefix = %q, want %q", got, "\u22121 \u00b5")
	}
	if got := (EngFormatter{Unit: "Hz"}).Format(0); got != "0 Hz" {
		t.Fatalf("EngFormatter zero with unit = %q, want %q", got, "0 Hz")
	}
	if got := (EngFormatter{Unit: "Hz", Places: 1}).Format(0); got != "0.0 Hz" {
		t.Fatalf("EngFormatter zero with fixed places = %q, want %q", got, "0.0 Hz")
	}
	if got := (EngFormatter{Unit: "Hz", SepSet: true}).Format(0); got != "0Hz" {
		t.Fatalf("EngFormatter explicit empty separator at zero = %q, want %q", got, "0Hz")
	}
	if got := (EngFormatter{Places: 0, PlacesSet: true}).Format(1234); got != "1 k" {
		t.Fatalf("EngFormatter explicit zero places = %q, want %q", got, "1 k")
	}
}

func TestEngFormatterEdgeBehavior(t *testing.T) {
	if got := (EngFormatter{Places: 2, Sep: " ", UseUnicodeMicro: true}).Format(-1e-6); got != "\u22121.00 \u00b5" {
		t.Fatalf("EngFormatter unicode micro = %q, want %q", got, "\u22121.00 \u00b5")
	}
	if got := (EngFormatter{Places: 1, Sep: " "}).Format(999.95); got != "1.0 k" {
		t.Fatalf("EngFormatter rounding rollover = %q, want %q", got, "1.0 k")
	}
	if got := (EngFormatter{Places: 1, Sep: " "}).Format(1e31); got != "10.0 Q" {
		t.Fatalf("EngFormatter extreme prefix = %q, want %q", got, "10.0 Q")
	}
	if got := (EngFormatter{Places: 1, Sep: " "}).Format(0); got != "0.0" {
		t.Fatalf("EngFormatter zero without suffix = %q, want %q", got, "0.0")
	}
	if got := (EngFormatter{Places: 1, Sep: " ", Unit: "Hz", UseMathText: true}).FormatEng(-1200); got != `$\mathdefault{−1.2}$ kHz` {
		t.Fatalf("EngFormatter mathtext alias = %q", got)
	}
}

func TestPercentFormatterMatplotlibStyleDefaults(t *testing.T) {
	if got := (PercentFormatter{}).Format(50); got != "50%" {
		t.Fatalf("PercentFormatter zero value = %q, want %q", got, "50%")
	}
	if got := (PercentFormatter{XMax: 1, DisplayRange: 0.01}).Format(-0.0035); got != "\u22120.35%" {
		t.Fatalf("PercentFormatter default auto decimals = %q, want %q", got, "\u22120.35%")
	}
	if got := (PercentFormatter{XMax: 1, Decimals: 0, DecimalsSet: true, DisplayRange: 0.01}).Format(0.0035); got != "0%" {
		t.Fatalf("PercentFormatter explicit zero decimals = %q, want %q", got, "0%")
	}
}

func TestPercentFormatterAutoDecimals(t *testing.T) {
	if got := (PercentFormatter{XMax: 1, Decimals: -1, DisplayRange: 1}).Format(0.345); got != "34%" {
		t.Fatalf("PercentFormatter wide range = %q, want %q", got, "34%%")
	}
	if got := (PercentFormatter{XMax: 1, Decimals: -1, DisplayRange: 0.01}).Format(-0.0035); got != "\u22120.35%" {
		t.Fatalf("PercentFormatter narrow negative range = %q, want %q", got, "\u22120.35%")
	}
	if got := (PercentFormatter{XMax: 1, Decimals: 1, NoSymbol: true}).Format(0.125); got != "12.5" {
		t.Fatalf("PercentFormatter no symbol = %q, want %q", got, "12.5")
	}
	if got := (PercentFormatter{XMax: 1, Decimals: 0, Symbol: "%_done", UseTeX: true}).Format(0.5); got != `50\%\_done` {
		t.Fatalf("PercentFormatter TeX escaped symbol = %q", got)
	}
	if got := (PercentFormatter{XMax: 1, Decimals: 0, Symbol: `% raw`, UseTeX: true, IsLaTeX: true}).Format(0.5); got != `50% raw` {
		t.Fatalf("PercentFormatter raw LaTeX symbol = %q", got)
	}
}

func TestLogFormatterFormatsBaseTenDecadesAsPowers(t *testing.T) {
	formatter := LogFormatter{Base: 10}
	cases := []struct {
		value float64
		want  string
	}{
		{value: 1000, want: "10³"},
		{value: 1, want: "10⁰"},
		{value: 0.1, want: "10⁻¹"},
	}

	for _, tc := range cases {
		if got := formatter.Format(tc.value); got != tc.want {
			t.Fatalf("LogFormatter.Format(%v) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestLogFormatterExponent(t *testing.T) {
	formatter := LogFormatterExponent{Base: 10}
	cases := []struct {
		value float64
		want  string
	}{
		{value: 1000, want: "3"},
		{value: 0.1, want: "−1"},
		{value: 0, want: "0"},
	}
	for _, tc := range cases {
		if got := formatter.Format(tc.value); got != tc.want {
			t.Fatalf("LogFormatterExponent.Format(%v) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestLogFormatterMathText(t *testing.T) {
	formatter := LogFormatterMathText{Base: 10}
	if got, want := formatter.Format(1000), `$\mathdefault{10^{3}}$`; got != want {
		t.Fatalf("LogFormatterMathText.Format(1000) = %q, want %q", got, want)
	}
	sci := LogFormatterMathText{Base: 10, SciNotation: true}
	if got, want := sci.Format(20), `$\mathdefault{2\times10^{1}}$`; got != want {
		t.Fatalf("LogFormatterMathText scientific Format(20) = %q, want %q", got, want)
	}
}

func TestLogFormattersMinorThresholds(t *testing.T) {
	denseTicks := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	dense := LogFormatterMathText{
		Base:              10,
		SciNotation:       true,
		UseMinorThreshold: true,
		MinorThresholds:   [2]float64{1, 0.4},
	}
	if got := dense.FormatTick(2, 1, denseTicks); got != "" {
		t.Fatalf("dense minor label = %q, want empty", got)
	}
	if got, want := dense.FormatTick(10, 9, denseTicks), `$\mathdefault{10^{1}}$`; got != want {
		t.Fatalf("dense decade label = %q, want %q", got, want)
	}

	sparseTicks := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sparse := LogFormatter{Base: 10, UseMinorThreshold: true, MinorThresholds: [2]float64{2, 0.4}}
	if got := sparse.FormatTick(5, 4, sparseTicks); got != "" {
		t.Fatalf("sparse omitted minor label = %q, want empty", got)
	}
	if got, want := sparse.FormatTick(6, 5, sparseTicks), "6"; got != want {
		t.Fatalf("sparse selected minor label = %q, want %q", got, want)
	}
	if got := (LogFormatterExponent{Base: 10, LabelOnlyBase: true}).FormatTick(20, 0, []float64{10, 20, 100}); got != "" {
		t.Fatalf("label-only-base exponent formatter = %q, want empty", got)
	}
}

func TestDefaultLogScaleFormatterUsesMatplotlibMinorThresholds(t *testing.T) {
	axis := &Axis{}
	configureScaleAxis(axis, "log", transform.ScaleOptions{Base: 10})

	formatter, ok := axis.Formatter.(LogFormatterMathText)
	if !ok {
		t.Fatalf("log scale formatter = %T, want LogFormatterMathText", axis.Formatter)
	}
	ticks := []float64{200, 300, 500, 700, 850, 1000}
	want := []string{
		`$\mathdefault{2\times10^{2}}$`,
		`$\mathdefault{3\times10^{2}}$`,
		"",
		"",
		"",
		`$\mathdefault{10^{3}}$`,
	}
	for i, tick := range ticks {
		if got := formatter.FormatTick(tick, i, ticks); got != want[i] {
			t.Fatalf("default log formatter tick %v = %q, want %q", tick, got, want[i])
		}
	}
}

func TestLogitFormatter(t *testing.T) {
	formatter := LogitFormatter{}
	cases := []struct {
		value float64
		want  string
	}{
		{value: 0.001, want: `$\mathdefault{10^{-3}}$`},
		{value: 0.5, want: `$\mathdefault{\frac{1}{2}}$`},
		{value: 0.9, want: `$\mathdefault{1-10^{-1}}$`},
		{value: 0.25, want: `$\mathdefault{0.25}$`},
	}

	for _, tc := range cases {
		if got := formatter.Format(tc.value); got != tc.want {
			t.Fatalf("LogitFormatter.Format(%v) = %q, want %q", tc.value, got, tc.want)
		}
	}
	if got := (LogitFormatter{Minor: true}).Format(0.2); got != "" {
		t.Fatalf("minor LogitFormatter.Format = %q, want empty", got)
	}
}
