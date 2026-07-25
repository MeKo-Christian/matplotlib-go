package main_test

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type stableAPIArtifact struct {
	SchemaVersion int                `json:"schema_version"`
	Packages      []stableAPIPackage `json:"packages"`
}

type stableAPIPackage struct {
	ImportPath string            `json:"import_path"`
	Dir        string            `json:"dir"`
	Symbols    []stableAPISymbol `json:"symbols"`
}

type stableAPISymbol struct {
	ID          string `json:"id"`
	Declaration string `json:"declaration"`
}

func TestStablePublicAPIMatchesFrozenAudit(t *testing.T) {
	root := repoRootForAPIAudit(t)
	got := collectStablePublicAPI(t, root)
	gotJSON := marshalStableAPIArtifact(t, got)

	path := filepath.Join(root, "test", "testdata", "public_api", "stable_public_api.json")
	if os.Getenv("UPDATE_PUBLIC_API_AUDIT") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create public API artifact dir: %v", err)
		}
		if err := os.WriteFile(path, gotJSON, 0o644); err != nil {
			t.Fatalf("write frozen public API artifact %s: %v", path, err)
		}
		t.Logf("updated frozen public API artifact %s", path)
		return
	}

	wantJSON, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read frozen public API artifact %s: %v", path, err)
	}
	if !bytes.Equal(bytes.TrimSpace(gotJSON), bytes.TrimSpace(wantJSON)) {
		t.Fatalf("stable public API differs from frozen audit; inspect exported changes and update %s only for intentional v1 API changes", path)
	}
}

func TestStablePublicAPIDoesNotExposeInternalGeometry(t *testing.T) {
	root := repoRootForAPIAudit(t)
	for _, dir := range stablePublicPackageDirs() {
		pkgDir := filepath.Join(root, filepath.FromSlash(dir))
		_, err := build.ImportDir(pkgDir, build.FindOnly)
		if err != nil {
			t.Fatalf("load package %s: %v", dir, err)
		}

		fset := token.NewFileSet()
		for _, name := range allNonTestGoFiles(t, pkgDir) {
			file, err := parser.ParseFile(fset, filepath.Join(pkgDir, name), nil, parser.ParseComments)
			if err != nil {
				t.Fatalf("parse %s/%s: %v", dir, name, err)
			}
			internalAliases := internalGeometryImportAliases(file)
			if len(internalAliases) == 0 {
				continue
			}
			for _, symbol := range exportedDeclsForFile(fset, file) {
				for alias := range internalAliases {
					if strings.Contains(symbol.Declaration, alias+".") {
						t.Fatalf("%s exports %s with unimportable internal geometry type in %q", dir, symbol.ID, symbol.Declaration)
					}
				}
			}
		}
	}
}

func collectStablePublicAPI(t *testing.T, root string) stableAPIArtifact {
	t.Helper()

	artifact := stableAPIArtifact{
		SchemaVersion: 1,
		Packages:      make([]stableAPIPackage, 0, len(stablePublicPackageDirs())),
	}
	for _, dir := range stablePublicPackageDirs() {
		pkg := collectStablePackageAPI(t, root, dir)
		artifact.Packages = append(artifact.Packages, pkg)
	}
	return artifact
}

