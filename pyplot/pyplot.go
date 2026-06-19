package pyplot

const (
	// DefaultFigureWidth matches Matplotlib's default 6.4in figure width at
	// the repository's default 100 DPI.
	DefaultFigureWidth = 640
	// DefaultFigureHeight matches Matplotlib's default 4.8in figure height at
	// the repository's default 100 DPI.
	DefaultFigureHeight = 480
)

// TightLayout enables tight layout on the current figure.
func TightLayout() {
	GCF().TightLayout()
}
