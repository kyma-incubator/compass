package fixtures

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/json"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func CreateLabelDefinitionWithinTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, key string, schema interface{}, tenantID string) *graphql.LabelDefinition {
	output, err := CreateLabelDefinitionWithinTenantError(t, ctx, gqlClient, key, schema, tenantID)
	require.NoError(t, err)

	return output
}

func CreateLabelDefinitionWithinTenantError(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, key string, schema interface{}, tenantID string) (*graphql.LabelDefinition, error) {
	input := graphql.LabelDefinitionInput{
		Key:    key,
		Schema: json.MarshalJSONSchema(t, schema),
	}

	in, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(input)
	if err != nil {
		return nil, err
	}

	createRequest := FixCreateLabelDefinitionRequest(in)

	output := graphql.LabelDefinition{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, createRequest, &output)
	return &output, err
}

func UpsertScenariosLabelDefinitionWithinTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantID string, scenarios []string) *graphql.LabelDefinition {
	jsonSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"enum": scenarios,
			"type": "string",
		},
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
	}

	output, err := CreateLabelDefinitionWithinTenantError(t, ctx, gqlClient, "scenarios", jsonSchema, tenantID)
	if err != nil {
		if strings.Contains(err.Error(), "Object is not unique") {
			return UpdateLabelDefinitionWithinTenant(t, ctx, gqlClient, "scenarios", jsonSchema, tenantID)
		}
	}
	require.NoError(t, err)
	return output
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

func CreateScenariosLabelDefinitionWithinTenantError(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantID string, scenarios []string) (*graphql.LabelDefinition, error) {
	jsonSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"enum": scenarios,
			"type": "string",
		},
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
	}

	return CreateLabelDefinitionWithinTenantError(t, ctx, gqlClient, "scenarios", jsonSchema, tenantID)
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
	// Listing scenarios for LabelDefinition creates a default record if there are no scenarios. This invocation is needed so that we can later update the scenarios instead of creating a new LabelDefinition
	_, err := ListLabelDefinitionByKeyWithinTenant(ctx, gqlClient, "scenarios", tenantID)
	require.NoError(t, err)

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

func ListLabelDefinitionsWithinTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantID string) ([]*graphql.LabelDefinition, error) {
	labelDefinitionsRequest := FixLabelDefinitionsRequest()

	var labelDefinitions []*graphql.LabelDefinition

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, labelDefinitionsRequest, &labelDefinitions)
	return labelDefinitions, err
}

func ListLabelDefinitionByKeyWithinTenant(ctx context.Context, gqlClient *gcli.Client, key, tenantID string) (*graphql.LabelDefinition, error) {
	labelDefinitionRequest := FixLabelDefinitionRequest(key)

	var labelDefinition *graphql.LabelDefinition

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, labelDefinitionRequest, &labelDefinition)
	return labelDefinition, err
}
