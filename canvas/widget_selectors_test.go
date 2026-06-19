package canvas

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

func TestWidgetInteractionSpanSelectorMouseAndKeyboard(t *testing.T) {
	fig := core.NewFigure(160, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 100)
	span := ax.SpanSelector("horizontal")

	var got [][2]float64
	span.OnSelect(func(_ *core.SpanSelector, min, max float64) {
		got = append(got, [2]float64{min, max})
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	startPx := geom.Pt{X: 40, Y: 50}
	endPx := geom.Pt{X: 100, Y: 50}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: startPx, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("span press: %v", err)
	}
	if !span.Active {
		t.Fatal("span should become active on drag start")
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: endPx, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("span drag: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: endPx, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("span release: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("span select callbacks = %d, want 1", len(got))
	}
	dStart, _ := ax.PixelToData(startPx)
	dEnd, _ := ax.PixelToData(endPx)
	wantMin, wantMax := sortedPair(dStart.X, dEnd.X)
	assertCloseEnough(t, got[0][0], wantMin)
	assertCloseEnough(t, got[0][1], wantMax)

	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "left"}); err != nil {
		t.Fatalf("span left key: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "right", Modifiers: ModifierControl}); err != nil {
		t.Fatalf("span right ctrl key: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("span callback count = %d, want 3", len(got))
	}
	expectedAfter := wantMin - 5 + 50
	assertCloseEnough(t, span.Start, expectedAfter)
	assertCloseEnough(t, span.End, wantMax-5+50)
}

func TestWidgetInteractionRectangleSelectorMouseAndKeyboard(t *testing.T) {
	fig := core.NewFigure(160, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 100)
	rect := ax.RectangleSelector()

	var rectBounds [][2]float64
	rect.OnSelect(func(_ *core.RectangleSelector, bounds geom.Rect) {
		rectBounds = append(rectBounds, [2]float64{
			bounds.W(),
			bounds.H(),
		})
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	press := geom.Pt{X: 30, Y: 30}
	move := geom.Pt{X: 120, Y: 90}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: press, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("rectangle press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: move, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("rectangle drag: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: move, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("rectangle release: %v", err)
	}
	if !rect.Active {
		t.Fatal("rectangle should be active after drag")
	}
	if len(rectBounds) != 1 {
		t.Fatalf("rectangle select callbacks = %d, want 1", len(rectBounds))
	}
	if rectBounds[0][0] != rectBounds[0][1] {
		t.Fatalf("rectangle shift drag expected square bounds, got w=%g h=%g", rectBounds[0][0], rectBounds[0][1])
	}

	beforeMin, beforeMax := rect.Min, rect.Max
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "left"}); err != nil {
		t.Fatalf("rectangle left key: %v", err)
	}
	assertCloseEnough(t, rect.Min.X-beforeMin.X, -5)
	assertCloseEnough(t, rect.Max.X-beforeMax.X, -5)
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "right", Modifiers: ModifierControl}); err != nil {
		t.Fatalf("rectangle ctrl-right key: %v", err)
	}
	assertCloseEnough(t, rect.Min.X-(beforeMin.X+45), 0)
	assertCloseEnough(t, rect.Max.X-(beforeMax.X+45), 0)
}

