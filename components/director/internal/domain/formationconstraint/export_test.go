package formationconstraint

import "context"

func (e *ConstraintEngine) SetOperator(operator func(ctx context.Context, input OperatorInput) (bool, error)) {
	e.operators = map[OperatorName]OperatorFunc{IsNotAssignedToAnyFormationOfTypeOperator: operator}
}

func (e *ConstraintEngine) SetEmptyOperatorMap() {
	e.operators = map[OperatorName]OperatorFunc{}
}

func (e *ConstraintEngine) SetEmptyOperatorInputBuilderMap() {
	e.operatorInputConstructors = map[OperatorName]OperatorInputConstructor{}
}
