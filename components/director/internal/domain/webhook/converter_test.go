package webhook_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *model.ApplicationWebhook
		Expected *graphql.ApplicationWebhook
	}{
		{
			Name:     "All properties given",
			Input:    fixModelWebhook("1", "foo", "bar"),
			Expected: fixGQLWebhook("foo", "bar"),
		},
		{
			Name:     "Empty",
			Input:    &model.ApplicationWebhook{},
			Expected: &graphql.ApplicationWebhook{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			authConv := &automock.AuthConverter{}
			if testCase.Input != nil {
				authConv.On("ToGraphQL", testCase.Input.Auth).Return(testCase.Expected.Auth)
			}
			converter := webhook.NewConverter(authConv)
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	input := []*model.ApplicationWebhook{
		fixModelWebhook("1", "foo", "baz"),
		fixModelWebhook("2", "bar", "bez"),
		{},
		nil,
	}
	expected := []*graphql.ApplicationWebhook{
		fixGQLWebhook("foo", "baz"),
		fixGQLWebhook("bar", "bez"),
		{},
	}
	var nilAuth *model.Auth

	// when
	authConv := &automock.AuthConverter{}
	authConv.On("ToGraphQL", input[0].Auth).Return(expected[0].Auth)
	authConv.On("ToGraphQL", nilAuth).Return(nil)
	converter := webhook.NewConverter(authConv)
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *graphql.ApplicationWebhookInput
		Expected *model.ApplicationWebhookInput
	}{
		{
			Name:     "All properties given",
			Input:    fixGQLWebhookInput("foo"),
			Expected: fixModelWebhookInput("foo"),
		},
		{
			Name:     "Empty",
			Input:    &graphql.ApplicationWebhookInput{},
			Expected: &model.ApplicationWebhookInput{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			authConv := &automock.AuthConverter{}
			if testCase.Input != nil {
				authConv.On("InputFromGraphQL", testCase.Input.Auth).Return(testCase.Expected.Auth)
			}
			converter := webhook.NewConverter(authConv)
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	// given
	input := []*graphql.ApplicationWebhookInput{
		fixGQLWebhookInput("foo"),
		fixGQLWebhookInput("bar"),
		{},
		nil,
	}
	expected := []*model.ApplicationWebhookInput{
		fixModelWebhookInput("foo"),
		fixModelWebhookInput("bar"),
		{},
	}
	var nilAuthInput *graphql.AuthInput

	// when
	authConv := &automock.AuthConverter{}
	authConv.On("InputFromGraphQL", input[0].Auth).Return(expected[0].Auth)
	authConv.On("InputFromGraphQL", nilAuthInput).Return(nil)
	converter := webhook.NewConverter(authConv)
	res := converter.MultipleInputFromGraphQL(input)

	// then
	assert.Equal(t, expected, res)
}
