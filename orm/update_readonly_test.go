package orm

import (
	"context"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestUpdateRunnerRestoresReadOnlyBasicFieldsFromStoredModel(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)

	object := &remote.Object{
		Name:    "Demo",
		PkgPath: "/vmi/test",
		Fields: []*remote.Field{
			{
				Name: "id",
				Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &remote.SpecImpl{PrimaryKey: true, ValueDeclare: models.AutoIncrement, ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}, Constraint: "ro"},
			},
			{
				Name: "name",
				Type: &remote.TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &remote.SpecImpl{ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}},
			},
			{
				Name: "createTime",
				Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &remote.SpecImpl{ViewDeclare: []models.ViewDeclare{models.DetailView}, Constraint: "ro"},
			},
		},
	}
	if _, err := remoteProvider.RegisterModel(object); err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}

	updateModel := &remote.ObjectValue{
		Name:    "Demo",
		PkgPath: "/vmi/test",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(101), Assigned: true},
			{Name: "name", Value: "updated", Assigned: true},
			{Name: "createTime", Value: int64(999999), Assigned: true},
		},
	}
	entityModel, err := remoteProvider.GetEntityModel(updateModel, true)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}

	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, `FROM "tenant_Demo"`) && len(args) == 1 && args[0] == int64(101)
				},
				rows: [][]any{
					{int64(101), "updated", int64(123456)},
				},
			},
		},
	}

	updateRunner := NewUpdateRunner(context.Background(), entityModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"))
	updatedModel, err := updateRunner.Update()
	if err != nil {
		t.Fatalf("UpdateRunner.Update failed: %v", err)
	}

	updatedValue := updatedModel.Interface(true).(*remote.ObjectValue)
	if got := updatedValue.GetFieldValue("createTime"); got != int64(123456) {
		t.Fatalf("readonly createTime should be restored from stored model, got %#v", got)
	}
	if got := updatedValue.GetFieldValue("name"); got != "updated" {
		t.Fatalf("updated name mismatch, got %#v", got)
	}
	if !containsSQLCall(executor.execCalls, "exec", `UPDATE "tenant_Demo" SET "name" = $1 WHERE "id" = $2`, []any{"updated", int64(101)}) {
		t.Fatalf("missing host update call: %#v", executor.execCalls)
	}
}

func TestUpdateRunnerReturnsStoredReadOnlyFieldsEvenWhenNotAssigned(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)

	object := &remote.Object{
		Name:    "StockDoc",
		PkgPath: "/vmi/test",
		Fields: []*remote.Field{
			{
				Name: "id",
				Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &remote.SpecImpl{PrimaryKey: true, ValueDeclare: models.AutoIncrement, ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}, Constraint: "ro"},
			},
			{
				Name: "name",
				Type: &remote.TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &remote.SpecImpl{ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}},
			},
			{
				Name: "sn",
				Type: &remote.TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &remote.SpecImpl{ViewDeclare: []models.ViewDeclare{models.DetailView}, Constraint: "ro"},
			},
		},
	}
	if _, err := remoteProvider.RegisterModel(object); err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}

	updateValue := &remote.ObjectValue{
		Name:    "StockDoc",
		PkgPath: "/vmi/test",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(301), Assigned: true},
			{Name: "name", Value: "updated", Assigned: true},
		},
	}
	entityModel, err := remoteProvider.GetEntityModel(updateValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}

	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, `FROM "tenant_StockDoc"`) && len(args) == 1 && args[0] == int64(301)
				},
				rows: [][]any{
					{int64(301), "updated", "SN-001"},
				},
			},
		},
	}

	updateRunner := NewUpdateRunner(context.Background(), entityModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"))
	updatedModel, err := updateRunner.Update()
	if err != nil {
		t.Fatalf("UpdateRunner.Update failed: %v", err)
	}

	updatedValue := updatedModel.Interface(true).(*remote.ObjectValue)
	if got := updatedValue.GetFieldValue("name"); got != "updated" {
		t.Fatalf("updated name mismatch, got %#v", got)
	}
	if got := updatedValue.GetFieldValue("sn"); got != "SN-001" {
		t.Fatalf("update response should include stored readonly sn, got %#v", got)
	}
}

func TestInsertRunnerProjectsWriteResponseByDetailView(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)

	object := &remote.Object{
		Name:    "Demo",
		PkgPath: "/vmi/test",
		Fields: []*remote.Field{
			{
				Name: "id",
				Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &remote.SpecImpl{PrimaryKey: true, ValueDeclare: models.AutoIncrement, ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}},
			},
			{
				Name: "name",
				Type: &remote.TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &remote.SpecImpl{ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}},
			},
			{
				Name: "createTime",
				Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &remote.SpecImpl{ViewDeclare: []models.ViewDeclare{models.DetailView}, Constraint: "ro"},
			},
			{
				Name: "namespace",
				Type: &remote.TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &remote.SpecImpl{Constraint: "ro"},
			},
		},
	}
	if _, err := remoteProvider.RegisterModel(object); err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}

	insertValue := &remote.ObjectValue{
		Name:    "Demo",
		PkgPath: "/vmi/test",
		Fields: []*remote.FieldValue{
			{Name: "name", Value: "created"},
			{Name: "createTime", Value: int64(123456), Assigned: true},
			{Name: "namespace", Value: "tenant-a", Assigned: true},
		},
	}
	entityModel, err := remoteProvider.GetEntityModel(insertValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}

	executor := &fakeExecutor{
		insertIDs: []any{int64(201)},
	}

	insertRunner := NewInsertRunner(context.Background(), entityModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"))
	insertedModel, err := insertRunner.Insert()
	if err != nil {
		t.Fatalf("InsertRunner.Insert failed: %v", err)
	}

	insertedValue := insertedModel.Interface(true).(*remote.ObjectValue)
	if got := insertedValue.GetFieldValue("id"); got != int64(201) {
		t.Fatalf("inserted id mismatch, got %#v", got)
	}
	if got := insertedValue.GetFieldValue("name"); got != "created" {
		t.Fatalf("inserted name mismatch, got %#v", got)
	}
	if got := insertedValue.GetFieldValue("createTime"); got != int64(123456) {
		t.Fatalf("detail createTime mismatch, got %#v", got)
	}
	if got := insertedValue.GetFieldValue("namespace"); got != nil {
		t.Fatalf("namespace without detail view should not be returned, got %#v", got)
	}
}
