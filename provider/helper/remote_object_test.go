package helper

import (
	"encoding/json"
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/stretchr/testify/assert"
)

type Simple struct {
	ID     int64   `orm:"id key" view:"detail,lite"`
	Name   string  `orm:"name" view:"detail,lite"`
	Desc   *string `orm:"desc" view:"detail"`
	Age    uint8   `orm:"age" view:"detail,lite"`
	Flag   bool    `orm:"flag" view:"detail,lite"`
	Add    []int   `orm:"add" view:"detail,lite"`
	AddPtr *[]int  `orm:"addPtr" view:"detail,lite"`
}

type ExtInfo struct {
	ID       int64     `orm:"id key" view:"detail,lite"`
	Name     string    `orm:"name" view:"detail,lite"`
	Obj      Simple    `orm:"obj" view:"detail,lite"`
	ObjPtr   *Simple   `orm:"objPtr" view:"detail,lite"`
	ObjArray []*Simple `orm:"array" view:"detail,lite"`
}

type Complex struct {
	ID       int64      `orm:"id key" view:"detail,lite"`
	Name     string     `orm:"name" view:"detail,lite"`
	Info     ExtInfo    `orm:"info" view:"detail"`
	InfoPtr  *ExtInfo   `orm:"infoPtr" view:"detail"`
	Array    []ExtInfo  `orm:"array" view:"detail"`
	ArrayPtr []*ExtInfo `orm:"arrayPtr" view:"detail"`
}

type qualifiedSubject string
type revisionID int64

type publicJSONSubject string

func (s publicJSONSubject) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

type NamedBasicAlias struct {
	ID        int64              `orm:"id key" view:"detail,lite"`
	Subject   qualifiedSubject   `orm:"subject" view:"detail"`
	Reviewer  *qualifiedSubject  `orm:"reviewer" view:"detail"`
	Revision  revisionID         `orm:"revision" view:"detail"`
	Audiences []qualifiedSubject `orm:"audiences" view:"detail"`
}

type NamedJSONBasicAlias struct {
	ID        int64               `orm:"id key" view:"detail,lite"`
	Subject   publicJSONSubject   `orm:"subject" view:"detail"`
	Reviewer  *publicJSONSubject  `orm:"reviewer" view:"detail"`
	Audiences []publicJSONSubject `orm:"audiences" view:"detail"`
}

type IgnoredProjection struct {
	ID          int64    `orm:"id key" view:"detail,lite"`
	Name        string   `orm:"name" view:"detail,lite"`
	Credentials []string `json:"credentials" orm:"-" view:"detail"`
}

func TestSpec(t *testing.T) {
	spec := ""
	_, err := getOrmSpec(spec)
	if err != nil {
		t.Errorf("illegal spec value")
		return
	}

	spec = "test"
	itemSpec, err := getOrmSpec(spec)
	if err != nil {
		t.Errorf("illegal spec value")
		return
	}
	if itemSpec.GetFieldName() != "test" {
		t.Errorf("illegal spec name")
		return
	}
	if itemSpec.IsPrimaryKey() {
		t.Errorf("illegal spec define")
		return
	}
	if itemSpec.GetValueDeclare() == models.AutoIncrement {
		t.Errorf("illegal spec define")
		return
	}

	spec = "test auto key"
	itemSpec, err = getOrmSpec(spec)
	if err != nil {
		t.Errorf("illegal spec value")
		return
	}
	if itemSpec.GetFieldName() != "test" {
		t.Errorf("illegal spec name")
		return
	}
	if !itemSpec.IsPrimaryKey() {
		t.Errorf("illegal spec define")
		return
	}
	if itemSpec.GetValueDeclare() != models.AutoIncrement {
		t.Errorf("illegal spec define")
		return
	}
}

func TestSimpleObjInfo(t *testing.T) {
	desc := "obj_desc"
	obj := Simple{Name: "obj", Desc: &desc, Age: 240}

	info, err := GetObject(obj)
	if err != nil {
		t.Errorf("GetObject failed, err:%s", err.Error())
		return
	}
	if info.GetName() != "Simple" {
		t.Errorf("GetObject failed")
	}

	byteVal, byteErr := json.Marshal(info)
	if byteErr != nil {
		t.Errorf("marshal info failed, err:%s", byteErr.Error())
		return
	}

	info2 := &remote.Object{}
	byteErr = json.Unmarshal(byteVal, info2)
	if byteErr != nil {
		t.Errorf("marshal info failed, err:%s", byteErr.Error())
		return
	}

	if !remote.CompareObject(info, info2) {
		t.Errorf("unmarshal failed")
		return
	}
}

