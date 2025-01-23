package test

// Status status
type Status struct {
	ID    int `orm:"id key auto" view:"detail,lite"`
	Value int `orm:"value" view:"detail,lite"`
}

// Group Group
type Group struct {
	ID     int      `orm:"gid key auto" view:"detail,lite"`
	Name   string   `orm:"name" view:"detail,lite"`
	Users  *[]*User `orm:"users" view:"detail,lite"`
	Parent *Group   `orm:"parent" view:"detail,lite"`
}

// User User
type User struct {
	ID     int      `orm:"uid key auto" view:"detail,lite"`
	Name   string   `orm:"name" view:"detail,lite"`
	EMail  string   `orm:"email" view:"detail,lite"`
	Status *Status  `orm:"status" view:"detail,lite"`
	Group  []*Group `orm:"group" view:"detail,lite"`
}

// System System
type System struct {
	ID    int      `orm:"id key auto" view:"detail,lite"`
	Name  string   `orm:"name" view:"detail,lite"`
	Users *[]User  `orm:"users" view:"detail,lite"`
	Tags  []string `orm:"tags" view:"detail,lite"`
}

// Equal Equal
func (s *Group) Equal(r *Group) bool {
	if s.ID != r.ID {
		return false
	}
	if s.Name != r.Name {
		return false
	}

	if s.Users == nil && r.Users != nil {
		return false
	}

	if s.Users != nil && r.Users == nil {
		return false
	}

	if s.Users != nil && r.Users != nil {
		if len(*(s.Users)) != len(*(r.Users)) {
			return false
		}

		for idx := 0; idx < len(*(s.Users)); idx++ {
			l := (*(s.Users))[idx]
			r := (*(r.Users))[idx]
			if !l.Equal(r) {
				return false
			}
		}
	}
	if s.Parent == nil && r.Parent != nil {
		return false
	}

	if s.Parent != nil && r.Parent == nil {
		return false
	}
	if s.Parent != nil && r.Parent != nil {
		if !s.Parent.Equal(r.Parent) {
			return false
		}
	}

	return true
}

// Equal check user Equal
func (s *User) Equal(r *User) bool {
	if s.ID != r.ID {
		return false
	}
	if s.Name != r.Name {
		return false
	}
	if s.EMail != r.EMail {
		return false
	}
	if len(s.Group) != len(r.Group) {
		return false
	}

	for idx := 0; idx < len(s.Group); idx++ {
		l := s.Group[idx]
		r := r.Group[idx]
		if !l.Equal(r) {
			return false
		}
	}

	return true
}

// Equal Equal
func (s *System) Equal(r *System) bool {
	if s.ID != r.ID {
		return false
	}
	if s.Name != r.Name {
		return false
	}

	if s.Users == nil && r.Users != nil {
		return false
	}

	if s.Users != nil && r.Users == nil {
		return false
	}

	if s.Users != nil && r.Users != nil {
		if len(*(s.Users)) != len(*(r.Users)) {
			return false
		}

		for idx := 0; idx < len(*(s.Users)); idx++ {
			l := (*(s.Users))[idx]
			r := (*(r.Users))[idx]
			if !l.Equal(&r) {
				return false
			}
		}
	}
	if len(s.Tags) != len(r.Tags) {
		return false
	}

	for idx := 0; idx < len(s.Tags); idx++ {
		l := s.Tags[idx]
		r := r.Tags[idx]
		if l != r {
			return false
		}
	}

	return true
}

const (
	// ByPiece 按件数
	ByPiece = iota
	// ByMoney 按金额
	ByMoney
)

const (
	// CheckSingle 考核单项
	CheckSingle = iota
	// CheckTwice 考核两项
	CheckTwice
)

// Goal 考核目标
type Goal struct {
	ID    int     `json:"id" orm:"id key auto" view:"detail,lite"` // ID
	Type  int     `json:"type" orm:"type" view:"detail,lite"`
	Value float32 `json:"value" orm:"value" view:"detail,lite"`
}

// SpecialGoal 特殊目标
type SpecialGoal struct {
	ID            int      `json:"id" orm:"id key auto" view:"detail,lite"` // ID
	CheckDistrict []string `json:"checkDistrict" orm:"checkDistrict" view:"detail,lite"`
	CheckProduct  []string `json:"checkProduct" orm:"checkProduct" view:"detail,lite"`
	CheckType     int      `json:"checkType" orm:"checkType" view:"detail,lite"`
	CheckValue    Goal     `json:"checkValue" orm:"checkValue" view:"detail,lite"`
}

// KPI 代理商考核指标
type KPI struct {
	ID            int         `json:"id" orm:"id key auto" view:"detail,lite"`              // ID
	Title         string      `json:"title" orm:"title" view:"detail,lite"`                 // 名称
	JoinValue     Goal        `json:"joinValue" orm:"joinValue" view:"detail,lite"`         // 加盟目标
	PerMonthValue Goal        `json:"perMonthValue" orm:"perMonthValue" view:"detail,lite"` // 每月目标
	SpecialValue  SpecialGoal `json:"specialValue" orm:"specialValue" view:"detail,lite"`   // 特殊地区或产品目标
	Default       bool        `json:"default" orm:"default" view:"detail,lite"`
}

type ValueItem struct {
	ID    int     `json:"id" orm:"id key auto" view:"detail,lite"`
	Level int     `json:"level" orm:"level" view:"detail,lite"`
	Type  int     `json:"type" orm:"type" view:"detail,lite"`
	Value float64 `json:"value" orm:"value" view:"detail,lite"`
}

type ValueScope struct {
	ID        int     `json:"id" orm:"id key auto" view:"detail,lite"`
	LowValue  float64 `json:"lowValue" orm:"lowValue" view:"detail,lite"`
	HighValue float64 `json:"highValue" orm:"highValue" view:"detail,lite"`
}

type RewardPolicy struct {
	ID          int         `json:"id" orm:"id key auto" view:"detail,lite"`
	Name        string      `json:"name" orm:"name" view:"detail,lite"`
	Description string      `json:"description" orm:"description" view:"detail,lite"`
	ValueItem   []ValueItem `json:"item" orm:"item" view:"detail,lite"`
	ValueScope  ValueScope  `json:"scope" orm:"scope" view:"detail,lite"`
	Status      *Status     `json:"status" orm:"status" view:"detail,lite"`
	Creater     int         `json:"creater" orm:"creater" view:"detail,lite"`
	UpdateTime  int64       `json:"updateTime" orm:"updateTime" view:"detail,lite"`
	Namespace   string      `json:"namespace" orm:"namespace" view:"detail,lite"`
}
