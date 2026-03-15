package orm

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/database/postgres"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

type queryMaskViewModel struct {
	ID          int64  `orm:"id key auto" view:"detail,lite"`
	Name        string `orm:"name" view:"detail,lite"`
	Description string `orm:"description" view:"detail"`
}

type queryMaskReference struct {
	ID          int       `orm:"id key auto"`
	Name        string    `orm:"name"`
	FValue      float32   `orm:"value"`
	Flag        bool      `orm:"flag"`
	StrArray    []string  `orm:"strArray"`
	PtrArray    *[]string `orm:"ptrArray"`
	PtrStrArray *[]string `orm:"ptrStrArray"`
}

type implicitQuerySliceModel struct {
	ID    int64    `orm:"id key auto"`
	Token string   `orm:"token"`
	Tags  []string `orm:"tags"`
}

func TestBuildFullQueryMaskModelExpandsRemoteBasicFields(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)

	definition, err := helper.GetObject(&queryMaskReference{})
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if _, err := remoteProvider.RegisterModel(definition); err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}

	queryValue := &remote.ObjectValue{
		Name:    "queryMaskReference",
		PkgPath: "github.com/muidea/magicOrm/orm",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: 1, Assigned: true},
		},
	}

	queryModel, err := remoteProvider.GetEntityModel(queryValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}

	maskModel, err := buildFullQueryMaskModel(queryModel)
	if err != nil {
		t.Fatalf("buildFullQueryMaskModel failed: %v", err)
	}

	for _, fieldName := range []string{"name", "value", "flag", "strArray"} {
		field := maskModel.GetField(fieldName)
		if field == nil {
			t.Fatalf("mask field %q should exist", fieldName)
		}
		if !models.IsValidField(field) {
			t.Fatalf("mask field %q should be valid for full-row query", fieldName)
		}
	}

	for _, fieldName := range []string{"ptrArray", "ptrStrArray"} {
		field := maskModel.GetField(fieldName)
		if field == nil {
			t.Fatalf("mask field %q should exist", fieldName)
		}
		if models.IsValidField(field) {
			t.Fatalf("pointer field %q should keep query mask semantics", fieldName)
		}
	}
}

func registerQueryMaskViewRemoteModel(t *testing.T, remoteProvider provider.Provider) {
	t.Helper()

	definition, err := helper.GetObject(&queryMaskViewModel{})
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if _, err := remoteProvider.RegisterModel(definition); err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}
}

func TestQueryRunnerUsesLiteViewResponseMask(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerQueryMaskViewRemoteModel(t, remoteProvider)

	queryValue := &remote.ObjectValue{
		Name:    "queryMaskViewModel",
		PkgPath: "github.com/muidea/magicOrm/orm",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1), Assigned: true},
		},
	}
	filter, err := remoteProvider.GetEntityFilter(queryValue, models.LiteView)
	if err != nil {
		t.Fatalf("GetEntityFilter failed: %v", err)
	}
	queryModel, err := remoteProvider.GetEntityModel(queryValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}
	modelCodec := codec.New(remoteProvider, "tenant")
	queryFilter, err := getModelFilter(queryModel, remoteProvider, modelCodec)
	if err != nil {
		t.Fatalf("getModelFilter failed: %v", err)
	}

	responseModel, responseByMask, err := buildQueryResponseModel(nil, filter)
	if err != nil {
		t.Fatalf("buildQueryResponseModel failed: %v", err)
	}
	queryMask, err := buildFullQueryMaskModel(responseModel)
	if err != nil {
		t.Fatalf("buildFullQueryMaskModel failed: %v", err)
	}

	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_QueryMaskViewModel") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(1)})
				},
				rows: [][]any{{int64(1), "alpha", "detail-description"}},
			},
		},
	}

	queryRunner := NewQueryRunner(context.Background(), queryMask, responseModel, responseByMask, executor, remoteProvider, modelCodec, false, 0)
	modelList, err := queryRunner.Query(queryFilter)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(modelList) != 1 {
		t.Fatalf("Query got %d rows", len(modelList))
	}

	queryResult := modelList[0].Interface(true).(*remote.ObjectValue)
	if queryResult.GetFieldValue("id") != int64(1) {
		t.Fatalf("id mismatch: %#v", queryResult.GetFieldValue("id"))
	}
	if queryResult.GetFieldValue("name") != "alpha" {
		t.Fatalf("name mismatch: %#v", queryResult.GetFieldValue("name"))
	}
	if queryResult.GetFieldValue("description") != nil {
		t.Fatalf("LiteView should exclude description, got %#v", queryResult.GetFieldValue("description"))
	}
}

