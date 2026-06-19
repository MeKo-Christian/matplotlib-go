package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
)

type cliOptions struct {
	Port           string
	RepoRoot       string
	UseParity      bool
	ParityDir      string
	BaselineDir    string
	ArtifactDir    string
	IncludeWebdemo bool
	WebBaselineDir string
	WebArtifactDir string
	NameFilter     string
	NamePrefix     string
	PrintOnly      bool
}

func parseCLIOptions() (cliOptions, error) {
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
		return cliOptions{}, err
	}

	return cliOptions{
		Port:           *port,
		RepoRoot:       root,
		UseParity:      *parityDir != "",
		ParityDir:      resolveCLIPath(root, *parityDir),
		BaselineDir:    resolveCLIPath(root, *baselineDir),
		ArtifactDir:    resolveCLIPath(root, *artifactDir),
		IncludeWebdemo: *includeWebdemo,
		WebBaselineDir: resolveCLIPath(root, *webBaselineDir),
		WebArtifactDir: resolveCLIPath(root, *webArtifactDir),
		NameFilter:     *nameFilter,
		NamePrefix:     *namePrefix,
		PrintOnly:      *printOnly,
	}, nil
}

func resolveCLIPath(root, path string) string {
	cleaned := filepath.Clean(path)
	if path != "" && !filepath.IsAbs(cleaned) {
		return filepath.Join(root, cleaned)
	}
	return cleaned
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
