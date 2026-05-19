package tex

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
)

// DVIFont describes a font definition encountered in a DVI file.
type DVIFont struct {
	ID       uint32
	Name     string
	Checksum uint32
	Scale    uint32
	Design   uint32
}

// DVIGlyphMetrics are DVI-space glyph extents in TeX scaled points.
type DVIGlyphMetrics struct {
	Width  int32
	Height int32
	Depth  int32
}

// DVIMetricsProvider resolves glyph metrics for DVI font/glyph pairs.
type DVIMetricsProvider interface {
	GlyphMetrics(font DVIFont, glyph uint32) (DVIGlyphMetrics, bool)
}

// DVIPage is one parsed DVI page with coordinates converted to dpi units.
type DVIPage struct {
	Text    []DVIText
	Boxes   []DVIBox
	Width   float64
	Height  float64
	Descent float64
}

// DVIText is one DVI glyph placement after page-origin normalization.
type DVIText struct {
	X, Y  float64
	Font  DVIFont
	Glyph uint32
	Width float64
}

// DVIBox is one DVI rule after page-origin normalization.
type DVIBox struct {
	X, Y          float64
	Height, Width float64
}

type dviParser struct {
	r       *bytes.Reader
	dpi     float64
	metrics DVIMetricsProvider

	state dviState
	fonts map[uint32]DVIFont
	font  uint32

	h, v, w, x, y, z int32
	stack            []dviCursor
	text             []dviTextRaw
	boxes            []dviBoxRaw
	pages            []DVIPage
}

type dviState uint8

const (
	dviPre dviState = iota
	dviOuter
	dviInPage
	dviDone
)

type dviCursor struct {
	h, v, w, x, y, z int32
}

type dviTextRaw struct {
	x, y    int32
	font    DVIFont
	glyph   uint32
	metrics DVIGlyphMetrics
}

type dviBoxRaw struct {
	x, y          int32
	height, width int32
}

// ParseDVI parses the page geometry of a TeX DVI file. It follows the same
// extent model Matplotlib's dviread uses for text bounds: glyph heights extend
// above the baseline, glyph depths contribute descent, and rules have zero
// descent.
func ParseDVI(data []byte, dpi float64, metrics DVIMetricsProvider) ([]DVIPage, error) {
	if len(data) == 0 {
		return nil, errors.New("tex: empty dvi data")
	}
	if dpi <= 0 {
		dpi = 72
	}
	p := &dviParser{
		r:       bytes.NewReader(data),
		dpi:     dpi,
		metrics: metrics,
		state:   dviPre,
		fonts:   map[uint32]DVIFont{},
	}
	for p.state != dviDone {
		op, err := p.readByte()
		if err != nil {
			if errors.Is(err, io.EOF) && p.state == dviDone {
				break
			}
			return nil, err
		}
		if err := p.dispatch(op); err != nil {
			return nil, err
		}
	}
	return p.pages, nil
}

