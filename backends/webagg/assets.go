package webagg

import (
	"embed"
	"io/fs"
)

//go:embed static/index.html static/mpl.js static/mpl.css
var staticFS embed.FS

// bundledAssets returns the embedded static asset tree, rooted at the
// directory holding index.html so callers can ask for "index.html"
// directly rather than "static/index.html".
func bundledAssets() (fs.FS, error) {
	return fs.Sub(staticFS, "static")
}
