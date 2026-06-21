//go:build js && wasm

package main

import (
	"encoding/base64"
	"fmt"
	"runtime/debug"
	"syscall/js"

	plotcanvas "github.com/cwbudde/matplotlib-go/canvas"
	wasmcanvas "github.com/cwbudde/matplotlib-go/canvas/wasm"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/webdemo"
)

var (
	callbacks      []js.Func
	currentManager *wasmcanvas.Manager
)

type wasmCallback func(js.Value, []js.Value) any

func main() {
	callbacks = append(
		callbacks,
		safeCallback("listDemos", func(string) any { return js.Global().Get("Array").New() }, listDemos),
		safeCallback("listBackends", func(string) any { return js.Global().Get("Array").New() }, listBackends),
		safeCallback("mountDemo", errorResult, mountDemo),
		safeCallback("resizeDemo", errorResult, resizeDemo),
		safeCallback("unmountDemo", errorResult, unmountDemo),
		safeCallback("defaultDemoID", func(string) any { return webdemo.DefaultDemoID() }, defaultDemoID),
		safeCallback("defaultBackendID", func(string) any { return webdemo.DefaultBackendID() }, defaultBackendID),
		safeCallback("renderDemoPNG", errorResult, renderDemoPNG),
		safeCallback("demoSource", errorResult, demoSource),
		safeCallback("setNavigationMode", errorResult, setNavigationMode),
		safeCallback("navigationMode", func(string) any { return "" }, navigationMode),
		safeCallback("triggerToolbar", errorResult, triggerToolbar),
	)

	api := js.Global().Get("Object").New()
	api.Set("listDemos", callbacks[0])
	api.Set("listBackends", callbacks[1])
	api.Set("mountDemo", callbacks[2])
	api.Set("resizeDemo", callbacks[3])
	api.Set("unmountDemo", callbacks[4])
	api.Set("defaultDemoID", callbacks[5])
	api.Set("defaultBackendID", callbacks[6])
	api.Set("renderDemoPNG", callbacks[7])
	api.Set("demoSource", callbacks[8])
	api.Set("setNavigationMode", callbacks[9])
	api.Set("navigationMode", callbacks[10])
	api.Set("triggerToolbar", callbacks[11])
	js.Global().Set("matplotlibGoWASM", api)
	js.Global().Get("console").Call("log", "matplotlib-go wasm ready")

	select {}
}

func safeCallback(name string, fallback func(string) any, fn wasmCallback) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) (result any) {
		defer func() {
			if recovered := recover(); recovered != nil {
				message := fmt.Sprintf("matplotlib-go wasm: %s panicked: %v", name, recovered)
				consoleError(message)
				consoleError(string(debug.Stack()))
				if fallback != nil {
					result = fallback(message)
					return
				}
				result = nil
			}
		}()

		return fn(this, args)
	})
}

func errorResult(message string) any {
	result := js.Global().Get("Object").New()
	result.Set("error", message)
	return result
}

func consoleError(message string) {
	console := js.Global().Get("console")
	if console.IsUndefined() || console.IsNull() {
		return
	}
	console.Call("error", message)
}

func listDemos(_ js.Value, _ []js.Value) any {
	result := js.Global().Get("Array").New()
	for _, descriptor := range webdemo.Catalog() {
		item := js.Global().Get("Object").New()
		item.Set("id", descriptor.ID)
		item.Set("title", descriptor.Title)
		item.Set("description", descriptor.Description)
		result.Call("push", item)
	}
	return result
}

func listBackends(_ js.Value, _ []js.Value) any {
	result := js.Global().Get("Array").New()
	for _, descriptor := range webdemo.Backends() {
		item := js.Global().Get("Object").New()
		item.Set("id", descriptor.ID)
		item.Set("name", descriptor.Name)
		item.Set("description", descriptor.Description)
		result.Call("push", item)
	}
	return result
}

