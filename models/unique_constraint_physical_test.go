package models_test

import (
	"strings"
	"testing"

	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/internal/database/mysql"
	"github.com/muidea/magicOrm/internal/database/postgres"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestUniqueConstraintUsesLocalPhysicalFields(t *testing.T) {
	type target struct {
		ID int64 `orm:"id key"`
	}
	type sample struct {
		ID      int64     `orm:"id key"`
		Stamp   int64     `orm:"stamp"`
		Target  *target   `orm:"target"`
		Targets []*target `orm:"targets"`
		Values  []int64   `orm:"values"`
	}
	model, err := local.GetEntityModel(&sample{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"stamp", "target", "targets", "values"} {
		constraint := models.UniqueConstraint{Name: "uq_field", Fields: []string{field}}
		allowed := field == "stamp" || field == "values" || field == "target"
		if err := constraint.Verify(model.GetFields()); (err == nil) != allowed {
			t.Fatal("local 唯一约束与同一物理列规则不一致", field, err)
		}
	}
}

func TestUniqueConstraintRejectsNonMainTableFields(t *testing.T) {
	for _, kind := range []string{"scalar", "scalar-pointer", "scalar-slice", "reference", "owned", "references", "untyped"} {
		t.Run(kind, func(t *testing.T) {
			typ := &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}
			switch kind {
			case "scalar-pointer":
				typ.IsPtr = true
			case "scalar-slice":
				typ = &remote.TypeImpl{Name: "[]int64", Value: models.TypeSliceValue, ElemType: typ}
			case "reference", "owned", "references":
				typ = &remote.TypeImpl{Name: "Target", PkgPath: "/generic", Value: models.TypeStructValue, IsPtr: kind != "owned"}
				if kind == "references" {
					typ = &remote.TypeImpl{Name: "[]*Target", Value: models.TypeSliceValue, ElemType: typ}
				}
			case "untyped":
				typ = nil
			}
			object := &remote.Object{Name: "Sample", PkgPath: "/generic", Fields: []*remote.Field{
				{Name: "id", Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}, Spec: &remote.SpecImpl{PrimaryKey: true}},
				{Name: "stamp", Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}},
				{Name: "target", Type: typ},
			}, UniqueConstraints: []models.UniqueConstraint{{Name: "uq_target_stamp", Fields: []string{"target", "stamp"}}}}
			constraint := object.UniqueConstraints[0]
			allowed := strings.HasPrefix(kind, "scalar") || kind == "reference"
			if err := constraint.Verify(object.GetFields()); (err == nil) != allowed {
				t.Fatal("唯一约束准入与主表物理字段不一致", kind, err)
			}
			// 无类型字段不能交给 DDL 遍历；该负例只验证准入不会 panic。
			if typ == nil {
				return
			}
			if err := object.Verify(); (err == nil) != allowed {
				t.Fatal("remote 模型没有应用同一唯一约束边界", err)
			}
			p := provider.NewRemoteProvider("generic", nil)
			target := &remote.Object{Name: "Target", PkgPath: "/generic", Fields: []*remote.Field{{Name: "id", Type: &remote.TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}, Spec: &remote.SpecImpl{PrimaryKey: true}}}}
			if _, err := p.RegisterModel(target); err != nil {
				t.Fatal(err)
			}
			builders := []database.Builder{postgres.NewBuilder(p, codec.New(p, "generic")), mysql.NewBuilder(p, codec.New(p, "generic"))}
			for _, builder := range builders {
				result, err := builder.BuildCreateTable(object)
				if allowed && (err != nil || result == nil) || !allowed && (err == nil || result != nil) {
					t.Fatal("建表生成了不可用的关系唯一列", kind, err)
				}
				result, err = builder.BuildCreateUniqueConstraint(object, constraint)
				if allowed && (err != nil || result == nil) || !allowed && (err == nil || result != nil) {
					t.Fatal("补建唯一约束绕过主表列检查", kind, err)
				}
				schema, ok := builder.(database.SchemaBuilder)
				if !ok {
					t.Fatal("数据库 builder 缺少 Schema 描述合同")
				}
				declaration, err := schema.DescribeTable(object)
				if (err == nil) != allowed {
					t.Fatal("Schema 期望模型仍接受不存在的关系列", declaration, err)
				}
			}
		})
	}
}
