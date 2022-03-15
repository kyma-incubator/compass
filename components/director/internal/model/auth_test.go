package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"

	"github.com/stretchr/testify/assert"
)

var accessStrategy = "accessStrategy"

func TestAuthInput_ToAuth(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *auth.AuthInput
		Expected *auth.Auth
	}{
		{
			Name: "All properties given",
			Input: &auth.AuthInput{
				Credential: &auth.CredentialDataInput{
					Basic: &auth.BasicCredentialDataInput{
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
				RequestAuth: &auth.CredentialRequestAuthInput{
					Csrf: &auth.CSRFTokenCredentialRequestAuthInput{
						TokenEndpointURL: "test",
					},
				},
			},
			Expected: &auth.Auth{
				Credential: auth.CredentialData{
					Basic: &auth.BasicCredentialData{
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
				RequestAuth: &auth.CredentialRequestAuth{
					Csrf: &auth.CSRFTokenCredentialRequestAuth{
						TokenEndpointURL: "test",
					},
				},
			},
		},
		{
			Name:     "Empty",
			Input:    &auth.AuthInput{},
			Expected: &auth.Auth{},
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
		Input    *auth.CredentialDataInput
		Expected *auth.CredentialData
	}{
		{
			Name: "All properties given",
			Input: &auth.CredentialDataInput{
				Basic: &auth.BasicCredentialDataInput{
					Username: "user",
				},
				Oauth: &auth.OAuthCredentialDataInput{
					URL: "test",
				},
			},
			Expected: &auth.CredentialData{
				Basic: &auth.BasicCredentialData{
					Username: "user",
				},
				Oauth: &auth.OAuthCredentialData{
					URL: "test",
				},
			},
		},
		{
			Name:     "Empty",
			Input:    &auth.CredentialDataInput{},
			Expected: &auth.CredentialData{},
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
		Input    *auth.BasicCredentialDataInput
		Expected *auth.BasicCredentialData
	}{
		{
			Name: "All properties given",
			Input: &auth.BasicCredentialDataInput{
				Username: "user",
				Password: "pass",
			},
			Expected: &auth.BasicCredentialData{
				Username: "user",
				Password: "pass",
			},
		},
		{
			Name:     "Empty",
			Input:    &auth.BasicCredentialDataInput{},
			Expected: &auth.BasicCredentialData{},
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
		Input    *auth.OAuthCredentialDataInput
		Expected *auth.OAuthCredentialData
	}{
		{
			Name: "All properties given",
			Input: &auth.OAuthCredentialDataInput{
				URL:          "test",
				ClientID:     "id",
				ClientSecret: "secret",
			},
			Expected: &auth.OAuthCredentialData{
				URL:          "test",
				ClientID:     "id",
				ClientSecret: "secret",
			},
		},
		{
			Name:     "Empty",
			Input:    &auth.OAuthCredentialDataInput{},
			Expected: &auth.OAuthCredentialData{},
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
		Input    *auth.CredentialRequestAuthInput
		Expected *auth.CredentialRequestAuth
	}{
		{
			Name: "All properties given",
			Input: &auth.CredentialRequestAuthInput{
				Csrf: &auth.CSRFTokenCredentialRequestAuthInput{
					TokenEndpointURL: "foo.bar",
				},
			},
			Expected: &auth.CredentialRequestAuth{
				Csrf: &auth.CSRFTokenCredentialRequestAuth{
					TokenEndpointURL: "foo.bar",
				},
			},
		},
		{
			Name:     "Empty",
			Input:    &auth.CredentialRequestAuthInput{},
			Expected: &auth.CredentialRequestAuth{},
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
		Input    *auth.CSRFTokenCredentialRequestAuthInput
		Expected *auth.CSRFTokenCredentialRequestAuth
	}{
		{
			Name: "All properties given",
			Input: &auth.CSRFTokenCredentialRequestAuthInput{
				Credential: &auth.CredentialDataInput{
					Basic: &auth.BasicCredentialDataInput{
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
			Expected: &auth.CSRFTokenCredentialRequestAuth{
				Credential: auth.CredentialData{
					Basic: &auth.BasicCredentialData{
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
			Input:    &auth.CSRFTokenCredentialRequestAuthInput{},
			Expected: &auth.CSRFTokenCredentialRequestAuth{},
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
