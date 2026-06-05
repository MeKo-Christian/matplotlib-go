package canvas

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/core"
)

func (w *WidgetInteraction) handleCheckKey(checks *core.CheckButtons, focusedIndex int, ev KeyEvent, key string) bool {
	if checks == nil || len(checks.Labels) == 0 {
		return false
	}
	if !checks.Enabled {
		return false
	}
	command := strings.ToLower(key)
	focusedIndex = clampInt(focusedIndex, 0, len(checks.Labels)-1)
	switch command {
	case "up", "left":
		if focusedIndex > 0 {
			focusedIndex--
		} else {
			focusedIndex = len(checks.Labels) - 1
		}
		w.focusedCheckIndex = focusedIndex
		return true
	case "down", "right":
		if focusedIndex < len(checks.Labels)-1 {
			focusedIndex++
		} else {
			focusedIndex = 0
		}
		w.focusedCheckIndex = focusedIndex
		return true
	case "space", "enter":
		_ = ev
		if focusedIndex < 0 {
			focusedIndex = 0
		}
		before := checks.Values[focusedIndex]
		checks.Toggle(focusedIndex)
		w.focusedCheckIndex = focusedIndex
		return checks.Values[focusedIndex] != before
	}
	return false
}

func (w *WidgetInteraction) handleRadioKey(radios *core.RadioButtons, focusedIndex int, key string) bool {
	if radios == nil || len(radios.Labels) == 0 {
		return false
	}
	if !radios.Enabled {
		return false
	}
	focusedIndex = clampInt(focusedIndex, 0, len(radios.Labels)-1)

	switch strings.ToLower(key) {
	case "up", "left":
		if focusedIndex > 0 {
			focusedIndex--
		} else {
			focusedIndex = len(radios.Labels) - 1
		}
		w.focusedRadioIndex = focusedIndex
		radios.SetActive(focusedIndex)
		return true
	case "down", "right":
		if focusedIndex < len(radios.Labels)-1 {
			focusedIndex++
		} else {
			focusedIndex = 0
		}
		w.focusedRadioIndex = focusedIndex
		radios.SetActive(focusedIndex)
		return true
	case "space", "enter":
		w.focusedRadioIndex = focusedIndex
		radios.SetActive(focusedIndex)
		return true
	}
	return false
}
