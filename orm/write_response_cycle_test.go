package orm

import (
	"reflect"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

type responseCycleNode struct {
	ID       int                  `orm:"id key" view:"detail,lite"`
	Next     *responseCycleNode   `orm:"next" view:"detail,lite"`
	Children []*responseCycleNode `orm:"children" view:"detail,lite"`
}

func TestWriteResponseRejectsPointerCyclesBeforeCopy(t *testing.T) {
	for _, providerKind := range []string{"local", "remote"} {
		for _, mode := range []string{"direct", "two-nodes", "slice-cycle", "shared", "chain", "nil"} {
			t.Run(providerKind+"/"+mode, func(t *testing.T) {
				root := &responseCycleNode{ID: 1}
				child := &responseCycleNode{ID: 2}
				cyclic := true
				switch mode {
				case "direct":
					root.Next = root
				case "two-nodes":
					root.Next, child.Next = child, root
				case "slice-cycle":
					root.Children = []*responseCycleNode{root}
				case "shared":
					root.Next, root.Children = child, []*responseCycleNode{child, child}
					cyclic = false
				case "chain":
					root.Next, child.Next = child, &responseCycleNode{ID: 3}
					cyclic = false
				case "nil":
					cyclic = false
				}
				var p provider.Provider
				var model models.Model
				var err *cd.Error
				if providerKind == "local" {
					p = provider.NewLocalProvider("response-cycle", nil)
					if _, err = p.RegisterModel(&responseCycleNode{}); err == nil {
						model, err = local.GetEntityModel(root, nil)
					}
				} else {
					p = provider.NewRemoteProvider("response-cycle", nil)
					var object *remote.Object
					object, err = helper.GetObject(&responseCycleNode{})
					if err == nil {
						_, err = p.RegisterModel(object)
					}
					if err == nil {
						model = object
						err = model.SetFieldValue("id", 1)
					}
					if err == nil && cyclic {
						// 用有限远端元数据携带真正有环的 ObjectValue，而非循环类型定义。
						value := &remote.ObjectValue{Name: object.Name, PkgPath: object.PkgPath}
						other := &remote.ObjectValue{Name: object.Name, PkgPath: object.PkgPath}
						value.Fields = []*remote.FieldValue{{Name: "id", Value: 2}}
						switch mode {
						case "direct":
							value.Fields = append(value.Fields, &remote.FieldValue{Name: "next", Value: value})
						case "two-nodes":
							value.Fields = append(value.Fields, &remote.FieldValue{Name: "next", Value: other})
							other.Fields = []*remote.FieldValue{{Name: "id", Value: 3}, {Name: "next", Value: value}}
						case "slice-cycle":
							value.Fields = append(value.Fields, &remote.FieldValue{Name: "children", Value: &remote.SliceObjectValue{Name: object.Name, PkgPath: object.PkgPath, Values: []*remote.ObjectValue{value}}})
						}
						err = model.SetFieldValue("next", value)
					} else if err == nil {
						var value *remote.ObjectValue
						value, err = helper.GetObjectValue(root)
						if err == nil && mode == "shared" {
							shared := value.GetFieldValue("next").(*remote.ObjectValue)
							value.SetFieldValue("children", &remote.SliceObjectValue{Name: object.Name, PkgPath: object.PkgPath, Values: []*remote.ObjectValue{shared, shared}})
						}
						if err == nil {
							model, err = p.GetEntityModel(value, true)
						}
					}
				}
				if err != nil {
					t.Fatal("构造源码夹具失败", err)
				}
				response, err := projectWriteResponseModel(model, p)
				if cyclic {
					if err == nil || err.Code != cd.IllegalParam || response != nil {
						t.Fatal("有环值在复制前未被拒绝", err)
					}
				} else if err != nil || response == nil {
					t.Fatal("有限链或共享无环对象被误拒绝", err)
				}
			})
		}
	}
}

type responseIgnoredGraph struct {
	ID    int                   `orm:"id key" view:"detail,lite"`
	Extra *responseIgnoredGraph `orm:"-"`
	Data  map[any]any           `orm:"-"`
}

func TestWriteResponseRejectsCycleInCopiedMemoryFields(t *testing.T) {
	for _, mode := range []string{"pointer", "map-key", "map-value"} {
		t.Run(mode, func(t *testing.T) {
			value := &responseIgnoredGraph{ID: 1}
			switch mode {
			case "pointer":
				value.Extra = value
			case "map-key":
				value.Data = map[any]any{value: "self"}
			case "map-value":
				value.Data = map[any]any{"self": value}
			}
			model, err := local.GetEntityModel(value, nil)
			if err != nil {
				t.Fatal(err)
			}
			response, err := buildWriteResponseModel(model)
			if err == nil || err.Code != cd.IllegalParam || response != nil {
				t.Fatal("复制包含有环内存字段时未明确拒绝", err)
			}
		})
	}
}

func TestWriteResponseValueNestingLimit(t *testing.T) {
	root := &responseCycleNode{ID: 1}
	current := root
	for idx := 0; idx < 300; idx++ {
		current.Next = &responseCycleNode{ID: idx + 2}
		current = current.Next
	}
	if err := validateWriteResponseValue(reflect.ValueOf(root), map[responseValuePointer]bool{}, 0); err == nil || err.Code != cd.IllegalParam {
		t.Fatal("过深输入未明确拒绝", err)
	}
}