func TestWidgetInteractionRectangleSelectorModifierMouseCreate(t *testing.T) {
	fig := core.NewFigure(200, 120)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 200)
	ax.SetYLim(0, 120)
	rect := ax.RectangleSelector()

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	startCenter := geom.Pt{X: 120, Y: 30}
	endCenter := geom.Pt{X: 160, Y: 70}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: startCenter, Button: MouseButtonLeft, Modifiers: ModifierControl}); err != nil {
		t.Fatalf("ctrl press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: endCenter, Button: MouseButtonLeft, Modifiers: ModifierControl}); err != nil {
		t.Fatalf("ctrl move: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: endCenter, Button: MouseButtonLeft, Modifiers: ModifierControl}); err != nil {
		t.Fatalf("ctrl release: %v", err)
	}

	startData, ok := ax.PixelToData(startCenter)
	if !ok {
		t.Fatal("ctrl start pixelToData failed")
	}
	endData, ok := ax.PixelToData(endCenter)
	if !ok {
		t.Fatal("ctrl end pixelToData failed")
	}

	expMin, expMax := selectorBoundsFromDrag(startData, endData, false, true)
	assertCloseEnough(t, rect.Min.X, expMin.X)
	assertCloseEnough(t, rect.Max.X, expMax.X)
	assertCloseEnough(t, rect.Min.Y, expMin.Y)
	assertCloseEnough(t, rect.Max.Y, expMax.Y)
	rect.Clear()

	startSquare := geom.Pt{X: 40, Y: 80}
	endSquare := geom.Pt{X: 100, Y: 60}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: startSquare, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("shift press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: endSquare, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("shift move: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: endSquare, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("shift release: %v", err)
	}

	startSquareData, ok := ax.PixelToData(startSquare)
	if !ok {
		t.Fatal("shift start pixelToData failed")
	}
	endSquareData, ok := ax.PixelToData(endSquare)
	if !ok {
		t.Fatal("shift end pixelToData failed")
	}
	expShiftMin, expShiftMax := selectorBoundsFromDrag(startSquareData, endSquareData, true, false)
	if math.Abs((rect.Max.X-rect.Min.X)-(rect.Max.Y-rect.Min.Y)) > 1e-9 {
		t.Fatalf("rectangle with shift should be square, got width=%g height=%g", rect.Max.X-rect.Min.X, rect.Max.Y-rect.Min.Y)
	}
	assertCloseEnough(t, rect.Min.X, expShiftMin.X)
	assertCloseEnough(t, rect.Max.X, expShiftMax.X)
	assertCloseEnough(t, rect.Min.Y, expShiftMin.Y)
	assertCloseEnough(t, rect.Max.Y, expShiftMax.Y)
	rect.Clear()

	startSquareFromCenter := geom.Pt{X: 150, Y: 40}
	endSquareFromCenter := geom.Pt{X: 170, Y: 100}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: startSquareFromCenter, Button: MouseButtonLeft, Modifiers: ModifierShift | ModifierControl}); err != nil {
		t.Fatalf("shift+ctrl press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: endSquareFromCenter, Button: MouseButtonLeft, Modifiers: ModifierShift | ModifierControl}); err != nil {
		t.Fatalf("shift+ctrl move: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: endSquareFromCenter, Button: MouseButtonLeft, Modifiers: ModifierShift | ModifierControl}); err != nil {
		t.Fatalf("shift+ctrl release: %v", err)
	}

	shiftCenterStart, ok := ax.PixelToData(startSquareFromCenter)
	if !ok {
		t.Fatal("shift+ctrl start pixelToData failed")
	}
	shiftCenterEnd, ok := ax.PixelToData(endSquareFromCenter)
	if !ok {
		t.Fatal("shift+ctrl end pixelToData failed")
	}
	expCenterMin, expCenterMax := selectorBoundsFromDrag(shiftCenterStart, shiftCenterEnd, true, true)
	assertCloseEnough(t, rect.Min.X, expCenterMin.X)
	assertCloseEnough(t, rect.Max.X, expCenterMax.X)
	assertCloseEnough(t, rect.Min.Y, expCenterMin.Y)
	assertCloseEnough(t, rect.Max.Y, expCenterMax.Y)
	if math.Abs((rect.Min.X+rect.Max.X)/2-shiftCenterStart.X) > 1e-9 {
		t.Fatalf("rectangle with shift+ctrl should stay centered on press, got center x %g want %g", (rect.Min.X+rect.Max.X)/2, shiftCenterStart.X)
	}
	if math.Abs((rect.Max.X-rect.Min.X)-(rect.Max.Y-rect.Min.Y)) > 1e-9 {
		t.Fatalf("rectangle with shift+ctrl should remain square, got width=%g height=%g", rect.Max.X-rect.Min.X, rect.Max.Y-rect.Min.Y)
	}
}

func TestWidgetInteractionPolygonSelectorEdit(t *testing.T) {
	fig := core.NewFigure(160, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 100)
	polygon := ax.PolygonSelector()

	var onSelectCount int
	polygon.OnSelect(func(*core.PolygonSelector, []geom.Pt) {
		onSelectCount++
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	p1 := geom.Pt{X: 30, Y: 20}
	p2 := geom.Pt{X: 130, Y: 20}
	p3 := geom.Pt{X: 80, Y: 80}
	p1Data, ok := ax.PixelToData(p1)
	if !ok {
		t.Fatal("pixelToData failed for first polygon point")
	}
	p2Data, ok := ax.PixelToData(p2)
	if !ok {
		t.Fatal("pixelToData failed for second polygon point")
	}
	p3Data, ok := ax.PixelToData(p3)
	if !ok {
		t.Fatal("pixelToData failed for third polygon point")
	}
	if !polygon.AppendPoint(p1Data) || !polygon.AppendPoint(p2Data) || !polygon.AppendPoint(p3Data) {
		t.Fatal("manual polygon point append should succeed")
	}
	if !polygon.Close() {
		t.Fatal("polygon close should succeed")
	}
	if !polygon.Closed {
		t.Fatal("polygon should be closed")
	}
	points := []geom.Pt{p1Data, p2Data, p3Data}
	if onSelectCount == 0 {
		polygon.TriggerOnSelect()
		if onSelectCount != 1 {
			t.Fatal("expected polygon onSelect after close")
		}
	}

	// Shift+drag should move all vertices together.
	center := geom.Pt{X: 80, Y: 40}
	beforeData0 := polygon.Points[0]
	beforeMoveData, _ := ax.PixelToData(center)
	afterMoveData, _ := ax.PixelToData(geom.Pt{X: 90, Y: 40})
	delta := geom.Pt{
		X: afterMoveData.X - beforeMoveData.X,
		Y: afterMoveData.Y - beforeMoveData.Y,
	}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: center, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("polygon move press: %v", err)
	}
	translated := geom.Pt{X: 90, Y: 40}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: translated, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("polygon move drag: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: translated, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("polygon move release: %v", err)
	}
	if len(polygon.Points) != len(points) {
		t.Fatalf("polygon points = %d, want %d", len(polygon.Points), len(points))
	}
	assertCloseEnough(t, polygon.Points[0].X-beforeData0.X-delta.X, 0)
	assertCloseEnough(t, polygon.Points[0].Y-beforeData0.Y-delta.Y, 0)
}

func TestWidgetInteractionPolygonSelectorPreCompleteMoveModes(t *testing.T) {
	fig := core.NewFigure(200, 120)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 200)
	ax.SetYLim(0, 120)

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	poly := ax.PolygonSelector()
	p1ShiftPx := geom.Pt{X: 50, Y: 70}
	p2ShiftPx := geom.Pt{X: 150, Y: 70}
	p1ShiftData, ok := ax.PixelToData(p1ShiftPx)
	if !ok {
		t.Fatal("polygon shift pre-complete start pixelToData failed")
	}
	p2ShiftData, ok := ax.PixelToData(p2ShiftPx)
	if !ok {
		t.Fatal("polygon shift pre-complete second pixelToData failed")
	}
	poly.AppendPoint(p1ShiftData)
	poly.AppendPoint(p2ShiftData)

	// Move all vertices before completion (shift).
	shiftPressPx := geom.Pt{X: 100, Y: 70}
	shiftMovePx := geom.Pt{X: 100, Y: 120}
	shiftPressData, ok := ax.PixelToData(shiftPressPx)
	if !ok {
		t.Fatal("polygon shift move press pixelToData failed")
	}
	shiftMoveData, ok := ax.PixelToData(shiftMovePx)
	if !ok {
		t.Fatal("polygon shift move pixelToData failed")
	}
	shiftDelta := geom.Pt{
		X: shiftMoveData.X - shiftPressData.X,
		Y: shiftMoveData.Y - shiftPressData.Y,
	}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: shiftPressPx, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("polygon shift press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: shiftMovePx, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("polygon shift move: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: shiftMovePx, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("polygon shift release: %v", err)
	}

	if len(poly.Points) != 2 {
		t.Fatalf("polygon points = %d, want 2", len(poly.Points))
	}
	assertCloseEnough(t, poly.Points[0].X-p1ShiftData.X-shiftDelta.X, 0)
	assertCloseEnough(t, poly.Points[0].Y-p1ShiftData.Y-shiftDelta.Y, 0)
	assertCloseEnough(t, poly.Points[1].X-p2ShiftData.X-shiftDelta.X, 0)
	assertCloseEnough(t, poly.Points[1].Y-p2ShiftData.Y-shiftDelta.Y, 0)

	// Move a vertex before completion (control).
	poly2 := ax.PolygonSelector()
	ctrlV0 := geom.Pt{X: 20, Y: 20}
	ctrlV1 := geom.Pt{X: 120, Y: 20}
	ctrlV0Data, ok := ax.PixelToData(ctrlV0)
	if !ok {
		t.Fatal("polygon control point pixelToData failed")
	}
	ctrlV1Data, ok := ax.PixelToData(ctrlV1)
	if !ok {
		t.Fatal("polygon control second pixelToData failed")
	}
	poly2.AppendPoint(ctrlV0Data)
	poly2.AppendPoint(ctrlV1Data)

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: ctrlV0, Button: MouseButtonLeft, Modifiers: ModifierControl}); err != nil {
		t.Fatalf("polygon control press: %v", err)
	}
	ctrlMovePx := geom.Pt{X: 30, Y: 20}
	ctrlMoveData, ok := ax.PixelToData(ctrlMovePx)
	if !ok {
		t.Fatal("polygon control move pixelToData failed")
	}
	ctrlDelta := geom.Pt{
		X: ctrlMoveData.X - ctrlV0Data.X,
		Y: ctrlMoveData.Y - ctrlV0Data.Y,
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: ctrlMovePx, Button: MouseButtonLeft, Modifiers: ModifierControl}); err != nil {
		t.Fatalf("polygon control move: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: ctrlMovePx, Button: MouseButtonLeft, Modifiers: ModifierControl}); err != nil {
		t.Fatalf("polygon control release: %v", err)
	}

	if len(poly2.Points) != 2 {
		t.Fatalf("polygon control points = %d, want 2", len(poly2.Points))
	}
	assertCloseEnough(t, poly2.Points[0].X-(ctrlV0Data.X+ctrlDelta.X), 0)
	assertCloseEnough(t, poly2.Points[0].Y-(ctrlV0Data.Y+ctrlDelta.Y), 0)
	assertCloseEnough(t, poly2.Points[1].X-ctrlV1Data.X, 0)
	assertCloseEnough(t, poly2.Points[1].Y-ctrlV1Data.Y, 0)
	if math.Abs((poly2.Points[0].X - (ctrlV0Data.X + ctrlDelta.X))) > 1e-9 {
		t.Fatalf("polygon control move should shift first vertex x, got %g", poly2.Points[0].X)
	}
}

