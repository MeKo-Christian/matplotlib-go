package test

import (
	"image"
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

func TestGoBasicSmokeCasesRenderNonBlank(t *testing.T) {
	for _, c := range goBasicSmokeCases() {
		c := c
		t.Run(c.ID, func(t *testing.T) {
			fig, _, err := parity.Figure(c.ID)
			if err != nil {
				t.Fatalf("GoBasic smoke case %s has no parity figure: %v", c.ID, err)
			}
			r := gobasic.New(int(fig.SizePx.X), int(fig.SizePx.Y), render.Color{R: 1, G: 1, B: 1, A: 1})
			core.DrawFigure(fig, r)
			if !hasNonWhitePixel(r.GetImage()) {
				t.Fatal("GoBasic smoke render is blank")
			}
		})
	}
}

func TestGoBasicSmokeFamiliesCoverPhase74StaticSurface(t *testing.T) {
	want := []string{
		"annotation",
		"axes",
		"axes_grid1",
		"axisartist",
		"line",
		"scatter",
		"bar",
		"colorbar",
		"fill",
		"errorbar",
		"geo",
		"histogram",
		"layout",
		"matrix",
		"mplot3d",
		"text",
		"mathtext",
		"image",
		"mesh",
		"collection",
		"patch_hatch",
		"polar",
		"signal",
		"specialty",
		"statistics",
		"units",
		"unstructured",
		"variants",
		"vectors",
	}
	got := map[string]string{}
	for _, c := range goBasicSmokeCases() {
		got[c.GoBasicSmokeFamily] = c.ID
	}
	for _, family := range want {
		if got[family] == "" {
			t.Fatalf("missing GoBasic smoke family %q", family)
		}
	}
}

func goBasicSmokeCases() []examplecatalog.Case {
	out := []examplecatalog.Case{}
	for _, c := range allCases() {
		if c.GoBasicSmokeFamily != "" {
			out = append(out, c)
		}
	}
	return out
}

func hasNonWhitePixel(img image.Image) bool {
	if img == nil {
		return false
	}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if color.RGBAModel.Convert(img.At(x, y)).(color.RGBA) != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
				return true
			}
		}
	}
	return false
}
