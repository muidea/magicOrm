package constraints

import (
	"fmt"
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
	"github.com/muidea/magicOrm/validation/errors"
)

func TestCachedValidationDoesNotReuseAnotherValue(t *testing.T) {
	v := NewConstraintValidator(true)
	c := utils.ParseConstraints("min=1")
	if err := v.ValidateConstraints(int32(5), c, errors.ScenarioInsert); err != nil {
		t.Fatal(err)
	}
	if err := v.ValidateConstraints(int32(0), c, errors.ScenarioInsert); err == nil {
		t.Fatal("cached success admitted invalid value")
	}
	if err := v.RegisterCustomConstraint(models.KeyMin, func(any, []string) error { return fmt.Errorf("replacement rule") }); err != nil {
		t.Fatal(err)
	}
	if err := v.ValidateConstraints(int32(5), c, errors.ScenarioInsert); err == nil {
		t.Fatal("registered rule reused obsolete cached success")
	}
}
