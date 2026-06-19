package svg

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestSaveSVGPreservesClip(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	r.ClipRect(geom.Rect{
		Min: geom.Pt{X: 10, Y: 10},
		Max: geom.Pt{X: 50, Y: 50},
	})
	r.DrawText("clipped", geom.Pt{X: 20, Y: 20}, 12, render.Color{R: 1})
	r.End()

	tmp, err := os.CreateTemp("", "matplotlib-go-svg-clip-*.svg")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	t.Cleanup(func() { _ = os.Remove(tmpPath) })

	if err := r.SaveSVG(tmpPath); err != nil {
		t.Fatalf("SaveSVG failed: %v", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<clipPath") {
		t.Fatal("SVG output should contain clipPath definitions")
	}
	if !strings.Contains(content, "clip-path=\"url(#") {
		t.Fatal("SVG output should apply clip-path to content")
	}
}

func TestRenderSVGPreservesClipStackAcrossSaveRestore(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipRect(geom.Rect{
			Min: geom.Pt{X: 5, Y: 5},
			Max: geom.Pt{X: 50, Y: 60},
		})
		r.Save()
		r.ClipRect(geom.Rect{
			Min: geom.Pt{X: 10, Y: 20},
			Max: geom.Pt{X: 30, Y: 40},
		})
		r.DrawText("inner", geom.Pt{X: 15, Y: 25}, 12, render.Color{A: 1})
		r.Restore()
		r.DrawText("outer", geom.Pt{X: 15, Y: 25}, 12, render.Color{A: 1})
	})

	if strings.Count(content, "<clipPath") != 2 {
		t.Fatalf("expected two clip path definitions after nested clipping, got %q", content)
	}
	if !strings.Contains(content, `<rect x="5" y="60" width="45" height="55" />`) {
		t.Fatalf("missing outer clip rect in SVG defs: %q", content)
	}
	if !strings.Contains(content, `<rect x="10" y="80" width="20" height="20" />`) {
		t.Fatalf("missing intersected inner clip rect in SVG defs: %q", content)
	}

	re := regexp.MustCompile(`<g clip-path="url\(#(clip\d+)\)"><text[^>]*>inner</text></g>\s*<g clip-path="url\(#(clip\d+)\)"><text[^>]*>outer</text></g>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) != 3 {
		t.Fatalf("expected clipped inner and restored outer groups, got %q", content)
	}
	if matches[1] == matches[2] {
		t.Fatalf("expected restore to switch back to the outer clip, got %q", content)
	}
}

func TestClipPathNestsInsideActiveRectClip(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipRect(geom.Rect{
			Min: geom.Pt{X: 10, Y: 15},
			Max: geom.Pt{X: 70, Y: 75},
		})

		var clipPath geom.Path
		clipPath.MoveTo(geom.Pt{X: 0, Y: 0})
		clipPath.LineTo(geom.Pt{X: 30, Y: 0})
		clipPath.LineTo(geom.Pt{X: 30, Y: 30})
		clipPath.Close()
		r.ClipPath(clipPath)

		r.DrawText("clipped", geom.Pt{X: 20, Y: 30}, 12, render.Color{A: 1})
	})

	if got := strings.Count(content, "<clipPath"); got != 2 {
		t.Fatalf("expected two clip defs (rect + path), got %d in %q", got, content)
	}
	if !strings.Contains(content, `<rect x="10" y="45" width="60" height="60" />`) {
		t.Fatalf("expected rect clip def, got %q", content)
	}
	if !strings.Contains(content, `<path d="M 0 120 L 30 120 L 30 90 Z" />`) {
		t.Fatalf("expected path clip def, got %q", content)
	}

	// Rect clip wraps the outer <g>; path clip nests inside.
	re := regexp.MustCompile(`<g clip-path="url\(#(clip\d+)\)"><g clip-path="url\(#(clip\d+)\)"><text[^>]*>clipped</text></g></g>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) != 3 {
		t.Fatalf("expected rect-clip outer group wrapping path-clip inner group, got %q", content)
	}
	if matches[1] == matches[2] {
		t.Fatalf("nested groups should reference distinct clip IDs, got %q", content)
	}
}

func TestClipPathDedupesIdenticalPathsAcrossNodes(t *testing.T) {
	makePath := func() geom.Path {
		var p geom.Path
		p.MoveTo(geom.Pt{X: 0, Y: 0})
		p.LineTo(geom.Pt{X: 40, Y: 0})
		p.LineTo(geom.Pt{X: 40, Y: 40})
		p.Close()
		return p
	}

	content := renderSVGDocument(t, func(r *Renderer) {
		r.Save()
		r.ClipPath(makePath())
		r.DrawText("first", geom.Pt{X: 10, Y: 10}, 12, render.Color{A: 1})
		r.Restore()

		r.Save()
		r.ClipPath(makePath())
		r.DrawText("second", geom.Pt{X: 10, Y: 30}, 12, render.Color{A: 1})
		r.Restore()
	})

	if got := strings.Count(content, "<clipPath"); got != 1 {
		t.Fatalf("identical path clips should share one def, got %d in %q", got, content)
	}
	if got := strings.Count(content, `clip-path="url(#clip1)"`); got != 2 {
		t.Fatalf("both nodes should reference the shared clip def, got %d refs in %q", got, content)
	}
}

