# magicOrm

Golang对象的ORM框架，支持PostgreSQL和MySQL数据库。一个所见即所得的ORM框架。

## 特性

- **多数据库支持**: PostgreSQL和MySQL
- **类型安全**: 完整的Go类型映射支持
- **灵活的关系映射**: 支持一对一、一对多、多对多关系
- **强大的约束系统**: 内置数据验证和业务规则
- **视图模式**: 支持detail/lite视图控制字段输出
- **事务支持**: 完整的ACID事务处理
- **高性能**: 连接池和批量操作优化

## 安装

```bash
go get github.com/muidea/magicOrm
```

## 快速开始

### 1. 定义模型

```go
type User struct {
    ID     int      `orm:"uid key auto" view:"detail,lite"`
    Name   string   `orm:"name" view:"detail,lite"`
    EMail  string   `orm:"email" view:"detail,lite"`
    Status *Status  `orm:"status" view:"detail,lite"`
    Group  []*Group `orm:"group" view:"detail,lite"`
}

type Status struct {
    ID    int `orm:"id key auto" view:"detail,lite"`
    Value int `orm:"value" view:"detail,lite"`
}

type Group struct {
    ID     int      `orm:"gid key auto" view:"detail,lite"`
    Name   string   `orm:"name" view:"detail,lite"`
    Users  *[]*User `orm:"users" view:"detail,lite"`
    Parent *Group   `orm:"parent" view:"detail,lite"`
}
```

### 2. 初始化ORM

```go
import (
    "github.com/muidea/magicOrm/orm"
    "github.com/muidea/magicOrm/provider"
)

// 数据库配置
config := &orm.Options{
    Driver: "postgres", // 或 "mysql"
    DSN:    "数据库连接字符串",
}

// 初始化
orm.Initialize()
defer orm.Uninitialized()

// 创建Provider
localProvider := provider.NewLocalProvider("default", nil)

// 创建ORM实例
o1, err := orm.NewOrm(localProvider, config, "schema_prefix")
defer o1.Release()
if err != nil {
    log.Fatal(err)
}
```

### 3. 注册模型

```go
// 注册所有模型
entityList := []any{&User{}, &Status{}, &Group{}}
modelList, err := registerLocalModel(localProvider, entityList)
if err != nil {
    log.Fatal(err)
}

// 创建数据表
err = createModel(o1, modelList)
if err != nil {
    log.Fatal(err)
}
```

### 4. 基本CRUD操作

#### 插入数据
```go
user := &User{
    Name:  "demo", 
    EMail: "123@demo.com", 
    Group: []*Group{},
}

userModel, err := localProvider.GetEntityModel(user, true)
if err != nil {
    log.Fatal(err)
}

userModel, err = o1.Insert(userModel)
if err != nil {
    log.Fatal(err)
}

user = userModel.Interface(true).(*User)
fmt.Printf("插入成功，ID: %d\n", user.ID)
```

#### 查询数据
```go
// 单个查询
queryUser := &User{ID: user.ID}
queryModel, err := localProvider.GetEntityModel(queryUser, true)
if err != nil {
    log.Fatal(err)
}

queryModel, err = o1.Query(queryModel)
if err != nil {
    log.Fatal(err)
}

result := queryModel.Interface(true).(*User)
fmt.Printf("查询结果: %+v\n", result)
```

#### 更新数据
```go
user.Name = "updated name"
userModel, err = localProvider.GetEntityModel(user, true)
if err != nil {
    log.Fatal(err)
}

userModel, err = o1.Update(userModel)
if err != nil {
    log.Fatal(err)
}
```

#### 删除数据
```go
_, err = o1.Delete(userModel)
if err != nil {
    log.Fatal(err)
}
```

## 核心功能

### CRUD操作

- **Insert** - 插入单个对象
- **Update** - 更新指定对象  
- **Delete** - 删除指定对象
- **Query** - 查询指定对象（根据主键）
- **BatchQuery** - 按条件批量查询对象

### 查询过滤器

```go
filter, _ := localProvider.GetModelFilter(model)

// 基础比较
filter.Equal("name", "value")        // 等于
filter.NotEqual("name", "value")     // 不等于
filter.Below("age", 18)              // 小于
filter.Above("age", 18)              // 大于
filter.In("id", []int{1, 2, 3})      // 在指定集合内
filter.NotIn("id", []int{1, 2, 3})   // 在指定集合外
filter.Like("name", "%demo%")        // 模糊匹配

// 组合查询
filter.Equal("status", "active")
filter.Above("created_at", startTime)
```

