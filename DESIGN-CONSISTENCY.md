# magicOrm 数据一致性设计文档

## 1. 项目概述

magicOrm 是一个支持 PostgreSQL 和 MySQL 的 Go ORM 框架，采用双 Provider 架构：

- **Local Provider**：进程内 reflect 操作，保持 Go 原始类型
- **Remote Provider**：跨服务 JSON 传输，需要类型转换

**核心目标**：确保通过 models.Model 方式使用时，无论数据来源（local/remote），数据表现完全一致。

**Provider 约束**：**Local Provider** 与 **Remote Provider** 均**严格遵循** models.Model、models.Field、models.Value、models.Type 等接口定义；所有对模型数据的访问、转换与赋值均通过上述接口进行，**不得绕过或仅部分实现**。对外暴露的 API 与数据转换规则以本规范为准，无歧义。

**文档范围**：本文档**只**描述 **Remote 与 Local 对 models.Model 实现的一致性**，以及**数据转换的一致性**（含类型系统、编解码与转换规则）。**数据库与存储**由**单独文档**描述，本文档**不包含**数据库与存储的任何信息，也不引用其具体内容。

## 2. 架构设计

### 2.1 核心抽象

```go
// 当前统一的模型抽象（models/model.go）
type Model interface {
    GetName() string
    GetShowName() string
    GetPkgPath() string
    GetPkgKey() string
    GetDescription() string
    GetFields() Fields
    SetFieldValue(name string, val any) *cd.Error
    SetPrimaryFieldValue(val any) *cd.Error
    GetPrimaryField() Field
    GetField(name string) Field
    Interface(ptrValue bool) any
    Copy(viewSpec ViewDeclare) Model
    Reset()
}
```

其中 `Model` 由不同 Provider **完整实现**（严格遵循接口，不遗漏、不绕过）：

- **Local Provider**：`provider/local/object.go` 中的 `objectImpl`，基于 `reflect.Value` 包装原生结构体，对外仅通过 Model/Field/Value 暴露数据。
- **Remote Provider**：`provider/remote/object.go` 中的 `Object`，基于 JSON 友好的结构定义和字段集合，对外仅通过 Model/Field/Value 暴露数据。

### 2.2 Provider 架构

```
┌─────────────────────────────────────────────┐
│                models.Model                 │
├─────────────────────────────────────────────┤
│           Local Provider                    │
│  ┌─────────────┐  ┌─────────────┐          │
│  │   Object    │  │ ObjectValue │          │
│  │ (reflect)   │  │ (reflect)   │          │
│  └─────────────┘  └─────────────┘          │
├─────────────────────────────────────────────┤
│           Remote Provider                   │
│  ┌─────────────┐  ┌─────────────┐          │
│  │   Object    │  │ ObjectValue │          │
│  │ (JSON def)  │  │ (JSON data) │          │
│  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────┘
```

Local 侧图中 Object/ObjectValue 为**逻辑等价**，实际实现为 `objectImpl` + reflect，无独立 Object 结构体（见 2.1）。**数据访问与转换**：无论 Local 还是 Remote，调用方对模型数据的读写、转换必须通过 **models.Model / models.Field / models.Value** 进行，禁止直接操作 reflect 或 Object/ObjectValue 的底层表示。

### 2.3 Field（字段抽象）

**models.Field** 用于对**对象结构中的字段**做统一抽象，屏蔽 Local 与 Remote 的实现差异，对上暴露一致接口。通过 Field 可获取或设置：

- **字段类型**（Type）：该字段的类型信息（对应 TypeDeclare、指针/切片等），与 4.x 类型系统一致。
- **当前值**（Value）：该字段的当前取值，以 models.Value 表示，便于在不同 Provider 间统一读写。
- **Spec 声明**：字段的元声明（如 orm tag 解析结果），包含字段名、是否主键、值生成方式（auto/uuid/datetime 等）、视图（ViewDeclare）等。

这样，无论数据来自 Local（reflect）还是 Remote（Object/FieldValue），上层仅依赖 **Field** 即可按“类型 + 值 + Spec”操作字段，无需关心底层是反射还是 JSON 结构。

### 2.4 Value（字段值抽象）

**models.Value** 是**字段值的抽象**，定义对字段值的访问行为，屏蔽 Local 与 Remote 的取值/存值差异：

- **Get / Set**：读写当前值；**Get 与 Set 均使用 Local 或 Remote 对应的实际值**（Local 侧为 reflect 取到的 Go 值，Remote 侧为 ObjectValue/FieldValue 中存储的可序列化值）。
- **TypeDateTimeValue 的接口约定**：为保持通过 **models.Model 及对应接口**访问时 Local 与 Remote 数据一致，**TypeDateTimeValue 类型字段在 Model/Field/Value 接口层以 string 传递**（Get 返回 string、Set 接受 string）。Local 内部仍使用 `time.Time` 与 reflect 传递，在**接口边界**（如 Get/Set、SetFieldValue 等）做 string ↔ `time.Time` 转换；Remote 侧接口层即为 string，无需额外转换。
- **UnpackValue**：若该值为 **slice 类型**，可通过 **UnpackValue** 将整体拆解为 **Value 的 slice**（即 `[]Value`），便于逐元素访问或遍历，再结合各元素的 Get/Set 与类型信息做编解码。

