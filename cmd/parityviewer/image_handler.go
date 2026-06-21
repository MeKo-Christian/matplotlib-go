package main

import (
	"errors"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	imageKindBaseline = "baseline"
	imageKindArtifact = "artifact"
	imageKindDiffRaw  = "diff-raw"
	imageKindDiffAmp  = "diff-amp"
)

func handleImage(w http.ResponseWriter, r *http.Request, opts cliOptions) {
	query := r.URL.Query()
	kind := strings.TrimSpace(query.Get("kind"))
	source, err := resolveImageSource(opts, query.Get("suite"), query.Get("baseline"), query.Get("name"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Content-Type", "image/png")
	switch kind {
	case imageKindBaseline:
		http.ServeFile(w, r, source.baselinePath)
	case imageKindArtifact:
		http.ServeFile(w, r, source.artifactPath)
	case imageKindDiffRaw, imageKindDiffAmp:
		ref, err := readPNGAsRGBA(source.baselinePath)
		if err != nil {
			http.Error(w, "read baseline: "+err.Error(), http.StatusInternalServerError)
			return
		}
		act, err := readPNGAsRGBA(source.artifactPath)
		if err != nil {
			http.Error(w, "read artifact: "+err.Error(), http.StatusInternalServerError)
			return
		}
		img := rawDiffImage(ref, act)
		if kind == imageKindDiffAmp {
			img = amplifiedDiffImage(ref, act)
		}
		if err := png.Encode(w, img); err != nil {
			http.Error(w, "encode diff: "+err.Error(), http.StatusInternalServerError)
		}
	default:
		http.Error(w, "unknown image kind", http.StatusBadRequest)
	}
}

type imageSource struct {
	baselinePath string
	artifactPath string
}

func resolveImageSource(opts cliOptions, suite, baseline, name string) (imageSource, error) {
	suite = strings.TrimSpace(suite)
	baseline = strings.TrimSpace(baseline)
	name = strings.TrimSpace(name)
	if !validImageComponent(suite) || !validImageComponent(baseline) || !validImageComponent(name) {
		return imageSource{}, errors.New("invalid image request")
	}

	var baselineDir, artifactDir string
	if opts.UseParity {
		baselineDir = filepath.Join(opts.ParityDir, suite, baseline)
		artifactDir = filepath.Join(opts.ParityDir, suite, "artifacts")
		source := imageSource{
			baselinePath: filepath.Join(baselineDir, name+".png"),
			artifactPath: filepath.Join(artifactDir, name+".png"),
		}
		if _, err := os.Stat(source.artifactPath); os.IsNotExist(err) {
			source.artifactPath = filepath.Join(artifactDir, baseline, name+".png")
		}
		return source, requireImageFiles(source)
	}

	if opts.IncludeWebdemo && suite == "webdemo" {
		baselineDir = opts.WebBaselineDir
		artifactDir = opts.WebArtifactDir
	} else {
		baselineDir = opts.BaselineDir
		artifactDir = opts.ArtifactDir
	}

	source := imageSource{
		baselinePath: filepath.Join(baselineDir, name+".png"),
		artifactPath: filepath.Join(artifactDir, name+".png"),
	}
	return source, requireImageFiles(source)
}

func requireImageFiles(source imageSource) error {
	if _, err := os.Stat(source.baselinePath); err != nil {
		return err
	}
	if _, err := os.Stat(source.artifactPath); err != nil {
		return err
	}
	return nil
}

func validImageComponent(s string) bool {
	if s == "" {
		return false
	}
	return filepath.Base(s) == s && !strings.Contains(s, string(filepath.Separator))
}
