package core

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type textLayoutVerticalAlign uint8

const (
	textLayoutVAlignTop textLayoutVerticalAlign = iota
	textLayoutVAlignBottom
	textLayoutVAlignCenter
	textLayoutVAlignBaseline
	textLayoutVAlignCenterBaseline
)

type singleLineTextLayout struct {
	render.TextLineLayout
	MathLayout *MathTextLayout
}

func measureSingleLineTextLayout(r render.Renderer, text string, size float64, fontKey string, useTeX ...bool) singleLineTextLayout {
	return measureSingleLineTextLayoutParseMath(r, text, size, fontKey, true, useTeX...)
}

func measureSingleLineTextLayoutParseMath(r render.Renderer, text string, size float64, fontKey string, parseMath bool, useTeX ...bool) singleLineTextLayout {
	if texEnabled(useTeX) {
		if metricer, ok := r.(render.TeXMetricer); ok {
			if metrics, ok := metricer.MeasureTeX(text, size, fontKey); ok {
				return singleLineTextLayout{
					TextLineLayout: render.TextLineLayout{
						Width:   metrics.W,
						Ascent:  metrics.Ascent,
						Descent: metrics.Descent,
						Height:  metrics.H,
					},
				}
			}
		}
	}

	if parseMath {
		if layout, ok := layoutDisplayText(r, text, size, fontKey); ok {
			width, ascent, descent := layout.Width, layout.Ascent, layout.Descent
			height := layout.Height
			// On the Agg raster backend, matplotlib aligns mathtext by the
			// ink-image bbox (get_text_width_height_descent → to_raster), not the
			// advance box. Override the metrics so centered/right-aligned math
			// anchors to the same pixel as matplotlib. Vector/purego keep the box
			// metrics (matplotlib's to_vector path).
			if _, isRaster := r.(render.RGBAExporter); isRaster {
				if w, a, d, ok := mathLayoutImageMetrics(r, layout, fontKey); ok {
					width, ascent, descent, height = w, a, d, a+d
				}
			}
			lp := r.MeasureText("lp", size, fontKey)
			if lp.H > height {
				height = lp.H
			}
			if lp.Descent > descent {
				descent = lp.Descent
			}
			ascent = height - descent
			if ascent < 0 {
				ascent = 0
			}
			return singleLineTextLayout{
				TextLineLayout: render.TextLineLayout{
					Width:   width,
					Ascent:  ascent,
					Descent: descent,
					Height:  height,
				},
				MathLayout: &layout,
			}
		}
	}

	display := displayTextForMathParsing(text, parseMath)
	return singleLineTextLayout{
		TextLineLayout: render.MeasureTextLineLayout(r, display, size, fontKey),
	}
}

func texEnabled(useTeX []bool) bool {
	return len(useTeX) > 0 && useTeX[0]
}

func textBaselineOffset(layout singleLineTextLayout, align textLayoutVerticalAlign) float64 {
	switch align {
	case textLayoutVAlignTop:
		return -layout.Ascent
	case textLayoutVAlignBottom:
		return layout.Descent
	case textLayoutVAlignCenter:
		return -(layout.Ascent - layout.Descent) / 2
	case textLayoutVAlignCenterBaseline:
		return -layout.Ascent / 2
	default:
		return 0
	}
}

func textHorizontalOriginOffset(layout singleLineTextLayout, align TextAlign) float64 {
	switch align {
	case TextAlignLeft:
		return 0
	case TextAlignRight:
		return layout.Width
	default:
		return layout.Width / 2
	}
}

func alignedSingleLineOrigin(anchor geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign) geom.Pt {
	return geom.Pt{
		X: anchor.X - textHorizontalOriginOffset(layout, hAlign),
		Y: anchor.Y + textBaselineOffset(layout, vAlign),
	}
}

func layoutVerticalAlign(vAlign TextVerticalAlign, preferCenterBaseline bool) textLayoutVerticalAlign {
	switch vAlign {
	case TextVAlignTop:
		return textLayoutVAlignTop
	case TextVAlignBottom:
		return textLayoutVAlignBottom
	case TextVAlignBaseline:
		return textLayoutVAlignBaseline
	case TextVAlignCenterBaseline:
		return textLayoutVAlignCenterBaseline
	case TextVAlignMiddle:
		if preferCenterBaseline {
			return textLayoutVAlignCenterBaseline
		}
		return textLayoutVAlignCenter
	default:
		return textLayoutVAlignBaseline
	}
}
