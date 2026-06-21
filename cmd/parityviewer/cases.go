package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	RefImageURL string
	ActImageURL string
	RawDiffURL  string
	AmpDiffURL  string
}

type loadResult struct {
	Cases         []caseEntry
	ComparedCount int
	SkippedCount  int
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

func imageURL(suite, baseline, name, kind string) string {
	values := url.Values{}
	values.Set("suite", suite)
	values.Set("baseline", baseline)
	values.Set("name", name)
	values.Set("kind", kind)
	return "/image?" + values.Encode()
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
