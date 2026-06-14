package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	goldenUpdateBuildTag               = "freetype"
	goldenUpdateOptionalVisualTestsEnv = "RUN_OPTIONAL_VISUAL_TESTS"
	goldenUpdateFallbackGoCache        = "/tmp/mpl-parity-gocache"
	rerenderAllButtonPlaceholder       = "__RERENDER_ALL_BUTTON__"
	goldenUpdateRunPatternAll          = "^TestGolden$"
	goldenUpdateTimeout                = 5 * time.Minute
)

type caseEntry struct {
	Suite       string
	Baseline    string
	Name        string
	RMSE        float64
	AvgDiff     float64
	MaxDiff     uint8
	DiffPixels  int
	TotalPixels int
	DiffRatio   float64
	RefWidth    int
	RefHeight   int
	ActWidth    int
	ActHeight   int
	RefB64      string
	ActB64      string
	RawDiffB64  string
	AmpDiffB64  string
}

type loadResult struct {
	Cases         []caseEntry
	ComparedCount int
	SkippedCount  int
}

type viewOptions struct {
	CanRerender         bool
	RerenderDisabledMsg string
}

type metrics struct {
	RMSE        float64
	AvgDiff     float64
	MaxDiff     uint8
	DiffPixels  int
	TotalPixels int
	DiffRatio   float64
}

func main() {
	port := flag.String("port", envOr("PORT", "8090"), "Port to listen on")
	parityDir := flag.String("parity-dir", "", "Optional parity directory with suite/baseline-* / artifacts")
	baselineDir := flag.String("baseline-dir", filepath.Join("testdata", "matplotlib_ref"), "Baseline PNG directory (used when --parity-dir is not set)")
	artifactDir := flag.String("artifact-dir", filepath.Join("testdata", "golden"), "Artifact PNG directory (used when --parity-dir is not set)")
	includeWebdemo := flag.Bool("include-webdemo", false, "Also include web demo parity artifacts")
	webBaselineDir := flag.String("web-baseline-dir", filepath.Join("testdata", "_artifacts", "webdemo", "matplotlib"), "Web demo baseline PNG directory")
	webArtifactDir := flag.String("web-artifact-dir", filepath.Join("testdata", "_artifacts", "webdemo", "go"), "Web demo artifact PNG directory")
	nameFilter := flag.String("name-filter", "", "Optional case-name substring filter")
	namePrefix := flag.String("name-prefix", "", "Optional case-name prefix filter")
	printOnly := flag.Bool("print", false, "Print comparison rows and exit")
	flag.Parse()

	root, err := detectRepoRoot()
	if err != nil {
		log.Fatalf("detect repo root: %v", err)
	}

	baseDir := filepath.Clean(*baselineDir)
	if !filepath.IsAbs(baseDir) {
		baseDir = filepath.Join(root, baseDir)
	}
	artDir := filepath.Clean(*artifactDir)
	if !filepath.IsAbs(artDir) {
		artDir = filepath.Join(root, artDir)
	}
	webBaseDir := filepath.Clean(*webBaselineDir)
	if !filepath.IsAbs(webBaseDir) {
		webBaseDir = filepath.Join(root, webBaseDir)
	}
	webArtDir := filepath.Clean(*webArtifactDir)
	if !filepath.IsAbs(webArtDir) {
		webArtDir = filepath.Join(root, webArtDir)
	}
	parDir := filepath.Clean(*parityDir)
	if *parityDir != "" && !filepath.IsAbs(parDir) {
		parDir = filepath.Join(root, parDir)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		result, err := loadCasesUnified(*parityDir != "", *includeWebdemo, parDir, baseDir, artDir, webBaseDir, webArtDir, *nameFilter, *namePrefix)
		if err != nil {
			http.Error(w, fmt.Sprintf("load parity cases: %v", err), http.StatusInternalServerError)
			return
		}
		renderPage(w, result, buildViewOptionsWithExtra(*parityDir != "", *includeWebdemo, root, artDir))
	})
	http.HandleFunc("/rerender", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if r.FormValue("all") == "1" {
			if err := ensureRerenderSupportedWithExtra(*parityDir != "", *includeWebdemo, root, artDir); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := rerenderAllArtifacts(root); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Error(w, "missing name parameter", http.StatusBadRequest)
			return
		}
		if err := ensureRerenderSupportedWithExtra(*parityDir != "", *includeWebdemo, root, artDir); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := rerenderArtifact(root, name); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if *printOnly {
		result, err := loadCasesUnified(*parityDir != "", *includeWebdemo, parDir, baseDir, artDir, webBaseDir, webArtDir, *nameFilter, *namePrefix)
		if err != nil {
			log.Fatalf("load parity cases: %v", err)
		}
		printCases(os.Stdout, result)
		return
	}

	addr := ":" + *port
	log.Printf("Parity viewer running at http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func buildViewOptions(useParity bool, repoRoot, artifactDir string) viewOptions {
	return buildViewOptionsWithExtra(useParity, false, repoRoot, artifactDir)
}

func buildViewOptionsWithExtra(useParity, includeExtra bool, repoRoot, artifactDir string) viewOptions {
	if useParity {
		return viewOptions{
			CanRerender:         false,
			RerenderDisabledMsg: "Re-render is disabled in --parity-dir mode because the button only regenerates ./test goldens.",
		}
	}
	if includeExtra {
		return viewOptions{
			CanRerender:         false,
			RerenderDisabledMsg: "Re-render is disabled in unified mode because web demo artifacts need a different regeneration command.",
		}
	}
	defaultArtifactDir := filepath.Join(repoRoot, "testdata", "golden")
	if !samePath(defaultArtifactDir, artifactDir) {
		return viewOptions{
			CanRerender:         false,
			RerenderDisabledMsg: fmt.Sprintf("Re-render is disabled because the viewer is reading artifacts from %s, but rerender only updates %s.", artifactDir, defaultArtifactDir),
		}
	}
	return viewOptions{CanRerender: true}
}

func ensureRerenderSupported(useParity bool, repoRoot, artifactDir string) error {
	return ensureRerenderSupportedWithExtra(useParity, false, repoRoot, artifactDir)
}

func ensureRerenderSupportedWithExtra(useParity, includeExtra bool, repoRoot, artifactDir string) error {
	opts := buildViewOptionsWithExtra(useParity, includeExtra, repoRoot, artifactDir)
	if opts.CanRerender {
		return nil
	}
	return errors.New(opts.RerenderDisabledMsg)
}

func samePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

func rerenderArtifact(repoRoot, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("missing name")
	}
	return runGoParityRender(repoRoot, name)
}

func rerenderAllArtifacts(repoRoot string) error {
	return runGoGoldenUpdate(repoRoot, goldenUpdateRunPatternAll)
}

