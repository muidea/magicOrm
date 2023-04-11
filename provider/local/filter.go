package local

import "github.com/muidea/magicCommon/foundation/util"

type itemValue struct {
	name  string     `json:"name"`
	value *valueImpl `json:"value"`
}

type filterImpl struct {
	equalFilter    []*itemValue     `json:"equal"`
	notEqualFilter []*itemValue     `json:"noEqual"`
	belowFilter    []*itemValue     `json:"below"`
	aboveFilter    []*itemValue     `json:"above"`
	inFilter       []*itemValue     `json:"in"`
	notInFilter    []*itemValue     `json:"notIn"`
	likeFilter     []*itemValue     `json:"like"`
	pageFilter     *util.Pagination `json:"page"`
	sortFilter     *util.SortFilter `json:"sort"`
}