func TestNamedBasicAliasUsesCanonicalRemoteType(t *testing.T) {
	subject := qualifiedSubject("panel/account:1026173723334912")
	info, err := GetObject(NamedBasicAlias{Subject: subject, Reviewer: &subject, Revision: 7})
	if err != nil {
		t.Fatalf("GetObject failed: %s", err.Error())
	}

	tests := []struct {
		field   string
		name    string
		pointer bool
	}{
		{field: "subject", name: models.TypeStringName},
		{field: "reviewer", name: models.TypeStringName, pointer: true},
		{field: "revision", name: models.TypeBigIntegerName},
	}
	for _, test := range tests {
		field := info.GetField(test.field)
		if field == nil {
			t.Fatalf("field %s is unavailable", test.field)
		}
		if got := field.GetType().GetName(); got != test.name {
			t.Errorf("field %s type name = %q, want %q", test.field, got, test.name)
		}
		if got := field.GetType().GetPkgPath(); got != "" {
			t.Errorf("field %s package path = %q, want empty", test.field, got)
		}
		if got := field.GetType().IsPtrType(); got != test.pointer {
			t.Errorf("field %s pointer = %v, want %v", test.field, got, test.pointer)
		}
	}

	audiences := info.GetField("audiences")
	if audiences == nil || audiences.GetType().Elem().GetName() != models.TypeStringName || audiences.GetType().Elem().GetPkgPath() != "" {
		t.Fatalf("named string slice element was not canonicalized: %#v", audiences)
	}

	raw, marshalErr := json.Marshal(info)
	if marshalErr != nil {
		t.Fatalf("marshal object failed: %v", marshalErr)
	}
	roundTrip := &remote.Object{}
	if unmarshalErr := json.Unmarshal(raw, roundTrip); unmarshalErr != nil {
		t.Fatalf("unmarshal object failed: %v", unmarshalErr)
	}
	if !remote.CompareObject(info, roundTrip) {
		t.Fatal("named primitive type contract changed after JSON round trip")
	}
	if got := roundTrip.GetField("subject").GetType().GetValue(); got != models.TypeStringValue {
		t.Fatalf("round-trip subject type = %v, want string", got)
	}

	value, valueErr := GetObjectValue(NamedBasicAlias{Subject: subject, Reviewer: &subject, Revision: 7})
	if valueErr != nil {
		t.Fatalf("GetObjectValue failed: %s", valueErr.Error())
	}
	if got := value.GetFieldValue("subject"); got != "panel/account:1026173723334912" {
		t.Fatalf("subject value = %#v", got)
	}
}

func TestViewMaskUsesCanonicalValuesForNamedBasicTypes(t *testing.T) {
	mask, err := BuildViewMask(NamedJSONBasicAlias{}, models.DetailView)
	if err != nil {
		t.Fatalf("BuildViewMask failed: %s", err.Error())
	}

	if got := mask.GetFieldValue("subject"); got != "" {
		t.Fatalf("subject mask value = %#v, want canonical empty string", got)
	}
	if got := mask.GetFieldValue("reviewer"); got != "" {
		t.Fatalf("reviewer mask value = %#v, want canonical empty string", got)
	}
	if got, ok := mask.GetFieldValue("audiences").([]string); !ok || len(got) != 0 {
		t.Fatalf("audiences mask value = %#v, want canonical empty string slice", mask.GetFieldValue("audiences"))
	}

	raw, marshalErr := json.Marshal(mask)
	if marshalErr != nil {
		t.Fatalf("marshal view mask failed: %v", marshalErr)
	}
	roundTrip, decodeErr := remote.DecodeObjectValue(raw)
	if decodeErr != nil {
		t.Fatalf("decode view mask failed: %s", decodeErr.Error())
	}
	if got := roundTrip.GetFieldValue("subject"); got != "" {
		t.Fatalf("round-trip subject mask value = %#v, want empty string", got)
	}
	if got := roundTrip.GetFieldValue("reviewer"); got != "" {
		t.Fatalf("round-trip reviewer mask value = %#v, want empty string", got)
	}
}

func TestIgnoredORMFieldIsAbsentFromRemoteContracts(t *testing.T) {
	entity := IgnoredProjection{ID: 7, Name: "subscription", Credentials: []string{"secret"}}

	object, err := GetObject(entity)
	if err != nil {
		t.Fatalf("GetObject failed: %s", err.Error())
	}
	if object.GetField("credentials") != nil {
		t.Fatal("ignored field was included in remote object definition")
	}
	if object.GetField("name") == nil {
		t.Fatal("persisted field was omitted from remote object definition")
	}

	value, valueErr := GetObjectValue(entity)
	if valueErr != nil {
		t.Fatalf("GetObjectValue failed: %s", valueErr.Error())
	}
	if got := value.GetFieldValue("credentials"); got != nil {
		t.Fatalf("ignored field value = %#v, want nil", got)
	}
	if got := value.GetFieldValue("name"); got != "subscription" {
		t.Fatalf("persisted name = %#v", got)
	}

	mask, maskErr := BuildViewMask(entity, models.DetailView)
	if maskErr != nil {
		t.Fatalf("BuildViewMask failed: %s", maskErr.Error())
	}
	if got := mask.GetFieldValue("credentials"); got != nil {
		t.Fatalf("ignored field mask = %#v, want nil", got)
	}
}

