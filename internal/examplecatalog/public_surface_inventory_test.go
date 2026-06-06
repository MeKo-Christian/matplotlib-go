package examplecatalog

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type publicSurfaceArtifact struct {
	SchemaVersion int                `json:"schema_version"`
	SourceRoot    string             `json:"source_root"`
	Modules       []string           `json:"modules"`
	Rows          []PublicSurfaceRow `json:"rows"`
}

func TestPublicSurfaceInventoryContainsLandmarkUpstreamRows(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	if artifact.SchemaVersion != 1 {
		t.Fatalf("schema version = %d, want 1", artifact.SchemaVersion)
	}
	want := []string{
		"artist.py:class:Artist",
		"axes/_axes.py:class:Axes",
		"axes/_axes.py:method:Axes.plot",
		"axes/_axes.py:method:Axes.violin",
		"axes/_base.py:method:_AxesBase.set_xlim",
		"axis.py:class:Axis",
		"ticker.py:class:EngFormatter",
		"scale.py:registry:scale:log",
		"transforms.py:class:Transform",
		"lines.py:class:Line2D",
		"markers.py:registry:marker:*",
		"markers.py:registry:fillstyle:left",
		"collections.py:class:PathCollection",
		"patches.py:registry:arrowstyle:->",
		"hatch.py:function:get_path",
		"text.py:class:Text",
		"font_manager.py:class:FontProperties",
		"textpath.py:class:TextPath",
		"legend.py:class:Legend",
		"legend_handler.py:class:HandlerBase",
		"offsetbox.py:class:OffsetBox",
		"image.py:registry:interpolation:lanczos",
		"colorbar.py:class:Colorbar",
		"cm.py:class:ColormapRegistry",
		"colors.py:class:Normalize",
		"pyplot.py:function:plot",
		"_pylab_helpers.py:class:Gcf",
		"backend_bases.py:class:FigureCanvasBase",
		"backend_tools.py:registry:tool:home",
		"widgets.py:class:Button",
		"animation.py:class:FuncAnimation",
		"mpl_toolkits/mplot3d/axes3d.py:class:Axes3D",
		"mpl_toolkits/mplot3d/axes3d.py:method:Axes3D.plot_surface",
	}
	for _, id := range want {
		if !publicSurfaceArtifactHasRow(artifact, id) {
			t.Fatalf("public surface inventory missing landmark row %q", id)
		}
	}
}

func TestPublicSurfaceInventoryMatchesExtractor(t *testing.T) {
	root := repoRoot(t)
	cmd := exec.Command("python3", filepath.Join(root, "internal", "examplecatalog", "extract_public_surface.py"))
	cmd.Dir = root
	got, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("run public surface extractor: %v\n%s", err, exitErr.Stderr)
		}
		t.Fatalf("run public surface extractor: %v", err)
	}
	want, err := os.ReadFile(filepath.Join(root, "test", "testdata", "parity_surface", "upstream_public_surface.json"))
	if err != nil {
		t.Fatalf("read committed public surface artifact: %v", err)
	}
	if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(want)) {
		t.Fatal("committed public surface artifact differs from extractor output; rerun internal/examplecatalog/extract_public_surface.py")
	}
}

func TestPublicSurfaceParityRowsClassifyLandmarkRows(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	for _, row := range PublicSurfaceParityRows() {
		validatePublicSurfaceParityRow(t, artifact, row)
	}
	want := []string{
		"artist.py:class:Artist",
		"lines.py:class:Line2D",
		"markers.py:registry:marker:*",
		"image.py:registry:interpolation:lanczos",
		"pyplot.py:function:plot",
		"widgets.py:class:Button",
		"animation.py:class:FuncAnimation",
	}
	for _, upstreamID := range want {
		if _, ok := LookupPublicSurfaceParityByUpstreamID(upstreamID); !ok {
			t.Fatalf("missing public surface parity classification for %q", upstreamID)
		}
	}
}

func TestImageClassAndIOOmissionsHaveExplicitRows(t *testing.T) {
	want := []string{
		"image.py:class:BboxImage",
		"image.py:class:FigureImage",
		"image.py:class:NonUniformImage",
		"image.py:class:PcolorImage",
		"image.py:function:imread",
		"image.py:function:imsave",
		"image.py:function:thumbnail",
	}
	for _, upstreamID := range want {
		row, ok := LookupPublicSurfaceParityByUpstreamID(upstreamID)
		if !ok {
			t.Fatalf("missing explicit Phase 12.5 image public-surface classification for %q", upstreamID)
		}
		if row.Status == PublicSurfaceNotStarted {
			t.Fatalf("%s status = %s, want an implemented, partial, idiomatic, or intentional-omission decision", upstreamID, row.Status)
		}
	}
}

