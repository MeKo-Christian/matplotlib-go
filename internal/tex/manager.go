package tex

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/render"
)

// ManagerConfig configures the external LaTeX and dvipng command pipeline.
type ManagerConfig struct {
	CacheDir      string
	LaTeXCommand  string
	DVIPNGCommand string
	TFMDirs       []string
}

// Manager converts TeX strings to cached tight PNG artifacts.
type Manager struct {
	mu            sync.Mutex
	cacheDir      string
	latexCommand  string
	dvipngCommand string
	tfmDirs       []string
}

// RenderResult is one cached TeX raster artifact and its renderer metrics.
type RenderResult struct {
	PNGPath string
	Image   *image.RGBA
	Metrics render.TextMetrics
}

// NewManager creates a TeX manager with Matplotlib-like command defaults.
func NewManager(cfg ManagerConfig) *Manager {
	if cfg.CacheDir == "" {
		cfg.CacheDir = DefaultCacheDir()
	}
	if cfg.LaTeXCommand == "" {
		cfg.LaTeXCommand = "latex"
	}
	if cfg.DVIPNGCommand == "" {
		cfg.DVIPNGCommand = "dvipng"
	}
	return &Manager{
		cacheDir:      cfg.CacheDir,
		latexCommand:  cfg.LaTeXCommand,
		dvipngCommand: cfg.DVIPNGCommand,
		tfmDirs:       append([]string(nil), cfg.TFMDirs...),
	}
}

// DefaultCacheDir returns the default on-disk cache directory for TeX artifacts.
func DefaultCacheDir() string {
	if dir, err := os.UserCacheDir(); err == nil && dir != "" {
		return filepath.Join(dir, "matplotlib-go", "tex.cache")
	}
	return filepath.Join(os.TempDir(), "matplotlib-go", "tex.cache")
}

