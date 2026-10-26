// Package apicheck guards the source contract without freezing private implementation.
package apicheck

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var updateAPI = flag.Bool("update-api", false, "update the reviewed public API snapshot")

// TestPublicAPI covers both dialect source sets, including support packages.
// Changes to this snapshot require reviewing callers and compatibility first.
func TestPublicAPI(t *testing.T) {
	got := publicAPI(t, filepath.Join("..", ".."))
	data, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	path := filepath.Join("testdata", "public-api.json")
	if *updateAPI {
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotJSON(t, want), gotJSON(t, data)) {
		t.Fatal("public API changed; review compatibility before updating with go test ./internal/apicheck -update-api")
	}
}

func gotJSON(t *testing.T, data []byte) []byte {
	t.Helper()
	var value map[string]string
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func publicAPI(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "vendor", "internal", "test", "examples", "example":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		if file.Name.Name == "main" {
			return nil
		}
		pkg, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		imports := map[string]string{}
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			name := filepath.Base(importPath)
			if spec.Name != nil {
				name = spec.Name.Name
			} else if strings.HasPrefix(importPath, "github.com/muidea/magicOrm/internal/") {
				// Package names may differ from directory names (e.g. shared).
				dir := filepath.Join(root, strings.TrimPrefix(importPath, "github.com/muidea/magicOrm/"))
				entries, err := os.ReadDir(dir)
				if err != nil {
					return err
				}
				for _, entry := range entries {
					if strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
						declaration, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, parser.PackageClauseOnly)
						if err != nil {
							return err
						}
						name = declaration.Name.Name
						break
					}
				}
			}
			imports[name] = importPath
		}
		add := func(name string, node ast.Node) {
			var buf bytes.Buffer
			if err := format.Node(&buf, fset, node); err != nil {
				t.Fatal(err)
			}
			key, signature := filepath.ToSlash(pkg)+" "+name, buf.String()
			used := map[string]string{}
			ast.Inspect(node, func(n ast.Node) bool {
				if selector, ok := n.(*ast.SelectorExpr); ok {
					if ident, ok := selector.X.(*ast.Ident); ok {
						if imported, ok := imports[ident.Name]; ok {
							if strings.HasPrefix(imported, "github.com/muidea/magicOrm/internal/") {
								t.Errorf("public signature %s exposes internal type from %s", key, imported)
							}
							used[ident.Name] = imported
						}
					}
				}
				return true
			})
			// Resolve imported names as well as syntax, so switching a signature
			// to another package under the same alias cannot evade the snapshot.
			var names []string
			for name := range used {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				signature += "\nimport " + name + " " + strconv.Quote(used[name])
			}
			if previous, exists := result[key]; exists && previous != signature {
				t.Fatalf("dialects disagree on %s: %s / %s", key, previous, signature)
			}
			result[key] = signature
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				name := "func " + d.Name.Name
				if d.Recv != nil {
					receiver := d.Recv.List[0].Type
					pointer := ""
					if ptr, ok := receiver.(*ast.StarExpr); ok {
						pointer = "*"
						receiver = ptr.X
					}
					ident, ok := receiver.(*ast.Ident)
					if !ok || !ident.IsExported() {
						continue
					}
					name = "method " + pointer + ident.Name + "." + d.Name.Name
				} else if strings.Contains(d.Name.Name, "ForTest") {
					t.Errorf("test-only entry leaked into %s: %s", pkg, d.Name.Name)
				}
				add(name, d.Type)
			case *ast.GenDecl:
				var previousValues []ast.Expr
				var previousType ast.Expr
				for index, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if !s.Name.IsExported() {
							continue
						}
						if typ, ok := s.Type.(*ast.StructType); ok {
							fields := &ast.FieldList{}
							for _, field := range typ.Fields.List {
								if len(field.Names) == 0 || field.Names[0].IsExported() {
									fields.List = append(fields.List, field)
								}
							}
							s.Type = &ast.StructType{Fields: fields}
						}
						add("type "+s.Name.Name, s)
					case *ast.ValueSpec:
						if len(s.Values) != 0 {
							previousValues, previousType = s.Values, s.Type
						}
						for _, name := range s.Names {
							if !name.IsExported() {
								continue
							}
							copy := *s
							if d.Tok == token.CONST && len(copy.Values) == 0 {
								copy.Values, copy.Type = previousValues, previousType
							}
							key := d.Tok.String() + " " + name.Name
							// Preserve iota positions, which affect constant values.
							if d.Tok == token.CONST {
								key += " [" + strconv.Itoa(index) + "]"
							}
							add(key, &copy)
						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(fmt.Errorf("inspect public packages: %w", err))
	}
	return result
}
