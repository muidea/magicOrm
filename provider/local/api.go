package local

import (
	"fmt"
	"log/slog"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/utils"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation"
)

func EncodeValue(vVal any, vType models.Type) (ret any, err *cd.Error) {
	if !models.IsBasic(vType) {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("EncodeValue failed, illegal type, type pkgKey:%s", vType.GetPkgKey()))
		return
	}

	rVal := reflect.ValueOf(vVal)
	ret, err = encodeValue(rVal, vType)
	return
}

func DecodeValue(vVal any, vType models.Type) (ret any, err *cd.Error) {
	if !models.IsBasic(vType) {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal type, type pkgKey:%s", vType.GetPkgKey()))
		return
	}

	rVal := reflect.ValueOf(vVal)
	ret, err = decodeValue(rVal, vType)
	return
}

func GetEntityType(entity any) (ret models.Type, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is invalid")
		return
	}

	typeImplPtr, typeImplErr := checkEntityType(entity)
	if typeImplErr != nil {
		err = typeImplErr
		slog.Error("GetEntityType checkEntityType failed", "error", err.Error())
		return
	}

	ret = typeImplPtr
	return
}

func GetEntityValue(entity any) (ret models.Value, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is invalid")
		return
	}

	_, err = checkEntityType(entity)
	if err != nil {
		slog.Error("GetEntityValue checkEntityType failed", "error", err.Error())
		return
	}

	ret = NewValue(reflect.ValueOf(entity))
	return
}

func GetEntityModel(entity any, valueValidator models.ValueValidator) (ret models.Model, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is invalid")
		return
	}

	reallyVal := reflect.Indirect(reflect.ValueOf(entity))
	if !reallyVal.IsValid() {
		return nil, cd.NewError(cd.IllegalParam, "model value is nil or invalid")
	}
	if !reallyVal.CanSet() {
		newVal := reflect.New(reallyVal.Type()).Elem()
		newVal.Set(reallyVal)
		reallyVal = newVal
	}

	implPtr, implErr := getValueModel(reallyVal, models.OriginView)
	if implErr != nil {
		err = implErr
		slog.Error("GetEntityModel getValueModel failed", "entityType", reallyVal.Type().String(), "error", err.Error())
		return
	}

	implPtr.valueValidator = valueValidator
	ret = implPtr
	return
}

func GetValueModel(vVal models.Value) (ret models.Model, err *cd.Error) {
	if vVal == nil {
		err = cd.NewError(cd.IllegalParam, "value is invalid")
		return
	}
	valueImplPtr, valueImplOK := vVal.(*ValueImpl)
	if !valueImplOK {
		err = cd.NewError(cd.IllegalParam, "value is invalid")
		return
	}

	implPtr, implErr := getValueModel(valueImplPtr.value, models.MetaView)
	if implErr != nil {
		err = implErr
		slog.Error("GetValueModel getValueModel failed", "error", err.Error())
		return
	}

	ret = implPtr
	return
}

func GetModelFilter(vModel models.Model) (ret models.Filter, err *cd.Error) {
	valuePtr := NewValue(reflect.ValueOf(vModel.Interface(true)))
	ret = newFilter(valuePtr, vModel)
	return
}

func SetModelValue(vModel models.Model, vVal models.Value, disableValidator bool) (ret models.Model, err *cd.Error) {
	valImplPtr, valImplOK := vVal.(*ValueImpl)
	if !valImplOK {
		err = cd.NewError(cd.IllegalParam, "value is invalid")
		slog.Error("SetModelValue: value is not *ValueImpl")
		return
	}
	valueModel, valueModelErr := getValueModel(valImplPtr.value, models.OriginView)
	if valueModelErr != nil {
		err = valueModelErr
		slog.Error("SetModelValue getValueModel failed", "error", err.Error())
		return
	}

	vModelImplPtr := vModel.(*objectImpl)
	fields := valueModel.GetFields()
	for _, field := range fields {
		if !models.IsValidField(field) && !models.IsAssignedField(field) {
			continue
		}

		err = vModelImplPtr.innerSetFieldValue(field.GetName(), field.GetValue().Get(), disableValidator)
		if err != nil {
			slog.Error("SetModelValue innerSetFieldValue failed", "model", vModel.GetPkgKey(), "field", field.GetName(), "error", err.Error())
			return
		}
	}

	ret = vModel
	return
}

// NewSpec name[key][auto]
func NewSpec(tag reflect.StructTag) (ret *SpecImpl, err *cd.Error) {
	ormSpec := tag.Get(ormTag)
	ret, err = getOrmSpec(ormSpec)
	if err != nil {
		return
	}
	constraints := tag.Get(constraintsTag)
	if constraints != "" {
		ret.constraints = utils.ParseConstraints(constraints)
	}

	viewSpec := tag.Get(viewTag)
	ret.viewDeclare = getViewItems(viewSpec)
	return
}

func NewType(val reflect.Type) (ret *TypeImpl, err *cd.Error) {
	rType := val
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}
	_, tErr := utils.GetTypeEnum(rType)
	if tErr != nil {
		err = tErr
		return
	}

	ret = &TypeImpl{typeVal: val}
	return
}

// NewValidationExtension creates a new validation extension
func NewValidationExtension(valueValidator models.ValueValidator) ValidationExtension {
	// Create default validation configuration
	config := validation.DefaultConfig()
	config.EnableCaching = false

	// Create validation factory
	factory := validation.NewValidationFactory()

	// Create validation manager
	validationManager := factory.CreateValidationManager(config)

	return &validationExtensionImpl{
		validationManager: validationManager,
		valueValidator:    valueValidator,
	}
}

func NewValue(val reflect.Value) (ret *ValueImpl) {
	ret = &ValueImpl{value: val}
	ret.assigned = !ret.IsZero()
	return
}
