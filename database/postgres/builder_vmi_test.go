package postgres

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/remote"
)

func loadVMIRemoteObject(t *testing.T, relativePath string) *remote.Object {
	t.Helper()

	filePath := filepath.Join("..", "..", relativePath)
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

func registerAllVMIRemoteModels(t *testing.T, remoteProvider provider.Provider) {
	t.Helper()

	root := filepath.Join("..", "..", "test", "vmi", "entity")
	paths := make([]string, 0, 16)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		relativePath, err := filepath.Rel(filepath.Join("..", ".."), path)
		if err != nil {
			return err
		}
		paths = append(paths, relativePath)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(%s) failed: %v", root, err)
	}

	sort.Strings(paths)
	for _, path := range paths {
		if _, err := remoteProvider.RegisterModel(loadVMIRemoteObject(t, path)); err != nil {
			t.Fatalf("RegisterModel(%s) failed: %v", path, err)
		}
	}
}

func buildVMIOrderFilter(t *testing.T) (provider.Provider, *Builder, *remote.Object, *remote.ObjectFilter, *remote.SliceObjectValue, *remote.ObjectValue) {
	t.Helper()

	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerAllVMIRemoteModels(t, remoteProvider)

	orderObject := loadVMIRemoteObject(t, "test/vmi/entity/order/order.json")
	orderModel, err := remoteProvider.GetEntityModel(orderObject, true)
	if err != nil {
		t.Fatalf("GetEntityModel(order) failed: %v", err)
	}

	filter, err := remoteProvider.GetModelFilter(orderModel)
	if err != nil {
		t.Fatalf("GetModelFilter(order) failed: %v", err)
	}
	objectFilter, ok := filter.(*remote.ObjectFilter)
	if !ok {
		t.Fatalf("expected *remote.ObjectFilter, got %T", filter)
	}

	goodsSlice := &remote.SliceObjectValue{
		Name:    "goodsItem",
		PkgPath: "/vmi/order",
		Values: []*remote.ObjectValue{
			{
				Name:    "goodsItem",
				PkgPath: "/vmi/order",
				Fields: []*remote.FieldValue{
					{Name: "id", Value: int64(501)},
				},
			},
			{
				Name:    "goodsItem",
				PkgPath: "/vmi/order",
				Fields: []*remote.FieldValue{
					{Name: "id", Value: int64(502)},
				},
			},
		},
	}
	statusValue := &remote.ObjectValue{
		Name:    "status",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(9)},
		},
	}

	builder := NewBuilder(remoteProvider, codec.New(remoteProvider, "tenant"))
	return remoteProvider, builder, orderModel.(*remote.Object), objectFilter, goodsSlice, statusValue
}

func buildVMIProductValueModel(t *testing.T) (provider.Provider, *Builder, *remote.Object) {
	t.Helper()

	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerAllVMIRemoteModels(t, remoteProvider)

	productValue := &remote.ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
			{Name: "name", Value: "apple"},
			{Name: "description", Value: "fresh apple"},
			{Name: "image", Value: []string{"main.png", "detail.png"}},
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
			{Name: "tags", Value: []string{"fruit", "fresh"}},
		},
	}
	productModel, err := remoteProvider.GetEntityModel(productValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(productValue) failed: %v", err)
	}

	builder := NewBuilder(remoteProvider, codec.New(remoteProvider, "tenant"))
	productObject, ok := productModel.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", productModel)
	}
	return remoteProvider, builder, productObject
}

