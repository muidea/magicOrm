package orm

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

// Both dialect files are inspected, independently of the current build tag.
// New capabilities require a deliberate update of the public contract here.
func TestPublicFacadeDoesNotExposeExecutionImplementations(t *testing.T) {
	functions := strings.Fields("GetOrm NewOrm BeginSerializableTransaction LockWithFilter Aggregate ValidateSchema PreflightSchemaChange SchemaTables SchemaSetTables SchemaTableName Initialize Uninitialized AddDatabase DelDatabase EnsureSchema NewConfig OwnerPoolStats EnsureORMMetricProviderRegistered")
	allowed := map[string]bool{}
	for _, name := range functions {
		allowed[name] = true
	}
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".go") || strings.HasSuffix(file.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), file.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		internalImports := map[string]bool{}
		for _, imp := range f.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if strings.HasPrefix(path, "github.com/muidea/magicOrm/internal/") {
				name := path[strings.LastIndex(path, "/")+1:]
				if imp.Name != nil {
					name = imp.Name.Name
				}
				internalImports[name] = true
			}
		}
		for _, decl := range f.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if !decl.Name.IsExported() {
					continue
				}
				if decl.Recv == nil {
					if !allowed[decl.Name.Name] {
						t.Errorf("unexpected exported factory/function: %s", decl.Name.Name)
					}
					seen[decl.Name.Name] = true
				}
				ast.Inspect(decl.Type, func(node ast.Node) bool {
					if sel, ok := node.(*ast.SelectorExpr); ok {
						if pkg, ok := sel.X.(*ast.Ident); ok && internalImports[pkg.Name] {
							t.Errorf("%s leaks internal type %s.%s", decl.Name.Name, pkg.Name, sel.Sel.Name)
						}
					}
					return true
				})
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					if values, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range values.Names {
							if name.IsExported() {
								t.Errorf("unexpected exported state: %s", name.Name)
							}
						}
					}
					if typ, ok := spec.(*ast.TypeSpec); ok && typ.Name.IsExported() {
						if typ.Name.Name != "Orm" && typ.Name.Name != "SchemaReport" {
							t.Errorf("unexpected exported implementation type: %s", typ.Name.Name)
						}
						if typ.Assign == token.NoPos {
							t.Errorf("%s must refer to the models contract", typ.Name.Name)
						}
						contract, ok := typ.Type.(*ast.SelectorExpr)
						if !ok {
							t.Errorf("%s must alias a model contract", typ.Name.Name)
							continue
						}
						pkg, ok := contract.X.(*ast.Ident)
						if !ok || pkg.Name != "models" || (contract.Sel.Name != "ORM" && contract.Sel.Name != "SchemaReport") {
							t.Errorf("unexpected public alias target: %s", typ.Name.Name)
						}
					}
				}
			}
		}
	}
	for name := range allowed {
		if !seen[name] {
			t.Errorf("public capability was removed: %s", name)
		}
	}
	allowedMethods := strings.Fields("Create Drop Reconcile InspectSchema InspectSchemaSet DropSchemaSet RepairSchema Insert Update UpdateWithFilter Delete Query Count BatchQuery BeginTransaction CommitTransaction RollbackTransaction Release BeginSerializableTransaction LockWithFilter Aggregate")
	methodSet := map[string]bool{}
	for _, name := range allowedMethods {
		methodSet[name] = true
	}
	typ := reflect.TypeOf((*impl)(nil))
	for i := 0; i < typ.NumMethod(); i++ {
		if !methodSet[typ.Method(i).Name] {
			t.Errorf("implementation exposes mutable/internal control: %s", typ.Method(i).Name)
		}
	}
}

type observedConfig struct {
	models.DatabaseConfig
	read bool
}

func (c *observedConfig) GetDsn() string {
	c.read = true
	return c.DatabaseConfig.GetDsn()
}

func TestNewOrmPreservesCallerConfigurationContract(t *testing.T) {
	config := &observedConfig{DatabaseConfig: NewConfig("127.0.0.1:1", "sample", "tenant", "user", "synthetic")}
	// NewOrm opens an idle database/sql handle; no query, ping or connection is
	// performed. The caller's DSN must not be reconstructed from other fields.
	handler, err := NewOrm(provider.NewLocalProvider("config-contract", nil), config, "")
	if err != nil {
		t.Fatal(err)
	}
	defer handler.Release()
	if !config.read {
		t.Fatal("caller-provided connection description was replaced")
	}
}

func TestPublicObservationAndConfigurationCompatibility(t *testing.T) {
	var _ Orm = (models.ORM)(nil)
	var _ func(models.Model, bool) (*SchemaReport, *cd.Error) = (&impl{}).InspectSchema
	config := NewConfig("db:5432", "sample", "tenant", "user", "secret")
	if config.Server() != "db:5432" || config.Database() != "sample" || config.Schema() != "tenant" || config.Username() != "user" || config.Password() != "secret" || config.GetDsn() != newConfig("db:5432", "sample", "tenant", "user", "secret").GetDsn() {
		t.Fatal("connection-description contract changed")
	}
	report := &SchemaReport{Matches: false, Tables: []string{"tenant_Host"}, Issues: []models.SchemaIssue{{Table: "tenant_Host", Object: "value", Kind: "column-missing"}}}
	encoded, err := json.Marshal(report)
	if err != nil || string(encoded) != `{"matches":false,"tables":["tenant_Host"],"issues":[{"table":"tenant_Host","object":"value","kind":"column-missing"}]}` {
		t.Fatal("diagnostic wire contract changed", string(encoded), err)
	}
	if _, err := NewOrm(nil, config, ""); err == nil {
		t.Fatal("nil provider accepted")
	}
}

func TestSchemaHostIdentityDoesNotExpandOwnedGraph(t *testing.T) {
	handler, model, _ := inspectionFixture(t)
	name, err := SchemaTableName(model, "tenant_a")
	if err != nil || name != "tenant_a_InspectHost" {
		t.Fatal(name, err)
	}
	names, err := SchemaTables(handler.context, model, handler.modelProvider, "tenant_a")
	if err != nil || len(names) != 4 {
		t.Fatal("owned graph observation changed", names, err)
	}
	if _, err := SchemaTableName(nil, "tenant_a"); err == nil {
		t.Fatal("nil host accepted")
	}
}
