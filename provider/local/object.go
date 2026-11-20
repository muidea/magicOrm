package local

import (
	"fmt"
	"path"
	"reflect"
	"strings"
	"sync"

	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/internal/utils"
	"github.com/muidea/magicOrm/models"
)

type objectFieldSchema struct {
	structIndex  int
	index        int
	name         string
	typePtr      *TypeImpl
	specPtr      *SpecImpl
	isPtr        bool
	isSlice      bool
	enableDetail bool
	enableLite   bool
}

type objectSchema struct {
	name       string
	pkgPath    string
	pkgKey     string
	fields     []objectFieldSchema
	fieldByIdx map[string]int
	primaryIdx int
}

var objectSchemaCache sync.Map // reflect.Type -> *objectSchema

type objectImpl struct {
	objectPtr   bool
	objectValue reflect.Value
	fields      []*field
	schema      *objectSchema
	viewSpec    models.ViewDeclare

	// 临时变量不进行数据序列化传递
	valueValidator models.ValueValidator
}

func (s *objectImpl) GetName() string {
	if s.schema != nil {
		return s.schema.name
	}
	return reflect.Indirect(s.objectValue).Type().Name()
}

func (s *objectImpl) GetShowName() string {
	return s.GetName()
}

func (s *objectImpl) GetPkgPath() string {
	if s.schema != nil {
		return s.schema.pkgPath
	}
	rType := reflect.Indirect(s.objectValue).Type()
	return rType.PkgPath()
}

func (s *objectImpl) GetPkgKey() string {
	if s.schema != nil {
		return s.schema.pkgKey
	}
	return path.Join(s.GetPkgPath(), s.GetName())
}

func (s *objectImpl) GetDescription() string {
	return ""
}

func (s *objectImpl) GetUniqueConstraints() []models.UniqueConstraint {
	declaration, ok := s.Interface(true).(models.UniqueConstraintDeclarer)
	if !ok {
		return nil
	}
	constraints := declaration.GetUniqueConstraints()
	ret := make([]models.UniqueConstraint, len(constraints))
	for idx, constraint := range constraints {
		ret[idx] = models.UniqueConstraint{Name: constraint.Name, Fields: append([]string(nil), constraint.Fields...)}
	}
	return ret
}

