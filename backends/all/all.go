// Package all imports the built-in rendering backends for side effects.
package all

import (
	_ "github.com/cwbudde/matplotlib-go/backends/agg"
	_ "github.com/cwbudde/matplotlib-go/backends/gobasic"
	_ "github.com/cwbudde/matplotlib-go/backends/pdf"
	_ "github.com/cwbudde/matplotlib-go/backends/pgf"
	_ "github.com/cwbudde/matplotlib-go/backends/ps"
	_ "github.com/cwbudde/matplotlib-go/backends/skia"
	_ "github.com/cwbudde/matplotlib-go/backends/svg"
)
