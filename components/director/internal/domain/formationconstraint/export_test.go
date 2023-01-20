package formationconstraint

import "context"

func (e *ConstraintEngine) Set() {
	e.operators = map[OperatorName]OperatorFunc{IsNotAssignedToAnyFormationOfTypeOperator: func(ctx context.Context, input OperatorInput) (bool, error) { return true, nil }}
}
