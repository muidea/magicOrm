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
	viewSpec    models.ViewDeclare

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
		slog.Error("objectImpl.SetFieldValue failed", "field", name, "error", err.Error())
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

	slog.Warn("objectImpl.innerSetFieldValue: field not found", "field", name)
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

func (s *objectImpl) ResponseIncludesField(name string) bool {
	field := s.GetField(name)
	if field == nil {
		return false
	}

	switch s.viewSpec {
	case models.DetailView, models.LiteView:
		return field.GetSpec().EnableView(s.viewSpec)
	default:
		return true
	}
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
		slog.Error("getValueModel NewType failed", "entityType", entityType.String(), "error", err.Error())
		return
	}
	if typePtr.GetValue() != models.TypeStructValue {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal type, must be a struct entity, type:%s", entityType.String()))
		slog.Error("getValueModel: not a struct type", "entityType", entityType.String())
		return
	}

	hasPrimaryKey := false
	impl := &objectImpl{objectValue: entityValue, objectPtr: isPtr, fields: []*field{}, viewSpec: viewSpec}
	fieldNum := entityType.NumField()
	for idx := range fieldNum {
		fieldVal := entityValue.Field(idx)
		fieldInfo := entityType.Field(idx)
		tField, tErr := getFieldInfo(idx, fieldInfo, fieldVal, viewSpec)
		if tErr != nil {
			err = tErr
			slog.Error("getValueModel getFieldInfo failed", "struct", impl.GetName(), "fieldIndex", idx, "error", err.Error())
			return
		}

		if models.IsPrimaryField(tField) {
			if hasPrimaryKey {
				err = cd.NewError(cd.Unexpected, fmt.Sprintf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldInfo.Name, impl.GetName()))
				slog.Error("getValueModel: duplicate primary key", "struct", impl.GetName(), "field", fieldInfo.Name)
				return
			}

			hasPrimaryKey = true
		}

		impl.fields = append(impl.fields, tField)
	}

	if len(impl.fields) == 0 {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("no define orm field, struct name:%s", impl.GetName()))
		slog.Error("getValueModel: no orm fields", "struct", impl.GetName())
		return
	}
	if !hasPrimaryKey {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("no define primary key field, struct name:%s", impl.GetName()))
		slog.Error("getValueModel: no primary key", "struct", impl.GetName())
		return
	}

	err = models.VerifyModel(impl)
	if err != nil {
		slog.Error("getValueModel VerifyModel failed", "struct", impl.GetName(), "error", err.Error())
		return
	}
	ret = impl
	return
}
