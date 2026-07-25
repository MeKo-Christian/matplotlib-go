// Package style defines rc-like defaults, named themes, palette presets,
// style-library discovery, runtime rcParam overrides, and a supported-subset
// Matplotlib style/rc loader.
//
// Runtime defaults are synchronized process-wide, but PushContext scopes share
// one global LIFO stack and must not overlap across independent goroutines
// without external serialization. See docs/concurrency.md.
package style
