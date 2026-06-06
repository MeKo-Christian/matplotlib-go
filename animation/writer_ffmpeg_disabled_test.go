//go:build !ffmpeg

package animation

import (
	"path/filepath"
	"testing"
)

func TestFFmpegWritersUnsupportedWithoutBuildTag(t *testing.T) {
	cnv := newRasterFakeCanvas(2, 2)
	anim := newGIFAnimation(t, cnv, 2)

	for _, ext := range []string{".mp4", ".webm"} {
		out := filepath.Join(t.TempDir(), "anim"+ext)
		if err := anim.Save(out); err != ErrWriterUnsupported {
			t.Fatalf("Save(%s) = %v, want ErrWriterUnsupported", ext, err)
		}
	}

	for _, name := range []string{"ffmpeg", "ffmpeg-webm"} {
		if _, err := WriterByName(name, 5); err != ErrWriterUnsupported {
			t.Fatalf("WriterByName(%q) = %v, want ErrWriterUnsupported", name, err)
		}
	}
}