func TestPyplotImageHelpersHaveExplicitRows(t *testing.T) {
	want := []string{
		"pyplot.py:function:figimage",
		"pyplot.py:function:imshow",
		"pyplot.py:function:matshow",
		"pyplot.py:function:pcolor",
		"pyplot.py:function:pcolormesh",
		"pyplot.py:function:spy",
	}
	for _, upstreamID := range want {
		row, ok := LookupPublicSurfaceParityByUpstreamID(upstreamID)
		if !ok {
			t.Fatalf("missing explicit Phase 12.5 pyplot image-helper classification for %q", upstreamID)
		}
		if row.Status == PublicSurfaceNotStarted {
			t.Fatalf("%s status = %s, want an implemented, partial, idiomatic, or intentional-omission decision", upstreamID, row.Status)
		}
	}
}

func TestPyplotDynamicShortcutsHaveExplicitRows(t *testing.T) {
	want := []string{
		"pyplot.py:function:clabel",
		"pyplot.py:function:clim",
		"pyplot.py:function:fignum_exists",
		"pyplot.py:function:findobj",
		"pyplot.py:function:get_figlabels",
		"pyplot.py:function:get_fignums",
		"pyplot.py:function:ginput",
		"pyplot.py:function:margins",
		"pyplot.py:function:new_figure_manager",
		"pyplot.py:function:polar",
		"pyplot.py:function:rgrids",
		"pyplot.py:function:sci",
		"pyplot.py:function:set_cmap",
		"pyplot.py:function:switch_backend",
		"pyplot.py:function:thetagrids",
		"pyplot.py:function:waitforbuttonpress",
	}
	for _, upstreamID := range want {
		row, ok := LookupPublicSurfaceParityByUpstreamID(upstreamID)
		if !ok {
			t.Fatalf("missing explicit Phase 12.5 pyplot dynamic/global shortcut classification for %q", upstreamID)
		}
		if row.Status == PublicSurfaceNotStarted {
			t.Fatalf("%s status = %s, want an implemented, partial, idiomatic, or intentional-omission decision", upstreamID, row.Status)
		}
	}
}

// TestPyplotStateSurfaceRowsAreExplicitlyDecided locks the Phase 17.6.7 closure
// of the stateful pyplot wrapper surface: every upstream pyplot.py and
// _pylab_helpers.py row must resolve to a parity decision that is not
// not-started. Implemented wrappers, documented partials, idiomatic equivalents,
// and intentional omissions are all acceptable; an unclassified or not-started
// row is not.
func TestPyplotStateSurfaceRowsAreExplicitlyDecided(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	checked := 0
	for _, surface := range artifact.Rows {
		if surface.Module != "pyplot.py" && surface.Module != "_pylab_helpers.py" {
			continue
		}
		checked++
		row, ok := PublicSurfaceParityForRow(surface)
		if !ok {
			t.Fatalf("missing Phase 17.6.7 pyplot state classification for %q", surface.ID)
		}
		if row.Status == PublicSurfaceNotStarted {
			t.Fatalf("%s status = %s, want an implemented, partial, idiomatic, or intentional-omission decision", surface.ID, row.Status)
		}
	}
	if checked == 0 {
		t.Fatal("no pyplot.py/_pylab_helpers.py surface rows found; artifact may be stale")
	}
}

func TestPatchStyleClosureRowsAreNotLeftPartial(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	for _, surface := range artifact.Rows {
		if surface.Module != "patches.py" {
			continue
		}
		if surface.Kind != "class" && !strings.HasPrefix(surface.Kind, "registry:") {
			continue
		}
		row, ok := PublicSurfaceParityForRow(surface)
		if !ok {
			t.Fatalf("missing Phase 17.6.6 patch-style classification for %q", surface.ID)
		}
		if row.Status == PublicSurfacePartial || row.Status == PublicSurfaceNotStarted {
			t.Fatalf("%s status = %s, want closed patch class/registry decision", surface.ID, row.Status)
		}
	}
}