func TestWidgetInteractionEllipseSelectorMouse(t *testing.T) {
	fig := core.NewFigure(160, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 100)
	ellipse := ax.EllipseSelector()

	var got float64
	var selected bool
	ellipse.OnSelect(func(_ *core.EllipseSelector, bounds geom.Rect) {
		got = bounds.W()
		selected = true
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	press := geom.Pt{X: 80, Y: 40}
	move := geom.Pt{X: 120, Y: 80}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: press, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("ellipse press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: move, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("ellipse drag: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: move, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("ellipse release: %v", err)
	}
	if !ellipse.Active || !selected {
		t.Fatal("ellipse should become active and selected after drag")
	}
	if got <= 0 {
		t.Fatalf("expected positive ellipse width on select, got %g", got)
	}
}

func TestWidgetInteractionEllipseSelectorKeyboard(t *testing.T) {
	fig := core.NewFigure(160, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 100)
	ellipse := ax.EllipseSelector()

	var selected int
	ellipse.OnSelect(func(_ *core.EllipseSelector, got geom.Rect) {
		selected++
		if got.Min.X == 0 && got.Min.Y == 0 && got.Max.X == 0 && got.Max.Y == 0 {
			t.Fatalf("ellipse callback reported empty bounds")
		}
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	press := geom.Pt{X: 80, Y: 40}
	move := geom.Pt{X: 120, Y: 80}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: press, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("ellipse press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: move, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("ellipse drag: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: move, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("ellipse release: %v", err)
	}
	if !ellipse.Active {
		t.Fatal("ellipse should be active after drag")
	}
	if selected != 1 {
		t.Fatalf("ellipse callback count = %d, want 1", selected)
	}

	beforeMin := ellipse.Min
	beforeMax := ellipse.Max
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "left"}); err != nil {
		t.Fatalf("ellipse left key: %v", err)
	}
	if selected != 2 {
		t.Fatalf("ellipse callback count after left = %d, want 2", selected)
	}
	assertCloseEnough(t, ellipse.Min.X-beforeMin.X+5, 0)
	assertCloseEnough(t, ellipse.Max.X-beforeMax.X+5, 0)
	assertCloseEnough(t, ellipse.Min.Y-beforeMin.Y, 0)
	assertCloseEnough(t, ellipse.Max.Y-beforeMax.Y, 0)

	beforeMin = ellipse.Min
	beforeMax = ellipse.Max
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "right", Modifiers: ModifierControl}); err != nil {
		t.Fatalf("ellipse ctrl right key: %v", err)
	}
	if selected != 3 {
		t.Fatalf("ellipse callback count after ctrl-right = %d, want 3", selected)
	}
	assertCloseEnough(t, ellipse.Min.X-beforeMin.X-50, 0)
	assertCloseEnough(t, ellipse.Max.X-beforeMax.X-50, 0)
	assertCloseEnough(t, ellipse.Min.Y-beforeMin.Y, 0)
	assertCloseEnough(t, ellipse.Max.Y-beforeMax.Y, 0)
}