func TestBuilderVMIBuildQueryRelationFilters(t *testing.T) {
	_, builder, orderModel, filter, goodsSlice, statusValue := buildVMIOrderFilter(t)

	if err := filter.In("goods", goodsSlice); err != nil {
		t.Fatalf("filter.In(goods) failed: %v", err)
	}
	if err := filter.Equal("status", statusValue); err != nil {
		t.Fatalf("filter.Equal(status) failed: %v", err)
	}

	result, err := builder.BuildQuery(orderModel, filter)
	if err != nil {
		t.Fatalf("BuildQuery failed: %v", err)
	}

	sql := result.SQL()
	if !strings.Contains(sql, `FROM "tenant_Order"`) {
		t.Fatalf("unexpected query sql: %s", sql)
	}
	if !strings.Contains(sql, `"id" IN (SELECT DISTINCT("left") "id"  FROM "tenant_OrderGoods2GoodsItem" WHERE "right" IN ($1,$2))`) {
		t.Fatalf("missing goods relation filter in sql: %s", sql)
	}
	if !strings.Contains(sql, `"id" IN (SELECT DISTINCT("left") "id"  FROM "tenant_OrderStatus3Status" WHERE "right" = $3)`) {
		t.Fatalf("missing status relation filter in sql: %s", sql)
	}

	wantArgs := []any{int64(501), int64(502), int64(9)}
	if !reflect.DeepEqual(result.Args(), wantArgs) {
		t.Fatalf("unexpected query args: got=%#v want=%#v", result.Args(), wantArgs)
	}
}

func TestBuilderVMIBuildBatchQueryRelation(t *testing.T) {
	_, builder, orderModel, _, _, _ := buildVMIOrderFilter(t)

	result, err := builder.BuildBatchQueryRelation(orderModel, orderModel.GetField("goods"), []any{int64(101), int64(102)})
	if err != nil {
		t.Fatalf("BuildBatchQueryRelation failed: %v", err)
	}

	sql := result.SQL()
	if !strings.Contains(sql, `SELECT "left","right" FROM "tenant_OrderGoods2GoodsItem" WHERE "left" IN ($1,$2)`) {
		t.Fatalf("unexpected batch relation sql: %s", sql)
	}

	wantArgs := []any{int64(101), int64(102)}
	if !reflect.DeepEqual(result.Args(), wantArgs) {
		t.Fatalf("unexpected batch relation args: got=%#v want=%#v", result.Args(), wantArgs)
	}
}

func TestBuilderVMINotInRelationFilter(t *testing.T) {
	_, builder, orderModel, filter, goodsSlice, _ := buildVMIOrderFilter(t)

	if err := filter.NotIn("goods", goodsSlice); err != nil {
		t.Fatalf("filter.NotIn(goods) failed: %v", err)
	}

	result, err := builder.BuildQuery(orderModel, filter)
	if err != nil {
		t.Fatalf("BuildQuery failed: %v", err)
	}

	sql := result.SQL()
	if !strings.Contains(sql, `"id" IN (SELECT DISTINCT("left") "id"  FROM "tenant_OrderGoods2GoodsItem" WHERE "right" NOT IN ($1,$2))`) {
		t.Fatalf("missing goods not-in relation filter in sql: %s", sql)
	}

	wantArgs := []any{int64(501), int64(502)}
	if !reflect.DeepEqual(result.Args(), wantArgs) {
		t.Fatalf("unexpected query args: got=%#v want=%#v", result.Args(), wantArgs)
	}
}

func TestBuilderVMIBuildCountRelationFilters(t *testing.T) {
	_, builder, orderModel, filter, goodsSlice, statusValue := buildVMIOrderFilter(t)

	if err := filter.In("goods", goodsSlice); err != nil {
		t.Fatalf("filter.In(goods) failed: %v", err)
	}
	if err := filter.Equal("status", statusValue); err != nil {
		t.Fatalf("filter.Equal(status) failed: %v", err)
	}

	result, err := builder.BuildCount(orderModel, filter)
	if err != nil {
		t.Fatalf("BuildCount failed: %v", err)
	}

	sql := result.SQL()
	if !strings.Contains(sql, `SELECT COUNT(*) FROM "tenant_Order"`) {
		t.Fatalf("unexpected count sql: %s", sql)
	}
	if !strings.Contains(sql, `"tenant_OrderGoods2GoodsItem"`) || !strings.Contains(sql, `"tenant_OrderStatus3Status"`) {
		t.Fatalf("missing relation count filters in sql: %s", sql)
	}

	wantArgs := []any{int64(501), int64(502), int64(9)}
	if !reflect.DeepEqual(result.Args(), wantArgs) {
		t.Fatalf("unexpected count args: got=%#v want=%#v", result.Args(), wantArgs)
	}
}

