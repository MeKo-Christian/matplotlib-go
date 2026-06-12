package text_layout_gallery

import (
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
)

func TestWrappedSampleUsesAutomaticWrap(t *testing.T) {
	fig := Plot()
	var found []*core.Text
	for _, artist := range fig.Children[0].Artists {
		text, ok := artist.(*core.Text)
		if ok && strings.HasPrefix(text.Content, "wrapped text uses") {
			found = append(found, text)
		}
	}
	if len(found) != 1 {
		t.Fatalf("found %d wrapped sample text artists, want 1", len(found))
	}
	if !found[0].Wrap || found[0].WrapWidth != 0 {
		t.Fatalf("wrapped sample Wrap=%v WrapWidth=%v, want automatic wrap", found[0].Wrap, found[0].WrapWidth)
	}
}
