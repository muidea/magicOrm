package orm

import (
	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type filterItem struct {
	oprFun   func(name, value string) string
	oprValue model.Value
}

func (s *filterItem) OprFunc() model.OprFunc {
	return s.oprFun
}

func (s *filterItem) OprValue() model.Value {
	return s.oprValue
}

// queryFilter queryFilter
type queryFilter struct {
	params        map[string]model.FilterItem
	maskValue     model.Model
	pageFilter    *util.PageFilter
	sortFilter    *util.SortFilter
	modelProvider provider.Provider
}

func (s *queryFilter) Equal(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{oprFun: builder.EqualOpr, oprValue: vVal}
	s.equalInternal(key, vVal)
	return
}

func (s *queryFilter) equalInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{oprFun: builder.EqualOpr, oprValue: vVal}
}

func (s *queryFilter) NotEqual(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{oprFun: builder.NotEqualOpr, oprValue: vVal}
	s.notEqualInternal(key, vVal)
	return
}

func (s *queryFilter) notEqualInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{oprFun: builder.NotEqualOpr, oprValue: vVal}
}

func (s *queryFilter) Below(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{oprFun: builder.BelowOpr, oprValue: vVal}
	s.belowInternal(key, vVal)
	return
}

func (s *queryFilter) belowInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{oprFun: builder.BelowOpr, oprValue: vVal}
}

func (s *queryFilter) Above(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{oprFun: builder.AboveOpr, oprValue: vVal}
	s.aboveInternal(key, vVal)
	return
}

func (s *queryFilter) aboveInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{oprFun: builder.AboveOpr, oprValue: vVal}
}

func (s *queryFilter) In(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{oprFun: builder.InOpr, oprValue: vVal}
	s.inInternal(key, vVal)
	return
}

func (s *queryFilter) inInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{oprFun: builder.InOpr, oprValue: vVal}
}

func (s *queryFilter) NotIn(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{oprFun: builder.NotInOpr, oprValue: vVal}
	s.notInInternal(key, vVal)
	return
}

func (s *queryFilter) notInInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{oprFun: builder.NotInOpr, oprValue: vVal}
}

func (s *queryFilter) Like(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{oprFun: builder.LikeOpr, oprValue: vVal}
	s.likeInternal(key, vVal)
	return
}

func (s *queryFilter) likeInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{oprFun: builder.LikeOpr, oprValue: vVal}
}

func (s *queryFilter) ValueMask(val interface{}) (err error) {
	vModel, vErr := s.modelProvider.GetEntityModel(val)
	if vErr != nil {
		err = vErr
		return
	}

	s.maskValue = vModel
	return
}

func (s *queryFilter) Page(filter *util.PageFilter) {
	s.pageFilter = filter
}

func (s *queryFilter) Sort(sorter *util.SortFilter) {
	s.sortFilter = sorter
}

func (s *queryFilter) GetFilterItem(name string) model.FilterItem {
	v, ok := s.params[name]
	if ok {
		return v
	}

	return nil
}

func (s *queryFilter) Pagination() (limit, offset int, paging bool) {
	paging = false
	if s.pageFilter == nil {
		return
	}

	paging = true
	limit = s.pageFilter.PageSize
	offset = s.pageFilter.PageSize * (s.pageFilter.PageNum - 1)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100
	}

	return
}

func (s *queryFilter) MaskModel() (ret model.Model) {
	ret = s.maskValue
	return
}

func (s *queryFilter) Sorter() model.Sorter {
	if s.sortFilter == nil {
		return nil
	}

	return s.sortFilter
}
