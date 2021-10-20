package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFormation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	labelKey := "scenarios"
	defaultValue := conf.DefaultScenario
	additionalValue := "ADDITIONAL"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application")
	app, err := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app", tenantId)
	defer func() {
		fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, tenantId, app.ID, conf.DefaultScenarioEnabled)
		fixtures.CleanupApplication(t, ctx, dexGraphQLClient, tenantId, &app)
		fixtures.CleanupFormation(t, ctx, dexGraphQLClient, tenantId, additionalValue)
	}()
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	t.Logf("Update Label Definition scenarios enum with additional value %s", additionalValue)

	var formation graphql.Formation
	createReq := fixtures.FixCreateFormationRequest(additionalValue)

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createReq, &formation)
	require.NoError(t, err)

	formations := []string{additionalValue}
	if conf.DefaultScenarioEnabled {
		formations = []string{defaultValue, additionalValue}
	}
	var labelValue interface{} = formations

	t.Logf("Set scenario label value %s on application", additionalValue)
	fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, app.ID, labelKey, labelValue)

	t.Log("Check if new scenario label value was set correctly")
	appRequest := fixtures.FixGetApplicationRequest(app.ID)
	app = graphql.ApplicationExt{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, appRequest, &app)
	require.NoError(t, err)

	scenariosLabel, ok := app.Labels[labelKey].([]interface{})
	require.True(t, ok)

	var actualScenariosEnum []string
	for _, v := range scenariosLabel {
		actualScenariosEnum = append(actualScenariosEnum, v.(string))
	}
	assert.Equal(t, formations, actualScenariosEnum)
}

//
//func TestDeleteFormation(t *testing.T) {
//	// GIVEN
//	ctx := context.Background()
//
//	tenantId := tenant.TestTenants.GetDefaultTenantID()
//
//	t.Log("Create application")
//	app, err := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app", tenantId)
//	defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, tenantId, &app)
//	defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, tenantId, app.ID, conf.DefaultScenarioEnabled)
//	require.NoError(t, err)
//	require.NotEmpty(t, app.ID)
//
//	labelKey := "scenarios"
//	defaultValue := conf.DefaultScenario
//	additionalValue := "ADDITIONAL"
//
//	t.Logf("Update Label Definition scenarios enum with additional value %s", additionalValue)
//
//	var formation graphql.Formation
//	createReq := fixtures.FixCreateFormationRequest(additionalValue)
//
//	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createReq, &formation)
//	require.NoError(t, err)
//
//	formations := []string{additionalValue}
//	if conf.DefaultScenarioEnabled {
//		formations = []string{defaultValue, additionalValue}
//	}
//	var labelValue interface{} = formations
//
//	t.Logf("Set scenario label value %s on application", additionalValue)
//	fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, app.ID, labelKey, labelValue)
//
//	t.Log("Check if new scenario label value was set correctly")
//	appRequest := fixtures.FixGetApplicationRequest(app.ID)
//	app = graphql.ApplicationExt{}
//
//	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, appRequest, &app)
//	require.NoError(t, err)
//
//	scenariosLabel, ok := app.Labels[labelKey].([]interface{})
//	require.True(t, ok)
//
//	var actualScenariosEnum []string
//	for _, v := range scenariosLabel {
//		actualScenariosEnum = append(actualScenariosEnum, v.(string))
//	}
//	assert.Equal(t, formations, actualScenariosEnum)
//}
