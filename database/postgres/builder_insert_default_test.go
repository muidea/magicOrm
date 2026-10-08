package postgres

import (
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

type DefaultInsertHost struct {
	ID int64 `orm:"id key auto"`
}

func TestBuildInsertAutoOnlyUsesDefaultValues(t *testing.T) {
	for _, isRemote := range []bool{false, true} {
		p := provider.NewLocalProvider("default-insert", nil)
		var def any = &DefaultInsertHost{}
		var input any = &DefaultInsertHost{}
		if isRemote {
			p = provider.NewRemoteProvider("default-insert", nil)
			obj, err := helper.GetObject(def)
			if err != nil {
				t.Fatal(err)
			}
			def = obj
			val, err := helper.GetObjectValue(input)
			if err != nil {
				t.Fatal(err)
			}
			input = val
		}
		if _, err := p.RegisterModel(def); err != nil {
			t.Fatal(err)
		}
		m, err := p.GetEntityModel(input, true)
		if err != nil {
			t.Fatal(err)
		}
		result, err := NewBuilder(p, codec.New(p, "default")).BuildInsert(m)
		if err != nil {
			t.Fatal(err)
		}
		if result.SQL() != `INSERT INTO "default_DefaultInsertHost" DEFAULT VALUES RETURNING id` || len(result.Args()) != 0 {
			t.Fatalf("remote=%v invalid auto-only insert: %s %v", isRemote, result.SQL(), result.Args())
		}
	}
}
