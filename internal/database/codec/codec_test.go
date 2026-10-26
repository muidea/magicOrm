package codec

import (
	"reflect"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type codecStatus struct {
	ID   int    `orm:"id key auto"`
	Name string `orm:"name"`
}

type codecUser struct {
	ID     int            `orm:"id key auto"`
	Name   string         `orm:"name"`
	Tags   []string       `orm:"tags"`
	Status *codecStatus   `orm:"status"`
	Groups []*codecStatus `orm:"groups"`
}

type codecIdentifier struct {
	name string
}

func (i codecIdentifier) GetName() string        { return i.name }
func (i codecIdentifier) GetPkgPath() string     { return "codec.test" }
func (i codecIdentifier) GetDescription() string { return i.name }

type codecTestField struct {
	name string
	typ  models.Type
}

func (f *codecTestField) GetName() string                { return f.name }
func (f *codecTestField) GetShowName() string            { return f.name }
func (f *codecTestField) GetDescription() string         { return f.name }
func (f *codecTestField) GetType() models.Type           { return f.typ }
func (f *codecTestField) GetSpec() models.Spec           { return nil }
func (f *codecTestField) GetValue() models.Value         { return nil }
func (f *codecTestField) SetValue(any) *cd.Error         { return nil }
func (f *codecTestField) GetSliceValue() []models.Value  { return nil }
func (f *codecTestField) AppendSliceValue(any) *cd.Error { return nil }
func (f *codecTestField) Reset()                         {}

type codecTestModel struct {
	name string
}

func (m *codecTestModel) GetName() string                      { return m.name }
func (m *codecTestModel) GetShowName() string                  { return m.name }
func (m *codecTestModel) GetPkgPath() string                   { return "codec.test" }
func (m *codecTestModel) GetPkgKey() string                    { return m.GetPkgPath() + "/" + m.name }
func (m *codecTestModel) GetDescription() string               { return m.name }
func (m *codecTestModel) GetFields() models.Fields             { return nil }
func (m *codecTestModel) SetFieldValue(string, any) *cd.Error  { return nil }
func (m *codecTestModel) SetPrimaryFieldValue(any) *cd.Error   { return nil }
func (m *codecTestModel) GetPrimaryField() models.Field        { return nil }
func (m *codecTestModel) GetField(string) models.Field         { return nil }
func (m *codecTestModel) Interface(bool) any                   { return nil }
func (m *codecTestModel) Copy(models.ViewDeclare) models.Model { return m }
func (m *codecTestModel) Reset()                               {}

func buildCodecModel(t *testing.T) (Codec, models.Model) {
	t.Helper()

	modelProvider := provider.NewLocalProvider("tenant", nil)
	if _, err := modelProvider.RegisterModel(&codecStatus{}); err != nil {
		t.Fatalf("register status model failed: %v", err)
	}
	if _, err := modelProvider.RegisterModel(&codecUser{}); err != nil {
		t.Fatalf("register user model failed: %v", err)
	}

	model, err := modelProvider.GetEntityModel(&codecUser{
		ID:   10,
		Name: "demo",
		Tags: []string{"alpha", "beta"},
		Status: &codecStatus{
			ID:   7,
			Name: "active",
		},
		Groups: []*codecStatus{
			{ID: 1, Name: "g1"},
			{ID: 2, Name: "g2"},
		},
	}, true)
	if err != nil {
		t.Fatalf("build entity model failed: %v", err)
	}

	return New(modelProvider, "tenant"), model
}

func TestCodecTableNameConstruction(t *testing.T) {
	codec, _ := buildCodecModel(t)

	if got := codec.ConstructModelTableName(codecIdentifier{name: "user"}); got != "tenant_User" {
		t.Fatalf("unexpected model table name: %s", got)
	}
	if got := codec.ConstructModelTableName(codecIdentifier{}); got != "" {
		t.Fatalf("expected empty identifier name to produce empty table name, got %q", got)
	}
}

func TestCodecValuePackingAndExtraction(t *testing.T) {
	codec, model := buildCodecModel(t)

	nameField := model.GetField("name")
	packedName, err := codec.PackedBasicFieldValue(nameField, nameField.GetValue())
	if err != nil || packedName.(string) != "demo" {
		t.Fatalf("pack scalar basic value failed: %v %v", packedName, err)
	}

	tagsField := model.GetField("tags")
	packedTags, err := codec.PackedBasicFieldValue(tagsField, tagsField.GetValue())
	if err != nil {
		t.Fatalf("pack basic slice value failed: %v", err)
	}
	if packedTags.(string) == "" {
		t.Fatal("expected packed slice value to be a JSON string")
	}

	extractedTags, err := codec.ExtractBasicFieldValue(tagsField, []byte(`["alpha","beta"]`))
	if err != nil {
		t.Fatalf("extract basic slice value failed: %v", err)
	}
	if !reflect.DeepEqual(extractedTags, []string{"alpha", "beta"}) {
		t.Fatalf("unexpected extracted slice value: %#v", extractedTags)
	}
	if _, err := codec.ExtractBasicFieldValue(tagsField, ""); err != nil {
		t.Fatalf("expected empty slice payload to be handled, got %v", err)
	}

	statusField := model.GetField("status")
	packedStatus, err := codec.PackedStructFieldValue(statusField, statusField.GetValue())
	if err != nil {
		t.Fatalf("pack struct field failed: %v", err)
	}
	if packedStatus.(int) != 7 {
		t.Fatalf("unexpected packed struct primary value: %v", packedStatus)
	}

	groupsField := model.GetField("groups")
	packedGroups, err := codec.PackedSliceStructFieldValue(groupsField, groupsField.GetValue())
	if err != nil {
		t.Fatalf("pack slice struct field failed: %v", err)
	}
	if !reflect.DeepEqual(packedGroups, []any{1, 2}) {
		t.Fatalf("unexpected packed slice struct values: %#v", packedGroups)
	}
}

func TestCodecErrorPaths(t *testing.T) {
	codec, model := buildCodecModel(t)

	if _, err := codec.PackedBasicFieldValue(model.GetField("status"), model.GetField("status").GetValue()); err == nil {
		t.Fatal("expected packing non-basic field as basic to fail")
	}
	if _, err := codec.PackedStructFieldValue(model.GetField("name"), model.GetField("name").GetValue()); err == nil {
		t.Fatal("expected struct packing on basic field to fail")
	}
	if _, err := codec.PackedSliceStructFieldValue(model.GetField("name"), model.GetField("name").GetValue()); err == nil {
		t.Fatal("expected slice struct packing on basic field to fail")
	}
	if _, err := codec.ExtractBasicFieldValue(model.GetField("tags"), 123); err == nil {
		t.Fatal("expected invalid raw slice value to fail extraction")
	}

	groupField := model.GetField("groups")
	if _, err := codec.ConstructRelationTableName(&codecTestModel{}, &codecTestField{name: groupField.GetName(), typ: groupField.GetType()}); err == nil {
		t.Fatal("expected invalid relation identifier to fail")
	}
}
