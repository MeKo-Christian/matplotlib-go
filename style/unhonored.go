package style

import (
	"slices"
	"sort"
	"sync"

	"github.com/cwbudde/matplotlib-go/internal/diag"
)

// This file implements PLAN.md Phase 16 ("rcParams Honesty & Coverage"): of the
// rcParams matplotlib-go parses, a sizeable subset is stored on RC but never
// read by any drawing or backend code, so setting them has no effect. That
// silent no-op gives callers porting real matplotlib scripts false confidence a
// setting took effect.
//
// unhonoredRCParams is the audited source of truth for that subset. Each entry
// reports whether the parsed RC value differs from the library Default for its
// field(s); applyMPLStyleEntry consults it on every successfully-applied entry
// and emits a one-shot warning the first time such a param is set to a
// non-default value, turning the silence into a signal.
//
// An entry here is an explicit "parsed but not honored" assertion. When a param
// is later wired into the drawing/backend code, delete its entry so the warning
// stops firing (and add a consumption read; see style_test's
// TestUnhonoredRCParamsAreActuallyUnread guard).

// unhonoredRCParam describes one parsed-but-not-consumed rcParam.
type unhonoredRCParam struct {
	// differs reports whether the param's RC field(s) on rc differ from the
	// same field(s) on def (typically the library Default).
	differs func(rc, def *RC) bool
}

// unhonoredRCParams maps an rcParam key to its store-only descriptor. The keys
// must match the case labels in applyMPLStyleEntry exactly. Honored siblings
// (e.g. image.cmap, image.interpolation, mathtext.fontset, boxplot.notch,
// boxplot.patchartist) are intentionally absent.
var unhonoredRCParams = map[string]unhonoredRCParam{
	// image.* — cmap and interpolation are consumed in core/image_api.go;
	// origin and aspect in the imshow front-ends (core/matrix_helpers.go).
	// image.resample stays unhonored: the Go raster path has no adaptive
	// (downscale-aware) resampling engine to toggle — it structurally matches
	// matplotlib's resample=False branch only.
	"image.interpolation_stage": {differs: func(rc, def *RC) bool { return rc.Image.InterpolationStage != def.Image.InterpolationStage }},
	"image.resample":            {differs: func(rc, def *RC) bool { return rc.Image.Resample != def.Image.Resample }},
	"image.composite_image":     {differs: func(rc, def *RC) bool { return rc.Image.CompositeImage != def.Image.CompositeImage }},
	"image.lut":                 {differs: func(rc, def *RC) bool { return rc.Image.LUT != def.Image.LUT }},

	// mathtext.* — only fontset is consumed (core/mathtext.go). Honoring the
	// rest is blocked upstream: the cwbudde/mathtext engine exposes no hook
	// for the implicit-italic default or per-class family names (PLAN.md
	// Phase 16), and cal/bfit/fallback additionally need per-fontset glyph
	// maps (Phase 17).
	"mathtext.default":  {differs: func(rc, def *RC) bool { return rc.Mathtext.Default != def.Mathtext.Default }},
	"mathtext.fallback": {differs: func(rc, def *RC) bool { return rc.Mathtext.Fallback != def.Mathtext.Fallback }},
	"mathtext.bf":       {differs: func(rc, def *RC) bool { return rc.Mathtext.BF != def.Mathtext.BF }},
	"mathtext.bfit":     {differs: func(rc, def *RC) bool { return rc.Mathtext.BFit != def.Mathtext.BFit }},
	"mathtext.cal":      {differs: func(rc, def *RC) bool { return rc.Mathtext.Cal != def.Mathtext.Cal }},
	"mathtext.it":       {differs: func(rc, def *RC) bool { return rc.Mathtext.It != def.Mathtext.It }},
	"mathtext.rm":       {differs: func(rc, def *RC) bool { return rc.Mathtext.RM != def.Mathtext.RM }},
	"mathtext.sf":       {differs: func(rc, def *RC) bool { return rc.Mathtext.SF != def.Mathtext.SF }},
	"mathtext.tt":       {differs: func(rc, def *RC) bool { return rc.Mathtext.TT != def.Mathtext.TT }},

	// date.* — all consumed: autoformatter.* + interval_multiples + converter
	// in core/date_tick.go & core/units.go, epoch lazily via core.GetEpoch.

	// pdf.* — the PDF backend uses render.Config.PDF, not RC.PDF; none consumed.
	"pdf.fonttype":       {differs: func(rc, def *RC) bool { return rc.PDF.FontType != def.PDF.FontType }},
	"pdf.use14corefonts": {differs: func(rc, def *RC) bool { return rc.PDF.Use14CoreFonts != def.PDF.Use14CoreFonts }},
	"pdf.compression":    {differs: func(rc, def *RC) bool { return rc.PDF.Compression != def.PDF.Compression }},
	"pdf.inheritcolor":   {differs: func(rc, def *RC) bool { return rc.PDF.InheritColor != def.PDF.InheritColor }},

	// ps.* — the PS backend uses its own config; none consumed.
	"ps.fonttype":     {differs: func(rc, def *RC) bool { return rc.PS.FontType != def.PS.FontType }},
	"ps.useafm":       {differs: func(rc, def *RC) bool { return rc.PS.UseAFM != def.PS.UseAFM }},
	"ps.papersize":    {differs: func(rc, def *RC) bool { return rc.PS.PaperSize != def.PS.PaperSize }},
	"ps.usedistiller": {differs: func(rc, def *RC) bool { return rc.PS.UseDistiller != def.PS.UseDistiller }},
	"ps.distiller.res": {differs: func(rc, def *RC) bool {
		return rc.PS.DistillerRes != def.PS.DistillerRes
	}},

	// svg.* — the SVG backend uses render.Config.SVG; none consumed.
	"svg.fonttype":     {differs: func(rc, def *RC) bool { return rc.SVG.FontType != def.SVG.FontType }},
	"svg.image_inline": {differs: func(rc, def *RC) bool { return rc.SVG.ImageInline != def.SVG.ImageInline }},
	"svg.hashsalt":     {differs: func(rc, def *RC) bool { return rc.SVG.HashSalt != def.SVG.HashSalt }},
	"svg.id":           {differs: func(rc, def *RC) bool { return rc.SVG.ID != def.SVG.ID }},

	// animation.* — the animation writers carry their own config; none consumed.
	"animation.html":         {differs: func(rc, def *RC) bool { return rc.Animation.HTML != def.Animation.HTML }},
	"animation.writer":       {differs: func(rc, def *RC) bool { return rc.Animation.Writer != def.Animation.Writer }},
	"animation.codec":        {differs: func(rc, def *RC) bool { return rc.Animation.Codec != def.Animation.Codec }},
	"animation.bitrate":      {differs: func(rc, def *RC) bool { return rc.Animation.Bitrate != def.Animation.Bitrate }},
	"animation.frame_format": {differs: func(rc, def *RC) bool { return rc.Animation.FrameFormat != def.Animation.FrameFormat }},
	"animation.ffmpeg_path":  {differs: func(rc, def *RC) bool { return rc.Animation.FFmpegPath != def.Animation.FFmpegPath }},
	"animation.ffmpeg_args":  {differs: func(rc, def *RC) bool { return !slices.Equal(rc.Animation.FFmpegArgs, def.Animation.FFmpegArgs) }},
	"animation.convert_path": {differs: func(rc, def *RC) bool { return rc.Animation.ConvertPath != def.Animation.ConvertPath }},
	"animation.convert_args": {differs: func(rc, def *RC) bool { return !slices.Equal(rc.Animation.ConvertArgs, def.Animation.ConvertArgs) }},
	"animation.embed_limit":  {differs: func(rc, def *RC) bool { return rc.Animation.EmbedLimit != def.Animation.EmbedLimit }},

	// boxplot.* — vertical and whiskers are stored but never read; the rest of
	// boxplot.* (notch, patchartist, show*, *props.*) is consumed in core/plot.go
	// and core/boxplot.go.
	"boxplot.vertical": {differs: func(rc, def *RC) bool { return rc.Boxplot.Vertical != def.Boxplot.Vertical }},
	"boxplot.whiskers": {differs: func(rc, def *RC) bool { return rc.Boxplot.Whiskers != def.Boxplot.Whiskers }},
}

