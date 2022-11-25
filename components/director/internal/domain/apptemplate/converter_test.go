package apptemplate_test

import (
	"database/sql"
	"testing"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mockedError = errors.New("test-error")

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	appConv := &automock.AppConverter{}
	modelWebhooks := fixModelApplicationWebhooks(testWebhookID, testID)
	GQLWebhooks := fixGQLApplicationWebhooks(testWebhookID, testID)

	testCases := []struct {
		Name               string
		Input              *model.ApplicationTemplate
		Expected           *graphql.ApplicationTemplate
		ExpectedError      bool
		WebhookConverterFn func() *automock.WebhookConverter
	}{
		{
			Name:          "All properties given",
			Input:         fixModelApplicationTemplate(testID, testName, modelWebhooks),
			Expected:      fixGQLAppTemplate(testID, testName, GQLWebhooks),
			ExpectedError: false,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks).Return(GQLWebhooks, nil)
				return conv
			},
		},
		{
			Name: "Error when graphqlising Application Create Input",
			Input: &model.ApplicationTemplate{
				ID:                   testID,
				Name:                 testName,
				ApplicationInputJSON: "{abc",
			},
			Expected:      nil,
			ExpectedError: true,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks).Return(GQLWebhooks, nil)
				return conv
			},
		},
		{
			Name:          "Error when converting Webhooks",
			Input:         fixModelApplicationTemplate(testID, testName, modelWebhooks),
			Expected:      nil,
			ExpectedError: true,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks).Return(nil, mockedError)
				return conv
			},
		},
		{
			Name:          "Empty",
			Input:         &model.ApplicationTemplate{},
			Expected:      nil,
			ExpectedError: true,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks).Return(GQLWebhooks, nil)
				return conv
			},
		},
		{
			Name:          "Nil",
			Input:         nil,
			Expected:      nil,
			ExpectedError: false,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks).Return(GQLWebhooks, nil)
				return conv
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			webhookConverter := testCase.WebhookConverterFn()
			converter := apptemplate.NewConverter(appConv, webhookConverter)

			res, err := converter.ToGraphQL(testCase.Input)
			if testCase.ExpectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// THEN
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// GIVEN
	modelWebhooks := [][]*model.Webhook{
		fixModelApplicationTemplateWebhooks("webhook-id-1", "id1"),
		fixModelApplicationTemplateWebhooks("webhook-id-2", "id2"),
	}
	GQLWebhooks := [][]*graphql.Webhook{
		fixGQLApplicationTemplateWebhooks("webhook-id-1", "id1"),
		fixGQLApplicationTemplateWebhooks("webhook-id-2", "id2"),
	}

	testCases := []struct {
		Name               string
		Input              []*model.ApplicationTemplate
		Expected           []*graphql.ApplicationTemplate
		ExpectedError      bool
		WebhookConverterFn func() *automock.WebhookConverter
	}{
		{
			Name: "All properties given",
			Input: []*model.ApplicationTemplate{
				fixModelApplicationTemplate("id1", "name1", modelWebhooks[0]),
				fixModelApplicationTemplate("id2", "name2", modelWebhooks[1]),
				nil,
			},
			Expected: []*graphql.ApplicationTemplate{
				fixGQLAppTemplate("id1", "name1", GQLWebhooks[0]),
				fixGQLAppTemplate("id2", "name2", GQLWebhooks[1]),
			},
			ExpectedError: false,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks[0]).Return(GQLWebhooks[0], nil)
				conv.On("MultipleToGraphQL", modelWebhooks[1]).Return(GQLWebhooks[1], nil)
				return conv
			},
		},
		{
			Name: "Error when application input is empty",
			Input: []*model.ApplicationTemplate{
				fixModelApplicationTemplate("id1", "name1", modelWebhooks[0]),
				{
					ID:                   testID,
					Name:                 testName,
					ApplicationInputJSON: "",
				},
			},
			Expected:      nil,
			ExpectedError: true,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks[0]).Return(GQLWebhooks[0], nil)
				return conv
			},
		},
		{
			Name: "Error when converting application template",
			Input: []*model.ApplicationTemplate{
				fixModelApplicationTemplate("id1", "name1", modelWebhooks[0]),
				{
					ID:                   testID,
					Name:                 testName,
					ApplicationInputJSON: "{abc",
				},
			},
			Expected:      nil,
			ExpectedError: true,
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks[0]).Return(GQLWebhooks[0], nil)
				return conv
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			webhookConverter := testCase.WebhookConverterFn()
			converter := apptemplate.NewConverter(nil, webhookConverter)
			res, err := converter.MultipleToGraphQL(testCase.Input)
			if testCase.ExpectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// THEN
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// GIVEN
	appTemplateInputGQL := fixGQLAppTemplateInput(testName)
	appTemplateInputModel := fixModelAppTemplateInput(testName, "{\"name\":\"foo\",\"description\":\"Lorem ipsum\"}")

	testCases := []struct {
		Name               string
		AppConverterFn     func() *automock.AppConverter
		WebhookConverterFn func() *automock.WebhookConverter
		Input              graphql.ApplicationTemplateInput
		Expected           model.ApplicationTemplateInput
		ExpectedError      error
	}{
		{
			Name: "All properties given",
			AppConverterFn: func() *automock.AppConverter {
				appConverter := automock.AppConverter{}
				appConverter.On("CreateInputGQLToJSON", appTemplateInputGQL.ApplicationInput).Return(appTemplateInputModel.ApplicationInputJSON, nil).Once()
				return &appConverter
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", []*graphql.WebhookInput(nil)).Return([]*model.WebhookInput(nil), nil)
				return conv
			},
			Input:         *appTemplateInputGQL,
			Expected:      *appTemplateInputModel,
			ExpectedError: nil,
		},
		{
			Name: "Error when converting Webhook",
			AppConverterFn: func() *automock.AppConverter {
				appConverter := automock.AppConverter{}
				appConverter.On("CreateInputGQLToJSON", appTemplateInputGQL.ApplicationInput).Return(appTemplateInputModel.ApplicationInputJSON, nil).Once()
				return &appConverter
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", []*graphql.WebhookInput(nil)).Return(nil, mockedError)
				return conv
			},
			Input:         *appTemplateInputGQL,
			ExpectedError: mockedError,
		},
		{
			Name: "Empty",
			AppConverterFn: func() *automock.AppConverter {
				appConverter := automock.AppConverter{}
				return &appConverter
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", []*graphql.WebhookInput(nil)).Return([]*model.WebhookInput(nil), nil)
				return conv
			},
			Input: graphql.ApplicationTemplateInput{},
			Expected: model.ApplicationTemplateInput{
				Placeholders: []model.ApplicationTemplatePlaceholder{},
			},
			ExpectedError: nil,
		},
		{
			Name: "Error when converting",
			AppConverterFn: func() *automock.AppConverter {
				appConverter := automock.AppConverter{}
				appConverter.On("CreateInputGQLToJSON", appTemplateInputGQL.ApplicationInput).Return("", testError).Once()
				return &appConverter
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", []*graphql.WebhookInput(nil)).Return([]*model.WebhookInput(nil), nil)
				return conv
			},
			Input:         *appTemplateInputGQL,
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appConv := testCase.AppConverterFn()
			webhookConv := testCase.WebhookConverterFn()
			converter := apptemplate.NewConverter(appConv, webhookConv)
			// WHEN
			res, err := converter.InputFromGraphQL(testCase.Input)

			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				assert.Equal(t, testCase.Expected, res)
			} else {
				require.Error(t, err)
			}

			appConv.AssertExpectations(t)
		})
	}
}

