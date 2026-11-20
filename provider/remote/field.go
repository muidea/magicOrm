package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"log/slog"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/utils"
	"github.com/muidea/magicOrm/models"
)

// JSONValue returns the public model-value projection without changing Value.
// A uint8 model collection is a numeric array, including when nested in a DTO.
func (s FieldValue) JSONValue() any {
	return utils.BasicValueForJSON(s.Value)
}

func isSimpleJSONASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c >= 0x7f || c == '"' || c == '\\' || c == '<' || c == '>' || c == '&' {
			return false
		}
	}
	return true
}

func appendJSONString(dst []byte, s string) ([]byte, error) {
	if isSimpleJSONASCII(s) {
		dst = append(dst, '"')
		dst = append(dst, s...)
		return append(dst, '"'), nil
	}
	raw, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return append(dst, raw...), nil
}

func appendFieldValueJSON(dst []byte, val any) ([]byte, error) {
	switch v := val.(type) {
	case nil:
		return append(dst, "null"...), nil
	case string:
		return appendJSONString(dst, v)
	case bool:
		return strconv.AppendBool(dst, v), nil
	case int:
		return strconv.AppendInt(dst, int64(v), 10), nil
	case int64:
		return strconv.AppendInt(dst, v, 10), nil
	case int32:
		return strconv.AppendInt(dst, int64(v), 10), nil
	case int16:
		return strconv.AppendInt(dst, int64(v), 10), nil
	case int8:
		return strconv.AppendInt(dst, int64(v), 10), nil
	case uint:
		return strconv.AppendUint(dst, uint64(v), 10), nil
	case uint64:
		return strconv.AppendUint(dst, v, 10), nil
	case uint32:
		return strconv.AppendUint(dst, uint64(v), 10), nil
	case uint16:
		return strconv.AppendUint(dst, uint64(v), 10), nil
	case uint8:
		return strconv.AppendUint(dst, uint64(v), 10), nil
	case *string:
		if v == nil {
			return append(dst, "null"...), nil
		}
		return appendJSONString(dst, *v)
	case *bool:
		if v == nil {
			return append(dst, "null"...), nil
		}
		return strconv.AppendBool(dst, *v), nil
	case *int:
		if v == nil {
			return append(dst, "null"...), nil
		}
		return strconv.AppendInt(dst, int64(*v), 10), nil
	case *int64:
		if v == nil {
			return append(dst, "null"...), nil
		}
		return strconv.AppendInt(dst, *v, 10), nil
	default:
		raw, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		return append(dst, raw...), nil
	}
}

func (s FieldValue) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, 64+len(s.Name))
	buf = append(buf, `{"name":`...)
	var err error
	buf, err = appendJSONString(buf, s.Name)
	if err != nil {
		return nil, err
	}
	buf = append(buf, `,"value":`...)
	buf, err = appendFieldValueJSON(buf, s.JSONValue())
	if err != nil {
		return nil, err
	}
	if s.Assigned {
		buf = append(buf, `,"assigned":true}`...)
	} else {
		buf = append(buf, '}')
	}
	return buf, nil
}

