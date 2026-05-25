package core

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestImReadDecodesPNGAsImageData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.png")
	src := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	src.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
	src.SetNRGBA(1, 0, color.NRGBA{R: 40, G: 50, B: 60, A: 128})
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := png.Encode(file, src); err != nil {
		_ = file.Close()
		t.Fatalf("Encode() error = %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	img, err := ImRead(path)
	if err != nil {
		t.Fatalf("ImRead() error = %v", err)
	}
	if w, h := img.Size(); w != 2 || h != 1 {
		t.Fatalf("decoded size = %dx%d, want 2x1", w, h)
	}
	if got := img.RGBA().RGBAAt(1, 0); got.R != 40 || got.G != 50 || got.B != 60 || got.A != 128 {
		t.Fatalf("decoded pixel = %+v, want rgba(40,50,60,128)", got)
	}
}

func TestImSaveWritesPNGImageData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output.png")
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.SetRGBA(0, 0, color.RGBA{R: 12, G: 34, B: 56, A: 200})

	if err := ImSave(path, render.NewImageData(src)); err != nil {
		t.Fatalf("ImSave() error = %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer file.Close()
	decoded, err := png.Decode(file)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got := color.NRGBAModel.Convert(decoded.At(0, 0)).(color.NRGBA); got.R != 12 || got.G != 34 || got.B != 56 || got.A != 200 {
		t.Fatalf("saved pixel = %+v, want rgba(12,34,56,200)", got)
	}
}

func TestImSaveReportsUnsupportedInputs(t *testing.T) {
	if err := ImSave(filepath.Join(t.TempDir(), "out.jpg"), render.NewImageData(image.NewRGBA(image.Rect(0, 0, 1, 1)))); !errors.Is(err, ErrImageIOUnsupported) {
		t.Fatalf("ImSave jpg error = %v, want ErrImageIOUnsupported", err)
	}
	if err := ImSave(filepath.Join(t.TempDir(), "out.png"), nil); !errors.Is(err, ErrImageIOUnsupported) {
		t.Fatalf("ImSave nil error = %v, want ErrImageIOUnsupported", err)
	}
}
