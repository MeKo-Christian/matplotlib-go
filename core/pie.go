package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// PieOptions configures Axes.Pie.
type PieOptions struct {
	Center           geom.Pt
	Radius           float64
	Width            float64
	StartAngle       float64
	CounterClockwise *bool
	Colors           []render.Color
	Labels           []string
	Explode          []float64
	AutoPct          string
	LabelDistance    float64
	PctDistance      float64
	RotateLabels     bool
	Normalize        *bool
	Hatch            string
	Hatches          []string
	HatchColor       render.Color
	HatchWidth       float64
	Shadow           bool
	ShadowOffset     geom.Pt
	ShadowColor      render.Color
	EdgeColor        *render.Color
	LineWidth        float64
	Antialiased      *bool
	Alpha            float64
	Coords           CoordinateSpec
	Frame            *bool
}

// PieContainer groups the wedge and text artists created by Axes.Pie.
type PieContainer struct {
	Wedges      []*Wedge
	Shadows     []*Wedge
	Labels      []*Text
	AutoText    []*Text
	LabelAngles []float64
}

type PieLabelOptions struct {
	Distance  float64
	Rotate    bool
	Alignment string
	Coords    CoordinateSpec
}

// Wedge draws a pie-slice-style circular sector or ring segment.
type Wedge struct {
	Patch
	Center geom.Pt
	Radius float64
	Width  float64
	Theta1 float64
	Theta2 float64
	Coords CoordinateSpec
}

