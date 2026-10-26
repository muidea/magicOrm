package utils

type Pagination struct {
	PageNum  int `json:"pageNum"`
	PageSize int `json:"pageSize"`
}

func (s *Pagination) Limit() int64 {
	if s.PageNum < 1 {
		s.PageNum = 1
	}

	if s.PageSize < 1 {
		s.PageSize = 10
	}

	return int64(s.PageSize)
}

func (s *Pagination) Offset() int64 {
	if s.PageNum < 1 {
		s.PageNum = 1
	}

	if s.PageSize < 1 {
		s.PageSize = 10
	}

	return int64(s.PageNum-1) * int64(s.PageSize)
}

type SortFilter struct {
	// true:升序,false:降序
	AscFlag bool `json:"ascFlag"`
	// 排序字段
	FieldName string `json:"fieldName"`
}

func (s *SortFilter) Name() string {
	return s.FieldName
}

func (s *SortFilter) AscSort() bool {
	return s.AscFlag
}
