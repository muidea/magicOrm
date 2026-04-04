package orm

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unsafe"

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

type queryMaskRelationChild struct {
	ID     int64  `orm:"id key auto" view:"detail,lite"`
	Name   string `orm:"name" view:"detail,lite"`
	Secret string `orm:"secret" view:"detail"`
}

type queryMaskRelationParent struct {
	ID        int64                    `orm:"id key auto" view:"detail,lite"`
	Name      string                   `orm:"name" view:"detail,lite"`
	Child     *queryMaskRelationChild  `orm:"child" view:"detail,lite"`
	ChildList []queryMaskRelationChild `orm:"childList" view:"detail,lite"`
}

type countingValidator struct {
	calls int
}

func (s *countingValidator) Register(models.Key, models.ValidatorFunc) {}

func (s *countingValidator) ValidateValue(val any, directives []models.Directive) error {
	s.calls++
	for _, directive := range directives {
		if directive.Key() != models.KeyRequired {
			continue
		}
		if val == nil {
			return fmt.Errorf("required")
		}
		if strVal, ok := val.(string); ok && strVal == "" {
			return fmt.Errorf("required")
		}
	}
	return nil
}

func setRemoteObjectValidator(object *remote.Object, validator models.ValueValidator) {
	fieldVal := reflect.ValueOf(object).Elem().FieldByName("valueValidator")
	reflect.NewAt(fieldVal.Type(), unsafe.Pointer(fieldVal.UnsafeAddr())).Elem().Set(reflect.ValueOf(validator))
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

func TestBuildQueryExecutionModelKeepsExplicitMaskNarrow(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerQueryMaskViewRemoteModel(t, remoteProvider)

	filter, err := remoteProvider.GetEntityFilter(&remote.ObjectValue{
		Name:    "queryMaskViewModel",
		PkgPath: "github.com/muidea/magicOrm/orm",
	}, models.LiteView)
	if err != nil {
		t.Fatalf("GetEntityFilter failed: %v", err)
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
	if !responseByMask {
		t.Fatal("explicit value mask should mark responseByMask")
	}

	queryMask, err := buildQueryExecutionModel(responseModel, !responseByMask)
	if err != nil {
		t.Fatalf("buildQueryExecutionModel failed: %v", err)
	}

	nameField := queryMask.GetField("name")
	if nameField == nil {
		t.Fatal("name field should exist")
	}
	if models.IsValidField(nameField) || models.IsAssignedField(nameField) {
		t.Fatalf("explicit mask should not expand unrequested name field, got valid=%v assigned=%v", models.IsValidField(nameField), models.IsAssignedField(nameField))
	}

	descriptionField := queryMask.GetField("description")
	if descriptionField == nil || !models.IsValidField(descriptionField) {
		t.Fatalf("description field should stay queryable, got %#v", descriptionField)
	}
}

func TestRelationResponseModelUsesLiteViewForLocalMaskSlice(t *testing.T) {
	localProvider := provider.NewLocalProvider("tenant", nil)

	for _, entity := range []any{&queryMaskRelationChild{}, &queryMaskRelationParent{}} {
		if _, err := localProvider.RegisterModel(entity); err != nil {
			t.Fatalf("RegisterModel(%T) failed: %v", entity, err)
		}
	}

	baseModel, err := localProvider.GetEntityModel(&queryMaskRelationParent{}, true)
	if err != nil {
		t.Fatalf("GetEntityModel(base) failed: %v", err)
	}
	responseModel, err := localProvider.GetEntityModel(&queryMaskRelationParent{
		ChildList: []queryMaskRelationChild{},
	}, true)
	if err != nil {
		t.Fatalf("GetEntityModel(mask) failed: %v", err)
	}

	queryRunner := &QueryRunner{
		baseRunner: baseRunner{
			modelProvider: localProvider,
		},
		responseModel:  responseModel,
		responseByMask: true,
		relationCache:  map[string]models.Model{},
		relationMisses: map[string]struct{}{},
		relationEdges:  map[string][]any{},
		relationWarns:  map[string]struct{}{},
	}

	relationModel, relationByMask, err := queryRunner.relationResponseModel(baseModel.GetField("childList"))
	if err != nil {
		t.Fatalf("relationResponseModel(childList) failed: %v", err)
	}
	if relationByMask {
		t.Fatal("child relation should not inherit explicit nested response mask")
	}
	if relationModel == nil {
		t.Fatal("relationResponseModel(childList) returned nil")
	}
	if !fieldIncludedInResponse(relationModel, relationModel.GetField("name"), false) {
		t.Fatal("child lite response should include lite field name")
	}
	if fieldIncludedInResponse(relationModel, relationModel.GetField("secret"), false) {
		t.Fatal("child relation should be constrained to lite view")
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
	queryMask, err := buildQueryExecutionModel(responseModel, !responseByMask)
	if err != nil {
		t.Fatalf("buildQueryExecutionModel failed: %v", err)
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

func TestCanSkipProjectResponse(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerQueryMaskViewRemoteModel(t, remoteProvider)

	filter, err := remoteProvider.GetEntityFilter(&remote.ObjectValue{
		Name:    "queryMaskViewModel",
		PkgPath: "github.com/muidea/magicOrm/orm",
	}, models.LiteView)
	if err != nil {
		t.Fatalf("GetEntityFilter failed: %v", err)
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
	if !responseByMask {
		t.Fatal("explicit value mask should enable responseByMask")
	}
	queryMask, err := buildQueryExecutionModel(responseModel, !responseByMask)
	if err != nil {
		t.Fatalf("buildQueryExecutionModel failed: %v", err)
	}

	if !canSkipProjectResponse(queryMask, responseModel, responseByMask) {
		t.Fatal("expected same response mask shape to skip projection")
	}

	fullMask, err := buildFullQueryMaskModel(responseModel)
	if err != nil {
		t.Fatalf("buildFullQueryMaskModel failed: %v", err)
	}
	if canSkipProjectResponse(fullMask, responseModel, true) {
		t.Fatal("expanded query mask should still require projection")
	}
}

func TestORMBatchQueryUsesLiteViewResponseMask(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerQueryMaskViewRemoteModel(t, remoteProvider)

	filter, err := remoteProvider.GetEntityFilter(&remote.ObjectValue{
		Name:    "queryMaskViewModel",
		PkgPath: "github.com/muidea/magicOrm/orm",
	}, models.LiteView)
	if err != nil {
		t.Fatalf("GetEntityFilter failed: %v", err)
	}
	if err := filter.Equal("id", int64(1)); err != nil {
		t.Fatalf("filter.Equal(id) failed: %v", err)
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

	queryImpl := &impl{
		context:       context.Background(),
		executor:      executor,
		modelProvider: remoteProvider,
		modelCodec:    codec.New(remoteProvider, "tenant"),
	}
	queryResultList, err := queryImpl.BatchQuery(filter)
	if err != nil {
		t.Fatalf("BatchQuery failed: %v", err)
	}
	if len(queryResultList) != 1 {
		t.Fatalf("expected one result, got %d", len(queryResultList))
	}

	value := queryResultList[0].Interface(true).(*remote.ObjectValue)
	if value.GetFieldValue("id") != int64(1) {
		t.Fatalf("id mismatch: %#v", value.GetFieldValue("id"))
	}
	if value.GetFieldValue("name") != "alpha" {
		t.Fatalf("name mismatch: %#v", value.GetFieldValue("name"))
	}
	if value.GetFieldValue("description") != nil {
		t.Fatalf("LiteView should exclude description, got %#v", value.GetFieldValue("description"))
	}
}

func TestORMQueryUsesDetailViewResponse(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerQueryMaskViewRemoteModel(t, remoteProvider)

	queryType, err := remoteProvider.GetEntityType(&remote.ObjectValue{
		Name:    "queryMaskViewModel",
		PkgPath: "github.com/muidea/magicOrm/orm",
	})
	if err != nil {
		t.Fatalf("GetEntityType failed: %v", err)
	}
	queryModel, err := remoteProvider.GetTypeModel(queryType)
	if err != nil {
		t.Fatalf("GetTypeModel failed: %v", err)
	}
	queryModel = queryModel.Copy(models.LiteView)
	if err := queryModel.SetFieldValue("id", int64(1)); err != nil {
		t.Fatalf("SetFieldValue(id) failed: %v", err)
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

	queryImpl := &impl{
		context:       context.Background(),
		executor:      executor,
		modelProvider: remoteProvider,
		modelCodec:    codec.New(remoteProvider, "tenant"),
	}
	queryResult, err := queryImpl.Query(queryModel)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	value := queryResult.Interface(true).(*remote.ObjectValue)
	if value.GetFieldValue("id") != int64(1) {
		t.Fatalf("id mismatch: %#v", value.GetFieldValue("id"))
	}
	if value.GetFieldValue("name") != "alpha" {
		t.Fatalf("name mismatch: %#v", value.GetFieldValue("name"))
	}
	if value.GetFieldValue("description") != "detail-description" {
		t.Fatalf("Query should return detail field description, got %#v", value.GetFieldValue("description"))
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
	queryMask, err := buildQueryExecutionModel(responseModel, !responseByMask)
	if err != nil {
		t.Fatalf("buildQueryExecutionModel failed: %v", err)
	}

	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_QueryMaskViewModel") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(1)})
				},
				rows: [][]any{{int64(1), "detail-description"}},
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

func TestORMBatchQueryValueMaskOverridesLiteView(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerQueryMaskViewRemoteModel(t, remoteProvider)

	filter, err := remoteProvider.GetEntityFilter(&remote.ObjectValue{
		Name:    "queryMaskViewModel",
		PkgPath: "github.com/muidea/magicOrm/orm",
	}, models.LiteView)
	if err != nil {
		t.Fatalf("GetEntityFilter failed: %v", err)
	}
	if err := filter.Equal("id", int64(1)); err != nil {
		t.Fatalf("filter.Equal(id) failed: %v", err)
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

	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_QueryMaskViewModel") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(1)})
				},
				rows: [][]any{{int64(1), "detail-description"}},
			},
		},
	}

	queryImpl := &impl{
		context:       context.Background(),
		executor:      executor,
		modelProvider: remoteProvider,
		modelCodec:    codec.New(remoteProvider, "tenant"),
	}
	queryResultList, err := queryImpl.BatchQuery(filter)
	if err != nil {
		t.Fatalf("BatchQuery failed: %v", err)
	}
	if len(queryResultList) != 1 {
		t.Fatalf("expected one result, got %d", len(queryResultList))
	}

	value := queryResultList[0].Interface(true).(*remote.ObjectValue)
	if value.GetFieldValue("id") != int64(1) {
		t.Fatalf("id mismatch: %#v", value.GetFieldValue("id"))
	}
	if value.GetFieldValue("name") != nil {
		t.Fatalf("ValueMask should override LiteView and exclude name, got %#v", value.GetFieldValue("name"))
	}
	if value.GetFieldValue("description") != "detail-description" {
		t.Fatalf("description mismatch: %#v", value.GetFieldValue("description"))
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

func TestApplyQueryResponseModelSkipsValidationDuringProjection(t *testing.T) {
	validator := &countingValidator{}
	responseModel := &remote.Object{
		Name:    "ProjectionModel",
		PkgPath: "github.com/muidea/magicOrm/orm",
		Fields: []*remote.Field{
			{
				Name: "id",
				Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &remote.SpecImpl{FieldName: "id", PrimaryKey: true},
			},
			{
				Name: "name",
				Type: &remote.TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &remote.SpecImpl{FieldName: "name", Constraint: "req"},
			},
		},
	}
	setRemoteObjectValidator(responseModel, validator)

	sourceModel := responseModel.Copy(models.OriginView)
	if err := sourceModel.SetPrimaryFieldValue(int64(7)); err != nil {
		t.Fatalf("SetPrimaryFieldValue failed: %v", err)
	}
	sourceName := sourceModel.GetField("name")
	sourceName.Reset()
	if err := sourceName.GetValue().Set(nil); err != nil {
		t.Fatalf("clear source name failed: %v", err)
	}

	projected := applyQueryResponseModel(sourceModel, responseModel, false)
	if validator.calls != 0 {
		t.Fatalf("projection should not invoke validator, got %d calls", validator.calls)
	}

	nameField := projected.GetField("name")
	if nameField == nil {
		t.Fatal("projected name field is nil")
	}
	if nameField.GetValue().Get() != nil {
		t.Fatalf("projected name should stay nil, got %#v", nameField.GetValue().Get())
	}
}
