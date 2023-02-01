package formationconstraint

import "fmt"

// ConstraintError is structured error containing the name of the Constraint and the reason for the error
type ConstraintError struct {
	ConstraintName string
	Reason         string
}

// Error implements the Error interface
func (e ConstraintError) Error() string {
	return fmt.Sprintf("Formation constraint %q is not satisfied due to: %s", e.ConstraintName, e.Reason)
}
