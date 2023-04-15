package remote

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
)

func TestFilter(t *testing.T) {
	type Unit struct {
		ID      int     `orm:"id key auto"`
		Name    string  `orm:"name"`
		SubUnit []*Unit `orm:"subUnit"`
	}

	filter := NewFilter()

	err := filter.Above("aaa", 123)
	if err != nil {
		t.Errorf("above failed, err:%s", err.Error())
		return
	}

	val := time.Now()
	err = filter.NotEqual("abb", val)
	if err != nil {
		t.Errorf("noEqual failed, err:%s", err.Error())
		return
	}

	u := &Unit{}
	err = filter.Equal("acc", u)
	if err != nil {
		t.Errorf("equal failed, err:%s", err.Error())
		return
	}

	uList := []Unit{*u}
	err = filter.In("add", uList)
	if err != nil {
		t.Errorf("In failed, err:%s", err.Error())
		return
	}

	idList := []int{10, 11}
	err = filter.In("id", idList)
	if err != nil {
		t.Errorf("In failed, err:%s", err.Error())
		return
	}

	dtList := []*time.Time{&val}
	err = filter.NotIn("aee", dtList)
	if err != nil {
		t.Errorf("noIn failed, err:%s", err.Error())
		return
	}

	pageFilter := &util.Pagination{PageSize: 100, PageNum: 1}
	filter.Page(pageFilter)

	data, dataErr := json.Marshal(filter)
	if dataErr != nil {
		t.Errorf("marshal failed, err:%s", dataErr.Error())
		return
	}

	log.Print(*filter)
	log.Print(string(data))

	filter2 := NewFilter()
	err = json.Unmarshal(data, filter2)
	if err != nil {
		t.Errorf("unmarshal failed, err:%s", err.Error())
		return
	}

	log.Print(*filter2)
	val2 := filter2.NotEqualFilter[0]
	{
		vType := reflect.TypeOf(val2.Value)
		log.Print(vType.String())
		log.Print(val2)
	}
	val3 := filter2.EqualFilter[0]
	{
		vType := reflect.TypeOf(val3.Value)
		log.Print(vType.String())
		log.Print(val3)
	}
	val4 := filter2.InFilter[0]
	{
		vType := reflect.TypeOf(val4.Value)
		log.Print(vType.String())
		log.Print(val4)
	}
	val5 := filter2.NotInFilter[0]
	{
		vType := reflect.TypeOf(val5.Value)
		log.Print(vType.String())
		log.Print(val5)
	}

	if filter2.PageFilter != nil {
		val6 := filter2.PageFilter
		vType := reflect.TypeOf(*val6)
		log.Print(vType.String())
		log.Print(*val6)
	}
}

func TestQueryFilter_FromContentFilter(t *testing.T) {
	filter := util.NewFilter()
	filter.Set("set", -123)
	filter.Set("setB", true)
	filter.Equal("equal", -123)
	filter.NotEqual("notequal", "123")
	filter.Below("below", 40)
	filter.Above("above", 40)
	filter.In("in", []float32{12.23, 23.45})
	filter.In("inB", []bool{true, false, true})
	filter.NotIn("notin", []float32{12.23, 23.45})
	filter.Like("like", "hello world")

	bcFilter := NewFilter()
	bcFilter.FromContentFilter(filter)
	if len(bcFilter.EqualFilter) != 3 {
		t.Errorf("get equal failed")
		return
	}

	if len(bcFilter.NotEqualFilter) != 1 {
		t.Errorf("get notequal failed")
		return
	}

	if len(bcFilter.BelowFilter) != 1 {
		t.Errorf("get below failed")
		return
	}

	if len(bcFilter.AboveFilter) != 1 {
		t.Errorf("get above failed")
		return
	}

	if len(bcFilter.InFilter) != 2 {
		t.Errorf("get in failed")
		return
	}

	if len(bcFilter.NotInFilter) != 1 {
		t.Errorf("get notin failed")
		return
	}

	if len(bcFilter.LikeFilter) != 1 {
		t.Errorf("get like failed")
		return
	}
}