func TestWidgetInteractionPolygonSelectorKeyboard(t *testing.T) {
	fig := core.NewFigure(200, 120)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 200)
	ax.SetYLim(0, 120)
	poly := ax.PolygonSelector()

	before := []geom.Pt{
		{X: 30, Y: 30},
		{X: 170, Y: 30},
		{X: 100, Y: 90},
	}
	points := make([]geom.Pt, len(before))
	for i, px := range before {
		data, ok := ax.PixelToData(px)
		if !ok {
			t.Fatalf("polygon point pixelToData %v failed", px)
		}
		points[i] = data
		poly.AppendPoint(data)
	}
	if !poly.Close() {
		t.Fatal("polygon close should succeed")
	}
	if !poly.Closed {
		t.Fatal("polygon should be closed")
	}
	if len(poly.Points) != len(points) {
		t.Fatalf("polygon points = %d, want %d", len(poly.Points), len(points))
	}

	var callbacks int
	poly.OnSelect(func(*core.PolygonSelector, []geom.Pt) {
		callbacks++
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: geom.Pt{X: 100, Y: 50}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("polygon focus press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: geom.Pt{X: 100, Y: 50}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("polygon focus release: %v", err)
	}

	beforeMove := make([]geom.Pt, len(poly.Points))
	copy(beforeMove, poly.Points)
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "right"}); err != nil {
		t.Fatalf("polygon right key: %v", err)
	}
	if callbacks < 1 {
		t.Fatalf("polygon callback count = %d, want >= 1 after move", callbacks)
	}
	for i := range poly.Points {
		assertCloseEnough(t, poly.Points[i].X-beforeMove[i].X-10, 0)
		assertCloseEnough(t, poly.Points[i].Y-beforeMove[i].Y, 0)
	}

	copy(beforeMove, poly.Points)
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "down"}); err != nil {
		t.Fatalf("polygon down key: %v", err)
	}
	if callbacks < 2 {
		t.Fatalf("polygon callback count after move = %d, want >= 2 after two key moves", callbacks)
	}
	for i := range poly.Points {
		assertCloseEnough(t, poly.Points[i].X-beforeMove[i].X, 0)
		assertCloseEnough(t, poly.Points[i].Y-beforeMove[i].Y+6, 0)
	}
}