func TestConverter_UpdateInputFromGraphQL(t *testing.T) {
	// GIVEN
	appTemplateInputGQL := fixGQLAppTemplateUpdateInput(testName)
	appTemplateInputModel := fixModelAppTemplateUpdateInput(testName, "{\"name\":\"foo\",\"description\":\"Lorem ipsum\"}")

	testCases := []struct {
		Name           string
		AppConverterFn func() *automock.AppConverter
		Input          graphql.ApplicationTemplateUpdateInput
		Expected       model.ApplicationTemplateUpdateInput
		ExpectedError  error
	}{
		{
			Name: "All properties given",
			AppConverterFn: func() *automock.AppConverter {
				appConverter := automock.AppConverter{}
				appConverter.On("CreateInputGQLToJSON", appTemplateInputGQL.ApplicationInput).Return(appTemplateInputModel.ApplicationInputJSON, nil).Once()
				return &appConverter
			},
			Input:         *appTemplateInputGQL,
			Expected:      *appTemplateInputModel,
			ExpectedError: nil,
		},
		{
			Name: "Empty",
			AppConverterFn: func() *automock.AppConverter {
				appConverter := automock.AppConverter{}
				return &appConverter
			},
			Input: graphql.ApplicationTemplateUpdateInput{},
			Expected: model.ApplicationTemplateUpdateInput{
				Placeholders: []model.ApplicationTemplatePlaceholder{},
			},
			ExpectedError: nil,
		},
		{
			Name: "Error when converting app",
			AppConverterFn: func() *automock.AppConverter {
				appConverter := automock.AppConverter{}
				appConverter.On("CreateInputGQLToJSON", appTemplateInputGQL.ApplicationInput).Return("", testError).Once()
				return &appConverter
			},
			Input:         *appTemplateInputGQL,
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appConv := testCase.AppConverterFn()
			converter := apptemplate.NewConverter(appConv, nil)
			// WHEN
			res, err := converter.UpdateInputFromGraphQL(testCase.Input)

			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				assert.Equal(t, testCase.Expected, res)
			} else {
				require.Error(t, err)
			}

			appConv.AssertExpectations(t)
		})
	}
}