func (p *dviParser) dispatch(op byte) error {
	switch {
	case op <= 127:
		return p.setChar(uint32(op), true)
	case op >= 128 && op <= 131:
		glyph, err := p.readArg(int(op-128)+1, op == 131)
		if err != nil {
			return err
		}
		return p.setChar(uint32(glyph), true)
	case op == 132:
		a, err := p.readS4()
		if err != nil {
			return err
		}
		b, err := p.readS4()
		if err != nil {
			return err
		}
		p.putRule(a, b)
		p.h += b
	case op >= 133 && op <= 136:
		glyph, err := p.readArg(int(op-133)+1, op == 136)
		if err != nil {
			return err
		}
		return p.setChar(uint32(glyph), false)
	case op == 137:
		a, err := p.readS4()
		if err != nil {
			return err
		}
		b, err := p.readS4()
		if err != nil {
			return err
		}
		p.putRule(a, b)
	case op == 138:
		return nil
	case op == 139:
		return p.bop()
	case op == 140:
		return p.eop()
	case op == 141:
		p.stack = append(p.stack, dviCursor{h: p.h, v: p.v, w: p.w, x: p.x, y: p.y, z: p.z})
	case op == 142:
		if len(p.stack) == 0 {
			return errors.New("tex: dvi pop with empty stack")
		}
		top := p.stack[len(p.stack)-1]
		p.stack = p.stack[:len(p.stack)-1]
		p.h, p.v, p.w, p.x, p.y, p.z = top.h, top.v, top.w, top.x, top.y, top.z
	case op >= 143 && op <= 146:
		value, err := p.readArg(int(op-143)+1, true)
		if err != nil {
			return err
		}
		p.h += int32(value)
	case op >= 147 && op <= 151:
		return p.moveCached(op, 147, &p.w, &p.h)
	case op >= 152 && op <= 156:
		return p.moveCached(op, 152, &p.x, &p.h)
	case op >= 157 && op <= 160:
		value, err := p.readArg(int(op-157)+1, true)
		if err != nil {
			return err
		}
		p.v += int32(value)
	case op >= 161 && op <= 165:
		return p.moveCached(op, 161, &p.y, &p.v)
	case op >= 166 && op <= 170:
		return p.moveCached(op, 166, &p.z, &p.v)
	case op >= 171 && op <= 234:
		p.font = uint32(op - 171)
	case op >= 235 && op <= 238:
		value, err := p.readArg(int(op-235)+1, op == 238)
		if err != nil {
			return err
		}
		p.font = uint32(value)
	case op >= 239 && op <= 242:
		n, err := p.readArg(int(op-239)+1, false)
		if err != nil {
			return err
		}
		return p.skip(int64(n))
	case op >= 243 && op <= 246:
		return p.fontDef(op)
	case op == 247:
		return p.pre()
	case op == 248:
		p.state = dviDone
	default:
		return fmt.Errorf("tex: unsupported dvi opcode %d", op)
	}
	return nil
}

func (p *dviParser) pre() error {
	if p.state != dviPre {
		return errors.New("tex: dvi preamble outside pre state")
	}
	id, err := p.readByte()
	if err != nil {
		return err
	}
	num, err := p.readU4()
	if err != nil {
		return err
	}
	den, err := p.readU4()
	if err != nil {
		return err
	}
	mag, err := p.readU4()
	if err != nil {
		return err
	}
	commentLen, err := p.readByte()
	if err != nil {
		return err
	}
	if err := p.skip(int64(commentLen)); err != nil {
		return err
	}
	if id != 2 && id != 7 {
		return fmt.Errorf("tex: unsupported dvi format %d", id)
	}
	if num != 25400000 || den != 7227*65536 {
		return errors.New("tex: unsupported nonstandard dvi units")
	}
	if mag != 1000 {
		return errors.New("tex: unsupported dvi magnification")
	}
	p.state = dviOuter
	return nil
}

func (p *dviParser) bop() error {
	if p.state != dviOuter {
		return errors.New("tex: dvi bop outside page boundary")
	}
	for i := 0; i < 11; i++ {
		if _, err := p.readS4(); err != nil {
			return err
		}
	}
	p.state = dviInPage
	p.h, p.v, p.w, p.x, p.y, p.z = 0, 0, 0, 0, 0, 0
	p.stack = p.stack[:0]
	p.text = p.text[:0]
	p.boxes = p.boxes[:0]
	return nil
}

func (p *dviParser) eop() error {
	if p.state != dviInPage {
		return errors.New("tex: dvi eop outside page")
	}
	p.pages = append(p.pages, p.outputPage())
	p.state = dviOuter
	return nil
}

func (p *dviParser) fontDef(op byte) error {
	id, err := p.readArg(int(op-243)+1, op == 246)
	if err != nil {
		return err
	}
	checksum, err := p.readU4()
	if err != nil {
		return err
	}
	scale, err := p.readU4()
	if err != nil {
		return err
	}
	design, err := p.readU4()
	if err != nil {
		return err
	}
	areaLen, err := p.readByte()
	if err != nil {
		return err
	}
	nameLen, err := p.readByte()
	if err != nil {
		return err
	}
	nameBytes := make([]byte, int(areaLen)+int(nameLen))
	if _, err := io.ReadFull(p.r, nameBytes); err != nil {
		return err
	}
	p.fonts[uint32(id)] = DVIFont{
		ID:       uint32(id),
		Name:     string(nameBytes),
		Checksum: checksum,
		Scale:    scale,
		Design:   design,
	}
	return nil
}

