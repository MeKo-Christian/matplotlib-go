package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestColorbarParentAndLayoutModesAreDocumented(t *testing.T) {
	sourceRequirements := map[string][]string{
		filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "figure.py"): {
			"def colorbar(\n            self, mappable, cax=None, ax=None, use_gridspec=True, **kwargs):",
			"ax : `~matplotlib.axes.Axes` or iterable or `numpy.ndarray` of Axes",
			"cax : `~matplotlib.axes.Axes`, optional",
			"cbar.make_axes_gridspec(ax, **kwargs)",
			"cbar.make_axes(ax, **kwargs)",
		},
		filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "colorbar.py"): {
			"def make_axes(parents, location=None, orientation=None, fraction=0.15,",
			"parents_bbox = mtransforms.Bbox.union",
			"for ax in parents:",
			"def make_axes_gridspec(parent, *, location=None, orientation=None,",
			"parent.set_subplotspec(ss_main)",
			"cax = fig.add_subplot(ss_cb, label=\"<colorbar>\")",
		},
		filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "tests", "test_colorbar.py"): {
			"fig.colorbar(im, ax=ax)",
			"fig.colorbar(im, ax=[_ax for _ax in ax])",
			"fig.colorbar(im, ax=(ax[0], ax[1]))",
			"cax = ax.inset_axes([1.02, 0.1, 0.03, 0.8])",
			"cb = fig.colorbar(pc, cax=cax)",
		},
		filepath.Join("..", "core", "colorbar.go"): {
			"func (f *Figure) AddColorbar(parent *Axes, mappable ScalarMappable",
			"f.AddAxes(rect)",
			"ax.colorbarParent = parent",
			"func colorbarUsesResolvedSlot(fig *Figure, parent *Axes) bool",
			"func colorbarBaseRect(parent *Axes) geom.Rect",
		},
		filepath.Join("..", "core", "layout_engine.go"): {
			"func syncColorbarAxesMeasured(fig *Figure, r render.Renderer, vp geom.Rect)",
			"if ax == nil || ax.colorbarParent == nil",
			"parent := ax.colorbarParent",
			"if parent.subplotSpec == nil",
			"colorbarUsesResolvedSlot(fig, parent)",
		},
		filepath.Join("..", "core", "gridspec.go"): {
			"func (f *Figure) AddSubplot",
			"func (f *Figure) AddSubplotSpec",
		},
		filepath.Join("..", "core", "axes_locator.go"): {
			"func (a *Axes) InsetAxes(bounds geom.Rect, opts ...InsetAxesOption) *Axes",
			"inset.SetAxesLocator(locator)",
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
				t.Fatalf("%s missing colorbar parent/layout marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Colorbar Parent and Layout Modes",
		"Matplotlib `Figure.colorbar` accepts `ax` as a single axes, iterable, tuple, dict-values view, or numpy array",
		"`make_axes` unions all parent axes positions and shrinks each parent",
		"`make_axes_gridspec` is single-parent and replaces the parent subplotspec",
		"`cax` may be a manually supplied axes, including an inset axes",
		"Go `Figure.AddColorbar` is intentionally single-parent through `parent *Axes`",
		"the Go colorbar axes is created internally with `Figure.AddAxes`",
		"subplot and constrained-layout colorbars track the parent through `syncColorbarAxesMeasured`",
		"Go has inset axes through `Axes.InsetAxes`, but no colorbar `cax` parameter",
		"multi-parent and inset-`cax` colorbar placement remain unsupported placement modes",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("colorbar parent/layout mode docs missing %q", phrase)
		}
	}
}

func TestColorbarSizeAndAnchorOptionsAreDocumented(t *testing.T) {
	sourceRequirements := map[string][]string{
		filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "colorbar.py"): {
			"fraction : float, default: 0.15",
			"shrink : float, default: 1.0",
			"aspect : float, default: 20",
			"pad : float, default: 0.05 if vertical, 0.15 if horizontal",
			"anchor : (float, float), optional",
			"panchor : (float, float), or *False*, optional",
			"\"left\":   {\"location\": \"left\", \"anchor\": (1.0, 0.5)",
			"\"right\":  {\"location\": \"right\", \"anchor\": (0.0, 0.5)",
			"\"top\":    {\"location\": \"top\", \"anchor\": (0.5, 0.0)",
			"\"bottom\": {\"location\": \"bottom\", \"anchor\": (0.5, 1.0)",
			"raise TypeError(\"location and orientation are mutually exclusive\")",
		},
		filepath.Join("..", "core", "colorbar.go"): {
			"Width       float64",
			"Padding     float64",
			"Aspect      float64",
			"Shrink      float64",
			"Anchor      *geom.Pt",
			"Location    string",
			"Orientation string",
			"defaultColorbarFraction          = 0.15",
			"defaultColorbarPadding           = 0.05",
			"defaultHorizontalColorbarPadding = 0.15",
			"defaultColorbarAspect            = 20.0",
			"func resolvedColorbarPadding(base geom.Rect, padding float64, location ...string) float64",
			"func resolvedColorbarThickness(fig *Figure, base geom.Rect, width, aspect float64, location string) float64",
			"func applyColorbarShrinkAnchor(rect geom.Rect, shrink float64, anchor *geom.Pt, location string) geom.Rect",
			"func normalizeColorbarLocation(location, orientation string) string",
		},
		filepath.Join("..", "core", "colorbar_test.go"): {
			"TestFigureAddHorizontalColorbarConfiguresBottomAxes",
			"TestFigureAddHorizontalColorbarConfiguresTopAxes",
			"TestFigureAddColorbarShrinkAnchorsVerticalLongAxis",
			"TestFigureAddColorbarShrinkAnchorsHorizontalLongAxis",
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
				t.Fatalf("%s missing colorbar size/anchor marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Colorbar Size and Anchor Options",
		"Matplotlib placement defaults are `fraction=0.15`, `shrink=1.0`, `aspect=20`, vertical `pad=0.05`, and horizontal `pad=0.15`",
		"Matplotlib `location` also chooses orientation and rejects incompatible `location`/`orientation` combinations",
		"Go records the default slot fraction with `defaultColorbarFraction=0.15`",
		"`defaultColorbarPadding=0.05`, `defaultHorizontalColorbarPadding=0.15`, and `defaultColorbarAspect=20`",
		"Go exposes explicit colorbar thickness as `ColorbarOptions.Width` rather than a Matplotlib `fraction` option",
		"`ColorbarOptions.Shrink` and `ColorbarOptions.Anchor` shrink only the long colorbar axis",
		"`ColorbarOptions.Padding`, `Aspect`, `Location`, and `Orientation` cover the supported placement defaults",
		"`panchor` is not supported because Go axes do not expose a parent-anchor colorbar option",
		"incompatible `Location` and `Orientation` values are normalized by location precedence rather than rejected",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("colorbar size/anchor docs missing %q", phrase)
		}
	}
}
