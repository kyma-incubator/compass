package webhook_test

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *model.Webhook
		Expected *graphql.Webhook
	}{
		{
			Name:     "All properties given",
			Input:    fixModelWebhook("1", "foo", "", "bar"),
			Expected: fixGQLWebhook("1", "foo", "bar"),
		},
		{
			Name:     "Empty",
			Input:    &model.Webhook{},
			Expected: &graphql.Webhook{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := &automock.AuthConverter{}
			if testCase.Input != nil {
				authConv.On("ToGraphQL", testCase.Input.Auth).Return(testCase.Expected.Auth)
			}
			converter := webhook.NewConverter(authConv)

			// when
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			authConv.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	input := []*model.Webhook{
		fixModelWebhook("1", "foo", "", "baz"),
		fixModelWebhook("2", "bar", "", "bez"),
		{},
		nil,
	}
	expected := []*graphql.Webhook{
		fixGQLWebhook("1", "foo", "baz"),
		fixGQLWebhook("2", "bar", "bez"),
		{},
	}
	authConv := &automock.AuthConverter{}
	authConv.On("ToGraphQL", input[0].Auth).Return(expected[0].Auth)
	authConv.On("ToGraphQL", (*model.Auth)(nil)).Return(nil)
	converter := webhook.NewConverter(authConv)

	// when
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
	authConv.AssertExpectations(t)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *graphql.WebhookInput
		Expected *model.WebhookInput
	}{
		{
			Name:     "All properties given",
			Input:    fixGQLWebhookInput("foo"),
			Expected: fixModelWebhookInput("foo"),
		},
		{
			Name:     "Empty",
			Input:    &graphql.WebhookInput{},
			Expected: &model.WebhookInput{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := &automock.AuthConverter{}
			if testCase.Input != nil {
				authConv.On("InputFromGraphQL", testCase.Input.Auth).Return(testCase.Expected.Auth)
			}
			converter := webhook.NewConverter(authConv)

			// when
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			authConv.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	// given
	input := []*graphql.WebhookInput{
		fixGQLWebhookInput("foo"),
		fixGQLWebhookInput("bar"),
		{},
		nil,
	}
	expected := []*model.WebhookInput{
		fixModelWebhookInput("foo"),
		fixModelWebhookInput("bar"),
		{},
	}
	authConv := &automock.AuthConverter{}
	authConv.On("InputFromGraphQL", input[0].Auth).Return(expected[0].Auth)
	authConv.On("InputFromGraphQL", (*graphql.AuthInput)(nil)).Return(nil)
	converter := webhook.NewConverter(authConv)

	// when
	res := converter.MultipleInputFromGraphQL(input)

	// then
	assert.Equal(t, expected, res)
	authConv.AssertExpectations(t)
}

func TestConverter_ToEntity(t *testing.T) {
	sut := webhook.NewConverter(nil)

	b, err := json.Marshal(givenBasicAuth())
	require.NoError(t, err)
	expectedBasicAuthAsString := string(b)

	testCases := map[string]struct {
		in       model.Webhook
		expected webhook.Entity
	}{
		"success when Auth not provided": {
			in: model.Webhook{
				ID:            "givenID",
				ApplicationID: "givenApplicationID",
				URL:           "givenURL",
				Tenant:        "givenTenant",
				Type:          model.WebhookTypeConfigurationChanged,
			},
			expected: webhook.Entity{
				ID:       "givenID",
				AppID:    "givenApplicationID",
				URL:      "givenURL",
				TenantID: "givenTenant",
				Type:     "CONFIGURATION_CHANGED",
				Auth:     sql.NullString{Valid: false},
			},
		},
		"success when Auth provided": {
			in: model.Webhook{
				Auth: givenBasicAuth(),
			},
			expected: webhook.Entity{
				Auth: sql.NullString{Valid: true, String: expectedBasicAuthAsString},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			// WHEN
			actual, err := sut.ToEntity(tc.in)
			// THEN
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}

}

func TestConverter_FromEntity(t *testing.T) {
	// GIVEN
	sut := webhook.NewConverter(nil)
	b, err := json.Marshal(givenBasicAuth())
	require.NoError(t, err)

	testCases := map[string]struct {
		inEntity      webhook.Entity
		expectedModel model.Webhook
		expectedErr   error
	}{
		"success when Auth not provided": {
			inEntity: webhook.Entity{
				ID:       "givenID",
				TenantID: "givenTenant",
				Type:     "CONFIGURATION_CHANGED",
				URL:      "givenURL",
				AppID:    "givenAppID",
			},
			expectedModel: model.Webhook{
				ID:            "givenID",
				Tenant:        "givenTenant",
				Type:          "CONFIGURATION_CHANGED",
				URL:           "givenURL",
				ApplicationID: "givenAppID",
				Auth:          nil,
			},
		},
		"success when Auth provided": {
			inEntity: webhook.Entity{
				ID: "givenID",
				Auth: sql.NullString{
					Valid:  true,
					String: string(b),
				},
			},
			expectedModel: model.Webhook{
				ID:   "givenID",
				Auth: givenBasicAuth(),
			},
		},
		"got error on unmarshaling JSON": {
			inEntity: webhook.Entity{
				Auth: sql.NullString{
					Valid:  true,
					String: "it is not even a proper JSON!",
				},
			},
			expectedErr: errors.New("while unmarshaling Auth: invalid character 'i' looking for beginning of value"),
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			actual, err := sut.FromEntity(tc.inEntity)
			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedModel, actual)
		})
	}
}

func givenBasicAuth() *model.Auth {
	return &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: "aaa",
				Password: "bbb",
			},
		},
	}
}
