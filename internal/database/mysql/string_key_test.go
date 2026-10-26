package mysql

import (
	"strings"
	"testing"

	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestExplicitStringKeyCapacity(t *testing.T) {
	for _, test := range []struct {
		constraint, width string
		invalid           bool
	}{
		{"", "32", false}, {"max=64", "64", false}, {"max=255", "255", false},
		{"max=0", "", true}, {"max=256", "", true}, {"max=bogus", "", true},
	} {
		t.Run(test.constraint, func(t *testing.T) {
			p := provider.NewRemoteProvider("key-capacity", nil)
			model, err := p.RegisterModel(&remote.Object{Name: "operation", PkgPath: "/test", Fields: []*remote.Field{{Name: "id", Type: &remote.TypeImpl{Name: "string"}, Spec: &remote.SpecImpl{FieldName: "id", PrimaryKey: true, Constraint: test.constraint}}}})
			if err != nil {
				t.Fatal(err)
			}
			b := NewBuilder(p, codec.New(p, "acceptance"))
			ddl, err := b.BuildCreateTable(model)
			if test.invalid {
				if err == nil {
					t.Fatal("invalid key bound accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(ddl.SQL(), "VARCHAR("+test.width+")") {
				t.Fatal("key DDL capacity mismatch", ddl.SQL())
			}
			schema, err := b.(*Builder).DescribeTable(model)
			if err != nil {
				t.Fatal(err)
			}
			if len(schema.Columns) != 1 || !strings.HasSuffix(schema.Columns[0].Type, "("+test.width+")") {
				t.Fatal("schema capacity mismatch", schema)
			}
		})
	}
}