func runGoGoldenUpdate(repoRoot, runPattern string) error {
	ctx, cancel := context.WithTimeout(context.Background(), goldenUpdateTimeout)
	defer cancel()

	cmd := newGoldenUpdateCommandContext(ctx, repoRoot, runPattern)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(out.String())
		if msg == "" {
			msg = err.Error()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("go test timed out after %s for %q: %s", goldenUpdateTimeout, runPattern, msg)
		}
		return fmt.Errorf("go test failed: %w: %s", err, msg)
	}
	if runPattern != "" {
		outText := out.String()
		if strings.Contains(outText, "testing: warning: no tests to run") {
			return fmt.Errorf("no matching test for case %q", strings.TrimSuffix(strings.TrimPrefix(runPattern, "^"), "$"))
		}
	}
	return nil
}

func runGoParityRender(repoRoot, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), goldenUpdateTimeout)
	defer cancel()

	cmd := newParityRenderCommandContext(ctx, repoRoot, name)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(out.String())
		if msg == "" {
			msg = err.Error()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("go run timed out after %s for %q: %s", goldenUpdateTimeout, name, msg)
		}
		return fmt.Errorf("go run failed: %w: %s", err, msg)
	}
	return nil
}

func newGoldenUpdateCommand(repoRoot, runPattern string) *exec.Cmd {
	return newGoldenUpdateCommandContext(context.Background(), repoRoot, runPattern)
}

func newGoldenUpdateCommandContext(ctx context.Context, repoRoot, runPattern string) *exec.Cmd {
	args := []string{"test", "-tags", goldenUpdateBuildTag, "-count", "1", "-timeout", goldenUpdateTimeout.String()}
	if runPattern != "" {
		args = append(args, "-run", runPattern)
	}
	args = append(args, "./test", "-update-golden")

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = repoRoot

	cmd.Env = goCommandEnv()
	cmd.Env = setEnv(cmd.Env, goldenUpdateOptionalVisualTestsEnv, "true")

	return cmd
}

func newParityRenderCommand(repoRoot, name string) *exec.Cmd {
	return newParityRenderCommandContext(context.Background(), repoRoot, name)
}

func newParityRenderCommandContext(ctx context.Context, repoRoot, name string) *exec.Cmd {
	args := []string{
		"run",
		"-tags",
		goldenUpdateBuildTag,
		"./test/parity/cmd",
		"--id",
		name,
		"--output-dir",
		filepath.Join("testdata", "golden"),
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = repoRoot
	cmd.Env = goCommandEnv()

	return cmd
}

func goCommandEnv() []string {
	env := os.Environ()
	env = setEnv(env, "CGO_ENABLED", "1")
	if cacheDir, hasCache := os.LookupEnv("GOCACHE"); !hasCache || strings.TrimSpace(cacheDir) == "" {
		env = setEnv(env, "GOCACHE", goldenUpdateDefaultGoCache())
	}
	return env
}

func goldenUpdateDefaultGoCache() string {
	out, err := exec.Command("go", "env", "GOCACHE").Output()
	if err == nil {
		if cacheDir := strings.TrimSpace(string(out)); cacheDir != "" {
			return cacheDir
		}
	}
	return goldenUpdateFallbackGoCache
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	found := false
	for i, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			env[i] = prefix + value
			found = true
		}
	}
	if !found {
		env = append(env, prefix+value)
	}
	return env
}

func testNameFromCaseName(name string) string {
	return "^TestGolden/" + name + "$"
}

func loadCasesUnified(useParity, includeWebdemo bool, parityDir, baselineDir, artifactDir, webBaselineDir, webArtifactDir, nameFilter, namePrefix string) (loadResult, error) {
	if useParity {
		return loadCasesFromParityDir(parityDir, nameFilter, namePrefix)
	}
	if !includeWebdemo {
		return loadCasesFromDirectories(baselineDir, artifactDir, nameFilter, namePrefix)
	}
	return loadCasesFromDirectorySources([]directorySource{
		{
			Suite:       "plots",
			Baseline:    filepath.Base(baselineDir),
			BaselineDir: baselineDir,
			ArtifactDir: artifactDir,
		},
		{
			Suite:       "webdemo",
			Baseline:    "matplotlib",
			BaselineDir: webBaselineDir,
			ArtifactDir: webArtifactDir,
		},
	}, nameFilter, namePrefix)
}

func printCases(w io.Writer, result loadResult) {
	fmt.Fprintf(w, "found=%d skipped=%d\n", result.ComparedCount, result.SkippedCount)
	for i := range result.Cases {
		c := &result.Cases[i]
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\trmse=%.4f\tavg=%.4f\tmax=%d\tdiff_pixels=%d\tdiff_ratio=%.6f\tsize=%dx%d->%dx%d\n",
			c.Suite,
			c.Baseline,
			c.Name,
			c.RMSE,
			c.AvgDiff,
			c.MaxDiff,
			c.DiffPixels,
			c.DiffRatio,
			c.RefWidth,
			c.RefHeight,
			c.ActWidth,
			c.ActHeight,
		)
	}
}

func detectRepoRoot() (string, error) {
	wd, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}

	for range 12 {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		next := filepath.Dir(wd)
		if next == wd {
			break
		}
		wd = next
	}
	return "", errors.New("go.mod not found from cwd")
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func loadCasesFromDirectories(baselineDir, artifactDir, nameFilter, namePrefix string) (loadResult, error) {
	return loadCasesFromDirectorySources([]directorySource{
		{
			Suite:       "plots",
			Baseline:    filepath.Base(baselineDir),
			BaselineDir: baselineDir,
			ArtifactDir: artifactDir,
		},
	}, nameFilter, namePrefix)
}

type directorySource struct {
	Suite       string
	Baseline    string
	BaselineDir string
	ArtifactDir string
}

func loadCasesFromDirectorySources(sources []directorySource, nameFilter, namePrefix string) (loadResult, error) {
	filter := strings.ToLower(strings.TrimSpace(nameFilter))
	prefix := strings.ToLower(strings.TrimSpace(namePrefix))
	var result loadResult

	for _, source := range sources {
		baseline, err := filepath.Glob(filepath.Join(source.BaselineDir, "*.png"))
		if err != nil {
			return loadResult{}, fmt.Errorf("glob baseline %s: %w", source.BaselineDir, err)
		}
		sort.Strings(baseline)

		for _, baselinePath := range baseline {
			name := strings.TrimSuffix(filepath.Base(baselinePath), filepath.Ext(baselinePath))
			artifactPath := filepath.Join(source.ArtifactDir, filepath.Base(baselinePath))
			_, statErr := os.Stat(artifactPath)
			if os.IsNotExist(statErr) {
				result.SkippedCount++
				continue
			}
			if statErr != nil {
				return loadResult{}, fmt.Errorf("stat artifact %s: %w", artifactPath, statErr)
			}
			if filter != "" && !strings.Contains(strings.ToLower(name), filter) {
				continue
			}
			if prefix != "" && !strings.HasPrefix(strings.ToLower(name), prefix) {
				continue
			}

			entry, err := buildEntry(source.Suite, source.Baseline, name, baselinePath, artifactPath)
			if err != nil {
				return loadResult{}, fmt.Errorf("build entry for %s/%s/%s: %w", source.Suite, source.Baseline, name, err)
			}
			result.Cases = append(result.Cases, entry)
			result.ComparedCount++
		}
	}

	sort.Slice(result.Cases, func(i, j int) bool {
		if result.Cases[i].RMSE == result.Cases[j].RMSE {
			if result.Cases[i].Suite == result.Cases[j].Suite {
				return result.Cases[i].Name < result.Cases[j].Name
			}
			return result.Cases[i].Suite < result.Cases[j].Suite
		}
		return result.Cases[i].RMSE > result.Cases[j].RMSE
	})

	return result, nil
}

