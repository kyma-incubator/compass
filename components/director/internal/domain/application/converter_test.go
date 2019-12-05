package application_test

import (
	"testing"

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
			Input: &model.Application{},
			Expected: &graphql.Application{
				Status: &graphql.ApplicationStatus{
					Condition: graphql.ApplicationStatusConditionInitial,
				},
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
			converter := application.NewConverter(nil, nil, nil, nil)
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
		{},
		nil,
	}
	expected := []*graphql.Application{
		fixGQLApplication("foo", "Foo", "Lorem ipsum"),
		fixGQLApplication("bar", "Bar", "Dolor sit amet"),
		{
			Status: &graphql.ApplicationStatus{
				Condition: graphql.ApplicationStatusConditionInitial,
			},
		},
	}

	// when
	converter := application.NewConverter(nil, nil, nil, nil)
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
}

func TestConverter_CreateInputFromGraphQL(t *testing.T) {
	allPropsInput := fixGQLApplicationCreateInput("foo", "Lorem ipsum")
	allPropsExpected := fixModelApplicationCreateInput("foo", "Lorem ipsum")

	// given
	testCases := []struct {
		Name                string
		Input               graphql.ApplicationCreateInput
		Expected            model.ApplicationCreateInput
		WebhookConverterFn  func() *automock.WebhookConverter
		DocumentConverterFn func() *automock.DocumentConverter
		APIConverterFn      func() *automock.APIConverter
		EventAPIConverterFn func() *automock.EventAPIConverter
	}{
		{
			Name:     "All properties given",
			Input:    allPropsInput,
			Expected: allPropsExpected,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", allPropsInput.Webhooks).Return(allPropsExpected.Webhooks)
				return conv
			},
			APIConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleInputFromGraphQL", allPropsInput.Apis).Return(allPropsExpected.Apis)
				return conv
			},
			EventAPIConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("MultipleInputFromGraphQL", allPropsInput.EventAPIs).Return(allPropsExpected.EventAPIs)
				return conv
			},
			DocumentConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("MultipleInputFromGraphQL", allPropsInput.Documents).Return(allPropsExpected.Documents)
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    graphql.ApplicationCreateInput{},
			Expected: model.ApplicationCreateInput{},
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", []*graphql.WebhookInput(nil)).Return(nil)
				return conv
			},
			APIConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleInputFromGraphQL", []*graphql.APIDefinitionInput(nil)).Return(nil)
				return conv
			},
			EventAPIConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("MultipleInputFromGraphQL", []*graphql.EventAPIDefinitionInput(nil)).Return(nil)
				return conv
			},
			DocumentConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("MultipleInputFromGraphQL", []*graphql.DocumentInput(nil)).Return(nil)
				return conv
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// when
			converter := application.NewConverter(
				testCase.WebhookConverterFn(),
				testCase.APIConverterFn(),
				testCase.EventAPIConverterFn(),
				testCase.DocumentConverterFn(),
			)
			res := converter.CreateInputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_UpdateInputFromGraphQL(t *testing.T) {
	allPropsInput := fixGQLApplicationUpdateInput("foo", "Lorem ipsum", testURL)
	allPropsExpected := fixModelApplicationUpdateInput("foo", "Lorem ipsum", testURL)

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
			converter := application.NewConverter(nil, nil, nil, nil)
			res := converter.UpdateInputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	conv := application.NewConverter(nil, nil, nil, nil)

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
	conv := application.NewConverter(nil, nil, nil, nil)

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
		appEntity := &application.Entity{}

		// WHEN
		appModel := conv.FromEntity(appEntity)

		// THEN
		assertApplicationDefinition(t, appModel, appEntity)
	})
}

func TestConverter_CreateInputGQLJSONConversion(t *testing.T) {
	// GIVEN
	conv := application.NewConverter(nil, nil, nil, nil)

	t.Run("Successful two-way conversion", func(t *testing.T) {
		inputGQL := fixGQLApplicationCreateInput("name", "description")
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

	t.Run("Successful empty JSON to GQL conversion", func(t *testing.T) {
		// WHEN
		outputGQL, err := conv.CreateInputJSONToGQL("")

		// THEN
		require.NoError(t, err)
		require.Empty(t, outputGQL)
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
		assert.Equal(t, string(model.ApplicationStatusConditionUnknown), string(entity.StatusCondition))
	}

	testdb.AssertSqlNullString(t, entity.Description, appModel.Description)
	testdb.AssertSqlNullString(t, entity.HealthCheckURL, appModel.HealthCheckURL)
}

func givenID() string {
	return "bd0646fa-3c30-4255-84f8-182f57742aa1"
}

func givenTenant() string {
	return "8f237125-50be-4bb4-96ce-389e2b931f46"
}
