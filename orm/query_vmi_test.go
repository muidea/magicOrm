package orm

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/remote"
)

type fakeQueryResponse struct {
	match func(sql string, args []any) bool
	rows  [][]any
}

type fakeExecCall struct {
	kind string
	sql  string
	args []any
}

type fakeExecutor struct {
	responses   []fakeQueryResponse
	currentRows [][]any
	index       int

	execCalls   []fakeExecCall
	insertIDs   []any
	insertIndex int

	beginCalls    int
	commitCalls   int
	rollbackCalls int
}

func (s *fakeExecutor) Release() {}

func (s *fakeExecutor) BeginTransaction() *cd.Error {
	s.beginCalls++
	return nil
}

func (s *fakeExecutor) CommitTransaction() *cd.Error {
	s.commitCalls++
	return nil
}

func (s *fakeExecutor) RollbackTransaction() *cd.Error {
	s.rollbackCalls++
	return nil
}

func (s *fakeExecutor) Query(sql string, _ bool, args ...any) ([]string, *cd.Error) {
	s.execCalls = append(s.execCalls, fakeExecCall{kind: "query", sql: sql, args: append([]any(nil), args...)})
	for _, response := range s.responses {
		if response.match(sql, args) {
			s.currentRows = response.rows
			s.index = -1
			return nil, nil
		}
	}

	return nil, cd.NewError(cd.Unexpected, "unexpected query: "+sql)
}

func (s *fakeExecutor) Next() bool {
	if s.index+1 >= len(s.currentRows) {
		return false
	}
	s.index++
	return true
}

func (s *fakeExecutor) Finish() {
	s.currentRows = nil
	s.index = -1
}

func (s *fakeExecutor) GetField(value ...any) *cd.Error {
	if s.index < 0 || s.index >= len(s.currentRows) {
		return cd.NewError(cd.Unexpected, "no active row")
	}

	row := s.currentRows[s.index]
	if len(row) != len(value) {
		return cd.NewError(cd.Unexpected, "field count mismatch")
	}

	for idx, rawPtr := range value {
		ptrValue := reflect.ValueOf(rawPtr)
		if ptrValue.Kind() != reflect.Ptr || ptrValue.IsNil() {
			return cd.NewError(cd.Unexpected, "target must be pointer")
		}

		assignValue := reflect.ValueOf(row[idx])
		target := ptrValue.Elem()
		if !assignValue.IsValid() {
			target.Set(reflect.Zero(target.Type()))
			continue
		}
		if assignValue.Type().AssignableTo(target.Type()) {
			target.Set(assignValue)
			continue
		}
		if assignValue.Type().ConvertibleTo(target.Type()) {
			target.Set(assignValue.Convert(target.Type()))
			continue
		}
		if target.Kind() == reflect.Interface {
			target.Set(assignValue)
			continue
		}

		return cd.NewError(cd.Unexpected, "can't assign field value")
	}

	return nil
}

func (s *fakeExecutor) Execute(sql string, args ...any) (int64, *cd.Error) {
	s.execCalls = append(s.execCalls, fakeExecCall{kind: "exec", sql: sql, args: append([]any(nil), args...)})
	return 0, nil
}

func (s *fakeExecutor) ExecuteInsert(sql string, pkValOut any, args ...any) *cd.Error {
	s.execCalls = append(s.execCalls, fakeExecCall{kind: "insert", sql: sql, args: append([]any(nil), args...)})
	if pkValOut == nil || s.insertIndex >= len(s.insertIDs) {
		return nil
	}

	ptrValue := reflect.ValueOf(pkValOut)
	if ptrValue.Kind() != reflect.Ptr || ptrValue.IsNil() {
		return cd.NewError(cd.Unexpected, "pkValOut must be pointer")
	}

	assignValue := reflect.ValueOf(s.insertIDs[s.insertIndex])
	s.insertIndex++
	target := ptrValue.Elem()
	if assignValue.IsValid() {
		if assignValue.Type().AssignableTo(target.Type()) {
			target.Set(assignValue)
			return nil
		}
		if assignValue.Type().ConvertibleTo(target.Type()) {
			target.Set(assignValue.Convert(target.Type()))
			return nil
		}
		if target.Kind() == reflect.Interface {
			target.Set(assignValue)
			return nil
		}
	}

	return cd.NewError(cd.Unexpected, "can't assign insert id")
}