func TestQueryRunnerValueMaskOverridesLiteViewResponseMask(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerQueryMaskViewRemoteModel(t, remoteProvider)

	queryValue := &remote.ObjectValue{
		Name:    "queryMaskViewModel",
		PkgPath: "github.com/muidea/magicOrm/orm",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1), Assigned: true},
		},
	}
	filter, err := remoteProvider.GetEntityFilter(queryValue, models.LiteView)
	if err != nil {
		t.Fatalf("GetEntityFilter failed: %v", err)
	}
	queryModel, err := remoteProvider.GetEntityModel(queryValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}
	modelCodec := codec.New(remoteProvider, "tenant")
	queryFilter, err := getModelFilter(queryModel, remoteProvider, modelCodec)
	if err != nil {
		t.Fatalf("getModelFilter failed: %v", err)
	}
	if err := filter.ValueMask(&remote.ObjectValue{
		Name:    "queryMaskViewModel",
		PkgPath: "github.com/muidea/magicOrm/orm",
		Fields: []*remote.FieldValue{
			{Name: "description", Value: "", Assigned: true},
		},
	}); err != nil {
		t.Fatalf("ValueMask failed: %v", err)
	}

	responseModel, responseByMask, err := buildQueryResponseModel(nil, filter)
	if err != nil {
		t.Fatalf("buildQueryResponseModel failed: %v", err)
	}
	queryMask, err := buildFullQueryMaskModel(responseModel)
	if err != nil {
		t.Fatalf("buildFullQueryMaskModel failed: %v", err)
	}

	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_QueryMaskViewModel") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(1)})
				},
				rows: [][]any{{int64(1), "alpha", "detail-description"}},
			},
		},
	}

	queryRunner := NewQueryRunner(context.Background(), queryMask, responseModel, responseByMask, executor, remoteProvider, modelCodec, false, 0)
	modelList, err := queryRunner.Query(queryFilter)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	queryResult := modelList[0].Interface(true).(*remote.ObjectValue)

	if queryResult.GetFieldValue("id") != int64(1) {
		t.Fatalf("id mismatch: %#v", queryResult.GetFieldValue("id"))
	}
	if queryResult.GetFieldValue("name") != nil {
		t.Fatalf("ValueMask should override LiteView and exclude name, got %#v", queryResult.GetFieldValue("name"))
	}
	if queryResult.GetFieldValue("description") != "detail-description" {
		t.Fatalf("description mismatch: %#v", queryResult.GetFieldValue("description"))
	}
}

func TestGetModelFilterSkipsImplicitLocalSliceConditions(t *testing.T) {
	localProvider := provider.NewLocalProvider("tenant", nil)
	if _, err := localProvider.RegisterModel(&implicitQuerySliceModel{}); err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}

	modelType, err := localProvider.GetEntityType(&implicitQuerySliceModel{})
	if err != nil {
		t.Fatalf("GetEntityType failed: %v", err)
	}
	queryModel, err := localProvider.GetTypeModel(modelType)
	if err != nil {
		t.Fatalf("GetTypeModel failed: %v", err)
	}
	if err := queryModel.SetFieldValue("token", "file-token"); err != nil {
		t.Fatalf("SetFieldValue(token) failed: %v", err)
	}
	if err := queryModel.SetFieldValue("tags", []string{}); err != nil {
		t.Fatalf("SetFieldValue(tags) failed: %v", err)
	}

	modelCodec := codec.New(localProvider, "tenant")
	queryFilter, err := getModelFilter(queryModel, localProvider, modelCodec)
	if err != nil {
		t.Fatalf("getModelFilter failed: %v", err)
	}

	builder := postgres.NewBuilder(localProvider, modelCodec)
	querySQL, err := builder.BuildQuery(queryModel, queryFilter)
	if err != nil {
		t.Fatalf("BuildQuery failed: %v", err)
	}

	whereClause := querySQL.SQL()
	if idx := strings.Index(whereClause, " WHERE "); idx >= 0 {
		whereClause = whereClause[idx+7:]
	}
	if strings.Contains(whereClause, `"tags"`) {
		t.Fatalf("implicit slice field should not appear in WHERE, sql=%s", querySQL.SQL())
	}
	if !strings.Contains(querySQL.SQL(), `"token" = $1`) {
		t.Fatalf("token condition missing, sql=%s", querySQL.SQL())
	}
	if !reflect.DeepEqual(querySQL.Args(), []any{"file-token"}) {
		t.Fatalf("unexpected args: %#v", querySQL.Args())
	}
}

func TestGetModelFilterSkipsImplicitRemoteSliceConditions(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	definition, err := helper.GetObject(&implicitQuerySliceModel{})
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if _, err := remoteProvider.RegisterModel(definition); err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}

	queryValue := &remote.ObjectValue{
		Name:    "implicitQuerySliceModel",
		PkgPath: "github.com/muidea/magicOrm/orm",
		Fields: []*remote.FieldValue{
			{Name: "token", Value: "file-token", Assigned: true},
			{Name: "tags", Value: []string{}, Assigned: true},
		},
	}
	queryModel, err := remoteProvider.GetEntityModel(queryValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}

	modelCodec := codec.New(remoteProvider, "tenant")
	queryFilter, err := getModelFilter(queryModel, remoteProvider, modelCodec)
	if err != nil {
		t.Fatalf("getModelFilter failed: %v", err)
	}

	builder := postgres.NewBuilder(remoteProvider, modelCodec)
	querySQL, err := builder.BuildQuery(queryModel, queryFilter)
	if err != nil {
		t.Fatalf("BuildQuery failed: %v", err)
	}

	whereClause := querySQL.SQL()
	if idx := strings.Index(whereClause, " WHERE "); idx >= 0 {
		whereClause = whereClause[idx+7:]
	}
	if strings.Contains(whereClause, `"tags"`) {
		t.Fatalf("implicit slice field should not appear in WHERE, sql=%s", querySQL.SQL())
	}
	if !strings.Contains(querySQL.SQL(), `"token" = $1`) {
		t.Fatalf("token condition missing, sql=%s", querySQL.SQL())
	}
	if !reflect.DeepEqual(querySQL.Args(), []any{"file-token"}) {
		t.Fatalf("unexpected args: %#v", querySQL.Args())
	}
}