func TestTextAnnotationOffsetboxRowsAreSplitByStaticAndGuiScope(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	rowFor := func(upstreamID string) PublicSurfaceParity {
		t.Helper()
		for _, surface := range artifact.Rows {
			if surface.ID != upstreamID {
				continue
			}
			row, ok := PublicSurfaceParityForRow(surface)
			if !ok {
				t.Fatalf("missing Phase 17.6.6 text/offsetbox classification for %q", upstreamID)
			}
			return row
		}
		t.Fatalf("missing upstream surface row %q", upstreamID)
		return PublicSurfaceParity{}
	}

	wantClosed := []string{
		"text.py:class:Annotation",
		"text.py:class:OffsetFrom",
		"text.py:class:Text",
		"offsetbox.py:class:AnchoredOffsetbox",
		"offsetbox.py:class:AnchoredText",
		"offsetbox.py:class:AnnotationBbox",
		"offsetbox.py:class:AuxTransformBox",
		"offsetbox.py:class:DrawingArea",
		"offsetbox.py:class:HPacker",
		"offsetbox.py:class:OffsetBox",
		"offsetbox.py:class:OffsetImage",
		"offsetbox.py:class:PackerBase",
		"offsetbox.py:class:PaddedBox",
		"offsetbox.py:class:TextArea",
		"offsetbox.py:class:VPacker",
	}
	for _, upstreamID := range wantClosed {
		row := rowFor(upstreamID)
		if row.Status == PublicSurfacePartial || row.Status == PublicSurfaceNotStarted {
			t.Fatalf("%s status = %s, want closed static text/offsetbox decision", upstreamID, row.Status)
		}
	}

	for _, upstreamID := range []string{
		"offsetbox.py:class:DraggableAnnotation",
		"offsetbox.py:class:DraggableBase",
		"offsetbox.py:class:DraggableOffsetBox",
		"offsetbox.py:constant:DEBUG",
	} {
		row := rowFor(upstreamID)
		if row.Status != PublicSurfaceIntentionalOmission {
			t.Fatalf("%s status = %s, want intentional omission", upstreamID, row.Status)
		}
	}
}

func TestLegendStaticRowsAreNotLeftPartial(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	rowFor := func(upstreamID string) PublicSurfaceParity {
		t.Helper()
		for _, surface := range artifact.Rows {
			if surface.ID != upstreamID {
				continue
			}
			row, ok := PublicSurfaceParityForRow(surface)
			if !ok {
				t.Fatalf("missing Phase 17.6.6 static legend classification for %q", upstreamID)
			}
			return row
		}
		t.Fatalf("missing upstream surface row %q", upstreamID)
		return PublicSurfaceParity{}
	}

	for _, upstreamID := range []string{
		"legend.py:class:Legend",
		"legend_handler.py:class:HandlerBase",
		"legend_handler.py:class:HandlerCircleCollection",
		"legend_handler.py:class:HandlerErrorbar",
		"legend_handler.py:class:HandlerLine2D",
		"legend_handler.py:class:HandlerLine2DCompound",
		"legend_handler.py:class:HandlerLineCollection",
		"legend_handler.py:class:HandlerNpoints",
		"legend_handler.py:class:HandlerNpointsYoffsets",
		"legend_handler.py:class:HandlerPatch",
		"legend_handler.py:class:HandlerPathCollection",
		"legend_handler.py:class:HandlerPolyCollection",
		"legend_handler.py:class:HandlerRegularPolyCollection",
		"legend_handler.py:class:HandlerStem",
		"legend_handler.py:class:HandlerStepPatch",
		"legend_handler.py:class:HandlerTuple",
		"legend_handler.py:function:update_from_first_child",
	} {
		row := rowFor(upstreamID)
		if row.Status == PublicSurfacePartial || row.Status == PublicSurfaceNotStarted {
			t.Fatalf("%s status = %s, want closed static legend decision", upstreamID, row.Status)
		}
	}
}

func TestDraggableLegendRowIsExplicitlyOmitted(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	for _, surface := range artifact.Rows {
		if surface.ID != "legend.py:class:DraggableLegend" {
			continue
		}
		row, ok := PublicSurfaceParityForRow(surface)
		if !ok {
			t.Fatal("missing Phase 17.6.6 DraggableLegend classification")
		}
		if row.Status != PublicSurfaceIntentionalOmission {
			t.Fatalf("DraggableLegend status = %s, want intentional omission", row.Status)
		}
		return
	}
	t.Fatal("missing upstream surface row legend.py:class:DraggableLegend")
}

