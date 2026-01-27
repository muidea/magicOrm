package provider

import (
	"testing"
	"time"

	"github.com/muidea/magicOrm/models"
)

type Simple struct {
	ID        int       `orm:"id key auto" view:"detail,lite"`
	I8        int8      `orm:"i8" view:"detail,lite"`
	I16       int16     `orm:"i16" view:"detail,lite"`
	I32       int32     `orm:"i32" view:"detail,lite"`
	I64       uint64    `orm:"i64" view:"detail,lite"`
	Name      string    `orm:"name" view:"detail,lite"`
	Value     float32   `orm:"value" view:"detail,lite"`
	F64       float64   `orm:"f64" view:"detail,lite"`
	TimeStamp time.Time `orm:"ts datetime" view:"detail,lite"`
	Flag      bool      `orm:"flag" view:"detail,lite"`
	Namespace string    `orm:"namespace"`
}

func TestNewLocalProvider(t *testing.T) {
	s001 := &Simple{}
	provider := NewLocalProvider("t001", nil)
	_, sModelErr := provider.RegisterModel(s001)
	if sModelErr != nil {
		t.Errorf("%s", sModelErr.Error())
		return
	}

	_, sModelErr = provider.GetEntityModel(s001, true)
	if sModelErr != nil {
		t.Errorf("%s", sModelErr.Error())
		return
	}

	s001.ID = 100
	s001.I8 = 8
	s001.I16 = 16
	s001.I32 = 32
	s001.I64 = 64
	s001.Name = "s001"
	s001.Value = 123.456
	s001.F64 = 456789.123
	s001.TimeStamp = time.Now()
	s001.Flag = true
	s001.Namespace = "t001"
	_, sValueErr := provider.GetEntityValue(s001)
	if sValueErr != nil {
		t.Errorf("%s", sValueErr.Error())
		return
	}

	_, sTypeErr := provider.GetEntityType(s001)
	if sTypeErr != nil {
		t.Errorf("%s", sTypeErr.Error())
		return
	}

	_, sFilterErr := provider.GetEntityFilter(s001, models.MetaView)
	if sFilterErr != nil {
		t.Errorf("%s", sFilterErr.Error())
		return
	}

}
