package local

import (
	"fmt"
	"path"
	"reflect"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
	"log/slog"
)

type objectImpl struct {
	objectPtr   bool
	objectValue reflect.Value
	fields      []*field

	// 临时变量不进行数据序列化传递
	valueValidator models.ValueValidator
}

func (s *objectImpl) GetName() string {
	return reflect.Indirect(s.objectValue).Type().Name()
}

func (s *objectImpl) GetShowName() string {
	return s.GetName()
}

func (s *objectImpl) GetPkgPath() string {
	rType := reflect.Indirect(s.objectValue).Type()
	return rType.PkgPath()
}

func (s *objectImpl) GetPkgKey() string {
	return path.Join(s.GetPkgPath(), s.GetName())
}

func (s *objectImpl) GetDescription() string {
	return ""
}

func (s *objectImpl) GetFields() (ret models.Fields) {
	for _, sf := range s.fields {
		ret = append(ret, sf)
	}

	return
}

func (s *objectImpl) SetFieldValue(name string, val any) (err *cd.Error) {
	err = s.innerSetFieldValue(name, val, false)
	if err != nil {
		slog.Error("error occurred", "error", "operation failed")
		return
	}

	//slog.Warn("warning", "message", "warning")
	return
}

func (s *objectImpl) innerSetFieldValue(name string, val any, disableValidator bool) (err *cd.Error) {
	for _, sf := range s.fields {
		sf.valueValidator = s.valueValidator
		if sf.GetName() == name {
			err = sf.innerSetValue(val, disableValidator)
			return
		}
	}

	slog.Warn("warning", "message", "warning")
	return
}

func (s *objectImpl) SetPrimaryFieldValue(val any) (err *cd.Error) {
	for _, sf := range s.fields {
		if models.IsPrimaryField(sf) {
			err = sf.SetValue(val)
			return
		}
	}

	return
}

func (s *objectImpl) GetPrimaryField() (ret models.Field) {
	for _, sf := range s.fields {
		if models.IsPrimaryField(sf) {
			ret = sf
			return
		}
	}

	return
}

func (s *objectImpl) GetField(name string) (ret models.Field) {
	for _, sf := range s.fields {
		if sf.GetName() == name {
			ret = sf
			return
		}
	}

	return
}

func (s *objectImpl) Interface(ptrValue bool) (ret any) {
	if ptrValue {
		ret = s.objectValue.Addr().Interface()
		return
	}

	ret = s.objectValue.Interface()
	return
}

func (s *objectImpl) Copy(viewSpec models.ViewDeclare) models.Model {
	if !s.objectValue.IsValid() {
		return &objectImpl{}
	}

	modelImplPtr, _ := getValueModel(utils.DeepCopyForReflect(s.objectValue), viewSpec)

	modelImplPtr.valueValidator = s.valueValidator
	return modelImplPtr
}

func (s *objectImpl) Reset() {
	for _, sf := range s.fields {
		sf.Reset()
	}
}

func getValueModel(entityValue reflect.Value, viewSpec models.ViewDeclare) (ret *objectImpl, err *cd.Error) {
	isPtr := entityValue.Kind() == reflect.Ptr
	entityValue = reflect.Indirect(entityValue)
	entityType := entityValue.Type()
	typePtr, typeErr := NewType(entityType)
	if typeErr != nil {
		err = typeErr
		slog.Error("error occurred", "error", "operation failed")
		return
	}
	if typePtr.GetValue() != models.TypeStructValue {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal type, must be a struct entity, type:%s", entityType.String()))
		slog.Error("error occurred", "error", "operation failed")
		return
	}

	hasPrimaryKey := false
	impl := &objectImpl{objectValue: entityValue, objectPtr: isPtr, fields: []*field{}}
	fieldNum := entityType.NumField()
	for idx := range fieldNum {
		fieldVal := entityValue.Field(idx)
		fieldInfo := entityType.Field(idx)
		tField, tErr := getFieldInfo(idx, fieldInfo, fieldVal, viewSpec)
		if tErr != nil {
			err = tErr
			slog.Error("error occurred", "error", err.Error())
			return
		}

		if models.IsPrimaryField(tField) {
			if hasPrimaryKey {
				err = cd.NewError(cd.Unexpected, fmt.Sprintf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldInfo.Name, impl.GetName()))
				slog.Error("message")
				return
			}

			hasPrimaryKey = true
		}

		impl.fields = append(impl.fields, tField)
	}

	if len(impl.fields) == 0 {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("no define orm field, struct name:%s", impl.GetName()))
		slog.Error("error occurred", "error", "operation failed")
		return
	}
	if !hasPrimaryKey {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("no define primary key field, struct name:%s", impl.GetName()))
		slog.Error("error occurred", "error", "operation failed")
		return
	}

	err = models.VerifyModel(impl)
	if err != nil {
		slog.Error("error occurred", "error", "operation failed")
		return
	}
	ret = impl
	return
}