这样，上层通过 Value 的 Get/Set/UnpackValue 即可统一处理标量与切片，无需区分底层是 Local 还是 Remote 实现。

## 3. Remote Provider 核心概念

在 Remote 中：**Object 表示数据结构**，**ObjectValue 表示数据值**。二者通过 **Name** 与 **PkgPath** 关联：**Name 和 PkgPath 相同则表示同一数据结构、以及该结构下的同一种数据值**（即该 ObjectValue 是按该 Object 定义的结构所填的一组值）。

### 3.1 Object（数据结构）

Remote 的 **Object** 是 **models.Model** 的实现（见 2.1），**严格遵循** Model 接口；其字段为 **models.Field** 实现，类型为 **models.Type**。

描述结构体的元数据，包含：

- 名称（Name）、包路径（PkgPath）
- **字段定义（Fields）**：每个元素为 Remote 的 **Field**，遵循 **models.Field** 接口定义，包含 **Type**（字段类型）、**Name**（字段名）、**Spec**（声明，如主键、自增、视图等）等信息。

**用途**：定义数据结构，不包含实际数据值。

### 3.2 ObjectValue（数据值）

存储该结构下的实际数据值，包含：

- 名称（Name）、包路径（PkgPath）：与 Object 一致，用于标识对应的数据结构
- **字段值（Fields）**：每个元素为 **FieldValue**，定义**字段名称**与**该字段的取值**；**字段值采用 JSON 友好类型**（如 string、number、bool、嵌套对象/数组），便于跨服务序列化与传输
- 切片数据（SliceObjectValue 用于一对多关系）

**用途**：存储和传输实际数据。与 Object 通过 **Name + PkgPath** 对应：同一 Name、PkgPath 即同一数据结构及其数据值类型。

**跨服务 JSON 结构**（与 provider/remote 序列化一致）：**ObjectValue** 序列化包含 `name`、`pkgPath`、`fields`（字段名与值的数组）；**SliceObjectValue** 包含 `name`、`pkgPath`、`values`（ObjectValue 数组）。便于跨服务传输与反序列化。

### 3.3 Field 与 FieldValue（字段类型与字段值）

在 Remote 中，**Object 的 Field** 表示字段类型定义，包含 **Type、Name、Spec**，遵照 **models.Field** 接口；**ObjectValue 的 FieldValue** 表示该字段的名称及其当前值，**值采用 JSON 友好类型**。二者通过 **Name** 匹配：Name 相同即同一字段的类型定义与其取值。Object 的 Fields 为 Field 列表，ObjectValue 的 Fields 为 FieldValue 列表；按 Name 一一对应后，每个 FieldValue 的值需符合其对应 Field 的 Type 约定（如 TypeDateTimeValue 对应 string、TypeBooleanValue 对应 bool）。

### 3.4 关键区别


| 维度       | Object    | ObjectValue |
| -------- | --------- | ----------- |
| **内容**   | 元数据定义     | 实际数据值       |
| **序列化**  | 结构定义      | 数据内容        |
| **变化频率** | 低频（结构变化时） | 高频（数据操作时）   |
| **大小**   | 较小        | 可能很大        |


Field 与 FieldValue 的关系同理：Field 为字段类型，FieldValue 为字段值，以 **Name** 匹配。

### 3.5 Remote 跨服务解码与赋值约定

- **解码不强制转成 Go 类型**：Remote 在进行跨服务通讯时，解码（unmarshal）**不强制**将值转换为 Go 原生类型（如不要求解码后即为 `time.Time`、`bool` 等）；可保持 JSON/跨服务友好形式（如 string、number），以利于传输与序列化一致性。
- **仅在 Remote→Local 时做类型转换**：将“跨服务形式”转为 Go 类型（如 string→`time.Time`、数值→`bool`）**只在 Remote→Local 边界**进行（如 `UpdateEntity` 中经 `local.DecodeValue`、`SetValue` 等）。
- **通过 ObjectValue 给 Object 赋值**：对 Remote 的 Object 写入数据时，**必须通过 ObjectValue 进行赋值**（即用 ObjectValue 作为数据源赋给 Object）。赋值时：
  - 按 **Name** 进行字段匹配，将 ObjectValue 的每个 FieldValue 对应到 Object 的同名字段（Field）。
  - 要求**对应的值与该字段类型兼容**（如 TypeDateTimeValue 对应符合 DateTime 格式的 string，TypeBooleanValue 对应 bool 或可解析为 bool 的表示）。
  - 须**通过 Validate 校验**（类型兼容性、数值/格式合法、与类型匹配等）；**不符合或校验未通过则赋值失败**，不写入该字段或返回错误。校验发生在**“ObjectValue 赋给 Object”的赋值阶段**。
  - **赋值路径**（二者必居其一，无例外）：
    - **Provider 内部**：指 provider/remote 包内实现。直接采用 **ObjectValue → Object**，对各字段**遍历赋值**（按 Name 匹配、类型兼容与 Validate 后写回 Object 的对应 Field）。
    - **Provider 外部**：指所有非上述内部的调用方。**必须**通过 **models.Model** 进行赋值（如 `SetModelValue(vModel, vVal)`），**不得**直接操作 Object/ObjectValue 的字段或内部结构，以保证统一走 Model 接口与校验。
