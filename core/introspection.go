package core

import (
	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/render"
)

// This file ports Matplotlib's artist introspection helpers — setp/getp and
// findobj — to idiomatic Go.
//
// setp/getp use a lightweight, reflection-free property model: an artist opts
// into string-keyed access by implementing PropertyBag (mirroring the existing
// ArtistLabeler/ArtistLabelSetter pattern in artist_label.go). The embedded
// ArtistRasterization provides a baseline set of common properties, so every
// artist that embeds it gets generic get/set for free. Concrete artists may
// implement PropertyBag to expose type-specific properties, delegating unknown
// keys back to the embedded base.
//
// findobj traverses the artist tree via the unexported artistChildrener
// interface (implemented by Axes, Figure, and any container artist) and filters
// with a predicate; FindobjType adds an isinstance-style typed convenience.

// PropertyBag is the optional interface an artist implements to participate in
// generic string-keyed property access through Getp/Setp. Property reports a
// named property's value; SetProperty assigns one and reports whether the key
// was recognized; PropertyNames lists the available keys.
type PropertyBag interface {
	Property(name string) (any, bool)
	SetProperty(name string, v any) bool
	PropertyNames() []string
}

// Getp returns the value of a named property on art, mirroring Matplotlib's
// getp(obj, property). The "label" property is routed through the shared
// ArtistLabeler interface; all other keys fall through to PropertyBag. The
// second result reports whether the property was found.
func Getp(art Artist, name string) (any, bool) {
	if art == nil {
		return nil, false
	}
	if name == "label" {
		if labeler, ok := art.(ArtistLabeler); ok {
			return labeler.ArtistLabel(), true
		}
	}
	if bag, ok := art.(PropertyBag); ok {
		return bag.Property(name)
	}
	return nil, false
}

// GetpAll returns every readable property on art keyed by name, mirroring
// Matplotlib's getp(obj) with no property argument. The "label" property is
// included when art is labelable.
func GetpAll(art Artist) map[string]any {
	if art == nil {
		return nil
	}
	out := make(map[string]any)
	if labeler, ok := art.(ArtistLabeler); ok {
		out["label"] = labeler.ArtistLabel()
	}
	if bag, ok := art.(PropertyBag); ok {
		for _, name := range bag.PropertyNames() {
			if v, ok := bag.Property(name); ok {
				out[name] = v
			}
		}
	}
	return out
}

// Setp assigns each named property on art, mirroring Matplotlib's
// setp(obj, **kwargs). The "label" property is routed through the shared
// ArtistLabelSetter interface; all other keys fall through to PropertyBag.
//
// Unrecognized keys (or values of the wrong type for a recognized key) emit a
// one-shot diagnostic via internal/diag rather than being silently dropped:
// Matplotlib raises AttributeError, and silently swallowing a typo'd property is
// a quiet-failure footgun for callers porting real scripts.
func Setp(art Artist, props map[string]any) {
	if art == nil {
		return
	}
	bag, hasBag := art.(PropertyBag)
	for name, v := range props {
		if name == "label" {
			if s, ok := v.(string); ok {
				SetArtistLabel(art, s)
				continue
			}
		}
		if hasBag && bag.SetProperty(name, v) {
			continue
		}
		diag.Warnf("setp: artist %T has no settable property %q (value ignored)", art, name)
	}
}

// Property implements PropertyBag for the common artist metadata stored on
// ArtistRasterization.
func (a *ArtistRasterization) Property(name string) (any, bool) {
	if a == nil {
		return nil, false
	}
	switch name {
	case "visible":
		return a.Visible(), true
	case "animated":
		return a.Animated(), true
	case "alpha":
		alpha, _ := a.ArtistAlpha()
		return alpha, true
	case "in_layout":
		return a.InLayout(), true
	case "clip_on":
		return a.ClipOn(), true
	case "rasterized":
		return a.Rasterization().Mode == render.RasterizeAlways, true
	case "url":
		return a.URL(), true
	case "gid":
		return a.GID(), true
	}
	return nil, false
}

// SetProperty implements PropertyBag for the common artist metadata stored on
// ArtistRasterization. It reports whether the key was recognized.
func (a *ArtistRasterization) SetProperty(name string, v any) bool {
	if a == nil {
		return false
	}
	switch name {
	case "visible":
		if b, ok := v.(bool); ok {
			a.SetVisible(b)
			return true
		}
	case "animated":
		if b, ok := v.(bool); ok {
			a.SetAnimated(b)
			return true
		}
	case "alpha":
		if f, ok := toFloat(v); ok {
			a.SetAlpha(f)
			return true
		}
	case "in_layout":
		if b, ok := v.(bool); ok {
			a.SetInLayout(b)
			return true
		}
	case "clip_on":
		if b, ok := v.(bool); ok {
			a.SetClipOn(b)
			return true
		}
	case "rasterized":
		if b, ok := v.(bool); ok {
			a.SetRasterized(b)
			return true
		}
	case "url":
		if s, ok := v.(string); ok {
			a.SetURL(s)
			return true
		}
	case "gid":
		if s, ok := v.(string); ok {
			a.SetGID(s)
			return true
		}
	}
	return false
}

