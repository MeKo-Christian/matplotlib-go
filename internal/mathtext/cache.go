package mathtext

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Cache stores parsed MathText expressions and optionally final layouts. Layout
// entries are keyed by a caller-supplied MeasurementKey because final layout
// depends on renderer-specific text metrics and font resolution.
type Cache struct {
	mu          sync.RWMutex
	parsed      map[string]mathLayoutNode
	parsedOrder []string
	layouts     map[layoutCacheKey]MathTextLayout
	layoutOrder []layoutCacheKey
	maxParsed   int
	maxLayouts  int
}

// CacheConfig configures MathText cache bounds. Zero limits mean unbounded.
type CacheConfig struct {
	MaxParsed  int
	MaxLayouts int
}

type layoutCacheKey struct {
	kind           string
	text           string
	size           float64
	fontKey        string
	measurementKey string
}

var defaultCache = NewCache()

// NewCache creates an empty MathText cache.
func NewCache() *Cache {
	return NewCacheWithConfig(CacheConfig{})
}

// NewCacheWithConfig creates an empty MathText cache with optional entry
// bounds. Bounded caches evict the oldest stored entries first.
func NewCacheWithConfig(cfg CacheConfig) *Cache {
	return &Cache{
		parsed:     map[string]mathLayoutNode{},
		layouts:    map[layoutCacheKey]MathTextLayout{},
		maxParsed:  cfg.MaxParsed,
		maxLayouts: cfg.MaxLayouts,
	}
}

// DefaultCache returns the process-wide cache used for renderer-independent
// parsing. Callers that enable layout caching should normally own their own
// cache or provide a MeasurementKey that isolates renderer metric behavior.
func DefaultCache() *Cache {
	return defaultCache
}

// Clear removes all cached parsed expressions and layouts.
func (c *Cache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.parsed = map[string]mathLayoutNode{}
	c.parsedOrder = nil
	c.layouts = map[layoutCacheKey]MathTextLayout{}
	c.layoutOrder = nil
}

// Stats returns the current parsed-expression and layout entry counts.
func (c *Cache) Stats() (parsed, layouts int) {
	if c == nil {
		return 0, 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.parsed), len(c.layouts)
}

