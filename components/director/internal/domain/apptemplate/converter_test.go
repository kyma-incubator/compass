package apptemplate_test

import (
	"database/sql"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	appConv := &automock.AppConverter{}
	converter := apptemplate.NewConverter(appConv)

	testCases := []struct {
		Name          string
		Input         *model.ApplicationTemplate
		Expected      *graphql.ApplicationTemplate
		ExpectedError bool
	}{
		{
			Name:          "All properties given",
			Input:         fixModelAppTemplate(testID, testName),
			Expected:      fixGQLAppTemplate(testID, testName),
			ExpectedError: false,
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
		},
		{
			Name:          "Empty",
			Input:         &model.ApplicationTemplate{},
			Expected:      nil,
			ExpectedError: true,
		},
		{
			Name:          "Nil",
			Input:         nil,
			Expected:      nil,
			ExpectedError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
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
	converter := apptemplate.NewConverter(nil)

	testCases := []struct {
		Name          string
		Input         []*model.ApplicationTemplate
		Expected      []*graphql.ApplicationTemplate
		ExpectedError bool
	}{
		{
			Name: "All properties given",
			Input: []*model.ApplicationTemplate{
				fixModelAppTemplate("id1", "name1"),
				fixModelAppTemplate("id2", "name2"),
				nil,
			},
			Expected: []*graphql.ApplicationTemplate{
				fixGQLAppTemplate("id1", "name1"),
				fixGQLAppTemplate("id2", "name2"),
			},
			ExpectedError: false,
		},
		{
			Name: "Error when application input is empty",
			Input: []*model.ApplicationTemplate{
				fixModelAppTemplate("id1", "name1"),
				{
					ID:                   testID,
					Name:                 testName,
					ApplicationInputJSON: "",
				},
			},
			Expected:      nil,
			ExpectedError: true,
		},
		{
			Name: "Error when converting application template",
			Input: []*model.ApplicationTemplate{
				fixModelAppTemplate("id1", "name1"),
				{
					ID:                   testID,
					Name:                 testName,
					ApplicationInputJSON: "{abc",
				},
			},
			Expected:      nil,
			ExpectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
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
		Name           string
		AppConverterFn func() *automock.AppConverter
		Input          graphql.ApplicationTemplateInput
		Expected       model.ApplicationTemplateInput
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
			Input:         graphql.ApplicationTemplateInput{},
			Expected:      model.ApplicationTemplateInput{},
			ExpectedError: nil,
		},
		{
			Name: "Error when converting",
			AppConverterFn: func() *automock.AppConverter {
				appConverter := automock.AppConverter{}
				appConverter.On("CreateInputGQLToJSON", appTemplateInputGQL.ApplicationInput).Return("", testError).Once()
				return &appConverter
			},
			Input:         *appTemplateInputGQL,
			Expected:      model.ApplicationTemplateInput{},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appConv := testCase.AppConverterFn()
			converter := apptemplate.NewConverter(appConv)
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

func TestConverter_ApplicationFromTemplateInputFromGraphQL(t *testing.T) {
	// GIVEN
	conv := apptemplate.NewConverter(nil)

	in := fixGQLApplicationFromTemplateInput(testName)
	expected := fixModelApplicationFromTemplateInput(testName)

	// WHEN
	result := conv.ApplicationFromTemplateInputFromGraphQL(in)

	// THEN
	assert.Equal(t, expected, result)
}

func TestConverter_ToEntity(t *testing.T) {
	// given
	appTemplateModel := fixModelAppTemplate(testID, testName)
	appTemplateEntity := fixEntityAppTemplate(t, testID, testName)

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
			conv := apptemplate.NewConverter(nil)

			// when
			res, err := conv.ToEntity(testCase.Input)

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	// given
	id := "foo"
	name := "bar"

	appTemplateEntity := fixEntityAppTemplate(t, id, name)
	appTemplateModel := fixModelAppTemplate(id, name)

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
			conv := apptemplate.NewConverter(nil)

			// when
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