func TestArtistDynamicRowsAreSplitFromStaticArtistSurface(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	rowFor := func(upstreamID string) PublicSurfaceParity {
		t.Helper()
		for _, surface := range artifact.Rows {
			if surface.ID != upstreamID {
				continue
			}
			row, ok := PublicSurfaceParityForRow(surface)
			if !ok {
				t.Fatalf("missing Phase 17.6.6 artist classification for %q", upstreamID)
			}
			return row
		}
		t.Fatalf("missing upstream surface row %q", upstreamID)
		return PublicSurfaceParity{}
	}

	for _, upstreamID := range []string{
		"artist.py:class:Artist",
		"artist.py:function:allow_rasterization",
	} {
		row := rowFor(upstreamID)
		if row.Status == PublicSurfacePartial || row.Status == PublicSurfaceNotStarted {
			t.Fatalf("%s status = %s, want closed static artist decision", upstreamID, row.Status)
		}
	}
	for _, upstreamID := range []string{
		"artist.py:class:ArtistInspector",
		"artist.py:function:getp",
		"artist.py:function:kwdoc",
		"artist.py:function:setp",
	} {
		row := rowFor(upstreamID)
		if row.Status != PublicSurfaceIntentionalOmission {
			t.Fatalf("%s status = %s, want intentional omission", upstreamID, row.Status)
		}
	}
}

func TestHatchImplementationRowsAreSplitFromRendererHatchSurface(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	rowFor := func(upstreamID string) PublicSurfaceParity {
		t.Helper()
		for _, surface := range artifact.Rows {
			if surface.ID != upstreamID {
				continue
			}
			row, ok := PublicSurfaceParityForRow(surface)
			if !ok {
				t.Fatalf("missing Phase 17.6.6 hatch classification for %q", upstreamID)
			}
			return row
		}
		t.Fatalf("missing upstream surface row %q", upstreamID)
		return PublicSurfaceParity{}
	}

	for _, upstreamID := range []string{
		"hatch.py:class:Circles",
		"hatch.py:class:HatchPatternBase",
		"hatch.py:class:HorizontalHatch",
		"hatch.py:class:LargeCircles",
		"hatch.py:class:NorthEastHatch",
		"hatch.py:class:Shapes",
		"hatch.py:class:SmallCircles",
		"hatch.py:class:SmallFilledCircles",
		"hatch.py:class:SouthEastHatch",
		"hatch.py:class:Stars",
		"hatch.py:class:VerticalHatch",
	} {
		row := rowFor(upstreamID)
		if row.Status != PublicSurfaceIntentionalOmission {
			t.Fatalf("%s status = %s, want intentional omission", upstreamID, row.Status)
		}
	}

	row := rowFor("hatch.py:function:get_path")
	if row.Status == PublicSurfacePartial || row.Status == PublicSurfaceNotStarted {
		t.Fatalf("hatch.py:function:get_path status = %s, want closed renderer hatch decision", row.Status)
	}
}

func TestSketchAndFigureImageRowsHaveExplicitDecisions(t *testing.T) {
	for _, upstreamID := range []string{
		"pyplot.py:function:xkcd",
		"pyplot.py:function:figimage",
	} {
		row, ok := LookupPublicSurfaceParityByUpstreamID(upstreamID)
		if !ok {
			t.Fatalf("missing explicit 9.7 decision for %s", upstreamID)
		}
		if row.Status != PublicSurfaceIntentionalOmission {
			t.Fatalf("%s status = %s, want intentional omission", upstreamID, row.Status)
		}
		if !partialNoteDocumentsRemainingScope(row.Note) {
			t.Fatalf("%s note should document omission rationale: %q", upstreamID, row.Note)
		}
	}
}

func TestPatchDebugHelperRowsAreExplicitlyOmitted(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	for _, upstreamID := range []string{
		"patches.py:function:bbox_artist",
		"patches.py:function:draw_bbox",
	} {
		found := false
		for _, surface := range artifact.Rows {
			if surface.ID != upstreamID {
				continue
			}
			found = true
			row, ok := PublicSurfaceParityForRow(surface)
			if !ok {
				t.Fatalf("missing Phase 17.6.6 patch helper classification for %q", upstreamID)
			}
			if row.Status != PublicSurfaceIntentionalOmission {
				t.Fatalf("%s status = %s, want intentional omission", upstreamID, row.Status)
			}
		}
		if !found {
			t.Fatalf("missing upstream surface row %q", upstreamID)
		}
	}
}