func TestConverter_ApplicationFromTemplateInputFromGraphQL(t *testing.T) {
	// GIVEN
	conv := apptemplate.NewConverter(nil, nil)
	appTemplateModel := fixModelApplicationTemplate(testID, testName, fixModelApplicationTemplateWebhooks("webhook-id-1", testID))

	in := fixGQLApplicationFromTemplateInput(testName)
	expected := fixModelApplicationFromTemplateInput(testName)

	// WHEN
	result, err := conv.ApplicationFromTemplateInputFromGraphQL(appTemplateModel, in)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestConverter_ToEntity(t *testing.T) {
	// GIVEN
	appTemplateModel := fixModelApplicationTemplate(testID, testName, fixModelApplicationTemplateWebhooks("webhook-id-1", testID))
	appTemplateEntity := fixEntityApplicationTemplate(t, testID, testName)

	testCases := []struct {
		Name     string
		Input    *model.ApplicationTemplate
		Expected *apptemplate.Entity
	}{
		{
			Name:     "All properties given",
			Input:    appTemplateModel,
			Expected: appTemplateEntity,
		},
		{
			Name:     "Empty",
			Input:    &model.ApplicationTemplate{},
			Expected: &apptemplate.Entity{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := apptemplate.NewConverter(nil, nil)

			// WHEN
			res, err := conv.ToEntity(testCase.Input)

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	// GIVEN
	id := "foo"
	name := "bar"

	appTemplateEntity := fixEntityApplicationTemplate(t, id, name)
	appTemplateModel := fixModelApplicationTemplate(id, name, nil)

	testCases := []struct {
		Name               string
		Input              *apptemplate.Entity
		Expected           *model.ApplicationTemplate
		ExpectedErrMessage string
	}{
		{
			Name:               "All properties given",
			Input:              appTemplateEntity,
			Expected:           appTemplateModel,
			ExpectedErrMessage: "",
		},
		{
			Name:               "Empty",
			Input:              &apptemplate.Entity{},
			Expected:           &model.ApplicationTemplate{},
			ExpectedErrMessage: "",
		},
		{
			Name:               "Nil",
			Input:              nil,
			Expected:           nil,
			ExpectedErrMessage: "",
		},
		{
			Name: "PlaceholdersJSON Unmarshall Error",
			Input: &apptemplate.Entity{
				PlaceholdersJSON: sql.NullString{
					String: "{dasdd",
					Valid:  true,
				},
			},
			ExpectedErrMessage: "while converting placeholders from JSON to model: invalid character 'd' looking for beginning of object key string",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := apptemplate.NewConverter(nil, nil)

			// WHEN
			res, err := conv.FromEntity(testCase.Input)

			if testCase.ExpectedErrMessage != "" {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErrMessage, err.Error())
			} else {
				require.Nil(t, err)
			}

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