func (s *fakeExecutor) CheckTableExist(string) (bool, *cd.Error) { return false, nil }

func loadVMIObjectForORMTest(t *testing.T, relativePath string) *remote.Object {
	t.Helper()

	filePath := filepath.Join("..", relativePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile(%s) failed: %v", relativePath, err)
	}

	object := &remote.Object{}
	if err := json.Unmarshal(data, object); err != nil {
		t.Fatalf("Unmarshal(%s) failed: %v", relativePath, err)
	}

	return object
}

func registerVMIQueryModels(t *testing.T, remoteProvider provider.Provider) {
	t.Helper()

	for _, path := range []string{
		"test/vmi/entity/status.json",
		"test/vmi/entity/product/product.json",
	} {
		if _, err := remoteProvider.RegisterModel(loadVMIObjectForORMTest(t, path)); err != nil {
			t.Fatalf("RegisterModel(%s) failed: %v", path, err)
		}
	}
}

func buildProductQueryModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	productFilterValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
		},
	}
	productModel, err := remoteProvider.GetEntityModel(productFilterValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(productFilterValue) failed: %v", err)
	}

	object, ok := productModel.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", productModel)
	}

	return object
}

func countQueryCallsContaining(execCalls []fakeExecCall, token string) int {
	count := 0
	for _, call := range execCalls {
		if call.kind == "query" && strings.Contains(call.sql, token) {
			count++
		}
	}

	return count
}

func testQueryParentObject() *remote.Object {
	return &remote.Object{
		Name:    "parent",
		PkgPath: "/bench",
		Fields: []*remote.Field{
			{
				Name: "id",
				Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &remote.SpecImpl{FieldName: "id", PrimaryKey: true, ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}},
			},
			{
				Name: "name",
				Type: &remote.TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &remote.SpecImpl{FieldName: "name", ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}},
			},
			{
				Name: "status",
				Type: &remote.TypeImpl{Name: "status", PkgPath: "/bench", Value: models.TypeStructValue, IsPtr: true},
				Spec: &remote.SpecImpl{FieldName: "status", ViewDeclare: []models.ViewDeclare{models.DetailView}},
			},
			{
				Name: "children",
				Type: &remote.TypeImpl{
					Name:    "children",
					PkgPath: "/bench",
					Value:   models.TypeSliceValue,
					ElemType: &remote.TypeImpl{
						Name:    "child",
						PkgPath: "/bench",
						Value:   models.TypeStructValue,
						IsPtr:   true,
					},
				},
				Spec: &remote.SpecImpl{FieldName: "children", ViewDeclare: []models.ViewDeclare{models.DetailView}},
			},
		},
	}
}

func testQueryStatusObject() *remote.Object {
	return &remote.Object{
		Name:    "status",
		PkgPath: "/bench",
		Fields: []*remote.Field{
			{
				Name: "id",
				Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &remote.SpecImpl{FieldName: "id", PrimaryKey: true, ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}},
			},
			{
				Name: "name",
				Type: &remote.TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &remote.SpecImpl{FieldName: "name", ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}},
			},
			{
				Name: "owner",
				Type: &remote.TypeImpl{Name: "child", PkgPath: "/bench", Value: models.TypeStructValue, IsPtr: true},
				Spec: &remote.SpecImpl{FieldName: "owner", ViewDeclare: []models.ViewDeclare{models.DetailView}},
			},
		},
	}
}

func testQueryChildObject() *remote.Object {
	return &remote.Object{
		Name:    "child",
		PkgPath: "/bench",
		Fields: []*remote.Field{
			{
				Name: "id",
				Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &remote.SpecImpl{FieldName: "id", PrimaryKey: true, ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}},
			},
			{
				Name: "name",
				Type: &remote.TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &remote.SpecImpl{FieldName: "name", ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}},
			},
		},
	}
}