func (s *objectImpl) GetFields() (ret models.Fields) {
	ret = make(models.Fields, len(s.fields))
	for i, sf := range s.fields {
		ret[i] = sf
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
	if s.schema != nil && s.schema.primaryIdx >= 0 && s.schema.primaryIdx < len(s.fields) {
		return s.fields[s.schema.primaryIdx]
	}
	for _, sf := range s.fields {
		if models.IsPrimaryField(sf) {
			ret = sf
			return
		}
	}

	return
}

func (s *objectImpl) GetField(name string) (ret models.Field) {
	if s.schema != nil {
		if idx, ok := s.schema.fieldByIdx[name]; ok {
			return s.fields[idx]
		}
		return nil
	}
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

func getOrBuildObjectSchema(entityType reflect.Type) (*objectSchema, *cd.Error) {
	if cached, ok := objectSchemaCache.Load(entityType); ok {
		return cached.(*objectSchema), nil
	}
	typePtr, typeErr := NewType(entityType)
	if typeErr != nil {
		slog.Error("getValueModel NewType failed", "entityType", entityType.String(), "error", typeErr.Error())
		return nil, typeErr
	}
	if typePtr.GetValue() != models.TypeStructValue {
		err := cd.NewError(cd.Unexpected, fmt.Sprintf("illegal type, must be a struct entity, type:%s", entityType.String()))
		slog.Error("getValueModel: not a struct type", "entityType", entityType.String())
		return nil, err
	}

	name := entityType.Name()
	pkgPath := entityType.PkgPath()
	schema := &objectSchema{
		name:       name,
		pkgPath:    pkgPath,
		pkgKey:     path.Join(pkgPath, name),
		fieldByIdx: make(map[string]int),
		primaryIdx: -1,
	}

	fieldNum := entityType.NumField()
	zeroStruct := reflect.New(entityType).Elem()
	verifyFields := make([]*field, 0, fieldNum)
	for idx := range fieldNum {
		fieldInfo := entityType.Field(idx)
		if strings.TrimSpace(fieldInfo.Tag.Get(ormTag)) == "-" {
			continue
		}
		fTypePtr, err := NewType(fieldInfo.Type)
		if err != nil {
			slog.Error("getValueModel getFieldInfo failed", "struct", name, "fieldIndex", idx, "error", err.Error())
			return nil, err
		}
		fSpecPtr, err := NewSpec(fieldInfo.Tag)
		if err != nil {
			slog.Error("getValueModel getFieldInfo failed", "struct", name, "fieldIndex", idx, "error", err.Error())
			return nil, err
		}
		fName := fieldInfo.Name
		if fSpecPtr.GetFieldName() != "" {
			fName = fSpecPtr.GetFieldName()
		}
		fieldSchema := objectFieldSchema{
			structIndex:  idx,
			index:        idx,
			name:         fName,
			typePtr:      fTypePtr,
			specPtr:      fSpecPtr,
			isPtr:        fTypePtr.IsPtrType(),
			isSlice:      fTypePtr.GetValue() == models.TypeSliceValue,
			enableDetail: fSpecPtr.EnableView(models.DetailView),
			enableLite:   fSpecPtr.EnableView(models.LiteView),
		}
		if fSpecPtr.IsPrimaryKey() {
			if schema.primaryIdx >= 0 {
				err := cd.NewError(cd.Unexpected, fmt.Sprintf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldInfo.Name, name))
				slog.Error("getValueModel: duplicate primary key", "struct", name, "field", fieldInfo.Name)
				return nil, err
			}
			schema.primaryIdx = len(schema.fields)
		}
		schema.fieldByIdx[fName] = len(schema.fields)
		schema.fields = append(schema.fields, fieldSchema)
		verifyFields = append(verifyFields, &field{
			index:    idx,
			name:     fName,
			typePtr:  fTypePtr,
			specPtr:  fSpecPtr,
			valuePtr: NewValue(zeroStruct.Field(idx)),
		})
	}
	if len(schema.fields) == 0 {
		err := cd.NewError(cd.Unexpected, fmt.Sprintf("no define orm field, struct name:%s", name))
		slog.Error("getValueModel: no orm fields", "struct", name)
		return nil, err
	}
	if schema.primaryIdx < 0 {
		err := cd.NewError(cd.Unexpected, fmt.Sprintf("no define primary key field, struct name:%s", name))
		slog.Error("getValueModel: no primary key", "struct", name)
		return nil, err
	}
	verifyImpl := &objectImpl{objectValue: zeroStruct, fields: verifyFields, schema: schema}
	if err := models.VerifyModel(verifyImpl); err != nil {
		slog.Error("getValueModel VerifyModel failed", "struct", name, "error", err.Error())
		return nil, err
	}
	actual, _ := objectSchemaCache.LoadOrStore(entityType, schema)
	return actual.(*objectSchema), nil
}

func getValueModel(entityValue reflect.Value, viewSpec models.ViewDeclare) (ret *objectImpl, err *cd.Error) {
	if !entityValue.IsValid() || (entityValue.Kind() == reflect.Ptr && entityValue.IsNil()) {
		return nil, cd.NewError(cd.IllegalParam, "model value is nil or invalid")
	}
	isPtr := entityValue.Kind() == reflect.Ptr
	entityValue = reflect.Indirect(entityValue)
	entityType := entityValue.Type()
	schema, schemaErr := getOrBuildObjectSchema(entityType)
	if schemaErr != nil {
		return nil, schemaErr
	}

	fields := make([]*field, len(schema.fields))
	for i := range schema.fields {
		meta := &schema.fields[i]
		fieldVal := entityValue.Field(meta.structIndex)
		var valuePtr *ValueImpl
		switch viewSpec {
		case models.MetaView:
			valuePtr = &ValueImpl{value: fieldVal}
			if !meta.isPtr {
				if meta.isSlice {
					valuePtr.reset(false)
				} else {
					valuePtr.reset(true)
				}
			} else {
				valuePtr.reset(false)
			}
		case models.DetailView:
			valuePtr = NewValue(fieldVal)
			if !meta.enableDetail {
				valuePtr.reset(false)
			} else if !valuePtr.IsValid() {
				valuePtr.reset(true)
			}
		case models.LiteView:
			valuePtr = NewValue(fieldVal)
			if !meta.enableLite {
				valuePtr.reset(false)
			} else if !valuePtr.IsValid() {
				valuePtr.reset(true)
			}
		case models.OriginView:
			valuePtr = NewValue(fieldVal)
		default:
			valuePtr = NewValue(fieldVal)
			slog.Error("getFieldInfo: unsupported viewSpec", "viewSpec", viewSpec)
		}
		fields[i] = &field{
			index:    meta.index,
			name:     meta.name,
			typePtr:  meta.typePtr,
			specPtr:  meta.specPtr,
			valuePtr: valuePtr,
		}
	}

	ret = &objectImpl{
		objectValue: entityValue,
		objectPtr:   isPtr,
		fields:      fields,
		schema:      schema,
		viewSpec:    viewSpec,
	}
	return
}
