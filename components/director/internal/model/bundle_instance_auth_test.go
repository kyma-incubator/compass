package model

import (
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundleInstanceAuthRequestInput_ToBundleInstanceAuth(t *testing.T) {
	// GIVEN
	timestamp := time.Now()
	testID := "foo"
	testBundleID := "bar"
	testTenant := "baz"

	input := BundleInstanceAuthRequestInput{
		ID:          str.Ptr(`"foo"'`),
		Context:     str.Ptr(`"test"`),
		InputParams: str.Ptr(`"test"`),
	}
	inputStatus := BundleInstanceAuthStatus{
		Condition: BundleInstanceAuthStatusConditionPending,
		Timestamp: timestamp,
		Message:   "Credentials were not yet provided.",
		Reason:    "CredentialsNotProvided",
	}
	inputAuth := Auth{
		Credential: CredentialData{
			Basic: &BasicCredentialData{
				Username: "foo",
				Password: "bar",
			},
		},
	}

	expected := BundleInstanceAuth{
		ID:          testID,
		BundleID:    testBundleID,
		Tenant:      testTenant,
		Context:     str.Ptr(`"test"`),
		InputParams: str.Ptr(`"test"`),
		Auth:        &inputAuth,
		Status:      &inputStatus,
	}
	result := input.ToBundleInstanceAuth(testID, testBundleID, testTenant, &inputAuth, &inputStatus)
	// THEN
	require.Equal(t, expected, result)

}

func TestBundleInstanceAuthStatusInput_ToBundleInstanceAuthStatus(t *testing.T) {
	// GIVEN
	timestamp := time.Now()

	testCases := []struct {
		Name     string
		Input    *BundleInstanceAuthStatusInput
		Expected *BundleInstanceAuthStatus
	}{
		{
			Name:     "Success when nil",
			Input:    nil,
			Expected: nil,
		},
		{
			Name: "Success",
			Input: &BundleInstanceAuthStatusInput{
				Condition: BundleInstanceAuthSetStatusConditionInputSucceeded,
				Message:   "foo",
				Reason:    "bar",
			},
			Expected: &BundleInstanceAuthStatus{
				Condition: BundleInstanceAuthStatusConditionSucceeded,
				Timestamp: timestamp,
				Message:   "foo",
				Reason:    "bar",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToBundleInstanceAuthStatus(timestamp)
			// THEN
			require.Equal(t, testCase.Expected, result)
		})
	}
}

func TestBundleInstanceAuth_SetDefaultStatus(t *testing.T) {
	// GIVEN
	timestamp := time.Now()

	testCases := []struct {
		Name            string
		InputCondition  BundleInstanceAuthStatusCondition
		ExpectedReason  string
		ExpectedMessage string
		ExpectedError   error
	}{
		{
			Name:            "Success when succeeded",
			InputCondition:  BundleInstanceAuthStatusConditionSucceeded,
			ExpectedReason:  "CredentialsProvided",
			ExpectedMessage: "Credentials were provided.",
			ExpectedError:   nil,
		},
		{
			Name:            "Success when pending",
			InputCondition:  BundleInstanceAuthStatusConditionPending,
			ExpectedReason:  "CredentialsNotProvided",
			ExpectedMessage: "Credentials were not yet provided.",
			ExpectedError:   nil,
		},
		{
			Name:            "Success when unused",
			InputCondition:  BundleInstanceAuthStatusConditionUnused,
			ExpectedReason:  "PendingDeletion",
			ExpectedMessage: "Credentials for given Bundle Instance Auth are ready for being deleted by Application or Integration System.",
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
			instanceAuth := BundleInstanceAuth{}

			// WHEN
			err := instanceAuth.SetDefaultStatus(testCase.InputCondition, timestamp)

			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				assert.Equal(t, &BundleInstanceAuthStatus{
					Condition: testCase.InputCondition,
					Timestamp: timestamp,
					Message:   testCase.ExpectedMessage,
					Reason:    testCase.ExpectedReason,
				}, instanceAuth.Status)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}
		})
	}

	t.Run("Success if nil", func(t *testing.T) {
		var instanceAuth *BundleInstanceAuth

		// WHEN
		err := instanceAuth.SetDefaultStatus(BundleInstanceAuthStatusConditionSucceeded, timestamp)

		// THEN
		require.NoError(t, err)
		assert.Nil(t, instanceAuth)
	})
}