// Pie draws pie-slice wedges and returns the artists grouped in a container.
func (a *Axes) Pie(values []float64, opts ...PieOptions) *PieContainer {
	if a == nil || len(values) == 0 {
		return nil
	}
	rc := a.resolvedRC()
	cfg := PieOptions{
		Center:        geom.Pt{},
		Radius:        1,
		StartAngle:    0,
		LabelDistance: 1.1,
		PctDistance:   0.6,
		LineWidth:     rc.Patch.LineWidth,
		Alpha:         1,
		Coords:        Coords(CoordData),
	}
	if len(opts) > 0 {
		cfg = opts[0]
		if cfg.Radius <= 0 {
			cfg.Radius = 1
		}
		if cfg.LabelDistance <= 0 {
			cfg.LabelDistance = 1.1
		}
		if cfg.PctDistance <= 0 {
			cfg.PctDistance = 0.6
		}
		if cfg.LineWidth <= 0 {
			cfg.LineWidth = rc.Patch.LineWidth
		}
		if cfg.Alpha <= 0 {
			cfg.Alpha = 1
		}
		if cfg.Coords == (CoordinateSpec{}) {
			cfg.Coords = Coords(CoordData)
		}
	}

	total := 0.0
	filtered := make([]float64, len(values))
	for i, value := range values {
		if value >= 0 && isFinite(value) {
			filtered[i] = value
			total += value
		}
	}
	if total == 0 {
		return nil
	}
	normalize := specialtyBool(cfg.Normalize, true)
	if !normalize && total > 1 {
		return nil
	}

	container := &PieContainer{}
	edgeColor := render.Color{}
	if rc.Patch.ForceEdgeColor {
		edgeColor = rc.Patch.EdgeColor
	}
	if cfg.EdgeColor != nil {
		edgeColor = *cfg.EdgeColor
	}

	theta := cfg.StartAngle
	counterClockwise := specialtyBool(cfg.CounterClockwise, true)
	for i, value := range filtered {
		if value <= 0 {
			continue
		}
		fraction := value
		if normalize {
			fraction = value / total
		}
		span := fraction * 360
		end := theta + span
		if !counterClockwise {
			end = theta - span
		}
		mid := (theta + end) / 2
		explode := floatAt(cfg.Explode, i, 0)
		offset := geom.Pt{
			X: math.Cos(mid*math.Pi/180) * cfg.Radius * explode,
			Y: math.Sin(mid*math.Pi/180) * cfg.Radius * explode,
		}
		color := colorAt(a.NextColor(), cfg.Colors, i)
		color.A *= clampOneToOne(cfg.Alpha)
		hatchColor := cfg.HatchColor
		if hatchColor == (render.Color{}) {
			hatchColor = edgeColor
		}
		hatchWidth := cfg.HatchWidth
		if hatchWidth <= 0 {
			hatchWidth = rc.Hatch.LineWidth
		}
		if cfg.Shadow {
			shadowOffset := cfg.ShadowOffset
			if shadowOffset == (geom.Pt{}) {
				offset := pointsToPixels(a.figure.RC, 0.02) * cfg.Radius
				shadowOffset = geom.Pt{X: -offset, Y: -offset}
			}
			shadowColor := cfg.ShadowColor
			if shadowColor == (render.Color{}) {
				// Matplotlib patches.Shadow darkens the source patch facecolor
				// with shade=0.7 and alpha=0.5.
				shadowColor = render.Color{R: color.R * 0.3, G: color.G * 0.3, B: color.B * 0.3, A: 0.5}
			}
			shadow := &Wedge{
				Patch: Patch{
					FaceColor: shadowColor,
					// Matplotlib patches.Shadow updates both facecolor and
					// edgecolor to the darkened patch color.
					EdgeColor: shadowColor,
					EdgeWidth: cfg.LineWidth,
					Antialias: patchAntialiasMode(&rc.Patch, cfg.Antialiased),
					Alpha:     1,
					Label:     "_nolegend_",
					z:         1.8,
				},
				Center: geom.Pt{X: cfg.Center.X + offset.X + shadowOffset.X, Y: cfg.Center.Y + offset.Y + shadowOffset.Y},
				Radius: cfg.Radius,
				Width:  cfg.Width,
				Theta1: theta,
				Theta2: end,
				Coords: cfg.Coords,
			}
			a.AddPatch(shadow)
			container.Shadows = append(container.Shadows, shadow)
		}
		wedge := &Wedge{
			Patch: Patch{
				FaceColor:  color,
				EdgeColor:  edgeColor,
				EdgeWidth:  cfg.LineWidth,
				Antialias:  patchAntialiasMode(&rc.Patch, cfg.Antialiased),
				Alpha:      1,
				Label:      stringAt("", cfg.Labels, i),
				Hatch:      stringAt(cfg.Hatch, cfg.Hatches, i),
				HatchColor: hatchColor,
				HatchWidth: hatchWidth,
				z:          2,
			},
			Center: geom.Pt{X: cfg.Center.X + offset.X, Y: cfg.Center.Y + offset.Y},
			Radius: cfg.Radius,
			Width:  cfg.Width,
			Theta1: theta,
			Theta2: end,
			Coords: cfg.Coords,
		}
		a.AddPatch(wedge)
		container.Wedges = append(container.Wedges, wedge)

		if label := stringAt("", cfg.Labels, i); label != "" {
			labelPt := piePoint(wedge.Center, cfg.Radius*cfg.LabelDistance, mid)
			clipOn := false
			labelHAlign := pieLabelHAlign(labelPt.X)
			labelVAlign := pieLabelVAlign(labelPt.Y, cfg.RotateLabels)
			labelSize := a.effectiveRC(a.figure).TickLabelSize("x")
			container.LabelAngles = append(container.LabelAngles, pieLabelRotation(mid, cfg.RotateLabels))
			container.Labels = append(container.Labels, a.Text(labelPt.X, labelPt.Y, label, TextOptions{
				Coords:   cfg.Coords,
				FontSize: labelSize,
				HAlign:   labelHAlign,
				VAlign:   labelVAlign,
				Angle:    pieLabelRotation(mid, cfg.RotateLabels),
				ClipOn:   &clipOn,
			}))
		}
		if cfg.AutoPct != "" {
			autoPt := piePoint(wedge.Center, cfg.Radius*cfg.PctDistance, mid)
			autoText := fmt.Sprintf(cfg.AutoPct, fraction*100)
			clipOn := false
			container.AutoText = append(container.AutoText, a.Text(autoPt.X, autoPt.Y, autoText, TextOptions{
				Coords: cfg.Coords,
				HAlign: TextAlignCenter,
				VAlign: TextVAlignMiddle,
				ClipOn: &clipOn,
			}))
		}

		theta = end
	}

	if cfg.Coords == Coords(CoordData) {
		a.SetAxisEqual()
		if specialtyBool(cfg.Frame, false) {
			// frame=true: keep the frame and let autoscale fit the wedges,
			// mirroring matplotlib's pie() self._request_autoscale_view().
		} else {
			// Default frame=false: a fixed ±1.25 data window around the
			// center, independent of radius, with the frame hidden. Matches
			// matplotlib's xlim/ylim = (-1.25 + center, 1.25 + center); the
			// window must not scale with the pie radius.
			a.SetXLim(cfg.Center.X-1.25, cfg.Center.X+1.25)
			a.SetYLim(cfg.Center.Y-1.25, cfg.Center.Y+1.25)
			hideAxesFrame(a)
		}
	}
	return container
}

// PieLabel adds labels to an existing pie container, mirroring Matplotlib's pie_label helper.
func (a *Axes) PieLabel(container *PieContainer, labels []string, opts ...PieLabelOptions) []*Text {
	if a == nil || container == nil || len(container.Wedges) == 0 {
		return nil
	}
	cfg := PieLabelOptions{
		Distance:  0.6,
		Alignment: "auto",
		Coords:    Coords(CoordData),
	}
	if len(opts) > 0 {
		cfg = opts[0]
		if cfg.Distance <= 0 {
			cfg.Distance = 0.6
		}
		if cfg.Coords == (CoordinateSpec{}) {
			cfg.Coords = Coords(CoordData)
		}
	}
	out := make([]*Text, 0, len(container.Wedges))
	for i, wedge := range container.Wedges {
		if wedge == nil {
			continue
		}
		label := stringAt("", labels, i)
		if label == "" {
			continue
		}
		mid := (wedge.Theta1 + wedge.Theta2) / 2
		pt := piePoint(wedge.Center, wedge.Radius*cfg.Distance, mid)
		clipOn := false
		alignMode := strings.ToLower(strings.TrimSpace(cfg.Alignment))
		if alignMode == "" || alignMode == "auto" {
			if cfg.Distance > 1 {
				alignMode = "outer"
			} else {
				alignMode = "center"
			}
		}
		hAlign := TextAlignCenter
		vAlign := TextVAlignMiddle
		if alignMode == "outer" {
			hAlign, vAlign = pieOuterLabelAlignment(pt, cfg.Rotate)
		}
		txt := a.Text(pt.X, pt.Y, label, TextOptions{
			Coords: cfg.Coords,
			HAlign: hAlign,
			VAlign: vAlign,
			Angle:  pieLabelRotation(mid, cfg.Rotate),
			ClipOn: &clipOn,
		})
		out = append(out, txt)
		container.Labels = append(container.Labels, txt)
		container.LabelAngles = append(container.LabelAngles, pieLabelRotation(mid, cfg.Rotate))
	}
	return out
}

