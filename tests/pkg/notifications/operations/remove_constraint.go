package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	gcli "github.com/machinebox/graphql"
)

type RemoveConstraintOperation struct {
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

func NewRemoveConstraintOperation(name string) *RemoveConstraintOperation {
	return &RemoveConstraintOperation{name: name, constraintType: graphql.ConstraintTypePre, targetOperation: graphql.TargetOperationAssignFormation, constraintScope: graphql.ConstraintScopeFormationType}
}

func (o *RemoveConstraintOperation) WithType(constraintType graphql.ConstraintType) *RemoveConstraintOperation {
	o.constraintType = constraintType
	return o
}

func (o *RemoveConstraintOperation) WithTargetOperation(targetOperation graphql.TargetOperation) *RemoveConstraintOperation {
	o.targetOperation = targetOperation
	return o
}

func (o *RemoveConstraintOperation) WithOperator(operator string) *RemoveConstraintOperation {
	o.operator = operator
	return o
}

func (o *RemoveConstraintOperation) WithResourceType(resourceType graphql.ResourceType) *RemoveConstraintOperation {
	o.resourceType = resourceType
	return o
}

func (o *RemoveConstraintOperation) WithResourceSubtype(resourceSubtype string) *RemoveConstraintOperation {
	o.resourceSubtype = resourceSubtype
	return o
}

func (o *RemoveConstraintOperation) WithInputTemplate(inputTemplate string) *RemoveConstraintOperation {
	o.inputTemplate = inputTemplate
	return o
}

func (o *RemoveConstraintOperation) WithScope(constraintScope graphql.ConstraintScope) *RemoveConstraintOperation {
	o.constraintScope = constraintScope
	return o
}
func (o *RemoveConstraintOperation) WithTenant(tenantID string) *RemoveConstraintOperation {
	o.tenantID = tenantID
	return o
}

func (o *RemoveConstraintOperation) WithAsserters(asserters ...asserters.Asserter) *RemoveConstraintOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *RemoveConstraintOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationTemplateID := ctx.Value(context_keys.FormationTemplateIDKey).(string)
	formationTemplateName := ctx.Value(context_keys.FormationTemplateNameKey).(string)

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
	t.Logf("Created formation constraint with name: %s and type: %s for operation: %s", constraint.Name, constraint.ConstraintType, constraint.TargetOperation)
	o.constraintID = constraint.ID
	fixtures.AttachConstraintToFormationTemplate(t, ctx, gqlClient, constraint.ID, constraint.Name, formationTemplateID, formationTemplateName)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *RemoveConstraintOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationTemplateID := ctx.Value(context_keys.FormationTemplateIDKey).(string)

	fixtures.DetachConstraintFromFormationTemplate(t, ctx, gqlClient, o.constraintID, formationTemplateID)
	fixtures.CleanupFormationConstraint(t, ctx, gqlClient, o.constraintID)
}

func (o *RemoveConstraintOperation) Operation() Operation {
	return o
}