func TestFontManagerRowsAreSplitByTypedFontSurfaceAndRuntimeHelpers(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	rowFor := func(upstreamID string) PublicSurfaceParity {
		t.Helper()
		for _, surface := range artifact.Rows {
			if surface.ID != upstreamID {
				continue
			}
			row, ok := PublicSurfaceParityForRow(surface)
			if !ok {
				t.Fatalf("missing Phase 17.6.6 font-manager classification for %q", upstreamID)
			}
			return row
		}
		t.Fatalf("missing upstream surface row %q", upstreamID)
		return PublicSurfaceParity{}
	}

	for _, upstreamID := range []string{
		"font_manager.py:class:FontEntry",
		"font_manager.py:class:FontManager",
		"font_manager.py:class:FontProperties",
		"font_manager.py:function:get_font",
	} {
		row := rowFor(upstreamID)
		if row.Status == PublicSurfacePartial || row.Status == PublicSurfaceNotStarted {
			t.Fatalf("%s status = %s, want closed typed font-manager decision", upstreamID, row.Status)
		}
	}
	for _, upstreamID := range []string{
		"font_manager.py:function:afmFontProperty",
		"font_manager.py:function:findSystemFonts",
		"font_manager.py:function:get_fontext_synonyms",
		"font_manager.py:function:is_opentype_cff_font",
		"font_manager.py:function:json_dump",
		"font_manager.py:function:json_load",
		"font_manager.py:function:list_fonts",
		"font_manager.py:function:ttfFontProperty",
		"font_manager.py:function:win32FontDirectory",
	} {
		row := rowFor(upstreamID)
		if row.Status != PublicSurfaceIntentionalOmission {
			t.Fatalf("%s status = %s, want intentional omission", upstreamID, row.Status)
		}
	}
}

func TestTextPathRowsUseRendererLevelEquivalent(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	for _, upstreamID := range []string{
		"textpath.py:class:TextPath",
		"textpath.py:class:TextToPath",
	} {
		found := false
		for _, surface := range artifact.Rows {
			if surface.ID != upstreamID {
				continue
			}
			found = true
			row, ok := PublicSurfaceParityForRow(surface)
			if !ok {
				t.Fatalf("missing Phase 17.6.6 textpath classification for %q", upstreamID)
			}
			if row.Status == PublicSurfacePartial || row.Status == PublicSurfaceNotStarted {
				t.Fatalf("%s status = %s, want renderer-level textpath equivalent", upstreamID, row.Status)
			}
		}
		if !found {
			t.Fatalf("missing upstream surface row %q", upstreamID)
		}
	}
}

// assertModuleRowsClosed verifies every upstream row in the given modules
// resolves to a parity decision that is neither not-started nor partial. It is
// the Phase 17.6.8 closure guard for the backend/widget/animation tail: those
// rows used to be allowed to stay partial, but 17.6.8 resolved each into a
// direct/idiomatic equivalent or an explicit intentional omission.
func assertModuleRowsClosed(t *testing.T, modules ...string) {
	t.Helper()
	artifact := loadPublicSurfaceArtifact(t)
	wanted := map[string]bool{}
	for _, m := range modules {
		wanted[m] = true
	}
	checked := 0
	for _, surface := range artifact.Rows {
		if !wanted[surface.Module] {
			continue
		}
		checked++
		row, ok := PublicSurfaceParityForRow(surface)
		if !ok {
			t.Fatalf("missing Phase 17.6.8 classification for %q", surface.ID)
		}
		if row.Status == PublicSurfacePartial || row.Status == PublicSurfaceNotStarted {
			t.Fatalf("%s status = %s, want a closed backend/widget/animation decision (idiomatic, direct, or intentional-omission)", surface.ID, row.Status)
		}
	}
	if checked == 0 {
		t.Fatalf("no surface rows found for %v; artifact may be stale", modules)
	}
}

func TestWidgetClassesHaveExplicitRows(t *testing.T) {
	assertModuleRowsClosed(t, "widgets.py")
}

func TestAnimationSurfaceHasExplicitRows(t *testing.T) {
	assertModuleRowsClosed(t, "animation.py")
}

func TestBackendLifecycleAndToolRowsAreExplicit(t *testing.T) {
	// Landmark rows that must exist with a concrete classification, plus a full
	// sweep that rejects any leftover partial/not-started backend or tool row.
	want := []string{
		"backend_bases.py:class:Event",
		"backend_bases.py:class:DrawEvent",
		"backend_bases.py:class:FigureCanvasBase",
		"backend_bases.py:class:FigureManagerBase",
		"backend_bases.py:class:NavigationToolbar2",
		"backend_bases.py:class:TimerBase",
		"backend_tools.py:class:ToolBase",
		"backend_tools.py:registry:tool:home",
		"backend_tools.py:registry:tool:save",
	}
	for _, upstreamID := range want {
		row, ok := LookupPublicSurfaceParityByUpstreamID(upstreamID)
		if !ok {
			t.Fatalf("missing explicit backend/tool classification for %q", upstreamID)
		}
		if row.Status == PublicSurfaceNotStarted || row.Status == PublicSurfacePartial {
			t.Fatalf("%s status = %s, want a closed idiomatic, direct, or intentional-omission decision", upstreamID, row.Status)
		}
	}
	assertModuleRowsClosed(t, "backend_bases.py", "backend_tools.py")
}