// Render returns a cached PNG rendering of a TeX string.
func (m *Manager) Render(text string, size float64, dpi uint, fontKey string) (RenderResult, error) {
	if m == nil {
		return RenderResult{}, errors.New("tex: manager is unavailable")
	}
	if strings.TrimSpace(text) == "" || size <= 0 {
		return RenderResult{}, nil
	}
	if dpi == 0 {
		dpi = 72
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	pngPath, err := m.makePNG(text, size, dpi, fontKey)
	if err != nil {
		return RenderResult{}, err
	}
	img, err := decodePNG(pngPath)
	if err != nil {
		return RenderResult{}, err
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	metrics := render.TextMetrics{
		W:       float64(w),
		H:       float64(h),
		Ascent:  float64(h),
		Descent: 0,
	}
	if dviMetrics, ok := m.cachedDVIMetrics(text, size, dpi, fontKey); ok {
		metrics = dviMetrics
	}
	return RenderResult{
		PNGPath: pngPath,
		Image:   img,
		Metrics: metrics,
	}, nil
}

func (m *Manager) makeDVI(text string, size float64, fontKey string) (string, error) {
	source := Source(text, size, fontKey)
	base, err := m.basePath(source, 0)
	if err != nil {
		return "", err
	}
	dviPath := base + ".dvi"
	if fileExists(dviPath) {
		return dviPath, nil
	}

	tmp, err := os.MkdirTemp(filepath.Dir(dviPath), "tex-build-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)

	texPath := filepath.Join(tmp, "file.tex")
	if err := os.WriteFile(texPath, []byte(source), 0o644); err != nil {
		return "", err
	}
	if err := runCommand([]string{m.latexCommand, "-interaction=nonstopmode", "-halt-on-error", "-no-shell-escape", "file.tex"}, text, tmp); err != nil {
		return "", err
	}
	tmpDVI := filepath.Join(tmp, "file.dvi")
	if !fileExists(tmpDVI) {
		return "", fmt.Errorf("tex: latex did not produce %s for TeX string %q", tmpDVI, text)
	}
	if err := os.Rename(tmpDVI, dviPath); err != nil {
		return "", err
	}
	_ = os.Rename(texPath, base+".tex")
	return dviPath, nil
}

func (m *Manager) makePNG(text string, size float64, dpi uint, fontKey string) (string, error) {
	source := Source(text, size, fontKey)
	base, err := m.basePath(source, dpi)
	if err != nil {
		return "", err
	}
	pngPath := base + ".png"
	if fileExists(pngPath) {
		return pngPath, nil
	}

	dviPath, err := m.makeDVI(text, size, fontKey)
	if err != nil {
		return "", err
	}

	tmp, err := os.MkdirTemp(filepath.Dir(pngPath), "tex-raster-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)

	cmd := []string{
		m.dvipngCommand,
		"-bg", "Transparent",
		"-D", strconv.FormatUint(uint64(dpi), 10),
		"-T", "tight",
		"-o", "file.png",
		dviPath,
	}
	if err := runCommand(cmd, text, tmp); err != nil {
		return "", err
	}
	tmpPNG := filepath.Join(tmp, "file.png")
	if !fileExists(tmpPNG) {
		return "", fmt.Errorf("tex: dvipng did not produce %s for TeX string %q", tmpPNG, text)
	}
	if err := os.Rename(tmpPNG, pngPath); err != nil {
		return "", err
	}
	return pngPath, nil
}

func (m *Manager) basePath(source string, dpi uint) (string, error) {
	key := source + "\ndpi=" + strconv.FormatUint(uint64(dpi), 10)
	sum := sha256.Sum256([]byte(key))
	hash := hex.EncodeToString(sum[:])
	dir := filepath.Join(m.cacheDir, hash[:2], hash[2:4])
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, hash), nil
}

func (m *Manager) cachedDVIMetrics(text string, size float64, dpi uint, fontKey string) (render.TextMetrics, bool) {
	source := Source(text, size, fontKey)
	base, err := m.basePath(source, 0)
	if err != nil {
		return render.TextMetrics{}, false
	}
	dviPath := base + ".dvi"
	if !fileExists(dviPath) {
		return render.TextMetrics{}, false
	}
	data, err := os.ReadFile(dviPath)
	if err != nil {
		return render.TextMetrics{}, false
	}
	pages, err := ParseDVI(data, float64(dpi), newTFMMetricProvider(m.tfmDirs))
	if err != nil || len(pages) != 1 {
		return render.TextMetrics{}, false
	}
	page := pages[0]
	if page.Width <= 0 && page.Height <= 0 && page.Descent <= 0 {
		return render.TextMetrics{}, false
	}
	return render.TextMetrics{
		W:       page.Width,
		H:       page.Height + page.Descent,
		Ascent:  page.Height,
		Descent: page.Descent,
	}, true
}

func runCommand(command []string, text, cwd string) error {
	if len(command) == 0 || command[0] == "" {
		return errors.New("tex: empty command")
	}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("tex: failed to process TeX string %q because %s could not be found", text, command[0])
	}
	return fmt.Errorf("tex: %s failed while processing TeX string %q\ncommand: %s\noutput:\n%s", filepath.Base(command[0]), text, strings.Join(command, " "), strings.TrimSpace(string(out)))
}

// Source returns the complete TeX source for processing one TeX string.
func Source(text string, size float64, fontKey string) string {
	baselineSkip := size * 1.25
	fontCommand := fontCommand(fontKey)
	return strings.Join([]string{
		`\RequirePackage{fix-cm}`,
		`\documentclass{article}`,
		`\newcommand{\mathdefault}[1]{#1}`,
		`\usepackage[utf8]{inputenc}`,
		`\DeclareUnicodeCharacter{2212}{\ensuremath{-}}`,
		`\usepackage[papersize=72in, margin=1in]{geometry}`,
		usePackageIfNotLoaded("underscore", "strings"),
		usePackageIfNotLoaded("textcomp", ""),
		`\pagestyle{empty}`,
		`\begin{document}`,
		fmt.Sprintf(`\fontsize{%s}{%s}%%`, formatFloat(size), formatFloat(baselineSkip)),
		`\ifdefined\psfrag\else\hbox{}\fi%`,
		fmt.Sprintf(`{%s %s}%%`, fontCommand, text),
		`\end{document}`,
		"",
	}, "\n")
}

func usePackageIfNotLoaded(pkg, option string) string {
	if option != "" {
		option = "[" + option + "]"
	}
	return fmt.Sprintf(`\makeatletter\@ifpackageloaded{%s}{}{\usepackage%s{%s}}\makeatother`, pkg, option, pkg)
}

func fontCommand(fontKey string) string {
	key := strings.ToLower(fontKey)
	switch {
	case strings.Contains(key, "mono"), strings.Contains(key, "courier"), strings.Contains(key, "typewriter"):
		return `\ttfamily`
	case strings.Contains(key, "sans"), strings.Contains(key, "helvetica"), strings.Contains(key, "arial"):
		return `\sffamily`
	default:
		return `\rmfamily`
	}
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func decodePNG(path string) (*image.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(dst, dst.Bounds(), src, bounds.Min, draw.Src)
	return dst, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
