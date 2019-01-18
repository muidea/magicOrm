package test

import (
	"log"
	"testing"

	orm "muidea.com/magicOrm"
	"muidea.com/magicOrm/model"
)

// Person Person
type Person struct {
	ID      int     `orm:"id key auto"`
	Name    string  `orm:"name"`
	EMail   string  `orm:"email"`
	Age     int     `orm:"age"`
	Payward float32 `orm:"payward"`
}

func TestFilterOpr(t *testing.T) {
	cache := model.NewCache()

	p := &Person{}
	info, err := model.GetObjectStructInfo(p, cache)
	if err != nil {
		t.Errorf("GetObjectStructInfo failed, err:%s", err.Error())
		return
	}

	filter := orm.NewFilter()
	name := "hello"
	err = filter.Equle("Name", &name)
	if err != nil {
		t.Errorf("Equle failed, err:%s", err.Error())
		return
	}

	age := 10
	err = filter.NotEqule("Age", &age)
	if err != nil {
		t.Errorf("NotEqule failed, err:%s", err.Error())
		return
	}

	payward := float32(12.34)
	err = filter.Above("Payward", &payward)
	if err != nil {
		t.Errorf("Above failed, err:%s", err.Error())
		return
	}

	strVal, valErr := filter.Builder(info)
	if valErr != nil {
		t.Errorf("Builder failed, err:%s", valErr.Error())
		return
	}
	if strVal != "`Name` = 'hello' AND `Age` != 10 AND `Payward` > 12.340000" {
		t.Errorf("Builder failed, strVal:%s", strVal)
	}

	log.Print(strVal)
}
