// matplotlib-go WebAgg client.
//
// Speaks the wire protocol defined in
//   backends/webagg/protocol.go (Go side)
//   third_party/matplotlib/.../backend_webagg_core.py (upstream side).
//
// Responsibilities:
//   - open a WebSocket to the server
//   - render full / diff PNG frames pushed as binary frames
//   - send mouse, scroll, key, resize, and toolbar events as JSON
//   - reflect navigate_mode and history_buttons events in the toolbar
//
// This client is intentionally smaller than upstream mpl.js: it
// implements the subset of the protocol our Go server emits and the
// toolbar actions the manager understands. Hosts who need richer
// behavior can swap mpl.js for upstream's own bundle — the wire
// protocol is compatible.

(function (global) {
  "use strict";

  function MPLFigure(opts) {
    this.root = document.getElementById(opts.rootId);
    this.canvas = document.getElementById(opts.canvasId);
    this.overlay = document.getElementById(opts.overlayId);
    this.status = document.getElementById(opts.statusId);
    this.ctx = this.canvas.getContext("2d");
    this.octx = this.overlay.getContext("2d");
    this.wsUrl = opts.wsUrl;

    this.imageMode = "full";
    this.imageObj = new Image();
    this.imageObj.onload = this._applyImage.bind(this);
    this.devicePixelRatio = window.devicePixelRatio || 1;

    this._installToolbar();
    this._installCanvasEvents();
    this._connect();
  }

  // ---- WebSocket lifecycle ---------------------------------------------

  MPLFigure.prototype._connect = function () {
    var self = this;
    var ws = new WebSocket(this.wsUrl);
    ws.binaryType = "blob";
    this.ws = ws;

    ws.onopen = function () {
      self.send({ type: "send_image_mode" });
      self.send({
        type: "set_device_pixel_ratio",
        device_pixel_ratio: self.devicePixelRatio,
      });
    };

    ws.onmessage = function (evt) {
      if (typeof evt.data === "string") {
        try {
          self._handleJSON(JSON.parse(evt.data));
        } catch (e) {
          console.error("matplotlib-go webagg: bad JSON", e, evt.data);
        }
        return;
      }
      // Binary frame → PNG payload. Re-decode via a blob URL so the
      // browser handles the format.
      var url = URL.createObjectURL(evt.data);
      self.imageObj.src = url;
      // Revoke once the image has loaded; _applyImage handles paint.
      self.imageObj.onload = function () {
        self._applyImage();
        URL.revokeObjectURL(url);
      };
    };

    ws.onclose = function () {
      // Try once to reconnect — useful for dev where the server
      // restarts during edits.
      setTimeout(function () {
        if (!self._closedByUser) self._connect();
      }, 1000);
    };

    ws.onerror = function (e) {
      console.error("matplotlib-go webagg: ws error", e);
    };
  };

  MPLFigure.prototype.send = function (payload) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(payload));
    }
  };

  // ---- Incoming JSON dispatch ------------------------------------------

  MPLFigure.prototype._handleJSON = function (msg) {
    switch (msg.type) {
      case "image_mode":
        this.imageMode = msg.mode;
        return;
      case "resize":
        this._handleResize(msg.size);
        return;
      case "figure_label":
        document.title = msg.label || document.title;
        return;
      case "cursor":
        this.canvas.style.cursor = msg.cursor || "default";
        return;
      case "rubberband":
        this._drawRubberband(msg.x0, msg.y0, msg.x1, msg.y1);
        return;
      case "message":
        if (this.status) this.status.textContent = msg.message || "";
        return;
      case "history_buttons":
        this._setToolbarEnabled("back", !!msg.Back);
        this._setToolbarEnabled("forward", !!msg.Forward);
        return;
      case "navigate_mode":
        this._reflectMode(msg.mode);
        return;
      case "refresh":
      case "draw":
        // Server-side bookkeeping; nothing to do client-side.
        return;
      case "save":
        // The server fired the save handler; nothing to do here unless
        // we eventually receive a download URL alongside.
        return;
    }
  };

  MPLFigure.prototype._handleResize = function (size) {
    if (!size || size.length !== 2) return;
    var w = size[0],
      h = size[1];
    this.canvas.width = w;
    this.canvas.height = h;
    this.canvas.style.width = w + "px";
    this.canvas.style.height = h + "px";
    this.overlay.width = w;
    this.overlay.height = h;
    this.overlay.style.width = w + "px";
    this.overlay.style.height = h + "px";
  };

  // ---- Image painting --------------------------------------------------

  MPLFigure.prototype._applyImage = function () {
    var img = this.imageObj;
    if (!img.width || !img.height) return;
    if (this.imageMode === "full") {
      this.ctx.drawImage(img, 0, 0);
    } else {
      // Diff: composite-source-over by default already only paints
      // non-transparent pixels, which is exactly what the diff mode
      // encodes. Reset any prior compositeOperation just in case.
      this.ctx.globalCompositeOperation = "source-over";
      this.ctx.drawImage(img, 0, 0);
    }
    // Acknowledge — keeps the server's flow control happy.
    this.send({ type: "ack" });
  };

  // ---- Toolbar ---------------------------------------------------------

  MPLFigure.prototype._installToolbar = function () {
    var self = this;
    var buttons = this.root.querySelectorAll(
      ".mpl-toolbar button[data-action]",
    );
    this.toolbar = {};
    buttons.forEach(function (btn) {
      var action = btn.getAttribute("data-action");
      self.toolbar[action] = btn;
      btn.addEventListener("click", function () {
        self.send({ type: "toolbar_button", name: action });
      });
    });
  };

  MPLFigure.prototype._setToolbarEnabled = function (name, enabled) {
    var btn = this.toolbar[name];
    if (!btn) return;
    btn.disabled = !enabled;
  };

  MPLFigure.prototype._reflectMode = function (mode) {
    var panBtn = this.toolbar["pan"];
    var zoomBtn = this.toolbar["zoom"];
    if (panBtn) panBtn.setAttribute("aria-pressed", mode === "PAN");
    if (zoomBtn) zoomBtn.setAttribute("aria-pressed", mode === "ZOOM");
  };

  // ---- Canvas events ---------------------------------------------------

  MPLFigure.prototype._installCanvasEvents = function () {
    var self = this;
    var canvas = this.canvas;

    canvas.addEventListener("mousedown", function (e) {
      self._sendMouse("button_press", e);
      canvas.focus();
    });
    canvas.addEventListener("mouseup", function (e) {
      self._sendMouse("button_release", e);
    });
    canvas.addEventListener("mousemove", function (e) {
      self._sendMouse("motion_notify", e);
    });
    canvas.addEventListener("mouseenter", function (e) {
      self._sendMouse("figure_enter", e);
    });
    canvas.addEventListener("mouseleave", function (e) {
      self._sendMouse("figure_leave", e);
    });
    canvas.addEventListener("dblclick", function (e) {
      self._sendMouse("dblclick", e);
    });
    canvas.addEventListener(
      "wheel",
      function (e) {
        e.preventDefault();
        self._sendScroll(e);
      },
      { passive: false },
    );

    canvas.addEventListener("keydown", function (e) {
      self._sendKey("key_press", e);
    });
    canvas.addEventListener("keyup", function (e) {
      self._sendKey("key_release", e);
    });
  };

  MPLFigure.prototype._sendMouse = function (type, e) {
    var rect = this.canvas.getBoundingClientRect();
    var x = e.clientX - rect.left;
    var y = e.clientY - rect.top;
    this.send({
      type: type,
      x: x,
      y: y,
      button: e.button,
      buttons: e.buttons,
      modifiers: collectModifiers(e),
    });
  };

  MPLFigure.prototype._sendScroll = function (e) {
    var rect = this.canvas.getBoundingClientRect();
    // Match upstream convention: positive step zooms in.
    var step = e.deltaY < 0 ? 1 : -1;
    this.send({
      type: "scroll",
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
      step: step,
      modifiers: collectModifiers(e),
    });
  };

  MPLFigure.prototype._sendKey = function (type, e) {
    this.send({
      type: type,
      key: e.key,
      modifiers: collectModifiers(e),
    });
  };

  // ---- Rubberband ------------------------------------------------------

  MPLFigure.prototype._drawRubberband = function (x0, y0, x1, y1) {
    var c = this.octx;
    c.clearRect(0, 0, this.overlay.width, this.overlay.height);
    if (x0 < 0 || y0 < 0 || x1 < 0 || y1 < 0) return;
    c.strokeStyle = "#000";
    c.lineWidth = 1;
    c.setLineDash([4, 3]);
    c.strokeRect(x0, y0, x1 - x0, y1 - y0);
  };

  // ---- Helpers ---------------------------------------------------------

  function collectModifiers(e) {
    var out = [];
    if (e.shiftKey) out.push("shift");
    if (e.ctrlKey) out.push("ctrl");
    if (e.altKey) out.push("alt");
    if (e.metaKey) out.push("meta");
    return out;
  }

  global.MPLFigure = MPLFigure;
})(window);
