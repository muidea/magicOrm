package helper

import (
	"reflect"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider/remote"
)

type maskNode struct {
	children map[string]*maskNode
}

type viewMaskOptions struct {
	IncludeUntaggedTopLevel bool
}

func (s *maskNode) child(name string) *maskNode {
	if s.children == nil {
		s.children = map[string]*maskNode{}
	}
	if s.children[name] == nil {
		s.children[name] = &maskNode{}
	}
	return s.children[name]
}

func buildFieldMask(entity any, fields ...string) (ret *remote.ObjectValue, err *cd.Error) {
	ret, err = GetObjectValue(entity)
	if err != nil || len(fields) == 0 {
		return
	}

	tree := &maskNode{}
	for _, fieldPath := range fields {
		fieldPath = strings.TrimSpace(fieldPath)
		if fieldPath == "" {
			continue
		}

		current := tree
		for _, item := range strings.Split(fieldPath, ".") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			current = current.child(item)
		}
	}

	ret = trimObjectValue(ret, tree)
	return
}

func BuildViewMask(entity any, view models.ViewDeclare) (ret *remote.ObjectValue, err *cd.Error) {
	return buildViewMaskWithOptions(entity, view, &viewMaskOptions{})
}

func buildViewMaskWithOptions(entity any, view models.ViewDeclare, options *viewMaskOptions) (ret *remote.ObjectValue, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is nil")
		return
	}

	entityType := reflect.TypeOf(entity)
	ret, err = buildObjectMaskByView(entityType, "", view, options, map[string]struct{}{})
	return
}

func trimObjectValue(value *remote.ObjectValue, node *maskNode) *remote.ObjectValue {
	if value == nil || node == nil || len(node.children) == 0 {
		return value
	}

	ret := &remote.ObjectValue{
		ID:      value.ID,
		Name:    value.Name,
		PkgPath: value.PkgPath,
		Fields:  []*remote.FieldValue{},
	}

	for _, field := range value.Fields {
		childNode, ok := node.children[field.Name]
		if !ok {
			continue
		}

		newField := &remote.FieldValue{Name: field.Name, Assigned: field.Assigned, Value: field.Value}
		if len(childNode.children) > 0 {
			switch raw := field.Value.(type) {
			case *remote.ObjectValue:
				newField.Value = trimObjectValue(raw, childNode)
			case remote.ObjectValue:
				newField.Value = trimObjectValue(&raw, childNode)
			case *remote.SliceObjectValue:
				newField.Value = trimSliceObjectValue(raw, childNode)
			case remote.SliceObjectValue:
				newField.Value = trimSliceObjectValue(&raw, childNode)
			}
		}

		ret.Fields = append(ret.Fields, newField)
	}

	return ret
}

func trimSliceObjectValue(value *remote.SliceObjectValue, node *maskNode) *remote.SliceObjectValue {
	if value == nil || node == nil || len(node.children) == 0 {
		return value
	}

	ret := &remote.SliceObjectValue{
		Name:    value.Name,
		PkgPath: value.PkgPath,
		Values:  make([]*remote.ObjectValue, 0, len(value.Values)),
	}
	for _, item := range value.Values {
		ret.Values = append(ret.Values, trimObjectValue(item, node))
	}
	return ret
}

func buildObjectMaskByView(entityType reflect.Type, pathPrefix string, view models.ViewDeclare, options *viewMaskOptions, visiting map[string]struct{}) (ret *remote.ObjectValue, err *cd.Error) {
	entityType = indirectMaskType(entityType)
	if entityType.Kind() != reflect.Struct {
		ret = nil
		return
	}

	ret = &remote.ObjectValue{
		Name:    entityType.Name(),
		PkgPath: entityType.PkgPath(),
		Fields:  []*remote.FieldValue{},
	}
	typeKey := entityType.PkgPath() + "." + entityType.Name()
	if _, ok := visiting[typeKey]; ok {
		return
	}
	visiting[typeKey] = struct{}{}
	defer delete(visiting, typeKey)

	for idx := 0; idx < entityType.NumField(); idx++ {
		fieldType := entityType.Field(idx)
		if isIgnoredField(fieldType) {
			continue
		}
		fieldName, fieldErr := getFieldName(fieldType)
		if fieldErr != nil {
			err = fieldErr
			return
		}

		specPtr, specErr := newSpec(fieldType.Tag)
		if specErr != nil {
			err = specErr
			return
		}
		includeUntagged := options != nil && options.IncludeUntaggedTopLevel && pathPrefix == "" && len(specPtr.ViewDeclare) == 0
		if view != models.OriginView && !specPtr.EnableView(view) && !includeUntagged {
			continue
		}

		childPath := fieldName
		if pathPrefix != "" {
			childPath = pathPrefix + "." + fieldName
		}

		fieldValue, fieldValueErr := buildFieldMaskValue(fieldName, fieldType.Type, childPath, childViewOf(view), options, visiting)
		if fieldValueErr != nil {
			err = fieldValueErr
			return
		}
		ret.Fields = append(ret.Fields, fieldValue)
	}

	return
}

