package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"
)

const (
	IsNotAssignedToAnyFormationOfTypeOperator = "IsNotAssignedToAnyFormationOfType"
	DoesNotContainResourceOfSubtypeOperator   = "DoesNotContainResourceOfSubtype"
)

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
	saveExample(t, createRequest.Query(), "create formation constraint")

	formationConstraint := graphql.FormationConstraint{}
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, createRequest, &formationConstraint))
	require.NotEmpty(t, formationConstraint.ID)
	defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, formationConstraint.ID)

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
	saveExample(t, deleteRequest.Query(), "delete formation constraint")

	formationConstraint := graphql.FormationConstraint{}
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, deleteRequest, &formationConstraint)

	assertions.AssertNoErrorForOtherThanNotFound(t, err)
}

func TestFormationConstraint(t *testing.T) {
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

	queryRequest := fixtures.FixQueryFormationConstraintRequest(constraint.ID)
	saveExample(t, queryRequest.Query(), "query formation constraint")

	var actualFormationConstraint *graphql.FormationConstraint
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryRequest, &actualFormationConstraint))

	expectedConstraint := &graphql.FormationConstraint{
		ID:              constraint.ID,
		Name:            "test_constraint",
		ConstraintType:  string(graphql.ConstraintTypePre),
		TargetOperation: string(graphql.TargetOperationAssignFormation),
		Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
		ResourceType:    string(graphql.ResourceTypeTenant),
		ResourceSubtype: "subaccount",
		InputTemplate:   "{\"formation_template_id\": \"{{.FormationTemplateID}}\",\"resource_type\": \"{{.ResourceType}}\",\"resource_subtype\": \"{{.ResourceSubtype}}\",\"resource_id\": \"{{.ResourceID}}\",\"tenant\": \"{{.TenantID}}\"}",
		ConstraintScope: string(graphql.ConstraintScopeFormationType),
	}
	require.Equal(t, expectedConstraint, actualFormationConstraint)
}

