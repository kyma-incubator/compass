package formationtemplate_test

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_FromRegisterInputGraphQL(t *testing.T) {
	modelWebhookInputs := fixModelWebhookInput()
	GQLWebhooksInputs := fixGQLWebhookInput()

	testCases := []struct {
		Name               string
		Input              *graphql.FormationTemplateRegisterInput
		WebhookConverterFn func() *automock.WebhookConverter
		Expected           *model.FormationTemplateRegisterInput
		ExpectedErr        bool
	}{
		{
			Name:  "Success",
			Input: &formationTemplateRegisterInputGraphQL,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", GQLWebhooksInputs).Return(modelWebhookInputs, nil)
				return conv
			},
			Expected: &formationTemplateRegisterInputModel,
		},
		{
			Name:  "Success for app only templates",
			Input: &formationTemplateGraphQLInputAppOnly,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", GQLWebhooksInputs).Return(modelWebhookInputs, nil)
				return conv
			},
			Expected: &formationTemplateModelInputAppOnly,
		},
		{
			Name:  "Success with reset",
			Input: &formationTemplateWithResetGraphQLInput,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", GQLWebhooksInputs).Return(modelWebhookInputs, nil)
				return conv
			},
			Expected: &formationTemplateModelWithResetInput,
		},
		{
			Name:  "Error when converting webhooks",
			Input: &formationTemplateRegisterInputGraphQL,
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
			result, err := converter.FromRegisterInputGraphQL(testCase.Input)
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

func TestConverter_FromUpdateInputGraphQL(t *testing.T) {
	testCases := []struct {
		Name        string
		Input       *graphql.FormationTemplateUpdateInput
		Expected    *model.FormationTemplateUpdateInput
		ExpectedErr bool
	}{
		{
			Name:     "Success",
			Input:    &formationTemplateUpdateInputGraphQL,
			Expected: &formationTemplateUpdateInputModel,
		},
		{
			Name:        "Empty",
			Input:       nil,
			Expected:    nil,
			ExpectedErr: false,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			converter := formationtemplate.NewConverter(nil)
			// WHEN
			result, err := converter.FromUpdateInputGraphQL(testCase.Input)
			if testCase.ExpectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, result, testCase.Expected)
		})
	}
}

func TestConverter_FromModelInputToModel(t *testing.T) {
	ftModel := fixFormationTemplateModel(time.Time{}, nil)

	testCases := []struct {
		Name     string
		Input    *model.FormationTemplateRegisterInput
		Expected *model.FormationTemplate
	}{{
		Name:     "Success",
		Input:    &formationTemplateRegisterInputModel,
		Expected: ftModel,
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
			result := converter.FromModelRegisterInputToModel(testCase.Input, testFormationTemplateID, testTenantID)

			if testCase.Expected != nil {
				testCase.Expected.Webhooks[0].ID = result.Webhooks[0].ID // id is generated ad-hoc and can't be mocked
			}

			assert.Equal(t, result, testCase.Expected)
		})
	}
}

func TestConverter_FromModelUpdateInputToModel(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *model.FormationTemplateUpdateInput
		Expected *model.FormationTemplate
	}{{
		Name:     "Success",
		Input:    &formationTemplateUpdateInputModel,
		Expected: &formationTemplateModelWithoutWebhooks,
	}, {
		Name:     "Empty",
		Input:    nil,
		Expected: nil,
	},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := formationtemplate.NewConverter(nil)

			result := converter.FromModelUpdateInputToModel(testCase.Input, testFormationTemplateID, testTenantID)

			require.Equal(t, result, testCase.Expected)
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
}

func TestConverter_ToEntity(t *testing.T) {
	ftModel := fixFormationTemplateModel(time.Time{}, nil)
	ftEntity := fixFormationTemplateEntity(time.Time{}, nil)

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter(nil)
		// WHEN
		result, err := converter.ToEntity(ftModel)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, result, ftEntity)
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
	formationTemplateModelWithoutWebhooks := fixFormationTemplateModel(time.Time{}, nil)
	formationTemplateModelWithoutWebhooks.Webhooks = nil

	ftEntity := fixFormationTemplateEntity(time.Time{}, nil)

	testCases := []struct {
		Name     string
		Input    *formationtemplate.Entity
		Expected *model.FormationTemplate
	}{{
		Name:     "Success",
		Input:    ftEntity,
		Expected: formationTemplateModelWithoutWebhooks,
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