func registerMinimalRelationModels(t *testing.T, remoteProvider provider.Provider) {
	t.Helper()

	for _, object := range []*remote.Object{
		testQueryParentObject(),
		testQueryStatusObject(),
		testQueryChildObject(),
	} {
		if _, err := remoteProvider.RegisterModel(object); err != nil {
			t.Fatalf("RegisterModel(%s) failed: %v", object.GetPkgKey(), err)
		}
	}
}

func TestQueryRunnerVMIRemoteRelations(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductQueryModel(t, remoteProvider)
	modelCodec := codec.New(remoteProvider, "tenant")
	filter, err := getModelFilter(productModel, remoteProvider, modelCodec)
	if err != nil {
		t.Fatalf("getModelFilter(product) failed: %v", err)
	}

	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_Product") &&
						!strings.Contains(sql, "tenant_ProductStatus3Status") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(1001)})
				},
				rows: [][]any{
					{int64(1001), "apple", "fresh apple", `["main.png","detail.png"]`, 30, `["fruit","fresh"]`, int64(0), int64(0), int64(0), ""},
				},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ProductStatus3Status") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(1001)})
				},
				rows: [][]any{{int64(9)}},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_Status") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(9)})
				},
				rows: [][]any{
					{int64(9), 2, "published"},
				},
			},
		},
	}

	responseModel, responseByMask, err := buildQueryResponseModel(nil, filter)
	if err != nil {
		t.Fatalf("buildQueryResponseModel failed: %v", err)
	}
	queryRunner := NewQueryRunner(context.Background(), filter.MaskModel(), responseModel, responseByMask, executor, remoteProvider, modelCodec, false, 0)
	modelsList, err := queryRunner.Query(filter)
	if err != nil {
		t.Fatalf("QueryRunner.Query(product) failed: %v", err)
	}
	if len(modelsList) != 1 {
		t.Fatalf("QueryRunner.Query(product) got %d results", len(modelsList))
	}

	productValue, ok := modelsList[0].Interface(true).(*remote.ObjectValue)
	if !ok {
		t.Fatalf("Interface(true) should return *remote.ObjectValue, got %T", modelsList[0].Interface(true))
	}

	if productValue.ID != "1001" {
		t.Fatalf("product id mismatch: %s", productValue.ID)
	}
	if productValue.GetFieldValue("name") != "apple" || productValue.GetFieldValue("description") != "fresh apple" {
		t.Fatalf("product basic fields mismatch: %#v", productValue)
	}
	image, ok := productValue.GetFieldValue("image").([]string)
	if !ok || !reflect.DeepEqual(image, []string{"main.png", "detail.png"}) {
		t.Fatalf("product image mismatch: %#v", productValue.GetFieldValue("image"))
	}
	tags, ok := productValue.GetFieldValue("tags").([]string)
	if !ok || !reflect.DeepEqual(tags, []string{"fruit", "fresh"}) {
		t.Fatalf("product tags mismatch: %#v", productValue.GetFieldValue("tags"))
	}

	statusValue, ok := productValue.GetFieldValue("status").(*remote.ObjectValue)
	if !ok {
		t.Fatalf("product status should be *remote.ObjectValue, got %T", productValue.GetFieldValue("status"))
	}
	if statusValue.GetFieldValue("id") != int64(9) || statusValue.GetFieldValue("name") != "published" {
		t.Fatalf("status relation mismatch: %#v", statusValue)
	}

	if productValue.GetFieldValue("skuInfo") != nil {
		t.Fatalf("product should not expose removed skuInfo relation, got %#v", productValue.GetFieldValue("skuInfo"))
	}
}