// PropertyNames implements PropertyBag, listing the common artist metadata keys.
func (a *ArtistRasterization) PropertyNames() []string {
	return []string{"visible", "animated", "alpha", "in_layout", "clip_on", "rasterized", "url", "gid"}
}

// Property implements PropertyBag for Line2D, exposing the line-specific
// properties on top of the common metadata promoted from ArtistRasterization.
// This is the documented extension pattern for other concrete artists: handle
// the type's own keys, then delegate the rest to the embedded base.
func (l *Line2D) Property(name string) (any, bool) {
	if l == nil {
		return nil, false
	}
	switch name {
	case "linewidth":
		return l.W, true
	case "color":
		return l.Col, true
	}
	return l.ArtistRasterization.Property(name)
}

// SetProperty implements PropertyBag for Line2D.
func (l *Line2D) SetProperty(name string, v any) bool {
	if l == nil {
		return false
	}
	switch name {
	case "linewidth":
		if f, ok := toFloat(v); ok {
			l.W = f
			l.SetStale(true)
			return true
		}
	case "color":
		if c, ok := v.(render.Color); ok {
			l.Col = c
			l.SetStale(true)
			return true
		}
	default:
		return l.ArtistRasterization.SetProperty(name, v)
	}
	return false
}

// PropertyNames implements PropertyBag for Line2D.
func (l *Line2D) PropertyNames() []string {
	if l == nil {
		return nil
	}
	return append([]string{"linewidth", "color"}, l.ArtistRasterization.PropertyNames()...)
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	}
	return 0, false
}

// artistChildrener is implemented by container objects whose children Findobj
// should recurse into (Axes, Figure, and container artists).
type artistChildrener interface {
	artistChildren() []Artist
}

// artistChildren returns the artists owned by the axes — plot artists and
// widget-layer artists — plus the artists of any inset/child axes.
func (a *Axes) artistChildren() []Artist {
	if a == nil {
		return nil
	}
	children := make([]Artist, 0, len(a.Artists)+len(a.WidgetArtists))
	children = append(children, a.Artists...)
	children = append(children, a.WidgetArtists...)
	for _, child := range a.childAxes {
		children = append(children, child.artistChildren()...)
	}
	return children
}

// artistChildren returns every artist contained in the figure: the artists of
// each child axes plus the figure-level artists.
func (f *Figure) artistChildren() []Artist {
	if f == nil {
		return nil
	}
	var children []Artist
	for _, ax := range f.Children {
		children = append(children, ax.artistChildren()...)
	}
	children = append(children, f.Artists...)
	return children
}

// Findobj recursively collects artists reachable from root that satisfy match,
// mirroring Matplotlib's Artist.findobj. A nil match matches every artist. root
// may be an Artist or a container (*Axes, *Figure); a container root contributes
// its children only (it is not itself a drawable Artist here, a minor divergence
// from Matplotlib where Figure/Axes are Artists).
func Findobj(root any, match func(Artist) bool) []Artist {
	var out []Artist
	if art, ok := root.(Artist); ok {
		if match == nil || match(art) {
			out = append(out, art)
		}
	}
	if container, ok := root.(artistChildrener); ok {
		for _, child := range container.artistChildren() {
			out = append(out, Findobj(child, match)...)
		}
	}
	return out
}

// FindobjType collects every artist of concrete type T reachable from root,
// mirroring Matplotlib's findobj(match=SomeClass). root may be an Artist or a
// container (*Axes, *Figure).
func FindobjType[T any](root any) []T {
	var out []T
	for _, art := range Findobj(root, func(a Artist) bool {
		_, ok := a.(T)
		return ok
	}) {
		if typed, ok := art.(T); ok {
			out = append(out, typed)
		}
	}
	return out
}

// Findobj returns the figure's artists matching match (nil matches all).
func (f *Figure) Findobj(match func(Artist) bool) []Artist { return Findobj(f, match) }

// Findobj returns the axes' artists matching match (nil matches all).
func (a *Axes) Findobj(match func(Artist) bool) []Artist { return Findobj(a, match) }
