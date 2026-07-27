package style

import "github.com/cwbudde/matplotlib-go/internal/diag"

// This file sets the minimum bar for unparsed-but-known rcParam keys: a key
// that matplotlib 3.10.9 accepts but matplotlib-go
// does not parse used to land silently in MPLStyleReport.Unsupported —
// indistinguishable from a typo. applyMPLStyleEntry now consults
// knownUpstreamRCParams on that fallthrough path and emits a one-shot warning,
// so silence means "genuinely unknown key", never "known key we ignore".
//
// A subset of those keys is intentionally unsupported (nonGoalRCParams): they
// configure behavior this port pins for parity or does not model at all. Their
// warning carries the rationale instead of the generic "not parsed" text.

// knownUpstreamRCParams is the complete user-facing rcParam key set of
// matplotlib 3.10.9 (sorted per family), generated from
// `sorted(matplotlib.rcParams)` of the pinned reference version with the
// internal `_internal.classic_mode` key removed. Keys that matplotlib-go
// parses also appear here (they are handled by their applyMPLStyleEntry case
// and never reach the fallthrough lookup); guard tests in unparsed_test.go
// keep the two sets consistent.
var knownUpstreamRCParams = map[string]struct{}{
	"agg.path.chunksize": {},

	"animation.bitrate":      {},
	"animation.codec":        {},
	"animation.convert_args": {},
	"animation.convert_path": {},
	"animation.embed_limit":  {},
	"animation.ffmpeg_args":  {},
	"animation.ffmpeg_path":  {},
	"animation.frame_format": {},
	"animation.html":         {},
	"animation.writer":       {},

	"axes.autolimit_mode":             {},
	"axes.axisbelow":                  {},
	"axes.edgecolor":                  {},
	"axes.facecolor":                  {},
	"axes.formatter.limits":           {},
	"axes.formatter.min_exponent":     {},
	"axes.formatter.offset_threshold": {},
	"axes.formatter.use_locale":       {},
	"axes.formatter.use_mathtext":     {},
	"axes.formatter.useoffset":        {},
	"axes.grid":                       {},
	"axes.grid.axis":                  {},
	"axes.grid.which":                 {},
	"axes.labelcolor":                 {},
	"axes.labelpad":                   {},
	"axes.labelsize":                  {},
	"axes.labelweight":                {},
	"axes.linewidth":                  {},
	"axes.prop_cycle":                 {},
	"axes.spines.bottom":              {},
	"axes.spines.left":                {},
	"axes.spines.right":               {},
	"axes.spines.top":                 {},
	"axes.titlecolor":                 {},
	"axes.titlelocation":              {},
	"axes.titlepad":                   {},
	"axes.titlesize":                  {},
	"axes.titleweight":                {},
	"axes.titley":                     {},
	"axes.unicode_minus":              {},
	"axes.xmargin":                    {},
	"axes.ymargin":                    {},
	"axes.zmargin":                    {},

	"axes3d.automargin":         {},
	"axes3d.grid":               {},
	"axes3d.mouserotationstyle": {},
	"axes3d.trackballborder":    {},
	"axes3d.trackballsize":      {},
	"axes3d.xaxis.panecolor":    {},
	"axes3d.yaxis.panecolor":    {},
	"axes3d.zaxis.panecolor":    {},

	"backend": {},

	"backend_fallback": {},

	"boxplot.bootstrap":                  {},
	"boxplot.boxprops.color":             {},
	"boxplot.boxprops.linestyle":         {},
	"boxplot.boxprops.linewidth":         {},
	"boxplot.capprops.color":             {},
	"boxplot.capprops.linestyle":         {},
	"boxplot.capprops.linewidth":         {},
	"boxplot.flierprops.color":           {},
	"boxplot.flierprops.linestyle":       {},
	"boxplot.flierprops.linewidth":       {},
	"boxplot.flierprops.marker":          {},
	"boxplot.flierprops.markeredgecolor": {},
	"boxplot.flierprops.markeredgewidth": {},
	"boxplot.flierprops.markerfacecolor": {},
	"boxplot.flierprops.markersize":      {},
	"boxplot.meanline":                   {},
	"boxplot.meanprops.color":            {},
	"boxplot.meanprops.linestyle":        {},
	"boxplot.meanprops.linewidth":        {},
	"boxplot.meanprops.marker":           {},
	"boxplot.meanprops.markeredgecolor":  {},
	"boxplot.meanprops.markerfacecolor":  {},
	"boxplot.meanprops.markersize":       {},
	"boxplot.medianprops.color":          {},
	"boxplot.medianprops.linestyle":      {},
	"boxplot.medianprops.linewidth":      {},
	"boxplot.notch":                      {},
	"boxplot.patchartist":                {},
	"boxplot.showbox":                    {},
	"boxplot.showcaps":                   {},
	"boxplot.showfliers":                 {},
	"boxplot.showmeans":                  {},
	"boxplot.vertical":                   {},
	"boxplot.whiskerprops.color":         {},
	"boxplot.whiskerprops.linestyle":     {},
	"boxplot.whiskerprops.linewidth":     {},
	"boxplot.whiskers":                   {},

	"contour.algorithm":          {},
	"contour.corner_mask":        {},
	"contour.linewidth":          {},
	"contour.negative_linestyle": {},

	"date.autoformatter.day":         {},
	"date.autoformatter.hour":        {},
	"date.autoformatter.microsecond": {},
	"date.autoformatter.minute":      {},
	"date.autoformatter.month":       {},
	"date.autoformatter.second":      {},
	"date.autoformatter.year":        {},
	"date.converter":                 {},
	"date.epoch":                     {},
	"date.interval_multiples":        {},

	"docstring.hardcopy": {},

	"errorbar.capsize": {},

	"figure.autolayout":                {},
	"figure.constrained_layout.h_pad":  {},
	"figure.constrained_layout.hspace": {},
	"figure.constrained_layout.use":    {},
	"figure.constrained_layout.w_pad":  {},
	"figure.constrained_layout.wspace": {},
	"figure.dpi":                       {},
	"figure.edgecolor":                 {},
	"figure.facecolor":                 {},
	"figure.figsize":                   {},
	"figure.frameon":                   {},
	"figure.hooks":                     {},
	"figure.labelsize":                 {},
	"figure.labelweight":               {},
	"figure.max_open_warning":          {},
	"figure.raise_window":              {},
	"figure.subplot.bottom":            {},
	"figure.subplot.hspace":            {},
	"figure.subplot.left":              {},
	"figure.subplot.right":             {},
	"figure.subplot.top":               {},
	"figure.subplot.wspace":            {},
	"figure.titlesize":                 {},
	"figure.titleweight":               {},

	"font.cursive":    {},
	"font.family":     {},
	"font.fantasy":    {},
	"font.monospace":  {},
	"font.sans-serif": {},
	"font.serif":      {},
	"font.size":       {},
	"font.stretch":    {},
	"font.style":      {},
	"font.variant":    {},
	"font.weight":     {},

	"grid.alpha":     {},
	"grid.color":     {},
	"grid.linestyle": {},
	"grid.linewidth": {},

	"hatch.color":     {},
	"hatch.linewidth": {},

	"hist.bins": {},

	"image.aspect":              {},
	"image.cmap":                {},
	"image.composite_image":     {},
	"image.interpolation":       {},
	"image.interpolation_stage": {},
	"image.lut":                 {},
	"image.origin":              {},
	"image.resample":            {},

	"interactive": {},

	"keymap.back":       {},
	"keymap.copy":       {},
	"keymap.forward":    {},
	"keymap.fullscreen": {},
	"keymap.grid":       {},
	"keymap.grid_minor": {},
	"keymap.help":       {},
	"keymap.home":       {},
	"keymap.pan":        {},
	"keymap.quit":       {},
	"keymap.quit_all":   {},
	"keymap.save":       {},
	"keymap.xscale":     {},
	"keymap.yscale":     {},
	"keymap.zoom":       {},

	"legend.borderaxespad":  {},
	"legend.borderpad":      {},
	"legend.columnspacing":  {},
	"legend.edgecolor":      {},
	"legend.facecolor":      {},
	"legend.fancybox":       {},
	"legend.fontsize":       {},
	"legend.framealpha":     {},
	"legend.frameon":        {},
	"legend.handleheight":   {},
	"legend.handlelength":   {},
	"legend.handletextpad":  {},
	"legend.labelcolor":     {},
	"legend.labelspacing":   {},
	"legend.loc":            {},
	"legend.markerscale":    {},
	"legend.numpoints":      {},
	"legend.scatterpoints":  {},
	"legend.shadow":         {},
	"legend.title_fontsize": {},

	"lines.antialiased":     {},
	"lines.color":           {},
	"lines.dash_capstyle":   {},
	"lines.dash_joinstyle":  {},
	"lines.dashdot_pattern": {},
	"lines.dashed_pattern":  {},
	"lines.dotted_pattern":  {},
	"lines.linestyle":       {},
	"lines.linewidth":       {},
	"lines.marker":          {},
	"lines.markeredgecolor": {},
	"lines.markeredgewidth": {},
	"lines.markerfacecolor": {},
	"lines.markersize":      {},
	"lines.scale_dashes":    {},
	"lines.solid_capstyle":  {},
	"lines.solid_joinstyle": {},

	"macosx.window_mode": {},

	"markers.fillstyle": {},

	"mathtext.bf":       {},
	"mathtext.bfit":     {},
	"mathtext.cal":      {},
	"mathtext.default":  {},
	"mathtext.fallback": {},
	"mathtext.fontset":  {},
	"mathtext.it":       {},
	"mathtext.rm":       {},
	"mathtext.sf":       {},
	"mathtext.tt":       {},

	"patch.antialiased":     {},
	"patch.edgecolor":       {},
	"patch.facecolor":       {},
	"patch.force_edgecolor": {},
	"patch.linewidth":       {},

	"path.effects":            {},
	"path.simplify":           {},
	"path.simplify_threshold": {},
	"path.sketch":             {},
	"path.snap":               {},

	"pcolor.shading": {},

	"pcolormesh.snap": {},

	"pdf.compression":    {},
	"pdf.fonttype":       {},
	"pdf.inheritcolor":   {},
	"pdf.use14corefonts": {},

	"pgf.preamble":  {},
	"pgf.rcfonts":   {},
	"pgf.texsystem": {},

	"polaraxes.grid": {},

	"ps.distiller.res": {},
	"ps.fonttype":      {},
	"ps.papersize":     {},
	"ps.useafm":        {},
	"ps.usedistiller":  {},

	"savefig.bbox":        {},
	"savefig.directory":   {},
	"savefig.dpi":         {},
	"savefig.edgecolor":   {},
	"savefig.facecolor":   {},
	"savefig.format":      {},
	"savefig.orientation": {},
	"savefig.pad_inches":  {},
	"savefig.transparent": {},

	"scatter.edgecolors": {},
	"scatter.marker":     {},

	"svg.fonttype":     {},
	"svg.hashsalt":     {},
	"svg.id":           {},
	"svg.image_inline": {},

	"text.antialiased":    {},
	"text.color":          {},
	"text.hinting":        {},
	"text.hinting_factor": {},
	"text.kerning_factor": {},
	"text.latex.preamble": {},
	"text.parse_math":     {},
	"text.usetex":         {},

	"timezone": {},

	"tk.window_focus": {},

	"toolbar": {},

	"webagg.address":         {},
	"webagg.open_in_browser": {},
	"webagg.port":            {},
	"webagg.port_retries":    {},

	"xaxis.labellocation": {},

	"xtick.alignment":     {},
	"xtick.bottom":        {},
	"xtick.color":         {},
	"xtick.direction":     {},
	"xtick.labelbottom":   {},
	"xtick.labelcolor":    {},
	"xtick.labelsize":     {},
	"xtick.labeltop":      {},
	"xtick.major.bottom":  {},
	"xtick.major.pad":     {},
	"xtick.major.size":    {},
	"xtick.major.top":     {},
	"xtick.major.width":   {},
	"xtick.minor.bottom":  {},
	"xtick.minor.ndivs":   {},
	"xtick.minor.pad":     {},
	"xtick.minor.size":    {},
	"xtick.minor.top":     {},
	"xtick.minor.visible": {},
	"xtick.minor.width":   {},
	"xtick.top":           {},

	"yaxis.labellocation": {},

	"ytick.alignment":     {},
	"ytick.color":         {},
	"ytick.direction":     {},
	"ytick.labelcolor":    {},
	"ytick.labelleft":     {},
	"ytick.labelright":    {},
	"ytick.labelsize":     {},
	"ytick.left":          {},
	"ytick.major.left":    {},
	"ytick.major.pad":     {},
	"ytick.major.right":   {},
	"ytick.major.size":    {},
	"ytick.major.width":   {},
	"ytick.minor.left":    {},
	"ytick.minor.ndivs":   {},
	"ytick.minor.pad":     {},
	"ytick.minor.right":   {},
	"ytick.minor.size":    {},
	"ytick.minor.visible": {},
	"ytick.minor.width":   {},
	"ytick.right":         {},
}

