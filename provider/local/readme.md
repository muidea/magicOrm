# MagicORM Struct转换模块说明

## 功能概述

本模块通过反射机制，将包含基础数据类型的Go结构体转换为ORM模型定义（model.Model对象），转换结果遵循magicOrm/model目录下的模型规范。

## 转换规则详解

### 一、结构体转Model对象

1. **模型命名规则**
   - `model.Name` = 结构体类型名（reflect.Type.Name()）
   - `model.PkgPath` = 结构体包路径（reflect.Type.PkgPath()）
   - `model.PkgKey` = 组合标识：`${PkgPath}/${Name}`

2. **字段处理规则**
   - 每个结构体字段转换为model.Field对象：
     - `field.Name` ← 字段名称
     - `field.Type` ← 字段类型
     - `field.Value` ← 字段值
     - `field.Spec` ← 字段标签的`orm`部分

### 二、类型约束规范

1. **允许的字段类型**
   - 基础类型：整型/浮点/布尔/字符串
   - 复合类型：
     - 结构体（自动递归转换）
     - 时间类型（time.Time）
     - 指针类型（指向上述类型）
     - Slice类型（元素需符合上述类型要求）

2. **特殊类型处理**
   Slice类型 → 元素类型必须为：
   - 基础数值类型
   - 合规结构体
   - time.Time
   - 上述类型的指针形式
   - 所有的描述返回值都是""

### 三、其他处理规则

1. 字段标签包含Key，则该Field为PrimaryField

2. 字段标签view，表示Field支持的View种类可以是origin/detail/lite 其他字段非法

3. Model Interface函数根据要求返回对应的struct值,每个字段要求匹配view定义，如果字段没有定义，则默认为origin，如果ptrValue为true则表示返回struct值的指针

4. Model Copy函数如果传入true，返回一个Model副本，但是Fied的值都初始化成对应类型的初始值，否则返回一个完整的Model各个Field的值保持与源Model一致