// TestBackendWidgetAnimationTailDecisions locks the specific Phase 17.6.8
// closure decisions for the most load-bearing rows, including the ported GIF
// writer stack and the static-vs-GUI tool split.
func TestBackendWidgetAnimationTailDecisions(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	rowFor := func(upstreamID string) (PublicSurfaceParity, bool) {
		for _, surface := range artifact.Rows {
			if surface.ID == upstreamID {
				return PublicSurfaceParityForRow(surface)
			}
		}
		return PublicSurfaceParity{}, false
	}

	cases := []struct {
		upstreamID   string
		wantStatus   PublicSurfaceParityStatus
		noteContains string
	}{
		{"animation.py:class:PillowWriter", PublicSurfaceDirectEquivalent, "image/gif"},
		{"animation.py:class:AbstractMovieWriter", PublicSurfaceIdiomaticEquivalent, "MovieWriter interface"},
		{"animation.py:class:MovieWriterRegistry", PublicSurfaceIdiomaticEquivalent, "registry"},
		{"animation.py:class:FuncAnimation", PublicSurfaceIdiomaticEquivalent, "17.6.8"},
		{"animation.py:class:FFMpegWriter", PublicSurfaceIntentionalOmission, "external"},
		{"animation.py:class:HTMLWriter", PublicSurfaceIntentionalOmission, "HTML"},
		{"widgets.py:class:Button", PublicSurfaceIdiomaticEquivalent, "17.6.8"},
		{"backend_bases.py:class:FigureCanvasBase", PublicSurfaceIdiomaticEquivalent, "17.6.8"},
		{"backend_tools.py:registry:tool:fullscreen", PublicSurfaceIntentionalOmission, "intentionally omitted"},
	}
	for _, tc := range cases {
		row, ok := rowFor(tc.upstreamID)
		if !ok {
			t.Fatalf("missing classification for %q", tc.upstreamID)
		}
		if row.Status != tc.wantStatus {
			t.Fatalf("%s status = %s, want %s", tc.upstreamID, row.Status, tc.wantStatus)
		}
		if !strings.Contains(row.Note, tc.noteContains) {
			t.Fatalf("%s note %q does not contain %q", tc.upstreamID, row.Note, tc.noteContains)
		}
	}
}

func TestColorsNormSurfaceRowsHaveExplicitDecisions(t *testing.T) {
	want := []string{
		"colors.py:class:Normalize",
		"colors.py:class:SymLogNorm",
		"colors.py:class:PowerNorm",
		"colors.py:class:TwoSlopeNorm",
		"colors.py:class:CenteredNorm",
		"colors.py:class:BoundaryNorm",
		"colors.py:class:NoNorm",
		"colors.py:class:AsinhNorm",
		"colors.py:class:FuncNorm",
		"colors.py:function:make_norm_from_scale",
	}
	for _, upstreamID := range want {
		row, ok := LookupPublicSurfaceParityByUpstreamID(upstreamID)
		if !ok {
			t.Fatalf("missing explicit Phase 17.6.5 norm classification for %q", upstreamID)
		}
		if row.Status == PublicSurfaceNotStarted {
			t.Fatalf("%s status = %s, want an implemented, partial, idiomatic, or intentional-omission decision", upstreamID, row.Status)
		}
		if !strings.Contains(row.Note, "17.6.5") && !strings.Contains(row.Note, "ScalarNormalizer") {
			t.Fatalf("%s note does not reference the Phase 17.6.5 norm decision or ScalarNormalizer contract: %q", upstreamID, row.Note)
		}
	}
}

