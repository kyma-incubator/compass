package model

import (
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageInstanceAuthRequestInput_ToPackageInstanceAuth(t *testing.T) {
	// GIVEN
	timestamp := time.Now()
	testID := "foo"
	testPackageID := "bar"
	testTenant := "baz"

	input := PackageInstanceAuthRequestInput{
		Context:     str.Ptr(`"test"`),
		InputParams: str.Ptr(`"test"`),
	}
	inputStatus := PackageInstanceAuthStatus{
		Condition: PackageInstanceAuthStatusConditionPending,
		Timestamp: timestamp,
		Message:   str.Ptr("Credentials were not yet provided."),
		Reason:    str.Ptr("CredentialsNotProvided"),
	}
	inputAuth := Auth{
		Credential: CredentialData{
			Basic: &BasicCredentialData{
				Username: "foo",
				Password: "bar",
			},
		},
	}

	expected := PackageInstanceAuth{
		ID:          testID,
		PackageID:   testPackageID,
		Tenant:      testTenant,
		Context:     str.Ptr(`"test"`),
		InputParams: str.Ptr(`"test"`),
		Auth:        &inputAuth,
		Status:      &inputStatus,
	}
	result := input.ToPackageInstanceAuth(testID, testPackageID, testTenant, &inputAuth, inputStatus)
	// THEN
	require.Equal(t, expected, result)

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

func TestPackageInstanceAuth_SetDefaultStatus(t *testing.T) {
	// GIVEN
	timestamp := time.Now()

	testCases := []struct {
		Name            string
		InputCondition  PackageInstanceAuthStatusCondition
		ExpectedReason  string
		ExpectedMessage string
		ExpectedError   error
	}{
		{
			Name:            "Success when succeeded",
			InputCondition:  PackageInstanceAuthStatusConditionSucceeded,
			ExpectedReason:  "CredentialsProvided",
			ExpectedMessage: "Credentials were provided.",
			ExpectedError:   nil,
		},
		{
			Name:            "Success when pending",
			InputCondition:  PackageInstanceAuthStatusConditionPending,
			ExpectedReason:  "CredentialsNotProvided",
			ExpectedMessage: "Credentials were not yet provided.",
			ExpectedError:   nil,
		},
		{
			Name:            "Success when unused",
			InputCondition:  PackageInstanceAuthStatusConditionUnused,
			ExpectedReason:  "PendingDeletion",
			ExpectedMessage: "Credentials for given Package Instance Auth are ready for being deleted by Application or Integration System.",
			ExpectedError:   nil,
		},
		{
			Name:           "Error when unknown status condition",
			InputCondition: "INVALID",
			ExpectedError:  errors.New("invalid status condition"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			instanceAuth := PackageInstanceAuth{}

			// WHEN
			err := instanceAuth.SetDefaultStatus(testCase.InputCondition, timestamp)

			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				assert.Equal(t, &PackageInstanceAuthStatus{
					Condition: testCase.InputCondition,
					Timestamp: timestamp,
					Message:   &testCase.ExpectedMessage,
					Reason:    &testCase.ExpectedReason,
				}, instanceAuth.Status)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}
		})
	}

	t.Run("Success if nil", func(t *testing.T) {
		var instanceAuth *PackageInstanceAuth

		// WHEN
		err := instanceAuth.SetDefaultStatus(PackageInstanceAuthStatusConditionSucceeded, timestamp)

		// THEN
		require.NoError(t, err)
		assert.Nil(t, instanceAuth)
	})
}
