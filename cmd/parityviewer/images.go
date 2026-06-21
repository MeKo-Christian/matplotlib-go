package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

type metrics struct {
	RMSE        float64
	AvgDiff     float64
	MaxDiff     uint8
	DiffPixels  int
	TotalPixels int
	DiffRatio   float64
}

func buildEntry(suite, baseline, name, baselinePath, artifactPath string) (caseEntry, error) {
	cacheKey, hasCacheKey := newCaseEntryCacheKey(baselinePath, artifactPath)
	if hasCacheKey {
		if cached, ok := defaultCaseEntryCache.lookup(cacheKey); ok {
			return caseEntryFromCached(suite, baseline, name, cached), nil
		}
	}

	ref, err := readPNGAsRGBA(baselinePath)
	if err != nil {
		return caseEntry{}, fmt.Errorf("read baseline: %w", err)
	}
	act, err := readPNGAsRGBA(artifactPath)
	if err != nil {
		return caseEntry{}, fmt.Errorf("read artifact: %w", err)
	}

	stats := compareImages(ref, act)
	cached := cachedCaseEntry{
		Stats:     stats,
		RefWidth:  ref.Bounds().Dx(),
		RefHeight: ref.Bounds().Dy(),
		ActWidth:  act.Bounds().Dx(),
		ActHeight: act.Bounds().Dy(),
	}
	if hasCacheKey {
		defaultCaseEntryCache.store(cacheKey, cached)
	}

	return caseEntryFromCached(suite, baseline, name, cached), nil
}

func caseEntryFromCached(suite, baseline, name string, cached cachedCaseEntry) caseEntry {
	return caseEntry{
		Suite:       suite,
		Baseline:    baseline,
		Name:        name,
		RMSE:        cached.Stats.RMSE,
		AvgDiff:     cached.Stats.AvgDiff,
		MaxDiff:     cached.Stats.MaxDiff,
		DiffPixels:  cached.Stats.DiffPixels,
		TotalPixels: cached.Stats.TotalPixels,
		DiffRatio:   cached.Stats.DiffRatio,
		RefWidth:    cached.RefWidth,
		RefHeight:   cached.RefHeight,
		ActWidth:    cached.ActWidth,
		ActHeight:   cached.ActHeight,
		RefImageURL: imageURL(suite, baseline, name, "baseline"),
		ActImageURL: imageURL(suite, baseline, name, "artifact"),
		RawDiffURL:  imageURL(suite, baseline, name, "diff-raw"),
		AmpDiffURL:  imageURL(suite, baseline, name, "diff-amp"),
	}
}

func compareImages(ref, act *image.RGBA) metrics {
	bounds := unionBounds(ref.Bounds(), act.Bounds())
	totalPixels := bounds.Dx() * bounds.Dy()
	if totalPixels <= 0 {
		return metrics{}
	}

	var (
		sumSq      float64
		totalDiff  float64
		diffPixels int
		maxDiff    uint8
	)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rr, rg, rb, ra := rgbaAt(ref, x, y)
			ar, ag, ab, aa := rgbaAt(act, x, y)

			dr := absDiff8(rr, ar)
			dg := absDiff8(rg, ag)
			db := absDiff8(rb, ab)
			da := absDiff8(ra, aa)

			pixelMax := max4(dr, dg, db, da)
			if pixelMax > maxDiff {
				maxDiff = pixelMax
			}
			if pixelMax != 0 {
				diffPixels++
			}

			totalDiff += float64(dr) + float64(dg) + float64(db) + float64(da)
			sumSq += sqDiff(rr, ar) + sqDiff(rg, ag) + sqDiff(rb, ab) + sqDiff(ra, aa)
		}
	}

	return metrics{
		RMSE:        math.Sqrt(sumSq / float64(totalPixels*4)),
		AvgDiff:     totalDiff / float64(totalPixels*4),
		MaxDiff:     maxDiff,
		DiffPixels:  diffPixels,
		TotalPixels: totalPixels,
		DiffRatio:   float64(diffPixels) / float64(totalPixels),
	}
}