func TestBuilderVMIQueryAndDeleteRelations(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerAllVMIRemoteModels(t, remoteProvider)
	orderModel, err := remoteProvider.GetEntityModel(&remote.ObjectValue{
		Name:    "order",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
		},
	}, true)
	if err != nil {
		t.Fatalf("GetEntityModel(orderValue) failed: %v", err)
	}
	orderObject, ok := orderModel.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", orderModel)
	}

	builder := NewBuilder(remoteProvider, codec.New(remoteProvider, "tenant"))
	if orderObject.GetField("goods") == nil || orderObject.GetField("status") == nil {
		t.Fatal("goods/status field should exist")
	}

	queryRelation, err := builder.BuildQueryRelation(orderObject, orderObject.GetField("goods"))
	if err != nil {
		t.Fatalf("BuildQueryRelation(goods) failed: %v", err)
	}
	if queryRelation.SQL() != `SELECT "right" FROM "tenant_OrderGoods2GoodsItem" WHERE "left"= $1` {
		t.Fatalf("unexpected query relation sql: %s", queryRelation.SQL())
	}
	if !reflect.DeepEqual(queryRelation.Args(), []any{int64(1001)}) {
		t.Fatalf("unexpected query relation args: %#v", queryRelation.Args())
	}

	deleteRelationByRights, err := builder.BuildDeleteRelationByRights(orderObject, orderObject.GetField("goods"), []any{int64(501), int64(502)})
	if err != nil {
		t.Fatalf("BuildDeleteRelationByRights(goods) failed: %v", err)
	}
	if deleteRelationByRights.SQL() != `DELETE FROM "tenant_OrderGoods2GoodsItem" WHERE "left"=$1 AND "right" IN ($2,$3)` {
		t.Fatalf("unexpected delete relation by rights sql: %s", deleteRelationByRights.SQL())
	}
	if !reflect.DeepEqual(deleteRelationByRights.Args(), []any{int64(1001), int64(501), int64(502)}) {
		t.Fatalf("unexpected delete relation by rights args: %#v", deleteRelationByRights.Args())
	}

	delHost, delRelation, err := builder.BuildDeleteRelation(orderObject, orderObject.GetField("status"))
	if err != nil {
		t.Fatalf("BuildDeleteRelation(status) failed: %v", err)
	}
	if delHost.SQL() != `DELETE FROM "tenant_Status" WHERE "id" IN (SELECT "right" FROM "tenant_OrderStatus3Status" WHERE "left"=$1)` {
		t.Fatalf("unexpected delete host sql: %s", delHost.SQL())
	}
	if delRelation.SQL() != `DELETE FROM "tenant_OrderStatus3Status" WHERE "left"=$1` {
		t.Fatalf("unexpected delete relation sql: %s", delRelation.SQL())
	}
	if !reflect.DeepEqual(delHost.Args(), []any{int64(1001)}) || !reflect.DeepEqual(delRelation.Args(), []any{int64(1001)}) {
		t.Fatalf("unexpected delete relation args: host=%#v relation=%#v", delHost.Args(), delRelation.Args())
	}
}

