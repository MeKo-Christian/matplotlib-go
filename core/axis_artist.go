package core

import "fmt"

// AxisArtist is a host-linked auxiliary axis drawn on top of an existing axes.
type AxisArtist struct {
	Host *Axes
	Axis *Axis
}

// NewAxisArtist clones a side-specific axis and registers it as an auxiliary axis.
func (a *Axes) NewAxisArtist(side AxisSide) *AxisArtist {
	if a == nil {
		return nil
	}

	var base *Axis
	switch side {
	case AxisBottom, AxisTop:
		base = a.XAxis
		if side == AxisTop && a.XAxisTop != nil {
			base = a.XAxisTop
		}
	case AxisLeft, AxisRight:
		base = a.YAxis
		if side == AxisRight && a.YAxisRight != nil {
			base = a.YAxisRight
		}
	default:
		return nil
	}

	axis := cloneAxisForSide(base, side)
	if axis == nil {
		return nil
	}
	a.ExtraAxes = append(a.ExtraAxes, axis)
	return &AxisArtist{
		Host: a,
		Axis: axis,
	}
}

// FloatingXAxis creates an auxiliary x-axis drawn through the requested y data value.
func (a *Axes) FloatingXAxis(y float64) *AxisArtist {
	artist := a.NewAxisArtist(AxisBottom)
	if artist == nil || artist.Axis == nil {
		return nil
	}
	artist.Axis.SetSpinePositionData(y)
	return artist
}

// FloatingYAxis creates an auxiliary y-axis drawn through the requested x data value.
func (a *Axes) FloatingYAxis(x float64) *AxisArtist {
	artist := a.NewAxisArtist(AxisLeft)
	if artist == nil || artist.Axis == nil {
		return nil
	}
	artist.Axis.SetSpinePositionData(x)
	return artist
}

// SetTickDirection forwards tick-direction updates to the underlying auxiliary axis.
func (a *AxisArtist) SetTickDirection(direction string) error {
	if a == nil || a.Axis == nil {
		return nil
	}
	parsed, err := ParseTickDirection(direction)
	if err != nil {
		return err
	}
	a.Axis.TickDirection = parsed
	return nil
}

// SetSpinePositionData forwards data-position updates to the underlying auxiliary axis.
func (a *AxisArtist) SetSpinePositionData(value float64) error {
	if a == nil || a.Axis == nil {
		return fmt.Errorf("axis artist is nil")
	}
	a.Axis.SetSpinePositionData(value)
	return nil
}