func TestClipPathStackUnwindsOnRestore(t *testing.T) {
	makePath := func(offset float64) geom.Path {
		var p geom.Path
		p.MoveTo(geom.Pt{X: offset, Y: offset})
		p.LineTo(geom.Pt{X: offset + 20, Y: offset})
		p.LineTo(geom.Pt{X: offset + 20, Y: offset + 20})
		p.Close()
		return p
	}

	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipPath(makePath(0))
		r.Save()
		r.ClipPath(makePath(50))
		r.DrawText("nested", geom.Pt{X: 5, Y: 5}, 12, render.Color{A: 1})
		r.Restore()
		r.DrawText("popped", geom.Pt{X: 5, Y: 30}, 12, render.Color{A: 1})
	})

	// "nested" sees both clips; "popped" sees only the outer clip after Restore.
	nestedRe := regexp.MustCompile(`<g clip-path="url\(#clip1\)"><g clip-path="url\(#clip2\)"><text[^>]*>nested</text></g></g>`)
	if !nestedRe.MatchString(content) {
		t.Fatalf("expected nested element wrapped in both clips, got %q", content)
	}
	poppedRe := regexp.MustCompile(`<g clip-path="url\(#clip1\)"><text[^>]*>popped</text></g>`)
	if !poppedRe.MatchString(content) {
		t.Fatalf("expected popped element wrapped in only the outer clip, got %q", content)
	}
	if strings.Contains(content, `<g clip-path="url(#clip2)"><text x="5" y="30"`) {
		t.Fatalf("inner clip leaked past Restore, got %q", content)
	}
}

func TestClipPathRejectsEmptyPath(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipPath(geom.Path{}) // empty path → no clip, no def
		r.DrawText("unclipped", geom.Pt{X: 5, Y: 5}, 12, render.Color{A: 1})
	})

	if strings.Contains(content, "<clipPath") {
		t.Fatalf("empty path should not register clip defs, got %q", content)
	}
	if strings.Contains(content, "clip-path=") {
		t.Fatalf("empty path should not wrap content in clip groups, got %q", content)
	}
}

func TestClipPathTransformedEmitsPathTransformAndDedupesByTransform(t *testing.T) {
	makePath := func() geom.Path {
		var p geom.Path
		p.MoveTo(geom.Pt{X: 0, Y: 0})
		p.LineTo(geom.Pt{X: 20, Y: 0})
		p.LineTo(geom.Pt{X: 20, Y: 20})
		p.Close()
		return p
	}
	shift := geom.Affine{A: 1, D: 1, E: 10, F: 15}
	scale := geom.Affine{A: 2, D: 2}

	content := renderSVGDocument(t, func(r *Renderer) {
		r.Save()
		r.ClipPathTransformed(makePath(), shift)
		r.DrawText("first", geom.Pt{X: 5, Y: 5}, 12, render.Color{A: 1})
		r.Restore()

		r.Save()
		r.ClipPathTransformed(makePath(), shift)
		r.DrawText("second", geom.Pt{X: 5, Y: 20}, 12, render.Color{A: 1})
		r.Restore()

		r.Save()
		r.ClipPathTransformed(makePath(), scale)
		r.DrawText("third", geom.Pt{X: 5, Y: 35}, 12, render.Color{A: 1})
		r.Restore()
	})

	if got := strings.Count(content, "<clipPath"); got != 2 {
		t.Fatalf("path clips should dedupe by path plus transform, got %d defs in %q", got, content)
	}
	for _, want := range []string{
		`<path d="M 0 0 L 20 0 L 20 20 Z" transform="matrix(1 0 0 -1 10 105)" />`,
		`<path d="M 0 0 L 20 0 L 20 20 Z" transform="matrix(2 0 0 -2 0 120)" />`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected transformed clip def %q in %q", want, content)
		}
	}
	if got := strings.Count(content, `clip-path="url(#clip1)"`); got != 2 {
		t.Fatalf("first transform should share clip1 across two nodes, got %d refs in %q", got, content)
	}
	if got := strings.Count(content, `clip-path="url(#clip2)"`); got != 1 {
		t.Fatalf("second transform should use clip2 once, got %d refs in %q", got, content)
	}
}
