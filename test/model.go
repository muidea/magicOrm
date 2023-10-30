package test

// Status status
type Status struct {
	ID    int `orm:"id key auto" view:"view,lite"`
	Value int `orm:"value" view:"view,lite"`
}

// Group Group
type Group struct {
	ID     int      `orm:"gid key auto" view:"view,lite"`
	Name   string   `orm:"name" view:"view,lite"`
	Users  *[]*User `orm:"users" view:"view"`
	Parent *Group   `orm:"parent" view:"view"`
}

// User User
type User struct {
	ID     int      `orm:"uid key auto" view:"view,lite"`
	Name   string   `orm:"name" view:"view,lite"`
	EMail  string   `orm:"email" view:"view,lite"`
	Status *Status  `orm:"status" view:"view,lite"`
	Group  []*Group `orm:"group" view:"view,lite"`
}

// System System
type System struct {
	ID    int      `orm:"id key auto"`
	Name  string   `orm:"name"`
	Users *[]User  `orm:"users"`
	Tags  []string `orm:"tags"`
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
	ID    int     `json:"id" orm:"id key auto" view:"view,lite"` // ID
	Type  int     `json:"type" orm:"type" view:"view,lite"`
	Value float32 `json:"value" orm:"value" view:"view,lite"`
}

// SpecialGoal 特殊目标
type SpecialGoal struct {
	ID            int      `json:"id" orm:"id key auto" view:"view,lite"` // ID
	CheckDistrict []string `json:"checkDistrict" orm:"checkDistrict" view:"view,lite"`
	CheckProduct  []string `json:"checkProduct" orm:"checkProduct" view:"view,lite"`
	CheckType     int      `json:"checkType" orm:"checkType" view:"view,lite"`
	CheckValue    Goal     `json:"checkValue" orm:"checkValue" view:"view,lite"`
}

// KPI 代理商考核指标
type KPI struct {
	ID            int         `json:"id" orm:"id key auto" view:"view,lite"`              // ID
	Title         string      `json:"title" orm:"title" view:"view,lite"`                 // 名称
	JoinValue     Goal        `json:"joinValue" orm:"joinValue" view:"view,lite"`         // 加盟目标
	PerMonthValue Goal        `json:"perMonthValue" orm:"perMonthValue" view:"view,lite"` // 每月目标
	SpecialValue  SpecialGoal `json:"specialValue" orm:"specialValue" view:"view,lite"`   // 特殊地区或产品目标
	Default       bool        `json:"default" orm:"default" view:"view,lite"`
}

type ValueItem struct {
	ID    int     `json:"id" orm:"id key auto" view:"view,lite"`
	Level int     `json:"level" orm:"level" view:"view,lite"`
	Type  int     `json:"type" orm:"type" view:"view,lite"`
	Value float64 `json:"value" orm:"value" view:"view,lite"`
}

type ValueScope struct {
	ID        int     `json:"id" orm:"id key auto" view:"view,lite"`
	LowValue  float64 `json:"lowValue" orm:"lowValue" view:"view,lite"`
	HighValue float64 `json:"highValue" orm:"highValue" view:"view,lite"`
}

type RewardPolicy struct {
	ID          int         `json:"id" orm:"id key auto" view:"view,lite"`
	Name        string      `json:"name" orm:"name" view:"view,lite"`
	Description string      `json:"description" orm:"description" view:"view,lite"`
	ValueItem   []ValueItem `json:"item" orm:"item" view:"view,lite"`
	ValueScope  ValueScope  `json:"scope" orm:"scope" view:"view,lite"`
	Status      *Status     `json:"status" orm:"status" view:"view,lite"`
	Creater     int         `json:"creater" orm:"creater" view:"view,lite"`
	UpdateTime  int64       `json:"updateTime" orm:"updateTime" view:"view,lite"`
	Namespace   string      `json:"namespace" orm:"namespace" view:"view,lite"`
}
