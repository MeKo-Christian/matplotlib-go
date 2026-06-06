package canvas

import (
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

var (
	clipboardMu   sync.Mutex
	clipboardText string
)

func (w *WidgetInteraction) handleTextKey(tb *core.TextBox, ev KeyEvent, key string) bool {
	if tb == nil {
		return false
	}

	raw := ev.Key
	if raw == "" {
		return false
	}

	command := strings.ToLower(key)
	shift := ev.Modifiers&ModifierShift != 0
	ctrlOrMeta := ev.Modifiers&(ModifierControl|ModifierMeta) != 0
	word := ev.Modifiers&(ModifierControl|ModifierMeta) != 0
	before := tb.Value
	beforeSelStart, beforeSelEnd := tb.Selection()

	switch command {
	case "left":
		tb.MoveCaretLeft(word, shift)
	case "right":
		tb.MoveCaretRight(word, shift)
	case "home":
		tb.MoveCaretToStart(shift)
	case "end":
		tb.MoveCaretToEnd(shift)
	case "backspace":
		tb.Backspace()
	case "delete", "del":
		tb.Delete()
	case "enter":
		tb.Submit()
		return false
	case "escape":
		tb.Cancel()
		return false
	case "a":
		if ctrlOrMeta {
			tb.SelectAll()
			start, end := tb.Selection()
			return beforeSelStart != start || beforeSelEnd != end
		}
		tb.InsertText(raw)
	case "c":
		if !ctrlOrMeta {
			tb.InsertText(raw)
			break
		}
		text := tb.SelectedText()
		if text == "" {
			return false
		}
		setClipboard(text)
	case "x":
		if !ctrlOrMeta {
			tb.InsertText(raw)
			break
		}
		text := tb.SelectedText()
		if text == "" {
			return false
		}
		setClipboard(text)
		tb.Delete()
	case "v":
		if !ctrlOrMeta {
			tb.InsertText(raw)
			break
		}
		tb.InsertText(getClipboard())
		return tb.Value != before
	case " ":
		tb.InsertText(" ")
	default:
		if ctrlOrMeta {
			return false
		}
		if len(raw) == 1 {
			tb.InsertText(raw)
			return tb.Value != before
		}
		return false
	}
	return tb.Value != before
}

func (w *WidgetInteraction) blurFocusedTextLocked() bool {
	if w.focusedText == nil || !w.focusedText.Active {
		return false
	}
	w.focusedText.Active = false
	w.focusedText = nil
	return true
}

func cursorIndexFromTextPoint(tb *core.TextBox, axes *Axes, fig *Figure, point geom.Pt) int {
	if tb == nil || axes == nil || fig == nil {
		return 0
	}
	ctx := core.AxesDrawContext(axes, fig)
	return tb.CaretForPoint(point, ctx)
}

func setClipboard(value string) {
	clipboardMu.Lock()
	defer clipboardMu.Unlock()
	clipboardText = value
}

func getClipboard() string {
	clipboardMu.Lock()
	defer clipboardMu.Unlock()
	return clipboardText
}
