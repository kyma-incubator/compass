package application_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *model.Application
		Expected *graphql.Application
	}{
		{
			Name:     "All properties given",
			Input:    fixDetailedModelApplication(t, givenID(), givenTenant(), "Foo", "Lorem ipsum"),
			Expected: fixDetailedGQLApplication(t, givenID(), "Foo", "Lorem ipsum"),
		},
		{
			Name:  "Empty",
			Input: &model.Application{BaseEntity: &model.BaseEntity{}},
			Expected: &graphql.Application{
				Status: &graphql.ApplicationStatus{
					Condition: graphql.ApplicationStatusConditionInitial,
				},
				BaseEntity: &graphql.BaseEntity{},
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// when
			converter := application.NewConverter(nil, nil)
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	input := []*model.Application{
		fixModelApplication("foo", givenTenant(), "Foo", "Lorem ipsum"),
		fixModelApplication("bar", givenTenant(), "Bar", "Dolor sit amet"),
		{BaseEntity: &model.BaseEntity{}},
		nil,
	}
	expected := []*graphql.Application{
		fixGQLApplication("foo", "Foo", "Lorem ipsum"),
		fixGQLApplication("bar", "Bar", "Dolor sit amet"),
		{
			BaseEntity: &graphql.BaseEntity{},
			Status: &graphql.ApplicationStatus{
				Condition: graphql.ApplicationStatusConditionInitial,
			},
		},
	}

	// when
	converter := application.NewConverter(nil, nil)
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
}

func TestConverter_CreateInputFromGraphQL(t *testing.T) {
	allPropsInput := fixGQLApplicationRegisterInput("foo", "Lorem ipsum")
	allPropsExpected := fixModelApplicationRegisterInput("foo", "Lorem ipsum")

	// given
	testCases := []struct {
		Name               string
		Input              graphql.ApplicationRegisterInput
		Expected           model.ApplicationRegisterInput
		WebhookConverterFn func() *automock.WebhookConverter
		BundleConverterFn  func() *automock.BundleConverter
	}{
		{
			Name:     "All properties given",
			Input:    allPropsInput,
			Expected: allPropsExpected,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", allPropsInput.Webhooks).Return(allPropsExpected.Webhooks, nil)
				return conv
			},
			BundleConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("MultipleCreateInputFromGraphQL", allPropsInput.Bundles).Return(allPropsExpected.Bundles, nil)
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    graphql.ApplicationRegisterInput{},
			Expected: model.ApplicationRegisterInput{},
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", []*graphql.WebhookInput(nil)).Return(nil, nil)
				return conv
			},
			BundleConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("MultipleCreateInputFromGraphQL", []*graphql.BundleCreateInput(nil)).Return(nil, nil)
				return conv
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// when
			converter := application.NewConverter(
				testCase.WebhookConverterFn(),
				testCase.BundleConverterFn(),
			)
			res, err := converter.CreateInputFromGraphQL(context.TODO(), testCase.Input)

			// then
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_UpdateInputFromGraphQL_StatusCondition(t *testing.T) {
	testCases := []struct {
		Name           string
		CondtionGQL    graphql.ApplicationStatusCondition
		ConditionModel model.ApplicationStatusCondition
	}{
		{
			Name:           "When status condition is FAILED",
			CondtionGQL:    graphql.ApplicationStatusConditionFailed,
			ConditionModel: model.ApplicationStatusConditionFailed,
		},
		{
			Name:           "When status condition is CONNECTED",
			CondtionGQL:    graphql.ApplicationStatusConditionConnected,
			ConditionModel: model.ApplicationStatusConditionConnected,
		},
		{
			Name:           "When status condition is INITIAL",
			CondtionGQL:    graphql.ApplicationStatusConditionInitial,
			ConditionModel: model.ApplicationStatusConditionInitial,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			gqlApp := graphql.ApplicationUpdateInput{StatusCondition: &testCase.CondtionGQL}

			converter := application.NewConverter(nil, nil)
			modelApp := converter.UpdateInputFromGraphQL(gqlApp)

			require.Equal(t, &testCase.ConditionModel, modelApp.StatusCondition)
		})
	}
}

func TestConverter_UpdateInputFromGraphQL(t *testing.T) {
	allPropsInput := fixGQLApplicationUpdateInput("foo", "Lorem ipsum", testURL, graphql.ApplicationStatusConditionConnected)
	allPropsExpected := fixModelApplicationUpdateInput("foo", "Lorem ipsum", testURL, model.ApplicationStatusConditionConnected)

	// given
	testCases := []struct {
		Name     string
		Input    graphql.ApplicationUpdateInput
		Expected model.ApplicationUpdateInput
	}{
		{
			Name:     "All properties given",
			Input:    allPropsInput,
			Expected: allPropsExpected,
		},
		{
			Name:     "Empty",
			Input:    graphql.ApplicationUpdateInput{},
			Expected: model.ApplicationUpdateInput{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// when
			converter := application.NewConverter(nil, nil)
			res := converter.UpdateInputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	conv := application.NewConverter(nil, nil)

	t.Run("All properties given", func(t *testing.T) {
		// GIVEN
		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "app-name", "app-description")

		// WHEN
		appEntity, err := conv.ToEntity(appModel)

		// THEN
		assert.NoError(t, err)
		assertApplicationDefinition(t, appModel, appEntity)
	})

	t.Run("Nil", func(t *testing.T) {
		// WHEN
		appEntity, err := conv.ToEntity(nil)

		// THEN
		assert.NoError(t, err)
		assert.Nil(t, appEntity)
	})

	t.Run("Empty", func(t *testing.T) {
		// GIVEN
		appModel := &model.Application{}

		// WHEN
		appEntity, err := conv.ToEntity(appModel)

		// THEN
		if err != nil {
			assert.Contains(t, err.Error(), "invalid input model")
		} else {
			assertApplicationDefinition(t, appModel, appEntity)
		}
	})
}

func TestConverter_FromEntity(t *testing.T) {
	conv := application.NewConverter(nil, nil)

	t.Run("All properties given", func(t *testing.T) {
		// GIVEN
		appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "app-name", "app-description")

		// WHEN
		appModel := conv.FromEntity(appEntity)

		// THEN
		assertApplicationDefinition(t, appModel, appEntity)
	})

	t.Run("Nil", func(t *testing.T) {
		// WHEN
		appModel := conv.FromEntity(nil)

		// THEN
		assert.Nil(t, appModel)
	})

	t.Run("Empty", func(t *testing.T) {
		// GIVEN
		appEntity := &application.Entity{BaseEntity: &repo.BaseEntity{}}

		// WHEN
		appModel := conv.FromEntity(appEntity)

		// THEN
		assertApplicationDefinition(t, appModel, appEntity)
	})
}

func TestConverter_CreateInputGQLJSONConversion(t *testing.T) {
	// GIVEN
	conv := application.NewConverter(nil, nil)

	t.Run("Successful two-way conversion", func(t *testing.T) {
		inputGQL := fixGQLApplicationRegisterInput("name", "description")
		inputGQL.Labels = &graphql.Labels{"test": "test"}

		// WHEN
		// GQL -> JSON
		json, err := conv.CreateInputGQLToJSON(&inputGQL)
		require.NoError(t, err)

		// JSON -> GQL
		outputGQL, err := conv.CreateInputJSONToGQL(json)
		require.NoError(t, err)

		// THEN
		require.Equal(t, inputGQL, outputGQL)
	})

	t.Run("Error while JSON to GQL conversion", func(t *testing.T) {
		// WHEN
		expectedErr := "invalid character 'a' looking for beginning of value"
		_, err := conv.CreateInputJSONToGQL("ad[sd")

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), expectedErr)
	})
}

func TestConverter_ConvertToModel(t *testing.T) {
	conv := application.NewConverter(nil, nil)

	t.Run("Successful full model", func(t *testing.T) {
		tenantID := uuid.New().String()
		appGraphql := fixGQLApplication(uuid.New().String(), "app", "desc")

		//WHEN
		appModel := conv.GraphQLToModel(appGraphql, tenantID)
		outputGraphql := conv.ToGraphQL(appModel)

		//THEN
		assert.Equal(t, appGraphql, outputGraphql)
	})

	t.Run("Success empty model", func(t *testing.T) {
		//GIVEN
		appGraphql := &graphql.Application{BaseEntity: &graphql.BaseEntity{}}

		//WHEN
		appModel := conv.GraphQLToModel(appGraphql, uuid.New().String())
		outputGraphql := conv.ToGraphQL(appModel)

		//THEN
		appGraphql.Status = &graphql.ApplicationStatus{Condition: graphql.ApplicationStatusConditionInitial}
		assert.Equal(t, appGraphql, outputGraphql)
	})

	t.Run("Nil model", func(t *testing.T) {
		//WHEN
		output := conv.GraphQLToModel(nil, uuid.New().String())
		//THEN
		require.Nil(t, output)
	})
}

func assertApplicationDefinition(t *testing.T, appModel *model.Application, entity *application.Entity) {
	assert.Equal(t, appModel.ID, entity.ID)
	assert.Equal(t, appModel.Tenant, entity.TenantID)
	assert.Equal(t, appModel.Name, entity.Name)

	if appModel.Status != nil {
		assert.Equal(t, appModel.Status.Condition, model.ApplicationStatusCondition(entity.StatusCondition))
		assert.Equal(t, appModel.Status.Timestamp, entity.StatusTimestamp)
	} else {
		assert.Equal(t, string(model.ApplicationStatusConditionInitial), string(entity.StatusCondition))
	}

	testdb.AssertSqlNullStringEqualTo(t, entity.Description, appModel.Description)
	testdb.AssertSqlNullStringEqualTo(t, entity.HealthCheckURL, appModel.HealthCheckURL)
	testdb.AssertSqlNullStringEqualTo(t, entity.ProviderName, appModel.ProviderName)
}

func givenID() string {
	return "bd0646fa-3c30-4255-84f8-182f57742aa1"
}

func givenTenant() string {
	return "8f237125-50be-4bb4-96ce-389e2b931f46"
}
