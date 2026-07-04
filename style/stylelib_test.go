package style

import (
	"strings"
	"testing"
)

func TestBundledStylesRegistered(t *testing.T) {
	// First color of each sheet's axes.prop_cycle, from Matplotlib 3.10.9.
	cases := map[string]string{
		"fivethirtyeight":      "008fd5",
		"bmh":                  "348ABD",
		"solarize_light2":      "268BD2",
		"seaborn-v0_8":         "4C72B0",
		"seaborn-v0_8-deep":    "4C72B0",
		"seaborn-v0_8-bright":  "003FFF",
		"tableau-colorblind10": "006BA4",
	}
	for name, wantHex := range cases {
		theme, ok := GetTheme(name)
		if !ok {
			t.Errorf("theme %q not registered", name)
			continue
		}
		palette := theme.RC.Palette()
		if len(palette) == 0 {
			t.Errorf("theme %q has empty palette", name)
			continue
		}
		if got, want := palette[0], mustParseTestColor(t, wantHex); got != want {
			t.Errorf("theme %q palette[0] = %+v, want %+v", name, got, want)
		}
	}
}

func TestBundledStylesDoNotOverrideBuiltins(t *testing.T) {
	// ggplot and dark_background are hand-tuned built-ins; the embedded sheets of
	// the same name must not clobber them (golden images depend on the built-ins).
	for _, name := range []string{"default", "ggplot", "dark_background", "publication"} {
		for _, bundled := range BundledStyleNames() {
			if bundled == name {
				t.Errorf("built-in theme %q was overridden by an embedded sheet", name)
			}
		}
	}
}

func TestRegisterBundledStylesEmitsNoWarnings(t *testing.T) {
	// The bundled sheets are a fixed, known set; parsing them at init must not
	// spam consumers' stderr with unparsed/unhonored rcParam warnings.
	// registerBundledStyles is idempotent (already-registered themes are
	// skipped before registration but still parsed), so re-running it here
	// exercises the same parse path as package init.
	msgs := captureWarnings(t)

	registerBundledStyles()

	if len(*msgs) != 0 {
		t.Fatalf("got %d warnings from bundled style registration, want 0: %v", len(*msgs), *msgs)
	}
}

func TestUserSheetsStillWarnAfterBundledRegistration(t *testing.T) {
	// The suppressed init-time parse must not consume the one-shot warning
	// dedup: a user-supplied sheet setting a key that also appears in the
	// bundled sheets still warns.
	msgs := captureWarnings(t)

	registerBundledStyles()
	if _, _, err := ParseMPLStyle("user.mplstyle", "patch.antialiased: False\n"); err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}

	if len(*msgs) != 1 {
		t.Fatalf("got %d warnings for user sheet, want 1: %v", len(*msgs), *msgs)
	}
	if !strings.Contains((*msgs)[0], "patch.antialiased") {
		t.Errorf("warning %q does not mention the rcParam key", (*msgs)[0])
	}
}

func TestBundledStylesExcludePrivateSheets(t *testing.T) {
	for _, name := range BundledStyleNames() {
		if name != "" && name[0] == '_' {
			t.Errorf("private sheet %q should not be registered", name)
		}
	}
}
