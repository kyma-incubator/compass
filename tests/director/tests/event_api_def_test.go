package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/director/tests/example"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventDefinitionInApplication(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	appName := "app-test-event-def-in-application"

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	event := fixtures.AddEventToApplication(t, ctx, certSecuredGraphQLClient, application.ID)

	queryEventForApplication := fixtures.FixEventForApplicationWithDefaultPaginationRequest(application.ID)

	eventDefPage := graphql.EventAPIDefinitionPageExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryEventForApplication, &eventDefPage)
	require.NoError(t, err)

	assert.Equal(t, 1, eventDefPage.TotalCount)
	assert.Equal(t, event.ID, eventDefPage.Data[0].ID)
	example.SaveExample(t, queryEventForApplication.Query(), "query event definition for application")
}

func TestAddEventDefinitionToApplication(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	appName := "app-test-event-def-in-application"

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	inputGQL, err := testctx.Tc.Graphqlizer.EventDefinitionInputToGQL(fixtures.FixEventDefinitionInput())
	require.NoError(t, err)

	eventAddRequest := fixtures.FixAddEventToApplicationRequest(application.ID, inputGQL)
	eventDef := graphql.EventAPIDefinitionExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, eventAddRequest, &eventDef)
	require.NoError(t, err)

	assert.NotNil(t, eventDef.ID)
	example.SaveExample(t, eventAddRequest.Query(), "add event definition to application")
}

func TestUpdateEventDefinitionToApplication(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-event-def-in-application"
	newName := "TestUpdateAPIDefinitionToApplication"
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	addedEvent := fixtures.AddEventToApplication(t, ctx, certSecuredGraphQLClient, application.ID)
	assert.NotNil(t, addedEvent.ID)

	inputGQL, err := testctx.Tc.Graphqlizer.EventDefinitionInputToGQL(fixtures.FixEventAPIDefinitionInputWithName(newName))
	require.NoError(t, err)

	eventUpdateRequest := fixtures.FixUpdateEventToApplicationRequest(addedEvent.ID, inputGQL)
	updatedEvent := graphql.EventAPIDefinitionExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, eventUpdateRequest, &updatedEvent)
	require.NoError(t, err)

	assert.Equal(t, newName, updatedEvent.Name)
	assert.NotEqual(t, addedEvent.Name, updatedEvent.Name)
	example.SaveExample(t, eventUpdateRequest.Query(), "update event definition to application")
}