func TestColorsLightSourceSurfaceRowIsIntentionalOmission(t *testing.T) {
	row, ok := LookupPublicSurfaceParityByUpstreamID("colors.py:class:LightSource")
	if !ok {
		t.Fatal("missing explicit Phase 17.6.5 LightSource classification")
	}
	if row.Status != PublicSurfaceIntentionalOmission {
		t.Fatalf("LightSource status = %s, want %s", row.Status, PublicSurfaceIntentionalOmission)
	}
	required := []string{
		"Phase 17.6.5",
		"intentional omission",
		"LightSource.hillshade",
		"shade_rgb",
		"mplot3d face shading",
	}
	for _, phrase := range required {
		if !strings.Contains(row.Note, phrase) {
			t.Fatalf("LightSource row note missing %q: %q", phrase, row.Note)
		}
	}
	if !stringSliceContains(row.CatalogIDs, "mplot3d_terrain") {
		t.Fatalf("LightSource row catalog IDs = %v, want mplot3d_terrain audit anchor", row.CatalogIDs)
	}
}

func TestBivarMultivarColormapRowsAreIntentionalOmissions(t *testing.T) {
	tests := []struct {
		upstreamID string
		required   []string
	}{
		{
			upstreamID: "colors.py:class:BivarColormap",
			required:   []string{"Phase 17.6.5", "intentional omission", "two-component", "2D colorbar"},
		},
		{
			upstreamID: "colors.py:class:BivarColormapFromImage",
			required:   []string{"Phase 17.6.5", "intentional omission", "image-backed bivariate", "two-dimensional LUT"},
		},
		{
			upstreamID: "colors.py:class:MultivarColormap",
			required:   []string{"Phase 17.6.5", "intentional omission", "tuple-valued component arrays", "multi-component colorbar"},
		},
		{
			upstreamID: "colors.py:class:SegmentedBivarColormap",
			required:   []string{"Phase 17.6.5", "intentional omission", "segmented bivariate", "visual fixture"},
		},
	}
	for _, tc := range tests {
		row, ok := LookupPublicSurfaceParityByUpstreamID(tc.upstreamID)
		if !ok {
			t.Fatalf("missing explicit Phase 17.6.5 classification for %s", tc.upstreamID)
		}
		if row.Status != PublicSurfaceIntentionalOmission {
			t.Fatalf("%s status = %s, want %s", tc.upstreamID, row.Status, PublicSurfaceIntentionalOmission)
		}
		for _, phrase := range tc.required {
			if !strings.Contains(row.Note, phrase) {
				t.Fatalf("%s note missing %q: %q", tc.upstreamID, phrase, row.Note)
			}
		}
	}
}

func TestPublicSurfaceParityRowsCoverCommittedInventory(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	rows := PublicSurfaceParityRowsForSurface(artifact.Rows)
	if len(rows) != len(artifact.Rows) {
		classified := map[string]bool{}
		for _, row := range rows {
			classified[row.UpstreamID] = true
		}
		var missing []string
		for _, surface := range artifact.Rows {
			if !classified[surface.ID] {
				missing = append(missing, surface.ID)
			}
		}
		t.Fatalf("classified %d public-surface rows, want %d; missing: %v", len(rows), len(artifact.Rows), missing)
	}

	seenIDs := map[string]bool{}
	seenUpstream := map[string]bool{}
	for _, row := range rows {
		validatePublicSurfaceParityRow(t, artifact, row)
		if seenIDs[row.ID] {
			t.Fatalf("duplicate public surface parity row ID %q", row.ID)
		}
		seenIDs[row.ID] = true
		if seenUpstream[row.UpstreamID] {
			t.Fatalf("duplicate public surface upstream classification for %q", row.UpstreamID)
		}
		seenUpstream[row.UpstreamID] = true
	}
}

func TestPublicSurfaceParityRowsReferenceExistingLocalArtifacts(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	root := repoRoot(t)
	for _, row := range PublicSurfaceParityRowsForSurface(artifact.Rows) {
		for _, path := range row.GoFiles {
			requireFile(t, filepath.Join(root, path))
		}
		for _, id := range row.CatalogIDs {
			if _, ok := Lookup(id); !ok {
				t.Fatalf("%s references missing catalog case %q", row.ID, id)
			}
		}
		for _, id := range row.ExampleIDs {
			c, ok := Lookup(id)
			if !ok {
				t.Fatalf("%s references missing example case %q", row.ID, id)
			}
			if !c.Showcase {
				t.Fatalf("%s references non-showcase example %q", row.ID, id)
			}
		}
	}
}

