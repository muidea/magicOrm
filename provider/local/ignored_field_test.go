package local

import (
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/models"
)

type ignoredRelation struct{ State string }

type ignoredOpaqueModel struct {
	ID   int64    `orm:"id key"`
	Work chan int `orm:"-"`
}

type persistedOpaqueModel struct {
	ID   int64    `orm:"id key"`
	Work chan int `json:"-"`
}

type ignoredFieldModel struct {
	ID        int64           `orm:"id key" view:"detail,lite"`
	Transient ignoredRelation `orm:"-" json:"-" view:"detail,lite"`
	Name      string          `orm:"name" view:"detail,lite"`
	Counter   int64           `orm:" - " view:"detail"`
	Hidden    string          `orm:"hidden" json:"-" view:"detail"`
}

func TestIgnoredFieldsAreAbsentFromAllORMViews(t *testing.T) {
	for _, view := range []models.ViewDeclare{models.OriginView, models.MetaView, models.DetailView, models.LiteView} {
		t.Run(string(view), func(t *testing.T) {
			entity := &ignoredFieldModel{ID: 7, Name: "before", Transient: ignoredRelation{State: "private"}, Counter: 9, Hidden: "stored"}
			model, err := getValueModel(reflect.ValueOf(entity), view)
			if err != nil {
				t.Fatal(err)
			}
			fields := model.GetFields()
			if len(fields) != 3 || fields[0].GetName() != "id" || fields[1].GetName() != "name" || fields[2].GetName() != "hidden" {
				t.Fatal("ignored fields entered the ORM model", fields)
			}
			// The reflection index after skipped fields must still map correctly.
			if err := model.SetFieldValue("name", "after"); err != nil || entity.Name != "after" {
				t.Fatal("skipped field changed the mapping of later fields", err)
			}
			if entity.Transient.State != "private" || entity.Counter != 9 {
				t.Fatal("ORM view reset an ignored in-memory value")
			}
		})
	}
}

func TestIgnoredOpaqueTypesDoNotNeedORMTypeSupport(t *testing.T) {
	entity := &ignoredOpaqueModel{ID: 7, Work: make(chan int)}
	model, err := GetEntityModel(entity, nil)
	if err != nil || model == nil || len(model.GetFields()) != 1 {
		t.Fatal("ignored field was parsed as a persisted type", err)
	}
	if _, err := GetEntityModel(&persistedOpaqueModel{ID: 7}, nil); err == nil {
		t.Fatal("JSON-only exclusion weakened persisted field type validation")
	}
}
