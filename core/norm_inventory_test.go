package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/transform"
)

type normInventoryEntry struct {
	upstream          string
	upstreamGenerated bool
	goNorm            ScalarNormalizer
	scaleNames        []string
	colorbarRoute     string
	omission          string
}

func TestNormInventoryMatchesMatplotlibColorsSurface(t *testing.T) {
	src := readUpstreamColorsPy(t)
	scaleNames := map[string]bool{}
	for _, name := range transform.ScaleNames() {
		scaleNames[name] = true
	}

	seen := map[string]bool{}
	for _, entry := range normInventory {
		if entry.upstream == "" {
			t.Fatal("norm inventory entry has empty upstream name")
		}
		if seen[entry.upstream] {
			t.Fatalf("duplicate norm inventory entry %q", entry.upstream)
		}
		seen[entry.upstream] = true

		if entry.upstreamGenerated {
			if !strings.Contains(src, entry.upstream+" = make_norm_from_scale(") {
				t.Fatalf("upstream generated norm %s not found in colors.py", entry.upstream)
			}
		} else if !strings.Contains(src, "class "+entry.upstream) {
			t.Fatalf("upstream norm class %s not found in colors.py", entry.upstream)
		}

		for _, scaleName := range entry.scaleNames {
			if !scaleNames[scaleName] {
				t.Fatalf("%s references missing scale registry name %q", entry.upstream, scaleName)
			}
		}
		if entry.colorbarRoute == "" {
			t.Fatalf("%s missing colorbar interaction inventory", entry.upstream)
		}
		if entry.goNorm == nil {
			if entry.omission == "" {
				t.Fatalf("%s has no Go norm and no omission note", entry.upstream)
			}
			continue
		}
		if entry.omission != "" {
			t.Fatalf("%s has both Go norm and omission note", entry.upstream)
		}
		if entry.goNorm.NormName() == "" {
			t.Fatalf("%s Go norm reports empty NormName", entry.upstream)
		}
		if err := entry.goNorm.Validate(); err != nil {
			t.Fatalf("%s Go norm Validate() = %v", entry.upstream, err)
		}
	}

	for _, required := range []string{
		"Normalize",
		"LogNorm",
		"TwoSlopeNorm",
		"CenteredNorm",
		"FuncNorm",
		"SymLogNorm",
		"AsinhNorm",
		"PowerNorm",
		"BoundaryNorm",
		"NoNorm",
	} {
		if !seen[required] {
			t.Fatalf("norm inventory missing %s", required)
		}
	}
}

func TestNormGapListDocumentsSupportedFixtureNeeds(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	required := []string{
		"Norm Gap List",
		"`lognorm_imshow`",
		"`twoslope_norm_image`",
		"`boundarynorm_pcolormesh`",
		"`asinh_norm_image`",
		"No supported parity fixture currently requires `FuncNorm`",
		"`MultiNorm`",
		"scalar-mappable callback",
		"clip",
	}
	for _, phrase := range required {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("norm gap list missing %q", phrase)
		}
	}
}

