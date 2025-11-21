# Remote 数据结构说明


## Object 数据模型

Object描述数据模型信息，模型信息使用json格式进行序列化保存以及服务间传递

### 数据模型定义格式如下

```go
    type TypeDeclare int

    // Define the Type enum
    const (
        // TypeBooleanValue bool
        TypeBooleanValue = iota + 100
        // TypeBitValue int8
        TypeBitValue
        // TypeSmallIntegerValue int16
        TypeSmallIntegerValue
        // TypeInteger32Value int32
        TypeInteger32Value
        // TypeIntegerValue int
        TypeIntegerValue
        // TypeBigIntegerValue int64
        TypeBigIntegerValue
        // TypePositiveBitValue uint8
        TypePositiveBitValue
        // TypePositiveSmallIntegerValue uint16
        TypePositiveSmallIntegerValue
        // TypePositiveInteger32Value uint32
        TypePositiveInteger32Value
        // TypePositiveIntegerValue uint
        TypePositiveIntegerValue
        // TypePositiveBigIntegerValue uint64
        TypePositiveBigIntegerValue
        // TypeFloatValue float32
        TypeFloatValue
        // TypeDoubleValue float64
        TypeDoubleValue
        // TypeStringValue string
        TypeStringValue
        // TypeDateTimeValue time.Time
        TypeDateTimeValue
        // TypeStructValue struct
        TypeStructValue
        // TypeSliceValue slice
        TypeSliceValue
    )

    type ValueDeclare int

    const (
        Customer ValueDeclare = iota
        AutoIncrement
        UUID
        Snowflake
        DateTime
    )

    type ViewDeclare string

    const (
        OriginView = "origin"
        DetailView = "detail"
        LiteView   = "lite"
    )

    type TypeImpl struct {
        Name        string            `json:"name"`
        PkgPath     string            `json:"pkgPath"`
        Description string            `json:"description"`
        Value       models.TypeDeclare `json:"value"`
        IsPtr       bool              `json:"isPtr"`
        ElemType    *TypeImpl         `json:"elemType"`
    }

    type SpecImpl struct {
        PrimaryKey   bool                `json:"primaryKey"`
        ValueDeclare models.ValueDeclare  `json:"valueDeclare"`
        ViewDeclare  []models.ViewDeclare `json:"viewDeclare"`
        DefaultValue any                 `json:"defaultValue"`
    }

    type ValueImpl struct {
        value any
    }

    type Field struct {
        Name        string    `json:"name"`
        ShowName    string    `json:"showName"`
        Description string    `json:"description"`
        Type        *TypeImpl `json:"type"`
        Spec        *SpecImpl `json:"spec"`
        value       *ValueImpl
    }
```

### 模型信息说明

#### 基本信息

* ID 表示当前模型值对应的唯一标识，ID必须唯一，在模型持久化时使用，ID不能为空，ID不能重复
* Name 表示当前模型值对应的名称，名称必须唯一，名称不能为空，名称不能重复
* ShowName 表示当前模型值对应的显示名称，名称不能为空
* Icon 表示当前模型值对应的图标，图标不能为空
* PkgPath 表示当前模型值对应的包路径，参照http的url规则，在本地使用时可以只包含http url的path部分
* Description 表示当前模型值对应的描述信息

#### Field说明

* Name 表示当前Field值对应的名称，名称必须唯一，名称不能为空，名称不能重复

* ShowName 表示当前Field值对应的显示名称，名称不能为空

* Description 表示当前Field值对应的描述信息

* TypeImpl说明

1. TypeImpl->ElemType 表示当前TypeImpl所依赖的TypeImpl类型 是可选字段，只有在Value的值为TypeStructValue和TypeSliceValue时必须

2. TypeImpl->Value的值为TypeStructValue 表示当前Fileld值是另外一个模型值。此时TypeImpl->ElemType表示当前Fileld值对应的模型

3. TypeImpl->Value的值为TypeSliceValue 表示当前Field值是一个列表，此时TypeImpl->ElemType表示该列表的元素类型

4. TypeImpl->Value的值如果是其他的类型(TypeBooleanValue-TypeDateTimeValue)，则表示当前Field值是一个基本类型，此时TypeImpl->ElemType可以为空

5. TypeImpl->IsPtr 表示当前Field值是否是可选字段，如果IsPtr为True表示模型值对应的Field值可能为空，在业务中不是必选字段

* SpecImpl说明

1. SpecImpl->PrimaryKey 表示当前Field值是否是主键，如果为True则表示当前Field值是主键，否则不是主键

2. SpecImpl->ValueDeclare 表示当前Field值对应的值声明类型，如果为Customer表示当前Field值对应的值声明类型为自定义类型，否则为系统内置类型

3. SpecImpl->ViewDeclare 表示当前Field值对应的视图声明类型，OriginView是默认的视图类型，DetailView表示详细视图类型，LiteView表示精简视图类型，在不指定视图类型时，默认为OriginView

4. SpecImpl->DefaultValue 表示当前Field值对应的默认值，如果为空表示当前Field值没有默认值，否则则使用该值为默认值，在模型对象值对应的Field值为为空时，使用该值为默认值

