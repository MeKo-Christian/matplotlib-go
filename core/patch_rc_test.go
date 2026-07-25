package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestStandalonePatchUsesRCDefaultsAndExplicitPaintWins(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.Patch = style.PatchRC{
		LineWidth:      2.5,
		FaceColor:      render.Color{R: 1, A: 1},
		FaceColorRaw:   "",
		EdgeColor:      render.Color{B: 1, A: 1},
		ForceEdgeColor: true,
		Antialiased:    false,
	}
	rect := &Rectangle{Width: 1, Height: 1}
	rec := &recordingRenderer{}
	rect.Draw(rec, ctx)
	if len(rec.pathCalls) != 1 {
		t.Fatalf("path calls = %d, want 1", len(rec.pathCalls))
	}
	paint := rec.pathCalls[0].paint
	if paint.Fill != ctx.RC.Patch.FaceColor || paint.Stroke != ctx.RC.Patch.EdgeColor ||
		paint.LineWidth != pointsToPixels(ctx.RC, 2.5) || paint.Antialias != render.AntialiasOff {
		t.Fatalf("default patch paint = %+v", paint)
	}

	explicitFace := render.Color{G: 1, A: 1}
	explicitEdge := render.Color{R: 1, G: 1, A: 1}
	rect.Patch = Patch{
		FaceColor: explicitFace,
		EdgeColor: explicitEdge,
		EdgeWidth: 4,
		Antialias: render.AntialiasOn,
	}
	rec.pathCalls = nil
	rect.Draw(rec, ctx)
	paint = rec.pathCalls[0].paint
	if paint.Fill != explicitFace || paint.Stroke != explicitEdge ||
		paint.LineWidth != pointsToPixels(ctx.RC, 4) || paint.Antialias != render.AntialiasOn {
		t.Fatalf("explicit patch paint = %+v", paint)
	}
}

func TestStandalonePatchC0TracksCurrentColorCycle(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.Patch.FaceColorRaw = "C0"
	ctx.RC.ColorCycle = color.Palette{{G: 1, B: 1, A: 1}}
	rec := &recordingRenderer{}
	(&Rectangle{Width: 1, Height: 1}).Draw(rec, ctx)
	if got := rec.pathCalls[0].paint.Fill; got != ctx.RC.ColorCycle[0] {
		t.Fatalf("patch C0 fill = %+v, want current cycle %+v", got, ctx.RC.ColorCycle[0])
	}
}

func TestStandalonePatchExplicitTransparentAndZeroOverrideRC(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.Patch = style.PatchRC{
		LineWidth:      5,
		FaceColor:      render.Color{R: 1, A: 1},
		EdgeColor:      render.Color{B: 1, A: 1},
		ForceEdgeColor: true,
		Antialiased:    true,
	}
	rect := &Rectangle{Width: 1, Height: 1}
	rect.SetFaceColor(render.Color{})
	rect.SetEdgeColor(render.Color{})
	rect.SetEdgeWidth(0)
	rec := &recordingRenderer{}
	rect.Draw(rec, ctx)
	if len(rec.pathCalls) != 0 {
		t.Fatalf("explicit transparent/zero patch drew %d paths", len(rec.pathCalls))
	}
}

func TestPatchLegendEntryUsesBoundRCDefaults(t *testing.T) {
	fig := NewFigure(100, 100)
	fig.RC.Patch.FaceColorRaw = ""
	fig.RC.Patch.FaceColor = render.Color{G: 1, A: 1}
	fig.RC.Patch.EdgeColor = render.Color{B: 1, A: 1}
	fig.RC.Patch.LineWidth = 2.25
	fig.RC.Patch.ForceEdgeColor = true
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	rect := &Rectangle{Patch: Patch{Label: "rc"}, Width: 1, Height: 1}
	ax.AddPatch(rect)
	entry, ok := rect.legendEntry()
	if !ok {
		t.Fatal("legendEntry() = false")
	}
	if entry.patchFill != fig.RC.Patch.FaceColor ||
		entry.patchEdge != fig.RC.Patch.EdgeColor ||
		entry.patchEdgeWidth != fig.RC.Patch.LineWidth {
		t.Fatalf("legend patch = fill %+v edge %+v width %v", entry.patchFill, entry.patchEdge, entry.patchEdgeWidth)
	}
}

func TestPatchProducingMethodsHonorRCAndExplicitZeroWidth(t *testing.T) {
	fig := NewFigure(200, 150)
	fig.RC.Patch.LineWidth = 3
	fig.RC.Patch.EdgeColor = render.Color{R: 1, A: 1}
	fig.RC.Patch.ForceEdgeColor = true
	fig.RC.Patch.Antialiased = false
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})

	bar, err := ax.Bar([]float64{0}, []float64{1})
	if err != nil {
		t.Fatalf("Bar() returned error: %v", err)
	}
	fill, err := ax.FillBetween([]float64{0, 1}, []float64{0, 0}, []float64{1, 1})
	if err != nil {
		t.Fatalf("FillBetween() returned error: %v", err)
	}
	hist := ax.Hist([]float64{0, 1})
	span := ax.AxHSpan(0, 1)
	if bar.EdgeWidth != 3 || bar.EdgeColor != fig.RC.Patch.EdgeColor ||
		bar.Antialias != render.AntialiasOff {
		t.Fatalf("bar defaults = edge %+v width %v aa %v", bar.EdgeColor, bar.EdgeWidth, bar.Antialias)
	}
	if fill.EdgeWidth != 3 || fill.EdgeColor != fig.RC.Patch.EdgeColor ||
		fill.Antialias != render.AntialiasOff {
		t.Fatalf("fill defaults = edge %+v width %v aa %v", fill.EdgeColor, fill.EdgeWidth, fill.Antialias)
	}
	if hist.EdgeWidth != 3 || hist.EdgeColor != fig.RC.Patch.EdgeColor ||
		hist.Antialias != render.AntialiasOff {
		t.Fatalf("hist defaults = edge %+v width %v aa %v", hist.EdgeColor, hist.EdgeWidth, hist.Antialias)
	}
	if span.Color != fig.RC.DefaultPatchFaceColor() ||
		span.EdgeColor != fig.RC.Patch.EdgeColor || span.EdgeWidth != 3 ||
		span.Antialias != render.AntialiasOff {
		t.Fatalf("span defaults = fill %+v edge %+v width %v aa %v", span.Color, span.EdgeColor, span.EdgeWidth, span.Antialias)
	}

	zero := 0.0
	bar, err = ax.Bar([]float64{0}, []float64{1}, BarOptions{EdgeWidth: &zero})
	if err != nil {
		t.Fatalf("Bar() returned error: %v", err)
	}
	fill, err = ax.FillBetween(
		[]float64{0, 1}, []float64{0, 0}, []float64{1, 1},
		FillOptions{EdgeWidth: &zero},
	)
	if err != nil {
		t.Fatalf("FillBetween() returned error: %v", err)
	}
	hist = ax.Hist([]float64{0, 1}, HistOptions{EdgeWidth: &zero})
	span = ax.AxHSpan(0, 1, HSpanOptions{EdgeWidth: &zero})
	if bar.EdgeWidth != 0 || fill.EdgeWidth != 0 || hist.EdgeWidth != 0 || span.EdgeWidth != 0 {
		t.Fatalf("explicit zero widths = bar %v fill %v hist %v span %v", bar.EdgeWidth, fill.EdgeWidth, hist.EdgeWidth, span.EdgeWidth)
	}
}
