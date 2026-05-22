package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestGoIdentifiersUseStableFeatureNames(t *testing.T) {
	roadmapLabel := regexp.MustCompile(`(?i)phase[0-9]`)
	fset := token.NewFileSet()

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "third_party", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(n ast.Node) bool {
			ident, ok := n.(*ast.Ident)
			if !ok || !roadmapLabel.MatchString(ident.Name) {
				return true
			}
			t.Errorf("%s uses roadmap phase in Go identifier %s; use a stable feature name", fset.Position(ident.Pos()), ident.Name)
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("scan Go files: %v", err)
	}
}
