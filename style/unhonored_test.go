package style

import (
	"slices"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/diag"
)

// captureWarnings installs a diag handler that records messages and resets the
// one-shot unhonored-param dedup so each test starts from a clean slate.
func captureWarnings(t *testing.T) *[]string {
	t.Helper()
	var msgs []string
	restore := diag.SetHandler(func(m string) { msgs = append(msgs, m) })
	resetUnhonoredRCParamWarnings()
	t.Cleanup(func() {
		restore()
		resetUnhonoredRCParamWarnings()
	})
	return &msgs
}

func TestUnhonoredRCParamWarnsOnNonDefault(t *testing.T) {
	msgs := captureWarnings(t)

	// date.epoch is store-only; set it to a non-default value.
	if _, _, err := ParseMPLStyle("t.mplstyle", "date.epoch: 0000-12-31T00:00:00\n"); err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}

	if len(*msgs) != 1 {
		t.Fatalf("got %d warnings, want 1: %v", len(*msgs), *msgs)
	}
	if !strings.Contains((*msgs)[0], "date.epoch") {
		t.Errorf("warning %q does not mention the rcParam key", (*msgs)[0])
	}
	if !strings.Contains((*msgs)[0], "not honored") {
		t.Errorf("warning %q does not explain the param is unhonored", (*msgs)[0])
	}
}

func TestUnhonoredRCParamSilentOnDefaultValue(t *testing.T) {
	msgs := captureWarnings(t)

	// Setting the param to its library default is a no-op-as-intended; do not warn.
	src := "date.epoch: " + Default.Date.Epoch + "\n"
	if _, _, err := ParseMPLStyle("t.mplstyle", src); err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}

	if len(*msgs) != 0 {
		t.Fatalf("got %d warnings, want 0: %v", len(*msgs), *msgs)
	}
}

func TestUnhonoredRCParamWarnsOncePerKey(t *testing.T) {
	msgs := captureWarnings(t)

	for range 3 {
		if _, _, err := ParseMPLStyle("t.mplstyle", "svg.id: my-id\n"); err != nil {
			t.Fatalf("ParseMPLStyle() error = %v", err)
		}
	}

	if len(*msgs) != 1 {
		t.Fatalf("got %d warnings across 3 applies, want 1 (deduped): %v", len(*msgs), *msgs)
	}
}

func TestUnhonoredRCParamWarnsViaUpdateParams(t *testing.T) {
	msgs := captureWarnings(t)
	t.Cleanup(ResetDefaults)

	// The runtime rcParam path funnels through applyMPLStyleEntry too.
	if _, err := UpdateParams(Params{"animation.writer": "imagemagick"}); err != nil {
		t.Fatalf("UpdateParams() error = %v", err)
	}

	if len(*msgs) != 1 || !strings.Contains((*msgs)[0], "animation.writer") {
		t.Fatalf("want one warning mentioning animation.writer, got %v", *msgs)
	}
}

func TestHonoredRCParamsDoNotWarn(t *testing.T) {
	msgs := captureWarnings(t)

	// Consumed params must never appear in the unhonored registry; setting them
	// to non-default values must stay silent.
	src := strings.Join([]string{
		"image.cmap: plasma",
		"mathtext.fontset: cm",
		"boxplot.notch: true",
		"boxplot.patchartist: true",
	}, "\n") + "\n"
	if _, _, err := ParseMPLStyle("t.mplstyle", src); err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}

	if len(*msgs) != 0 {
		t.Fatalf("honored params warned unexpectedly: %v", *msgs)
	}
}

// TestUnhonoredRCParamsAreParseable guards that every registry key is actually a
// recognized rcParam (present in supportedMPLStyleKeys). A typo'd key would
// otherwise silently never fire.
func TestUnhonoredRCParamsAreParseable(t *testing.T) {
	supported := SupportedMPLStyleKeys()
	for _, key := range unhonoredRCParamKeys() {
		if !slices.Contains(supported, key) {
			t.Errorf("unhonored key %q is not in supportedMPLStyleKeys (typo or removed param?)", key)
		}
	}
}

// TestHonoredRCParamsNotInRegistry guards the audit: params that ARE consumed by
// drawing/backend code must not be listed as unhonored.
func TestHonoredRCParamsNotInRegistry(t *testing.T) {
	honored := []string{
		"image.cmap", "image.interpolation",
		"mathtext.fontset",
		"boxplot.notch", "boxplot.patchartist",
		"boxplot.showmeans", "boxplot.showcaps", "boxplot.showbox", "boxplot.showfliers",
	}
	for _, key := range honored {
		if _, ok := unhonoredRCParams[key]; ok {
			t.Errorf("honored rcParam %q must not be in the unhonored registry", key)
		}
	}
}
