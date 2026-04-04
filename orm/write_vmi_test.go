package orm

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/remote"
)

func buildProductInsertModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "name", Value: "apple"},
			{Name: "description", Value: "fresh apple"},
			{Name: "image", Value: []string{"main.png"}},
			{Name: "expire", Value: 30},
			{
				Name: "status",
				Value: &remote.ObjectValue{
					Name:    "status",
					PkgPath: "/vmi",
					Fields: []*remote.FieldValue{
						{Name: "id", Value: int64(9)},
					},
				},
			},
			{Name: "tags", Value: []string{"fruit"}},
		},
	}

	model, err := remoteProvider.GetEntityModel(productValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(product insert value) failed: %v", err)
	}

	object, ok := model.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", model)
	}
	return object
}

func buildProductUpdateModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
			{Name: "name", Value: "apple-updated"},
			{
				Name: "status",
				Value: &remote.ObjectValue{
					Name:    "status",
					PkgPath: "/vmi",
					Fields: []*remote.FieldValue{
						{Name: "id", Value: int64(10)},
					},
				},
			},
		},
	}

	model, err := remoteProvider.GetEntityModel(productValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(product update value) failed: %v", err)
	}

	object, ok := model.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", model)
	}
	return object
}

func buildProductBasicUpdateModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
			{Name: "description", Value: "fresh apple updated"},
		},
	}

	model, err := remoteProvider.GetEntityModel(productValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(product basic update value) failed: %v", err)
	}

	object, ok := model.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", model)
	}
	return object
}

func buildProductDeleteModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
		},
	}

	model, err := remoteProvider.GetEntityModel(productValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(product delete value) failed: %v", err)
	}

	object, ok := model.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", model)
	}
	return object
}

func buildProductClearSkuInfoModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
			{
				Name: "skuInfo",
				Value: &remote.SliceObjectValue{
					Name:    "skuInfo",
					PkgPath: "/vmi/product",
					Values:  []*remote.ObjectValue{},
				},
			},
		},
	}

	model, err := remoteProvider.GetEntityModel(productValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(product clear skuInfo value) failed: %v", err)
	}

	object, ok := model.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", model)
	}
	return object
}

func buildProductNoopSkuInfoModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
		},
	}

	model, err := remoteProvider.GetEntityModel(productValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(product noop skuInfo value) failed: %v", err)
	}

	object, ok := model.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", model)
	}
	return object
}

func buildProductClearStatusModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
			{Name: "status", Value: nil, Assigned: true},
		},
	}

	model, err := remoteProvider.GetEntityModel(productValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(product clear status value) failed: %v", err)
	}

	object, ok := model.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", model)
	}
	return object
}

func buildShelfReadOnlyWarehouseUpdateModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	shelfValue := &remote.ObjectValue{
		Name:    "shelf",
		PkgPath: "/vmi/warehouse",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
			{Name: "capacity", Value: 20},
			{
				Name: "warehouse",
				Value: &remote.ObjectValue{
					Name:    "warehouse",
					PkgPath: "/vmi/warehouse",
					Fields: []*remote.FieldValue{
						{Name: "id", Value: int64(2002)},
					},
				},
			},
		},
	}

	model, err := remoteProvider.GetEntityModel(shelfValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(shelf readonly warehouse update value) failed: %v", err)
	}

	object, ok := model.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", model)
	}
	return object
}

func buildShelfInsertWithoutWarehouseModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	shelfValue := &remote.ObjectValue{
		Name:    "shelf",
		PkgPath: "/vmi/warehouse",
		Fields: []*remote.FieldValue{
			{Name: "capacity", Value: 20},
			{
				Name: "status",
				Value: &remote.ObjectValue{
					Name:    "status",
					PkgPath: "/vmi",
					Fields: []*remote.FieldValue{
						{Name: "id", Value: int64(19)},
					},
				},
			},
		},
	}

	model, err := remoteProvider.GetEntityModel(shelfValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(shelf insert without warehouse) failed: %v", err)
	}

	object, ok := model.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", model)
	}
	return object
}

func buildProductSameSkuInfoModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
			{
				Name: "skuInfo",
				Value: &remote.SliceObjectValue{
					Name:    "skuInfo",
					PkgPath: "/vmi/product",
					Values: []*remote.ObjectValue{
						{
							Name:    "skuInfo",
							PkgPath: "/vmi/product",
							Fields: []*remote.FieldValue{
								{Name: "sku", Value: "sku-001"},
								{Name: "description", Value: "default sku"},
								{Name: "image", Value: []string{"sku-a.png"}},
							},
						},
					},
				},
			},
		},
	}

	model, err := remoteProvider.GetEntityModel(productValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(product same skuInfo value) failed: %v", err)
	}

	object, ok := model.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", model)
	}
	return object
}

func registerRewardPolicyQueryModels(t *testing.T, remoteProvider provider.Provider) {
	t.Helper()

	for _, path := range []string{
		"test/vmi/entity/status.json",
		"test/vmi/entity/bill/rewardPolicy/valueItem.json",
		"test/vmi/entity/bill/rewardPolicy/valueIScope.json",
		"test/vmi/entity/bill/rewardPolicy/rewardPolicy.json",
	} {
		if _, err := remoteProvider.RegisterModel(loadVMIObjectForORMTest(t, path)); err != nil {
			t.Fatalf("RegisterModel(%s) failed: %v", path, err)
		}
	}
}

func buildRewardPolicyUpdateScopeModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	rewardPolicyValue := &remote.ObjectValue{
		Name:    "rewardPolicy",
		PkgPath: "/vmi/bill",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(2001)},
			{
				Name: "scope",
				Value: &remote.ObjectValue{
					Name:    "valueScope",
					PkgPath: "/vmi/bill/rewardPolicy",
					Fields: []*remote.FieldValue{
						{Name: "id", Value: int64(21)},
						{Name: "lowValue", Value: 120.0},
						{Name: "highValue", Value: 1080.0},
					},
				},
			},
		},
	}

	model, err := remoteProvider.GetEntityModel(rewardPolicyValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(rewardPolicy scope update value) failed: %v", err)
	}

	object, ok := model.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", model)
	}
	return object
}

func buildProductPreciseSkuInfoModel(t *testing.T, remoteProvider provider.Provider) *remote.Object {
	t.Helper()

	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
			{
				Name: "skuInfo",
				Value: &remote.SliceObjectValue{
					Name:    "skuInfo",
					PkgPath: "/vmi/product",
					Values: []*remote.ObjectValue{
						{
							Name:    "skuInfo",
							PkgPath: "/vmi/product",
							Fields: []*remote.FieldValue{
								{Name: "sku", Value: "sku-001"},
								{Name: "description", Value: "updated sku"},
								{Name: "image", Value: []string{"sku-a.png"}},
							},
						},
						{
							Name:    "skuInfo",
							PkgPath: "/vmi/product",
							Fields: []*remote.FieldValue{
								{Name: "sku", Value: "sku-003"},
								{Name: "description", Value: "new sku"},
								{Name: "image", Value: []string{"sku-c.png"}},
							},
						},
					},
				},
			},
		},
	}

	model, err := remoteProvider.GetEntityModel(productValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(product precise skuInfo value) failed: %v", err)
	}

	object, ok := model.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", model)
	}
	return object
}

func countCallsByKind(calls []fakeExecCall, kind string) int {
	count := 0
	for _, call := range calls {
		if call.kind == kind {
			count++
		}
	}
	return count
}

func containsSQLCall(calls []fakeExecCall, kind, snippet string, wantArgs []any) bool {
	for _, call := range calls {
		if call.kind != kind {
			continue
		}
		if !strings.Contains(call.sql, snippet) {
			continue
		}
		if wantArgs != nil && !argsEquivalent(call.args, wantArgs) {
			continue
		}
		return true
	}
	return false
}

