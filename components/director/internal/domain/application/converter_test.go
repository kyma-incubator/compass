package application_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application/automock"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
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
			Input:    fixDetailedModelApplication(t, "foo", "Foo", "Lorem ipsum"),
			Expected: fixDetailedGQLApplication(t, "foo", "Foo", "Lorem ipsum"),
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
		fixModelApplication("foo", "Foo", "Lorem ipsum"),
		fixModelApplication("bar", "Bar", "Dolor sit amet"),
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

func TestConverter_InputFromGraphQL(t *testing.T) {
	allPropsInput := fixGQLApplicationInput("foo", "Lorem ipsum")
	allPropsExpected := fixModelApplicationInput("foo", "Lorem ipsum")

	// given
	testCases := []struct {
		Name                string
		Input               graphql.ApplicationInput
		Expected            model.ApplicationInput
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
			Input:    graphql.ApplicationInput{},
			Expected: model.ApplicationInput{},
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleInputFromGraphQL", []*graphql.ApplicationWebhookInput(nil)).Return(nil)
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
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