- **转换到 Local 时做解码与校验**：上述约定使得在 **Remote→Local** 转换时，可在此处做**解码**（将 Remote 值转为 Go 类型）及与类型匹配的**合法性校验**，再写入 Local。

## 4. 数据类型系统

### 4.1 TypeDeclare 枚举（models/const.go）

```go
type TypeDeclare int

const (
    // 基础数值类型（见 models/const.go）
    TypeBooleanValue TypeDeclare = iota + 100 // bool
    TypeByteValue                             // int8
    TypeSmallIntegerValue                     // int16
    TypeInteger32Value                        // int32
    TypeIntegerValue                          // int
    TypeBigIntegerValue                       // int64
    TypePositiveByteValue                     // uint8
    TypePositiveSmallIntegerValue             // uint16
    TypePositiveInteger32Value                // uint32
    TypePositiveIntegerValue                  // uint
    TypePositiveBigIntegerValue               // uint64
    TypeFloatValue                            // float32
    TypeDoubleValue                           // float64
    TypeStringValue                           // string
    TypeDateTimeValue                         // time.Time / datetime

    // 复合类型
    TypeStructValue                           // struct
    TypeSliceValue                            // slice/array
)
```

### 4.2 models.Type（类型抽象）

**models.Type** 根据实际类型表示字段类型信息，与 TypeDeclare 配合使用：

- **IsPtr**：若该字段为**指针类型**（如 `*int`、`*string`），则 **IsPtr 为 true**；否则为 false。
- **Elem**：用于标识该类型是否为 **slice**，并返回元素类型：
  - 若字段类型为 **slice**（如 `[]T`、`[]*T`），则 **Elem() 返回的是 slice 里单个元素的类型**（即 `T` 或 `*T` 对应的 Type）；
  - 非 slice 时，**Elem() 返回自身**。

这样，通过 Type 的 **GetValue()（TypeDeclare）**、**IsPtr**、**Elem()** 即可区分标量、指针、切片及元素类型，与 4.1 的 TypeDeclare 枚举一致。

### 4.3 类型包装

- **指针类型**：`*T` 包装为 TypeDeclare + **IsPtr = true**
- **切片类型**：`[]T` 包装为 TypeDeclare（TypeSliceValue）+ **Elem() 为元素类型 T**
- **复合类型**：struct/slice 需要递归处理

### 4.4 Local 侧 TypeDeclare 与 ViewDeclare 的来源

在 **Local** 中，**TypeDeclare** 与 **ViewDeclare** 是根据 **Go 结构体字段上声明的 tag** 解析得到的，用于标识该字段的**类型**以及**在各类视图下的特殊语义/用途**（如主键、自增、视图可见性等）。

- **TypeDeclare**：由 Go 类型（reflect）结合 tag（如 `orm`）得到，用于标识字段类型（含基础类型、指针、切片等）以及值生成方式（如 `key`、`auto`、`uuid`、`datetime`）。
- **ViewDeclare**：由字段上的 **view** tag 解析得到（如 `detail`、`lite`），表示该字段在哪些视图（OriginView、DetailView、LiteView 等）下参与序列化或展示，用于区分“完整/详细/精简”等不同数据视图。

Remote 侧的结构定义（Object/Field）在跨服务协商或从 Local 推导时，会与上述 tag 约定对齐，保证同一结构在 Local 与 Remote 上语义一致。

## 5. 编解码规范

### 5.1 Local Provider（保持 Go 原始类型）

- **编码**：reflect 获取原始值，不做类型转换
- **解码**：直接赋值到目标变量
- **特点**：零转换开销，类型安全
- **与 Model 接口的边界**：通过 **models.Model / Field / Value** 访问 TypeDateTimeValue 时，**接口层统一以 string 传递**，以与 Remote 一致；Local **内部**仍使用 `time.Time` 与 reflect，在 Get/Set、SetFieldValue 等边界处做 string ↔ `time.Time` 转换。

### 5.2 Remote Provider（JSON 友好）