### 模型管理

```go
// 创建表
err := o1.Create(model)

// 删除表
err := o1.Drop(model)

// 检查表是否存在
exists, err := o1.Exist(model)
```

### 事务支持

```go
// 开始事务
tx, err := o1.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

// 在事务中执行操作
userModel, err := tx.Insert(userModel)
if err != nil {
    return err
}

// 提交事务
err = tx.Commit()
if err != nil {
    return err
}
```

## 数据类型和标签

### 支持的数据类型

#### 基础数据类型
- 整数: `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- 浮点数: `float32`, `float64`
- 布尔值: `bool`
- 字符串: `string`
- 时间: `time.Time`
- UUID: `string` (配合 `uuid` 标签使用)
- 以及对应的指针类型

#### 复合数据类型
- 结构体: `struct`
- 切片: `slice`
- 指针: `pointer`

### ORM标签说明

```go
type User struct {
    ID     int      `orm:"uid key auto" view:"detail,lite"`                // 主键，自增
    Name   string   `orm:"name" constraint:"req,min=3,max=50" view:"detail,lite"` // 必填，长度3-50
    EMail  string   `orm:"email" constraint:"re=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,64}$" view:"detail,lite"` // 邮箱格式
    UU_ID  string   `orm:"uuid key uuid" view:"detail,lite"`                // UUID主键
    Status *Status  `orm:"status" view:"detail,lite"`                       // 可选关联
    Group  []*Group `orm:"group" view:"detail,lite"`                        // 多对多关联
    Time   time.Time `orm:"created_at datetime" constraint:"ro" view:"detail"` // 不可变时间
}
```

**ORM标签说明:**
- `key`: 主键标识
- `auto`: 自增标识 (PostgreSQL使用SERIAL/BIGSERIAL)
- `uuid`: UUID类型主键 (PostgreSQL VARCHAR(32))
- `datetime`: 时间类型
- `view`: 视图声明，支持 `detail` 和 `lite` 模式

### 约束系统

约束标签用于定义字段的业务规则和数据验证，支持访问行为约束和内容值约束。

#### 语法规范
```
constraint:"指令1,指令2=参数1,指令3=参数1:参数2"
```

- **指令分隔符**: `,` (英文逗号)
- **键值分隔符**: `=` (等号)
- **参数分隔符**: `:` (冒号)

#### 访问行为约束

| 约束 | 参数 | 描述 | 使用场景 |
| :--- | :--- | :--- | :--- |
| **`req`** | 无 | **Required**: 必填/必传 | 校验值不能为零值（0, "", nil） |
| **`ro`** | 无 | **Read-Only**: 只读 | 输出接口展示，更新接口忽略此字段 |
| **`wo`** | 无 | **Write-Only**: 只写 | 敏感字段（如密码），禁止在展示接口输出 |

#### 内容值约束

| 约束 | 参数示例 | 描述 | 适用类型 |
| :--- | :--- | :--- | :--- |
| **`min`** | `min=1` | **最小值/最小长度** | 数字、字符串、数组 |
| **`max`** | `max=100` | **最大值/最大长度** | 数字、字符串、数组 |
| **`range`** | `range=1:100` | **区间约束**: 定义数值的闭区间 `[min, max]` | 数字、浮点数 |
| **`in`** | `in=active:inactive:pending` | **枚举约束**: 字段值必须在定义的参数集合内 | 字符串、数字 |
| **`re`** | `re=^[a-z]+$` | **正则约束**: 字段值必须匹配指定的正则表达式 | 字符串 |

#### 约束示例

```go
type UserAccount struct {
    // 访问行为约束示例
    ID         int    `orm:"id key auto" constraint:"ro"`           // 自增主键，只读
    Name       string `orm:"name" constraint:"req"`                // 必填
    Password   string `orm:"password" constraint:"wo"`              // 只写（敏感字段）
    CreateTime int64  `orm:"create_time" constraint:"ro"`         // 不可变
    Email      string `orm:"email" constraint:"req,re=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,64}$"`
    UpdateTime int64  `orm:"update_time"`                          // 普通字段
    Status     int    `orm:"status" constraint:"req,ro"`           // 必填且只读
}

