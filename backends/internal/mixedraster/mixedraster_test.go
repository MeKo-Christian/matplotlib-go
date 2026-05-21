package mixedraster

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestStartScalesRasterSurfaceByDPI(t *testing.T) {
	session, ok := Start(
		20,
		10,
		geom.Rect{Max: geom.Pt{X: 20, Y: 10}},
		render.Rasterization{Mode: render.RasterizeAlways, DPI: 144},
		72,
		nil,
		nil,
	)
	if !ok {
		t.Fatal("Start failed")
	}

	var path geom.Path
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.LineTo(geom.Pt{X: 20, Y: 0})
	path.LineTo(geom.Pt{X: 20, Y: 10})
	path.LineTo(geom.Pt{X: 0, Y: 10})
	path.Close()
	session.Renderer().Path(path, &render.Paint{Fill: render.Color{R: 1, A: 1}})

	img, rect, ok := session.Stop()
	if !ok {
		t.Fatal("Stop failed")
	}
	if gotW, gotH := img.Size(); gotW != 40 || gotH != 20 {
		t.Fatalf("raster image size = %dx%d, want 40x20", gotW, gotH)
	}
	if rect.W() != 20 || rect.H() != 10 {
		t.Fatalf("placement rect = %v, want original 20x10 display units", rect)
	}
	alpha := img.RGBA().RGBAAt(30, 10).A
	if alpha == 0 {
		t.Fatal("scaled path did not fill high-DPI right half")
	}
}
