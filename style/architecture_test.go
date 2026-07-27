package style

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSupportedMPLStyleKeysReturnSortedCopy(t *testing.T) {
	keys := SupportedMPLStyleKeys()
	if len(keys) == 0 {
		t.Fatal("supported key list is empty")
	}
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Fatalf("keys are not sorted at %d: %q > %q", i, keys[i-1], keys[i])
		}
	}

	keys[0] = "mutated"
	fresh := SupportedMPLStyleKeys()
	if fresh[0] == "mutated" {
		t.Fatal("SupportedMPLStyleKeys returned mutable backing storage")
	}
}

func TestMPLStyleParamsApplySupportedKeysAndReportUnsupported(t *testing.T) {
	params := Params{
		"figure.dpi":              "144",
		"path.simplify":           "False",
		"path.simplify_threshold": "0.2",
		"agg.path.chunksize":      "8192",
		"unsupported.option":      "value",
	}
	rc, report, err := applyMPLStyleParams(Default, params)
	if err != nil {
		t.Fatal(err)
	}
	if rc.DPI != 144 {
		t.Fatalf("DPI = %v, want 144", rc.DPI)
	}
	if rc.PathSimplify {
		t.Fatal("path.simplify should parse false")
	}
	if rc.PathSimplifyThreshold != 0.2 {
		t.Fatalf("PathSimplifyThreshold = %v, want 0.2", rc.PathSimplifyThreshold)
	}
	if rc.AggPathChunkSize != 8192 {
		t.Fatalf("AggPathChunkSize = %d, want 8192", rc.AggPathChunkSize)
	}
	if len(report.Applied) != 4 {
		t.Fatalf("applied report = %+v", report.Applied)
	}
	if len(report.Unsupported) != 1 || report.Unsupported[0].Key != "unsupported.option" {
		t.Fatalf("unsupported report = %+v", report.Unsupported)
	}
}

func TestSupportedMPLStyleKeysAreAuditedAgainstUpstream(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "mpl-data", "matplotlibrc"))
	if err != nil {
		t.Fatalf("read upstream matplotlibrc: %v", err)
	}
	upstream := upstreamMatplotlibRCKeys(string(data))
	if len(upstream) == 0 {
		t.Fatal("upstream matplotlibrc key inventory is empty")
	}

	for _, key := range SupportedMPLStyleKeys() {
		upstreamKey := supportedMPLStyleKeyUpstreamEquivalent(key)
		if !upstream[upstreamKey] {
			t.Fatalf("supported style key %q maps to upstream key %q, which is not present in upstream matplotlibrc", key, upstreamKey)
		}
	}
	if len(upstream) <= len(SupportedMPLStyleKeys()) {
		t.Fatalf("upstream rcParams inventory = %d keys, supported = %d; expected an audited subset",
			len(upstream), len(SupportedMPLStyleKeys()))
	}

	doc, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(doc)), " ")
	for _, phrase := range []string{
		"rcParams Key Audit",
		"`style.SupportedMPLStyleKeys()` is the supported rcParams subset",
		"Unsupported rcParams are reported through `MPLStyleReport.Unsupported` and ignored",
		"unsupported keys are intentional typed-API divergences unless a fixture needs them",
	} {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("migration notes missing rcParams audit phrase %q", phrase)
		}
	}
}

func upstreamMatplotlibRCKeys(src string) map[string]bool {
	keys := map[string]bool{}
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "#")
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, ":") {
			continue
		}
		key := strings.TrimSpace(strings.SplitN(line, ":", 2)[0])
		if key == "" || strings.ContainsAny(key, " \t") {
			continue
		}
		keys[key] = true
	}
	return keys
}

func supportedMPLStyleKeyUpstreamEquivalent(key string) string {
	switch key {
	case "grid.major.color", "grid.minor.color":
		return "grid.color"
	case "grid.major.linestyle", "grid.minor.linestyle":
		return "grid.linestyle"
	default:
		return key
	}
}
