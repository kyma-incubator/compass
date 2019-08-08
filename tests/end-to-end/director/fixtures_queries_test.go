package director

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func createRandomApplication(t *testing.T, ctx context.Context, name string) graphql.Application {
	in := generateSampleApplicationInputWithName("first", name)
	return createApplication(t, ctx, in, defaultTenant)
}

func createApplicationFromInput(t *testing.T, ctx context.Context, in graphql.ApplicationInput) graphql.Application {
	return createApplication(t, ctx, in, defaultTenant)
}

func createApplicationWithinTenantFromInput(t *testing.T, ctx context.Context, in graphql.ApplicationInput, tenant string) graphql.Application {
	return createApplication(t, ctx, in, tenant)
}

func createApplication(t *testing.T, ctx context.Context, in graphql.ApplicationInput, tenant string) graphql.Application {
	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)

	createReq := fixCreateApplicationRequest(appInputGQL)
	app := graphql.Application{}

	require.NoError(t, tc.RunQuery(ctx, createReq, &app))
	require.NotEmpty(t, app.ID)
	return app
}

func deleteLabelDefinition(t *testing.T, ctx context.Context, labelDefinitionKey string, deleteRelatedResources bool) {
	deleteRequest := fixDeleteLabelDefinition(labelDefinitionKey, deleteRelatedResources)

	require.NoError(t, tc.RunQuery(ctx, deleteRequest, nil))
}

func deleteApplicationLabel(t *testing.T, ctx context.Context, applicationID, labelKey string) {
	deleteRequest := fixDeleteApplicationLabel(applicationID, labelKey)

	require.NoError(t, tc.RunQuery(ctx, deleteRequest, nil))
}

func setApplicationLabel(t *testing.T, ctx context.Context, applicationID string, labelKey string, labelValue interface{}) graphql.Label {
	setLabelRequest := fixSetApplicationLabelRequest(applicationID, labelKey, labelValue)
	label := graphql.Label{}

	err := tc.RunQuery(ctx, setLabelRequest, &label)
	require.NoError(t, err)

	return label
}