func TestBuilderVMICreateInsertAndDropRelations(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerAllVMIRemoteModels(t, remoteProvider)
	orderModel, err := remoteProvider.GetEntityModel(&remote.ObjectValue{
		Name:    "order",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(1001)},
		},
	}, true)
	if err != nil {
		t.Fatalf("GetEntityModel(orderValue) failed: %v", err)
	}
	orderObject := orderModel.(*remote.Object)
	builder := NewBuilder(remoteProvider, codec.New(remoteProvider, "tenant"))

	createRelation, err := builder.BuildCreateRelationTable(orderObject, orderObject.GetField("goods"))
	if err != nil {
		t.Fatalf("BuildCreateRelationTable(goods) failed: %v", err)
	}
	if !strings.Contains(createRelation.SQL(), `CREATE TABLE IF NOT EXISTS "tenant_OrderGoods2GoodsItem"`) {
		t.Fatalf("unexpected create relation sql: %s", createRelation.SQL())
	}
	if !strings.Contains(createRelation.SQL(), `"left" BIGINT NOT NULL`) || !strings.Contains(createRelation.SQL(), `"right" BIGINT NOT NULL`) {
		t.Fatalf("unexpected create relation column types: %s", createRelation.SQL())
	}

	goodsItemModel, err := remoteProvider.GetEntityModel(&remote.ObjectValue{
		Name:    "goodsItem",
		PkgPath: "/vmi/order",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(909)},
		},
	}, true)
	if err != nil {
		t.Fatalf("GetEntityModel(goodsItemValue) failed: %v", err)
	}

	insertRelation, err := builder.BuildInsertRelation(orderObject, orderObject.GetField("goods"), goodsItemModel)
	if err != nil {
		t.Fatalf("BuildInsertRelation(goods) failed: %v", err)
	}
	if insertRelation.SQL() != `INSERT INTO "tenant_OrderGoods2GoodsItem" ("left", "right") VALUES ($1,$2) RETURNING id` {
		t.Fatalf("unexpected insert relation sql: %s", insertRelation.SQL())
	}
	if !reflect.DeepEqual(insertRelation.Args(), []any{int64(1001), int64(909)}) {
		t.Fatalf("unexpected insert relation args: %#v", insertRelation.Args())
	}

	dropRelation, err := builder.BuildDropRelationTable(orderObject, orderObject.GetField("goods"))
	if err != nil {
		t.Fatalf("BuildDropRelationTable(goods) failed: %v", err)
	}
	if dropRelation.SQL() != "DROP INDEX IF EXISTS \"tenant_OrderGoods2GoodsItem_index\";\nDROP TABLE IF EXISTS \"tenant_OrderGoods2GoodsItem\"" {
		t.Fatalf("unexpected drop relation sql: %s", dropRelation.SQL())
	}
}

func TestBuilderVMIMainTableSQL(t *testing.T) {
	_, builder, productModel := buildVMIProductValueModel(t)

	createTable, err := builder.BuildCreateTable(productModel)
	if err != nil {
		t.Fatalf("BuildCreateTable(product) failed: %v", err)
	}
	createSQL := createTable.SQL()
	if !strings.Contains(createSQL, `CREATE TABLE IF NOT EXISTS "tenant_Product"`) {
		t.Fatalf("unexpected create table sql: %s", createSQL)
	}
	if !strings.Contains(createSQL, `"id" BIGSERIAL NOT NULL`) {
		t.Fatalf("create table should use BIGSERIAL primary key: %s", createSQL)
	}
	if !strings.Contains(createSQL, `"image" TEXT NOT NULL`) || !strings.Contains(createSQL, `"tags" TEXT NOT NULL`) {
		t.Fatalf("create table should include basic slice columns: %s", createSQL)
	}
	if !strings.Contains(createSQL, `"creater" BIGINT NOT NULL DEFAULT '0'`) || !strings.Contains(createSQL, `"createTime" BIGINT NOT NULL DEFAULT '0'`) {
		t.Fatalf("create table should materialize numeric reference defaults: %s", createSQL)
	}
	if strings.Contains(createSQL, `"status"`) {
		t.Fatalf("create table should not include relation fields: %s", createSQL)
	}

	insertResult, err := builder.BuildInsert(productModel)
	if err != nil {
		t.Fatalf("BuildInsert(product) failed: %v", err)
	}
	insertSQL := insertResult.SQL()
	if !strings.Contains(insertSQL, `INSERT INTO "tenant_Product" ("name","description","image","expire","tags","creater","createTime","modifyTime","namespace") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`) {
		t.Fatalf("unexpected insert sql: %s", insertSQL)
	}
	wantInsertArgs := []any{"apple", "fresh apple", `["main.png","detail.png"]`, 30, `["fruit","fresh"]`, int64(0), int64(0), int64(0), ""}
	if !reflect.DeepEqual(insertResult.Args(), wantInsertArgs) {
		t.Fatalf("unexpected insert args: got=%#v want=%#v", insertResult.Args(), wantInsertArgs)
	}

	updateResult, err := builder.BuildUpdate(productModel)
	if err != nil {
		t.Fatalf("BuildUpdate(product) failed: %v", err)
	}
	updateSQL := updateResult.SQL()
	if !strings.Contains(updateSQL, `UPDATE "tenant_Product" SET "name" = $1,"description" = $2,"image" = $3,"expire" = $4,"tags" = $5 WHERE "id" = $6`) {
		t.Fatalf("unexpected update sql: %s", updateSQL)
	}
	wantUpdateArgs := []any{"apple", "fresh apple", `["main.png","detail.png"]`, 30, `["fruit","fresh"]`, int64(1001)}
	if !reflect.DeepEqual(updateResult.Args(), wantUpdateArgs) {
		t.Fatalf("unexpected update args: got=%#v want=%#v", updateResult.Args(), wantUpdateArgs)
	}

	deleteResult, err := builder.BuildDelete(productModel)
	if err != nil {
		t.Fatalf("BuildDelete(product) failed: %v", err)
	}
	if deleteResult.SQL() != `DELETE FROM "tenant_Product" WHERE "id" = $1` {
		t.Fatalf("unexpected delete sql: %s", deleteResult.SQL())
	}
	if !reflect.DeepEqual(deleteResult.Args(), []any{int64(1001)}) {
		t.Fatalf("unexpected delete args: %#v", deleteResult.Args())
	}
}