// Draw renders the wedge path using the embedded patch styling.
func (w *Wedge) Draw(ren render.Renderer, ctx *DrawContext) {
	if w == nil || ctx == nil || ren == nil || w.Radius <= 0 {
		return
	}
	path := buildArtistDisplayPath(ctx, w, w.Coords, w.localPath(), geom.Identity())
	w.drawStyledPath(ren, &ctx.RC, path, geom.Path{})
}

// Bounds returns the wedge's data-space bounds when applicable.
func (w *Wedge) Bounds(*DrawContext) geom.Rect {
	if w == nil || w.Radius <= 0 || !artistUsesDataCoords(w, w.Coords) {
		return geom.Rect{}
	}
	bounds, _ := pathBounds(w.localPath())
	return bounds
}

func (w *Wedge) localPath() geom.Path {
	if w.Width <= 0 || w.Width >= w.Radius {
		path := matplotlibArcPath(w.Center, w.Radius, w.Theta1, w.Theta2)
		if len(path.C) == 0 {
			return geom.Path{}
		}
		path.LineTo(w.Center)
		path.C = append(path.C, geom.ClosePath)
		return path
	}
	path := matplotlibArcPath(w.Center, w.Radius, w.Theta1, w.Theta2)
	if len(path.C) == 0 {
		return geom.Path{}
	}
	inner := matplotlibArcPath(w.Center, w.Radius-w.Width, w.Theta2, w.Theta1)
	if len(inner.C) == 0 {
		return path
	}
	firstInner := inner.V[0]
	path.LineTo(firstInner)
	path.C = append(path.C, inner.C[1:]...)
	path.V = append(path.V, inner.V[1:]...)
	path.C = append(path.C, geom.ClosePath)
	return path
}

func matplotlibArcPath(center geom.Pt, radius, theta1, theta2 float64) geom.Path {
	if radius <= 0 {
		return geom.Path{}
	}
	// geom.Arc builds the unit arc with Matplotlib's algorithm; scale to the
	// requested radius and translate to center. radius*v + center is
	// float-identical to baking center/radius into the vertex math.
	return geom.Arc(theta1, theta2, 0).Transformed(geom.Affine{
		A: radius, D: radius, E: center.X, F: center.Y,
	})
}

func piePoint(center geom.Pt, radius, angleDeg float64) geom.Pt {
	rad := angleDeg * math.Pi / 180
	return geom.Pt{
		X: center.X + radius*math.Cos(rad),
		Y: center.Y + radius*math.Sin(rad),
	}
}

func pieLabelHAlign(x float64) TextAlign {
	if x > 0 {
		return TextAlignLeft
	}
	return TextAlignRight
}

func pieLabelVAlign(y float64, rotate bool) TextVerticalAlign {
	if !rotate {
		return TextVAlignMiddle
	}
	if y > 0 {
		return TextVAlignBottom
	}
	return TextVAlignTop
}

func pieOuterLabelAlignment(pt geom.Pt, rotate bool) (TextAlign, TextVerticalAlign) {
	// Matplotlib pie_label(alignment="outer") aligns horizontally away
	// from x=0 and, when rotated, vertically away from y=0.
	hAlign := TextAlignRight
	if pt.X > 0 {
		hAlign = TextAlignLeft
	}
	vAlign := TextVAlignMiddle
	if rotate {
		vAlign = TextVAlignTop
		if pt.Y > 0 {
			vAlign = TextVAlignBottom
		}
	}
	return hAlign, vAlign
}

func pieLabelRotation(angleDeg float64, enabled bool) float64 {
	if !enabled {
		return 0
	}
	angle := math.Mod(angleDeg, 360)
	if angle < 0 {
		angle += 360
	}
	if angle > 90 && angle < 270 {
		angle += 180
	}
	return math.Mod(angle, 360)
}

func hideAxesFrame(a *Axes) {
	if a == nil {
		return
	}
	a.ShowFrame = false
	for _, axis := range []*Axis{a.XAxis, a.YAxis, a.XAxisTop, a.YAxisRight} {
		if axis == nil {
			continue
		}
		axis.ShowSpine = false
		axis.ShowTicks = false
		axis.ShowLabels = false
	}
}
