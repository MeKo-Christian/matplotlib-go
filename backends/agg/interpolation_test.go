package agg

import (
	"testing"

	agglib "github.com/cwbudde/agg_go"
)

func TestParseInterpolationName(t *testing.T) {
	cases := []struct {
		name        string
		want        agglib.ImageFilter
		ok          bool
		useAdaptive bool
	}{
		{"", agglib.NoFilter, true, false},
		{"none", agglib.NoFilter, true, false},
		{"nearest", agglib.NoFilter, true, false},
		{"auto", agglib.FilterHanning, true, true},
		{"antialiased", agglib.FilterHanning, true, true},
		{"bilinear", agglib.Bilinear, true, false},
		{"bicubic", agglib.Bicubic, true, false},
		{"hanning", agglib.Hanning, true, false},
		{"hamming", agglib.FilterHamming, true, false},
		{"hermite", agglib.Hermite, true, false},
		{"kaiser", agglib.Blackman, true, false},
		{"quadric", agglib.Quadric, true, false},
		{"catrom", agglib.Catrom, true, false},
		{"gaussian", agglib.FilterGaussian, true, false},
		{"bessel", agglib.FilterBessel, true, false},
		{"mitchell", agglib.FilterMitchell, true, false},
		{"sinc", agglib.FilterSinc, true, false},
		{"lanczos", agglib.FilterLanczos, true, false},
		{"spline16", agglib.Spline16, true, false},
		{"spline36", agglib.Spline36, true, false},
		{"blackman", agglib.Blackman, true, false},
		{"BILINEAR", agglib.Bilinear, true, false},     // case-insensitive
		{"  bilinear  ", agglib.Bilinear, true, false}, // trim whitespace
	}
	for _, c := range cases {
		got, adaptive, ok := parseInterpolationName(c.name)
		if ok != c.ok {
			t.Errorf("parseInterpolationName(%q) ok = %v, want %v", c.name, ok, c.ok)
		}
		if adaptive != c.useAdaptive {
			t.Errorf("parseInterpolationName(%q) adaptive = %v, want %v", c.name, adaptive, c.useAdaptive)
		}
		if got != c.want {
			t.Errorf("parseInterpolationName(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestParseInterpolationName_Unknown(t *testing.T) {
	got, adaptive, ok := parseInterpolationName("definitely-not-a-filter")
	if ok {
		t.Fatal("expected ok=false for unknown filter")
	}
	if adaptive {
		t.Fatalf("adaptive = %v, want false", adaptive)
	}
	if got != agglib.NoFilter {
		t.Fatalf("fallback = %v, want NoFilter", got)
	}
}

func TestResolveInterpolationName(t *testing.T) {
	cases := []struct {
		name         string
		srcW, srcH   float64
		dstW, dstH   float64
		wantFilter   agglib.ImageFilter
		wantResolved bool
	}{
		{"nearest", 2, 2, 4, 4, agglib.NoFilter, true},
		{"auto", 2, 2, 4, 4, agglib.NoFilter, true},
		{"auto", 2, 2, 3, 3, agglib.FilterHanning, true},
		{"auto", 8, 8, 4, 4, agglib.FilterHanning, true},
		{"unknown", 2, 2, 64, 64, agglib.NoFilter, false},
	}
	for _, c := range cases {
		got, ok := resolveInterpolationName(c.name, c.srcW, c.srcH, c.dstW, c.dstH)
		if ok != c.wantResolved {
			t.Errorf("resolveInterpolationName(%q, %v, %v, %v, %v) ok = %v, want %v",
				c.name, c.srcW, c.srcH, c.dstW, c.dstH, ok, c.wantResolved)
		}
		if got != c.wantFilter {
			t.Errorf("resolveInterpolationName(%q, %v, %v, %v, %v) = %v, want %v",
				c.name, c.srcW, c.srcH, c.dstW, c.dstH, got, c.wantFilter)
		}
	}
}

func TestShouldUseNearestForAutoResample(t *testing.T) {
	cases := []struct {
		srcW, srcH, dstW, dstH float64
		want                   bool
	}{
		{2, 2, 4, 4, true},
		{2, 2, 3, 3, false},
		{2, 2, 10, 4, true},
		{2, 3, 3, 2, false},
		{2, 2, 6.0, 2.0, false},
		{2, 2, 6.1, 6.1, true},
	}
	for _, c := range cases {
		if got := shouldUseNearestForAutoResample(c.srcW, c.srcH, c.dstW, c.dstH); got != c.want {
			t.Errorf("shouldUseNearestForAutoResample(%v, %v, %v, %v) = %v, want %v",
				c.srcW, c.srcH, c.dstW, c.dstH, got, c.want)
		}
	}
}

func TestResampleForFilter(t *testing.T) {
	cases := []struct {
		filter agglib.ImageFilter
		want   agglib.ImageResample
	}{
		{agglib.NoFilter, agglib.NoResample},
		{agglib.Bilinear, agglib.ResampleAlways},
		{agglib.Bicubic, agglib.ResampleAlways},
		{agglib.Catrom, agglib.ResampleAlways},
	}
	for _, c := range cases {
		if got := resampleForFilter(c.filter); got != c.want {
			t.Errorf("resampleForFilter(%v) = %v, want %v", c.filter, got, c.want)
		}
	}
}