// SaveFile writes a deterministic JSON cache snapshot to path. The write is
// atomic within the destination directory so readers never observe a partial
// cache file.
func (c *Cache) SaveFile(path string) error {
	if c == nil {
		return fmt.Errorf("mathtext: nil cache")
	}
	if path == "" {
		return fmt.Errorf("mathtext: empty cache path")
	}

	snapshot := c.snapshot()
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("mathtext: marshal cache: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mathtext: create cache dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".mathtext-cache-*")
	if err != nil {
		return fmt.Errorf("mathtext: create temp cache: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("mathtext: write temp cache: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("mathtext: close temp cache: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("mathtext: replace cache: %w", err)
	}
	return nil
}

// LoadFile replaces the cache contents with a JSON snapshot produced by
// SaveFile. Loaded entries still honor this cache's configured entry bounds.
func (c *Cache) LoadFile(path string) error {
	if c == nil {
		return fmt.Errorf("mathtext: nil cache")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("mathtext: read cache: %w", err)
	}
	var snapshot cacheSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return fmt.Errorf("mathtext: parse cache: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.parsed = map[string]mathLayoutNode{}
	c.parsedOrder = nil
	c.layouts = map[layoutCacheKey]MathTextLayout{}
	c.layoutOrder = nil
	for _, expr := range snapshot.Parsed {
		if _, ok := c.parsed[expr]; ok {
			continue
		}
		c.parsed[expr] = parseMathLayoutNode(expr, nil)
		c.parsedOrder = append(c.parsedOrder, expr)
		c.evictParsedLocked()
	}
	for _, entry := range snapshot.Layouts {
		key := layoutCacheKey{
			kind:           entry.Kind,
			text:           entry.Text,
			size:           entry.Size,
			fontKey:        entry.FontKey,
			measurementKey: entry.MeasurementKey,
		}
		if _, ok := c.layouts[key]; !ok {
			c.layoutOrder = append(c.layoutOrder, key)
		}
		c.layouts[key] = cloneLayout(entry.Layout)
		c.evictLayoutsLocked()
	}
	return nil
}

func (c *Cache) parsedNode(expr string) (mathLayoutNode, bool) {
	if c == nil {
		return mathLayoutNode{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	node, ok := c.parsed[expr]
	return node, ok
}

func (c *Cache) snapshot() cacheSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	parsed := make([]string, 0, len(c.parsed))
	for expr := range c.parsed {
		parsed = append(parsed, expr)
	}
	sort.Strings(parsed)

	layouts := make([]cacheLayoutEntry, 0, len(c.layouts))
	for key, layout := range c.layouts {
		layouts = append(layouts, cacheLayoutEntry{
			Kind:           key.kind,
			Text:           key.text,
			Size:           key.size,
			FontKey:        key.fontKey,
			MeasurementKey: key.measurementKey,
			Layout:         cloneLayout(layout),
		})
	}
	sort.Slice(layouts, func(i, j int) bool {
		return layouts[i].less(layouts[j])
	})
	return cacheSnapshot{
		Version: 1,
		Parsed:  parsed,
		Layouts: layouts,
	}
}

type cacheSnapshot struct {
	Version int                `json:"version"`
	Parsed  []string           `json:"parsed,omitempty"`
	Layouts []cacheLayoutEntry `json:"layouts,omitempty"`
}

type cacheLayoutEntry struct {
	Kind           string         `json:"kind"`
	Text           string         `json:"text"`
	Size           float64        `json:"size"`
	FontKey        string         `json:"font_key,omitempty"`
	MeasurementKey string         `json:"measurement_key,omitempty"`
	Layout         MathTextLayout `json:"layout"`
}

func (e cacheLayoutEntry) less(other cacheLayoutEntry) bool {
	if e.Kind != other.Kind {
		return e.Kind < other.Kind
	}
	if e.Text != other.Text {
		return e.Text < other.Text
	}
	if e.Size != other.Size {
		return e.Size < other.Size
	}
	if e.FontKey != other.FontKey {
		return e.FontKey < other.FontKey
	}
	return e.MeasurementKey < other.MeasurementKey
}

func (c *Cache) storeParsedNode(expr string, node mathLayoutNode) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.parsed[expr]; !ok {
		c.parsedOrder = append(c.parsedOrder, expr)
	}
	c.parsed[expr] = node
	c.evictParsedLocked()
}

func (c *Cache) layout(key layoutCacheKey) (MathTextLayout, bool) {
	if c == nil {
		return MathTextLayout{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	layout, ok := c.layouts[key]
	if !ok {
		return MathTextLayout{}, false
	}
	return cloneLayout(layout), true
}

func (c *Cache) storeLayout(key layoutCacheKey, layout MathTextLayout) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.layouts[key]; !ok {
		c.layoutOrder = append(c.layoutOrder, key)
	}
	c.layouts[key] = cloneLayout(layout)
	c.evictLayoutsLocked()
}

func (c *Cache) evictParsedLocked() {
	if c.maxParsed <= 0 {
		return
	}
	for len(c.parsed) > c.maxParsed && len(c.parsedOrder) > 0 {
		oldest := c.parsedOrder[0]
		c.parsedOrder = c.parsedOrder[1:]
		delete(c.parsed, oldest)
	}
}

func (c *Cache) evictLayoutsLocked() {
	if c.maxLayouts <= 0 {
		return
	}
	for len(c.layouts) > c.maxLayouts && len(c.layoutOrder) > 0 {
		oldest := c.layoutOrder[0]
		c.layoutOrder = c.layoutOrder[1:]
		delete(c.layouts, oldest)
	}
}

func cloneLayout(layout MathTextLayout) MathTextLayout {
	layout.Runs = append([]MathTextLayoutRun(nil), layout.Runs...)
	layout.Rules = append([]MathTextLayoutRule(nil), layout.Rules...)
	return layout
}
