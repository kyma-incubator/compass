package webhook_test

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

var (
	givenAppID         = "givenApplicationID"
	modelWebhookMode   = model.WebhookModeSync
	graphqlWebhookMode = graphql.WebhookModeSync
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *model.Webhook
		Expected *graphql.Webhook
	}{
		{
			Name:     "All properties given",
			Input:    fixApplicationModelWebhook("1", "foo", "", "bar", time.Time{}),
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
				authConv.On("ToGraphQL", testCase.Input.Auth).Return(testCase.Expected.Auth, nil)
			}
			converter := webhook.NewConverter(authConv)

			// WHEN
			res, err := converter.ToGraphQL(testCase.Input)

			// THEN
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
			authConv.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// GIVEN
	input := []*model.Webhook{
		fixApplicationModelWebhook("1", "foo", "", "baz", time.Time{}),
		fixApplicationModelWebhook("2", "bar", "", "bez", time.Time{}),
		{},
		nil,
	}
	expected := []*graphql.Webhook{
		fixGQLWebhook("1", "foo", "baz"),
		fixGQLWebhook("2", "bar", "bez"),
		{},
	}
	authConv := &automock.AuthConverter{}
	authConv.On("ToGraphQL", input[0].Auth).Return(expected[0].Auth, nil)
	authConv.On("ToGraphQL", (*model.Auth)(nil)).Return(nil, nil)
	converter := webhook.NewConverter(authConv)

	// WHEN
	res, err := converter.MultipleToGraphQL(input)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
	authConv.AssertExpectations(t)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *graphql.WebhookInput
		Expected *model.WebhookInput
		Error    error
	}{
		{
			Name:     "All properties given",
			Input:    fixGQLWebhookInput("https://test-domain.com"),
			Expected: fixModelWebhookInput("https://test-domain.com"),
			Error:    nil,
		},
		{
			Name:     "Empty",
			Input:    &graphql.WebhookInput{},
			Expected: &model.WebhookInput{},
			Error:    nil,
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
			Error:    nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			authConv := &automock.AuthConverter{}
			if testCase.Input != nil && testCase.Error == nil {
				authConv.On("InputFromGraphQL", testCase.Input.Auth).Return(testCase.Expected.Auth, nil)
			}
			converter := webhook.NewConverter(authConv)

			// WHEN
			res, err := converter.InputFromGraphQL(testCase.Input)

			// then
			if testCase.Error == nil {
				assert.NoError(t, err)
				assert.Equal(t, testCase.Expected, res)
			} else {
				assert.Error(t, err, testCase.Error)
			}
			authConv.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	// GIVEN
	input := []*graphql.WebhookInput{
		fixGQLWebhookInput("https://test-domain.com"),
		fixGQLWebhookInput("https://test-domain.com"),
		nil,
	}
	expected := []*model.WebhookInput{
		fixModelWebhookInput("https://test-domain.com"),
		fixModelWebhookInput("https://test-domain.com"),
	}
	authConv := &automock.AuthConverter{}
	authConv.On("InputFromGraphQL", input[0].Auth).Return(expected[0].Auth, nil)
	converter := webhook.NewConverter(authConv)

	// WHEN
	res, err := converter.MultipleInputFromGraphQL(input)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
	authConv.AssertExpectations(t)
}

func TestConverter_ToEntity(t *testing.T) {
	sut := webhook.NewConverter(nil)

	b, err := json.Marshal(fixBasicAuth())
	require.NoError(t, err)
	expectedBasicAuthAsString := string(b)

	testCases := map[string]struct {
		in       *model.Webhook
		expected *webhook.Entity
	}{
		"success when Auth not provided": {
			in: &model.Webhook{
				ID:             "givenID",
				ObjectID:       givenAppID,
				ObjectType:     model.ApplicationWebhookReference,
				URL:            stringPtr("https://test-domain.com"),
				Type:           model.WebhookTypeConfigurationChanged,
				Mode:           &modelWebhookMode,
				URLTemplate:    &emptyTemplate,
				InputTemplate:  &emptyTemplate,
				HeaderTemplate: &emptyTemplate,
				OutputTemplate: &emptyTemplate,
			},
			expected: &webhook.Entity{
				ID:             "givenID",
				ApplicationID:  repo.NewValidNullableString(givenAppID),
				URL:            repo.NewValidNullableString("https://test-domain.com"),
				Type:           "CONFIGURATION_CHANGED",
				Auth:           sql.NullString{Valid: false},
				Mode:           repo.NewValidNullableString("SYNC"),
				URLTemplate:    repo.NewValidNullableString(emptyTemplate),
				InputTemplate:  repo.NewValidNullableString(emptyTemplate),
				HeaderTemplate: repo.NewValidNullableString(emptyTemplate),
				OutputTemplate: repo.NewValidNullableString(emptyTemplate),
			},
		},
		"success when Auth provided": {
			in: &model.Webhook{
				Auth: fixBasicAuth(),
			},
			expected: &webhook.Entity{
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
	b, err := json.Marshal(fixBasicAuth())
	require.NoError(t, err)

	testCases := map[string]struct {
		inEntity      *webhook.Entity
		expectedModel *model.Webhook
		expectedErr   error
	}{
		"success when Auth not provided": {
			inEntity: &webhook.Entity{
				ID:             "givenID",
				Type:           "CONFIGURATION_CHANGED",
				URL:            repo.NewValidNullableString("https://test-domain.com"),
				ApplicationID:  repo.NewValidNullableString(givenAppID),
				Mode:           repo.NewValidNullableString("SYNC"),
				URLTemplate:    repo.NewValidNullableString(emptyTemplate),
				InputTemplate:  repo.NewValidNullableString(emptyTemplate),
				HeaderTemplate: repo.NewValidNullableString(emptyTemplate),
				OutputTemplate: repo.NewValidNullableString(emptyTemplate),
			},
			expectedModel: &model.Webhook{
				ID:             "givenID",
				Type:           "CONFIGURATION_CHANGED",
				URL:            stringPtr("https://test-domain.com"),
				ObjectID:       givenAppID,
				ObjectType:     model.ApplicationWebhookReference,
				Auth:           nil,
				Mode:           &modelWebhookMode,
				URLTemplate:    &emptyTemplate,
				InputTemplate:  &emptyTemplate,
				HeaderTemplate: &emptyTemplate,
				OutputTemplate: &emptyTemplate,
			},
		},
		"success when Auth provided": {
			inEntity: &webhook.Entity{
				ID:            "givenID",
				ApplicationID: repo.NewValidNullableString("appID"),
				Auth: sql.NullString{
					Valid:  true,
					String: string(b),
				},
			},
			expectedModel: &model.Webhook{
				ID:         "givenID",
				ObjectID:   "appID",
				ObjectType: model.ApplicationWebhookReference,
				Auth:       fixBasicAuth(),
			},
		},
		"got error on unmarshaling JSON": {
			inEntity: &webhook.Entity{
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
