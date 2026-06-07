package animation

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAPNGEncodesAnimationControlChunks(t *testing.T) {
	cnv := newRasterFakeCanvas(4, 3)
	anim := newGIFAnimation(t, cnv, 3)

	out := filepath.Join(t.TempDir(), "anim.apng")
	if err := anim.Save(out); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read apng: %v", err)
	}
	chunks := parsePNGChunks(t, data)

	if got := chunkCount(chunks, "acTL"); got != 1 {
		t.Fatalf("acTL chunks = %d, want 1", got)
	}
	actl := firstChunk(t, chunks, "acTL").data
	if got := binary.BigEndian.Uint32(actl[0:4]); got != 3 {
		t.Fatalf("acTL frame count = %d, want 3", got)
	}
	if got := binary.BigEndian.Uint32(actl[4:8]); got != 0 {
		t.Fatalf("acTL num plays = %d, want 0", got)
	}

	fctl := chunksByType(chunks, "fcTL")
	if got := len(fctl); got != 3 {
		t.Fatalf("fcTL chunks = %d, want 3", got)
	}
	for i, ch := range fctl {
		if got := binary.BigEndian.Uint32(ch.data[4:8]); got != 4 {
			t.Fatalf("frame %d width = %d, want 4", i, got)
		}
		if got := binary.BigEndian.Uint32(ch.data[8:12]); got != 3 {
			t.Fatalf("frame %d height = %d, want 3", i, got)
		}
		if got := binary.BigEndian.Uint16(ch.data[20:22]); got != 1 {
			t.Fatalf("frame %d delay numerator = %d, want 1", i, got)
		}
		if got := binary.BigEndian.Uint16(ch.data[22:24]); got != 5 {
			t.Fatalf("frame %d delay denominator = %d, want 5", i, got)
		}
		if got := ch.data[24]; got != 0 {
			t.Fatalf("frame %d dispose op = %d, want 0", i, got)
		}
		if got := ch.data[25]; got != 0 {
			t.Fatalf("frame %d blend op = %d, want 0", i, got)
		}
	}
	if got := chunkCount(chunks, "fdAT"); got != 2 {
		t.Fatalf("fdAT chunks = %d, want 2", got)
	}

	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("decode first frame as png: %v", err)
	}
}

func TestSaveAPNGRespectsSaveCountAndFPS(t *testing.T) {
	cnv := newRasterFakeCanvas(2, 2)
	anim := newGIFAnimation(t, cnv, 10)

	out := filepath.Join(t.TempDir(), "anim.apng")
	if err := anim.Save(out, WithSaveCount(4), WithFPS(10)); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read apng: %v", err)
	}
	chunks := parsePNGChunks(t, data)
	actl := firstChunk(t, chunks, "acTL").data
	if got := binary.BigEndian.Uint32(actl[0:4]); got != 4 {
		t.Fatalf("acTL frame count = %d, want 4", got)
	}
	for i, ch := range chunksByType(chunks, "fcTL") {
		if got := binary.BigEndian.Uint16(ch.data[20:22]); got != 1 {
			t.Fatalf("frame %d delay numerator = %d, want 1", i, got)
		}
		if got := binary.BigEndian.Uint16(ch.data[22:24]); got != 10 {
			t.Fatalf("frame %d delay denominator = %d, want 10", i, got)
		}
	}
}

func TestWriterByNameAPNGRegistry(t *testing.T) {
	w, err := WriterByName("apng", 5)
	if err != nil {
		t.Fatalf("WriterByName(apng): %v", err)
	}
	if _, ok := w.(*APNGWriter); !ok {
		t.Fatalf("WriterByName(apng) = %T, want *APNGWriter", w)
	}
}

func TestAPNGWriterRequiresRasterCanvas(t *testing.T) {
	w := NewAPNGWriter(5)
	if err := w.Setup(plainCanvas{}, "x.apng", 0); err != ErrWriterUnsupported {
		t.Fatalf("Setup(non-raster) = %v, want ErrWriterUnsupported", err)
	}
}

func TestAPNGWriterFinishWithoutFrames(t *testing.T) {
	w := NewAPNGWriter(5)
	if err := w.Setup(newRasterFakeCanvas(2, 2), filepath.Join(t.TempDir(), "empty.apng"), 0); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if err := w.Finish(); err != ErrNoFramesGrabbed {
		t.Fatalf("Finish without frames = %v, want ErrNoFramesGrabbed", err)
	}
}

type pngChunk struct {
	typ  string
	data []byte
}

func parsePNGChunks(t *testing.T, data []byte) []pngChunk {
	t.Helper()
	if len(data) < 8 || !bytes.Equal(data[:8], []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
		t.Fatalf("missing PNG signature")
	}
	var chunks []pngChunk
	off := 8
	for off < len(data) {
		if off+8 > len(data) {
			t.Fatalf("truncated chunk header at offset %d", off)
		}
		n := int(binary.BigEndian.Uint32(data[off : off+4]))
		typ := string(data[off+4 : off+8])
		off += 8
		if off+n+4 > len(data) {
			t.Fatalf("truncated %s chunk at offset %d", typ, off)
		}
		chunks = append(chunks, pngChunk{typ: typ, data: data[off : off+n]})
		off += n + 4
		if typ == "IEND" {
			break
		}
	}
	return chunks
}

func chunkCount(chunks []pngChunk, typ string) int {
	count := 0
	for _, ch := range chunks {
		if ch.typ == typ {
			count++
		}
	}
	return count
}

func chunksByType(chunks []pngChunk, typ string) []pngChunk {
	var matches []pngChunk
	for _, ch := range chunks {
		if ch.typ == typ {
			matches = append(matches, ch)
		}
	}
	return matches
}

func firstChunk(t *testing.T, chunks []pngChunk, typ string) pngChunk {
	t.Helper()
	for _, ch := range chunks {
		if ch.typ == typ {
			return ch
		}
	}
	t.Fatalf("missing %s chunk", typ)
	return pngChunk{}
}