* ValueImpl说明

1. 在使用Object进行业务处理时，ValueImpl->value 表示当前Field值对应的值，如果为空表示当前Field值没有值，否则则使用该值为值，该值不导出，只做临时数据传递


## ObjectValue 模型对象值

ObjectValue描述模型对象值，模型对象值使用json格式进行序列化保存以及服务间传递

### 模型值格式定义如下

```go
    type FieldValue struct {
        Name  string `json:"name"`
        Value any    `json:"value"`
    }

    type ObjectValue struct {
        ID      string        `json:"id"`
        Name    string        `json:"name"`
        PkgPath string        `json:"pkgPath"`
        Fields  []*FieldValue `json:"fields"`
    }
```

### 模型值信息说明

#### 基本信息

* ID 表示当前模型值对应的唯一标识，ID必须唯一，在模型持久化时使用，ID不能为空，ID不能重复
* Name 表示当前模型值对应的名称，名称必须唯一，名称不能为空，名称不能重复，保持与Object的一致
* PkgPath 表示当前模型值对应的包路径，参照http的url规则，在本地使用时可以只包含http url的path部分，保持与Object的一致
* Fields 表示当前模型值对应的字段列表，Fields不能重复

### FieldValue说明

* Name 表示当前FieldValue对应的名称，名称必须唯一，名称不能为空，名称不能重复，保持与Object的一致
* Value 表示当前FieldValue对应的值，根据模型对应Field的定义，该值允许为空，如果field对应的类型为TypeStructValue，则对应的值为ObjectValue，如果对应的类型为TypeSliceValue，并且ElemType对应的类型为TypeStructValue，则对应的值为SliceObjectValue，否则为对应类型的值


## SliceObjectValue 列表对象值

SliceObjectValue描述列表对象值，列表对象值使用json格式进行序列化保存以及服务间传递


### SliceObjectValue 格式定义如下

```go
    type SliceObjectValue struct {
        Name    string         `json:"name"`
        PkgPath string         `json:"pkgPath"`
        Values  []*ObjectValue `json:"values"`
    }
```

### SliceObjectValue 信息说明

#### SliceObjectValue 基本信息

* Name 表示当前列表对象值对应的名称，名称必须唯一，名称不能为空，名称不能重复，保持与Object的一致
* PkgPath 表示当前列表对象值对应的包路径，参照http的url规则，在本地使用时可以只包含http url的path部分，保持与Object的一致
* Values 表示当前列表对象值对应的值列表, 对应的ObjectValue参照前面的定义


## ObjectFilter 模型对象过滤

ObjectFilter描述模型对象值过滤条件，模型对象值过滤使用json格式进行序列化保存以及服务间传递


### 模型值过滤格式定义如下

```go
    type ObjectFilter struct {
        Name           string         `json:"name"`
        PkgPath        string         `json:"pkgPath"`
        EqualFilter    []*FieldValue  `json:"equal"`
        NotEqualFilter []*FieldValue  `json:"noEqual"`
        BelowFilter    []*FieldValue  `json:"below"`
        AboveFilter    []*FieldValue  `json:"above"`
        InFilter       []*FieldValue  `json:"in"`
        NotInFilter    []*FieldValue  `json:"notIn"`
        LikeFilter     []*FieldValue  `json:"like"`
        MaskValue      *ObjectValue   `json:"maskValue"`
        PageFilter     *pu.Pagination `json:"page"`
        SortFilter     *pu.SortFilter `json:"sort"`

        bindObject *Object
    }
```

### ObjectFilter 信息说明

#### ObjectFilter 基本信息

* Name 被过滤对象值对应的模型名称，保持与Object的一致
* PkgPath 被过滤对象值对应模型的PkgPath 保持与Object的一致

#### ObjectFilter 过滤条件说明

* EqualFilter 筛选指定属性值与过滤条件完全相等的对象，支持多条件组合，多个条件间为逻辑"与"关系

* NotEqualFilter 排除指定属性值与过滤条件相同的对象，支持多属性联合排除，空值视为特殊匹配项

* 范围过滤

 1. BelowFilter 小于过滤，属性值 > 条件值

 2. AboveFilter 大于过滤，属性值 < 条件值


* 集合过滤

 1. InFilter 包含过滤，属性值存在于条件集合

 2. NotInFilter 排除过滤，属性值不在条件集合

* 模糊匹配过滤
 
 1. LikeFilter 模糊匹配过滤，使用通配符进行模式匹配（*代表任意字符）

* MaskValue 字段掩码，控制返回字段的白名单，MaskValue数据格式为ObjectValue

* PageFilter 分页控制，实现将过滤结果分页返回, 当前页(从1开始)，每页记录数，是否返回总记录数

* SortFilter 排序控制，实现将过滤结果按照指定的字段进行排序

* bindObject 动态绑定，临时绑定的过滤值对应的对象模型

* 组合规则

 1. 多条件默认执行逻辑"与"操作

 2. 相同字段的多个条件按最后出现的生效

 3. 空值处理：需显式声明null匹配条件