// nonGoalRCParams maps rcParam keys matplotlib-go intentionally does not
// support to the rationale reported in their one-shot warning. These are the
// the "document, don't parse" non-goals: parsing them would suggest the
// behavior is configurable when it deliberately is not.
var nonGoalRCParams = map[string]string{
	"path.snap": "pixel snapping is baked into the renderers to byte-match matplotlib output; a global toggle would break golden parity",

	"figure.hooks":            "Python import hooks have no Go or headless-rendering equivalent and are not executed",
	"figure.max_open_warning": "the headless pyplot facade does not own GUI figure-manager resources, so an open-window warning threshold is not applicable",
	"figure.raise_window":     "raising a GUI window is not applicable to the headless renderer API",

	"text.hinting":        "text parity statically links FreeType 2.6.1 with matplotlib's autohinting setup; alternative hinting modes are not configurable",
	"text.hinting_factor": "text parity pins matplotlib's hinting_factor=8; other factors would break glyph-level parity",

	"polaraxes.grid": "polar axes draw their grid through the regular axes.grid machinery; use the grid API or axes.grid instead",

	"axes3d.grid":            "3D grid visibility is fixed to the matplotlib default; not configurable per rcParam",
	"axes3d.automargin":      "3D margin behavior is fixed to the matplotlib default; not configurable per rcParam",
	"axes3d.xaxis.panecolor": "3D pane styling is fixed to the matplotlib defaults; not configurable per rcParam",
	"axes3d.yaxis.panecolor": "3D pane styling is fixed to the matplotlib defaults; not configurable per rcParam",
	"axes3d.zaxis.panecolor": "3D pane styling is fixed to the matplotlib defaults; not configurable per rcParam",

	"axes3d.mouserotationstyle": "interactive 3D mouse navigation is not part of this library",
	"axes3d.trackballborder":    "interactive 3D mouse navigation is not part of this library",
	"axes3d.trackballsize":      "interactive 3D mouse navigation is not part of this library",
}

// maybeWarnUnparsedRCParam handles the applyMPLStyleEntry fallthrough: key was
// not matched by any parse case. If the key is a known matplotlib 3.10.9
// rcParam, emit a one-shot warning (deduped process-global, sharing the
// unhonored-param dedup state) — with the non-goal rationale when one is
// registered. Unknown keys stay silent; they are still recorded in
// MPLStyleReport.Unsupported by the caller.
func maybeWarnUnparsedRCParam(key string) {
	if _, known := knownUpstreamRCParams[key]; !known {
		return
	}
	unhonoredWarnMu.Lock()
	seen := unhonoredWarned[key]
	unhonoredWarned[key] = true
	unhonoredWarnMu.Unlock()
	if seen {
		return
	}
	if rationale, ok := nonGoalRCParams[key]; ok {
		diag.Warnf("rcParam %q is intentionally not supported by matplotlib-go: %s", key, rationale)
		return
	}
	diag.Warnf("rcParam %q is a known matplotlib rcParam but is not parsed by matplotlib-go: the setting is ignored", key)
}
