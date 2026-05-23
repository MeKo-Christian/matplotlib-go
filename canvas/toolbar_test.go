package canvas

import (
	"errors"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
)

func TestToolbarControllerDefaultActionsEnabled(t *testing.T) {
	c := NewToolbarController(nil)
	for _, a := range standardToolbarActions() {
		if !c.ActionEnabled(a) {
			t.Fatalf("action %q not enabled by default", a)
		}
	}
}

func TestToolbarControllerSetActionEnabled(t *testing.T) {
	c := NewToolbarController(nil)
	c.SetActionEnabled(ToolbarSave, false)
	if c.ActionEnabled(ToolbarSave) {
		t.Fatal("Save should be disabled")
	}
	if err := c.Trigger(ToolbarSave); err != nil {
		t.Fatal(err)
	}
}

func TestToolbarControllerTriggerHandler(t *testing.T) {
	c := NewToolbarController(nil)
	calls := 0
	c.SetHandler(ToolbarSave, func() error {
		calls++
		return nil
	})
	if err := c.Trigger(ToolbarSave); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1", calls)
	}
}

func TestToolbarControllerTriggerPropagatesError(t *testing.T) {
	c := NewToolbarController(nil)
	want := errors.New("boom")
	c.SetHandler(ToolbarHome, func() error { return want })
	if err := c.Trigger(ToolbarHome); err != want {
		t.Fatalf("Trigger err = %v, want %v", err, want)
	}
}

func TestToolbarControllerPanZoomTogglesNavigationMode(t *testing.T) {
	fig := &core.Figure{}
	nav := NewNavigation(fig, nil)
	c := NewToolbarController(nav)

	if c.Mode() != ToolbarModeNone {
		t.Fatalf("initial mode = %v, want None", c.Mode())
	}

	if err := c.Trigger(ToolbarPan); err != nil {
		t.Fatal(err)
	}
	if c.Mode() != ToolbarModePan {
		t.Fatalf("after pan trigger mode = %v, want Pan", c.Mode())
	}
	if nav.Mode() != NavPan {
		t.Fatalf("nav mode = %v, want NavPan", nav.Mode())
	}

	if err := c.Trigger(ToolbarZoom); err != nil {
		t.Fatal(err)
	}
	if c.Mode() != ToolbarModeZoom {
		t.Fatalf("after zoom trigger mode = %v, want Zoom", c.Mode())
	}
	if nav.Mode() != NavZoom {
		t.Fatalf("nav mode = %v, want NavZoom", nav.Mode())
	}

	// Second click of the same toggle returns to None.
	if err := c.Trigger(ToolbarZoom); err != nil {
		t.Fatal(err)
	}
	if c.Mode() != ToolbarModeNone {
		t.Fatalf("after second zoom trigger mode = %v, want None", c.Mode())
	}
	if nav.Mode() != NavNone {
		t.Fatalf("nav mode = %v, want NavNone", nav.Mode())
	}
}

func TestToolbarControllerSetMessage(t *testing.T) {
	c := NewToolbarController(nil)
	c.SetMessage("x=1, y=2")
	if got := c.Message(); got != "x=1, y=2" {
		t.Fatalf("Message = %q, want %q", got, "x=1, y=2")
	}
}

func TestToolbarControllerNilSafe(t *testing.T) {
	var c *ToolbarController
	if c.Mode() != ToolbarModeNone {
		t.Fatal("nil Mode != None")
	}
	if c.ActionEnabled(ToolbarHome) {
		t.Fatal("nil ActionEnabled returned true")
	}
	c.SetMode(ToolbarModePan) // must not panic
	c.SetActionEnabled(ToolbarHome, true)
	c.SetHandler(ToolbarHome, func() error { return nil })
	c.SetMessage("x")
	if err := c.Trigger(ToolbarHome); err == nil {
		t.Fatal("nil Trigger should error")
	}
}

// Compile-time check that ToolbarController satisfies NavigationToolbar.
var _ NavigationToolbar = (*ToolbarController)(nil)
