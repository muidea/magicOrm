package utils

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/models"
)

func TestConstraintFailurePreservesCustomCauseWithoutInput(t *testing.T) {
	validator := NewValueValidator()
	cause := errors.New("custom validation rejected")
	validator.Register(models.KeyMax, func(any, []string) error { return cause })
	err := validator.ValidateValue("private-input", ParseConstraints("max=10").Directives())
	if !errors.Is(err, cause) {
		t.Fatal("custom cause was lost")
	}
	key, message := ConstraintErrorDetails(fmt.Errorf("validation: %w", err))
	if key != models.KeyMax || !strings.Contains(message, cause.Error()) || strings.Contains(message, "private-input") {
		t.Fatalf("unexpected validation details: %s %s", key, message)
	}
	_, message = ConstraintErrorDetails(cause)
	if message != cause.Error() {
		t.Fatal("plain external validator reason lost")
	}
}
