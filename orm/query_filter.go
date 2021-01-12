package orm

import (
	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type filterItem struct {
	filterFun     func(name, value string) string
	value         model.Value
	modelProvider provider.Provider
}

func (s *filterItem) FilterStr(name string, fType model.Type) (ret string, err error) {
	itemStr, itemErr := s.modelProvider.GetFieldStrValue(s.value, fType)
	if itemErr != nil {
		err = itemErr
		return
	}

	ret = s.filterFun(name, itemStr)
	return
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

	//s.params[key] = &filterItem{filterFun: builder.EqualOpr, value: vVal, modelProvider: s.modelProvider}
	s.equalInternal(key, vVal)
	return
}

func (s *queryFilter) equalInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{filterFun: builder.EqualOpr, value: vVal, modelProvider: s.modelProvider}
}

func (s *queryFilter) NotEqual(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{filterFun: builder.NotEqualOpr, value: vVal, modelProvider: s.modelProvider}
	s.notEqualInternal(key, vVal)
	return
}

func (s *queryFilter) notEqualInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{filterFun: builder.NotEqualOpr, value: vVal, modelProvider: s.modelProvider}
}

func (s *queryFilter) Below(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{filterFun: builder.BelowOpr, value: vVal, modelProvider: s.modelProvider}
	s.belowInternal(key, vVal)
	return
}

func (s *queryFilter) belowInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{filterFun: builder.BelowOpr, value: vVal, modelProvider: s.modelProvider}
}

func (s *queryFilter) Above(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{filterFun: builder.AboveOpr, value: vVal, modelProvider: s.modelProvider}
	s.aboveInternal(key, vVal)
	return
}

func (s *queryFilter) aboveInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{filterFun: builder.AboveOpr, value: vVal, modelProvider: s.modelProvider}
}

func (s *queryFilter) In(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{filterFun: builder.InOpr, value: vVal, modelProvider: s.modelProvider}
	s.inInternal(key, vVal)
	return
}

func (s *queryFilter) inInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{filterFun: builder.InOpr, value: vVal, modelProvider: s.modelProvider}
}

func (s *queryFilter) NotIn(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{filterFun: builder.NotInOpr, value: vVal, modelProvider: s.modelProvider}
	s.notInInternal(key, vVal)
	return
}

func (s *queryFilter) notInInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{filterFun: builder.NotInOpr, value: vVal, modelProvider: s.modelProvider}
}

func (s *queryFilter) Like(key string, val interface{}) (err error) {
	vVal, vErr := s.modelProvider.GetEntityValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	//s.params[key] = &filterItem{filterFun: builder.LikeOpr, value: vVal, modelProvider: s.modelProvider}
	s.likeInternal(key, vVal)
	return
}

func (s *queryFilter) likeInternal(key string, vVal model.Value) {
	s.params[key] = &filterItem{filterFun: builder.LikeOpr, value: vVal, modelProvider: s.modelProvider}
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

func (s *queryFilter) Items() map[string]model.FilterItem {
	return s.params
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