func TestWidgetInteractionLassoSelectorMouse(t *testing.T) {
	fig := core.NewFigure(160, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 100)
	lasso := ax.LassoSelector()

	var got int
	lasso.OnSelect(func(_ *core.LassoSelector, points []geom.Pt) {
		got = len(points)
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	points := []geom.Pt{{X: 20, Y: 20}, {X: 60, Y: 20}, {X: 80, Y: 70}}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: points[0], Button: MouseButtonLeft}); err != nil {
		t.Fatalf("lasso press: %v", err)
	}
	for _, point := range points[1:] {
		if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: point, Button: MouseButtonLeft}); err != nil {
			t.Fatalf("lasso move: %v", err)
		}
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: points[len(points)-1], Button: MouseButtonLeft}); err != nil {
		t.Fatalf("lasso release: %v", err)
	}
	if !lasso.Active {
		t.Fatal("lasso should activate on release")
	}
	if got < len(points) {
		t.Fatalf("lasso onSelect points = %d, want >=%d", got, len(points))
	}
}

func TestWidgetInteractionLassoSelectorKeyboardEscape(t *testing.T) {
	fig := core.NewFigure(160, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 100)
	lasso := ax.LassoSelector()

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	points := []geom.Pt{{X: 20, Y: 20}, {X: 60, Y: 20}, {X: 80, Y: 70}}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: points[0], Button: MouseButtonLeft}); err != nil {
		t.Fatalf("lasso press: %v", err)
	}
	for _, point := range points[1:] {
		if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: point, Button: MouseButtonLeft}); err != nil {
			t.Fatalf("lasso move: %v", err)
		}
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: points[len(points)-1], Button: MouseButtonLeft}); err != nil {
		t.Fatalf("lasso release: %v", err)
	}
	if !lasso.Active {
		t.Fatal("lasso should be active after release")
	}

	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "escape"}); err != nil {
		t.Fatalf("lasso escape key: %v", err)
	}
	if lasso.Active {
		t.Fatal("lasso should clear on escape")
	}
	if lasso.Tracking {
		t.Fatal("lasso should not keep tracking after escape")
	}
}