func rawDiffImage(ref, act *image.RGBA) *image.RGBA {
	bounds := unionBounds(ref.Bounds(), act.Bounds())
	out := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			rr, rg, rb, ra := rgbaAt(ref, bounds.Min.X+x, bounds.Min.Y+y)
			ar, ag, ab, aa := rgbaAt(act, bounds.Min.X+x, bounds.Min.Y+y)
			dr := absDiff8(rr, ar)
			dg := absDiff8(rg, ag)
			db := absDiff8(rb, ab)
			da := absDiff8(ra, aa)
			if dr == 0 && dg == 0 && db == 0 && da == 0 {
				out.SetRGBA(x, y, color.RGBA{R: 0, G: 0xaa, B: 0, A: 255})
				continue
			}
			out.SetRGBA(x, y, color.RGBA{R: dr, G: dg, B: db, A: clampAlpha(da)})
		}
	}
	return out
}

func amplifiedDiffImage(ref, act *image.RGBA) *image.RGBA {
	bounds := unionBounds(ref.Bounds(), act.Bounds())
	out := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			rr, rg, rb, ra := rgbaAt(ref, bounds.Min.X+x, bounds.Min.Y+y)
			ar, ag, ab, aa := rgbaAt(act, bounds.Min.X+x, bounds.Min.Y+y)
			dr := absDiff8(rr, ar)
			dg := absDiff8(rg, ag)
			db := absDiff8(rb, ab)
			da := absDiff8(ra, aa)
			pixelMax := max4(dr, dg, db, da)
			if pixelMax == 0 {
				out.SetRGBA(x, y, color.RGBA{R: 0, G: 0xaa, B: 0, A: 255})
				continue
			}
			intensity := uint8(255)
			if pixelMax < 255 {
				intensity = uint8((float64(pixelMax) / 255.0) * 255.0)
			}
			out.SetRGBA(x, y, color.RGBA{R: intensity, G: 0, B: 0, A: 255})
		}
	}
	return out
}

func rgbaAt(img *image.RGBA, x, y int) (uint8, uint8, uint8, uint8) {
	if img == nil || !image.Pt(x, y).In(img.Bounds()) {
		return 0, 0, 0, 0
	}
	i := img.PixOffset(x, y)
	return img.Pix[i+0], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3]
}

func unionBounds(a, b image.Rectangle) image.Rectangle {
	minX := minInt(a.Min.X, b.Min.X)
	minY := minInt(a.Min.Y, b.Min.Y)
	maxX := maxInt(a.Max.X, b.Max.X)
	maxY := maxInt(a.Max.Y, b.Max.Y)
	if maxX < minX {
		maxX = minX
	}
	if maxY < minY {
		maxY = minY
	}
	return image.Rect(minX, minY, maxX, maxY)
}

func max4(a, b, c, d uint8) uint8 {
	if a < b {
		a = b
	}
	if a < c {
		a = c
	}
	if a < d {
		a = d
	}
	return a
}

func absDiff8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func sqDiff(a, b uint8) float64 {
	d := float64(int(a) - int(b))
	return d * d
}

func pngToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func readPNGAsRGBA(path string) (*image.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	rgba := image.NewRGBA(img.Bounds())
	for y := rgba.Bounds().Min.Y; y < rgba.Bounds().Max.Y; y++ {
		for x := rgba.Bounds().Min.X; x < rgba.Bounds().Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba, nil
}

func clampAlpha(a uint8) uint8 {
	if a == 0 {
		return 255
	}
	return a
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func compositeOverSolid(src *image.RGBA, bg color.RGBA) *image.RGBA {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := rgbaAt(src, x, y)
			out.SetRGBA(x-bounds.Min.X, y-bounds.Min.Y, compositePixel(r, g, b, a, bg))
		}
	}
	return out
}

func compositePixel(r, g, b, a uint8, bg color.RGBA) color.RGBA {
	srcA := int(a)
	invA := 255 - srcA
	return color.RGBA{
		R: uint8((int(r)*srcA + int(bg.R)*invA) / 255),
		G: uint8((int(g)*srcA + int(bg.G)*invA) / 255),
		B: uint8((int(b)*srcA + int(bg.B)*invA) / 255),
		A: 255,
	}
}
