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
		Name  string
		Value *graphql.CredentialDataInput
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: &credential,
			Valid: true,
		},
		{
			Name:  "Invalid - nil",
			Value: nil,
			Valid: false,
		},
		{
			Name:  "Invalid - nested validation error",
			Value: &graphql.CredentialDataInput{},
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAuthInput_Validate_AdditionalHeaders(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *graphql.HttpHeaders
		Valid bool
	}{
		{
			Name: "Valid",
			Value: &graphql.HttpHeaders{
				"Authorization": {"test", "asdf"},
				"Test":          {"test", "asdf"},
			},
			Valid: true,
		},
		{
			Name:  "Valid - nil",
			Value: nil,
			Valid: true,
		},
		{
			Name: "Invalid - empty key",
			Value: &graphql.HttpHeaders{
				inputvalidationtest.EmptyString: {"test"},
			},
			Valid: false,
		},
		{
			Name: "Invalid - nil value",
			Value: &graphql.HttpHeaders{
				"test": nil,
			},
			Valid: false,
		},
		{
			Name: "Invalid - empty slice element",
			Value: &graphql.HttpHeaders{
				"test": {inputvalidationtest.EmptyString},
			},
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAuthInput_Validate_AdditionalQueryParams(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *graphql.QueryParams
		Valid bool
	}{
		{
			Name: "Valid",
			Value: &graphql.QueryParams{
				"Param": {"test", "asdf"},
				"Test":  {"test", "asdf"},
			},
			Valid: true,
		},
		{
			Name:  "Valid - nil",
			Value: nil,
			Valid: true,
		},
		{
			Name: "Invalid - empty key",
			Value: &graphql.QueryParams{
				inputvalidationtest.EmptyString: {"test"},
			},
			Valid: false,
		},
		{
			Name: "Invalid - nil value",
			Value: &graphql.QueryParams{
				"test": nil,
			},
			Valid: false,
		},
		{
			Name: "Invalid - empty slice element",
			Value: &graphql.QueryParams{
				"test": {inputvalidationtest.EmptyString},
			},
			Valid: false,
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
			if testCase.Valid {
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
		Name  string
		Value *graphql.CredentialRequestAuthInput
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: &credRequest,
			Valid: true,
		},
		{
			Name:  "Valid - nil",
			Value: nil,
			Valid: true,
		},
		{
			Name: "Invalid - no auth provided",
			Value: &graphql.CredentialRequestAuthInput{
				Csrf: nil,
			},
			Valid: false,
		},
		{
			Name: "Invalid - nested validation error",
			Value: &graphql.CredentialRequestAuthInput{
				Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{},
			},
			Valid: false,
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
			if testCase.Valid {
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
		Name  string
		Value *graphql.CredentialDataInput
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: &credential,
			Valid: true,
		},
		{
			Name: "Invalid - no auth provided",
			Value: &graphql.CredentialDataInput{
				Basic: nil,
				Oauth: nil,
			},
			Valid: false,
		},
		{
			Name: "Invalid - multiple auths provided",
			Value: &graphql.CredentialDataInput{
				Basic: &basic,
				Oauth: &oauth,
			},
			Valid: false,
		},
		{
			Name: "Invalid - nested validation error in Basic",
			Value: &graphql.CredentialDataInput{
				Basic: &graphql.BasicCredentialDataInput{},
			},
			Valid: false,
		},
		{
			Name: "Invalid - nested validation error in Oauth",
			Value: &graphql.CredentialDataInput{
				Oauth: &graphql.OAuthCredentialDataInput{},
			},
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBasicCredentialDataInput_Validate_Username(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: "John",
			Valid: true,
		},
		{
			Name:  "Invalid - Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestBasicCredentialDataInput_Validate_Password(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: "John",
			Valid: true,
		},
		{
			Name:  "Invalid - Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestOAuthCredentialDataInput_Validate_ClientID(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: "John.2h2kj2k5gw6j3h5gjk34hg-g:0",
			Valid: true,
		},
		{
			Name:  "Invalid - Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestOAuthCredentialDataInput_Validate_ClientSecret(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: "Doe.2h2kj2k5gw6j3h5gjk34hg-g:0",
			Valid: true,
		},
		{
			Name:  "Invalid - Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestOAuthCredentialDataInput_Validate_URL(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: inputvalidationtest.ValidURL,
			Valid: true,
		},
		{
			Name:  "Invalid - Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "Invalid - Invalid URL",
			Value: inputvalidationtest.InvalidURL,
			Valid: false,
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
			if testCase.Valid {
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
		Name  string
		Value *graphql.CredentialRequestAuthInput
		Valid bool
	}{
		{
			Name: "Valid",
			Value: &graphql.CredentialRequestAuthInput{
				Csrf: &csrf,
			},
			Valid: true,
		},
		{
			Name: "Invalid - no auth provided",
			Value: &graphql.CredentialRequestAuthInput{
				Csrf: nil,
			},
			Valid: false,
		},
		{
			Name: "Invalid - nested validation error in Csrf",
			Value: &graphql.CredentialRequestAuthInput{
				Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{},
			},
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.Valid {
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
		Name  string
		Value *graphql.CredentialDataInput
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: &credential,
			Valid: true,
		},
		{
			Name:  "Invalid - nil",
			Value: nil,
			Valid: false,
		},
		{
			Name: "Invalid - nested validation error",
			Value: &graphql.CredentialDataInput{
				Basic: nil,
				Oauth: nil,
			},
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCSRFTokenCredentialRequestAuthInput_Validate_AdditionalHeaders(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *graphql.HttpHeaders
		Valid bool
	}{
		{
			Name: "Valid",
			Value: &graphql.HttpHeaders{
				"Authorization": {"test", "asdf"},
				"Test":          {"test", "asdf"},
			},
			Valid: true,
		},
		{
			Name:  "Valid - nil",
			Value: nil,
			Valid: true,
		},
		{
			Name: "Invalid - empty key",
			Value: &graphql.HttpHeaders{
				inputvalidationtest.EmptyString: {"test"},
			},
			Valid: false,
		},
		{
			Name: "Invalid - nil value",
			Value: &graphql.HttpHeaders{
				"test": nil,
			},
			Valid: false,
		},
		{
			Name: "Invalid - empty slice element",
			Value: &graphql.HttpHeaders{
				"test": {inputvalidationtest.EmptyString},
			},
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCSRFTokenCredentialRequestAuthInput_Validate_AdditionalQueryParams(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *graphql.QueryParams
		Valid bool
	}{
		{
			Name: "Valid",
			Value: &graphql.QueryParams{
				"Param": {"test", "asdf"},
				"Test":  {"test", "asdf"},
			},
			Valid: true,
		},
		{
			Name:  "Valid - nil",
			Value: nil,
			Valid: true,
		},
		{
			Name: "Invalid - empty key",
			Value: &graphql.QueryParams{
				inputvalidationtest.EmptyString: {"test"},
			},
			Valid: false,
		},
		{
			Name: "Invalid - nil value",
			Value: &graphql.QueryParams{
				"test": nil,
			},
			Valid: false,
		},
		{
			Name: "Invalid - empty slice element",
			Value: &graphql.QueryParams{
				"test": {inputvalidationtest.EmptyString},
			},
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCSRFTokenCredentialRequestAuthInput_Validate_TokenEndpointURL(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: inputvalidationtest.ValidURL,
			Valid: true,
		},
		{
			Name:  "Invalid - Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "Invalid - Invalid URL",
			Value: inputvalidationtest.InvalidURL,
			Valid: false,
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
			if testCase.Valid {
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
