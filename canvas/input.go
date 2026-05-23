package canvas

import "strings"

// ModifierSet composes a backend-neutral modifier bitfield from raw backend
// booleans.
func ModifierSet(shift, control, alt, meta bool) Modifier {
	var out Modifier
	if shift {
		out |= ModifierShift
	}
	if control {
		out |= ModifierControl
	}
	if alt {
		out |= ModifierAlt
	}
	if meta {
		out |= ModifierMeta
	}
	return out
}

// ModifierFromName maps common backend modifier names onto the canvas bitfield.
func ModifierFromName(name string) Modifier {
	switch strings.ToLower(name) {
	case "shift":
		return ModifierShift
	case "control", "ctrl":
		return ModifierControl
	case "alt":
		return ModifierAlt
	case "meta", "super", "cmd", "command":
		return ModifierMeta
	}
	return 0
}

// ModifiersFromNames maps backend modifier-name lists onto the canvas bitfield.
func ModifiersFromNames(names []string) Modifier {
	var out Modifier
	for _, name := range names {
		out |= ModifierFromName(name)
	}
	return out
}

// MouseButtonFromJSIndex maps browser MouseEvent.button values onto canvas
// buttons. Unknown indices map to MouseButtonNone instead of pretending to be
// left clicks.
func MouseButtonFromJSIndex(index int) MouseButton {
	switch index {
	case 0:
		return MouseButtonLeft
	case 1:
		return MouseButtonMiddle
	case 2:
		return MouseButtonRight
	}
	return MouseButtonNone
}

// NormalizeKey strips modifier prefixes that are represented separately in
// Event.Modifiers.
func NormalizeKey(key string) string {
	for {
		next := key
		for _, prefix := range []string{"ctrl+", "control+", "shift+", "alt+", "meta+", "super+", "cmd+", "command+"} {
			next = strings.TrimPrefix(next, prefix)
		}
		if next == key {
			return key
		}
		key = next
	}
}
