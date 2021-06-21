package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/json"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func CreateLabelDefinitionWithinTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, key string, schema interface{}, tenantID string) *graphql.LabelDefinition {
	input := graphql.LabelDefinitionInput{
		Key:    key,
		Schema: json.MarshalJSONSchema(t, schema),
	}

	in, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(input)
	if err != nil {
		return nil
	}

	createRequest := FixCreateLabelDefinitionRequest(in)

	output := graphql.LabelDefinition{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, createRequest, &output)
	require.NoError(t, err)

	return &output
}

func CreateScenariosLabelDefinitionWithinTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantID string, scenarios []string) *graphql.LabelDefinition {
	jsonSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"enum": scenarios,
			"type": "string",
		},
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
	}

	return CreateLabelDefinitionWithinTenant(t, ctx, gqlClient, "scenarios", jsonSchema, tenantID)
}

func UpdateLabelDefinitionWithinTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, key string, schema interface{}, tenantID string) *graphql.LabelDefinition {
	input := graphql.LabelDefinitionInput{
		Key:    key,
		Schema: json.MarshalJSONSchema(t, schema),
	}

	in, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(input)
	if err != nil {
		return nil
	}

	updateRequest := FixUpdateLabelDefinitionRequest(in)

	output := graphql.LabelDefinition{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, updateRequest, &output)
	require.NoError(t, err)

	return &output
}

func UpdateScenariosLabelDefinitionWithinTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantID string, scenarios []string) *graphql.LabelDefinition {
	jsonSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"enum": scenarios,
			"type": "string",
		},
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
	}

	return UpdateLabelDefinitionWithinTenant(t, ctx, gqlClient, "scenarios", jsonSchema, tenantID)
}

func DeleteLabelDefinition(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, labelDefinitionKey string, deleteRelatedResources bool, tenantID string) {
	deleteRequest := FixDeleteLabelDefinitionRequest(labelDefinitionKey, deleteRelatedResources)

	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, deleteRequest, nil))
}

func ListLabelDefinitionsWithinTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantID string) ([]*graphql.LabelDefinition, error) {
	labelDefinitionsRequest := FixLabelDefinitionsRequest()

	var labelDefinitions []*graphql.LabelDefinition

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, labelDefinitionsRequest, &labelDefinitions)
	return labelDefinitions, err
}