func TestQueryRunnerVMIRemoteMissingPointerRelation(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductQueryModel(t, remoteProvider)
	modelCodec := codec.New(remoteProvider, "tenant")
	filter, err := getModelFilter(productModel, remoteProvider, modelCodec)
	if err != nil {
		t.Fatalf("getModelFilter(product) failed: %v", err)
	}

	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_Product") &&
						!strings.Contains(sql, "tenant_ProductStatus3Status")
				},
				rows: [][]any{
					{int64(1001), "apple", "fresh apple", `["main.png"]`, 30, `["fruit"]`, int64(0), int64(0), int64(0), ""},
				},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ProductStatus3Status")
				},
				rows: [][]any{},
			},
		},
	}

	responseModel, responseByMask, err := buildQueryResponseModel(nil, filter)
	if err != nil {
		t.Fatalf("buildQueryResponseModel failed: %v", err)
	}
	queryRunner := NewQueryRunner(context.Background(), filter.MaskModel(), responseModel, responseByMask, executor, remoteProvider, modelCodec, false, 0)
	modelsList, err := queryRunner.Query(filter)
	if err != nil {
		t.Fatalf("QueryRunner.Query(product) failed: %v", err)
	}
	productValue := modelsList[0].Interface(true).(*remote.ObjectValue)

	if productValue.GetFieldValue("status") != nil {
		t.Fatalf("missing pointer relation should stay nil, got %#v", productValue.GetFieldValue("status"))
	}
	if productValue.GetFieldValue("skuInfo") != nil {
		t.Fatalf("product should not expose removed skuInfo relation, got %#v", productValue.GetFieldValue("skuInfo"))
	}
}

func TestBuildQueryRejectsRelationWithoutPrimaryKey(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productFilterValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{
				Name: "status",
				Value: &remote.ObjectValue{
					Name:    "status",
					PkgPath: "/vmi",
					Fields: []*remote.FieldValue{
						{Name: "name", Value: "published"},
					},
				},
			},
		},
	}
	productModel, err := remoteProvider.GetEntityModel(productFilterValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(productFilterValue) failed: %v", err)
	}

	modelCodec := codec.New(remoteProvider, "tenant")
	filter, err := getModelFilter(productModel, remoteProvider, modelCodec)
	if err != nil {
		t.Fatalf("getModelFilter(product) failed: %v", err)
	}

	builder := mysql.NewBuilder(remoteProvider, modelCodec)
	_, err = builder.BuildQuery(productModel, filter)
	if err == nil {
		t.Fatal("BuildQuery(product relation without primary key) should fail")
	}
	if err.Code != cd.IllegalParam {
		t.Fatalf("BuildQuery(product relation without primary key) code mismatch, got %v", err.Code)
	}
}

func TestQueryRejectsMultipleMatches(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductQueryModel(t, remoteProvider)
	modelCodec := codec.New(remoteProvider, "tenant")

	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_Product") &&
						!strings.Contains(sql, "tenant_ProductStatus3Status") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(1001)})
				},
				rows: [][]any{
					{int64(1001), "apple", "fresh apple", `["main.png"]`, 30, `["fruit"]`, int64(0), int64(0), int64(0), ""},
					{int64(1001), "apple-dup", "duplicate apple", `["dup.png"]`, 31, `["fruit"]`, int64(0), int64(0), int64(0), ""},
				},
			},
		},
	}

	queryImpl := &impl{
		context:       context.Background(),
		executor:      executor,
		modelProvider: remoteProvider,
		modelCodec:    modelCodec,
	}

	_, err := queryImpl.Query(productModel)
	if err == nil {
		t.Fatal("Query(product) with multiple matches should fail")
	}
	if err.Code != cd.Unexpected {
		t.Fatalf("Query(product) multiple matches code mismatch, got %v", err.Code)
	}
}

