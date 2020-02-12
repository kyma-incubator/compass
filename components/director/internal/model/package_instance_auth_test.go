package model

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/stretchr/testify/require"
)

func TestPackageInstanceAuthRequestInput_ToPackageInstanceAuth(t *testing.T) {
	// GIVEN
	timestamp := time.Now()
	testID := "foo"
	testPackageID := "bar"
	testTenant := "baz"

	testCases := []struct {
		Name      string
		Input     PackageInstanceAuthRequestInput
		InputAuth *Auth
		Expected  PackageInstanceAuth
	}{
		{
			Name: "Success when auth provided",
			Input: PackageInstanceAuthRequestInput{
				Context:     str.Ptr(`"test"`),
				InputParams: str.Ptr(`"test"`),
			},
			InputAuth: &Auth{
				Credential: CredentialData{
					Basic: &BasicCredentialData{
						Username: "foo",
						Password: "bar",
					},
				},
			},
			Expected: PackageInstanceAuth{
				ID:          testID,
				PackageID:   testPackageID,
				Tenant:      testTenant,
				Context:     str.Ptr(`"test"`),
				InputParams: str.Ptr(`"test"`),
				Auth: &Auth{
					Credential: CredentialData{
						Basic: &BasicCredentialData{
							Username: "foo",
							Password: "bar",
						},
					},
				},
				Status: &PackageInstanceAuthStatus{
					Condition: PackageInstanceAuthStatusConditionSucceeded,
					Timestamp: timestamp,
					Message:   str.Ptr("Credentials were provided."),
					Reason:    str.Ptr("CredentialsProvided"),
				},
			},
		},
		{
			Name: "Success when auth not provided",
			Input: PackageInstanceAuthRequestInput{
				Context:     str.Ptr(`"test"`),
				InputParams: str.Ptr(`"test"`),
			},
			InputAuth: nil,
			Expected: PackageInstanceAuth{
				ID:          testID,
				PackageID:   testPackageID,
				Tenant:      testTenant,
				Context:     str.Ptr(`"test"`),
				InputParams: str.Ptr(`"test"`),
				Auth:        nil,
				Status: &PackageInstanceAuthStatus{
					Condition: PackageInstanceAuthStatusConditionPending,
					Timestamp: timestamp,
					Message:   str.Ptr("Credentials were not yet provided."),
					Reason:    str.Ptr("CredentialsNotProvided"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToPackageInstanceAuth(testID, testPackageID, testTenant, testCase.InputAuth, timestamp)
			// THEN
			require.Equal(t, testCase.Expected, result)
		})
	}
}

func TestPackageInstanceAuthStatusInput_ToPackageInstanceAuthStatus(t *testing.T) {
	// GIVEN
	timestamp := time.Now()

	testCases := []struct {
		Name     string
		Input    *PackageInstanceAuthStatusInput
		Expected *PackageInstanceAuthStatus
	}{
		{
			Name:     "Success when nil",
			Input:    nil,
			Expected: nil,
		},
		{
			Name: "Success",
			Input: &PackageInstanceAuthStatusInput{
				Condition: PackageInstanceAuthSetStatusConditionInputSucceeded,
				Message:   str.Ptr("foo"),
				Reason:    str.Ptr("bar"),
			},
			Expected: &PackageInstanceAuthStatus{
				Condition: PackageInstanceAuthStatusConditionSucceeded,
				Timestamp: timestamp,
				Message:   str.Ptr("foo"),
				Reason:    str.Ptr("bar"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToPackageInstanceAuthStatus(timestamp)
			// THEN
			require.Equal(t, testCase.Expected, result)
		})
	}
}

func TestPackageInstanceAuth_SetAuth(t *testing.T) {
	// GIVEN
	timestamp := time.Now()

	testPackageInstanceAuth := PackageInstanceAuth{
		ID:          "foo",
		PackageID:   "bar",
		Tenant:      "baz",
		Context:     str.Ptr(`"test"`),
		InputParams: str.Ptr(`"test"`),
		Auth:        nil,
		Status:      nil,
	}

	testSetInput := PackageInstanceAuthSetInput{
		Auth: &AuthInput{
			Credential: &CredentialDataInput{
				Basic: &BasicCredentialDataInput{
					Username: "foo",
					Password: "bar",
				},
			},
		},
		Status: &PackageInstanceAuthStatusInput{
			Condition: PackageInstanceAuthSetStatusConditionInputSucceeded,
			Message:   str.Ptr("foo"),
			Reason:    str.Ptr("bar"),
		},
	}

	testExpectedPackageInstanceAuth := PackageInstanceAuth{
		ID:          "foo",
		PackageID:   "bar",
		Tenant:      "baz",
		Context:     str.Ptr(`"test"`),
		InputParams: str.Ptr(`"test"`),
		Auth: &Auth{
			Credential: CredentialData{
				Basic: &BasicCredentialData{
					Username: "foo",
					Password: "bar",
				},
			},
		},
		Status: &PackageInstanceAuthStatus{
			Condition: PackageInstanceAuthStatusConditionSucceeded,
			Timestamp: timestamp,
			Message:   str.Ptr("foo"),
			Reason:    str.Ptr("bar"),
		},
	}

	testCases := []struct {
		Name           string
		Input          *PackageInstanceAuth
		InputAuth      PackageInstanceAuthSetInput
		ExpectedOutput *PackageInstanceAuth
	}{
		{
			Name:           "Success when nil",
			Input:          nil,
			InputAuth:      PackageInstanceAuthSetInput{},
			ExpectedOutput: nil,
		},
		{
			Name:           "Success",
			Input:          &testPackageInstanceAuth,
			InputAuth:      testSetInput,
			ExpectedOutput: &testExpectedPackageInstanceAuth,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			testCase.Input.SetAuth(testCase.InputAuth, timestamp)
			// THEN
			require.Equal(t, testCase.ExpectedOutput, testCase.Input)
		})
	}
}
