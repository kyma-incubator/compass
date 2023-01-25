package formationtemplate_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testErr = errors.New("test-error")

func TestConverter_FromInputGraphQL(t *testing.T) {
	modelWebhookInputs := fixModelWebhookInput()
	GQLWebhooksInputs := fixGQLWebhookInput()

	testCases := []struct {
		Name               string
		Input              *graphql.FormationTemplateInput
		WebhookConverterFn func() *automock.WebhookConverter
		Expected           *model.FormationTemplateInput
		ExpectedErr        bool
	}{
		{
			Name:  "Success",
			Input: &formationTemplateGraphQLInput,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", GQLWebhooksInputs).Return(modelWebhookInputs, nil)
				return conv
			},
			Expected: &formationTemplateModelInput,
		},
		{
			Name:  "Error when converting webhooks",
			Input: &formationTemplateGraphQLInput,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", GQLWebhooksInputs).Return(nil, testErr)
				return conv
			},
			Expected:    nil,
			ExpectedErr: true,
		},
		{
			Name:  "Empty",
			Input: nil,
			WebhookConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			Expected:    nil,
			ExpectedErr: false,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			whConverter := testCase.WebhookConverterFn()
			converter := formationtemplate.NewConverter(whConverter)
			// WHEN
			result, err := converter.FromInputGraphQL(testCase.Input)
			if testCase.ExpectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, result, testCase.Expected)
			whConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_FromModelInputToModel(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *model.FormationTemplateInput
		Expected *model.FormationTemplate
	}{{
		Name:     "Success",
		Input:    &formationTemplateModelInput,
		Expected: &formationTemplateModel,
	}, {
		Name:     "Empty",
		Input:    nil,
		Expected: nil,
	},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			converter := formationtemplate.NewConverter(nil)
			// WHEN
			result := converter.FromModelInputToModel(testCase.Input, testID, testTenantID)

			if testCase.Expected != nil {
				testCase.Expected.Webhooks[0].ID = result.Webhooks[0].ID // id is generated ad-hoc and can't be mocked
			}

			assert.Equal(t, result, testCase.Expected)
		})
	}
}

func TestConverter_ToGraphQL(t *testing.T) {
	GQLWebhooks := []*graphql.Webhook{fixFormationTemplateGQLWebhook()}

	testCases := []struct {
		Name               string
		Input              *model.FormationTemplate
		WebhookConverterFn func() *automock.WebhookConverter
		Expected           *graphql.FormationTemplate
		ExpectedErr        bool
	}{
		{
			Name:  "Success",
			Input: &formationTemplateModel,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", formationTemplateModel.Webhooks).Return(GQLWebhooks, nil)
				return conv
			},
			Expected: &graphQLFormationTemplate,
		},
		{
			Name:  "Error when converting webhook",
			Input: &formationTemplateModel,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", formationTemplateModel.Webhooks).Return(nil, testErr)
				return conv
			},
			Expected:    nil,
			ExpectedErr: true,
		},
		{
			Name:  "Empty",
			Input: nil,
			WebhookConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			Expected: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			whConverter := testCase.WebhookConverterFn()
			converter := formationtemplate.NewConverter(whConverter)
			// WHEN
			result, err := converter.ToGraphQL(testCase.Input)
			if testCase.ExpectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, result, testCase.Expected)
			whConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	GQLWebhooks := []*graphql.Webhook{fixFormationTemplateGQLWebhook()}
	testCases := []struct {
		Name               string
		Input              []*model.FormationTemplate
		WebhookConverterFn func() *automock.WebhookConverter
		Expected           []*graphql.FormationTemplate
		ExpectedErr        bool
	}{
		{
			Name:  "Success",
			Input: formationTemplateModelPage.Data,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", formationTemplateModel.Webhooks).Return(GQLWebhooks, nil)
				return conv
			},
			Expected: graphQLFormationTemplatePage.Data,
		},
		{
			Name:  "Error when converting webhook",
			Input: formationTemplateModelPage.Data,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", formationTemplateModel.Webhooks).Return(nil, testErr)
				return conv
			},
			Expected:    nil,
			ExpectedErr: true,
		},
		{
			Name:  "Empty",
			Input: nil,
			WebhookConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			Expected: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			whConverter := testCase.WebhookConverterFn()
			converter := formationtemplate.NewConverter(whConverter)
			// WHEN
			result, err := converter.MultipleToGraphQL(testCase.Input)
			if testCase.ExpectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.ElementsMatch(t, result, testCase.Expected)
			whConverter.AssertExpectations(t)
		})
	}
	//t.Run("Success", func(t *testing.T) {
	//	// GIVEN
	//	converter := formationtemplate.NewConverter(nil)
	//	// WHEN
	//	result := converter.MultipleToGraphQL(formationTemplateModelPage.Data)
	//
	//	// THEN
	//	assert.ElementsMatch(t, result, graphQLFormationTemplatePage.Data)
	//})
	//t.Run("Returns nil when given empty model", func(t *testing.T) {
	//	// GIVEN
	//	converter := formationtemplate.NewConverter()
	//	// WHEN
	//	result := converter.MultipleToGraphQL(nil)
	//
	//	assert.Nil(t, result)
	//})
}

func TestConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter(nil)
		// WHEN
		result, err := converter.ToEntity(&formationTemplateModel)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, result, &formationTemplateEntity)
	})
	t.Run("Returns nil when given empty model", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter(nil)
		// WHEN
		result, err := converter.ToEntity(nil)

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestConverter_FromEntity(t *testing.T) {
	formationTemplateModelWithoutWebhooks := formationTemplateModel
	formationTemplateModelWithoutWebhooks.Webhooks = nil
	testCases := []struct {
		Name     string
		Input    *formationtemplate.Entity
		Expected *model.FormationTemplate
	}{{
		Name:     "Success",
		Input:    &formationTemplateEntity,
		Expected: &formationTemplateModelWithoutWebhooks,
	}, {
		Name:     "Empty",
		Input:    nil,
		Expected: nil,
	},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			converter := formationtemplate.NewConverter(nil)
			// WHEN
			result, err := converter.FromEntity(testCase.Input)

			assert.NoError(t, err)
			assert.Equal(t, result, testCase.Expected)
		})
	}
}