func TestQueryRunnerCachesRepeatedPointerRelations(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerMinimalRelationModels(t, remoteProvider)

	filter, err := remoteProvider.GetEntityFilter(testQueryParentObject(), models.DetailView)
	if err != nil {
		t.Fatalf("GetEntityFilter(parent) failed: %v", err)
	}

	modelCodec := codec.New(remoteProvider, "tenant")
	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_Parent") &&
						!strings.Contains(sql, "tenant_ParentStatus") &&
						!strings.Contains(sql, "tenant_ParentChildren") &&
						len(args) == 0
				},
				rows: [][]any{
					{int64(1), "parent-1"},
					{int64(2), "parent-2"},
				},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ParentStatus") &&
						len(args) == 2 && reflect.DeepEqual(args, []any{int64(1), int64(2)})
				},
				rows: [][]any{
					{int64(1), int64(9)},
					{int64(2), int64(9)},
				},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_Status") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(9)})
				},
				rows: [][]any{
					{int64(9), "shared-status"},
				},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ParentChildren") &&
						len(args) == 2 && reflect.DeepEqual(args, []any{int64(1), int64(2)})
				},
				rows: [][]any{},
			},
		},
	}

	responseModel, responseByMask, err := buildQueryResponseModel(nil, filter)
	if err != nil {
		t.Fatalf("buildQueryResponseModel failed: %v", err)
	}
	queryRunner := NewQueryRunner(context.Background(), filter.MaskModel(), responseModel, responseByMask, executor, remoteProvider, modelCodec, true, 0)
	modelsList, err := queryRunner.Query(filter)
	if err != nil {
		t.Fatalf("QueryRunner.Query(parent list) failed: %v", err)
	}
	if len(modelsList) != 2 {
		t.Fatalf("expected 2 parent results, got %d", len(modelsList))
	}
	if countQueryCallsContaining(executor.execCalls, "tenant_ParentStatus") != 1 {
		t.Fatalf("expected parent status relation keys to be prefetched once, got %#v", executor.execCalls)
	}
	if countQueryCallsContaining(executor.execCalls, "tenant_StatusOwner") != 0 {
		t.Fatalf("expected prefetched relation target to avoid nested relation loading, got %#v", executor.execCalls)
	}
	if countQueryCallsContaining(executor.execCalls, "tenant_Status") != 1 {
		t.Fatalf("expected shared status to be queried once, got %#v", executor.execCalls)
	}
}

func TestQueryRunnerBatchesSliceRelations(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerMinimalRelationModels(t, remoteProvider)

	filter, err := remoteProvider.GetEntityFilter(testQueryParentObject(), models.DetailView)
	if err != nil {
		t.Fatalf("GetEntityFilter(parent) failed: %v", err)
	}
	if err := filter.Equal("id", int64(1)); err != nil {
		t.Fatalf("filter.Equal(id) failed: %v", err)
	}

	modelCodec := codec.New(remoteProvider, "tenant")
	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_Parent") &&
						!strings.Contains(sql, "tenant_ParentStatus") &&
						!strings.Contains(sql, "tenant_ParentChildren") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(1)})
				},
				rows: [][]any{
					{int64(1), "parent-1"},
				},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ParentStatus") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(1)})
				},
				rows: [][]any{},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ParentChildren") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(1)})
				},
				rows: [][]any{{int64(101)}, {int64(102)}},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_Child") &&
						len(args) == 2 &&
						reflect.DeepEqual(args, []any{int64(101), int64(102)})
				},
				rows: [][]any{
					{int64(101), "child-1"},
					{int64(102), "child-2"},
				},
			},
		},
	}

	responseModel, responseByMask, err := buildQueryResponseModel(nil, filter)
	if err != nil {
		t.Fatalf("buildQueryResponseModel failed: %v", err)
	}
	queryRunner := NewQueryRunner(context.Background(), filter.MaskModel(), responseModel, responseByMask, executor, remoteProvider, modelCodec, false, 0)
	modelsList, err := queryRunner.Query(filter)
	if err != nil {
		t.Fatalf("QueryRunner.Query(parent) failed: %v", err)
	}
	if len(modelsList) != 1 {
		t.Fatalf("expected 1 parent result, got %d", len(modelsList))
	}

	parentValue := modelsList[0].Interface(true).(*remote.ObjectValue)
	children, ok := parentValue.GetFieldValue("children").(*remote.SliceObjectValue)
	if !ok || children == nil || len(children.Values) != 2 {
		t.Fatalf("expected 2 child relations, got %#v", parentValue.GetFieldValue("children"))
	}
	if countQueryCallsContaining(executor.execCalls, "tenant_Child") != 1 {
		t.Fatalf("expected child relation host query to be batched once, got %#v", executor.execCalls)
	}
}

