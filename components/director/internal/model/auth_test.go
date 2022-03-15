package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/stretchr/testify/assert"
)

var accessStrategy = "accessStrategy"

func TestAuthInput_ToAuth(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *model.AuthInput
		Expected *model.Auth
	}{
		{
			Name: "All properties given",
			Input: &model.AuthInput{
				Credential: &model.CredentialDataInput{
					Basic: &model.BasicCredentialDataInput{
						Username: "test",
					},
				},
				AccessStrategy: &accessStrategy,
				AdditionalQueryParams: map[string][]string{
					"key": {"value1", "value2"},
				},
				AdditionalHeaders: map[string][]string{
					"header": {"value1", "value2"},
				},
				RequestAuth: &model.CredentialRequestAuthInput{
					Csrf: &model.CSRFTokenCredentialRequestAuthInput{
						TokenEndpointURL: "test",
					},
				},
			},
			Expected: &model.Auth{
				Credential: model.CredentialData{
					Basic: &model.BasicCredentialData{
						Username: "test",
					},
				},
				AccessStrategy: &accessStrategy,
				AdditionalQueryParams: map[string][]string{
					"key": {"value1", "value2"},
				},
				AdditionalHeaders: map[string][]string{
					"header": {"value1", "value2"},
				},
				RequestAuth: &model.CredentialRequestAuth{
					Csrf: &model.CSRFTokenCredentialRequestAuth{
						TokenEndpointURL: "test",
					},
				},
			},
		},
		{
			Name:     "Empty",
			Input:    &model.AuthInput{},
			Expected: &model.Auth{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToAuth()

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestCredentialDataInput_ToCredentialData(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *model.CredentialDataInput
		Expected *model.CredentialData
	}{
		{
			Name: "All properties given",
			Input: &model.CredentialDataInput{
				Basic: &model.BasicCredentialDataInput{
					Username: "user",
				},
				Oauth: &model.OAuthCredentialDataInput{
					URL: "test",
				},
			},
			Expected: &model.CredentialData{
				Basic: &model.BasicCredentialData{
					Username: "user",
				},
				Oauth: &model.OAuthCredentialData{
					URL: "test",
				},
			},
		},
		{
			Name:     "Empty",
			Input:    &model.CredentialDataInput{},
			Expected: &model.CredentialData{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToCredentialData()

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestBasicCredentialDataInput_ToBasicCredentialData(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *model.BasicCredentialDataInput
		Expected *model.BasicCredentialData
	}{
		{
			Name: "All properties given",
			Input: &model.BasicCredentialDataInput{
				Username: "user",
				Password: "pass",
			},
			Expected: &model.BasicCredentialData{
				Username: "user",
				Password: "pass",
			},
		},
		{
			Name:     "Empty",
			Input:    &model.BasicCredentialDataInput{},
			Expected: &model.BasicCredentialData{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToBasicCredentialData()

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestOAuthCredentialDataInput_ToOAuthCredentialData(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *model.OAuthCredentialDataInput
		Expected *model.OAuthCredentialData
	}{
		{
			Name: "All properties given",
			Input: &model.OAuthCredentialDataInput{
				URL:          "test",
				ClientID:     "id",
				ClientSecret: "secret",
			},
			Expected: &model.OAuthCredentialData{
				URL:          "test",
				ClientID:     "id",
				ClientSecret: "secret",
			},
		},
		{
			Name:     "Empty",
			Input:    &model.OAuthCredentialDataInput{},
			Expected: &model.OAuthCredentialData{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToOAuthCredentialData()

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestCredentialRequestAuthInput_ToCredentialRequestAuth(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *model.CredentialRequestAuthInput
		Expected *model.CredentialRequestAuth
	}{
		{
			Name: "All properties given",
			Input: &model.CredentialRequestAuthInput{
				Csrf: &model.CSRFTokenCredentialRequestAuthInput{
					TokenEndpointURL: "foo.bar",
				},
			},
			Expected: &model.CredentialRequestAuth{
				Csrf: &model.CSRFTokenCredentialRequestAuth{
					TokenEndpointURL: "foo.bar",
				},
			},
		},
		{
			Name:     "Empty",
			Input:    &model.CredentialRequestAuthInput{},
			Expected: &model.CredentialRequestAuth{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToCredentialRequestAuth()

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestCSRFTokenCredentialRequestAuthInput_ToCSRFTokenCredentialRequestAuth(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *model.CSRFTokenCredentialRequestAuthInput
		Expected *model.CSRFTokenCredentialRequestAuth
	}{
		{
			Name: "All properties given",
			Input: &model.CSRFTokenCredentialRequestAuthInput{
				Credential: &model.CredentialDataInput{
					Basic: &model.BasicCredentialDataInput{
						Username: "test",
					},
				},
				TokenEndpointURL: "foo.bar",
				AdditionalQueryParams: map[string][]string{
					"key": {"value1", "value2"},
				},
				AdditionalHeaders: map[string][]string{
					"header": {"value1", "value2"},
				},
			},
			Expected: &model.CSRFTokenCredentialRequestAuth{
				Credential: model.CredentialData{
					Basic: &model.BasicCredentialData{
						Username: "test",
					},
				},
				TokenEndpointURL: "foo.bar",
				AdditionalQueryParams: map[string][]string{
					"key": {"value1", "value2"},
				},
				AdditionalHeaders: map[string][]string{
					"header": {"value1", "value2"},
				},
			},
		},
		{
			Name:     "Empty",
			Input:    &model.CSRFTokenCredentialRequestAuthInput{},
			Expected: &model.CSRFTokenCredentialRequestAuth{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToCSRFTokenCredentialRequestAuth()

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
