# magicOrm

Golang对象的orm框架，目前仅支持mySQL数据库。

一个所见即所得的orm框架。

## 基本约束

1. 如果Filed类型为isPtr为true,则标识该字段为可选类型

## 怎么使用


    type User struct {
	    ID    int      `orm:"id key auto"`
	    Name  string   `orm:"name"`
	    EMail string   `orm:"email"`
	    Group []*Group `orm:"group"`
    }
    
    o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
	}
	
    user1 := &User{Name: "demo", EMail: "123@demo.com", Group: []*Group{}}
    err = o1.Create(user1)
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
	}
	err = o1.Insert(user1)
	if err != nil {
		t.Errorf("insert user1 failed, err:%s", err.Error())
	}

	user2 := &User{ID: user1.ID}
	err = o1.Query(user2)
	if err != nil {
		t.Errorf("query user2 failed, err:%s", err.Error())
	}

## 支持操作
### insert
插入单个对象
### update
更新指定对象
### delete
删除指定对象
#### query
查询指定对象
### batchQuery
按条件查询对象，支持的查询如下

Equal 等于

NotEqual 不等于

Below 小于

Above 大于

In 在指定集合内

NotIn 在指定集合外

Like 模糊匹配


## 特殊说明
1. 支持的基础数据类型:int,int8,int16,int32,int64,uint,uint8,uint16,uint32,uint64,float32,float64,bool,string，以及对应的指针类型

2. 支持的复合数据类型为:slice,struct，以及对应的指针类型

3. 对于指针类型的复合类型成员，orm 处理该对象时，只处理对象与复合类型成员之间的关系，复合成员的保存由外部保证。

4. 对于普通复合类型成员，orm保存该对象时，会同步处理对象与复合类型成员之间的关系，并且保证对象与复合类型成员的对象

5. 不支持基础类型指针slice，例如[]*bool,[]*int这种表达方式

6. 支持slice指针，例如*[]bool, *[]int