func TestFuncNormUpstreamContractIsDocumented(t *testing.T) {
	src := readUpstreamColorsPy(t)
	upstreamRequired := []string{
		"def make_norm_from_scale(scale_cls, base_norm_cls=None",
		"scale_cls, *scale_args",
		"self._trf = self._scale.get_transform()",
		"def __call__(self, value, clip=None):",
		"if clip:",
		"def inverse(self, value):",
		"def autoscale_None(self, A):",
		"@make_norm_from_scale(\n    scale.FuncScale",
		"functions : (callable, callable)",
		"The forward function must be monotonic.",
		"Both functions must have the signature",
	}
	for _, phrase := range upstreamRequired {
		if !strings.Contains(src, phrase) {
			t.Fatalf("upstream FuncNorm contract missing %q", phrase)
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	docRequired := []string{
		"FuncNorm Upstream Contract",
		"`make_norm_from_scale`",
		"`scale.FuncScale`",
		"forward and inverse",
		"monotonic",
		"`clip`",
		"`autoscale_None`",
		"transform-domain finite values",
	}
	for _, phrase := range docRequired {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("FuncNorm contract docs missing %q", phrase)
		}
	}
}

func TestFuncNormGoCallbackShapeDecisionIsDocumented(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	required := []string{
		"FuncNorm Go Callback Shape",
		"`ScalarNormalizer`",
		"`Map(float64) float64`",
		"`Inverse(float64) (float64, bool)`",
		"`Autoscale([]float64) ScalarNormalizer`",
		"`Range() (float64, float64)`",
		"`Validate() error`",
		"`NormName() string`",
		"does not share the axis `transform.Scale` callback shape",
		"a concrete `core.FuncNorm` type",
	}
	for _, phrase := range required {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("FuncNorm callback-shape docs missing %q", phrase)
		}
	}
}

func TestFuncNormOmissionLedgerIsDocumented(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	required := []string{
		"FuncNorm Omission Ledger",
		"intentional omission",
		"no supported parity fixture",
		"`asinh_norm_image`",
		"`twoslope_norm_image`",
		"`lognorm_imshow`",
		"`boundarynorm_pcolormesh`",
		"implement `ScalarNormalizer` directly",
		"`transform.Scale`",
	}
	for _, phrase := range required {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("FuncNorm omission ledger missing %q", phrase)
		}
	}
}

func TestNormMetadataDocsAndStatusAreCurrent(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	required := []string{
		"Norm Public Surface Metadata",
		"`idiomatic-equivalent`",
		"`make_norm_from_scale` remains marked `intentional-omission`",
		"`LogNorm` remains covered by the",
		"norm inventory and `lognorm_imshow`",
	}
	for _, phrase := range required {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("norm metadata docs missing %q", phrase)
		}
	}
	for _, name := range []string{"Normalize", "SymLogNorm", "PowerNorm", "TwoSlopeNorm", "CenteredNorm", "BoundaryNorm", "NoNorm", "AsinhNorm"} {
		if !strings.Contains(doc, "`"+name+"`") {
			t.Fatalf("norm metadata docs missing implemented norm %q", name)
		}
	}

	statusData, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-parity-status.md"))
	if err != nil {
		t.Fatalf("read parity status doc: %v", err)
	}
	status := string(statusData)
	if strings.Contains(status, "| colors.py:class:Normalize | colors-cm | partial |") {
		t.Fatal("parity status still reports Normalize through the broad partial colors row")
	}
	if !strings.Contains(status, "norm classes and the dynamic norm factory have explicit rows") {
		t.Fatal("parity status broad colors note does not point to explicit norm rows")
	}
}

func TestScalarMappableNormUpdateAuditIsDocumented(t *testing.T) {
	colorizer := readUpstreamMatplotlibFile(t, "colorizer.py")
	colorbar := readUpstreamMatplotlibFile(t, "colorbar.py")
	upstreamRequired := []struct {
		name string
		src  string
	}{
		{"colorizer changed registry", "self.callbacks = cbook.CallbackRegistry(signals=[\"changed\"])"},
		{"norm callback connection", "self._id_norm = self.norm.callbacks.connect('changed'"},
		{"set clim", "def set_clim(self, vmin=None, vmax=None):"},
		{"blocked norm callbacks", "with self.norm.callbacks.blocked(signal='changed')"},
		{"colorizer changed", "def changed(self):"},
	}
	for _, required := range upstreamRequired {
		if !strings.Contains(colorizer, required.src) {
			t.Fatalf("upstream Colorizer audit missing %s", required.name)
		}
	}
	for _, phrase := range []string{
		"mappable.callbacks.connect(",
		"def update_normal(self, mappable=None):",
		"self._reset_locator_formatter_scale()",
	} {
		if !strings.Contains(colorbar, phrase) {
			t.Fatalf("upstream Colorbar audit missing %q", phrase)
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	for _, phrase := range []string{
		"Scalar-Mappable Norm Update Audit",
		"Matplotlib `Colorizer` connects norm callbacks",
		"`set_clim`",
		"blocks norm callbacks",
		"`Colorbar.update_normal`",
		"Go uses pull-based",
		"`SetNorm`",
		"`SetCLim`",
		"`syncColorbarMapping`",
		"no callback registry",
	} {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("scalar-mappable norm update audit missing %q", phrase)
		}
	}
}

func TestScalarMappableNormUpdateDecisionIsDocumented(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	for _, phrase := range []string{
		"Scalar-Mappable Norm Update Decision",
		"supported Go update path",
		"`SetArray`",
		"`SetColormap`",
		"`SetNorm`",
		"`SetCLim`",
		"redraw",
		"colorbar pulls",
		"Matplotlib-style callback registry remains intentionally omitted",
	} {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("scalar-mappable norm update decision missing %q", phrase)
		}
	}
}