func TestQueryRunnerCachesMissingPointerRelations(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerMinimalRelationModels(t, remoteProvider)

	filter, err := remoteProvider.GetEntityFilter(testQueryParentObject(), models.DetailView)
	if err != nil {
		t.Fatalf("GetEntityFilter(parent) failed: %v", err)
	}

	modelCodec := codec.New(remoteProvider, "tenant")
	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_Parent") &&
						!strings.Contains(sql, "tenant_ParentStatus") &&
						!strings.Contains(sql, "tenant_ParentChildren") &&
						len(args) == 0
				},
				rows: [][]any{
					{int64(1), "parent-1"},
					{int64(2), "parent-2"},
				},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ParentStatus") &&
						len(args) == 2 && reflect.DeepEqual(args, []any{int64(1), int64(2)})
				},
				rows: [][]any{
					{int64(1), int64(9)},
					{int64(2), int64(9)},
				},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ParentChildren") &&
						len(args) == 2 && reflect.DeepEqual(args, []any{int64(1), int64(2)})
				},
				rows: [][]any{},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_Status") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(9)})
				},
				rows: [][]any{},
			},
		},
	}

	responseModel, responseByMask, err := buildQueryResponseModel(nil, filter)
	if err != nil {
		t.Fatalf("buildQueryResponseModel failed: %v", err)
	}
	queryRunner := NewQueryRunner(context.Background(), filter.MaskModel(), responseModel, responseByMask, executor, remoteProvider, modelCodec, true, 0)
	modelsList, err := queryRunner.Query(filter)
	if err != nil {
		t.Fatalf("QueryRunner.Query(parent list) failed: %v", err)
	}
	if len(modelsList) != 2 {
		t.Fatalf("expected 2 parent results, got %d", len(modelsList))
	}
	if countQueryCallsContaining(executor.execCalls, "tenant_ParentStatus") != 1 {
		t.Fatalf("expected missing status relation keys to be prefetched once, got %#v", executor.execCalls)
	}
	if countQueryCallsContaining(executor.execCalls, "tenant_ParentChildren") != 1 {
		t.Fatalf("expected missing children relation keys to be prefetched once, got %#v", executor.execCalls)
	}

	for idx, modelVal := range modelsList {
		parentValue := modelVal.Interface(true).(*remote.ObjectValue)
		statusVal := parentValue.GetFieldValue("status")
		if statusVal == nil {
			continue
		}

		statusObject, ok := statusVal.(*remote.ObjectValue)
		if !ok {
			t.Fatalf("missing status relation on parent[%d] should not materialize non-object data, got %#v", idx, statusVal)
		}
		if statusObject.GetFieldValue("id") != nil || statusObject.GetFieldValue("name") != nil {
			t.Fatalf("missing status relation on parent[%d] should not produce assigned relation data, got %#v", idx, statusVal)
		}
	}

	if countQueryCallsContaining(executor.execCalls, "tenant_Status") != 1 {
		t.Fatalf("expected missing relation target query to execute once, got %#v", executor.execCalls)
	}
}

func TestQueryRunnerSkipsExcludedRelationsInLiteView(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerMinimalRelationModels(t, remoteProvider)

	filter, err := remoteProvider.GetEntityFilter(testQueryParentObject(), models.LiteView)
	if err != nil {
		t.Fatalf("GetEntityFilter(parent lite) failed: %v", err)
	}

	modelCodec := codec.New(remoteProvider, "tenant")
	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_Parent") &&
						!strings.Contains(sql, "tenant_ParentStatus") &&
						!strings.Contains(sql, "tenant_ParentChildren") &&
						len(args) == 0
				},
				rows: [][]any{
					{int64(1), "parent-1"},
					{int64(2), "parent-2"},
				},
			},
		},
	}

	responseModel, responseByMask, err := buildQueryResponseModel(nil, filter)
	if err != nil {
		t.Fatalf("buildQueryResponseModel failed: %v", err)
	}
	queryRunner := NewQueryRunner(context.Background(), filter.MaskModel(), responseModel, responseByMask, executor, remoteProvider, modelCodec, true, 0)
	modelsList, err := queryRunner.Query(filter)
	if err != nil {
		t.Fatalf("QueryRunner.Query(parent lite list) failed: %v", err)
	}
	if len(modelsList) != 2 {
		t.Fatalf("expected 2 parent results, got %d", len(modelsList))
	}
	if countQueryCallsContaining(executor.execCalls, "tenant_ParentStatus") != 0 {
		t.Fatalf("expected lite view to skip status relation loading, got %#v", executor.execCalls)
	}
	if countQueryCallsContaining(executor.execCalls, "tenant_ParentChildren") != 0 {
		t.Fatalf("expected lite view to skip children relation loading, got %#v", executor.execCalls)
	}
}
