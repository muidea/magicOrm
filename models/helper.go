package models

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
)

// IsBasic 判断是否是基本数值类型
// 基本数值类型见const.go的定义
func IsBasic(tType Type) bool {
	return tType.Elem().GetValue() < TypeStructValue
}

// IsStruct 判断是否是struct
// 当前类型是struct,或者slice的elem是struct
// 根据Type.Elem()的定义，非slice的elem是type本身，所以这里直接判断elem即可
func IsStruct(tType Type) bool {
	return tType.GetValue() == TypeStructValue
}

// IsSlice 判断是否是slice
// 不管elem是普通类型还是struct类型
func IsSlice(tType Type) bool {
	return tType.GetValue() == TypeSliceValue
}

func IsBasicType(typeValue TypeDeclare) bool {
	return typeValue < TypeStructValue
}

func IsDateTimeType(typeValue TypeDeclare) bool {
	return typeValue == TypeDateTimeValue
}

func IsStructType(typeValue TypeDeclare) bool {
	return typeValue == TypeStructValue
}

func IsSliceType(typeValue TypeDeclare) bool {
	return typeValue == TypeSliceValue
}

func GetTypeValue(typeName string) (ret TypeDeclare) {
	tVal, tOK := typeName2ValueMap[typeName]
	if tOK {
		ret = tVal
		return
	}

	ret = TypeStructValue
	return
}

func IsCustomerDeclare(val ValueDeclare) bool {
	return val == Customer
}

func IsAutoIncrementDeclare(val ValueDeclare) bool {
	return val == AutoIncrement
}

func IsUUIDDeclare(val ValueDeclare) bool {
	return val == UUID
}

func IsSnowFlakeDeclare(val ValueDeclare) bool {
	return val == SnowFlake
}

func IsDateTimeDeclare(val ValueDeclare) bool {
	return val == DateTime
}

// IsBasicField 判断Field对应的Type是否是基本类型
func IsBasicField(field Field) bool {
	return IsBasic(field.GetType())
}

// IsStructField 判断Field对应的Type是否是struct类型
func IsStructField(field Field) bool {
	return IsStruct(field.GetType())
}

// IsSliceField 判断Field对应的Type是否是slice类型
func IsSliceField(field Field) bool {
	return IsSlice(field.GetType())
}

// IsValidField 判断Field的值是否有效
// 1. Field的值必须有效 或者Field有默认值
func IsValidField(field Field) bool {
	return field.GetValue().IsValid()
}

// IsPtrField 判断Field对应的Type是否是指针类型
func IsPtrField(field Field) bool {
	return field.GetType().IsPtrType()
}

// IsAssignedField 判断Field是否已经被赋值
// 如果Field的值非初始值，认为是已经被赋值
func IsAssignedField(field Field) bool {
	return !field.GetValue().IsZero()
}

// IsPrimaryField 判断Field是否是主键
// Field的Spec申明是主键，认为是主键
func IsPrimaryField(field Field) bool {
	return field.GetSpec().IsPrimaryKey()
}

// VerifyModel 验证Model
// 1. Name和PkgPath不能为""
// 2. Fields 不能存在重名的Field
// 3. 至少有一个PrimaryField
// 4. 如果校验失败，则返回失败信息
// 5. 校验通过返回nil
func VerifyModel(vModel Model) (err *cd.Error) {
	if vModel.GetName() == "" {
		err = cd.NewError(cd.Unexpected, "model name is empty")
		return
	}

	if vModel.GetPkgPath() == "" {
		err = cd.NewError(cd.Unexpected, "model pkgPath is empty")
		return
	}

	fieldNum := len(vModel.GetFields())
	if fieldNum == 0 {
		err = cd.NewError(cd.Unexpected, "model fields is empty")
		return
	}

	primaryFieldNum := 0
	fieldNameMap := make(map[string]bool)
	for _, vField := range vModel.GetFields() {
		if vField.GetSpec().IsPrimaryKey() {
			primaryFieldNum++
		}

		if _, exists := fieldNameMap[vField.GetName()]; exists {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("model field name is duplicate, field name:%s", vField.GetName()))
			return
		}

		fieldNameMap[vField.GetName()] = true
	}

	if primaryFieldNum == 0 {
		err = cd.NewError(cd.Unexpected, "model no primary field")
		return
	}

	return
}