func TestBuilderVMIUpdateRequiresWritableFields(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerAllVMIRemoteModels(t, remoteProvider)

	productModel, err := remoteProvider.GetEntityModel(loadVMIRemoteObject(t, "test/vmi/entity/product/product.json"), true)
	if err != nil {
		t.Fatalf("GetEntityModel(product) failed: %v", err)
	}

	productObject, ok := productModel.(*remote.Object)
	if !ok {
		t.Fatalf("expected *remote.Object, got %T", productModel)
	}
	for _, field := range productObject.GetFields() {
		if models.IsPrimaryField(field) || !models.IsBasicField(field) {
			continue
		}
		field.Reset()
	}
	if err := productObject.SetPrimaryFieldValue(int64(1001)); err != nil {
		t.Fatalf("SetPrimaryFieldValue failed: %v", err)
	}
	if err := productObject.SetFieldValue("status", &remote.ObjectValue{
		Name:    "status",
		PkgPath: "/vmi",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: int64(9)},
		},
	}); err != nil {
		t.Fatalf("SetFieldValue(status) failed: %v", err)
	}

	builder := NewBuilder(remoteProvider, codec.New(remoteProvider, "tenant"))
	_, err = builder.BuildUpdate(productObject)
	if err == nil {
		t.Fatalf("BuildUpdate should fail when no writable basic fields are assigned")
	}
	if err.Code != cd.IllegalParam {
		t.Fatalf("BuildUpdate should return IllegalParam, got: %v", err)
	}
}

func TestBuilderVMIUpdateUsesAssignedBasicFieldsOnly(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	registerAllVMIRemoteModels(t, remoteProvider)

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

	productModel, err := remoteProvider.GetEntityModel(productValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel(product update value) failed: %v", err)
	}

	builder := NewBuilder(remoteProvider, codec.New(remoteProvider, "tenant"))
	updateResult, err := builder.BuildUpdate(productModel)
	if err != nil {
		t.Fatalf("BuildUpdate(partial product) failed: %v", err)
	}

	updateSQL := updateResult.SQL()
	if !strings.Contains(updateSQL, `UPDATE "tenant_Product" SET "name" = $1 WHERE "id" = $2`) {
		t.Fatalf("unexpected partial update sql: %s", updateSQL)
	}
	wantUpdateArgs := []any{"apple-updated", int64(1001)}
	if !reflect.DeepEqual(updateResult.Args(), wantUpdateArgs) {
		t.Fatalf("unexpected partial update args: got=%#v want=%#v", updateResult.Args(), wantUpdateArgs)
	}
}
