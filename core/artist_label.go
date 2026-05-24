package core

import "strings"

// ArtistLabeler exposes an artist label through a shared inspection path.
type ArtistLabeler interface {
	ArtistLabel() string
}

// ArtistLabelSetter updates an artist label through a shared metadata path.
type ArtistLabelSetter interface {
	SetArtistLabel(string)
}

// ArtistLabel returns an artist label when the artist exposes one.
func ArtistLabel(art Artist) string {
	if art == nil {
		return ""
	}
	provider, ok := art.(ArtistLabeler)
	if !ok {
		return ""
	}
	return provider.ArtistLabel()
}

// SetArtistLabel updates an artist label when the artist exposes one.
func SetArtistLabel(art Artist, label string) {
	if art == nil {
		return
	}
	setter, ok := art.(ArtistLabelSetter)
	if !ok {
		return
	}
	setter.SetArtistLabel(label)
}

func legendLabelVisible(label string) bool {
	return label != "" && !strings.HasPrefix(label, "_")
}

func (l *Line2D) ArtistLabel() string {
	if l == nil {
		return ""
	}
	return l.Label
}

func (l *Line2D) SetArtistLabel(label string) {
	if l == nil {
		return
	}
	l.Label = label
	l.SetStale(true)
}

func (s *Scatter2D) ArtistLabel() string {
	if s == nil {
		return ""
	}
	return s.Label
}

func (s *Scatter2D) SetArtistLabel(label string) {
	if s == nil {
		return
	}
	s.Label = label
	s.SetStale(true)
}

func (c *Collection) ArtistLabel() string {
	if c == nil {
		return ""
	}
	return c.Label
}

func (c *Collection) SetArtistLabel(label string) {
	if c == nil {
		return
	}
	c.Label = label
	c.SetStale(true)
}

func (p *Patch) ArtistLabel() string {
	if p == nil {
		return ""
	}
	return p.Label
}

func (p *Patch) SetArtistLabel(label string) {
	if p == nil {
		return
	}
	p.Label = label
	p.SetStale(true)
}

func (b *Bar2D) ArtistLabel() string {
	if b == nil {
		return ""
	}
	return b.Label
}

func (b *Bar2D) SetArtistLabel(label string) {
	if b == nil {
		return
	}
	b.Label = label
	b.SetStale(true)
}

func (f *Fill2D) ArtistLabel() string {
	if f == nil {
		return ""
	}
	return f.Label
}

func (f *Fill2D) SetArtistLabel(label string) {
	if f == nil {
		return
	}
	f.Label = label
	f.SetStale(true)
}

func (h *Hist2D) ArtistLabel() string {
	if h == nil {
		return ""
	}
	return h.Label
}

func (h *Hist2D) SetArtistLabel(label string) {
	if h == nil {
		return
	}
	h.Label = label
}

func (b *BoxPlot2D) ArtistLabel() string {
	if b == nil {
		return ""
	}
	return b.Label
}

func (b *BoxPlot2D) SetArtistLabel(label string) {
	if b == nil {
		return
	}
	b.Label = label
}

func (i *Image2D) ArtistLabel() string {
	if i == nil {
		return ""
	}
	return i.Label
}

func (i *Image2D) SetArtistLabel(label string) {
	if i == nil {
		return
	}
	i.Label = label
	i.SetStale(true)
}

func (e *ErrorBar) ArtistLabel() string {
	if e == nil {
		return ""
	}
	return e.Label
}

func (e *ErrorBar) SetArtistLabel(label string) {
	if e == nil {
		return
	}
	e.Label = label
}

func (s *Stairs2D) ArtistLabel() string {
	if s == nil {
		return ""
	}
	return s.Label
}

func (s *Stairs2D) SetArtistLabel(label string) {
	if s == nil {
		return
	}
	s.Label = label
}

func (q *Quiver) ArtistLabel() string {
	if q == nil {
		return ""
	}
	return q.Label
}

func (q *Quiver) SetArtistLabel(label string) {
	if q == nil {
		return
	}
	q.Label = label
}

func (b *Barbs) ArtistLabel() string {
	if b == nil {
		return ""
	}
	return b.Label
}

func (b *Barbs) SetArtistLabel(label string) {
	if b == nil {
		return
	}
	b.Label = label
}

func (s *StreamplotSet) ArtistLabel() string {
	if s == nil {
		return ""
	}
	return s.Label
}

func (s *StreamplotSet) SetArtistLabel(label string) {
	if s == nil {
		return
	}
	s.Label = label
}
