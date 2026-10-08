package models

// ConstraintViolation retains the failed directive without retaining the input
// value. Validators can wrap it while callers recover its identity with errors.As.
type ConstraintViolation struct {
	Key   Key
	Cause error
}

func (e *ConstraintViolation) Error() string { return e.Cause.Error() }
func (e *ConstraintViolation) Unwrap() error { return e.Cause }
