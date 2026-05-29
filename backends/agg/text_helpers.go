package agg

import (
	"errors"
	"math"
	"strings"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type textBackend uint8

const (
	textBackendUnavailable textBackend = iota
	textBackendRaster
	textBackendGSV
)

type configuredTextFont struct {
	backend  textBackend
	face     render.FontFace
	fontPath string
	size     float64
	fallback bool
}

func (r *Renderer) configureTextFont(size float64, fontKey string) configuredTextFont {
	if size <= 0 {
		return configuredTextFont{}
	}
	if face := r.resolveTextFontFace(fontKey); fontReference(face) != "" {
		if _, err := loadRasterFont(face); err != nil {
			if !r.emergencyTextFallback {
				return configuredTextFont{backend: textBackendUnavailable}
			}
			r.markEmergencyTextFallback()
			return configuredTextFont{
				backend:  textBackendGSV,
				size:     math.Max(size, 6),
				fallback: true,
			}
		}
		return configuredTextFont{
			backend:  textBackendRaster,
			face:     face,
			fontPath: face.Path,
			size:     size,
		}
	}
	if !r.emergencyTextFallback {
		return configuredTextFont{backend: textBackendUnavailable}
	}
	r.markEmergencyTextFallback()
	return configuredTextFont{
		backend:  textBackendGSV,
		size:     math.Max(size, 6),
		fallback: true,
	}
}

func (r *Renderer) markEmergencyTextFallback() {
	if r == nil {
		return
	}
	r.fallback = true
	r.emergencyTextFallbackUsed = true
}

func measureLocalGSVTextWidth(text string, size float64) float64 {
	gsv := agglib.NewGSVText()
	gsv.SetFlip(true)
	gsv.SetSize(size, 0)
	return gsv.MeasureText(text)
}

func measureLocalGSVTextBounds(text string, size float64) (x, y, width, height float64, ok bool) {
	if text == "" || size <= 0 {
		return 0, 0, 0, 0, false
	}

	gsv := agglib.NewGSVText()
	gsv.SetFlip(true)
	gsv.SetSize(size, 0)
	gsv.SetText(text)
	gsv.SetStartPoint(0, 0)
	gsv.Rewind(0)

	var minX, minY, maxX, maxY float64
	have := false
	for {
		vx, vy, cmd := gsv.Vertex()
		switch cmd {
		case agglib.GSVPathCmdStop:
			if !have {
				return 0, 0, 0, 0, false
			}
			return minX, minY, maxX - minX, maxY - minY, true
		case agglib.GSVPathCmdMoveTo, agglib.GSVPathCmdLineTo:
			if !have {
				minX, maxX = vx, vx
				minY, maxY = vy, vy
				have = true
				continue
			}
			minX = math.Min(minX, vx)
			minY = math.Min(minY, vy)
			maxX = math.Max(maxX, vx)
			maxY = math.Max(maxY, vy)
		}
	}
}

func measureTextPathBounds(text string, size float64, fontPath string) (x, y, width, height float64, ok bool) {
	path, ok := render.TextPath(text, geom.Pt{}, size, fontPath)
	if !ok || len(path.V) == 0 {
		return 0, 0, 0, 0, false
	}
	minX, minY := path.V[0].X, path.V[0].Y
	maxX, maxY := minX, minY
	for _, pt := range path.V[1:] {
		minX = math.Min(minX, pt.X)
		minY = math.Min(minY, pt.Y)
		maxX = math.Max(maxX, pt.X)
		maxY = math.Max(maxY, pt.Y)
	}
	// render.TextPath outlines are y-up (ascender at larger Y), but callers here
	// expect the y-down font-metric convention where the returned y is the top of
	// the ink box (so ascent = -y, descent = y+height). Reflect the vertical
	// extent back to y-down to preserve that contract.
	return minX, -maxY, maxX - minX, maxY - minY, true
}

func (r *Renderer) configureOutlineFont(fontPath string, size float64) (*agglib.FreeTypeOutlineText, error) {
	if fontPath == "" || size <= 0 {
		return nil, errors.New("outline font is unavailable")
	}

	if r.outlineText == nil {
		txt, err := agglib.NewFreeTypeOutlineText()
		if err != nil {
			return nil, err
		}
		r.outlineText = txt
	}

	r.outlineText.SetHinting(false)
	r.outlineText.SetFlip(true)
	r.outlineText.SetResolution(96)
	r.outlineText.SetSize(size, 0)
	if err := r.outlineText.LoadFont(fontPath); err != nil {
		return nil, err
	}
	return r.outlineText, nil
}

func appendLocalGSVText(ctx *aggSurface, x, y, size float64, text string) bool {
	if ctx == nil || text == "" {
		return false
	}

	gsv := agglib.NewGSVText()
	gsv.SetFlip(true)
	gsv.SetSize(size, 0)
	gsv.SetText(text)
	gsv.SetStartPoint(math.Round(x), math.Round(y))

	ctx.BeginPath()
	gsv.Rewind(0)

	hasVertices := false
	for {
		vx, vy, cmd := gsv.Vertex()
		switch cmd {
		case agglib.GSVPathCmdStop:
			return hasVertices
		case agglib.GSVPathCmdMoveTo:
			ctx.MoveTo(vx, vy)
			hasVertices = true
		case agglib.GSVPathCmdLineTo:
			ctx.LineTo(vx, vy)
			hasVertices = true
		}
	}
}

func drawTrueTypeOutlineText(ctx *aggSurface, face *agglib.FreeTypeOutlineText, x, y float64, text string) bool {
	if ctx == nil || face == nil || text == "" {
		return false
	}

	face.SetText(text)
	face.SetStartPoint(x, y)

	if appendFreeTypeOutlineText(ctx, face) {
		ctx.Fill()
		return true
	}
	return false
}

func appendFreeTypeOutlineText(ctx *aggSurface, text *agglib.FreeTypeOutlineText) bool {
	if ctx == nil || text == nil {
		return false
	}

	ctx.BeginPath()
	text.Rewind(0)

	hasVertices := false
	for {
		x1, y1, cmd := text.Vertex()
		switch {
		case cmd == agglib.PathCmdStop:
			return hasVertices
		case cmd == agglib.PathCmdMoveTo:
			ctx.MoveTo(x1, y1)
			hasVertices = true
		case cmd == agglib.PathCmdLineTo:
			ctx.LineTo(x1, y1)
			hasVertices = true
		case agglib.IsPathCurve3(cmd):
			x2, y2, cmd2 := text.Vertex()
			if !agglib.IsPathCurve3(cmd2) {
				return hasVertices
			}
			ctx.QuadricCurveTo(x1, y1, x2, y2)
			hasVertices = true
		case agglib.IsPathCurve4(cmd):
			x2, y2, cmd2 := text.Vertex()
			x3, y3, cmd3 := text.Vertex()
			if !agglib.IsPathCurve4(cmd2) || !agglib.IsPathCurve4(cmd3) {
				return hasVertices
			}
			ctx.CubicCurveTo(x1, y1, x2, y2, x3, y3)
			hasVertices = true
		case agglib.IsPathEndPoly(cmd):
			if agglib.IsPathClose(cmd) {
				ctx.ClosePath()
			}
		}
	}
}

func (r *Renderer) resolveTextFontPath(fontKey string) string {
	return r.resolveTextFontFace(fontKey).Path
}

func (r *Renderer) resolveTextFontFace(fontKey string) render.FontFace {
	if wantsDefaultDejaVuSans(fontKey) && fontReference(r.defaultFontFace) != "" {
		return r.defaultFontFace
	}
	if face, ok := resolveFontFace(fontKey); ok {
		return face
	}
	if fontReference(r.defaultFontFace) != "" {
		return r.defaultFontFace
	}
	if face, ok := resolveFontFace("DejaVu Sans"); ok {
		return face
	}
	return render.FontFace{}
}

func wantsDefaultDejaVuSans(fontKey string) bool {
	fontKey = strings.TrimSpace(fontKey)
	lower := strings.ToLower(fontKey)
	// "DejaVu Sans Display" (used for large math operators) must resolve to its
	// own .ttf, not the cached regular DejaVu Sans face — otherwise big operators
	// render at regular size. Exclude any Display variant from the fast path.
	if strings.Contains(lower, "display") {
		return false
	}
	return fontKey == "" || strings.Contains(lower, "dejavu sans")
}
