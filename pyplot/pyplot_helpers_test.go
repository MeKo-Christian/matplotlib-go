package pyplot

import (
	"math"

	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
)

type testFigureManager struct {
	canvas  canvas.FigureCanvas
	onShow  func()
	onClose func()
	tools   *canvas.ToolManager
}

func (m *testFigureManager) Canvas() canvas.FigureCanvas { return m.canvas }

func (m *testFigureManager) Show() error {
	if m.onShow != nil {
		m.onShow()
	}
	return nil
}

func (m *testFigureManager) Close() error {
	if m.onClose != nil {
		m.onClose()
	}
	return nil
}

func (m *testFigureManager) SetTitle(string) {}

func (m *testFigureManager) ToolManager() *canvas.ToolManager { return m.tools }

type testFigureCanvas struct {
	figure     *core.Figure
	onDraw     func()
	dispatcher canvas.Dispatcher
}

func (c *testFigureCanvas) Figure() *core.Figure { return c.figure }

func (c *testFigureCanvas) Draw() error {
	if c.onDraw != nil {
		c.onDraw()
	}
	return nil
}

func (c *testFigureCanvas) Resize(width, height int) error {
	if c.figure != nil {
		c.figure.SizePx.X = float64(width)
		c.figure.SizePx.Y = float64(height)
	}
	return nil
}

func (c *testFigureCanvas) Connect(t canvas.EventType, h canvas.Handler) canvas.ConnectionID {
	return c.dispatcher.Connect(t, h)
}

func (c *testFigureCanvas) Disconnect(id canvas.ConnectionID) {
	c.dispatcher.Disconnect(id)
}

func (c *testFigureCanvas) dispatch(ev canvas.Event) error {
	return c.dispatcher.Emit(ev)
}

func (c *testFigureCanvas) Close() error { return nil }

func approxFloat(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}