var normInventory = []normInventoryEntry{
	{
		upstream:      "Normalize",
		goNorm:        Normalize{VMin: 0, VMax: 1},
		scaleNames:    []string{"linear"},
		colorbarRoute: "default scalar-map norm and linear colorbar axis",
	},
	{
		upstream:          "LogNorm",
		upstreamGenerated: true,
		goNorm:            LogNorm{VMin: 1, VMax: 10},
		scaleNames:        []string{"log"},
		colorbarRoute:     "log colorbar scale and log tick locator",
	},
	{
		upstream:      "TwoSlopeNorm",
		goNorm:        TwoSlopeNorm{VMin: -1, VCenter: 0, VMax: 2},
		colorbarRoute: "function colorbar scale via norm inverse",
	},
	{
		upstream:      "CenteredNorm",
		goNorm:        CenteredNorm{VCenter: 0, HalfRange: 1},
		colorbarRoute: "function colorbar scale via norm inverse",
	},
	{
		upstream:      "FuncNorm",
		goNorm:        FuncNorm{Forward: func(v float64) float64 { return v }, Reverse: func(v float64) float64 { return v }, VMin: 0, VMax: 1},
		scaleNames:    []string{"function", "functionlog"},
		colorbarRoute: "function colorbar scale via norm inverse",
	},
	{
		upstream:      "SymLogNorm",
		goNorm:        SymLogNorm{VMin: -10, VMax: 10, LinThresh: 1, LinScale: 1, Base: 10},
		scaleNames:    []string{"symlog"},
		colorbarRoute: "function colorbar scale via norm inverse",
	},
	{
		upstream:      "AsinhNorm",
		goNorm:        AsinhNorm{LinearWidth: 1, VMin: -10, VMax: 10},
		scaleNames:    []string{"asinh"},
		colorbarRoute: "asinh colorbar scale with linear-width metadata",
	},
	{
		upstream:      "PowerNorm",
		goNorm:        PowerNorm{Gamma: 2, VMin: 0, VMax: 1},
		colorbarRoute: "function colorbar scale via norm inverse",
	},
	{
		upstream:      "BoundaryNorm",
		goNorm:        BoundaryNorm{Boundaries: []float64{0, 1, 2}, NColors: 2},
		colorbarRoute: "boundary colorbar ticks, boundaries, values, and extension handling",
	},
	{
		upstream:      "NoNorm",
		goNorm:        NoNorm{},
		colorbarRoute: "index-style scalar map with linear colorbar axis",
	},
}

func readUpstreamColorsPy(t *testing.T) string {
	t.Helper()
	return readUpstreamMatplotlibFile(t, "colors.py")
}

func readUpstreamMatplotlibFile(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", name))
	if err != nil {
		t.Fatalf("read upstream %s: %v", name, err)
	}
	return string(data)
}
