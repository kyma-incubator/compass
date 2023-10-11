package operations

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/asserters"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/context-keys"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
	"testing"
)

type AddConstraintOperation struct {
	name            string
	constraintType  graphql.ConstraintType
	targetOperation graphql.TargetOperation
	operator        string
	resourceType    graphql.ResourceType
	resourceSubtype string
	inputTemplate   string
	constraintScope graphql.ConstraintScope
	tenantID        string
	constraintID    string
	asserters       []asserters.Asserter
}

func NewAddConstraintOperation(name string) *AddConstraintOperation {
	return &AddConstraintOperation{name: name, constraintType: graphql.ConstraintTypePre, targetOperation: graphql.TargetOperationAssignFormation, constraintScope: graphql.ConstraintScopeFormationType}
}

func (o *AddConstraintOperation) WithType(constraintType graphql.ConstraintType) *AddConstraintOperation {
	o.constraintType = constraintType
	return o
}

func (o *AddConstraintOperation) WithTargetOperation(targetOperation graphql.TargetOperation) *AddConstraintOperation {
	o.targetOperation = targetOperation
	return o
}

func (o *AddConstraintOperation) WithOperator(operator string) *AddConstraintOperation {
	o.operator = operator
	return o
}

func (o *AddConstraintOperation) WithResourceType(resourceType graphql.ResourceType) *AddConstraintOperation {
	o.resourceType = resourceType
	return o
}

func (o *AddConstraintOperation) WithResourceSubtype(resourceSubtype string) *AddConstraintOperation {
	o.resourceSubtype = resourceSubtype
	return o
}

func (o *AddConstraintOperation) WithInputTemplate(inputTemplate string) *AddConstraintOperation {
	o.inputTemplate = inputTemplate
	return o
}

func (o *AddConstraintOperation) WithScope(constraintScope graphql.ConstraintScope) *AddConstraintOperation {
	o.constraintScope = constraintScope
	return o
}
func (o *AddConstraintOperation) WithTenant(tenantID string) *AddConstraintOperation {
	o.tenantID = tenantID
	return o
}

func (o *AddConstraintOperation) WithAsserters(asserters ...asserters.Asserter) *AddConstraintOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *AddConstraintOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationTemplateID := ctx.Value(context_keys.FormationTemplateIDKey).(string)

	in := graphql.FormationConstraintInput{
		Name:            o.name,
		ConstraintType:  o.constraintType,
		TargetOperation: o.targetOperation,
		Operator:        o.operator,
		ResourceType:    o.resourceType,
		ResourceSubtype: o.resourceSubtype,
		InputTemplate:   o.inputTemplate,
		ConstraintScope: o.constraintScope,
	}

	constraint := fixtures.CreateFormationConstraint(t, ctx, gqlClient, in)
	o.constraintID = constraint.ID
	fixtures.AttachConstraintToFormationTemplate(t, ctx, gqlClient, constraint.ID, formationTemplateID)
	t.Logf("Created formation constraint with name: %s with type: %s for operation: %s", constraint.Name, constraint.ConstraintType, constraint.TargetOperation)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *AddConstraintOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationTemplateID := ctx.Value(context_keys.FormationTemplateIDKey).(string)

	fixtures.DetachConstraintFromFormationTemplate(t, ctx, gqlClient, o.constraintID, formationTemplateID)
	fixtures.CleanupFormationConstraint(t, ctx, gqlClient, o.constraintID)
}

func (o *AddConstraintOperation) Operation() Operation {
	return o
}