func loadCasesFromParityDir(parityDir, nameFilter, namePrefix string) (loadResult, error) {
	filter := strings.ToLower(strings.TrimSpace(nameFilter))
	prefix := strings.ToLower(strings.TrimSpace(namePrefix))
	var result loadResult

	suiteDirs, err := os.ReadDir(parityDir)
	if err != nil {
		return loadResult{}, fmt.Errorf("read parity dir %s: %w", parityDir, err)
	}

	for _, suiteDir := range suiteDirs {
		if !suiteDir.IsDir() {
			continue
		}
		suiteName := suiteDir.Name()
		suitePath := filepath.Join(parityDir, suiteName)

		children, err := os.ReadDir(suitePath)
		if err != nil {
			return loadResult{}, fmt.Errorf("read suite dir %s: %w", suitePath, err)
		}

		for _, child := range children {
			if !child.IsDir() || !strings.HasPrefix(child.Name(), "baseline-") {
				continue
			}

			baselineName := child.Name()
			baselineDir := filepath.Join(suitePath, baselineName)
			artifactDir := filepath.Join(suitePath, "artifacts")
			artifactBaselineDir := filepath.Join(artifactDir, baselineName)

			files, err := filepath.Glob(filepath.Join(baselineDir, "*.png"))
			if err != nil {
				return loadResult{}, fmt.Errorf("glob baselines in %s: %w", baselineDir, err)
			}
			sort.Strings(files)

			for _, baselinePath := range files {
				name := strings.TrimSuffix(filepath.Base(baselinePath), filepath.Ext(baselinePath))
				artifactPath := filepath.Join(artifactDir, filepath.Base(baselinePath))
				if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
					artifactPath = filepath.Join(artifactBaselineDir, filepath.Base(baselinePath))
					if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
						result.SkippedCount++
						continue
					}
				}
				if filter != "" && !strings.Contains(strings.ToLower(name), filter) {
					continue
				}
				if prefix != "" && !strings.HasPrefix(strings.ToLower(name), prefix) {
					continue
				}
				entry, err := buildEntry(suiteName, baselineName, name, baselinePath, artifactPath)
				if err != nil {
					return loadResult{}, fmt.Errorf("build entry %s/%s/%s: %w", suiteName, baselineName, name, err)
				}
				result.Cases = append(result.Cases, entry)
				result.ComparedCount++
			}
		}
	}

	sort.Slice(result.Cases, func(i, j int) bool {
		if result.Cases[i].RMSE == result.Cases[j].RMSE {
			if result.Cases[i].Suite == result.Cases[j].Suite {
				if result.Cases[i].Baseline == result.Cases[j].Baseline {
					return result.Cases[i].Name < result.Cases[j].Name
				}
				return result.Cases[i].Baseline < result.Cases[j].Baseline
			}
			return result.Cases[i].Suite < result.Cases[j].Suite
		}
		return result.Cases[i].RMSE > result.Cases[j].RMSE
	})

	return result, nil
}

func buildEntry(suite, baseline, name, baselinePath, artifactPath string) (caseEntry, error) {
	ref, err := readPNGAsRGBA(baselinePath)
	if err != nil {
		return caseEntry{}, fmt.Errorf("read baseline: %w", err)
	}
	act, err := readPNGAsRGBA(artifactPath)
	if err != nil {
		return caseEntry{}, fmt.Errorf("read artifact: %w", err)
	}

	rawDiff := rawDiffImage(ref, act)
	ampDiff := amplifiedDiffImage(ref, act)
	stats := compareImages(ref, act)

	refB64, err := pngToBase64(compositeOverSolid(ref, color.RGBA{R: 255, G: 255, B: 255, A: 255}))
	if err != nil {
		return caseEntry{}, err
	}
	actB64, err := pngToBase64(compositeOverSolid(act, color.RGBA{R: 255, G: 255, B: 255, A: 255}))
	if err != nil {
		return caseEntry{}, err
	}
	rawDiffB64, err := pngToBase64(rawDiff)
	if err != nil {
		return caseEntry{}, err
	}
	ampDiffB64, err := pngToBase64(ampDiff)
	if err != nil {
		return caseEntry{}, err
	}

	return caseEntry{
		Suite:       suite,
		Baseline:    baseline,
		Name:        name,
		RMSE:        stats.RMSE,
		AvgDiff:     stats.AvgDiff,
		MaxDiff:     stats.MaxDiff,
		DiffPixels:  stats.DiffPixels,
		TotalPixels: stats.TotalPixels,
		DiffRatio:   stats.DiffRatio,
		RefWidth:    ref.Bounds().Dx(),
		RefHeight:   ref.Bounds().Dy(),
		ActWidth:    act.Bounds().Dx(),
		ActHeight:   act.Bounds().Dy(),
		RefB64:      refB64,
		ActB64:      actB64,
		RawDiffB64:  rawDiffB64,
		AmpDiffB64:  ampDiffB64,
	}, nil
}

func compareImages(ref, act *image.RGBA) metrics {
	bounds := unionBounds(ref.Bounds(), act.Bounds())
	totalPixels := bounds.Dx() * bounds.Dy()
	if totalPixels <= 0 {
		return metrics{}
	}

	var (
		sumSq      float64
		totalDiff  float64
		diffPixels int
		maxDiff    uint8
	)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rr, rg, rb, ra := rgbaAt(ref, x, y)
			ar, ag, ab, aa := rgbaAt(act, x, y)

			dr := absDiff8(rr, ar)
			dg := absDiff8(rg, ag)
			db := absDiff8(rb, ab)
			da := absDiff8(ra, aa)

			pixelMax := max4(dr, dg, db, da)
			if pixelMax > maxDiff {
				maxDiff = pixelMax
			}
			if pixelMax != 0 {
				diffPixels++
			}

			totalDiff += float64(dr) + float64(dg) + float64(db) + float64(da)
			sumSq += sqDiff(rr, ar) + sqDiff(rg, ag) + sqDiff(rb, ab) + sqDiff(ra, aa)
		}
	}

	return metrics{
		RMSE:        math.Sqrt(sumSq / float64(totalPixels*4)),
		AvgDiff:     totalDiff / float64(totalPixels*4),
		MaxDiff:     maxDiff,
		DiffPixels:  diffPixels,
		TotalPixels: totalPixels,
		DiffRatio:   float64(diffPixels) / float64(totalPixels),
	}
}

