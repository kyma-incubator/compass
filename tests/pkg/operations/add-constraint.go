package operations

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/asserters"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
	"testing"
)

//in := graphql.FormationConstraintInput{
//	Name:            "mutate",
//	ConstraintType:  graphql.ConstraintTypePre,
//	TargetOperation: graphql.TargetOperationNotificationStatusReturned,
//	Operator:        "ConfigMutator",
//	ResourceType:    graphql.ResourceTypeApplication,
//	ResourceSubtype: "app-type-1",
//	InputTemplate:   "{\\\"configuration\\\":\\\"{\\\\\\\"tmp\\\\\\\":\\\\\\\"tmpval\\\\\\\"}\\\",\\\"state\\\":\\\"DELETING\\\",\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .FormationAssignment }}\\\"details_formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"details_reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}",
//	ConstraintScope: graphql.ConstraintScopeFormationType,
//}

//constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, in)

type AddConstraintOperation struct {
	name                string
	constraintType      graphql.ConstraintType
	targetOperation     graphql.TargetOperation
	operator            string
	resourceType        graphql.ResourceType
	resourceSubtype     string
	inputTemplate       string
	constraintScope     graphql.ConstraintScope
	formationTemplateID string
	tenantID            string
	constraintID        string
	asserters           []asserters.Asserter
}

func NewAddConstraintOperation(name string, constraintType graphql.ConstraintType, targetOperation graphql.TargetOperation, operator string, resourceType graphql.ResourceType, resourceSubtype string, inputTemplate string, constraintScope graphql.ConstraintScope, formationTemplateID string, tenantID string) *AddConstraintOperation {
	return &AddConstraintOperation{name: name, constraintType: constraintType, targetOperation: targetOperation, operator: operator, resourceType: resourceType, resourceSubtype: resourceSubtype, inputTemplate: inputTemplate, constraintScope: constraintScope, formationTemplateID: formationTemplateID, tenantID: tenantID}
}

func (o *AddConstraintOperation) WithAsserter(asserter asserters.Asserter) *AddConstraintOperation {
	o.asserters = append(o.asserters, asserter)
	return o
}

func (o *AddConstraintOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
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
	fixtures.AttachConstraintToFormationTemplate(t, ctx, gqlClient, constraint.ID, o.formationTemplateID)
	t.Logf("Created formation constraint")
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *AddConstraintOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.DetachConstraintFromFormationTemplate(t, ctx, gqlClient, o.constraintID, o.formationTemplateID)
	fixtures.CleanupFormationConstraint(t, ctx, gqlClient, o.constraintID)
}

func (o *AddConstraintOperation) Operation() Operation {
	return o
}