- **编码**：转换为 JSON 友好类型，类型语义清晰
- **解码**：跨服务解码**不强制**转为 Go 类型，可保持 JSON 友好形式；**类型转换仅在 Remote→Local 时进行**（见 3.5）。
- **特点**：跨服务兼容；给 Object 赋值须通过 ObjectValue，以便在 Remote→Local 时统一做解码与校验。
- **时间类型**：在 Remote 编码和解码时，类型均标记为 **TypeDateTimeValue**，**值一律为 string**（便于跨服务 JSON）。通过 **models.Model 及对应接口**访问时，TypeDateTimeValue 在接口层也以 string 传递（与 Local 接口约定一致）。**仅在 Remote→Local 写回本地实体时**（如 `UpdateEntity` 经 `local.DecodeValue`）在 Local 内部将 string 转成 `time.Time`。
- **布尔类型**：针对 **TypeBooleanValue**，Remote 与 Local **均使用 bool**（编码、解码的值类型一致）。

### 5.3 编解码映射表


| Go 类型       | Local 编码    | Remote 编码 | Remote 解码    |
| ----------- | ----------- | --------- | ------------ |
| `bool`      | `bool`      | `bool`    | `bool`       |
| `int8`      | `int8`      | `int8`    | `int8`       |
| `int16`     | `int16`     | `int16`   | `int16`      |
| `int32`     | `int32`     | `int32`   | `int32`      |
| `int64`     | `int64`     | `int64`   | `int64`      |
| `uint8`     | `uint8`     | `uint8`   | `uint8`      |
| `uint16`    | `uint16`    | `uint16`  | `uint16`     |
| `uint32`    | `uint32`    | `uint32`  | `uint32`     |
| `uint64`    | `uint64`    | `uint64`  | `uint64`     |
| `float32`   | `float32`   | `float32` | `float32`    |
| `float64`   | `float64`   | `float64` | `float64`    |
| `string`    | `string`    | `string`  | `string`     |
| `time.Time` | `time.Time` | `string`  | `string`（见下） |


**说明**：Remote 层对 DateTime 的约定——**编码与解码时类型均为 TypeDateTimeValue，值均为 string**；表中「Remote 解码」在 Remote 内部仍为 string，**只有到 Remote→Local 写回时**（如 `UpdateEntity` / `local.DecodeValue`）才转为 `time.Time`。

**说明（Boolean）**：TypeBooleanValue 在设计中 Remote 与 Local 均使用 bool。

### 5.4 指针与切片、Remote 解码值类型

- **指针（*T）**：按元素类型 T** 编解码；`nil` 在编解码与赋值时按空值处理。编码时若为指针则产出 T 或 *T（由实现决定，Remote 侧常为 JSON 可表示形式）；解码时与 5.3 表中 T 的约定一致。
- **切片（[]T）**：按**元素类型 T** 逐元素编解码，与 4.3 类型包装及 5.3 映射表一致；复合类型（struct/slice）递归为 ObjectValue、SliceObjectValue。
- **成员类型为 []*T（指针切片）的补充说明**：实际业务中可以**不考虑** slice 内 **item 值为 nil** 的场景（即不要求支持或测试元素为 nil 的 []*T）。但必须保证：**成员类型为 []*T 的对象**在 **Local↔Remote 相互转换**（GetObjectValue / UpdateEntity、GetSliceObjectValue / UpdateSliceEntity 及 JSON 往返）中**正确无误**；且该字段在 **Model、Field、Type、Value** 上的表现均符合预期——即通过 Model.GetField 得到对应 Field，Field.GetType() 为切片类型（IsSlice 为 true、Elem 为元素类型），Field.GetValue() 与 Value.Get()/UnpackValue() 与设计一致，往返后数据与起点一致。
- **Remote 解码后的值类型**：与 5.3 表一致——标量对应 JSON 友好类型（如 number→float64/int、string、bool）；**TypeDateTimeValue** 为 **string**；嵌套结构为 **ObjectValue**，切片为 **SliceObjectValue**（其元素为 ObjectValue）。即 Remote 侧解码后不强制转为 Go 原生类型，保持跨服务可序列化形式（见 3.5）。

### 5.5 数据转换与处理规则摘要（无歧义约定）

以下为数据转换与处理的**统一约定**，Local 与 Remote 均按此执行，无例外。


| 维度                | 约定                                                                                                                                                                                     |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **接口边界**          | 凡通过 **models.Model / models.Field / models.Value** 暴露或传递的数据，均以本规范为准。**TypeDateTimeValue** 在接口层**一律为 string**（Get 返回 string、Set 接受 string）；其余类型与 5.3、5.4 一致。                            |
| **Local 内部**      | 使用 **time.Time** 与 **reflect** 表示与传递；仅在**接口边界**（如 Get/Set、SetFieldValue）做 TypeDateTimeValue 的 string ↔ `time.Time` 转换。编解码按 5.1、5.3、5.4。                                                |
| **Remote 内部**     | 使用 **Object / ObjectValue**，值为 **JSON 友好类型**（见 5.3、5.4）。对 Object 的写入**必须**通过 ObjectValue 赋值（3.5）；外部调用方**必须**通过 **SetModelValue(vModel, vVal)** 等 Model 接口，不得直接写 Object/ObjectValue 字段。 |
| **跨 Provider 转换** | **仅**通过 **provider/helper** 的 GetObject、GetObjectValue、GetSliceObjectValue、UpdateEntity、UpdateSliceEntity 以及 **models.Model** 相关接口；数据语义与类型以 5.x、6.x 为准。                                |


