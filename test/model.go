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
