package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/stretchr/testify/require"
)

func TestAuthInput_Validate_Credential(t *testing.T) {
	credential := fixValidCredentialDataInput()

	testCases := []struct {
		Name          string
		Value         *graphql.CredentialDataInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &credential,
			ExpectedValid: true,
		},
		{
			Name:          "Empty",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - nested validation error",
			Value:         &graphql.CredentialDataInput{},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidAuthInput()
			sut.Credential = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAuthInput_Validate_AdditionalHeaders(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *graphql.HttpHeaders
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: &graphql.HttpHeaders{
				"Authorization": {"test", "asdf"},
				"Test":          {"test", "asdf"},
			},
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid - nil",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name: "Invalid - empty key",
			Value: &graphql.HttpHeaders{
				inputvalidationtest.EmptyString: {"test"},
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - nil value",
			Value: &graphql.HttpHeaders{
				"test": nil,
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - empty slice element",
			Value: &graphql.HttpHeaders{
				"test": {inputvalidationtest.EmptyString},
			},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidAuthInput()
			sut.AdditionalHeaders = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAuthInput_Validate_AdditionalQueryParams(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *graphql.QueryParams
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: &graphql.QueryParams{
				"Param": {"test", "asdf"},
				"Test":  {"test", "asdf"},
			},
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid - nil",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name: "Invalid - empty key",
			Value: &graphql.QueryParams{
				inputvalidationtest.EmptyString: {"test"},
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - nil value",
			Value: &graphql.QueryParams{
				"test": nil,
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - empty slice element",
			Value: &graphql.QueryParams{
				"test": {inputvalidationtest.EmptyString},
			},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidAuthInput()
			sut.AdditionalQueryParams = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAuthInput_Validate_RequestAuth(t *testing.T) {
	credRequest := fixValidCredentialRequestAuthInput()
	testCases := []struct {
		Name          string
		Value         *graphql.CredentialRequestAuthInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &credRequest,
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid - nil",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name: "Invalid - no auth provided",
			Value: &graphql.CredentialRequestAuthInput{
				Csrf: nil,
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - nested validation error",
			Value: &graphql.CredentialRequestAuthInput{
				Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{},
			},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidAuthInput()
			sut.RequestAuth = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCredentialDataInput_Validate(t *testing.T) {
	credential := fixValidCredentialDataInput()
	basic := fixValidBasicCredentialDataInput()
	oauth := fixValidOAuthCredentialDataInput()

	testCases := []struct {
		Name          string
		Value         *graphql.CredentialDataInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &credential,
			ExpectedValid: true,
		},
		{
			Name: "Invalid - no auth provided",
			Value: &graphql.CredentialDataInput{
				Basic: nil,
				Oauth: nil,
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - multiple auths provided",
			Value: &graphql.CredentialDataInput{
				Basic: &basic,
				Oauth: &oauth,
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - nested validation error in Basic",
			Value: &graphql.CredentialDataInput{
				Basic: &graphql.BasicCredentialDataInput{},
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - nested validation error in Oauth",
			Value: &graphql.CredentialDataInput{
				Oauth: &graphql.OAuthCredentialDataInput{},
			},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidAuthInput()
			sut.Credential = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBasicCredentialDataInput_Validate_Username(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "John",
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidBasicCredentialDataInput()
			sut.Username = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBasicCredentialDataInput_Validate_Password(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "John",
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidBasicCredentialDataInput()
			sut.Password = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestOAuthCredentialDataInput_Validate_ClientID(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "John.2h2kj2k5gw6j3h5gjk34hg-g:0",
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidOAuthCredentialDataInput()
			sut.ClientID = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestOAuthCredentialDataInput_Validate_ClientSecret(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "Doe.2h2kj2k5gw6j3h5gjk34hg-g:0",
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidOAuthCredentialDataInput()
			sut.ClientSecret = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestOAuthCredentialDataInput_Validate_URL(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         inputvalidationtest.ValidURL,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - Invalid URL",
			Value:         inputvalidationtest.InvalidURL,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidOAuthCredentialDataInput()
			sut.URL = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCredentialRequestAuthInput_Validate(t *testing.T) {
	csrf := fixValidCSRFTokenCredentialRequestAuthInput()
	testCases := []struct {
		Name          string
		Value         *graphql.CredentialRequestAuthInput
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: &graphql.CredentialRequestAuthInput{
				Csrf: &csrf,
			},
			ExpectedValid: true,
		},
		{
			Name: "Invalid - no auth provided",
			Value: &graphql.CredentialRequestAuthInput{
				Csrf: nil,
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - Nested validation error",
			Value: &graphql.CredentialRequestAuthInput{
				Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{},
			},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCSRFTokenCredentialRequestAuthInput_Validate_Credential(t *testing.T) {
	credential := fixValidCredentialDataInput()

	testCases := []struct {
		Name          string
		Value         *graphql.CredentialDataInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &credential,
			ExpectedValid: true,
		},
		{
			Name:          "Valid - empty",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name: "Invalid - nested validation error",
			Value: &graphql.CredentialDataInput{
				Basic: nil,
				Oauth: nil,
			},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidCSRFTokenCredentialRequestAuthInput()
			sut.Credential = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCSRFTokenCredentialRequestAuthInput_Validate_AdditionalHeaders(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *graphql.HttpHeaders
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: &graphql.HttpHeaders{
				"Authorization": {"test", "asdf"},
				"Test":          {"test", "asdf"},
			},
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid - nil",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name: "Invalid - empty key",
			Value: &graphql.HttpHeaders{
				inputvalidationtest.EmptyString: {"test"},
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - nil value",
			Value: &graphql.HttpHeaders{
				"test": nil,
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - empty slice element",
			Value: &graphql.HttpHeaders{
				"test": {inputvalidationtest.EmptyString},
			},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidCSRFTokenCredentialRequestAuthInput()
			sut.AdditionalHeaders = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCSRFTokenCredentialRequestAuthInput_Validate_AdditionalQueryParams(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *graphql.QueryParams
		ExpectedValid bool
	}{
		{
			Name: "ExpectedValid",
			Value: &graphql.QueryParams{
				"Param": {"test", "asdf"},
				"Test":  {"test", "asdf"},
			},
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid - nil",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name: "Invalid - empty key",
			Value: &graphql.QueryParams{
				inputvalidationtest.EmptyString: {"test"},
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - nil value",
			Value: &graphql.QueryParams{
				"test": nil,
			},
			ExpectedValid: false,
		},
		{
			Name: "Invalid - empty slice element",
			Value: &graphql.QueryParams{
				"test": {inputvalidationtest.EmptyString},
			},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidCSRFTokenCredentialRequestAuthInput()
			sut.AdditionalQueryParams = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCSRFTokenCredentialRequestAuthInput_Validate_TokenEndpointURL(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         inputvalidationtest.ValidURL,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid - Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "Invalid - Invalid URL",
			Value:         inputvalidationtest.InvalidURL,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidCSRFTokenCredentialRequestAuthInput()
			sut.TokenEndpointURL = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidAuthInput() graphql.AuthInput {
	credential := fixValidCredentialDataInput()
	return graphql.AuthInput{
		Credential: &credential,
	}
}

func fixValidCredentialDataInput() graphql.CredentialDataInput {
	basic := fixValidBasicCredentialDataInput()
	return graphql.CredentialDataInput{
		Basic: &basic,
		Oauth: nil,
	}
}

func fixValidBasicCredentialDataInput() graphql.BasicCredentialDataInput {
	return graphql.BasicCredentialDataInput{
		Username: "John",
		Password: "P3!@2sklasdjkfla",
	}
}

func fixValidOAuthCredentialDataInput() graphql.OAuthCredentialDataInput {
	return graphql.OAuthCredentialDataInput{
		ClientID:     "client",
		ClientSecret: "secret",
		URL:          "http://valid.url",
	}
}

func fixValidCredentialRequestAuthInput() graphql.CredentialRequestAuthInput {
	csrf := fixValidCSRFTokenCredentialRequestAuthInput()
	return graphql.CredentialRequestAuthInput{
		Csrf: &csrf,
	}
}

func fixValidCSRFTokenCredentialRequestAuthInput() graphql.CSRFTokenCredentialRequestAuthInput {
	credential := fixValidCredentialDataInput()
	return graphql.CSRFTokenCredentialRequestAuthInput{
		TokenEndpointURL:      "http://valid.url",
		Credential:            &credential,
		AdditionalHeaders:     nil,
		AdditionalQueryParams: nil,
	}
}