// UnmarshalJSON keeps numeric tokens exact until the declared field type is
// known. This also applies when FieldValue is nested in an HTTP request DTO.
func (s *FieldValue) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name     string          `json:"name"`
		Value    json.RawMessage `json:"value"`
		Assigned bool            `json:"assigned,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	val, err := decodeFieldRawValue(raw.Value)
	if err != nil {
		return err
	}
	*s = FieldValue{
		Name:     raw.Name,
		Value:    val,
		Assigned: raw.Assigned,
	}
	return nil
}

func decodeFieldRawValue(raw []byte) (any, error) {
	v := bytes.TrimSpace(raw)
	if len(v) == 0 || bytes.Equal(v, []byte("null")) {
		return nil, nil
	}
	if bytes.Equal(v, []byte("true")) {
		return true, nil
	}
	if bytes.Equal(v, []byte("false")) {
		return false, nil
	}
	if v[0] == '"' {
		var str string
		if err := json.Unmarshal(v, &str); err != nil {
			return nil, err
		}
		return str, nil
	}
	if v[0] == '-' || (v[0] >= '0' && v[0] <= '9') {
		return json.Number(string(v)), nil
	}
	var val any
	decoder := json.NewDecoder(bytes.NewReader(v))
	decoder.UseNumber()
	if err := decoder.Decode(&val); err != nil {
		return nil, err
	}
	return val, nil
}

func (s *FieldValue) String() string {
	return fmt.Sprintf("name:%s,value:%+v", s.Name, s.Value)
}

func (s *Field) GetName() string {
	return s.Name
}

func (s *Field) GetShowName() string {
	return s.ShowName
}

func (s *Field) GetDescription() string {
	return s.Description
}

func (s *Field) GetType() (ret models.Type) {
	ret = s.Type
	return
}

func (s *Field) GetSpec() (ret models.Spec) {
	if s.Spec != nil {
		ret = s.Spec
		return
	}

	ret = &emptySpec
	return
}

func (s *Field) GetValue() (ret models.Value) {
	if s.value != nil {
		ret = s.value
		return
	}

	ret = &ValueImpl{}
	return
}

func (s *Field) SetValue(val any) *cd.Error {
	return s.innerSetValue(val, false)
}

func (s *Field) innerSetValue(val any, disableValidator bool) *cd.Error {
	if s.value == nil {
		s.value = &ValueImpl{}
	}

	if s.valueValidator != nil && !disableValidator {
		constraintVal := s.GetSpec().GetConstraints()
		if constraintVal != nil {
			err := s.valueValidator.ValidateValue(val, constraintVal.Directives())
			if err != nil {
				return utils.FieldValidationError(s.GetName(), err)
			}
		}
	}

	return s.value.Set(val)
}

func (s *Field) GetSliceValue() (ret []models.Value) {
	if !models.IsSlice(s.Type) || s.value == nil || !s.value.IsValid() {
		return
	}

	ret = s.value.UnpackValue()
	return
}

func (s *Field) AppendSliceValue(val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.Unexpected, "field append slice value is nil")
		return
	}
	if !models.IsSlice(s.Type) {
		err = cd.NewError(cd.Unexpected, "field is not slice")
		return
	}

	err = s.value.Append(val)
	return
}

func (s *Field) Reset() {
	if s.value != nil && s.value.IsValid() {
		s.value = NewValue(getInitializeValue(s.Type))
		return
	}

	s.value = &ValueImpl{}
}

func (s *Field) copy(viewSpec models.ViewDeclare) (ret *Field, err error) {
	ret = &Field{
		Name:        s.Name,
		ShowName:    s.ShowName,
		Description: s.Description,
		Type:        s.Type.Copy(),
		//Spec:        s.Spec.Copy(),
	}
	if s.Spec == nil {
		ret.Spec = &emptySpec
	} else {
		ret.Spec = s.Spec.Copy()
	}

	switch viewSpec {
	case models.MetaView:
		if !s.Type.IsPtrType() {
			ret.value = NewValue(getInitializeValue(s.Type))
		} else {
			ret.value = &ValueImpl{}
		}
	case models.DetailView, models.LiteView:
		if !ret.Spec.EnableView(viewSpec) {
			ret.value = &ValueImpl{}
		} else {
			ret.value = NewValue(getInitializeValue(s.Type))
		}
	case models.OriginView:
		if s.value != nil {
			ret.value, _ = s.value.copy()
		} else {
			if !s.Type.IsPtrType() {
				ret.value = NewValue(getInitializeValue(s.Type))
			}
		}
	default:
		slog.Warn("Field.Copy: unsupported viewSpec", "field", s.GetName(), "viewSpec", viewSpec)
	}

	return
}

func (s *Field) verifyAutoIncrement(typeVal models.TypeDeclare) *cd.Error {
	switch typeVal {
	case models.TypeBooleanValue,
		models.TypeStringValue,
		models.TypeDateTimeValue,
		models.TypeFloatValue,
		models.TypeDoubleValue,
		models.TypeStructValue,
		models.TypeSliceValue:
		return cd.NewError(cd.Unexpected, fmt.Sprintf("illegal auto_increment field type, type:%v", typeVal))
	default:
	}

	return nil
}

func (s *Field) verifyUUID(typeVal models.TypeDeclare) *cd.Error {
	if typeVal != models.TypeStringValue {
		return cd.NewError(cd.Unexpected, fmt.Sprintf("illegal uuid field type, type:%v", typeVal))
	}

	return nil
}

func (s *Field) verifySnowflake(typeVal models.TypeDeclare) *cd.Error {
	if typeVal != models.TypeBigIntegerValue {
		return cd.NewError(cd.Unexpected, fmt.Sprintf("illegal snowflake field type, type:%v", typeVal))
	}

	return nil
}

func (s *Field) verifyDateTime(typeVal models.TypeDeclare) *cd.Error {
	if typeVal != models.TypeDateTimeValue {
		return cd.NewError(cd.Unexpected, fmt.Sprintf("illegal datetime field type, type:%v", typeVal))
	}

	return nil
}

func (s *Field) verifyPK(typeVal models.TypeDeclare) *cd.Error {
	switch typeVal {
	case models.TypeStructValue, models.TypeSliceValue:
		return cd.NewError(cd.Unexpected, fmt.Sprintf("illegal primary key field type, type:%v", typeVal))
	default:
	}

	return nil
}

func (s *Field) verify() (err *cd.Error) {
	if s.Type == nil {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal filed, field type is null, name:%v", s.Name))
		return
	}

	if s.Spec == nil {
		return
	}

	val := s.Type.GetValue()
	if s.Spec.GetValueDeclare() == models.AutoIncrement {
		err = s.verifyAutoIncrement(val)
		if err != nil {
			return
		}
	}

	if s.Spec.GetValueDeclare() == models.UUID {
		err = s.verifyUUID(val)
		if err != nil {
			return
		}
	}

	if s.Spec.GetValueDeclare() == models.Snowflake {
		err = s.verifySnowflake(val)
		if err != nil {
			return
		}
	}

	if s.Spec.GetValueDeclare() == models.DateTime {
		err = s.verifyDateTime(val)
		if err != nil {
			return
		}
	}

	if s.Spec.IsPrimaryKey() {
		err = s.verifyPK(val)
	}

	return
}

func compareItem(l, r *Field) bool {
	if l.Name != r.Name {
		return false
	}
	if l.ShowName != r.ShowName {
		return false
	}

	if !compareType(l.Type, r.Type) {
		return false
	}
	if !compareSpec(l.Spec, r.Spec) {
		return false
	}

	return true
}

func (s *FieldValue) IsValid() bool {
	return isValid(s.Value)
}

func (s *FieldValue) IsZero() bool {
	return isZero(s.Value)
}

func (s *FieldValue) Set(val any) {
	s.Assigned = true
	if val == nil {
		s.Value = nil
		return
	}

	switch val.(type) {
	case *ObjectValue, *SliceObjectValue:
		s.Value = val
	case ObjectValue:
		objectVal := val.(ObjectValue)
		s.Value = &objectVal
	case SliceObjectValue:
		sliceObjectVal := val.(SliceObjectValue)
		s.Value = &sliceObjectVal
	default:
		if !isSupportedBasicValue(val) {
			panic(fmt.Sprintf("illegal value:%+v", val))
		}

		s.Value = val
	}
}

func (s *FieldValue) Get() any {
	return s.Value
}

func (s *FieldValue) GetName() string {
	return s.Name
}

func (s *FieldValue) GetValue() models.Value {
	return &ValueImpl{value: s.Value, assigned: s.Assigned || !isZero(s.Value)}
}

func (s *FieldValue) copy() (ret *FieldValue) {
	if s.Value == nil {
		ret = &FieldValue{
			Name:     s.Name,
			Value:    nil,
			Assigned: s.Assigned,
		}
		return
	}

	switch v := s.Value.(type) {
	case *ObjectValue:
		ret = &FieldValue{
			Name:     s.Name,
			Value:    v.Copy(),
			Assigned: s.Assigned,
		}
		return
	case *SliceObjectValue:
		ret = &FieldValue{
			Name:     s.Name,
			Value:    v.Copy(),
			Assigned: s.Assigned,
		}
		return ret
	default:
		copiedVal, copiedErr := utils.DeepCopy(s.Value)
		if copiedErr != nil {
			panic(copiedErr)
		}
		ret = &FieldValue{
			Name:     s.Name,
			Value:    copiedVal,
			Assigned: s.Assigned,
		}
		return
	}
}
