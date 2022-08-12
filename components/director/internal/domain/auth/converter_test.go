package auth_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	emptyCertCommonName := ""
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *model.Auth
		Expected *graphql.Auth
		Error    error
	}{
		{
			Name:     "All properties given for basic auth",
			Input:    fixDetailedAuthBasicCredentials(),
			Expected: fixDetailedBasicCredentialsGQLAuth(),
		},
		{
			Name:     "All properties given for certificate oauth auth",
			Input:    fixDetailedAuthCertificateOAuthCredentials(),
			Expected: fixDetailedCertificateOAuthCredentialsGQLAuth(),
		},
		{
			Name:     "All properties given for oauth auth",
			Input:    fixDetailedOAuthCredentials(),
			Expected: fixDetailedOAuthCredentialsGQLAuth(),
		},
		{
			Name:  "Empty",
			Input: &model.Auth{},
			Expected: &graphql.Auth{
				CertCommonName: &emptyCertCommonName,
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
			// WHEN
			converter := auth.NewConverter()
			res, err := converter.ToGraphQL(testCase.Input)

			// then
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *graphql.AuthInput
		Expected *model.AuthInput
	}{
		{
			Name:     "All properties given for basic oauth",
			Input:    fixDetailedBasicCredentialGQLAuthInput(),
			Expected: fixDetailedBasicCredentialAuthInput(),
		},
		{
			Name:     "All properties given for certificate oauth auth",
			Input:    fixDetailedCertificateOAuthGQLAuthInput(),
			Expected: fixDetailedCertificateOAuthAuthInput(),
		},
		{
			Name:     "All properties given - deprecated",
			Input:    fixDetailedGQLAuthInputDeprecated(),
			Expected: fixDetailedBasicCredentialAuthInput(),
		},
		{
			Name:     "Empty",
			Input:    &graphql.AuthInput{},
			Expected: &model.AuthInput{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			converter := auth.NewConverter()
			res, err := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_ModelFromGraphQLInput(t *testing.T) {
	// GIVEN
	gqlAuthInputWithInvalidHeaders := fixDetailedBasicCredentialGQLAuthInput()
	gqlAuthInputWithInvalidHeaders.AdditionalHeadersSerialized = &invalidAuthHeadersSerialized

	gqlAuthInputWithInvalidQueryParams := fixDetailedBasicCredentialGQLAuthInput()
	gqlAuthInputWithInvalidQueryParams.AdditionalQueryParamsSerialized = &invalidAuthParamsSerialized

	testCases := []struct {
		Name             string
		Input            graphql.AuthInput
		Expected         *model.Auth
		ExpectedErrorMsg string
	}{
		{
			Name:     "All properties given for basic auth",
			Input:    *fixDetailedBasicCredentialGQLAuthInput(),
			Expected: fixDetailedAuthBasicCredentials(),
		},
		{
			Name:     "All properties given for certificate oauth auth",
			Input:    *fixDetailedCertificateOAuthGQLAuthInput(),
			Expected: fixDetailedAuthCertificateOAuthCredentials(),
		},
		{
			Name:     "Empty",
			Input:    graphql.AuthInput{},
			Expected: &model.Auth{},
		},
		{
			Name:             "Error when HTTP serialized headers are invalid",
			Input:            *gqlAuthInputWithInvalidHeaders,
			Expected:         nil,
			ExpectedErrorMsg: "unable to unmarshal HTTP headers",
		},
		{
			Name:             "Error when query params are invalid",
			Input:            *gqlAuthInputWithInvalidQueryParams,
			Expected:         nil,
			ExpectedErrorMsg: "unable to unmarshal query parameters",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			converter := auth.NewConverter()
			res, err := converter.ModelFromGraphQLInput(testCase.Input)

			// then
			if testCase.ExpectedErrorMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
				assert.Equal(t, testCase.Expected, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.Expected, res)
			}
		})
	}
}

func TestConverter_ModelFromGraphQLTokenInput(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *graphql.OneTimeTokenInput
		Expected *model.OneTimeToken
	}{
		{
			Name:     "All properties given",
			Input:    fixOneTimeTokenGQLInput(),
			Expected: fixOneTimeTokenInput(),
		},
		{
			Name:     "Empty",
			Input:    &graphql.OneTimeTokenInput{},
			Expected: &model.OneTimeToken{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			converter := auth.NewConverter()
			res := converter.ModelFromGraphQLTokenInput(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
