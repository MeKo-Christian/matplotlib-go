package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestColorbarBoundariesAndExtensionsAreDocumented(t *testing.T) {
	sourceRequirements := map[string][]string{
		filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "colorbar.py"): {
			"extend : {'neither', 'both', 'min', 'max'}",
			"set for a given colormap using the colormap set_under and set_over methods",
			"extendfrac : {*None*, 'auto', length, lengths}",
			"extendrect : bool",
			"drawedges : bool",
			"boundaries, values : None or a sequence",
			"spacing : {'uniform', 'proportional'}",
			"self._process_values()",
			"self.vmin, self.vmax = self._boundaries[self._inside][[0, -1]]",
			"def _do_extends(self, ax=None):",
			"def _get_extension_lengths(self, frac, automin, automax, default=0.05):",
		},
		filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "tests", "test_colorbar.py"): {
			"cmap.set_under('darkred')",
			"cmap.set_over('crimson')",
			"spacings = ['uniform', 'proportional']",
			"extendfrac=0.5",
			"extendrect=True",
		},
		filepath.Join("..", "core", "colorbar.go"): {
			"Boundaries  []float64",
			"Values      []float64",
			"Spacing     string",
			"DrawEdges   bool",
			"ExtendRect  bool",
			"func colorbarOptionBoundaries(values, boundaries []float64) []float64",
			"func colorbarInteriorBoundaries(boundaries []float64, extend string) []float64",
			"func normalizeColorbarSpacing(spacing string) string",
			"func colorbarExtensionPaths(clip geom.Rect, extend, orientation string, extendRect bool)",
			"func (c *Colorbar) boundaryExtensionValue(mapping ScalarMapInfo, overRange bool) (float64, bool)",
			"func drawColorbarBoundaryDividers",
		},
		filepath.Join("..", "core", "colorbar_test.go"): {
			"TestFigureAddColorbarUsesBoundaryNormTicks",
			"TestFigureAddColorbarUsesExplicitBoundariesAsTicks",
			"TestFigureAddColorbarUsesInteriorBoundaryLimitsWithExtensions",
			"TestBoundaryColorbarDrawUsesUniformSpacingByDefault",
			"TestBoundaryColorbarDrawCanUseProportionalSpacing",
			"TestBoundaryColorbarDrawEdgesAddsInternalDividers",
			"TestBoundaryColorbarWithExtensionsDrawsOnlyInteriorCells",
			"TestColorbarExtendRectDrawsRectangularExtensions",
		},
		filepath.Join("..", "core", "norm.go"): {
			"type BoundaryNorm struct",
			"Clip       bool",
			"Extend     string",
			"clip=true is not compatible with extend",
			"boundary norm boundaries must be strictly increasing",
		},
		filepath.Join("..", "color", "colormap.go"): {
			"func (c Colormap) AtValue(t float64) render.Color",
			"values below zero use under",
			"if t > 1 {",
			"if c.hasOver {",
			"func (c Colormap) WithUnder",
			"func (c Colormap) WithOver",
		},
		filepath.Join("..", "test", "parity", "colorbar_boundary_values", "plot.go"): {
			"Extend:     \"both\"",
			"ExtendRect: true",
			"Boundaries: []float64{-0.5, -0.1, 0.4, 1.2}",
			"Values:     []float64{-0.35, 0.15, 0.8}",
			"Spacing:    \"uniform\"",
			"DrawEdges:  true",
		},
	}
	for path, phrases := range sourceRequirements {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := string(data)
		for _, phrase := range phrases {
			if !strings.Contains(src, phrase) {
				t.Fatalf("%s missing colorbar boundary/extension marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Colorbar Boundaries and Extensions",
		"Matplotlib boundary colorbars derive `_boundaries`, `_values`, and interior `vmin`/`vmax` through `_process_values`",
		"`spacing='uniform'` gives each discrete color equal space and `spacing='proportional'` sizes cells by data interval",
		"`extend` supports `neither`, `min`, `max`, and `both`",
		"`extendrect` switches extension patches from triangles to rectangles",
		"`extendfrac` supports default 5%, `auto`, scalar, and pair lengths upstream",
		"Go supports `ColorbarOptions.Boundaries`, `Values`, `Spacing`, `DrawEdges`, `Extend`, and `ExtendRect`",
		"Go trims colorbar scale limits to interior boundaries when extensions are active",
		"under/over color routing is covered through `Colormap.AtValue` and explicit boundary extension values",
		"custom `extendfrac` and `extendfrac='auto'` remain documented residuals",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("colorbar boundary/extension docs missing %q", phrase)
		}
	}
}

func TestColorbarTickAndLabelFormattingIsDocumented(t *testing.T) {
	sourceRequirements := map[string][]string{
		filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "colorbar.py"): {
			"ticks=None",
			"format=None",
			"label=''",
			"orientation=None, ticklocation='auto'",
			"def update_ticks(self):",
			"self.long_axis.set_major_locator(self._locator)",
			"self.long_axis.set_minor_locator(self._minorlocator)",
			"self.long_axis.set_major_formatter(self._formatter)",
			"def _get_ticker_locator_formatter(self):",
			"def set_ticks(self, ticks, *, labels=None, minor=False, **kwargs):",
			"def get_ticks(self, minor=False):",
			"def minorticks_on(self):",
			"def minorticks_off(self):",
			"def set_label(self, label, *, loc=None, **kwargs):",
		},
		filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "tests", "test_colorbar.py"): {
			"test_colorbar_minorticks_on_off",
			"test_colorbar_get_ticks",
			"test_colorbar_set_formatter_locator",
			"ticklocation=\"top\"",
			"StrMethodFormatter",
		},
		filepath.Join("..", "core", "colorbar.go"): {
			"Ticks       []float64",
			"Label       string",
			"Orientation string",
			"func configureColorbarScale",
			"target.Locator = FixedLocator{TicksList: cloneFloat64s(boundaries)}",
			"target.Formatter = ScalarFormatter{Prec: 6}",
			"target.Locator = LogLocator{Base: base}",
			"target.Formatter = LogFormatterMathText{Base: base, SciNotation: true}",
			"func applyExplicitColorbarTicks",
			"formatter := ScalarFormatter{Prec: 6}",
			"func configureColorbarAxes",
			"ax.XAxis.MinorLocator = nil",
			"top.MinorLocator = nil",
			"right.MinorLocator = nil",
			"_ = ax.SetYLabelPosition(\"right\")",
			"_ = ax.SetXLabelPosition(\"bottom\")",
		},
		filepath.Join("..", "core", "colorbar_test.go"): {
			"TestFigureAddColorbarUsesLogNormTicks",
			"TestFigureAddColorbarUsesAsinhNormScale",
			"TestFigureAddLeftColorbarUsesLeftBoundaryTicks",
			"TestFigureAddColorbarUsesExplicitTicks",
			"TestHorizontalColorbarUsesExplicitTicks",
			"TestFigureAddHorizontalColorbarConfiguresTopAxes",
		},
		filepath.Join("..", "core", "tick_formatters.go"): {
			"type FormatStrFormatter struct",
			"type StrMethodFormatter struct",
			"type LogFormatterMathText struct",
		},
		filepath.Join("..", "test", "parity", "colorbar_horizontal_ticks", "plot.go"): {
			"Location: \"bottom\"",
			"Label:    \"horizontal\"",
			"Ticks:    []float64{-1, 0, 1}",
		},
	}
	for path, phrases := range sourceRequirements {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := string(data)
		for _, phrase := range phrases {
			if !strings.Contains(src, phrase) {
				t.Fatalf("%s missing colorbar tick/label marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Colorbar Tick and Label Formatting",
		"Matplotlib colorbars expose `ticks`, `format`, `label`, `orientation`, and `ticklocation`",
		"`update_ticks` applies the long-axis major locator, minor locator, and major formatter",
		"`set_ticks` supports labels and `minor=True` upstream",
		"Go supports explicit major ticks through `ColorbarOptions.Ticks`",
		"Go routes boundary, explicit boundary, log, asinh, and nonlinear norm colorbars to focused locators and formatters",
		"Colorbar labels are placed on the active long axis for right, left, top, and bottom locations",
		"horizontal bottom colorbar tick fixtures are covered by `colorbar_horizontal_ticks`",
		"Go clears colorbar minor locators and does not expose colorbar minor tick controls",
		"custom colorbar formatter options, ticklocation independent from location, `set_ticks(labels=...)`, and minor formatter APIs remain documented residuals",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("colorbar tick/label docs missing %q", phrase)
		}
	}
}