func rawDiffImage(ref, act *image.RGBA) *image.RGBA {
	bounds := unionBounds(ref.Bounds(), act.Bounds())
	out := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			rr, rg, rb, ra := rgbaAt(ref, bounds.Min.X+x, bounds.Min.Y+y)
			ar, ag, ab, aa := rgbaAt(act, bounds.Min.X+x, bounds.Min.Y+y)
			dr := absDiff8(rr, ar)
			dg := absDiff8(rg, ag)
			db := absDiff8(rb, ab)
			da := absDiff8(ra, aa)
			if dr == 0 && dg == 0 && db == 0 && da == 0 {
				out.SetRGBA(x, y, color.RGBA{R: 0, G: 0xaa, B: 0, A: 255})
				continue
			}
			out.SetRGBA(x, y, color.RGBA{R: dr, G: dg, B: db, A: clampAlpha(da)})
		}
	}
	return out
}

func amplifiedDiffImage(ref, act *image.RGBA) *image.RGBA {
	bounds := unionBounds(ref.Bounds(), act.Bounds())
	out := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			rr, rg, rb, ra := rgbaAt(ref, bounds.Min.X+x, bounds.Min.Y+y)
			ar, ag, ab, aa := rgbaAt(act, bounds.Min.X+x, bounds.Min.Y+y)
			dr := absDiff8(rr, ar)
			dg := absDiff8(rg, ag)
			db := absDiff8(rb, ab)
			da := absDiff8(ra, aa)
			pixelMax := max4(dr, dg, db, da)
			if pixelMax == 0 {
				out.SetRGBA(x, y, color.RGBA{R: 0, G: 0xaa, B: 0, A: 255})
				continue
			}
			intensity := uint8(255)
			if pixelMax < 255 {
				intensity = uint8((float64(pixelMax) / 255.0) * 255.0)
			}
			out.SetRGBA(x, y, color.RGBA{R: intensity, G: 0, B: 0, A: 255})
		}
	}
	return out
}

func rgbaAt(img *image.RGBA, x, y int) (uint8, uint8, uint8, uint8) {
	if img == nil || !image.Pt(x, y).In(img.Bounds()) {
		return 0, 0, 0, 0
	}
	i := img.PixOffset(x, y)
	return img.Pix[i+0], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3]
}

func unionBounds(a, b image.Rectangle) image.Rectangle {
	minX := minInt(a.Min.X, b.Min.X)
	minY := minInt(a.Min.Y, b.Min.Y)
	maxX := maxInt(a.Max.X, b.Max.X)
	maxY := maxInt(a.Max.Y, b.Max.Y)
	if maxX < minX {
		maxX = minX
	}
	if maxY < minY {
		maxY = minY
	}
	return image.Rect(minX, minY, maxX, maxY)
}

func max4(a, b, c, d uint8) uint8 {
	if a < b {
		a = b
	}
	if a < c {
		a = c
	}
	if a < d {
		a = d
	}
	return a
}

func absDiff8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func sqDiff(a, b uint8) float64 {
	d := float64(int(a) - int(b))
	return d * d
}

func pngToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func readPNGAsRGBA(path string) (*image.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	rgba := image.NewRGBA(img.Bounds())
	for y := rgba.Bounds().Min.Y; y < rgba.Bounds().Max.Y; y++ {
		for x := rgba.Bounds().Min.X; x < rgba.Bounds().Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba, nil
}

func clampAlpha(a uint8) uint8 {
	if a == 0 {
		return 255
	}
	return a
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func compositeOverSolid(src *image.RGBA, bg color.RGBA) *image.RGBA {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := rgbaAt(src, x, y)
			out.SetRGBA(x-bounds.Min.X, y-bounds.Min.Y, compositePixel(r, g, b, a, bg))
		}
	}
	return out
}

func compositePixel(r, g, b, a uint8, bg color.RGBA) color.RGBA {
	srcA := int(a)
	invA := 255 - srcA
	return color.RGBA{
		R: uint8((int(r)*srcA + int(bg.R)*invA) / 255),
		G: uint8((int(g)*srcA + int(bg.G)*invA) / 255),
		B: uint8((int(b)*srcA + int(bg.B)*invA) / 255),
		A: 255,
	}
}

func renderPage(w io.Writer, result loadResult, opts viewOptions) {
	fmt.Fprint(w, strings.Replace(pageHeader, rerenderAllButtonPlaceholder, buildRerenderAllButton(opts), 1))
	fmt.Fprintf(
		w,
		`<div class="header-meta">%d comparisons loaded`,
		result.ComparedCount,
	)
	if result.SkippedCount > 0 {
		fmt.Fprintf(w, `, %d baselines skipped because no matching artifact exists`, result.SkippedCount)
	}
	if !opts.CanRerender && opts.RerenderDisabledMsg != "" {
		fmt.Fprintf(w, ` <span class="header-warning">%s</span>`, htmlEscape(opts.RerenderDisabledMsg))
	}
	fmt.Fprint(w, `.</div></div>`)

	fmt.Fprint(w, `<div class="container" id="cards-container">`)
	for i := range result.Cases {
		renderCard(w, &result.Cases[i], opts)
	}
	if len(result.Cases) == 0 {
		fmt.Fprint(w, `<div class="empty-state">No parity comparisons found. Point --baseline-dir/--artifact-dir at existing PNG sets.</div>`)
	}
	fmt.Fprint(w, pageFooter)
}

func buildRerenderAllButton(opts viewOptions) string {
	if opts.CanRerender {
		return `<button id="rerender-all-btn" class="rerender-btn" type="button">Re-render Artifacts</button>`
	}
	return fmt.Sprintf(
		`<button id="rerender-all-btn" class="rerender-btn" type="button" disabled title=%q>Re-render Artifacts</button>`,
		htmlEscape(opts.RerenderDisabledMsg),
	)
}

