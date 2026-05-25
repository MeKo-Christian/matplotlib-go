package webagg

// The WebAgg wire protocol carries two flavors of messages:
//
//   - JSON text frames in both directions for event traffic. All
//     payloads have a "type" field that selects the handler. Field
//     names match upstream Matplotlib so existing JS clients can speak
//     to this server (and our own static/mpl.js can speak to upstream).
//
//   - Binary frames server → client only, carrying a PNG buffer. The
//     buffer is either a full frame (when image_mode == "full") or a
//     sparse diff frame where unchanged pixels are encoded as fully
//     transparent. The current mode is announced via the "image_mode"
//     JSON event so the client knows how to apply the bytes.
//
// Message names in the constants below come straight from
// third_party/matplotlib/lib/matplotlib/backends/backend_webagg_core.py.

// Outbound message types — server → client.
const (
	MsgImageMode      = "image_mode"
	MsgCursor         = "cursor"
	MsgCaptureScroll  = "capture_scroll"
	MsgFigureLabel    = "figure_label"
	MsgResize         = "resize"
	MsgRubberband     = "rubberband"
	MsgSave           = "save"
	MsgNavigateMode   = "navigate_mode"
	MsgHistoryButtons = "history_buttons"
	MsgMessage        = "message"
	MsgRefresh        = "refresh"
	MsgDraw           = "draw"
)

// Inbound message types — client → server.
const (
	MsgAck                 = "ack"
	MsgDrawCmd             = "draw"
	MsgButtonPress         = "button_press"
	MsgButtonRelease       = "button_release"
	MsgDblClick            = "dblclick"
	MsgMotionNotify        = "motion_notify"
	MsgFigureEnter         = "figure_enter"
	MsgFigureLeave         = "figure_leave"
	MsgScroll              = "scroll"
	MsgKeyPress            = "key_press"
	MsgKeyRelease          = "key_release"
	MsgToolbarButton       = "toolbar_button"
	MsgRefreshCmd          = "refresh"
	MsgResizeCmd           = "resize"
	MsgSendImageMode       = "send_image_mode"
	MsgSetDevicePixelRatio = "set_device_pixel_ratio"
	MsgSetDPIRatio         = "set_dpi_ratio"
)

// imageMode names the active server → client image transport mode.
type imageMode string

const (
	imageModeFull imageMode = "full"
	imageModeDiff imageMode = "diff"
)

// inboundEvent is the loose-typed decoded form of every client → server
// JSON message. The set of fields is a union; only the ones relevant to
// a given message type are populated. We use float64 throughout because
// JSON has no integer type and a single decoded form keeps the dispatch
// branches readable.
type inboundEvent struct {
	Type             string   `json:"type"`
	X                float64  `json:"x"`
	Y                float64  `json:"y"`
	Button           int      `json:"button"`
	Buttons          int      `json:"buttons"`
	Step             float64  `json:"step"`
	Key              string   `json:"key"`
	Modifiers        []string `json:"modifiers"`
	Name             string   `json:"name"`
	Width            float64  `json:"width"`
	Height           float64  `json:"height"`
	DevicePixelRatio float64  `json:"device_pixel_ratio"`
	DPIRatio         float64  `json:"dpi_ratio"`
}

// outboundEvent is the structural form we marshal for every server →
// client JSON message. Anything we don't fill is omitted via the
// `omitempty` tag.
type outboundEvent struct {
	Type    string `json:"type"`
	Cursor  string `json:"cursor,omitempty"`
	Mode    string `json:"mode,omitempty"`
	Capture *bool  `json:"capture_scroll,omitempty"`
	Label   string `json:"label,omitempty"`
	// Size carries the {resize} payload — pixel width/height the client
	// should set on its canvas element.
	Size      []float64 `json:"size,omitempty"`
	ResizeFwd *bool     `json:"forward,omitempty"`
	// Rubberband corners. The upstream contract uses -1 for "remove".
	X0 *float64 `json:"x0,omitempty"`
	Y0 *float64 `json:"y0,omitempty"`
	X1 *float64 `json:"x1,omitempty"`
	Y1 *float64 `json:"y1,omitempty"`
	// Message body for the status line.
	Message string `json:"message,omitempty"`
	// History toggles.
	Back    *bool `json:"Back,omitempty"`
	HistFwd *bool `json:"Forward,omitempty"`
}