func argsEquivalent(got, want []any) bool {
	if len(got) != len(want) {
		return false
	}

	for idx := range got {
		if reflect.DeepEqual(got[idx], want[idx]) {
			continue
		}
		if fmt.Sprintf("%v", got[idx]) == fmt.Sprintf("%v", want[idx]) {
			continue
		}
		return false
	}

	return true
}

func TestInsertRunnerVMIRemoteRelations(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductInsertModel(t, remoteProvider)
	executor := &fakeExecutor{
		insertIDs: []any{int64(1001), int64(2001), int64(3001), int64(3002)},
	}
	insertRunner := NewInsertRunner(context.Background(), productModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"))

	insertedModel, err := insertRunner.Insert()
	if err != nil {
		t.Fatalf("InsertRunner.Insert(product) failed: %v", err)
	}

	productValue := insertedModel.Interface(true).(*remote.ObjectValue)
	if productValue.ID != "1001" || productValue.GetFieldValue("id") != int64(1001) {
		t.Fatalf("product primary key should be assigned from insert id, got %#v", productValue)
	}
	statusValue, ok := productValue.GetFieldValue("status").(*remote.ObjectValue)
	if !ok || statusValue.GetFieldValue("id") != int64(9) {
		t.Fatalf("status relation should remain reference by id, got %#v", productValue.GetFieldValue("status"))
	}
	if productValue.GetFieldValue("skuInfo") != nil {
		t.Fatalf("product should not expose removed skuInfo relation, got %#v", productValue.GetFieldValue("skuInfo"))
	}

	if countCallsByKind(executor.execCalls, "insert") != 2 {
		t.Fatalf("expected 2 insert calls, got %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "insert", "tenant_Product", []any{"apple", "fresh apple", `["main.png"]`, 30, `["fruit"]`, int64(0), int64(0), int64(0), ""}) {
		t.Fatalf("missing product host insert call: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "insert", "tenant_ProductStatus3Status", []any{int64(1001), int64(9)}) {
		t.Fatalf("missing status relation insert call: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "insert", "tenant_Status", nil) {
		t.Fatalf("status reference should not trigger host insert: %#v", executor.execCalls)
	}
}

func TestUpdateRunnerVMIRemoteReferenceDiff(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductUpdateModel(t, remoteProvider)
	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ProductStatus3Status") && reflect.DeepEqual(args, []any{int64(1001)})
				},
				rows: [][]any{{int64(9)}},
			},
		},
		insertIDs: []any{int64(5001)},
	}
	updateRunner := NewUpdateRunner(context.Background(), productModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"))

	updatedModel, err := updateRunner.Update()
	if err != nil {
		t.Fatalf("UpdateRunner.Update(product) failed: %v", err)
	}

	productValue := updatedModel.Interface(true).(*remote.ObjectValue)
	statusValue := productValue.GetFieldValue("status").(*remote.ObjectValue)
	if statusValue.GetFieldValue("id") != int64(10) {
		t.Fatalf("updated status relation mismatch: %#v", statusValue)
	}

	if !containsSQLCall(executor.execCalls, "exec", "UPDATE", []any{"apple-updated", int64(1001)}) {
		t.Fatalf("missing product update call: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "query", "tenant_ProductStatus3Status", []any{int64(1001)}) {
		t.Fatalf("missing existing relation query: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "exec", "DELETE FROM", []any{int64(1001), int64(9)}) {
		t.Fatalf("missing relation delete by rights call: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "insert", "tenant_ProductStatus3Status", []any{int64(1001), int64(10)}) {
		t.Fatalf("missing relation insert call: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "exec", "tenant_Status", nil) || containsSQLCall(executor.execCalls, "insert", "tenant_Status", nil) {
		t.Fatalf("status reference diff should not mutate status host table: %#v", executor.execCalls)
	}
}

func TestUpdateSkipsTransactionForBasicFieldOnlyUpdate(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductBasicUpdateModel(t, remoteProvider)
	executor := &fakeExecutor{}
	ormImpl := &impl{
		context:       context.Background(),
		executor:      executor,
		modelProvider: remoteProvider,
		modelCodec:    codec.New(remoteProvider, "tenant"),
	}

	updatedModel, err := ormImpl.Update(productModel)
	if err != nil {
		t.Fatalf("impl.Update(product basic update) failed: %v", err)
	}
	if updatedModel == nil {
		t.Fatal("impl.Update(product basic update) should return model")
	}

	if executor.beginCalls != 0 || executor.commitCalls != 0 || executor.rollbackCalls != 0 {
		t.Fatalf("basic field only update should not open transaction, got begin=%d commit=%d rollback=%d", executor.beginCalls, executor.commitCalls, executor.rollbackCalls)
	}
	if !containsSQLCall(executor.execCalls, "exec", "UPDATE", []any{"fresh apple updated", int64(1001)}) {
		t.Fatalf("missing basic update SQL call: %#v", executor.execCalls)
	}
}

func TestUpdateKeepsTransactionForRelationUpdate(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductUpdateModel(t, remoteProvider)
	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ProductStatus3Status") && reflect.DeepEqual(args, []any{int64(1001)})
				},
				rows: [][]any{{int64(9)}},
			},
		},
		insertIDs: []any{int64(5001)},
	}
	ormImpl := &impl{
		context:       context.Background(),
		executor:      executor,
		modelProvider: remoteProvider,
		modelCodec:    codec.New(remoteProvider, "tenant"),
	}

	updatedModel, err := ormImpl.Update(productModel)
	if err != nil {
		t.Fatalf("impl.Update(product relation update) failed: %v", err)
	}
	if updatedModel == nil {
		t.Fatal("impl.Update(product relation update) should return model")
	}

	if executor.beginCalls != 1 || executor.commitCalls != 1 || executor.rollbackCalls != 0 {
		t.Fatalf("relation update should keep transaction, got begin=%d commit=%d rollback=%d", executor.beginCalls, executor.commitCalls, executor.rollbackCalls)
	}
}

func TestDeleteRunnerVMIRemoteRelations(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductDeleteModel(t, remoteProvider)
	executor := &fakeExecutor{}
	deleteRunner := NewDeleteRunner(context.Background(), productModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"), 0)

	if err := deleteRunner.Delete(); err != nil {
		t.Fatalf("DeleteRunner.Delete(product) failed: %v", err)
	}

	if !containsSQLCall(executor.execCalls, "exec", "DELETE FROM", []any{int64(1001)}) {
		t.Fatalf("missing product host delete: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "exec", "tenant_ProductStatus3Status", []any{int64(1001)}) {
		t.Fatalf("missing status relation delete: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "exec", "tenant_Status", nil) {
		t.Fatalf("status pointer relation should not delete status host: %#v", executor.execCalls)
	}
}

func TestUpdateRunnerVMIRemoteNilReferenceClearsRelation(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductClearStatusModel(t, remoteProvider)
	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ProductStatus3Status") && reflect.DeepEqual(args, []any{int64(1001)})
				},
				rows: [][]any{{int64(9)}},
			},
		},
	}
	updateRunner := NewUpdateRunner(context.Background(), productModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"))

	updatedModel, err := updateRunner.Update()
	if err != nil {
		t.Fatalf("UpdateRunner.Update(product clear status) failed: %v", err)
	}

	productValue := updatedModel.Interface(true).(*remote.ObjectValue)
	foundStatus := false
	for _, field := range productValue.Fields {
		if field.Name == "status" {
			foundStatus = true
			break
		}
	}
	if !foundStatus {
		t.Fatalf("explicit nil status should remain exported for remote object value, got %#v", productValue)
	}
	if productValue.GetFieldValue("status") != nil {
		t.Fatalf("status should be explicit nil after clear, got %#v", productValue.GetFieldValue("status"))
	}

	if !containsSQLCall(executor.execCalls, "query", "tenant_ProductStatus3Status", []any{int64(1001)}) {
		t.Fatalf("missing existing status relation query: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "exec", "tenant_ProductStatus3Status", []any{int64(1001), int64(9)}) {
		t.Fatalf("missing status relation delete by rights: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "insert", "tenant_ProductStatus3Status", nil) {
		t.Fatalf("explicit nil status should not insert relation rows: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "exec", "tenant_Status", nil) || containsSQLCall(executor.execCalls, "insert", "tenant_Status", nil) {
		t.Fatalf("reference clear should not mutate status host table: %#v", executor.execCalls)
	}
}

func TestUpdateRunnerVMIRemoteReadOnlyReferenceIgnored(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	for _, path := range []string{
		"test/vmi/entity/status.json",
		"test/vmi/entity/warehouse/warehouse.json",
		"test/vmi/entity/warehouse/shelf.json",
	} {
		if _, err := remoteProvider.RegisterModel(loadVMIObjectForORMTest(t, path)); err != nil {
			t.Fatalf("RegisterModel(%s) failed: %v", path, err)
		}
	}

	shelfModel := buildShelfReadOnlyWarehouseUpdateModel(t, remoteProvider)
	executor := &fakeExecutor{}
	updateRunner := NewUpdateRunner(context.Background(), shelfModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"))

	updatedModel, err := updateRunner.Update()
	if err != nil {
		t.Fatalf("UpdateRunner.Update(shelf readonly warehouse) failed: %v", err)
	}

	shelfValue := updatedModel.Interface(true).(*remote.ObjectValue)
	warehouseValue, ok := shelfValue.GetFieldValue("warehouse").(*remote.ObjectValue)
	if !ok || warehouseValue.GetFieldValue("id") != int64(2002) {
		t.Fatalf("readonly warehouse relation should remain present on returned model, got %#v", shelfValue.GetFieldValue("warehouse"))
	}

	if !containsSQLCall(executor.execCalls, "exec", "UPDATE \"tenant_Shelf\"", []any{20, int64(1001)}) {
		t.Fatalf("missing shelf host update call: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "query", "tenant_ShelfWarehouse3Warehouse", nil) {
		t.Fatalf("readonly warehouse relation should not query relation rows: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "exec", "tenant_ShelfWarehouse3Warehouse", nil) || containsSQLCall(executor.execCalls, "insert", "tenant_ShelfWarehouse3Warehouse", nil) {
		t.Fatalf("readonly warehouse relation should not mutate relation rows: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "exec", "tenant_Warehouse", nil) || containsSQLCall(executor.execCalls, "insert", "tenant_Warehouse", nil) {
		t.Fatalf("readonly warehouse relation should not mutate warehouse host table: %#v", executor.execCalls)
	}
}

func TestInsertRunnerVMIRemoteMissingRequiredReferenceRejected(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	for _, path := range []string{
		"test/vmi/entity/status.json",
		"test/vmi/entity/warehouse/warehouse.json",
		"test/vmi/entity/warehouse/shelf.json",
	} {
		if _, err := remoteProvider.RegisterModel(loadVMIObjectForORMTest(t, path)); err != nil {
			t.Fatalf("RegisterModel(%s) failed: %v", path, err)
		}
	}

	shelfModel := buildShelfInsertWithoutWarehouseModel(t, remoteProvider)
	insertRunner := NewInsertRunner(context.Background(), shelfModel, &fakeExecutor{}, remoteProvider, codec.New(remoteProvider, "tenant"))

	if _, err := insertRunner.Insert(); err == nil {
		t.Fatal("InsertRunner.Insert should reject missing required warehouse relation")
	} else if !strings.Contains(err.Error(), "warehouse") {
		t.Fatalf("expected warehouse-related error, got %v", err)
	}
}