func renderCard(w io.Writer, entry *caseEntry, opts viewOptions) {
	if entry == nil {
		return
	}

	fmt.Fprintf(
		w,
		`<div class="card" data-name="%s" data-suite="%s" data-baseline="%s" data-rmse="%.4f" data-avg-diff="%.4f" data-max-diff="%d" data-diff-pixels="%d" data-diff-ratio="%.6f">`,
		htmlEscape(entry.Name),
		htmlEscape(entry.Suite),
		htmlEscape(entry.Baseline),
		entry.RMSE,
		entry.AvgDiff,
		entry.MaxDiff,
		entry.DiffPixels,
		entry.DiffRatio,
	)
	fmt.Fprint(w, `<div class="card-header">`)
	fmt.Fprint(w, `<span class="badge badge-neutral sort-metric-badge" style="display:none"></span>`)
	fmt.Fprintf(w, `<span class="card-title">%s</span>`, htmlEscape(entry.Name))
	if opts.CanRerender {
		fmt.Fprintf(
			w,
			`<button class="rerender-btn" data-suite="%s" data-name="%s" type="button">Re-render Artifact</button>`,
			htmlEscape(entry.Suite),
			htmlEscape(entry.Name),
		)
	} else {
		fmt.Fprintf(
			w,
			`<button class="rerender-btn" data-suite="%s" data-name="%s" type="button" disabled title="%s">Re-render Artifact</button>`,
			htmlEscape(entry.Suite),
			htmlEscape(entry.Name),
			htmlEscape(opts.RerenderDisabledMsg),
		)
	}
	fmt.Fprint(w, `<div class="right-badges">`)
	fmt.Fprintf(w, `<span class="badge badge-neutral">%s</span>`, htmlEscape(entry.Suite))
	fmt.Fprintf(w, `<span class="badge badge-neutral">%s</span>`, htmlEscape(entry.Baseline))
	fmt.Fprintf(w, `<span class="badge %s">RMSE %.2f</span>`, badgeClassRMSE(entry.RMSE), entry.RMSE)
	fmt.Fprintf(w, `<span class="badge %s">avg %.2f</span>`, badgeClassAvgDiff(entry.AvgDiff), entry.AvgDiff)
	fmt.Fprintf(w, `<span class="badge %s">max %d</span>`, badgeClassMaxDiff(entry.MaxDiff), entry.MaxDiff)
	fmt.Fprintf(w, `<span class="badge %s">diff %.2f%%</span>`, badgeClassDiffRatio(entry.DiffRatio), entry.DiffRatio*100)
	fmt.Fprint(w, `</div></div>`)

	fmt.Fprint(w, `<div class="card-body"><div class="card-meta">`)
	fmt.Fprintf(w, `size baseline %dx%d, artifact %dx%d`, entry.RefWidth, entry.RefHeight, entry.ActWidth, entry.ActHeight)
	fmt.Fprint(w, `</div><div class="img-grid">`)

	fmt.Fprint(w, `<div class="img-col col-ref">`)
	fmt.Fprint(w, `<label>Baseline</label>`)
	fmt.Fprint(w, `<div class="zoom-surface">`)
	fmt.Fprint(w, `<div class="zoom-transform">`)
	fmt.Fprintf(w, `<img class="parity-image" src="data:image/png;base64,%s" alt="baseline">`, entry.RefB64)
	fmt.Fprint(w, `</div><div class="zoom-selection"></div></div>`)
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `<div class="img-col col-artifact">`)
	fmt.Fprint(w, `<label>Artifact</label>`)
	fmt.Fprint(w, `<div class="zoom-surface">`)
	fmt.Fprint(w, `<div class="zoom-transform">`)
	fmt.Fprintf(w, `<img class="parity-image" src="data:image/png;base64,%s" alt="artifact">`, entry.ActB64)
	fmt.Fprint(w, `</div><div class="zoom-selection"></div></div>`)
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `<div class="img-col col-overlay">`)
	fmt.Fprint(w, `<label>Overlay</label>`)
	fmt.Fprint(w, `<div class="slider-wrap zoom-surface">`)
	fmt.Fprint(w, `<div class="zoom-transform zoom-base-layer">`)
	fmt.Fprintf(w, `<img class="base" src="data:image/png;base64,%s" alt="base">`, entry.RefB64)
	fmt.Fprint(w, `</div>`)
	fmt.Fprint(w, `<div class="slider-overlay">`)
	fmt.Fprint(w, `<div class="zoom-transform zoom-overlay-layer">`)
	fmt.Fprintf(w, `<img src="data:image/png;base64,%s" alt="overlay">`, entry.ActB64)
	fmt.Fprint(w, `</div></div><div class="slider-divider"></div><div class="zoom-selection"></div></div></div>`)

	fmt.Fprint(w, `<div class="img-col col-amp">`)
	fmt.Fprint(w, `<label>Diff amplified</label>`)
	fmt.Fprint(w, `<div class="zoom-surface">`)
	fmt.Fprint(w, `<div class="zoom-transform">`)
	fmt.Fprintf(w, `<img class="parity-image" src="data:image/png;base64,%s" alt="amplified-diff">`, entry.AmpDiffB64)
	fmt.Fprint(w, `</div><div class="zoom-selection"></div></div>`)
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `<div class="img-col col-raw">`)
	fmt.Fprint(w, `<label>Diff raw</label>`)
	fmt.Fprint(w, `<div class="zoom-surface">`)
	fmt.Fprint(w, `<div class="zoom-transform">`)
	fmt.Fprintf(w, `<img class="parity-image" src="data:image/png;base64,%s" alt="raw-diff">`, entry.RawDiffB64)
	fmt.Fprint(w, `</div><div class="zoom-selection"></div></div>`)
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `</div></div></div>`)
}

func htmlEscape(s string) string {
	return html.EscapeString(s)
}

const (
	badgeClassOK   = "badge-ok"
	badgeClassWarn = "badge-warn"
	badgeClassBad  = "badge-bad"
)

func badgeClassRMSE(v float64) string {
	if v <= 5 {
		return badgeClassOK
	}
	if v <= 20 {
		return badgeClassWarn
	}
	return badgeClassBad
}

func badgeClassAvgDiff(v float64) string {
	if v <= 2 {
		return badgeClassOK
	}
	if v <= 8 {
		return badgeClassWarn
	}
	return badgeClassBad
}

func badgeClassMaxDiff(v uint8) string {
	if v <= 10 {
		return badgeClassOK
	}
	if v <= 40 {
		return badgeClassWarn
	}
	return badgeClassBad
}

func badgeClassDiffRatio(v float64) string {
	if v <= 0.01 {
		return badgeClassOK
	}
	if v <= 0.05 {
		return badgeClassWarn
	}
	return badgeClassBad
}

