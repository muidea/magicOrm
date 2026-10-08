package test

import (
	"os"
	"strings"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

type NullRegressionTarget struct {
	ID int64 `orm:"id key" view:"detail,lite"`
}

type NullRegressionHost struct {
	ID   int64                 `orm:"id key auto" view:"detail,lite" constraint:"req,ro"`
	Name string                `orm:"name" view:"detail,lite"`
	Ref  *NullRegressionTarget `orm:"ref" view:"detail,lite"`
}

type NullRegressionRequired struct {
	ID  int64                 `orm:"id key auto" view:"detail,lite" constraint:"req,ro"`
	Ref *NullRegressionTarget `orm:"ref" view:"detail,lite" constraint:"req"`
}

// The explicit server prevents this regression from using a developer's database.
// Every schema and data operation below goes through ORM contracts.
func TestNullReferenceDatabaseContract(t *testing.T) {
	server := os.Getenv("MAGICORM_RELATION_TEST_SERVER")
	if server == "" {
		t.Skip("set MAGICORM_RELATION_TEST_SERVER to an isolated test database")
	}
	orm.Initialize()
	defer orm.Uninitialized()
	for _, isRemote := range []bool{false, true} {
		name := "local"
		setValue := local.SetModelValue
		if isRemote {
			name, setValue = "remote", remote.SetModelValue
		}
		t.Run(name, func(t *testing.T) {
			fault := ""
			option := provider.WithSetModelValueFunc(func(m models.Model, v models.Value, disable bool) (models.Model, *cd.Error) {
				if fault != "" && m.GetName() == "NullRegressionTarget" {
					if fault == "panic" {
						panic("related conversion fault")
					}
					return nil, cd.NewError(cd.Unexpected, "related conversion fault")
				}
				return setValue(m, v, disable)
			})
			p := provider.NewLocalProviderWithOptions("null-regression-"+name, option)
			if isRemote {
				p = provider.NewRemoteProviderWithOptions("null-regression-"+name, option)
			}
			db, err := orm.NewOrm(p, orm.NewConfig(server, "testdb", "public", os.Getenv("MAGICORM_RELATION_TEST_USER"), "rootkit"), "nullreg_"+name)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Release()
			definitions := []any{&NullRegressionTarget{}, &NullRegressionHost{}, &NullRegressionRequired{}}
			var registered []models.Model
			for _, entity := range definitions {
				if isRemote {
					entity, err = helper.GetObject(entity)
					if err != nil {
						t.Fatal(err)
					}
				}
				m, e := p.RegisterModel(entity)
				if e != nil {
					t.Fatal(e)
				}
				registered = append(registered, m)
				if e = db.Create(m); e != nil {
					t.Fatal(e)
				}
				defer func() {
					if e := db.Drop(m); e != nil {
						t.Errorf("drop: %v", e)
					}
				}()
			}
			model := func(v any) models.Model {
				t.Helper()
				if isRemote {
					v, err = helper.GetObjectValue(v)
					if err != nil {
						t.Fatal(err)
					}
				}
				m, e := p.GetEntityModel(v, true)
				if e != nil {
					t.Fatal(e)
				}
				return m
			}
			count := func(m models.Model) int {
				t.Helper()
				filter, e := p.GetModelFilter(m)
				if e != nil {
					t.Fatal(e)
				}
				rows, e := db.BatchQuery(filter)
				if e != nil {
					t.Fatal(e)
				}
				return len(rows)
			}
			query := func(id any) models.Model {
				t.Helper()
				m := model(&NullRegressionHost{})
				if e := m.SetFieldValue("id", id); e != nil {
					t.Fatal(e)
				}
				m, e := db.Query(m)
				if e != nil {
					t.Fatal(e)
				}
				return m
			}
			if _, err = db.Insert(model(&NullRegressionTarget{ID: 7})); err != nil {
				t.Fatal(err)
			}
			for _, mode := range []string{"omitted", "null", "bound"} {
				m := model(&NullRegressionHost{Name: mode})
				if isRemote && mode == "omitted" {
					// Omission differs from explicit JSON null in a remote request.
					v, e := helper.GetObjectValue(&NullRegressionHost{})
					if e != nil {
						t.Fatal(e)
					}
					v.Fields = []*remote.FieldValue{{Name: "name", Value: mode, Assigned: true}}
					m, e = p.GetEntityModel(v, true)
					if e != nil {
						t.Fatal(e)
					}
				}
				if mode == "null" {
					if e := m.SetFieldValue("ref", nil); e != nil {
						t.Fatal(e)
					}
				}
				if mode == "bound" {
					m = model(&NullRegressionHost{Name: mode, Ref: &NullRegressionTarget{ID: 7}})
				}
				inserted, e := db.Insert(m)
				if e != nil {
					t.Fatalf("%s insert: %v", mode, e)
				}
				id := inserted.GetPrimaryField().GetValue().Get()
				read := query(id)
				if models.IsValidField(read.GetField("ref")) != (mode == "bound") {
					t.Fatalf("%s wrong relation: %#v", mode, read.Interface(true))
				}
				if mode == "bound" {
					update := model(&NullRegressionHost{})
					if e = update.SetFieldValue("id", id); e != nil {
						t.Fatal(e)
					}
					if e = update.SetFieldValue("ref", nil); e != nil {
						t.Fatal(e)
					}
					if _, e = db.Update(update); e != nil {
						t.Fatal(e)
					}
					if models.IsValidField(query(id).GetField("ref")) {
						t.Fatal("clear did not persist")
					}
				}
			}
			if got := count(registered[1]); got != 3 {
				t.Fatalf("host count %d", got)
			}
			for _, mode := range []string{"omitted", "null"} {
				m := model(&NullRegressionRequired{})
				if mode == "null" {
					if e := m.SetFieldValue("ref", nil); e != nil {
						t.Fatal(e)
					}
				}
				if _, e := db.Insert(m); e == nil || !strings.Contains(e.Message, "ref") {
					t.Fatalf("required %s: %v", mode, e)
				}
				if got := count(registered[2]); got != 0 {
					t.Fatalf("required host residue %d", got)
				}
			}
			if _, e := db.Insert(model(&NullRegressionRequired{Ref: &NullRegressionTarget{ID: 7}})); e != nil {
				t.Fatalf("required bound reference: %v", e)
			}
			if got := count(registered[2]); got != 1 {
				t.Fatalf("required bound host count %d", got)
			}
			for _, mode := range []string{"error", "panic"} {
				m := model(&NullRegressionHost{Name: mode, Ref: &NullRegressionTarget{ID: 7}})
				fault = mode
				func() {
					if mode == "panic" {
						defer func() {
							if got := recover(); got != "related conversion fault" {
								t.Errorf("panic: %v", got)
							}
						}()
					}
					_, e := db.Insert(m)
					if mode == "error" && e == nil {
						t.Error("relation failure ignored")
					}
				}()
				fault = ""
				if got := count(registered[1]); got != 3 {
					t.Fatalf("%s host residue %d", mode, got)
				}
				if got := count(registered[0]); got != 1 {
					t.Fatalf("%s changed target: %d", mode, got)
				}
			}
		})
	}
}
