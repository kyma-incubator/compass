package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"
)

const IsNotAssignedToAnyFormationOfTypeOperator = "IsNotAssignedToAnyFormationOfType"

func TestCreateFormationConstraint(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	in := graphql.FormationConstraintInput{
		Name:            "test_constraint",
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        "IsNotAssignedToAnyFormationOfType",
		ResourceType:    graphql.ResourceTypeTenant,
		ResourceSubtype: "subaccount",
		InputTemplate:   "{\\\"formation_template_id\\\": \\\"{{.FormationTemplateID}}\\\",\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\"}",
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}

	t.Logf("Create formation constraint")

	formationConstraintInputGQLString, err := testctx.Tc.Graphqlizer.FormationConstraintInputToGQL(in)
	require.NoError(t, err)
	createRequest := fixtures.FixCreateFormationConstraintRequest(formationConstraintInputGQLString)
	formationConstraint := graphql.FormationConstraint{}
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, createRequest, &formationConstraint))
	require.NotEmpty(t, formationConstraint.ID)
	defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, formationConstraint.ID)

	saveExample(t, createRequest.Query(), "create formation constraint")

	expectedConstraint := &graphql.FormationConstraint{
		Name:            "test_constraint",
		ConstraintType:  string(graphql.ConstraintTypePre),
		TargetOperation: string(graphql.TargetOperationAssignFormation),
		Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
		ResourceType:    string(graphql.ResourceTypeTenant),
		ResourceSubtype: "subaccount",
		InputTemplate:   "{\"formation_template_id\": \"{{.FormationTemplateID}}\",\"resource_type\": \"{{.ResourceType}}\",\"resource_subtype\": \"{{.ResourceSubtype}}\",\"resource_id\": \"{{.ResourceID}}\",\"tenant\": \"{{.TenantID}}\"}",
		ConstraintScope: string(graphql.ConstraintScopeFormationType),
	}
	assertConstraint(t, expectedConstraint, &formationConstraint)
}

func TestDeleteFormationConstraint(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	in := graphql.FormationConstraintInput{
		Name:            "test_constraint",
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
		ResourceType:    graphql.ResourceTypeTenant,
		ResourceSubtype: "subaccount",
		InputTemplate:   "{\\\"formation_template_id\\\": \\\"{{.FormationTemplateID}}\\\",\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\"}",
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}

	t.Logf("Create formation constraint with name: %s", in.Name)
	constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, in)
	require.NotEmpty(t, constraint.ID)

	// WHEN
	t.Logf("Delete formation constraint with name: %s", in.Name)
	deleteRequest := fixtures.FixDeleteFormationConstraintRequest(constraint.ID)

	formationConstraint := graphql.FormationConstraint{}
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, deleteRequest, &formationConstraint)

	assertions.AssertNoErrorForOtherThanNotFound(t, err)
	saveExample(t, deleteRequest.Query(), "delete formation constraint")
}

func TestListFormationConstraints(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	firstConstraint := graphql.FormationConstraintInput{
		Name:            "test_constraint",
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
		ResourceType:    graphql.ResourceTypeTenant,
		ResourceSubtype: "subaccount",
		InputTemplate:   "{\\\"formation_template_id\\\": \\\"{{.FormationTemplateID}}\\\",\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\"}",
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}

	t.Logf("Create formation constraint with name: %s", firstConstraint.Name)
	constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, firstConstraint)
	defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraint.ID)
	require.NotEmpty(t, constraint.ID)

	secondConstraint := graphql.FormationConstraintInput{
		Name:            "test_constraint_second",
		ConstraintType:  graphql.ConstraintTypePost,
		TargetOperation: graphql.TargetOperationDeleteFormation,
		Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
		ResourceType:    graphql.ResourceTypeTenant,
		ResourceSubtype: "subaccount",
		InputTemplate:   "{\\\"formation_template_id\\\": \\\"{{.FormationTemplateID}}\\\",\\\"formation_type\\\": \\\"{{.FormationType}}\\\",\\\"formation_name\\\": \\\"{{.FormationName}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\"}",
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}

	t.Logf("Create formation constraint with name: %s", secondConstraint.Name)
	constraintSecond := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, secondConstraint)
	defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraintSecond.ID)
	require.NotEmpty(t, constraintSecond.ID)

	queryRequest := fixtures.FixQueryFormationConstraintsRequest()

	var formationConstraints []*graphql.FormationConstraint
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryRequest, &formationConstraints))
	require.Len(t, formationConstraints, 2)

	saveExample(t, queryRequest.Query(), "list formation constraints")

	expectedConstraints := map[string]*graphql.FormationConstraint{
		"test_constraint": {
			Name:            "test_constraint",
			ConstraintType:  string(graphql.ConstraintTypePre),
			TargetOperation: string(graphql.TargetOperationAssignFormation),
			Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
			ResourceType:    string(graphql.ResourceTypeTenant),
			ResourceSubtype: "subaccount",
			InputTemplate:   "{\"formation_template_id\": \"{{.FormationTemplateID}}\",\"resource_type\": \"{{.ResourceType}}\",\"resource_subtype\": \"{{.ResourceSubtype}}\",\"resource_id\": \"{{.ResourceID}}\",\"tenant\": \"{{.TenantID}}\"}",
			ConstraintScope: string(graphql.ConstraintScopeFormationType),
		},
		"test_constraint_second": {
			Name:            "test_constraint_second",
			ConstraintType:  string(graphql.ConstraintTypePost),
			TargetOperation: string(graphql.TargetOperationDeleteFormation),
			Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
			ResourceType:    string(graphql.ResourceTypeTenant),
			ResourceSubtype: "subaccount",
			InputTemplate:   "{\"formation_template_id\": \"{{.FormationTemplateID}}\",\"formation_type\": \"{{.FormationType}}\",\"formation_name\": \"{{.FormationName}}\",\"tenant\": \"{{.TenantID}}\"}",
			ConstraintScope: string(graphql.ConstraintScopeFormationType),
		},
	}

	for _, f := range formationConstraints {
		assertConstraint(t, expectedConstraints[f.Name], f)
	}
}

func TestListFormationConstraintsForFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "formation-template-name"
	formationTemplateInput := fixtures.FixFormationTemplateInput(formationTemplateName)

	formationTemplate := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplate.ID)

	secondFormationTemplateName := "second-formation-template-name"
	secondFormationTemplateInput := fixtures.FixFormationTemplateInput(secondFormationTemplateName)

	secondFormationTemplate := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, secondFormationTemplateInput)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, secondFormationTemplate.ID)

	firstConstraint := graphql.FormationConstraintInput{
		Name:            "test_constraint",
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
		ResourceType:    graphql.ResourceTypeTenant,
		ResourceSubtype: "subaccount",
		InputTemplate:   "{\\\"formation_template_id\\\": \\\"{{.FormationTemplateID}}\\\",\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\"}",
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}

	t.Logf("Create formation constraint with name: %s", firstConstraint.Name)
	constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, firstConstraint)
	defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraint.ID)
	require.NotEmpty(t, constraint.ID)

	t.Logf("Attaching constraint to formation template")
	fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, formationTemplate.ID)

	secondConstraint := graphql.FormationConstraintInput{
		Name:            "test_constraint_second",
		ConstraintType:  graphql.ConstraintTypePost,
		TargetOperation: graphql.TargetOperationDeleteFormation,
		Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
		ResourceType:    graphql.ResourceTypeTenant,
		ResourceSubtype: "subaccount",
		InputTemplate:   "{\\\"formation_template_id\\\": \\\"{{.FormationTemplateID}}\\\",\\\"formation_type\\\": \\\"{{.FormationType}}\\\",\\\"formation_name\\\": \\\"{{.FormationName}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\"}",
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}

	t.Logf("Create formation constraint with name: %s", secondConstraint.Name)
	constraintSecond := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, secondConstraint)
	defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraintSecond.ID)
	require.NotEmpty(t, constraintSecond.ID)

	t.Logf("Attaching second constraint to formation template")
	fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, constraintSecond.ID, formationTemplate.ID)

	constraintForOtherTemplateInput := graphql.FormationConstraintInput{
		Name:            "test_constraint_other_template",
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
		ResourceType:    graphql.ResourceTypeTenant,
		ResourceSubtype: "subaccount",
		InputTemplate:   "{\\\"formation_template_id\\\": \\\"{{.FormationTemplateID}}\\\",\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\"}",
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}

	t.Logf("Create formation constraint with name: %s", constraintForOtherTemplateInput.Name)
	constraintForOtherTemplate := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, constraintForOtherTemplateInput)
	defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraintForOtherTemplate.ID)
	require.NotEmpty(t, constraintForOtherTemplate.ID)

	t.Logf("Attaching constraintForOtherTemplate to formation template other")
	fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, constraintForOtherTemplate.ID, secondFormationTemplate.ID)

	queryRequest := fixtures.FixQueryFormationConstraintsForFormationTemplateRequest(formationTemplate.ID)

	var formationConstraints []*graphql.FormationConstraint
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryRequest, &formationConstraints))
	require.Len(t, formationConstraints, 2)

	saveExample(t, queryRequest.Query(), "list formation constraints for formation template")

	expectedConstraints := map[string]*graphql.FormationConstraint{
		"test_constraint": {
			Name:            "test_constraint",
			ConstraintType:  string(graphql.ConstraintTypePre),
			TargetOperation: string(graphql.TargetOperationAssignFormation),
			Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
			ResourceType:    string(graphql.ResourceTypeTenant),
			ResourceSubtype: "subaccount",
			InputTemplate:   "{\"formation_template_id\": \"{{.FormationTemplateID}}\",\"resource_type\": \"{{.ResourceType}}\",\"resource_subtype\": \"{{.ResourceSubtype}}\",\"resource_id\": \"{{.ResourceID}}\",\"tenant\": \"{{.TenantID}}\"}",
			ConstraintScope: string(graphql.ConstraintScopeFormationType),
		},
		"test_constraint_second": {
			Name:            "test_constraint_second",
			ConstraintType:  string(graphql.ConstraintTypePost),
			TargetOperation: string(graphql.TargetOperationDeleteFormation),
			Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
			ResourceType:    string(graphql.ResourceTypeTenant),
			ResourceSubtype: "subaccount",
			InputTemplate:   "{\"formation_template_id\": \"{{.FormationTemplateID}}\",\"formation_type\": \"{{.FormationType}}\",\"formation_name\": \"{{.FormationName}}\",\"tenant\": \"{{.TenantID}}\"}",
			ConstraintScope: string(graphql.ConstraintScopeFormationType),
		},
	}

	for _, f := range formationConstraints {
		assertConstraint(t, expectedConstraints[f.Name], f)
	}
}

func assertConstraint(t *testing.T, expected, actual *graphql.FormationConstraint) {
	require.NotEmpty(t, actual.ID)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.ConstraintType, actual.ConstraintType)
	require.Equal(t, expected.TargetOperation, actual.TargetOperation)
	require.Equal(t, expected.Operator, actual.Operator)
	require.Equal(t, expected.ResourceType, actual.ResourceType)
	require.Equal(t, expected.ResourceSubtype, actual.ResourceSubtype)
	require.Equal(t, expected.InputTemplate, actual.InputTemplate)
	require.Equal(t, expected.ConstraintScope, actual.ConstraintScope)
}
