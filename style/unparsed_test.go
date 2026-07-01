package style

import (
	"slices"
	"strings"
	"testing"
)

func TestUnparsedKnownRCParamWarns(t *testing.T) {
	msgs := captureWarnings(t)

	// axes.labelweight is a real matplotlib 3.10.9 rcParam that matplotlib-go
	// does not parse; it must warn instead of vanishing silently.
	_, report, err := ParseMPLStyle("t.mplstyle", "axes.labelweight: bold\n")
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}

	if len(*msgs) != 1 {
		t.Fatalf("got %d warnings, want 1: %v", len(*msgs), *msgs)
	}
	if !strings.Contains((*msgs)[0], "axes.labelweight") {
		t.Errorf("warning %q does not mention the rcParam key", (*msgs)[0])
	}
	if !strings.Contains((*msgs)[0], "not parsed") {
		t.Errorf("warning %q does not explain the param is unparsed", (*msgs)[0])
	}
	// The key still lands in the report so callers can inspect it.
	if len(report.Unsupported) != 1 || report.Unsupported[0].Key != "axes.labelweight" {
		t.Errorf("report.Unsupported = %v, want axes.labelweight", report.Unsupported)
	}
}

func TestUnparsedUnknownRCParamStaysSilent(t *testing.T) {
	msgs := captureWarnings(t)

	// A key matplotlib itself does not know (typo, plugin key) must not warn;
	// silence means "genuinely unknown key".
	_, report, err := ParseMPLStyle("t.mplstyle", "totally.made.up: 1\n")
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}

	if len(*msgs) != 0 {
		t.Fatalf("unknown key warned unexpectedly: %v", *msgs)
	}
	if len(report.Unsupported) != 1 || report.Unsupported[0].Key != "totally.made.up" {
		t.Errorf("report.Unsupported = %v, want totally.made.up", report.Unsupported)
	}
}

func TestUnparsedKnownRCParamWarnsOncePerKey(t *testing.T) {
	msgs := captureWarnings(t)

	for range 3 {
		if _, _, err := ParseMPLStyle("t.mplstyle", "axes.labelpad: 5\n"); err != nil {
			t.Fatalf("ParseMPLStyle() error = %v", err)
		}
	}

	if len(*msgs) != 1 {
		t.Fatalf("got %d warnings across 3 applies, want 1 (deduped): %v", len(*msgs), *msgs)
	}
}

func TestNonGoalRCParamWarnsWithRationale(t *testing.T) {
	msgs := captureWarnings(t)

	if _, _, err := ParseMPLStyle("t.mplstyle", "path.snap: False\n"); err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}

	if len(*msgs) != 1 {
		t.Fatalf("got %d warnings, want 1: %v", len(*msgs), *msgs)
	}
	if !strings.Contains((*msgs)[0], "intentionally not supported") {
		t.Errorf("warning %q does not flag the param as an explicit non-goal", (*msgs)[0])
	}
	if !strings.Contains((*msgs)[0], nonGoalRCParams["path.snap"]) {
		t.Errorf("warning %q does not carry the registered rationale", (*msgs)[0])
	}
}

// TestNonGoalRCParamsAreKnownUpstream guards that every documented non-goal is
// a real matplotlib rcParam (a typo'd key would silently never fire) and that
// none of them is parsed (a parsed key never reaches the fallthrough lookup).
func TestNonGoalRCParamsAreKnownUpstream(t *testing.T) {
	supported := SupportedMPLStyleKeys()
	for key := range nonGoalRCParams {
		if _, ok := knownUpstreamRCParams[key]; !ok {
			t.Errorf("non-goal key %q is not in knownUpstreamRCParams (typo?)", key)
		}
		if slices.Contains(supported, key) {
			t.Errorf("non-goal key %q is parsed by applyMPLStyleEntry; it would never reach the fallthrough warning", key)
		}
	}
}

// TestKnownUpstreamRCParamsConsistency pins the generated table against the
// keys matplotlib-go already handles: every parsed key except the documented
// Go-only grid.major/minor extensions must be a known upstream key, and every
// unhonored (parsed-but-ignored) key too. A violation means either a typo in a
// key list or that the table was regenerated against the wrong matplotlib.
func TestKnownUpstreamRCParamsConsistency(t *testing.T) {
	// matplotlib 3.10.9 ships 322 rcParams; the table drops the internal
	// _internal.classic_mode key.
	if got, want := len(knownUpstreamRCParams), 321; got != want {
		t.Errorf("len(knownUpstreamRCParams) = %d, want %d (regenerated against a different matplotlib?)", got, want)
	}

	goOnlyKeys := []string{
		"grid.major.color", "grid.major.linestyle",
		"grid.minor.color", "grid.minor.linestyle",
	}
	for _, key := range SupportedMPLStyleKeys() {
		if slices.Contains(goOnlyKeys, key) {
			continue
		}
		if _, ok := knownUpstreamRCParams[key]; !ok {
			t.Errorf("parsed key %q is not in knownUpstreamRCParams (typo or Go-only key missing from goOnlyKeys?)", key)
		}
	}
	for _, key := range unhonoredRCParamKeys() {
		if _, ok := knownUpstreamRCParams[key]; !ok {
			t.Errorf("unhonored key %q is not in knownUpstreamRCParams", key)
		}
	}
	for _, key := range goOnlyKeys {
		if _, ok := knownUpstreamRCParams[key]; ok {
			t.Errorf("Go-only key %q unexpectedly appeared in knownUpstreamRCParams; drop it from goOnlyKeys", key)
		}
	}
}