func (p *dviParser) setChar(glyph uint32, advance bool) error {
	if p.state != dviInPage {
		return errors.New("tex: dvi character outside page")
	}
	font, ok := p.fonts[p.font]
	if !ok {
		return fmt.Errorf("tex: dvi font %d is not defined", p.font)
	}
	var gm DVIGlyphMetrics
	if p.metrics != nil {
		gm, ok = p.metrics.GlyphMetrics(font, glyph)
	}
	if ok {
		p.text = append(p.text, dviTextRaw{x: p.h, y: p.v, font: font, glyph: glyph, metrics: gm})
	}
	if advance {
		p.h += gm.Width
	}
	return nil
}

func (p *dviParser) putRule(height, width int32) {
	if p.state == dviInPage && height > 0 && width > 0 {
		p.boxes = append(p.boxes, dviBoxRaw{x: p.h, y: p.v, height: height, width: width})
	}
}

func (p *dviParser) moveCached(op, base byte, cache, position *int32) error {
	delta := int(op - base)
	if delta > 0 {
		value, err := p.readArg(delta, true)
		if err != nil {
			return err
		}
		*cache = int32(value)
	}
	*position += *cache
	return nil
}

func (p *dviParser) outputPage() DVIPage {
	if len(p.text) == 0 && len(p.boxes) == 0 {
		return DVIPage{}
	}
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	maxYPure := math.Inf(-1)

	for _, text := range p.text {
		x := float64(text.x)
		y := float64(text.y)
		width := float64(text.metrics.Width)
		height := float64(text.metrics.Height)
		depth := float64(text.metrics.Depth)
		minX = math.Min(minX, x)
		minY = math.Min(minY, y-height)
		maxX = math.Max(maxX, x+width)
		maxY = math.Max(maxY, y+depth)
		maxYPure = math.Max(maxYPure, y)
	}
	for _, box := range p.boxes {
		x := float64(box.x)
		y := float64(box.y)
		height := float64(box.height)
		width := float64(box.width)
		minX = math.Min(minX, x)
		minY = math.Min(minY, y-height)
		maxX = math.Max(maxX, x+width)
		maxY = math.Max(maxY, y)
		maxYPure = math.Max(maxYPure, y)
	}

	scale := p.dpi / (72.27 * 65536)
	descent := (maxY - maxYPure) * scale
	page := DVIPage{
		Text:    make([]DVIText, 0, len(p.text)),
		Boxes:   make([]DVIBox, 0, len(p.boxes)),
		Width:   (maxX - minX) * scale,
		Height:  (maxYPure - minY) * scale,
		Descent: descent,
	}
	for _, text := range p.text {
		page.Text = append(page.Text, DVIText{
			X:     (float64(text.x) - minX) * scale,
			Y:     (maxY-float64(text.y))*scale - descent,
			Font:  text.font,
			Glyph: text.glyph,
			Width: float64(text.metrics.Width) * scale,
		})
	}
	for _, box := range p.boxes {
		page.Boxes = append(page.Boxes, DVIBox{
			X:      (float64(box.x) - minX) * scale,
			Y:      (maxY-float64(box.y))*scale - descent,
			Height: float64(box.height) * scale,
			Width:  float64(box.width) * scale,
		})
	}
	return page
}

func (p *dviParser) readByte() (byte, error) {
	b, err := p.r.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("tex: read dvi opcode: %w", err)
	}
	return b, nil
}

func (p *dviParser) readU4() (uint32, error) {
	value, err := p.readArg(4, false)
	return uint32(value), err
}

func (p *dviParser) readS4() (int32, error) {
	value, err := p.readArg(4, true)
	return int32(value), err
}

func (p *dviParser) readArg(n int, signed bool) (int64, error) {
	if n <= 0 || n > 4 {
		return 0, fmt.Errorf("tex: invalid dvi argument width %d", n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(p.r, buf); err != nil {
		return 0, fmt.Errorf("tex: read dvi argument: %w", err)
	}
	value := int64(0)
	for _, b := range buf {
		value = value<<8 | int64(b)
	}
	if signed && buf[0]&0x80 != 0 {
		value -= 1 << (8 * n)
	}
	return value, nil
}

func (p *dviParser) skip(n int64) error {
	if n < 0 {
		return fmt.Errorf("tex: invalid dvi skip %d", n)
	}
	if _, err := p.r.Seek(n, io.SeekCurrent); err != nil {
		return fmt.Errorf("tex: skip dvi bytes: %w", err)
	}
	return nil
}