func TestOpenPublicSurfaceParityRowsHaveClosureOwners(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	for _, row := range PublicSurfaceParityRowsForSurface(artifact.Rows) {
		if row.Status != PublicSurfacePartial && row.Status != PublicSurfaceNotStarted {
			continue
		}
		if row.ClosurePhase == "" && row.ClosureRationale == "" {
			t.Fatalf("%s (%s) is %s without a closure phase or omission rationale", row.ID, row.UpstreamID, row.Status)
		}
		if row.ClosurePhase != "" && !validPublicSurfaceClosurePhase(row.ClosurePhase) {
			t.Fatalf("%s has invalid closure phase %q", row.ID, row.ClosurePhase)
		}
	}
}

func TestPartialPublicSurfaceRowsHaveEvidenceOfRemainingScope(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	for _, row := range PublicSurfaceParityRowsForSurface(artifact.Rows) {
		if row.Status != PublicSurfacePartial {
			continue
		}
		hasFixtureEvidence := len(row.CatalogIDs) > 0 || len(row.ExampleIDs) > 0
		hasDocumentationEvidence := partialNoteDocumentsRemainingScope(row.Note)
		if !hasFixtureEvidence && !hasDocumentationEvidence {
			t.Fatalf("%s (%s) is partial without catalog/showcase evidence or a documented remaining scope", row.ID, row.UpstreamID)
		}
	}
}

func TestMatplotlibParityStatusDocIsCurrent(t *testing.T) {
	root := repoRoot(t)
	artifact := loadPublicSurfaceArtifact(t)
	got, err := os.ReadFile(filepath.Join(root, "docs", "matplotlib-parity-status.md"))
	if err != nil {
		t.Fatalf("read matplotlib parity status doc: %v", err)
	}
	want := MatplotlibParityStatusMarkdown(artifact.Rows)
	if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace([]byte(want))) {
		t.Fatal("docs/matplotlib-parity-status.md is stale; regenerate it from MatplotlibParityStatusMarkdown")
	}
}

func loadPublicSurfaceArtifact(t *testing.T) publicSurfaceArtifact {
	t.Helper()
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "test", "testdata", "parity_surface", "upstream_public_surface.json"))
	if err != nil {
		t.Fatalf("read public surface artifact: %v", err)
	}
	var artifact publicSurfaceArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fatalf("parse public surface artifact: %v", err)
	}
	return artifact
}

func publicSurfaceArtifactHasRow(artifact publicSurfaceArtifact, id string) bool {
	for _, row := range artifact.Rows {
		if row.ID == id {
			return true
		}
	}
	return false
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func validatePublicSurfaceParityRow(t *testing.T, artifact publicSurfaceArtifact, row PublicSurfaceParity) {
	t.Helper()
	if row.ID == "" || row.UpstreamID == "" || row.FeatureCoverageID == "" || row.Status == "" || row.Note == "" {
		t.Fatalf("incomplete public surface parity row: %+v", row)
	}
	if !validPublicSurfaceParityStatus(row.Status) {
		t.Fatalf("%s has invalid status %q", row.ID, row.Status)
	}
	if !publicSurfaceArtifactHasRow(artifact, row.UpstreamID) {
		t.Fatalf("%s references missing upstream public surface row %q", row.ID, row.UpstreamID)
	}
	if _, ok := LookupFeatureCoverage(row.FeatureCoverageID); !ok {
		t.Fatalf("%s references missing feature coverage row %q", row.ID, row.FeatureCoverageID)
	}
	if len(row.GoFiles) == 0 {
		t.Fatalf("%s has no local Go file reference", row.ID)
	}
}

func validPublicSurfaceClosurePhase(phase string) bool {
	switch phase {
	case "12.2C",
		"12.2D/E",
		"12.2G",
		"12.4C",
		"12.5",
		"17.6.2",
		"17.6.3",
		"17.6.4",
		"17.6.5",
		"17.6.6",
		"17.6.7",
		"17.6.8",
		"17.6.9":
		return true
	default:
		return false
	}
}

func partialNoteDocumentsRemainingScope(note string) bool {
	note = strings.ToLower(note)
	for _, marker := range []string{
		"remaining",
		"remains",
		"remain ",
		"documented",
		"omission",
		"omitted",
		"outside",
		"unsupported",
	} {
		if strings.Contains(note, marker) {
			return true
		}
	}
	return false
}

func validPublicSurfaceParityStatus(status PublicSurfaceParityStatus) bool {
	switch status {
	case PublicSurfaceDirectEquivalent,
		PublicSurfaceIdiomaticEquivalent,
		PublicSurfacePartial,
		PublicSurfaceNotStarted,
		PublicSurfaceIntentionalOmission:
		return true
	default:
		return false
	}
}
