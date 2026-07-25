package main_test

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStablePublicPackagesHaveGoDocAndExamples(t *testing.T) {
	for _, importPath := range stablePublicPackageDirs() {
		importPath := importPath
		t.Run(importPath, func(t *testing.T) {
			pkg, err := build.ImportDir(importPath, build.ImportComment)
			if err != nil {
				t.Fatalf("load package: %v", err)
			}
			if strings.TrimSpace(pkg.Doc) == "" {
				t.Fatalf("missing package documentation")
			}
			if !hasExternalExampleFile(t, importPath, pkg.Name) {
				t.Fatalf("missing external worked example in package %q", importPath)
			}
		})
	}
}

func stablePublicPackageDirs() []string {
	return []string{
		".",
		"animation",
		"backends",
		"backends/agg",
		"backends/all",
		"backends/desktop",
		"backends/desktop/gio",
		"backends/gobasic",
		"backends/pdf",
		"backends/pgf",
		"backends/ps",
		"backends/skia",
		"backends/svg",
		"backends/webagg",
		"canvas",
		"color",
		"core",
		"dates",
		"diag",
		"geom",
		"optional",
		"plot3d",
		"pyplot",
		"render",
		"style",
		"ticker",
		"transform",
		"tri",
		"widgets",
	}
}

func hasExternalExampleFile(t *testing.T, dir, packageName string) bool {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}

	examplePackage := packageName + "_test"
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(string(data), "package "+examplePackage) &&
			strings.Contains(string(data), "func Example") {
			return true
		}
	}
	return false
}