type Product struct {
    // 内容值约束示例
    ID          int     `orm:"id key auto" constraint:"ro"`          // 只读主键
    Name        string  `orm:"name" constraint:"req,min=3,max=50"`   // 必填，长度3-50
    Age         int     `orm:"age" constraint:"min=0,max=150"`      // 年龄0-150
    Score       float64 `orm:"score" constraint:"range=0.0:100.0"`  // 分数0.0-100.0
    Status      string  `orm:"status" constraint:"in=active:inactive:pending"` // 枚举值
    Description string  `orm:"description" constraint:"max=500"`    // 最大长度500
    Price       float64 `orm:"price" constraint:"range=0.01:9999.99"` // 价格范围
    Category    string  `orm:"category" constraint:"in=A:B:C:D"`     // 分类枚举
    Code        string  `orm:"code" constraint:"re=^[A-Z]{3}-\\d{3}$"` // 格式：ABC-123
}
```

## 高级特性

### 视图模式

```go
type User struct {
    ID     int      `orm:"uid key auto" view:"detail,lite"`
    Name   string   `orm:"name" view:"detail,lite"`
    EMail  string   `orm:"email" view:"detail"`           // 仅在detail视图
    Status *Status  `orm:"status" view:"detail"`          // 仅在detail视图
    Group  []*Group `orm:"group" view:"detail"`           // 仅在detail视图
}

// 使用lite视图查询（只包含基础字段）
liteModel := userModel.Copy(models.LiteView)
result, err := o1.Query(liteModel)

// 使用detail视图查询（包含所有字段）
detailModel := userModel.Copy(models.DetailView)
result, err = o1.Query(detailModel)
```

### 关联关系

```go
// 一对一关系
type User struct {
    ID     int     `orm:"uid key auto"`
    Profile *Profile `orm:"profile"`
}

// 一对多关系
type Group struct {
    ID    int      `orm:"gid key auto"`
    Name  string   `orm:"name"`
    Users *[]*User `orm:"users"`
}

// 多对多关系
type User struct {
    ID    int       `orm:"uid key auto"`
    Name  string    `orm:"name"`
    Groups []*Group `orm:"groups"`
}

type Group struct {
    ID    int      `orm:"gid key auto"`
    Name  string   `orm:"name"`
    Users []*User  `orm:"users"`
}
```

### 字段约束规则

1. **可选字段**: 如果字段类型为指针，则表示该字段为可选类型（可为NULL）
2. **复合类型指针**: 对于指针类型的复合类型成员，ORM只处理对象与复合类型成员之间的关系
3. **复合类型**: 对于普通复合类型成员，ORM会同步处理对象与复合类型成员之间的关系
4. **切片类型**: 
   - 不支持基础类型指针切片，如 `[]*bool`, `[]*int`
   - 支持切片指针，如 `*[]bool`, `*[]int`

## 配置选项

### 数据库配置

```go
// PostgreSQL配置
config := &orm.Options{
    Driver: "postgres",
    DSN:    "postgres://user:password@localhost:5432/dbname?sslmode=disable",
    MaxOpenConns: 25,
    MaxIdleConns: 5,
    ConnMaxLifetime: time.Hour,
}

// MySQL配置
config := &orm.Options{
    Driver: "mysql",
    DSN:    "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
    MaxOpenConns: 25,
    MaxIdleConns: 5,
    ConnMaxLifetime: time.Hour,
}
```

## 最佳实践

1. **合理选择数据库**: 根据项目需求选择合适的数据库
2. **合理使用视图**: 对于大对象，使用 `lite` 视图减少数据传输
3. **事务处理**: 对于复杂操作，使用事务确保数据一致性
4. **连接池配置**: 根据应用负载合理配置连接池参数
5. **错误处理**: 始终检查和处理ORM操作返回的错误
6. **约束验证**: 充分利用约束系统进行数据验证
7. **批量操作**: 对于大量数据操作，使用批量查询和插入提高性能

## 示例和参考

完整的使用示例请参考 `test/` 目录下的测试文件：

- `test/simple_test.go` - 基础CRUD操作
- `test/model_local_test.go` - 模型关系示例
- `test/batch_operation_local_test.go` - 批量操作示例
- `orm/builder_postgres_test.go` - PostgreSQL构建器测试

## 许可证

MIT License
