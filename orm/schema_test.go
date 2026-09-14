package orm

import (
	"context"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type inspectOwned struct {
	ID    int64   `orm:"id key auto"`
	Value *string `orm:"value"`
}
type inspectReferenced struct {
	ID int64 `orm:"id key auto"`
}
type inspectHost struct {
	ID         int64                `orm:"id key auto"`
	Owned      inspectOwned         `orm:"owned"`
	References []*inspectReferenced `orm:"references"`
}
type schemaExecutor struct {
	fakeExecutor
	tables map[string]database.TableSchema
	calls  []string
	err    *cd.Error
	cancel context.CancelFunc
}

func (s *schemaExecutor) ReadTableSchema(_ context.Context, name string) (database.TableSchema, *cd.Error) {
	s.calls = append(s.calls, name)
	if s.cancel != nil {
		s.cancel()
	}
	if s.err != nil {
		return database.TableSchema{}, s.err
	}
	if table, ok := s.tables[name]; ok {
		return table, nil
	}
	return database.TableSchema{Name: name}, nil
}

func inspectionFixture(t *testing.T) (*impl, models.Model, *schemaExecutor) {
	t.Helper()
	p := provider.NewLocalProvider("schema-inspection", nil)
	for _, entity := range []any{&inspectOwned{}, &inspectReferenced{}, &inspectHost{}} {
		if _, err := p.RegisterModel(entity); err != nil {
			t.Fatal(err)
		}
	}
	model, err := p.GetEntityModel(&inspectHost{}, true)
	if err != nil {
		t.Fatal(err)
	}
	c := codec.New(p, "tenant_a")
	builder := NewBuilder(p, c).(database.SchemaBuilder)
	executor := &schemaExecutor{tables: map[string]database.TableSchema{}}
	host, err := builder.DescribeTable(model)
	if err != nil {
		t.Fatal(err)
	}
	executor.tables[host.Name] = host
	for _, field := range model.GetFields() {
		if models.IsBasicField(field) {
			continue
		}
		relation, err := builder.DescribeRelationTable(model, field)
		if err != nil {
			t.Fatal(err)
		}
		executor.tables[relation.Name] = relation
		if !field.GetType().Elem().IsPtrType() {
			owned, err := p.GetTypeModel(field.GetType().Elem())
			if err != nil {
				t.Fatal(err)
			}
			table, err := builder.DescribeTable(owned)
			if err != nil {
				t.Fatal(err)
			}
			executor.tables[table.Name] = table
		}
	}
	return &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: c}, model, executor
}

func TestInspectSchemaIncludesOwnedAndRelationsNotReferenceHosts(t *testing.T) {
	handler, model, executor := inspectionFixture(t)
	report, err := handler.InspectSchema(model, true)
	if err != nil || !report.Matches || len(report.Tables) != 4 || len(executor.execCalls) != 0 {
		t.Fatal(report, err, executor.execCalls)
	}
	for _, name := range report.Tables {
		if name == "tenant_a_InspectReferenced" {
			t.Fatal("inspected unowned reference host")
		}
	}
	for name, table := range executor.tables {
		if name == "tenant_a_InspectHost" {
			continue
		}
		delete(executor.tables, name)
		report, err = handler.InspectSchema(model, true)
		if err != nil || report.Matches || len(report.Issues) != 1 || report.Issues[0].Table != name {
			t.Fatal(report, err)
		}
		executor.tables[name] = table
	}
}

func TestInspectSchemaDropChecksEveryOwnedTable(t *testing.T) {
	handler, model, executor := inspectionFixture(t)
	report, err := handler.InspectSchema(model, false)
	if err != nil || len(report.Issues) != 4 {
		t.Fatal(report, err)
	}
	executor.tables = map[string]database.TableSchema{}
	report, err = handler.InspectSchema(model, false)
	if err != nil || !report.Matches || len(report.Tables) != 4 {
		t.Fatal(report, err)
	}
}

func TestInspectSchemaReadFailureNeverReturnsSuccessReceipt(t *testing.T) {
	for _, mode := range []string{"read-error", "canceled", "wrong-table", "unsupported", "nil-model"} {
		t.Run(mode, func(t *testing.T) {
			handler, model, executor := inspectionFixture(t)
			switch mode {
			case "read-error":
				executor.err = cd.NewError(cd.Unexpected, "catalog failed")
			case "canceled":
				var cancel context.CancelFunc
				handler.context, cancel = context.WithCancel(context.Background())
				defer cancel()
				executor.cancel = cancel
			case "wrong-table":
				for name, table := range executor.tables {
					table.Name = "foreign"
					executor.tables[name] = table
				}
			case "unsupported":
				handler.executor = &fakeExecutor{}
			case "nil-model":
				model = nil
			}
			report, err := handler.InspectSchema(model, true)
			if err == nil || report != nil {
				t.Fatal(report, err)
			}
		})
	}
}

type schemaMappedProvider struct {
	provider.Provider
	mapped models.Model
}

func (s schemaMappedProvider) GetTypeModel(models.Type) (models.Model, *cd.Error) {
	return s.mapped, nil
}

type schemaNamedModel struct {
	models.Model
	name string
}

func (s schemaNamedModel) GetName() string { return s.name }

func TestInspectSchemaRejectsOwnedCycleAndPhysicalNameCollisionBeforeReading(t *testing.T) {
	for _, mode := range []string{"cycle", "collision"} {
		t.Run(mode, func(t *testing.T) {
			handler, model, executor := inspectionFixture(t)
			mapped := model
			if mode == "collision" {
				owned, err := handler.modelProvider.GetTypeModel(model.GetField("owned").GetType().Elem())
				if err != nil {
					t.Fatal(err)
				}
				mapped = schemaNamedModel{Model: owned, name: model.GetName()}
			}
			handler.modelProvider = schemaMappedProvider{Provider: handler.modelProvider, mapped: mapped}
			report, err := handler.InspectSchema(model, true)
			if err == nil || report != nil || len(executor.calls) != 0 || len(executor.execCalls) != 0 {
				t.Fatal(report, err, executor.calls)
			}
		})
	}
}
