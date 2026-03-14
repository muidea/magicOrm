package postgres

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
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

func buildVMIProductFilter(t *testing.T) (provider.Provider, *Builder, *remote.Object, *remote.ObjectFilter, *remote.SliceObjectValue, *remote.ObjectValue) {
	t.Helper()

	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	for _, path := range []string{
		"test/vmi/entity/status.json",
		"test/vmi/entity/product/skuinfo.json",
		"test/vmi/entity/product/product.json",
	} {
		if _, err := remoteProvider.RegisterModel(loadVMIRemoteObject(t, path)); err != nil {
			t.Fatalf("RegisterModel(%s) failed: %v", path, err)
		}
	}

	productObject := loadVMIRemoteObject(t, "test/vmi/entity/product/product.json")
	productModel, err := remoteProvider.GetEntityModel(productObject, true)
	if err != nil {
		t.Fatalf("GetEntityModel(product) failed: %v", err)
	}

	filter, err := remoteProvider.GetModelFilter(productModel)
	if err != nil {
		t.Fatalf("GetModelFilter(product) failed: %v", err)
	}
	objectFilter, ok := filter.(*remote.ObjectFilter)
	if !ok {
		t.Fatalf("expected *remote.ObjectFilter, got %T", filter)
	}

	skuInfoSlice := &remote.SliceObjectValue{
		Name:    "skuInfo",
		PkgPath: "/vmi/product",
		Values: []*remote.ObjectValue{
			{
				Name:    "skuInfo",
				PkgPath: "/vmi/product",
				Fields: []*remote.FieldValue{
					{Name: "sku", Value: "sku-001"},
				},
			},
			{
				Name:    "skuInfo",
				PkgPath: "/vmi/product",
				Fields: []*remote.FieldValue{
					{Name: "sku", Value: "sku-002"},
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
	return remoteProvider, builder, productModel.(*remote.Object), objectFilter, skuInfoSlice, statusValue
}

func buildVMIProductValueModel(t *testing.T) (provider.Provider, *Builder, *remote.Object) {
	t.Helper()

	remoteProvider := provider.NewRemoteProvider("tenant", nil)
	for _, path := range []string{
		"test/vmi/entity/status.json",
		"test/vmi/entity/product/skuinfo.json",
		"test/vmi/entity/product/product.json",
	} {
		if _, err := remoteProvider.RegisterModel(loadVMIRemoteObject(t, path)); err != nil {
			t.Fatalf("RegisterModel(%s) failed: %v", path, err)
		}
	}

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
	_, builder, productModel, filter, skuInfoSlice, statusValue := buildVMIProductFilter(t)

	if err := filter.In("skuInfo", skuInfoSlice); err != nil {
		t.Fatalf("filter.In(skuInfo) failed: %v", err)
	}
	if err := filter.Equal("status", statusValue); err != nil {
		t.Fatalf("filter.Equal(status) failed: %v", err)
	}

	result, err := builder.BuildQuery(productModel, filter)
	if err != nil {
		t.Fatalf("BuildQuery failed: %v", err)
	}

	sql := result.SQL()
	if !strings.Contains(sql, `FROM "tenant_Product"`) {
		t.Fatalf("unexpected query sql: %s", sql)
	}
	if !strings.Contains(sql, `"id" IN (SELECT DISTINCT("left") "id"  FROM "tenant_ProductSkuInfo2SkuInfo" WHERE "right" IN ($1,$2))`) {
		t.Fatalf("missing skuInfo relation filter in sql: %s", sql)
	}
	if !strings.Contains(sql, `"id" IN (SELECT DISTINCT("left") "id"  FROM "tenant_ProductStatus3Status" WHERE "right" = $3)`) {
		t.Fatalf("missing status relation filter in sql: %s", sql)
	}

	wantArgs := []any{"sku-001", "sku-002", int64(9)}
	if !reflect.DeepEqual(result.Args(), wantArgs) {
		t.Fatalf("unexpected query args: got=%#v want=%#v", result.Args(), wantArgs)
	}
}

func TestBuilderVMINotInRelationFilter(t *testing.T) {
	_, builder, productModel, filter, skuInfoSlice, _ := buildVMIProductFilter(t)

	if err := filter.NotIn("skuInfo", skuInfoSlice); err != nil {
		t.Fatalf("filter.NotIn(skuInfo) failed: %v", err)
	}

	result, err := builder.BuildQuery(productModel, filter)
	if err != nil {
		t.Fatalf("BuildQuery failed: %v", err)
	}

	sql := result.SQL()
	if !strings.Contains(sql, `"id" IN (SELECT DISTINCT("left") "id"  FROM "tenant_ProductSkuInfo2SkuInfo" WHERE "right" NOT IN ($1,$2))`) {
		t.Fatalf("missing skuInfo not-in relation filter in sql: %s", sql)
	}

	wantArgs := []any{"sku-001", "sku-002"}
	if !reflect.DeepEqual(result.Args(), wantArgs) {
		t.Fatalf("unexpected query args: got=%#v want=%#v", result.Args(), wantArgs)
	}
}

func TestBuilderVMIBuildCountRelationFilters(t *testing.T) {
	_, builder, productModel, filter, skuInfoSlice, statusValue := buildVMIProductFilter(t)

	if err := filter.In("skuInfo", skuInfoSlice); err != nil {
		t.Fatalf("filter.In(skuInfo) failed: %v", err)
	}
	if err := filter.Equal("status", statusValue); err != nil {
		t.Fatalf("filter.Equal(status) failed: %v", err)
	}

	result, err := builder.BuildCount(productModel, filter)
	if err != nil {
		t.Fatalf("BuildCount failed: %v", err)
	}

	sql := result.SQL()
	if !strings.Contains(sql, `SELECT COUNT(*) FROM "tenant_Product"`) {
		t.Fatalf("unexpected count sql: %s", sql)
	}
	if !strings.Contains(sql, `"tenant_ProductSkuInfo2SkuInfo"`) || !strings.Contains(sql, `"tenant_ProductStatus3Status"`) {
		t.Fatalf("missing relation count filters in sql: %s", sql)
	}

	wantArgs := []any{"sku-001", "sku-002", int64(9)}
	if !reflect.DeepEqual(result.Args(), wantArgs) {
		t.Fatalf("unexpected count args: got=%#v want=%#v", result.Args(), wantArgs)
	}
}

func TestBuilderVMIQueryAndDeleteRelations(t *testing.T) {
	_, builder, productModel := buildVMIProductValueModel(t)
	if productModel.GetField("skuInfo") == nil || productModel.GetField("status") == nil {
		t.Fatal("skuInfo/status field should exist")
	}

	queryRelation, err := builder.BuildQueryRelation(productModel, productModel.GetField("skuInfo"))
	if err != nil {
		t.Fatalf("BuildQueryRelation(skuInfo) failed: %v", err)
	}
	if queryRelation.SQL() != `SELECT "right" FROM "tenant_ProductSkuInfo2SkuInfo" WHERE "left"= $1` {
		t.Fatalf("unexpected query relation sql: %s", queryRelation.SQL())
	}
	if !reflect.DeepEqual(queryRelation.Args(), []any{int64(1001)}) {
		t.Fatalf("unexpected query relation args: %#v", queryRelation.Args())
	}

	deleteRelationByRights, err := builder.BuildDeleteRelationByRights(productModel, productModel.GetField("skuInfo"), []any{"sku-001", "sku-002"})
	if err != nil {
		t.Fatalf("BuildDeleteRelationByRights(skuInfo) failed: %v", err)
	}
	if deleteRelationByRights.SQL() != `DELETE FROM "tenant_ProductSkuInfo2SkuInfo" WHERE "left"=$1 AND "right" IN ($2,$3)` {
		t.Fatalf("unexpected delete relation by rights sql: %s", deleteRelationByRights.SQL())
	}
	if !reflect.DeepEqual(deleteRelationByRights.Args(), []any{int64(1001), "sku-001", "sku-002"}) {
		t.Fatalf("unexpected delete relation by rights args: %#v", deleteRelationByRights.Args())
	}

	delHost, delRelation, err := builder.BuildDeleteRelation(productModel, productModel.GetField("status"))
	if err != nil {
		t.Fatalf("BuildDeleteRelation(status) failed: %v", err)
	}
	if delHost.SQL() != `DELETE FROM "tenant_Status" WHERE "id" IN (SELECT "right" FROM "tenant_ProductStatus3Status" WHERE "left"=$1)` {
		t.Fatalf("unexpected delete host sql: %s", delHost.SQL())
	}
	if delRelation.SQL() != `DELETE FROM "tenant_ProductStatus3Status" WHERE "left"=$1` {
		t.Fatalf("unexpected delete relation sql: %s", delRelation.SQL())
	}
	if !reflect.DeepEqual(delHost.Args(), []any{int64(1001)}) || !reflect.DeepEqual(delRelation.Args(), []any{int64(1001)}) {
		t.Fatalf("unexpected delete relation args: host=%#v relation=%#v", delHost.Args(), delRelation.Args())
	}
}

func TestBuilderVMICreateInsertAndDropRelations(t *testing.T) {
	remoteProvider, builder, productModel := buildVMIProductValueModel(t)

	createRelation, err := builder.BuildCreateRelationTable(productModel, productModel.GetField("skuInfo"))
	if err != nil {
		t.Fatalf("BuildCreateRelationTable(skuInfo) failed: %v", err)
	}
	if !strings.Contains(createRelation.SQL(), `CREATE TABLE IF NOT EXISTS "tenant_ProductSkuInfo2SkuInfo"`) {
		t.Fatalf("unexpected create relation sql: %s", createRelation.SQL())
	}
	if !strings.Contains(createRelation.SQL(), `"left" BIGINT NOT NULL`) || !strings.Contains(createRelation.SQL(), `"right" VARCHAR(32) NOT NULL`) {
		t.Fatalf("unexpected create relation column types: %s", createRelation.SQL())
	}

	skuInfoModel, err := remoteProvider.GetEntityModel(&remote.ObjectValue{
		Name:    "skuInfo",
		PkgPath: "/vmi/product",
		Fields: []*remote.FieldValue{
			{Name: "sku", Value: "sku-009"},
		},
	}, true)
	if err != nil {
		t.Fatalf("GetEntityModel(skuInfoValue) failed: %v", err)
	}

	insertRelation, err := builder.BuildInsertRelation(productModel, productModel.GetField("skuInfo"), skuInfoModel)
	if err != nil {
		t.Fatalf("BuildInsertRelation(skuInfo) failed: %v", err)
	}
	if insertRelation.SQL() != `INSERT INTO "tenant_ProductSkuInfo2SkuInfo" ("left", "right") VALUES ($1,$2) RETURNING id` {
		t.Fatalf("unexpected insert relation sql: %s", insertRelation.SQL())
	}
	if !reflect.DeepEqual(insertRelation.Args(), []any{int64(1001), "sku-009"}) {
		t.Fatalf("unexpected insert relation args: %#v", insertRelation.Args())
	}

	dropRelation, err := builder.BuildDropRelationTable(productModel, productModel.GetField("skuInfo"))
	if err != nil {
		t.Fatalf("BuildDropRelationTable(skuInfo) failed: %v", err)
	}
	if dropRelation.SQL() != "DROP INDEX IF EXISTS \"tenant_ProductSkuInfo2SkuInfo_index\";\nDROP TABLE IF EXISTS \"tenant_ProductSkuInfo2SkuInfo\"" {
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
	if strings.Contains(createSQL, `"status"`) || strings.Contains(createSQL, `"skuInfo"`) {
		t.Fatalf("create table should not include relation fields: %s", createSQL)
	}

	insertResult, err := builder.BuildInsert(productModel)
	if err != nil {
		t.Fatalf("BuildInsert(product) failed: %v", err)
	}
	insertSQL := insertResult.SQL()
	if !strings.Contains(insertSQL, `INSERT INTO "tenant_Product" ("name","description","image","expire","tags","creater","createTime","namespace") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`) {
		t.Fatalf("unexpected insert sql: %s", insertSQL)
	}
	wantInsertArgs := []any{"apple", "fresh apple", `["main.png","detail.png"]`, 30, `["fruit","fresh"]`, int64(0), int64(0), ""}
	if !reflect.DeepEqual(insertResult.Args(), wantInsertArgs) {
		t.Fatalf("unexpected insert args: got=%#v want=%#v", insertResult.Args(), wantInsertArgs)
	}

	updateResult, err := builder.BuildUpdate(productModel)
	if err != nil {
		t.Fatalf("BuildUpdate(product) failed: %v", err)
	}
	updateSQL := updateResult.SQL()
	if !strings.Contains(updateSQL, `UPDATE "tenant_Product" SET "name" = $1,"description" = $2,"image" = $3,"expire" = $4,"tags" = $5,"creater" = $6,"createTime" = $7,"namespace" = $8 WHERE "id" = $9`) {
		t.Fatalf("unexpected update sql: %s", updateSQL)
	}
	wantUpdateArgs := []any{"apple", "fresh apple", `["main.png","detail.png"]`, 30, `["fruit","fresh"]`, int64(0), int64(0), "", int64(1001)}
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
