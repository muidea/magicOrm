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
							},
						},
					},
				},
			},
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
	skuInfoValue, ok := productValue.GetFieldValue("skuInfo").(*remote.SliceObjectValue)
	if !ok || len(skuInfoValue.Values) != 1 || skuInfoValue.Values[0].GetFieldValue("sku") != "sku-001" {
		t.Fatalf("skuInfo relation should remain assigned after insert, got %#v", productValue.GetFieldValue("skuInfo"))
	}

	if countCallsByKind(executor.execCalls, "insert") != 4 {
		t.Fatalf("expected 4 insert calls, got %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "insert", "tenant_Product", []any{"apple", "fresh apple", `["main.png"]`, 30, `["fruit"]`, int64(0), int64(0), ""}) {
		t.Fatalf("missing product host insert call: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "insert", "tenant_SkuInfo", []any{"sku-001", "default sku", "[]", 0, 0, ""}) {
		t.Fatalf("missing skuInfo host insert call: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "insert", "tenant_ProductSkuInfo2SkuInfo", []any{int64(1001), "sku-001"}) {
		t.Fatalf("missing skuInfo relation insert call: %#v", executor.execCalls)
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

	if !containsSQLCall(executor.execCalls, "exec", "UPDATE", []any{"apple-updated", "", "[]", 0, "[]", 0, 0, "", int64(1001)}) {
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

func TestDeleteRunnerVMIRemoteRelations(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductDeleteModel(t, remoteProvider)
	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ProductSkuInfo2SkuInfo") && reflect.DeepEqual(args, []any{int64(1001)})
				},
				rows: [][]any{{"sku-001"}},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_SkuInfo") && reflect.DeepEqual(args, []any{"sku-001"})
				},
				rows: [][]any{{"sku-001", "default sku", `[]`, int64(0), int64(0), ""}},
			},
		},
	}
	deleteRunner := NewDeleteRunner(context.Background(), productModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"), 0)

	if err := deleteRunner.Delete(); err != nil {
		t.Fatalf("DeleteRunner.Delete(product) failed: %v", err)
	}

	if !containsSQLCall(executor.execCalls, "exec", "DELETE FROM", []any{int64(1001)}) {
		t.Fatalf("missing product host delete: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "query", "tenant_ProductSkuInfo2SkuInfo", []any{int64(1001)}) {
		t.Fatalf("missing skuInfo relation query: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "exec", "tenant_SkuInfo", []any{"sku-001"}) {
		t.Fatalf("missing recursive skuInfo host delete: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "exec", "tenant_ProductSkuInfo2SkuInfo", []any{int64(1001)}) {
		t.Fatalf("missing skuInfo relation delete: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "exec", "tenant_ProductStatus3Status", []any{int64(1001)}) {
		t.Fatalf("missing status relation delete: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "exec", "tenant_Status", nil) {
		t.Fatalf("status pointer relation should not delete status host: %#v", executor.execCalls)
	}
}

func TestUpdateRunnerVMIRemoteEmptySliceClearsContainRelation(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductClearSkuInfoModel(t, remoteProvider)
	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ProductSkuInfo2SkuInfo") && reflect.DeepEqual(args, []any{int64(1001)})
				},
				rows: [][]any{{"sku-001"}},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_SkuInfo") && reflect.DeepEqual(args, []any{"sku-001"})
				},
				rows: [][]any{{"sku-001", "default sku", `[]`, int64(0), int64(0), ""}},
			},
		},
	}
	updateRunner := NewUpdateRunner(context.Background(), productModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"))

	updatedModel, err := updateRunner.Update()
	if err != nil {
		t.Fatalf("UpdateRunner.Update(product clear skuInfo) failed: %v", err)
	}

	productValue := updatedModel.Interface(true).(*remote.ObjectValue)
	skuInfoValue, ok := productValue.GetFieldValue("skuInfo").(*remote.SliceObjectValue)
	if !ok {
		t.Fatalf("skuInfo should stay as assigned slice object value, got %#v", productValue.GetFieldValue("skuInfo"))
	}
	if skuInfoValue.Values == nil || len(skuInfoValue.Values) != 0 {
		t.Fatalf("skuInfo should remain explicit empty slice, got %#v", skuInfoValue.Values)
	}

	if !containsSQLCall(executor.execCalls, "query", "tenant_ProductSkuInfo2SkuInfo", []any{int64(1001)}) {
		t.Fatalf("missing existing skuInfo relation query: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "exec", "tenant_SkuInfo", []any{"sku-001"}) {
		t.Fatalf("missing contained skuInfo host delete: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "exec", "tenant_ProductSkuInfo2SkuInfo", []any{int64(1001), "sku-001"}) {
		t.Fatalf("missing skuInfo relation delete by rights: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "insert", "tenant_ProductSkuInfo2SkuInfo", nil) {
		t.Fatalf("explicit empty skuInfo should not reinsert relation rows: %#v", executor.execCalls)
	}
}

func TestUpdateRunnerVMIRemoteNilSliceDoesNotTouchContainRelation(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductNoopSkuInfoModel(t, remoteProvider)
	executor := &fakeExecutor{}
	updateRunner := NewUpdateRunner(context.Background(), productModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"))

	updatedModel, err := updateRunner.Update()
	if err != nil {
		t.Fatalf("UpdateRunner.Update(product noop skuInfo) failed: %v", err)
	}

	productValue := updatedModel.Interface(true).(*remote.ObjectValue)
	skuInfoValue, ok := productValue.GetFieldValue("skuInfo").(*remote.SliceObjectValue)
	if !ok {
		t.Fatalf("unassigned skuInfo should keep slice object shell, got %#v", productValue.GetFieldValue("skuInfo"))
	}
	if skuInfoValue.Values != nil {
		t.Fatalf("unassigned skuInfo should remain nil-backed slice value, got %#v", skuInfoValue.Values)
	}

	if containsSQLCall(executor.execCalls, "query", "tenant_ProductSkuInfo2SkuInfo", []any{int64(1001)}) {
		t.Fatalf("unassigned skuInfo should not query existing relations: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "exec", "tenant_ProductSkuInfo2SkuInfo", nil) || containsSQLCall(executor.execCalls, "insert", "tenant_ProductSkuInfo2SkuInfo", nil) {
		t.Fatalf("unassigned skuInfo should not mutate relation rows: %#v", executor.execCalls)
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

func TestUpdateRunnerVMIRemoteSameContainRelationSkipsReplace(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductSameSkuInfoModel(t, remoteProvider)
	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ProductSkuInfo2SkuInfo") && reflect.DeepEqual(args, []any{int64(1001)})
				},
				rows: [][]any{{"sku-001"}},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_SkuInfo") && reflect.DeepEqual(args, []any{"sku-001"})
				},
				rows: [][]any{{"sku-001", "default sku", `["sku-a.png"]`, int64(0), int64(0), ""}},
			},
		},
	}
	updateRunner := NewUpdateRunner(context.Background(), productModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"))

	updatedModel, err := updateRunner.Update()
	if err != nil {
		t.Fatalf("UpdateRunner.Update(product same skuInfo) failed: %v", err)
	}

	productValue := updatedModel.Interface(true).(*remote.ObjectValue)
	skuInfoValue, ok := productValue.GetFieldValue("skuInfo").(*remote.SliceObjectValue)
	if !ok || len(skuInfoValue.Values) != 1 || skuInfoValue.Values[0].GetFieldValue("sku") != "sku-001" {
		t.Fatalf("skuInfo should remain assigned after noop contain update, got %#v", productValue.GetFieldValue("skuInfo"))
	}

	if !containsSQLCall(executor.execCalls, "query", "tenant_ProductSkuInfo2SkuInfo", []any{int64(1001)}) {
		t.Fatalf("missing existing skuInfo relation query: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "query", "tenant_SkuInfo", []any{"sku-001"}) {
		t.Fatalf("missing existing skuInfo host query: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "exec", "tenant_ProductSkuInfo2SkuInfo", nil) {
		t.Fatalf("unchanged skuInfo should not delete relation rows: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "insert", "tenant_ProductSkuInfo2SkuInfo", nil) {
		t.Fatalf("unchanged skuInfo should not insert relation rows: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "exec", "tenant_SkuInfo", nil) {
		t.Fatalf("unchanged skuInfo should not delete child hosts: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "insert", "tenant_SkuInfo", nil) {
		t.Fatalf("unchanged skuInfo should not insert child hosts: %#v", executor.execCalls)
	}
}

func TestUpdateRunnerVMIRemoteContainSliceUsesPrimaryDiff(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerVMIQueryModels(t, remoteProvider)

	productModel := buildProductPreciseSkuInfoModel(t, remoteProvider)
	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ProductSkuInfo2SkuInfo") && reflect.DeepEqual(args, []any{int64(1001)})
				},
				rows: [][]any{{"sku-001"}, {"sku-002"}},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_SkuInfo") && reflect.DeepEqual(args, []any{"sku-001"})
				},
				rows: [][]any{{"sku-001", "default sku", `["sku-a.png"]`, int64(0), int64(0), ""}},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_SkuInfo") && reflect.DeepEqual(args, []any{"sku-002"})
				},
				rows: [][]any{{"sku-002", "delete sku", `["sku-b.png"]`, int64(0), int64(0), ""}},
			},
		},
		insertIDs: []any{int64(7001), int64(7002)},
	}
	updateRunner := NewUpdateRunner(context.Background(), productModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"))

	updatedModel, err := updateRunner.Update()
	if err != nil {
		t.Fatalf("UpdateRunner.Update(product precise skuInfo) failed: %v", err)
	}

	productValue := updatedModel.Interface(true).(*remote.ObjectValue)
	skuInfoValue := productValue.GetFieldValue("skuInfo").(*remote.SliceObjectValue)
	if len(skuInfoValue.Values) != 2 {
		t.Fatalf("skuInfo should keep two rows after precise diff update, got %#v", skuInfoValue)
	}
	if skuInfoValue.Values[0].GetFieldValue("description") != "updated sku" {
		t.Fatalf("existing sku should be updated in place, got %#v", skuInfoValue.Values[0])
	}
	if skuInfoValue.Values[1].GetFieldValue("sku") != "sku-003" {
		t.Fatalf("new sku should be appended, got %#v", skuInfoValue.Values[1])
	}

	if !containsSQLCall(executor.execCalls, "exec", "tenant_SkuInfo", []any{"updated sku", `["sku-a.png"]`, int64(0), int64(0), "", "sku-001"}) {
		t.Fatalf("missing in-place skuInfo host update: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "exec", "tenant_SkuInfo", []any{"sku-002"}) {
		t.Fatalf("missing removed skuInfo host delete: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "exec", "tenant_ProductSkuInfo2SkuInfo", []any{int64(1001), "sku-002"}) {
		t.Fatalf("missing removed skuInfo relation delete by rights: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "insert", "tenant_SkuInfo", []any{"sku-003", "new sku", `["sku-c.png"]`, int64(0), int64(0), ""}) {
		t.Fatalf("missing new skuInfo host insert: %#v", executor.execCalls)
	}
	if !containsSQLCall(executor.execCalls, "insert", "tenant_ProductSkuInfo2SkuInfo", []any{int64(1001), "sku-003"}) {
		t.Fatalf("missing new skuInfo relation insert: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "exec", "tenant_ProductSkuInfo2SkuInfo", []any{int64(1001)}) {
		t.Fatalf("precise contain diff should not delete all relation rows: %#v", executor.execCalls)
	}
}

func TestUpdateRunnerVMIRemoteContainSingleUsesPrimaryDiff(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerRewardPolicyQueryModels(t, remoteProvider)

	rewardPolicyModel := buildRewardPolicyUpdateScopeModel(t, remoteProvider)
	executor := &fakeExecutor{
		responses: []fakeQueryResponse{
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_RewardPolicyScope1ValueScope") && reflect.DeepEqual(args, []any{int64(2001)})
				},
				rows: [][]any{{int64(21)}},
			},
			{
				match: func(sql string, args []any) bool {
					return strings.Contains(sql, "tenant_ValueScope") && reflect.DeepEqual(args, []any{int64(21)})
				},
				rows: [][]any{{int64(21), 100.0, 999.0}},
			},
		},
	}
	updateRunner := NewUpdateRunner(context.Background(), rewardPolicyModel, executor, remoteProvider, codec.New(remoteProvider, "tenant"))

	updatedModel, err := updateRunner.Update()
	if err != nil {
		t.Fatalf("UpdateRunner.Update(rewardPolicy scope) failed: %v", err)
	}

	rewardPolicyValue := updatedModel.Interface(true).(*remote.ObjectValue)
	scopeValue := rewardPolicyValue.GetFieldValue("scope").(*remote.ObjectValue)
	if scopeValue.GetFieldValue("id") != int64(21) || scopeValue.GetFieldValue("lowValue") != 120.0 {
		t.Fatalf("scope should be updated in place, got %#v", scopeValue)
	}

	if !containsSQLCall(executor.execCalls, "exec", "tenant_ValueScope", []any{120.0, 1080.0, int64(21)}) {
		t.Fatalf("missing in-place valueScope update: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "exec", "tenant_RewardPolicyScope1ValueScope", nil) || containsSQLCall(executor.execCalls, "insert", "tenant_RewardPolicyScope1ValueScope", nil) {
		t.Fatalf("single contain same primary should not recreate relation rows: %#v", executor.execCalls)
	}
	if containsSQLCall(executor.execCalls, "exec", "tenant_ValueScope", []any{int64(21)}) || containsSQLCall(executor.execCalls, "insert", "tenant_ValueScope", nil) {
		t.Fatalf("single contain same primary should not delete or insert child host: %#v", executor.execCalls)
	}
}
