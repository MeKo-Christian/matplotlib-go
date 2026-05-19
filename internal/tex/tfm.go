package tex

import (
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type tfmMetricProvider struct {
	dirs  []string
	cache map[string]*tfmFile
}

type tfmFile struct {
	glyphs map[uint32]DVIGlyphMetrics
}

func newTFMMetricProvider(dirs []string) *tfmMetricProvider {
	return &tfmMetricProvider{
		dirs:  append([]string(nil), dirs...),
		cache: map[string]*tfmFile{},
	}
}

func (p *tfmMetricProvider) GlyphMetrics(font DVIFont, glyph uint32) (DVIGlyphMetrics, bool) {
	if p == nil {
		return DVIGlyphMetrics{}, false
	}
	tfm, ok := p.load(font.Name)
	if !ok {
		return DVIGlyphMetrics{}, false
	}
	metrics, ok := tfm.glyphs[glyph]
	if !ok {
		return DVIGlyphMetrics{}, false
	}
	return DVIGlyphMetrics{
		Width:  scaleTFMMetric(metrics.Width, font.Scale),
		Height: scaleTFMMetric(metrics.Height, font.Scale),
		Depth:  scaleTFMMetric(metrics.Depth, font.Scale),
	}, true
}

func (p *tfmMetricProvider) load(name string) (*tfmFile, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, false
	}
	if tfm, ok := p.cache[name]; ok {
		return tfm, tfm != nil
	}
	path, ok := p.find(name)
	if !ok {
		p.cache[name] = nil
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		p.cache[name] = nil
		return nil, false
	}
	tfm, err := parseTFM(data)
	if err != nil {
		p.cache[name] = nil
		return nil, false
	}
	p.cache[name] = tfm
	return tfm, true
}

func (p *tfmMetricProvider) find(name string) (string, bool) {
	names := []string{name}
	if filepath.Ext(name) == "" {
		names = append(names, name+".tfm")
	}
	for _, candidate := range names {
		if filepath.IsAbs(candidate) || strings.ContainsAny(candidate, `/\`) {
			if fileExists(candidate) {
				return candidate, true
			}
		}
		for _, dir := range p.dirs {
			path := filepath.Join(dir, candidate)
			if fileExists(path) {
				return path, true
			}
		}
	}
	if path, ok := kpsewhichTFM(name); ok {
		return path, true
	}
	return "", false
}

func kpsewhichTFM(name string) (string, bool) {
	if filepath.Ext(name) == "" {
		name += ".tfm"
	}
	if _, err := exec.LookPath("kpsewhich"); err != nil {
		return "", false
	}
	out, err := exec.Command("kpsewhich", name).Output()
	if err != nil {
		return "", false
	}
	path := strings.TrimSpace(string(out))
	if path == "" || !fileExists(path) {
		return "", false
	}
	return path, true
}

func parseTFM(data []byte) (*tfmFile, error) {
	if len(data) < 24 {
		return nil, errInvalidTFM
	}
	lf := int(readTFMHalf(data, 0))
	lh := int(readTFMHalf(data, 2))
	bc := int(readTFMHalf(data, 4))
	ec := int(readTFMHalf(data, 6))
	nw := int(readTFMHalf(data, 8))
	nh := int(readTFMHalf(data, 10))
	nd := int(readTFMHalf(data, 12))
	if lf <= 0 || len(data) < lf*4 || ec < bc || nw <= 0 || nh <= 0 || nd <= 0 {
		return nil, errInvalidTFM
	}
	charCount := ec - bc + 1
	charInfoWord := 6 + lh
	widthWord := charInfoWord + charCount
	heightWord := widthWord + nw
	depthWord := heightWord + nh
	if depthWord+nd > lf {
		return nil, errInvalidTFM
	}

	widths := readTFMIntTable(data, widthWord, nw)
	heights := readTFMIntTable(data, heightWord, nh)
	depths := readTFMIntTable(data, depthWord, nd)
	glyphs := map[uint32]DVIGlyphMetrics{}
	for i := 0; i < charCount; i++ {
		offset := (charInfoWord + i) * 4
		widthIndex := int(data[offset])
		heightIndex := int(data[offset+1] >> 4)
		depthIndex := int(data[offset+1] & 0x0f)
		if widthIndex >= len(widths) || heightIndex >= len(heights) || depthIndex >= len(depths) {
			continue
		}
		glyphs[uint32(bc+i)] = DVIGlyphMetrics{
			Width:  widths[widthIndex],
			Height: heights[heightIndex],
			Depth:  depths[depthIndex],
		}
	}
	return &tfmFile{glyphs: glyphs}, nil
}

var errInvalidTFM = os.ErrInvalid

func readTFMHalf(data []byte, offset int) uint16 {
	return binary.BigEndian.Uint16(data[offset : offset+2])
}

func readTFMIntTable(data []byte, word, count int) []int32 {
	out := make([]int32, count)
	for i := range out {
		out[i] = int32(binary.BigEndian.Uint32(data[(word+i)*4 : (word+i+1)*4]))
	}
	return out
}

func scaleTFMMetric(value int32, scale uint32) int32 {
	return int32((int64(value) * int64(scale)) >> 20)
}