func mountDemo(_ js.Value, args []js.Value) any {
	canvasID := "plotCanvas"
	id := webdemo.DefaultDemoID()
	width := webdemo.DefaultWidth
	height := webdemo.DefaultHeight
	backendID := webdemo.DefaultBackendID()

	if len(args) > 0 && args[0].Type() == js.TypeString && args[0].String() != "" {
		canvasID = args[0].String()
	}
	if len(args) > 1 && args[1].Type() == js.TypeString && webdemo.ValidDemoID(args[1].String()) {
		id = args[1].String()
	}
	if len(args) > 2 && args[2].Type() == js.TypeNumber {
		width = args[2].Int()
	}
	if len(args) > 3 && args[3].Type() == js.TypeNumber {
		height = args[3].Int()
	}
	if len(args) > 4 && args[4].Type() == js.TypeString && webdemo.ValidBackendID(args[4].String()) {
		backendID = args[4].String()
	}

	return loadDemo(canvasID, id, backendID, width, height)
}

func resizeDemo(_ js.Value, args []js.Value) any {
	result := js.Global().Get("Object").New()
	if currentManager == nil {
		result.Set("error", "no mounted demo")
		return result
	}

	width := webdemo.DefaultWidth
	height := webdemo.DefaultHeight
	if len(args) > 0 && args[0].Type() == js.TypeNumber {
		width = args[0].Int()
	}
	if len(args) > 1 && args[1].Type() == js.TypeNumber {
		height = args[1].Int()
	}
	if err := currentManager.Canvas().Resize(width, height); err != nil {
		result.Set("error", err.Error())
		return result
	}
	result.Set("width", width)
	result.Set("height", height)
	return result
}

func unmountDemo(_ js.Value, _ []js.Value) any {
	if currentManager != nil {
		_ = currentManager.Close()
	}
	currentManager = nil
	return nil
}

func defaultDemoID(_ js.Value, _ []js.Value) any {
	return webdemo.DefaultDemoID()
}

func defaultBackendID(_ js.Value, _ []js.Value) any {
	return webdemo.DefaultBackendID()
}

func renderDemoPNG(_ js.Value, args []js.Value) any {
	result := js.Global().Get("Object").New()
	id := webdemo.DefaultDemoID()
	width := webdemo.DefaultWidth
	height := webdemo.DefaultHeight
	backendID := webdemo.DefaultBackendID()

	if len(args) > 0 && args[0].Type() == js.TypeString && webdemo.ValidDemoID(args[0].String()) {
		id = args[0].String()
	}
	if len(args) > 1 && args[1].Type() == js.TypeNumber {
		width = args[1].Int()
	}
	if len(args) > 2 && args[2].Type() == js.TypeNumber {
		height = args[2].Int()
	}
	if len(args) > 3 && args[3].Type() == js.TypeString && webdemo.ValidBackendID(args[3].String()) {
		backendID = args[3].String()
	}

	pngBytes, descriptor, err := webdemo.RenderPNGWithBackend(id, backendID, width, height)
	if err != nil {
		result.Set("error", err.Error())
		return result
	}

	result.Set("id", descriptor.ID)
	result.Set("backendID", backendID)
	result.Set("title", descriptor.Title)
	result.Set("description", descriptor.Description)
	result.Set("width", width)
	result.Set("height", height)
	result.Set("mimeType", "image/png")
	result.Set("base64", base64.StdEncoding.EncodeToString(pngBytes))
	return result
}

func demoSource(_ js.Value, args []js.Value) any {
	result := js.Global().Get("Object").New()
	id := webdemo.DefaultDemoID()

	if len(args) > 0 && args[0].Type() == js.TypeString && webdemo.ValidDemoID(args[0].String()) {
		id = args[0].String()
	}

	source, descriptor, err := webdemo.Source(id)
	if err != nil {
		result.Set("error", err.Error())
		return result
	}

	result.Set("id", descriptor.ID)
	result.Set("title", descriptor.Title)
	result.Set("description", descriptor.Description)
	result.Set("language", "go")
	result.Set("source", source)
	return result
}