**合法值与校验**：ObjectValue 赋给 Object 时，值须与字段类型兼容且通过 **Validate**（6.5）；不通过则赋值失败。

## 6. 特殊类型规范（编解码与类型语义）

以下仅约定**类型语义与编解码**。

### 6.1 时间类型（DateTime）

- **接口层约定**：为保持 Local 与 Remote 通过 **models.Model 及对应接口**访问数据一致，**TypeDateTimeValue 类型字段在 Model/Field/Value 的 Get/Set 上统一以 string 传递**。Local 在接口边界做 string ↔ `time.Time` 转换，内部仍用 `time.Time` 与 reflect；Remote 接口层即为 string。
- **编解码**：
  - **Local 内部**：类型与值均为 `time.Time`，reflect 直接读写。
  - **Remote**：编码与解码时**类型均标记为 TypeDateTimeValue，值均为 string**（便于跨服务 JSON）；通过 Model 接口访问时同样以 string 传递。
  - **Remote→Local 写回**：如 `UpdateEntity` 中经 `local.DecodeValue` 将 string 转成 `time.Time` 写入本地实体。

### 6.2 布尔类型（Boolean）

- **编解码**：针对 **TypeBooleanValue**，**Remote 与 Local 均使用 bool**（类型标记为 TypeBooleanValue，编码与解码的值均为 `bool`）。

### 6.3 指针类型

- **空值处理**：`nil` 指针在编解码与赋值时按空值处理。
- **编解码**：递归处理指向的类型。
- **约束**：支持所有基本类型的指针。

### 6.4 切片类型

- **一对多关系**：使用 `SliceObjectValue` 表示。
- **编解码**：递归处理元素类型。

### 6.5 合法值与校验

通过 ObjectValue 赋给 Object 时的**合法值**以 **Validate** 校验为准（见 3.5）：类型兼容、数值/格式合法、与 TypeDeclare 匹配。例如 TypeDateTimeValue 要求符合 DateTime 格式的 string，TypeBooleanValue 为 bool（或可解析为 bool 的表示）；其余基础类型遵循 Go/JSON 惯例，具体范围与格式以实现为准（如 5.3、5.4）。

## 7. 关键转换函数

以下 7.1～7.3 的对外入口及 7.4 所述实现均位于 **provider/helper**（主要为 `remote.go`），在遵循 **models.Model** 的前提下完成 Local 与 Remote 的互转。

### 7.1 数据结构获取（Local/任意实体 → Remote 结构或值）

```go
// 获取远程数据结构定义（基于 struct 标签）
func GetObject(entity any) (*remote.Object, *cd.Error)

// 获取远程单个数据值（结构体实例 → ObjectValue）
func GetObjectValue(entity any) (*remote.ObjectValue, *cd.Error)

// 获取远程切片数据值（切片实例 → SliceObjectValue）
func GetSliceObjectValue(sliceEntity any) (*remote.SliceObjectValue, *cd.Error)
```

### 7.2 远程编解码

上述 Encode/Decode 的**实现在 provider/remote** 包（ObjectValue/SliceObjectValue 的 JSON 序列化）；第 7 章的 Local↔Remote 转换入口在 provider/helper。

```go
// ObjectValue <-> JSON 字节流
func EncodeObjectValue(objVal *ObjectValue) ([]byte, *cd.Error)
func DecodeObjectValue(data []byte) (*ObjectValue, *cd.Error)

// SliceObjectValue <-> JSON 字节流
func EncodeSliceObjectValue(objVal *SliceObjectValue) ([]byte, *cd.Error)
func DecodeSliceObjectValue(data []byte) (*SliceObjectValue, *cd.Error)
```

### 7.3 远程 → 本地（写回实体）

```go
// 远程单对象值 → 本地实体
func UpdateEntity(remoteValuePtr *remote.ObjectValue, localEntity any) *cd.Error

// 远程切片对象值 → 本地切片
func UpdateSliceEntity(remoteSliceValuePtr *remote.SliceObjectValue, localSliceValue any) *cd.Error
```

### 7.4 provider/helper 与 models.Model 的转换约定

Local 与 Remote 的互转在 **provider/helper** 中完成，设计上**统一遵循 models.Model 与 models.Field 的接口**，保证两边数据在“模型视图”下一致。

#### 7.4.1 设计原则（严格遵循 models 接口）