func collectStablePackageAPI(t *testing.T, root, dir string) stableAPIPackage {
	t.Helper()

	pkgDir := filepath.Join(root, filepath.FromSlash(dir))
	buildPkg, err := build.ImportDir(pkgDir, build.FindOnly)
	if err != nil {
		t.Fatalf("load package %s: %v", dir, err)
	}

	fset := token.NewFileSet()
	symbolsByID := make(map[string]stableAPISymbol)
	for _, name := range allNonTestGoFiles(t, pkgDir) {
		file, err := parser.ParseFile(fset, filepath.Join(pkgDir, name), nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s/%s: %v", dir, name, err)
		}
		for _, symbol := range exportedDeclsForFile(fset, file) {
			if previous, ok := symbolsByID[symbol.ID]; ok {
				if stableDeclarationSignature(previous.Declaration) != stableDeclarationSignature(symbol.Declaration) {
					if merged, ok := mergeEmptyBuildStub(previous, symbol); ok {
						symbolsByID[symbol.ID] = merged
						continue
					}
					t.Fatalf(
						"%s has build-variant declarations for %s: %q and %q",
						dir,
						symbol.ID,
						previous.Declaration,
						symbol.Declaration,
					)
				}
				continue
			}
			symbolsByID[symbol.ID] = symbol
		}
	}
	symbols := make([]stableAPISymbol, 0, len(symbolsByID))
	for _, symbol := range symbolsByID {
		symbols = append(symbols, symbol)
	}
	sort.Slice(symbols, func(i, j int) bool { return symbols[i].ID < symbols[j].ID })

	return stableAPIPackage{
		ImportPath: buildPkg.ImportPath,
		Dir:        filepath.ToSlash(dir),
		Symbols:    symbols,
	}
}

func allNonTestGoFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read package directory %s: %v", dir, err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)
	return files
}

func stableDeclarationSignature(declaration string) string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "declaration.go", "package audit\n"+declaration, 0)
	if err != nil || len(file.Decls) != 1 {
		return declaration
	}
	fn, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		return declaration
	}
	return strings.Join([]string{
		"func",
		fieldListTypeSignature(fset, fn.Recv),
		fn.Name.Name,
		fieldListTypeSignature(fset, fn.Type.TypeParams),
		fieldListTypeSignature(fset, fn.Type.Params),
		fieldListTypeSignature(fset, fn.Type.Results),
	}, "\x00")
}

func fieldListTypeSignature(fset *token.FileSet, fields *ast.FieldList) string {
	if fields == nil {
		return ""
	}
	types := make([]string, 0, len(fields.List))
	for _, field := range fields.List {
		var out bytes.Buffer
		if err := printer.Fprint(&out, fset, field.Type); err != nil {
			return ""
		}
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			types = append(types, out.String())
		}
	}
	return strings.Join(types, ",")
}

func mergeEmptyBuildStub(left, right stableAPISymbol) (stableAPISymbol, bool) {
	leftEmpty := strings.HasSuffix(left.Declaration, " struct{}")
	rightEmpty := strings.HasSuffix(right.Declaration, " struct{}")
	switch {
	case leftEmpty && !rightEmpty:
		return right, true
	case rightEmpty && !leftEmpty:
		return left, true
	default:
		return stableAPISymbol{}, false
	}
}

func exportedDeclsForFile(fset *token.FileSet, file *ast.File) []stableAPISymbol {
	var symbols []stableAPISymbol
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			symbols = append(symbols, exportedGenDeclSymbols(fset, decl)...)
		case *ast.FuncDecl:
			symbol, ok := exportedFuncDeclSymbol(fset, decl)
			if ok {
				symbols = append(symbols, symbol)
			}
		}
	}
	return symbols
}

func exportedGenDeclSymbols(fset *token.FileSet, decl *ast.GenDecl) []stableAPISymbol {
	var symbols []stableAPISymbol
	for _, spec := range decl.Specs {
		switch spec := spec.(type) {
		case *ast.TypeSpec:
			if !spec.Name.IsExported() {
				continue
			}
			declaration := "type " + spec.Name.Name
			if spec.Assign.IsValid() {
				declaration += " ="
			}
			declaration += " " + stableTypeDeclString(fset, spec.Type)
			symbols = append(symbols, stableAPISymbol{
				ID:          "type " + spec.Name.Name,
				Declaration: declaration,
			})
		case *ast.ValueSpec:
			for i, name := range spec.Names {
				if !name.IsExported() {
					continue
				}
				declaration := decl.Tok.String() + " " + name.Name
				if spec.Type != nil {
					declaration += " " + nodeString(fset, spec.Type)
				}
				if i < len(spec.Values) {
					declaration += " = " + nodeString(fset, spec.Values[i])
				}
				symbols = append(symbols, stableAPISymbol{
					ID:          decl.Tok.String() + " " + name.Name,
					Declaration: declaration,
				})
			}
		}
	}
	return symbols
}

