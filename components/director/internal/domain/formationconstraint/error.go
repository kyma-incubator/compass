package formationconstraint

import "fmt"

type FormationConstraintError struct {
	ConstraintName string
	Reason         string
}

func (e FormationConstraintError) Error() string {
	return fmt.Sprintf("Formation constraint %q is not sattisfied due to: %s", e.ConstraintName, e.Reason)
}
