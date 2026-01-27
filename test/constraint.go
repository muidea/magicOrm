package test

// ConstraintTestModel 用于测试访问约束的模型
type ConstraintTestModel struct {
	ID         int    `orm:"id key auto" constraint:"ro"`  // 自增主键，只读（不应该有req约束，因为它是自增的）
	Name       string `orm:"name" constraint:"req"`        // 必填
	Password   string `orm:"password" constraint:"wo"`     // 只写（敏感字段）
	CreateTime int64  `orm:"create_time" constraint:"imm"` // 不可变
	UpdateTime int64  `orm:"update_time"`                  // 普通字段
	Email      string `orm:"email"`
	Status     int    `orm:"status" constraint:"req,ro"`  // 必填且只读
	ReadOnlyID int    `orm:"readonly_id" constraint:"ro"` // 只读
	WriteOnly  string `orm:"write_only" constraint:"wo"`  // 只写
}

// Equal 比较两个ConstraintTestModel是否相等
func (s *ConstraintTestModel) Equal(r *ConstraintTestModel) bool {
	if s.ID != r.ID {
		return false
	}
	if s.Name != r.Name {
		return false
	}
	// 密码字段是只写的，查询时不应该返回，所以不比较
	if s.CreateTime != r.CreateTime {
		return false
	}
	if s.UpdateTime != r.UpdateTime {
		return false
	}
	if s.Email != r.Email {
		return false
	}
	if s.Status != r.Status {
		return false
	}
	if s.ReadOnlyID != r.ReadOnlyID {
		return false
	}
	// WriteOnly字段是只写的，查询时不应该返回，所以不比较
	return true
}
