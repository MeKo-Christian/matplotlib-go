package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
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
	cfg := PieOptions{
		Center:        geom.Pt{},
		Radius:        1,
		StartAngle:    0,
		LabelDistance: 1.15,
		PctDistance:   0.62,
		LineWidth:     1,
		Alpha:         1,
		Coords:        Coords(CoordData),
	}
	if len(opts) > 0 {
		cfg = opts[0]
		if cfg.Radius <= 0 {
			cfg.Radius = 1
		}
		if cfg.LabelDistance <= 0 {
			cfg.LabelDistance = 1.15
		}
		if cfg.PctDistance <= 0 {
			cfg.PctDistance = 0.62
		}
		if cfg.LineWidth <= 0 {
			cfg.LineWidth = 1
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
	edgeColor := render.Color{R: 1, G: 1, B: 1, A: 1}
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
			hatchWidth = 1
		}
		if cfg.Shadow {
			shadowOffset := cfg.ShadowOffset
			if shadowOffset == (geom.Pt{}) {
				shadowOffset = geom.Pt{X: -0.02 * cfg.Radius, Y: -0.02 * cfg.Radius}
			}
			shadowColor := cfg.ShadowColor
			if shadowColor == (render.Color{}) {
				shadowColor = render.Color{R: 0, G: 0, B: 0, A: 0.25}
			}
			shadow := &Wedge{
				Patch: Patch{
					FaceColor: shadowColor,
					EdgeColor: render.Color{},
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
			container.LabelAngles = append(container.LabelAngles, pieLabelRotation(mid, cfg.RotateLabels))
			container.Labels = append(container.Labels, a.Text(labelPt.X, labelPt.Y, label, TextOptions{
				Coords: cfg.Coords,
				HAlign: pieAlign(mid),
				VAlign: TextVAlignMiddle,
				Angle:  pieLabelRotation(mid, cfg.RotateLabels),
				ClipOn: &clipOn,
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
		padding := cfg.Radius * 1.25
		a.SetXLim(cfg.Center.X-padding, cfg.Center.X+padding)
		a.SetYLim(cfg.Center.Y-padding, cfg.Center.Y+padding)
		a.SetAxisEqual()
		if !specialtyBool(cfg.Frame, false) {
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
		align := pieAlign(mid)
		if strings.EqualFold(cfg.Alignment, "center") {
			align = TextAlignCenter
		}
		txt := a.Text(pt.X, pt.Y, label, TextOptions{
			Coords: cfg.Coords,
			HAlign: align,
			VAlign: TextVAlignMiddle,
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
	path := buildDisplayPath(ctx, w.Coords, w.localPath(), geom.Identity())
	w.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the wedge's data-space bounds when applicable.
func (w *Wedge) Bounds(*DrawContext) geom.Rect {
	if w == nil || w.Radius <= 0 || !isDataCoords(w.Coords) {
		return geom.Rect{}
	}
	bounds, _ := pathBounds(w.localPath())
	return bounds
}

func (w *Wedge) localPath() geom.Path {
	outer := specialtyArcPoints(w.Center, w.Radius, w.Theta1, w.Theta2)
	if len(outer) == 0 {
		return geom.Path{}
	}
	if w.Width <= 0 || w.Width >= w.Radius {
		points := make([]geom.Pt, 0, len(outer)+1)
		points = append(points, w.Center)
		points = append(points, outer...)
		return polygonPath(points, true)
	}
	inner := specialtyArcPoints(w.Center, w.Radius-w.Width, w.Theta2, w.Theta1)
	points := make([]geom.Pt, 0, len(outer)+len(inner))
	points = append(points, outer...)
	points = append(points, inner...)
	return polygonPath(points, true)
}

func specialtyArcPoints(center geom.Pt, radius, theta1, theta2 float64) []geom.Pt {
	if radius <= 0 {
		return nil
	}
	span := theta2 - theta1
	steps := max(24, int(math.Ceil(math.Abs(span)/3))+1)
	points := make([]geom.Pt, 0, steps)
	for i := range steps {
		t := float64(i) / float64(max(steps-1, 1))
		angle := (theta1 + span*t) * math.Pi / 180
		points = append(points, geom.Pt{
			X: center.X + radius*math.Cos(angle),
			Y: center.Y + radius*math.Sin(angle),
		})
	}
	return points
}

func piePoint(center geom.Pt, radius, angleDeg float64) geom.Pt {
	rad := angleDeg * math.Pi / 180
	return geom.Pt{
		X: center.X + radius*math.Cos(rad),
		Y: center.Y + radius*math.Sin(rad),
	}
}

func pieAlign(angleDeg float64) TextAlign {
	rad := angleDeg * math.Pi / 180
	switch {
	case math.Cos(rad) < -0.2:
		return TextAlignRight
	case math.Cos(rad) > 0.2:
		return TextAlignLeft
	default:
		return TextAlignCenter
	}
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