func TestUpdateFormationConstraint(t *testing.T) {
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

	updateInput := graphql.FormationConstraintUpdateInput{
		InputTemplate: "{\\\"formation_template_id\\\": \\\"{{.FormationTemplateID}}\\\",\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\",\\\"key\\\": \\\"value\\\"}",
	}

	formationConstraintGQL, err := testctx.Tc.Graphqlizer.FormationConstraintUpdateInputToGQL(updateInput)
	require.NoError(t, err)

	updateRequest := fixtures.FixUpdateFormationConstraintRequest(constraint.ID, formationConstraintGQL)
	saveExample(t, updateRequest.Query(), "update formation constraint")

	var actualFormationConstraint *graphql.FormationConstraint
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, updateRequest, &actualFormationConstraint))

	expectedConstraint := &graphql.FormationConstraint{
		ID:              constraint.ID,
		Name:            "test_constraint",
		ConstraintType:  string(graphql.ConstraintTypePre),
		TargetOperation: string(graphql.TargetOperationAssignFormation),
		Operator:        IsNotAssignedToAnyFormationOfTypeOperator,
		ResourceType:    string(graphql.ResourceTypeTenant),
		ResourceSubtype: "subaccount",
		InputTemplate:   "{\"formation_template_id\": \"{{.FormationTemplateID}}\",\"resource_type\": \"{{.ResourceType}}\",\"resource_subtype\": \"{{.ResourceSubtype}}\",\"resource_id\": \"{{.ResourceID}}\",\"tenant\": \"{{.TenantID}}\",\"key\": \"value\"}",
		ConstraintScope: string(graphql.ConstraintScopeFormationType),
	}

	require.Equal(t, expectedConstraint, actualFormationConstraint)
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
	saveExample(t, queryRequest.Query(), "list formation constraints")

	var actualFormationConstraints []*graphql.FormationConstraint
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryRequest, &actualFormationConstraints))

	expectedConstraints := map[string]*graphql.FormationConstraint{
		"test_constraint": {
			ID:              constraint.ID,
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
			ID:              constraintSecond.ID,
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

	// Require there are at least len(expectedConstraints) as there may be more constraints created outside the test
	require.GreaterOrEqual(t, len(actualFormationConstraints), len(expectedConstraints))
	for _, f := range expectedConstraints {
		assert.Contains(t, actualFormationConstraints, f)
	}
}

func TestListFormationConstraintsForFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "formation-template-name"
	formationTemplateInput := fixtures.FixFormationTemplateInput(formationTemplateName)

	var formationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &formationTemplate)
	formationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)

	t.Logf("Assert there are no formation constraints for the formation template")
	constraintsForFormationTemplate := fixtures.ListFormationConstraintsForFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplate.ID)
	require.Len(t, constraintsForFormationTemplate, 0)

	secondFormationTemplateName := "second-formation-template-name"
	secondFormationTemplateInput := fixtures.FixFormationTemplateInput(secondFormationTemplateName)

	var secondFormationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &secondFormationTemplate)
	secondFormationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, secondFormationTemplateInput)

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

	// List all constraints and extract the global ones, so we can assert them later
	t.Log("List all formation constraints")
	queryRequest := fixtures.FixQueryFormationConstraintsRequest()
	var allFormationConstraints []*graphql.FormationConstraint
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryRequest, &allFormationConstraints))
	var globalFormationConstraints []graphql.FormationConstraint
	for _, fc := range allFormationConstraints {
		if fc.ConstraintScope == "GLOBAL" {
			globalFormationConstraints = append(globalFormationConstraints, *fc)
		}
	}

	// Assert no constraints attached
	t.Logf("Get formation template with name %q and id %q, and assert there are no constraints attached to it", formationTemplate.Name, formationTemplate.ID)
	ftOutput := graphql.FormationTemplateExt{}
	queryReq := fixtures.FixQueryFormationTemplateWithConstraintsRequest(formationTemplate.ID)
	saveExample(t, queryReq.Query(), "query formation template with constraints")
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryReq, &ftOutput))
	require.NotEmpty(t, ftOutput.ID)
	require.ElementsMatch(t, ftOutput.FormationConstraints, globalFormationConstraints) // only global constraints and no attached ones

	t.Logf("Get formation template with name %q and id %q, and assert there are no constraints attached to it", secondFormationTemplate.Name, secondFormationTemplate.ID)
	secondFormationTemplateOutput := fixtures.QueryFormationTemplateWithConstraints(t, ctx, certSecuredGraphQLClient, secondFormationTemplate.ID)
	require.ElementsMatch(t, secondFormationTemplateOutput.FormationConstraints, globalFormationConstraints) // only global constraints and no attached ones

	t.Logf("Attaching constraint to formation template")
	fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, formationTemplate.ID)

	// Assert the constraint is attached only to the first formation template
	t.Logf("Get formation template with name %q and id %q, and assert there is one constraint attached to it", formationTemplate.Name, formationTemplate.ID)
	formationTemplateOutput := fixtures.QueryFormationTemplateWithConstraints(t, ctx, certSecuredGraphQLClient, formationTemplate.ID)
	require.ElementsMatch(t, formationTemplateOutput.FormationConstraints, append(globalFormationConstraints, *constraint))

	t.Logf("Get formation template with name %q and id %q, and assert there are no constraints attached to it", secondFormationTemplate.Name, secondFormationTemplate.ID)
	secondFormationTemplateOutput = fixtures.QueryFormationTemplateWithConstraints(t, ctx, certSecuredGraphQLClient, secondFormationTemplate.ID)
	require.ElementsMatch(t, secondFormationTemplateOutput.FormationConstraints, globalFormationConstraints) // only global constraints and no attached ones

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

	// Assert the two constraints are attached only to the first formation template
	t.Logf("Get formation template with name %q and id %q, and assert there are are two constraints attached to it", formationTemplate.Name, formationTemplate.ID)
	formationTemplateOutput = fixtures.QueryFormationTemplateWithConstraints(t, ctx, certSecuredGraphQLClient, formationTemplate.ID)
	require.ElementsMatch(t, formationTemplateOutput.FormationConstraints, append([]graphql.FormationConstraint{*constraint, *constraintSecond}, globalFormationConstraints...))

	t.Logf("Get formation template with name %q and id %q, and assert there are no constraints attached to it", secondFormationTemplate.Name, secondFormationTemplate.ID)
	secondFormationTemplateOutput = fixtures.QueryFormationTemplateWithConstraints(t, ctx, certSecuredGraphQLClient, secondFormationTemplate.ID)
	require.ElementsMatch(t, secondFormationTemplateOutput.FormationConstraints, globalFormationConstraints) // only global constraints and no attached ones

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

	// Assert the two constraints are attached to the first formation template and one to the second formation template
	t.Logf("Get formation template with name %q and id %q, and assert there are are two constraints attached to it", formationTemplate.Name, formationTemplate.ID)
	formationTemplateOutput = fixtures.QueryFormationTemplateWithConstraints(t, ctx, certSecuredGraphQLClient, formationTemplate.ID)
	require.ElementsMatch(t, formationTemplateOutput.FormationConstraints, append([]graphql.FormationConstraint{*constraint, *constraintSecond}, globalFormationConstraints...))

	t.Logf("Get formation template with name %q and id %q, and assert there is one constraint attached to it", secondFormationTemplate.Name, secondFormationTemplate.ID)
	secondFormationTemplateOutput = fixtures.QueryFormationTemplateWithConstraints(t, ctx, certSecuredGraphQLClient, secondFormationTemplate.ID)
	require.ElementsMatch(t, secondFormationTemplateOutput.FormationConstraints, append(globalFormationConstraints, *constraintForOtherTemplate))

	queryRequest = fixtures.FixQueryFormationConstraintsForFormationTemplateRequest(formationTemplate.ID)
	saveExample(t, queryRequest.Query(), "list formation constraints for formation template")

	var actualFormationConstraints []*graphql.FormationConstraint
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryRequest, &actualFormationConstraints))

	expectedConstraints := map[string]*graphql.FormationConstraint{
		"test_constraint": {
			ID:              constraint.ID,
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
			ID:              constraintSecond.ID,
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

	// Require there are at least len(expectedConstraints) as there may be more constraints created outside the test
	require.GreaterOrEqual(t, len(actualFormationConstraints), len(expectedConstraints))
	for _, f := range expectedConstraints {
		assert.Contains(t, actualFormationConstraints, f)
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