func buildFieldMaskValue(fieldName string, fieldType reflect.Type, pathPrefix string, view models.ViewDeclare, options *viewMaskOptions, visiting map[string]struct{}) (ret *remote.FieldValue, err *cd.Error) {
	typePtr, typeErr := newType(fieldType)
	if typeErr != nil {
		err = typeErr
		return
	}

	ret = &remote.FieldValue{Name: fieldName}
	switch {
	case models.IsSlice(typePtr) && models.IsStruct(typePtr.Elem()):
		ret.Value, err = buildSliceObjectMaskByView(fieldType, pathPrefix, view, options, visiting)
	case models.IsStruct(typePtr):
		ret.Value, err = buildObjectMaskByView(fieldType, pathPrefix, view, options, visiting)
	default:
		ret.Value, err = zeroMaskValue(typePtr)
	}
	return
}

func buildSliceObjectMaskByView(fieldType reflect.Type, pathPrefix string, view models.ViewDeclare, options *viewMaskOptions, visiting map[string]struct{}) (ret *remote.SliceObjectValue, err *cd.Error) {
	elemType := indirectMaskType(fieldType)
	if elemType.Kind() == reflect.Slice {
		elemType = indirectMaskType(elemType.Elem())
	}
	if elemType.Kind() != reflect.Struct {
		ret = nil
		return
	}

	ret = &remote.SliceObjectValue{
		Name:    elemType.Name(),
		PkgPath: elemType.PkgPath(),
		Values:  []*remote.ObjectValue{},
	}
	childValue, childErr := buildObjectMaskByView(elemType, pathPrefix, view, options, visiting)
	if childErr != nil {
		err = childErr
		return
	}
	if childValue != nil {
		ret.Values = append(ret.Values, childValue)
	}
	return
}

func indirectMaskType(entityType reflect.Type) reflect.Type {
	for entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	return entityType
}

func zeroMaskValue(fieldType models.Type) (ret any, err *cd.Error) {
	if fieldType == nil {
		return nil, cd.NewError(cd.IllegalParam, "mask field type is nil")
	}
	if fieldType.GetValue() == models.TypeSliceValue {
		return zeroBasicSliceMaskValue(fieldType.Elem())
	}

	// A view mask is a remote persistence contract, not a public JSON DTO.
	// Named primitive types may implement json.Marshaler (for example a
	// namespace-qualified identity that exposes an object over HTTP). Keeping
	// the named zero value here would invoke that marshaler and can turn a
	// selected scalar field into JSON null. Use canonical wire primitives so
	// both named values and pointer-to-named values remain selected.
	switch fieldType.GetValue() {
	case models.TypeBooleanValue:
		ret = false
	case models.TypeStringValue, models.TypeDateTimeValue:
		ret = ""
	case models.TypeByteValue:
		ret = int8(0)
	case models.TypeSmallIntegerValue:
		ret = int16(0)
	case models.TypeInteger32Value:
		ret = int32(0)
	case models.TypeIntegerValue:
		ret = int(0)
	case models.TypeBigIntegerValue:
		ret = int64(0)
	case models.TypePositiveByteValue:
		ret = uint8(0)
	case models.TypePositiveSmallIntegerValue:
		ret = uint16(0)
	case models.TypePositiveInteger32Value:
		ret = uint32(0)
	case models.TypePositiveIntegerValue:
		ret = uint(0)
	case models.TypePositiveBigIntegerValue:
		ret = uint64(0)
	case models.TypeFloatValue:
		ret = float32(0)
	case models.TypeDoubleValue:
		ret = float64(0)
	default:
		err = cd.NewError(cd.Unexpected, "mask field is not a basic type")
	}
	return
}

func zeroBasicSliceMaskValue(elemType models.Type) (ret any, err *cd.Error) {
	if elemType == nil {
		return nil, cd.NewError(cd.IllegalParam, "mask slice element type is nil")
	}

	switch elemType.GetValue() {
	case models.TypeBooleanValue:
		ret = []bool{}
	case models.TypeStringValue, models.TypeDateTimeValue:
		ret = []string{}
	case models.TypeByteValue:
		ret = []int8{}
	case models.TypeSmallIntegerValue:
		ret = []int16{}
	case models.TypeInteger32Value:
		ret = []int32{}
	case models.TypeIntegerValue:
		ret = []int{}
	case models.TypeBigIntegerValue:
		ret = []int64{}
	case models.TypePositiveByteValue:
		ret = []uint8{}
	case models.TypePositiveSmallIntegerValue:
		ret = []uint16{}
	case models.TypePositiveInteger32Value:
		ret = []uint32{}
	case models.TypePositiveIntegerValue:
		ret = []uint{}
	case models.TypePositiveBigIntegerValue:
		ret = []uint64{}
	case models.TypeFloatValue:
		ret = []float32{}
	case models.TypeDoubleValue:
		ret = []float64{}
	default:
		err = cd.NewError(cd.Unexpected, "mask slice element is not a basic type")
	}
	return
}

func childViewOf(view models.ViewDeclare) models.ViewDeclare {
	return defaultRelationView(view)
}

func defaultRelationView(view models.ViewDeclare) models.ViewDeclare {
	switch view {
	case models.MetaView:
		return models.MetaView
	case models.OriginView, models.DetailView, models.LiteView:
		return models.LiteView
	default:
		return view
	}
}
