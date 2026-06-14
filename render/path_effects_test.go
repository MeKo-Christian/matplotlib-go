package render

import (
	"image"
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestDrawPathWithEffectsReplaysStrokeThenNormal(t *testing.T) {
	path := testEffectLinePath()
	paint := Paint{
		Stroke:    Color{B: 1, A: 1},
		LineWidth: 1,
		PathEffects: WithStrokePathEffects(
			Color{R: 1, A: 1},
			4,
			geom.Pt{X: 2, Y: 3},
		),
	}

	var draws []effectDraw
	if !DrawPathWithEffects(&NullRenderer{}, path, &paint, recordEffectDraw(&draws)) {
		t.Fatal("DrawPathWithEffects returned false")
	}
	if len(draws) != 2 {
		t.Fatalf("draw count = %d, want 2", len(draws))
	}
	if got := draws[0].paint.Stroke; got != (Color{R: 1, A: 1}) {
		t.Fatalf("stroke effect color = %+v", got)
	}
	if draws[0].paint.LineWidth != 4 {
		t.Fatalf("stroke effect linewidth = %v, want 4", draws[0].paint.LineWidth)
	}
	if len(draws[0].paint.PathEffects) != 0 {
		t.Fatalf("effect replay retained nested effects: %+v", draws[0].paint.PathEffects)
	}
	if got := draws[0].path.V[0]; got != (geom.Pt{X: 2, Y: 3}) {
		t.Fatalf("offset first vertex = %+v", got)
	}
	if got := draws[1].path.V[0]; got != (geom.Pt{}) {
		t.Fatalf("normal first vertex = %+v", got)
	}
	if got := draws[1].paint.Stroke; got != (Color{B: 1, A: 1}) {
		t.Fatalf("normal color = %+v", got)
	}
}

func TestDrawPathWithEffectsDerivesPatchShadowColor(t *testing.T) {
	path := testEffectLinePath()
	path.Close()
	paint := Paint{
		Fill: Color{R: 0.8, G: 0.4, B: 0.2, A: 1},
		PathEffects: []PathEffect{
			SimplePatchShadowPathEffect(geom.Pt{X: 1, Y: 1}, Color{}, 0.5, 0.25),
		},
	}

	var draws []effectDraw
	DrawPathWithEffects(&NullRenderer{}, path, &paint, recordEffectDraw(&draws))
	if len(draws) != 1 {
		t.Fatalf("draw count = %d, want 1", len(draws))
	}
	want := Color{R: 0.2, G: 0.1, B: 0.05, A: 0.5}
	if got := draws[0].paint.Fill; got != want {
		t.Fatalf("shadow fill = %+v, want %+v", got, want)
	}
	if draws[0].paint.Stroke.A != 0 {
		t.Fatalf("patch shadow unexpectedly stroked: %+v", draws[0].paint.Stroke)
	}
}

func TestDrawPathWithEffectsBuildsTickedStroke(t *testing.T) {
	var path geom.Path
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.LineTo(geom.Pt{X: 30, Y: 0})
	paint := Paint{
		Stroke:    Color{A: 1},
		LineWidth: 1,
		PathEffects: []PathEffect{
			TickedStrokePathEffect(Color{G: 1, A: 1}, 2, 10, 90, 1, geom.Pt{}),
		},
	}

	var draws []effectDraw
	DrawPathWithEffects(&NullRenderer{}, path, &paint, recordEffectDraw(&draws))
	if len(draws) != 1 {
		t.Fatalf("draw count = %d, want 1", len(draws))
	}
	if len(draws[0].path.C) != 6 {
		t.Fatalf("tick command count = %d, want 6", len(draws[0].path.C))
	}
	if got := draws[0].path.V[0]; got != (geom.Pt{X: 5, Y: 0}) {
		t.Fatalf("first tick start = %+v", got)
	}
	if got := draws[0].path.V[1]; math.Abs(got.X-5) > 1e-9 || got.Y >= 0 {
		t.Fatalf("first tick end = %+v, want same x and negative y", got)
	}
	if got := draws[0].paint.Stroke; got != (Color{G: 1, A: 1}) {
		t.Fatalf("tick stroke = %+v", got)
	}
}

func TestDrawPathWithEffectsUsesFilterRenderer(t *testing.T) {
	path := testEffectLinePath()
	path.Close()
	paint := Paint{
		Fill: Color{B: 1, A: 1},
		PathEffects: []PathEffect{
			FilterPathEffect(Color{R: 1, A: 1}, Color{}, 0, "blur", 1, geom.Pt{X: 2, Y: 0}),
			NormalPathEffect(),
		},
	}

	var draws []effectDraw
	renderer := &effectFilterRenderer{}
	if !DrawPathWithEffects(renderer, path, &paint, recordEffectDraw(&draws)) {
		t.Fatal("DrawPathWithEffects returned false")
	}
	if renderer.starts != 1 || renderer.stops != 1 {
		t.Fatalf("filter start/stop = %d/%d, want 1/1", renderer.starts, renderer.stops)
	}
	if !renderer.postProcessed {
		t.Fatal("filter post processor was not invoked")
	}
	if len(draws) != 2 {
		t.Fatalf("draw count = %d, want 2", len(draws))
	}
	if got := draws[0].path.V[0]; got != (geom.Pt{X: 2, Y: 0}) {
		t.Fatalf("filter pass offset first vertex = %+v", got)
	}
	if got := draws[0].paint.Fill; got != (Color{R: 1, A: 1}) {
		t.Fatalf("filter pass fill = %+v", got)
	}
	if got := draws[1].paint.Fill; got != (Color{B: 1, A: 1}) {
		t.Fatalf("normal pass fill = %+v", got)
	}
}

type effectDraw struct {
	path  geom.Path
	paint Paint
}

type effectFilterRenderer struct {
	NullRenderer
	starts        int
	stops         int
	postProcessed bool
}

func (r *effectFilterRenderer) StartFilter() {
	r.starts++
}

func (r *effectFilterRenderer) StopFilter(postProcess func(*image.RGBA, float64) (*image.RGBA, geom.Pt)) {
	r.stops++
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	img.Pix[img.PixOffset(1, 1)+3] = 255
	out, _ := postProcess(img, 72)
	r.postProcessed = out != nil && out.Pix[out.PixOffset(1, 1)+3] > 0
}

func recordEffectDraw(draws *[]effectDraw) func(geom.Path, *Paint) {
	return func(path geom.Path, paint *Paint) {
		*draws = append(*draws, effectDraw{
			path:  path,
			paint: *paint,
		})
	}
}

func testEffectLinePath() geom.Path {
	var path geom.Path
	path.MoveTo(geom.Pt{})
	path.LineTo(geom.Pt{X: 10})
	return path
}