- **写回本地**：**仅**通过 `models.Model` / `models.Field`（如 SetFieldValue、SetValue、AppendSliceValue）操作；**禁止**直接写 reflect 或目标结构体字段。
- **本地读出到远程表示**：结构信息与 `models.Type`/字段定义一致，值通过 Local 的编解码语义（5.1、5.3、5.4）生成 Remote 可序列化表示；**禁止**绕过 Model/Field/Value 直接读取 reflect 或拼装 ObjectValue。
- **TypeDateTimeValue 在接口层统一为 string**：通过 Model/Field/Value 访问 DateTime 字段时，Get/Set 均以 string 传递，保证 Local 与 Remote 行为一致；Local 内部在边界做 string ↔ `time.Time` 转换。
- 上述原则对 **Local Provider** 与 **Remote Provider** 均适用，与 2.x、5.5 一致。

#### 7.4.2 Local → Remote（helper 产出 Object / ObjectValue）


| 函数                                     | 位置               | 行为与 models 关系                                                                                                                                                                                                                                                                |
| -------------------------------------- | ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GetObject(entity any)`                | helper/remote.go | 由实体 `reflect.Type` 经 `type2Object` 得到 `*remote.Object`，并 `models.VerifyModel(objectPtr)`，保证结构满足 Model 约束。                                                                                                                                                                    |
| `GetObjectValue(entity any)`           | helper/remote.go | 若实体已是 `remote.Object`/`ObjectValue`，用 `Interface(true)` 取 `*remote.ObjectValue`；否则用 `getObjectValue(reflect.Value)` 按字段遍历，**基础类型** 用 `local.EncodeValue(fieldVal, itemType)` 生成可序列化值，**嵌套 struct/slice** 递归为 `ObjectValue`/`SliceObjectValue`。即：本地实体 → 符合 Type 定义的 Remote 值。 |
| `GetSliceObjectValue(sliceEntity any)` | helper/remote.go | 对切片逐元素调用 `getObjectValue`，得到 `*remote.SliceObjectValue`，元素类型与 `models.Type` 一致。                                                                                                                                                                                              |


内部依赖：

- `getFieldValue(fieldName, itemType, itemValue)`：按 `models.IsBasic` / `models.IsSlice` / 否则 struct 分支，基础类型走 **local.EncodeValue**，复合类型递归为 ObjectValue/SliceObjectValue。

#### 7.4.3 Remote → Local（helper 只通过 Model/Field 写回）


| 函数                                                          | 位置               | 行为与 models 关系                                                                                                                                                                                                                                                    |
| ----------------------------------------------------------- | ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UpdateEntity(remoteValuePtr, localEntity)`                 | helper/remote.go | 先 `local.GetEntityModel(localEntity, nil)` 得到 `models.Model`，再 `updateLocalModel(remoteValuePtr, localModel)`。全程仅用 Model 接口。                                                                                                                                     |
| `updateLocalModel(remoteValuePtr, localModel models.Model)` | helper/remote.go | 遍历 `remoteValuePtr.Fields`，用 `localModel.GetField(fieldValue.Name)` 取 `models.Field`，按 `models.IsBasicField` / `IsSliceField` / `IsStructField` 分支：基础类型用 **local.DecodeValue(fieldValue.Get(), localField.GetType())** 后 **localField.SetValue(rVal)**；切片/结构体见下。 |
| `updateSliceStructField(val, localField models.Field)`      | helper/remote.go | `localField.Reset()` 后对每个 `*remote.SliceObjectValue` 元素：`local.GetValueModel(localSubVal)` 得子 Model，`updateLocalModel(objectValuePtr, localSubModel)` 填值，再 **localField.AppendSliceValue(localSubModel.Interface(elemType.IsPtrType()))**。仅用 Field/Model 接口。     |
| `updateStructField(val, vField models.Field)`               | helper/remote.go | `local.GetValueModel(localFileVal)` 得子 Model，`updateLocalModel(objectValuePtr, localModelVal)` 填值，再 **vField.SetValue(localModelVal.Interface(elemType.IsPtrType()))**。                                                                                          |
| `UpdateSliceEntity(remoteSliceValuePtr, localSliceValue)`   | helper/remote.go | 对切片每个元素 `local.GetValueModel(localItemVal)` + `updateLocalModel(val, localItemModel)`，再 `localValuePtr.Append(...)` 写回切片。                                                                                                                                        |


要点：Remote → Local 的写入路径**不直接操作 reflect**，全部通过 `GetEntityModel` / `GetValueModel` 得到的 `models.Model` 以及 `GetField`、`SetValue`、`GetType`、`Interface`、`AppendSliceValue` 等 models 接口完成，从而与“遵循 models.Model 定义”的设计一致。

#### 7.4.4 编解码通用辅助（codec_helper）


| 函数                       | 位置                     | 用途                                                                                                                                |
| ------------------------ | ---------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `EncodeSliceTemplate[T]` | helper/codec_helper.go | 泛型 slice 编码模板，接收 `encodeValueFunc(reflect.Value, models.Type) (any, *cd.Error)`，在遵循 `models.Type` 的前提下对切片逐元素编码，供 local/remote 复用。 |
| `DecodeSliceValue`       | helper/codec_helper.go | 泛型 slice 解码模板，接收 `decodeValueFunc(any, models.Type) (any, *cd.Error)`，按 `models.Type` 逐元素解码。                                      |