// setNavigationMode switches the active interaction (pan / zoom / none)
// on the current demo. Invalid mode strings return an error.
func setNavigationMode(_ js.Value, args []js.Value) any {
	result := js.Global().Get("Object").New()
	if currentManager == nil {
		result.Set("error", "no mounted demo")
		return result
	}
	if len(args) == 0 || args[0].Type() != js.TypeString {
		result.Set("error", "matplotlib-go wasm: setNavigationMode expects a mode string")
		return result
	}
	mode := args[0].String()
	tb := currentManager.Toolbar()
	switch mode {
	case "", "none":
		tb.SetMode(plotcanvas.ToolbarModeNone)
	case "pan":
		tb.SetMode(plotcanvas.ToolbarModePan)
	case "zoom":
		tb.SetMode(plotcanvas.ToolbarModeZoom)
	default:
		result.Set("error", fmt.Sprintf("matplotlib-go wasm: unknown navigation mode %q", mode))
		return result
	}
	result.Set("mode", toolbarModeString(tb.Mode()))
	return result
}

// navigationMode returns the current interaction mode as one of "none",
// "pan", or "zoom".
func navigationMode(_ js.Value, _ []js.Value) any {
	if currentManager == nil {
		return ""
	}
	return toolbarModeString(currentManager.Toolbar().Mode())
}

// triggerToolbar fires a standard toolbar action by name. Supported
// names are "home", "pan", "zoom", "back", "forward", and "save_figure"
// (mirroring canvas.ToolbarAction values).
func triggerToolbar(_ js.Value, args []js.Value) any {
	result := js.Global().Get("Object").New()
	if currentManager == nil {
		result.Set("error", "no mounted demo")
		return result
	}
	if len(args) == 0 || args[0].Type() != js.TypeString {
		result.Set("error", "matplotlib-go wasm: triggerToolbar expects an action string")
		return result
	}
	action := plotcanvas.ToolbarAction(args[0].String())
	if err := currentManager.Toolbar().Trigger(action); err != nil {
		result.Set("error", err.Error())
		return result
	}
	result.Set("action", string(action))
	result.Set("mode", toolbarModeString(currentManager.Toolbar().Mode()))
	return result
}

func toolbarModeString(mode plotcanvas.ToolbarMode) string {
	switch mode {
	case plotcanvas.ToolbarModePan:
		return "pan"
	case plotcanvas.ToolbarModeZoom:
		return "zoom"
	default:
		return "none"
	}
}

func loadDemo(canvasID, id, backendID string, width, height int) any {
	result := js.Global().Get("Object").New()

	fig, descriptor, err := webdemo.Build(id, width, height)
	if err != nil {
		result.Set("error", err.Error())
		return result
	}
	if currentManager != nil {
		_ = currentManager.Close()
		currentManager = nil
	}

	manager, err := newManager(canvasID, backendID, fig)
	if err != nil {
		result.Set("error", err.Error())
		return result
	}
	currentManager = manager
	if err := manager.Canvas().Resize(width, height); err != nil {
		result.Set("error", err.Error())
		return result
	}

	result.Set("id", descriptor.ID)
	result.Set("backendID", backendID)
	result.Set("title", descriptor.Title)
	result.Set("description", descriptor.Description)
	result.Set("width", width)
	result.Set("height", height)
	result.Set("mode", toolbarModeString(manager.Toolbar().Mode()))
	return result
}

func newManager(canvasID, backendID string, fig *core.Figure) (*wasmcanvas.Manager, error) {
	switch backendID {
	case "agg":
		return wasmcanvas.NewAggManager(canvasID, fig)
	case "gobasic":
		return wasmcanvas.NewGoBasicManager(canvasID, fig)
	case "":
		return wasmcanvas.NewAggManager(canvasID, fig)
	default:
		return nil, fmt.Errorf("matplotlib-go wasm: unknown backend %q", backendID)
	}
}
