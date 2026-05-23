package canvas

import "testing"

func TestNormalizeKeyStripsKnownModifierPrefixes(t *testing.T) {
	cases := map[string]string{
		"a":              "a",
		"ctrl+a":         "a",
		"shift+x":        "x",
		"alt+meta+z":     "z",
		"super+cmd+home": "home",
		"unknown+key":    "unknown+key",
	}
	for in, want := range cases {
		if got := NormalizeKey(in); got != want {
			t.Fatalf("NormalizeKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestModifiersFromNamesNormalizesAliases(t *testing.T) {
	got := ModifiersFromNames([]string{"shift", "CONTROL", "cmd", "ignored"})
	want := ModifierShift | ModifierControl | ModifierMeta
	if got != want {
		t.Fatalf("ModifiersFromNames = %v, want %v", got, want)
	}
}

func TestModifierSetComposesBackendFlags(t *testing.T) {
	got := ModifierSet(true, false, true, true)
	want := ModifierShift | ModifierAlt | ModifierMeta
	if got != want {
		t.Fatalf("ModifierSet = %v, want %v", got, want)
	}
}

func TestMouseButtonFromJSIndex(t *testing.T) {
	cases := []struct {
		in   int
		want MouseButton
	}{
		{0, MouseButtonLeft},
		{1, MouseButtonMiddle},
		{2, MouseButtonRight},
		{99, MouseButtonNone},
	}
	for _, c := range cases {
		if got := MouseButtonFromJSIndex(c.in); got != c.want {
			t.Fatalf("MouseButtonFromJSIndex(%d) = %v, want %v", c.in, got, c.want)
		}
	}
}