func TestExtObjInfo(t *testing.T) {
	desc := "obj_desc"
	obj := Simple{Name: "obj", Desc: &desc}
	ext := &ExtInfo{Name: "extObj", Obj: obj, ObjArray: []*Simple{&obj, &obj}}

	info, err := GetObject(ext)
	if err != nil {
		t.Errorf("GetObject failed, err:%s", err.Error())
		return
	}

	if info.GetName() != "ExtInfo" {
		t.Errorf("get object failed")
		return
	}

	byteVal, byteErr := json.Marshal(info)
	if byteErr != nil {
		t.Errorf("marshal info failed, err:%s", byteErr.Error())
		return
	}

	eInfo := &remote.Object{}
	byteErr = json.Unmarshal(byteVal, eInfo)
	if byteErr != nil {
		t.Errorf("unmarshal ext failed, err:%s", byteErr.Error())
		return
	}

	if !remote.CompareObject(info, eInfo) {
		t.Errorf("unmarshal faile")
		return
	}
}

func TestGetObjectWithNilValue(t *testing.T) {
	var simplePtr *Simple = nil
	simpleValPtr, simpleValErr := GetObject(simplePtr)
	if simpleValErr != nil {
		t.Errorf("GetObject with nil value should return error, err:%s", simpleValErr.Error())
		return
	}
	assert.NotNil(t, simpleValPtr)

	var extPtr *ExtInfo = nil
	extValPtr, extValErr := GetObject(extPtr)
	if extValErr != nil {
		t.Errorf("GetObject with nil value should return error, err:%s", extValErr.Error())
		return
	}
	assert.NotNil(t, extValPtr)
}

func TestGetObjectWithStructPointers(t *testing.T) {
	desc := "obj_desc"
	obj := &Simple{Name: "obj", Desc: &desc, Age: 240}

	info, err := GetObject(obj)
	if err != nil {
		t.Errorf("GetObject failed with pointer, err:%s", err.Error())
		return
	}
	if info.GetName() != "Simple" {
		t.Errorf("GetObject failed with pointer")
	}

	// Test with a nested struct pointer
	ext := ExtInfo{Name: "extObj", ObjPtr: obj}
	extInfo, extErr := GetObject(ext)
	if extErr != nil {
		t.Errorf("GetObject failed with nested pointer, err:%s", extErr.Error())
		return
	}

	objPtrField := extInfo.GetField("objPtr")
	if objPtrField == nil {
		t.Errorf("Failed to get objPtr field")
		return
	}

	if !models.IsPtrField(objPtrField) {
		t.Errorf("objPtr field should be a pointer type")
		return
	}
}

func TestComplexObjInfo(t *testing.T) {
	desc := "obj_desc"
	simple := Simple{Name: "simple", Desc: &desc, Age: 240}
	ext := ExtInfo{Name: "extObj", Obj: simple, ObjPtr: &simple, ObjArray: []*Simple{&simple}}
	complex := Complex{
		Name:     "complex",
		Info:     ext,
		InfoPtr:  &ext,
		Array:    []ExtInfo{ext},
		ArrayPtr: []*ExtInfo{&ext},
	}

	info, err := GetObject(complex)
	if err != nil {
		t.Errorf("GetObject failed for complex obj, err:%s", err.Error())
		return
	}

	if info.GetName() != "Complex" {
		t.Errorf("GetObject failed for complex obj")
		return
	}

	// Check nested fields
	infoField := info.GetField("info")
	if infoField == nil {
		t.Errorf("Failed to get info field")
		return
	}

	if !models.IsStructField(infoField) {
		t.Errorf("info field should be a struct type")
		return
	}

	arrayField := info.GetField("array")
	if arrayField == nil {
		t.Errorf("Failed to get array field")
		return
	}

	if !models.IsSliceField(arrayField) {
		t.Errorf("array field should be a slice type")
		return
	}

	// Test serialization and deserialization
	byteVal, byteErr := json.Marshal(info)
	if byteErr != nil {
		t.Errorf("marshal complex info failed, err:%s", byteErr.Error())
		return
	}

	deserializedInfo := &remote.Object{}
	byteErr = json.Unmarshal(byteVal, deserializedInfo)
	if byteErr != nil {
		t.Errorf("unmarshal complex info failed, err:%s", byteErr.Error())
		return
	}

	if !remote.CompareObject(info, deserializedInfo) {
		t.Errorf("deserialization failed for complex object")
		return
	}
}