func exportedFuncDeclSymbol(fset *token.FileSet, decl *ast.FuncDecl) (stableAPISymbol, bool) {
	if !decl.Name.IsExported() {
		return stableAPISymbol{}, false
	}
	if decl.Recv == nil {
		return stableAPISymbol{
			ID:          "func " + decl.Name.Name,
			Declaration: "func " + decl.Name.Name + funcTypeString(fset, decl.Type),
		}, true
	}
	recv := receiverBaseName(decl.Recv)
	if recv == "" || !ast.IsExported(recv) {
		return stableAPISymbol{}, false
	}
	return stableAPISymbol{
		ID:          "method " + recv + "." + decl.Name.Name,
		Declaration: "func (" + recv + ") " + decl.Name.Name + funcTypeString(fset, decl.Type),
	}, true
}

func stableTypeDeclString(fset *token.FileSet, expr ast.Expr) string {
	switch expr := expr.(type) {
	case *ast.StructType:
		var fields []string
		if expr.Fields != nil {
			for _, field := range expr.Fields.List {
				fields = append(fields, exportedFieldStrings(fset, field)...)
			}
		}
		if len(fields) == 0 {
			return "struct{}"
		}
		return "struct{ " + strings.Join(fields, "; ") + " }"
	case *ast.InterfaceType:
		var methods []string
		if expr.Methods != nil {
			for _, method := range expr.Methods.List {
				methods = append(methods, exportedFieldStrings(fset, method)...)
			}
		}
		if len(methods) == 0 {
			return "interface{}"
		}
		return "interface{ " + strings.Join(methods, "; ") + " }"
	default:
		return nodeString(fset, expr)
	}
}

func exportedFieldStrings(fset *token.FileSet, field *ast.Field) []string {
	typeString := nodeString(fset, field.Type)
	if len(field.Names) == 0 {
		if exportedEmbeddedField(field.Type) {
			return []string{typeString}
		}
		return nil
	}
	var out []string
	for _, name := range field.Names {
		if name.IsExported() {
			out = append(out, name.Name+" "+typeString)
		}
	}
	return out
}

func exportedEmbeddedField(expr ast.Expr) bool {
	switch expr := expr.(type) {
	case *ast.Ident:
		return expr.IsExported()
	case *ast.StarExpr:
		return exportedEmbeddedField(expr.X)
	case *ast.SelectorExpr:
		return expr.Sel.IsExported()
	default:
		return false
	}
}

func internalGeometryImportAliases(file *ast.File) map[string]bool {
	aliases := map[string]bool{}
	for _, imp := range file.Imports {
		if strings.Trim(imp.Path.Value, `"`) != "github.com/cwbudde/matplotlib-go/internal/geom" {
			continue
		}
		alias := "geom"
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		if alias != "." && alias != "_" {
			aliases[alias] = true
		}
	}
	return aliases
}

func receiverBaseName(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	expr := fields.List[0].Type
	for {
		switch typed := expr.(type) {
		case *ast.StarExpr:
			expr = typed.X
		case *ast.Ident:
			return typed.Name
		default:
			return ""
		}
	}
}

func funcTypeString(fset *token.FileSet, typ *ast.FuncType) string {
	text := nodeString(fset, typ)
	return strings.TrimPrefix(text, "func")
}

func nodeString(fset *token.FileSet, node any) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		panic(err)
	}
	return strings.Join(strings.Fields(buf.String()), " ")
}

func marshalStableAPIArtifact(t *testing.T, artifact stableAPIArtifact) []byte {
	t.Helper()
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		t.Fatalf("marshal stable API artifact: %v", err)
	}
	return append(data, '\n')
}

func repoRootForAPIAudit(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	return wd
}
