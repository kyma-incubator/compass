package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/stretchr/testify/require"
)

func TestWebhookInput_Validate_Type(t *testing.T) {
	testCases := []struct {
		Name  string
		Value graphql.ApplicationWebhookType
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: graphql.ApplicationWebhookTypeConfigurationChanged,
			Valid: true,
		},
		{
			Name:  "Invalid - Empty",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "Invalid - Not enum",
			Value: "invalid",
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.Type = testCase.Value
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

func TestWebhookInput_Validate_URL(t *testing.T) {
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
		{
			Name:  "Invalid - Too long",
			Value: inputvalidationtest.URL257Long,
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
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

func TestWebhookInput_Validate_Auth(t *testing.T) {
	auth := fixValidAuthInput()
	testCases := []struct {
		Name  string
		Value *graphql.AuthInput
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: &auth,
			Valid: true,
		},
		{
			Name:  "Valid - nil",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "Invalid - Nested validation error",
			Value: &graphql.AuthInput{},
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidWebhookInput()
			sut.Auth = testCase.Value
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

func fixValidWebhookInput() graphql.WebhookInput {
	return graphql.WebhookInput{
		Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		URL:  inputvalidationtest.ValidURL,
	}
}