二者均以 **models.Type** 为类型依据，保证与类型系统一致。

#### 7.4.5 通过 ObjectValue 给 Object 赋值的 API 与调用链（对应 3.5）

对 Remote 的 Object 写入数据时，必须通过 ObjectValue 进行赋值。以下为当前实现中的 API 与内部调用关系。

**（1）Provider 外部：经 models.Model 统一入口**


| API                                                               | 位置                          | 说明                                                                                                                                                                                                                                              |
| ----------------------------------------------------------------- | --------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `provider.SetModelValue(vModel, vVal) (ret, err)`                 | provider/provider.go        | 统一接口：`vModel` 为 `*remote.Object` 时由 remote 实现；`vVal` 需为包装了 `*ObjectValue` 的 `models.Value`，可用 `remote.NewValue(objVal)` 得到。                                                                                                                     |
| `remote.SetModelValue(vModel, vVal, disableValidator) (ret, err)` | provider/remote/provider.go | Remote 实现。当 `vVal.Get()` 为 `*ObjectValue` 时，调用内部 `assignObjectValue(vObjectPtr, objectValuePtr, disableValidator)`，按 ObjectValue 的 Fields 遍历，将每个字段写回 Object；否则将 `vVal` 视为主键单值，调用 `vObjectPtr.innerSetPrimaryFieldValue(val, disableValidator)`。 |
| `remote.NewValue(val any) *ValueImpl`                             | provider/remote/value.go    | 将 `*ObjectValue`（或标量、`*SliceObjectValue` 等）包装为 `models.Value`，供 `SetModelValue` 使用。                                                                                                                                                             |


调用链（外部）：`remote.NewValue(objValue)` → 得到 `models.Value` → `provider.SetModelValue(remoteObjectModel, vVal)` → 内部 `remote.SetModelValue` → `assignObjectValue` → 对每个字段 `Object.innerSetFieldValue(name, val, disableValidator)`。

**（2）Provider 内部：ObjectValue → Object 直接赋值**


| API                                                                                         | 位置                          | 说明                                                                                                                                                                        |
| ------------------------------------------------------------------------------------------- | --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `assignObjectValue(vObjectPtr *Object, objectValuePtr *ObjectValue, disableValidator bool)` | provider/remote/provider.go | **未导出**。遍历 `objectValuePtr.Fields`，对每个 FieldValue 调用 `vObjectPtr.innerSetFieldValue(fieldVal.GetName(), fieldVal.Get(), disableValidator)`，完成整份 ObjectValue 到 Object 的写入。 |
| `(s *Object) SetFieldValue(name string, val any) *cd.Error`                                 | provider/remote/object.go   | 对 Object（即 Model）**单字段**赋值；内部调用 `innerSetFieldValue(name, val, false)`，**启用校验**。                                                                                          |
| `(s *Object) innerSetFieldValue(name, val any, disableValidator bool) *cd.Error`            | provider/remote/object.go   | **未导出**。按 Name 匹配 Object 的 Field，再按类型分支：基础类型 `setBasicFileValue`、切片 `setSliceStructValue`、结构体 `setStructValue`；若 `disableValidator == false` 则经校验后再写入。                    |


**（3）ObjectValue 自身（仅改数据容器，不写回 Object）**


| API                                                      | 位置                        | 说明                                                                |
| -------------------------------------------------------- | ------------------------- | ----------------------------------------------------------------- |
| `(s *ObjectValue) SetFieldValue(name string, value any)` | provider/remote/object.go | 仅设置 ObjectValue 的字段值，用于构造或修改待赋给 Object 的 ObjectValue；不会写回 Object。 |


**小结**：**外部**（非 provider/remote 包内）必须通过 **models.Model** 赋值（`SetModelValue(vModel, remote.NewValue(objVal))` 或 provider 暴露的等价入口），**不得**直接写 Object 或 ObjectValue 的字段。**内部**（provider/remote 包内）可直接用 **ObjectValue → Object** 的遍历赋值（`assignObjectValue`）或单字段 `Object.SetFieldValue`。校验在 `innerSetFieldValue` / `setBasicFileValue` 等内部执行，`disableValidator == true` 时跳过。与 3.5、5.5 一致，无例外。

## 8. 一致性要求

### 8.1 数据一致性

1. **值相等**：相同数据在不同 provider 中表现一致
2. **类型安全**：类型转换不丢失精度
3. **空值处理**：`nil`/`NULL` 处理一致

### 8.2 操作一致性