// unhonoredRCParamKeys returns the registered store-only rcParam keys, sorted.
// It exists for the audit/guard tests.
func unhonoredRCParamKeys() []string {
	keys := make([]string, 0, len(unhonoredRCParams))
	for k := range unhonoredRCParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

var (
	unhonoredWarnMu sync.Mutex
	unhonoredWarned = map[string]bool{}
)

// maybeWarnUnhonoredRCParam checks whether key is a store-only rcParam and, if
// the just-applied value (on rc) differs from the library Default, emits a
// one-shot warning (the first time per key, process-global so repeated style
// application does not spam the log). Called from applyMPLStyleEntry after a
// successful apply.
func maybeWarnUnhonoredRCParam(key string, rc *RC) {
	p, ok := unhonoredRCParams[key]
	if !ok || rc == nil {
		return
	}
	if !p.differs(rc, &Default) {
		return
	}
	unhonoredWarnMu.Lock()
	seen := unhonoredWarned[key]
	unhonoredWarned[key] = true
	unhonoredWarnMu.Unlock()
	if seen {
		return
	}
	diag.Warnf("rcParam %q is recognized but not honored by matplotlib-go: "+
		"the value is stored but does not affect rendering", key)
}

// resetUnhonoredRCParamWarnings clears the one-shot dedup state. Tests use it to
// re-trigger warnings; it is not part of the public API.
func resetUnhonoredRCParamWarnings() {
	unhonoredWarnMu.Lock()
	unhonoredWarned = map[string]bool{}
	unhonoredWarnMu.Unlock()
}
