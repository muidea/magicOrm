package utils

import (
	"github.com/muidea/magicOrm/models"
	"testing"
)

func TestPortableConstraintSyntaxAndGenericNumbers(t *testing.T) {
	valid := []string{"", "req,ro", "min=0,max=10", "range=-1:2", "in=a:b", "in=pending):done", `re=^[a-z]{2,4}$`, `re=^[),]+$`, `min=1,re=^hello,world$`, `re="^a,max=3$"`, `re=x,ro,max=4`}
	for _, s := range valid {
		if _, err := ParseConstraintsChecked(s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
	invalid := []string{"min=", "range=10:1", "min=2,max=1", "min=NaN", "ro,wo", "in=a:a", "in=", "req=1", "unknown", "min=1,min=2", "range=1:2,max=5", "re=["}
	for _, s := range invalid {
		if _, err := ParseConstraintsChecked(s); err == nil {
			t.Fatalf("accepted %s", s)
		}
	}
	c, _ := ParseConstraintsChecked("range=2:4")
	for _, v := range []any{int8(3), uint64(3), float32(3)} {
		if err := NewValueValidator().ValidateValue(v, c.Directives()); err != nil {
			t.Fatal(err)
		}
	}
	for _, v := range []any{int8(1), uint64(6), float32(6)} {
		if err := NewValueValidator().ValidateValue(v, c.Directives()); err == nil {
			t.Fatalf("accepted %v", v)
		}
	}
	malformed := ParseConstraints("min")
	if err := NewValueValidator().ValidateValue(3, malformed.Directives()); err == nil {
		t.Fatal("missing argument accepted")
	}
	r, _ := ParseConstraintsChecked(`re=^[),]+$`)
	d, _ := r.Get(models.KeyRegexp)
	if d.Args()[0] != `^[),]+$` {
		t.Fatal("regex damaged")
	}
}

func TestQuotedRegexRetainsLiteralDirectiveLikeText(t *testing.T) {
	constraints, err := ParseConstraintsChecked(`min=1,re="^a,max=3$"`)
	if err != nil {
		t.Fatal(err)
	}
	directive, _ := constraints.Get(models.KeyRegexp)
	if directive.Args()[0] != "^a,max=3$" {
		t.Fatal("quoted pattern damaged")
	}
	if err := NewValueValidator().ValidateValue("a,max=3", constraints.Directives()); err != nil {
		t.Fatal(err)
	}
}