1. **CRUD 操作**：相同语义，相同结果
2. **事务处理**：ACID 特性保持一致
3. **错误处理**：错误通过 **cd.Error**（`magicCommon/def`）返回；错误码如 `Unexpected`、`IllegalParam` 等由 cd 包定义。「错误类型和消息一致」指同一场景返回同类错误（相同错误码或分类）。

### 8.3 性能一致性

1. **响应时间**：相同操作耗时相近
2. **内存使用**：数据表示内存占用合理
3. **网络开销**：remote provider 序列化大小优化

具体量化指标（如响应时间、内存阈值）**待压测或运维约定后确定**。

### 8.4 实现级往返一致性（必达）

以下两条为**实现上的硬性要求**，必须满足且可被测试验证。

#### 8.4.1 Local → Remote → (marshal/unmarshal) → Remote → Local 完全一致

```
Local 实体
   → GetObjectValue / GetSliceObjectValue  →  Remote ObjectValue / SliceObjectValue
   → EncodeObjectValue / EncodeSliceObjectValue  →  []byte (marshal)
   → DecodeObjectValue / DecodeSliceObjectValue  →  Remote ObjectValue / SliceObjectValue (unmarshal)
   → UpdateEntity / UpdateSliceEntity  →  Local 实体（写回）
```

**要求**：最终 Local 实体与原始 Local 实体**在模型视图下完全一致**（值相等、类型无损、空值处理一致）。即经 marshal/unmarshal 后，Remote → Local 写回的结果必须与起点一致。

#### 8.4.2 Remote → (marshal/unmarshal) → Remote 完全一致

```
Remote Object / ObjectValue / SliceObjectValue
   → marshal (JSON 序列化)
   → unmarshal (JSON 反序列化)
   → Remote Object / ObjectValue / SliceObjectValue
```

**要求**：

- **Object**：marshal/unmarshal 前后 `*remote.Object` 完全一致（结构、字段定义、类型信息不变）。
- **ObjectValue / SliceObjectValue**：marshal/unmarshal 前后 `*remote.ObjectValue` / `*remote.SliceObjectValue` 完全一致（字段名、字段值、嵌套结构、切片元素等保持不变）。

即 Remote 表示本身经序列化/反序列化后必须**自洽且可逆**，不因 JSON 编解码产生语义或结构偏差。

## 9. 测试验证

### 9.1 测试目录结构

```
test/consistency/
├── codec_consistency_test.go      # 编解码一致性测试
├── roundtrip_test.go              # 往返转换测试
├── entity_conversion_test.go      # 实体转换测试
├── slice_conversion_test.go       # 切片转换测试
├── nested_struct_test.go          # 嵌套结构测试
├── edge_cases_test.go             # 边界情况测试
└── json_serialization_test.go     # JSON 序列化测试
```

### 9.2 测试策略

1. **单元测试**：每个类型单独测试
2. **集成测试**：完整实体转换测试
3. **边界测试**：空值、极值、特殊字符
4. **性能测试**：编解码性能对比
5. **往返一致性测试**（对应 8.4）：
  - **Local→Remote→marshal/unmarshal→Remote→Local**：验证 `GetObjectValue` / `UpdateEntity` 等链路的往返后 Local 与起点完全一致
  - **Remote→marshal/unmarshal→Remote**：验证 `Object` / `ObjectValue` / `SliceObjectValue` 经 `Encode`* / `Decode*` 后与起点完全一致

## 10. 待修复问题

### 10.1 已知不一致问题

（**Boolean 编解码**：已按 10.2 修复。Remote 现与 Local 一致，TypeBooleanValue 编码/解码均为 `bool`/`[]bool`。）

（**DateTime**：Remote 对 `time.Time` 采用 string 编码为既定设计，见 5.2、6.1，不列为待修复项。）

### 10.2 修复方案（Boolean 已实施）

1. ~~修改 remote `encodeValue`/`encodeValueConvertMap`，确保 Boolean 返回 `bool`~~ **已实施**：`provider/remote/codec.go` 中 `encodeValueConvertMap`、`encodeValueConvertSliceMap` 的 TypeBooleanValue 已改为使用 `utils.ConvertToBool` 与 `encodeSliceTemplate(..., false)`，产出 `bool`/`[]bool`。
2. 相关测试用例已通过（test/consistency、provider/remote）。

## 11. 后续开发计划

### Phase 1：设计确认

- 确认本设计文档内容（仅 models.Model 实现一致性与数据转换一致性）
- 确定特殊类型处理规则

### Phase 2：代码修复

- [x] 修复 Boolean 编解码不一致（已实施）
- 更新相关辅助函数（如无需则跳过）

### Phase 3：测试验证

- 运行现有测试套件
- 添加边界情况测试
- 验证跨数据库兼容性

### Phase 4：文档更新

- 更新 API 文档
- 添加使用示例
- 更新 README

---

**最后更新**：2026-02-28  
**版本**：1.4.0  
**状态**：本文档只描述 Remote/Local 对 models.Model 实现的一致性及数据转换一致性；数据库与存储由单独文档描述，本文档不包含其任何信息