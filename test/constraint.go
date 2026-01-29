package test

// ConstraintTestModel 用于测试访问约束的模型
type ConstraintTestModel struct {
	ID         int    `orm:"id key auto" constraint:"ro"` // 自增主键，只读（不应该有req约束，因为它是自增的）
	Name       string `orm:"name" constraint:"req"`       // 必填
	Password   string `orm:"password" constraint:"wo"`    // 只写（敏感字段）
	CreateTime int64  `orm:"create_time" constraint:"ro"` // 不可变
	UpdateTime int64  `orm:"update_time"`                 // 普通字段
	Email      string `orm:"email"`
	Status     int    `orm:"status" constraint:"req,ro"`  // 必填且只读
	ReadOnlyID int    `orm:"readonly_id" constraint:"ro"` // 只读
	WriteOnly  string `orm:"write_only" constraint:"wo"`  // 只写
}

// ContentConstraintTestModel 用于测试内容值约束的模型
type ContentConstraintTestModel struct {
	ID          int     `orm:"id key auto" constraint:"ro"`                                               // 自增主键，只读
	Name        string  `orm:"name" constraint:"req,min=3,max=50"`                                        // 必填，长度3-50
	Age         int     `orm:"age" constraint:"min=0,max=150"`                                            // 年龄0-150
	Score       float64 `orm:"score" constraint:"range=0.0:100.0"`                                        // 分数0.0-100.0
	Status      string  `orm:"status" constraint:"in=active:inactive:pending"`                            // 枚举值
	Email       string  `orm:"email" constraint:"re=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,64}$"` // 正则匹配邮箱，使用{2,64}避免逗号问题
	Description string  `orm:"description" constraint:"max=500"`                                          // 最大长度500
	ItemCount   int     `orm:"item_count" constraint:"min=1"`                                             // 最小值为1
	Price       float64 `orm:"price" constraint:"range=0.01:9999.99"`                                     // 价格范围
	Category    string  `orm:"category" constraint:"in=A:B:C:D"`                                          // 分类枚举
	Code        string  `orm:"code" constraint:"re=^[A-Z]{3}-\\d{3}$"`                                    // 正则匹配格式：ABC-123
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

// Equal 比较两个ContentConstraintTestModel是否相等
func (s *ContentConstraintTestModel) Equal(r *ContentConstraintTestModel) bool {
	if s.ID != r.ID {
		return false
	}
	if s.Name != r.Name {
		return false
	}
	if s.Age != r.Age {
		return false
	}
	if s.Score != r.Score {
		return false
	}
	if s.Status != r.Status {
		return false
	}
	if s.Email != r.Email {
		return false
	}
	if s.Description != r.Description {
		return false
	}
	if s.ItemCount != r.ItemCount {
		return false
	}
	if s.Price != r.Price {
		return false
	}
	if s.Category != r.Category {
		return false
	}
	if s.Code != r.Code {
		return false
	}
	return true
}