const pageHeader = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Matplotlib-Go Parity Viewer</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { background: #101216; color: #d7dce4; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 13px; }
.sticky-header { position: sticky; top: 0; z-index: 100; background: rgba(16,18,22,0.96); backdrop-filter: blur(10px); border-bottom: 1px solid #2d3440; padding: 10px 12px; }
.controls { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.controls h1 { font-size: 15px; color: #f4f7fb; margin-right: 8px; }
.controls input, .controls select { background: #171b22; color: #d7dce4; border: 1px solid #334155; padding: 5px 8px; font-family: inherit; font-size: 12px; border-radius: 6px; }
.controls label { display: flex; align-items: center; gap: 5px; font-size: 12px; color: #aeb8c8; }
.header-meta { margin-top: 8px; color: #93a1b5; font-size: 12px; }
#summary { margin-left: auto; color: #93a1b5; font-size: 12px; }
.container { padding: 12px; }
.card { background: #141922; border: 1px solid #293241; margin-bottom: 10px; border-radius: 10px; overflow: hidden; box-shadow: 0 8px 24px rgba(0,0,0,0.2); }
.card-header { padding: 9px 12px; cursor: pointer; display: flex; align-items: center; gap: 8px; background: #181f2a; user-select: none; }
.card-header:hover { background: #1d2531; }
.card-body { display: none; padding: 12px; }
.card.open .card-body { display: block; }
.card-title { font-size: 13px; color: #f4f7fb; flex: 1; }
.card-meta { color: #93a1b5; margin-bottom: 10px; }
.right-badges { display: flex; gap: 6px; align-items: center; flex-wrap: wrap; }
.badge { padding: 2px 7px; border-radius: 999px; font-size: 11px; font-weight: bold; }
.badge-neutral { background: #202734; color: #d7dce4; border: 1px solid #425168; }
.badge-ok { background: #163323; color: #7cf0a5; border: 1px solid #2e7250; }
.badge-warn { background: #3a2d13; color: #ffc66d; border: 1px solid #6f5521; }
.badge-bad { background: #3b1619; color: #ff8a8f; border: 1px solid #7c2d35; }
.img-grid { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 10px; align-items: start; overflow-x: auto; }
.img-col { display: flex; flex-direction: column; gap: 6px; min-width: 0; overflow: auto; }
.img-col label { font-size: 11px; color: #93a1b5; text-align: center; }
.zoom-surface { position: relative; overflow: hidden; border-radius: 6px; background-color: #fff; cursor: crosshair; max-width: 100%; }
.zoom-transform { position: relative; transform-origin: 0 0; will-change: transform; }
.zoom-selection { display: none; position: absolute; border: 1px solid #93c5fd; background: rgba(96, 165, 250, 0.18); pointer-events: none; z-index: 4; }
.parity-image { display: block; image-rendering: auto; width: 100%; height: auto; background-color: #fff; background-image: none; max-width: 100%; }
.resample-pixelated .parity-image { image-rendering: pixelated; }
.original-size .img-grid { grid-template-columns: repeat(5, max-content); }
.original-size .img-col { min-width: max-content; }
.original-size .zoom-surface, .original-size .slider-wrap { align-self: flex-start; }
.original-size .parity-image { width: auto; height: auto; max-width: none; }
.col-raw, .col-amp { display: none; }
.slider-wrap { width: 100%; background-color: #fff; background-image: none; }
.slider-wrap img.base, .slider-wrap .zoom-overlay-layer img { display: block; image-rendering: auto; width: 100%; height: auto; }
.resample-pixelated .slider-wrap img.base { image-rendering: pixelated; }
.resample-pixelated .slider-wrap .zoom-overlay-layer img { image-rendering: pixelated; }
.original-size .slider-wrap, .original-size .zoom-surface { width: auto; }
.original-size .slider-wrap img.base { width: auto; height: auto; max-width: none; }
.original-size .slider-wrap .zoom-overlay-layer img { width: auto; height: auto; max-width: none; }
.slider-overlay { position: absolute; top: 0; left: 0; height: 100%; overflow: hidden; width: 50%; z-index: 2; pointer-events: none; }
.slider-divider { position: absolute; top: 0; left: 50%; height: 100%; width: 3px; background: #f8fafc; cursor: col-resize; transform: translateX(-50%); }
.slider-divider::before {
  content: ''; position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%);
  width: 18px; height: 18px; border-radius: 50%; background: #f8fafc; border: 2px solid #111827;
}
.zoom-dragging, .zoom-dragging * { user-select: none; }
.empty-state { border: 1px dashed #425168; border-radius: 10px; padding: 24px; color: #93a1b5; background: #141922; }
button.rerender-btn { background: #0f172a; color: #e2e8f0; border: 1px solid #334155; border-radius: 6px; padding: 4px 8px; cursor: pointer; font-family: inherit; font-size: 11px; }
button.rerender-btn:hover { background: #273244; }
button.rerender-btn:disabled { opacity: 0.6; cursor: not-allowed; }
button.rerender-btn.is-rerendering { cursor: wait; }
code { color: #f4f7fb; }
@media (max-width: 1400px) { .img-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 800px) { .img-grid { grid-template-columns: 1fr; } }
</style>
</head>
<body>
<div class="sticky-header">
  <div class="controls">
    <h1>Matplotlib-Go Parity Viewer</h1>
    <input type="text" id="search" placeholder="Search case…" oninput="filterCards()" style="width:180px">
    <select id="sort-select" onchange="sortCards()">
      <option value="rmse-desc">Sort: RMSE ↓</option>
      <option value="rmse-asc">Sort: RMSE ↑</option>
      <option value="diff-pixels-desc">Sort: Different pixels ↓</option>
      <option value="diff-pixels-asc">Sort: Different pixels ↑</option>
      <option value="diff-ratio-desc">Sort: Diff ratio ↓</option>
      <option value="diff-ratio-asc">Sort: Diff ratio ↑</option>
      <option value="avg-diff-desc">Sort: Avg diff ↓</option>
      <option value="avg-diff-asc">Sort: Avg diff ↑</option>
      <option value="name-asc">Sort: Name ↑</option>
    </select>
    <select id="diff-mode" onchange="setDiffMode(this.value)">
      <option value="amp">Diff: amplified</option>
      <option value="raw">Diff: raw</option>
      <option value="both">Diff: both</option>
    </select>
    <select id="resample-mode" onchange="setResampleMode(this.value)">
      <option value="smooth">Scaling: smooth</option>
      <option value="pixelated">Scaling: pixelated</option>
    </select>
    __RERENDER_ALL_BUTTON__
    <label><input type="checkbox" id="original-size" onchange="setOriginalSize(this.checked)"> Original size</label>
    <span id="summary"></span>
  </div>
`

const pageFooter = `</div>
<script>
(function() {
  var viewerStateStorageKey = 'mpl-parity-viewer-state-v1';
  var viewerStateControlIDs = ['search', 'sort-select', 'diff-mode', 'resample-mode', 'original-size'];
  var minZoomDragPixels = 4;
  var activeSlider = null;
  var activeZoomSelection = null;

  function cardStateKey(card) {
    return [card.dataset.suite || '', card.dataset.baseline || '', card.dataset.name || ''].join('::');
  }

  function loadViewerState() {
    try {
      var raw = window.sessionStorage.getItem(viewerStateStorageKey);
      if (!raw) return {};
      var parsed = JSON.parse(raw);
      if (!parsed || typeof parsed !== 'object') return {};
      return parsed;
    } catch (err) {
      return {};
    }
  }

  function saveViewerState() {
    var state = {};
    viewerStateControlIDs.forEach(function(id) {
      var el = document.getElementById(id);
      if (!el) return;
      state[id] = el.type === 'checkbox' ? el.checked : el.value;
    });
    state.openCards = Array.from(document.querySelectorAll('.card.open')).map(cardStateKey);
    try {
      window.sessionStorage.setItem(viewerStateStorageKey, JSON.stringify(state));
    } catch (err) {
    }
  }

  function restoreViewerState() {
    var state = loadViewerState();
    viewerStateControlIDs.forEach(function(id) {
      if (!Object.prototype.hasOwnProperty.call(state, id)) return;
      var el = document.getElementById(id);
      if (!el) return;
      if (el.type === 'checkbox') {
        el.checked = !!state[id];
        return;
      }
      if (typeof state[id] === 'string') {
        el.value = state[id];
      }
    });
    var openCards = Array.isArray(state.openCards) ? state.openCards : [];
    var openCardSet = new Set(openCards);
    document.querySelectorAll('.card').forEach(function(card) {
      card.classList.toggle('open', openCardSet.has(cardStateKey(card)));
    });
  }

  function bindViewerStatePersistence() {
    viewerStateControlIDs.forEach(function(id) {
      var el = document.getElementById(id);
      if (!el) return;
      el.addEventListener(el.type === 'text' ? 'input' : 'change', saveViewerState);
    });
  }

  function metric(card, attr) {
    return parseFloat(card.dataset[attr] || 0);
  }

  function clamp(value, minValue, maxValue) {
    return Math.min(Math.max(value, minValue), maxValue);
  }

  function filterCards() {
    var q = document.getElementById('search').value.toLowerCase();
    document.querySelectorAll('.card').forEach(function(card) {
      var name = (card.dataset.name || '').toLowerCase();
      card.style.display = name.indexOf(q) >= 0 ? '' : 'none';
    });
    updateSummary();
  }

  function sortCards() {
    var mode = document.getElementById('sort-select').value;
    var container = document.getElementById('cards-container');
    var cards = Array.from(container.querySelectorAll('.card'));
    cards.sort(function(a, b) {
      if (mode === 'rmse-desc') return metric(b, 'rmse') - metric(a, 'rmse');
      if (mode === 'rmse-asc') return metric(a, 'rmse') - metric(b, 'rmse');
      if (mode === 'diff-pixels-desc') return metric(b, 'diffPixels') - metric(a, 'diffPixels');
      if (mode === 'diff-pixels-asc') return metric(a, 'diffPixels') - metric(b, 'diffPixels');
      if (mode === 'diff-ratio-desc') return metric(b, 'diffRatio') - metric(a, 'diffRatio');
      if (mode === 'diff-ratio-asc') return metric(a, 'diffRatio') - metric(b, 'diffRatio');
      if (mode === 'avg-diff-desc') return metric(b, 'avgDiff') - metric(a, 'avgDiff');
      if (mode === 'avg-diff-asc') return metric(a, 'avgDiff') - metric(b, 'avgDiff');
      return (a.dataset.name || '').localeCompare(b.dataset.name || '');
    });
    cards.forEach(function(card) { container.appendChild(card); });
    updateSortMetricBadges(mode);
  }

  function badgeColorClass(attr, value, diffRatio) {
    if (attr === 'rmse') {
      if (value <= 5) return 'badge-ok';
      if (value <= 20) return 'badge-warn';
      return 'badge-bad';
    }
    if (attr === 'avgDiff') {
      if (value <= 2) return 'badge-ok';
      if (value <= 8) return 'badge-warn';
      return 'badge-bad';
    }
    if (attr === 'maxDiff') {
      if (value <= 10) return 'badge-ok';
      if (value <= 40) return 'badge-warn';
      return 'badge-bad';
    }
    if (attr === 'diffPixels' || attr === 'diffRatio') {
      var ratio = parseFloat(diffRatio || 0);
      if (ratio <= 0.01) return 'badge-ok';
      if (ratio <= 0.05) return 'badge-warn';
      return 'badge-bad';
    }
    return 'badge-neutral';
  }

  function updateSortMetricBadges(mode) {
    var label = '';
    var attr = '';
    var format = function(v) { return String(v); };
    if (mode === 'rmse-desc' || mode === 'rmse-asc') {
      label = 'RMSE'; attr = 'rmse'; format = function(v) { return Number(v).toFixed(2); };
    } else if (mode === 'diff-pixels-desc' || mode === 'diff-pixels-asc') {
      label = 'diff px'; attr = 'diffPixels'; format = function(v) { return String(Math.round(Number(v))); };
    } else if (mode === 'diff-ratio-desc' || mode === 'diff-ratio-asc') {
      label = 'diff %'; attr = 'diffRatio'; format = function(v) { return (Number(v) * 100).toFixed(2); };
    } else if (mode === 'avg-diff-desc' || mode === 'avg-diff-asc') {
      label = 'avg'; attr = 'avgDiff'; format = function(v) { return Number(v).toFixed(2); };
    }
    document.querySelectorAll('.card').forEach(function(card) {
      var badge = card.querySelector('.sort-metric-badge');
      if (!badge) return;
      if (!attr) {
        badge.style.display = 'none';
        return;
      }
      var value = parseFloat(card.dataset[attr] || 0);
      badge.className = 'badge ' + badgeColorClass(attr, value, card.dataset.diffRatio) + ' sort-metric-badge';
      badge.textContent = label + ' ' + format(value);
      badge.style.display = '';
    });
  }

  function setDiffMode(mode) {
    document.querySelectorAll('.col-amp').forEach(function(el) {
      el.style.display = (mode === 'amp' || mode === 'both') ? 'flex' : 'none';
    });
    document.querySelectorAll('.col-raw').forEach(function(el) {
      el.style.display = (mode === 'raw' || mode === 'both') ? 'flex' : 'none';
    });
    refreshCardZooms();
  }

  function updateSummary() {
    var all = document.querySelectorAll('.card');
    var visible = Array.from(all).filter(function(card) { return card.style.display !== 'none'; });
    document.getElementById('summary').textContent = visible.length + ' / ' + all.length + ' cases';
  }

  function setResampleMode(mode) {
    var container = document.getElementById('cards-container');
    container.classList.remove('resample-smooth', 'resample-pixelated');
    container.classList.add(mode === 'pixelated' ? 'resample-pixelated' : 'resample-smooth');
  }

  function setOriginalSize(on) {
    var container = document.getElementById('cards-container');
    if (on) container.classList.add('original-size'); else container.classList.remove('original-size');
    refreshCardZooms();
  }

  function setRerenderButtonsDisabled(disabled) {
    document.querySelectorAll('.rerender-btn').forEach(function(button) {
      button.disabled = disabled;
      button.classList.toggle('is-rerendering', disabled);
    });
  }

  function rerenderArtifact(name) {
    return fetch('/rerender', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8' },
      body: new URLSearchParams({ name: name }).toString()
    }).then(function(response) {
      if (!response.ok) {
        return response.text().then(function(text) {
          throw new Error(text || 'rerender failed');
        });
      }
    });
  }

  function navigateToFreshPage() {
    var url = new URL(window.location.href);
    url.searchParams.set('_pv', String(Date.now()));
    window.location.assign(url.toString());
  }

  document.querySelectorAll('.card-header').forEach(function(header) {
    header.addEventListener('click', function() {
      header.closest('.card').classList.toggle('open');
      saveViewerState();
    });
  });

  document.querySelectorAll('.rerender-btn').forEach(function(button) {
    if (button.id === 'rerender-all-btn') {
      return;
    }
    button.addEventListener('click', function(event) {
      event.stopPropagation();
      var name = button.dataset.name || '';
      if (!name) return;
      saveViewerState();
      setRerenderButtonsDisabled(true);
      rerenderArtifact(name).then(function() {
        navigateToFreshPage();
      }).catch(function(err) {
        window.alert(err.message);
        setRerenderButtonsDisabled(false);
      });
    });
  });

  var bulkButton = document.getElementById('rerender-all-btn');
  bulkButton.addEventListener('click', function() {
    saveViewerState();
    setRerenderButtonsDisabled(true);
    var cards = Array.from(document.querySelectorAll('.card'));
    var chain = Promise.resolve();
    cards.forEach(function(card) {
      var name = card.dataset.name || '';
      if (!name) return;
      chain = chain.then(function() { return rerenderArtifact(name); });
    });
    chain.then(function() {
      navigateToFreshPage();
    }).catch(function(err) {
      window.alert(err.message);
      setRerenderButtonsDisabled(false);
    });
  });

  function ensureCardZoomState(card) {
    if (!card.__zoomState) {
      card.__zoomState = { scale: 1, x: 0, y: 0 };
    }
    return card.__zoomState;
  }

  function applyCardZoom(card) {
    var state = ensureCardZoomState(card);
    card.querySelectorAll('.zoom-surface').forEach(function(surface) {
      var width = surface.clientWidth;
      var height = surface.clientHeight;
      surface.querySelectorAll('.zoom-transform').forEach(function(layer) {
        if (!width || !height || state.scale <= 1) {
          layer.style.transform = '';
          return;
        }
        var tx = -state.x * width * state.scale;
        var ty = -state.y * height * state.scale;
        layer.style.transform = 'matrix(' + state.scale + ',0,0,' + state.scale + ',' + tx + ',' + ty + ')';
      });
    });
  }

  function refreshCardZooms() {
    document.querySelectorAll('.card').forEach(applyCardZoom);
  }

  function setCardZoomFromSelection(card, rect) {
    var width = clamp(rect.width, 0, 1);
    var height = clamp(rect.height, 0, 1);
    if (width <= 0 || height <= 0) {
      return;
    }
    var scale = 1 / Math.max(width, height);
    var visibleWidth = 1 / scale;
    var visibleHeight = 1 / scale;
    var centerX = rect.x + width / 2;
    var centerY = rect.y + height / 2;
    var state = ensureCardZoomState(card);
    state.scale = scale;
    state.x = clamp(centerX - visibleWidth / 2, 0, Math.max(0, 1 - visibleWidth));
    state.y = clamp(centerY - visibleHeight / 2, 0, Math.max(0, 1 - visibleHeight));
    applyCardZoom(card);
  }

  function resetCardZoom(card) {
    var state = ensureCardZoomState(card);
    state.scale = 1;
    state.x = 0;
    state.y = 0;
    applyCardZoom(card);
  }

  function showSelectionBox(selection, x0, y0, x1, y1) {
    var left = Math.min(x0, x1);
    var top = Math.min(y0, y1);
    selection.style.display = 'block';
    selection.style.left = left + 'px';
    selection.style.top = top + 'px';
    selection.style.width = Math.abs(x1 - x0) + 'px';
    selection.style.height = Math.abs(y1 - y0) + 'px';
  }

  function hideSelectionBox(selection) {
    if (!selection) return;
    selection.style.display = 'none';
    selection.style.left = '0';
    selection.style.top = '0';
    selection.style.width = '0';
    selection.style.height = '0';
  }

  document.querySelectorAll('.slider-wrap').forEach(function(wrap) {
    var divider = wrap.querySelector('.slider-divider');
    var overlay = wrap.querySelector('.slider-overlay');
    function applyPos(pct) {
      pct = Math.max(0, Math.min(1, pct));
      wrap.__sliderPct = pct;
      overlay.style.width = (pct * 100) + '%';
      divider.style.left = (pct * 100) + '%';
      var overlayLayer = wrap.querySelector('.zoom-overlay-layer');
      if (overlayLayer) {
        if (pct <= 0) {
          overlayLayer.style.width = '0%';
          return;
        }
        overlayLayer.style.width = (100 / pct) + '%';
      }
    }
    function setPos(x) {
      var rect = wrap.getBoundingClientRect();
      if (rect.width <= 0) {
        applyPos(0.5);
        return;
      }
      applyPos((x - rect.left) / rect.width);
    }
    wrap.__setSliderClientX = setPos;
    divider.addEventListener('mousedown', function(e) {
      activeSlider = { wrap: wrap, setPos: setPos };
      e.preventDefault();
      e.stopPropagation();
    });
    applyPos(0.5);
  });

  document.addEventListener('mousemove', function(e) {
    if (activeSlider) {
      activeSlider.setPos(e.clientX);
    }
    if (!activeZoomSelection) {
      return;
    }
    var rect = activeZoomSelection.rect;
    var x = clamp(e.clientX - rect.left, 0, rect.width);
    var y = clamp(e.clientY - rect.top, 0, rect.height);
    activeZoomSelection.currentX = x;
    activeZoomSelection.currentY = y;
    showSelectionBox(activeZoomSelection.selection, activeZoomSelection.startX, activeZoomSelection.startY, x, y);
  });

  document.addEventListener('mouseup', function(e) {
    if (activeSlider && e.button === 0) {
      activeSlider = null;
    }
    if (!activeZoomSelection || e.button !== 0) {
      return;
    }
    var drag = activeZoomSelection;
    var width = Math.abs(drag.currentX - drag.startX);
    var height = Math.abs(drag.currentY - drag.startY);
    hideSelectionBox(drag.selection);
    document.body.classList.remove('zoom-dragging');
    activeZoomSelection = null;

    if (width > minZoomDragPixels && height > minZoomDragPixels) {
      setCardZoomFromSelection(drag.card, {
        x: Math.min(drag.startX, drag.currentX) / drag.rect.width,
        y: Math.min(drag.startY, drag.currentY) / drag.rect.height,
        width: width / drag.rect.width,
        height: height / drag.rect.height,
      });
      return;
    }

    if (drag.surface.classList.contains('slider-wrap') && typeof drag.surface.__setSliderClientX === 'function') {
      drag.surface.__setSliderClientX(e.clientX);
    }
  });

  document.querySelectorAll('.zoom-surface').forEach(function(surface) {
    surface.addEventListener('mousedown', function(e) {
      if (e.button !== 0) {
        return;
      }
      if (e.target.closest('.slider-divider')) {
        return;
      }
      var rect = surface.getBoundingClientRect();
      if (rect.width <= 0 || rect.height <= 0) {
        return;
      }
      var selection = surface.querySelector('.zoom-selection');
      if (!selection) {
        return;
      }
      var startX = clamp(e.clientX - rect.left, 0, rect.width);
      var startY = clamp(e.clientY - rect.top, 0, rect.height);
      activeZoomSelection = {
        card: surface.closest('.card'),
        surface: surface,
        rect: rect,
        selection: selection,
        startX: startX,
        startY: startY,
        currentX: startX,
        currentY: startY,
      };
      showSelectionBox(selection, startX, startY, startX, startY);
      document.body.classList.add('zoom-dragging');
      e.preventDefault();
    });
    surface.addEventListener('contextmenu', function(e) {
      e.preventDefault();
      resetCardZoom(surface.closest('.card'));
    });
  });

  window.addEventListener('resize', refreshCardZooms);

  window.filterCards = filterCards;
  window.sortCards = sortCards;
  window.setDiffMode = setDiffMode;
  window.setOriginalSize = setOriginalSize;
  window.setResampleMode = setResampleMode;

  restoreViewerState();
  bindViewerStatePersistence();
  sortCards();
  setDiffMode(document.getElementById('diff-mode').value);
  setResampleMode(document.getElementById('resample-mode').value);
  setOriginalSize(document.getElementById('original-size').checked);
  refreshCardZooms();
  filterCards();
})();
</script>
</body>
</html>
`
