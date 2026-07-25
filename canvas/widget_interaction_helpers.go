package canvas

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/widgets"
)

func axesInteractionArtists(axes *core.Axes) []core.Artist {
	if axes == nil {
		return nil
	}
	if len(axes.WidgetArtists) == 0 {
		return axes.Artists
	}
	out := make([]core.Artist, 0, len(axes.Artists)+len(axes.WidgetArtists))
	out = append(out, axes.Artists...)
	out = append(out, axes.WidgetArtists...)
	return out
}

type widgetPick struct {
	axes   *Axes
	widget any
	info   core.PickInfo
}

func (w *WidgetInteraction) pickWidget(ev *Event) widgetPick {
	fig := ev.Figure
	if fig == nil {
		fig = w.figure
	}
	if fig == nil {
		return widgetPick{}
	}
	hits := Pick(fig, ev.Position)
	for _, hit := range hits {
		switch hit.Artist.(type) {
		case *widgets.Button, *widgets.Slider, *widgets.RangeSlider, *widgets.CheckButtons, *widgets.RadioButtons, *widgets.TextBox,
			*widgets.SpanSelector, *widgets.RectangleSelector, *widgets.EllipseSelector, *widgets.PolygonSelector, *widgets.LassoSelector,
			*widgets.Cursor, *widgets.MultiCursor:
			return widgetPick{
				axes:   hit.Axes,
				widget: hit.Artist,
				info:   hit.Info,
			}
		}
	}
	return widgetPick{}
}

func axesInAxesList(axesList []*Axes, axes *Axes) bool {
	for _, candidate := range axesList {
		if candidate == axes {
			return true
		}
	}
	return false
}

func widgetButton(v any) *widgets.Button {
	widget, ok := v.(*widgets.Button)
	if ok && widget.Enabled {
		return widget
	}
	return nil
}

func normalizeWidgetKey(key string) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "arrowleft", "left":
		return "left"
	case "arrowright", "right":
		return "right"
	case "arrowup", "up":
		return "up"
	case "arrowdown", "down":
		return "down"
	case "pageup":
		return "up"
	case "pagedown":
		return "down"
	case "home":
		return "home"
	case "end":
		return "end"
	case "escape":
		return "escape"
	case "return":
		return "enter"
	case "spacebar":
		return "space"
	default:
		return key
	}
}

func clampInt(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
