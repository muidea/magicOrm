package test

// Status status
type Status struct {
	ID    int `orm:"id key auto"`
	Value int `orm:"value"`
}

// Group Group
type Group struct {
	ID     int      `orm:"id key auto"`
	Name   string   `orm:"name"`
	Users  *[]*User `orm:"users"`
	Parent *Group   `orm:"parent"`
}

// User User
type User struct {
	ID     int      `orm:"id key auto"`
	Name   string   `orm:"name"`
	EMail  string   `orm:"email"`
	Status *Status  `orm:"status"`
	Group  []*Group `orm:"group"`
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
	ID    int     `json:"id" orm:"id key auto"` // ID
	Type  int     `json:"type" orm:"type"`
	Value float32 `json:"value" orm:"value"`
}

// SpecialGoal 特殊目标
type SpecialGoal struct {
	ID            int      `json:"id" orm:"id key auto"` // ID
	CheckDistrict []string `json:"checkDistrict" orm:"checkDistrict"`
	CheckProduct  []string `json:"checkProduct" orm:"checkProduct"`
	CheckType     int      `json:"checkType" orm:"checkType"`
	CheckValue    Goal     `json:"checkValue" orm:"checkValue"`
}

// KPI 代理商考核指标
type KPI struct {
	ID            int         `json:"id" orm:"id key auto"`              // ID
	Title         string      `json:"title" orm:"title"`                 // 名称
	JoinValue     Goal        `json:"joinValue" orm:"joinValue"`         // 加盟目标
	PerMonthValue Goal        `json:"perMonthValue" orm:"perMonthValue"` // 每月目标
	SpecialValue  SpecialGoal `json:"specialValue" orm:"specialValue"`   // 特殊地区或产品目标
	Default       bool        `json:"default" orm:"default"`
}

type ValueItem struct {
	ID    int     `json:"id" orm:"id key auto"`
	Level int     `json:"level" orm:"level"`
	Type  int     `json:"type" orm:"type"`
	Value float64 `json:"value" orm:"value"`
}

type ValueScope struct {
	ID        int     `json:"id" orm:"id key auto"`
	LowValue  float64 `json:"lowValue" orm:"lowValue"`
	HighValue float64 `json:"highValue" orm:"highValue"`
}

type RewardPolicy struct {
	ID          int         `json:"id" orm:"id key auto"`
	Name        string      `json:"name" orm:"name"`
	Description string      `json:"description" orm:"description"`
	ValueItem   []ValueItem `json:"item" orm:"item"`
	ValueScope  ValueScope  `json:"scope" orm:"scope"`
	Status      *Status     `json:"status" orm:"status"`
	Creater     int         `json:"creater" orm:"creater"`
	UpdateTime  int64       `json:"updateTime" orm:"updateTime"`
	Namespace   string      `json:"namespace" orm:"namespace"`
}
