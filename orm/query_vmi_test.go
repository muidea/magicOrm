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
}

func (s *fakeExecutor) Release() {}

func (s *fakeExecutor) BeginTransaction() *cd.Error { return nil }

func (s *fakeExecutor) CommitTransaction() *cd.Error { return nil }

func (s *fakeExecutor) RollbackTransaction() *cd.Error { return nil }

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
		"test/vmi/entity/product/skuinfo.json",
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
						!strings.Contains(sql, "tenant_ProductSkuInfo2SkuInfo") &&
						!strings.Contains(sql, "tenant_ProductStatus3Status") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(1001)})
				},
				rows: [][]any{
					{int64(1001), "apple", "fresh apple", `["main.png","detail.png"]`, 30, `["fruit","fresh"]`, int64(0), int64(0), ""},
				},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ProductSkuInfo2SkuInfo") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{int64(1001)})
				},
				rows: [][]any{{"sku-001"}},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_SkuInfo") &&
						len(args) == 1 && reflect.DeepEqual(args, []any{"sku-001"})
				},
				rows: [][]any{
					{"sku-001", "default sku", `["sku-a.png"]`, int64(0), int64(0), ""},
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

	skuInfoValue, ok := productValue.GetFieldValue("skuInfo").(*remote.SliceObjectValue)
	if !ok {
		t.Fatalf("product skuInfo should be *remote.SliceObjectValue, got %T", productValue.GetFieldValue("skuInfo"))
	}
	if len(skuInfoValue.Values) != 1 {
		t.Fatalf("product skuInfo length mismatch: %#v", skuInfoValue)
	}
	if skuInfoValue.Values[0].GetFieldValue("sku") != "sku-001" || skuInfoValue.Values[0].GetFieldValue("description") != "default sku" {
		t.Fatalf("skuInfo relation mismatch: %#v", skuInfoValue.Values[0])
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
						!strings.Contains(sql, "tenant_ProductSkuInfo2SkuInfo") &&
						!strings.Contains(sql, "tenant_ProductStatus3Status")
				},
				rows: [][]any{
					{int64(1001), "apple", "fresh apple", `["main.png"]`, 30, `["fruit"]`, int64(0), int64(0), ""},
				},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ProductSkuInfo2SkuInfo")
				},
				rows: [][]any{},
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
	skuInfoValue, ok := productValue.GetFieldValue("skuInfo").(*remote.SliceObjectValue)
	if !ok {
		t.Fatalf("skuInfo should remain *remote.SliceObjectValue, got %T", productValue.GetFieldValue("skuInfo"))
	}
	if skuInfoValue.Values != nil {
		t.Fatalf("missing slice relation should stay unassigned nil slice, got %#v", skuInfoValue.Values)
	}
}
