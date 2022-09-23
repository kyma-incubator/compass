package subject_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/hydrator/internal/subject"

	"github.com/stretchr/testify/require"
)

const (
	configTpl = `[{"consumer_type": "%s", "tenant_access_levels": ["%s"], "subject": "%s", "internal_consumer_id": "%s"}]`

	validConsumer           = "Integration System"
	validAccessLvl          = "account"
	validSubject            = "C=DE, OU=Compass Clients, OU=ed1f789b-1a85-4a63-b360-fac9d6484544, L=validate, CN=test-compass-integration"
	validInternalConsumerID = "3bfbb60f-d67d-4657-8f9e-2d73a6b24a10"

	invalidValue = "test"
)

var validConfig = fmt.Sprintf(configTpl, validConsumer, validAccessLvl, validSubject, validInternalConsumerID)

func TestNewProcessor(t *testing.T) {
	testCases := []struct {
		name             string
		config           string
		ouPattern        string
		ouRegionPattern  string
		expectedErrorMsg string
	}{
		{
			name:   "Success",
			config: validConfig,
		},
		{
			name:             "Returns error when configuration with invalid format is provided",
			config:           invalidValue,
			expectedErrorMsg: "while unmarshalling mappings",
		},
		{
			name:             "Returns error when configuration contains invalid consumer type",
			config:           fmt.Sprintf(configTpl, invalidValue, validAccessLvl, validSubject, validInternalConsumerID),
			expectedErrorMsg: "consumer type test is not valid",
		},
		{
			name:             "Returns error when configuration contains invalid access level",
			config:           fmt.Sprintf(configTpl, validConsumer, invalidValue, validSubject, validInternalConsumerID),
			expectedErrorMsg: fmt.Sprintf("tenant access level %s is not valid", invalidValue),
		},
		{
			name:             "Returns error when subject is not provided in configuration",
			config:           fmt.Sprintf(configTpl, validConsumer, validAccessLvl, "", validInternalConsumerID),
			expectedErrorMsg: "subject is not provided",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			_, err := subject.NewProcessor(test.config, test.ouPattern, test.ouRegionPattern)
			if len(test.expectedErrorMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedErrorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuthIDFromSubjectFunc(t *testing.T) {
	t.Run("Success when internal consumer id is provided", func(t *testing.T) {
		p, err := subject.NewProcessor(validConfig, "", "")
		require.NoError(t, err)

		res := p.AuthIDFromSubjectFunc()(validSubject)
		require.Equal(t, validInternalConsumerID, res)
	})
	t.Run("Success when internal consumer id is not provided", func(t *testing.T) {
		config := fmt.Sprintf(configTpl, validConsumer, validAccessLvl, validSubject, "")
		p, err := subject.NewProcessor(config, "Compass Clients", "")
		require.NoError(t, err)

		res := p.AuthIDFromSubjectFunc()(validSubject)
		require.Equal(t, "ed1f789b-1a85-4a63-b360-fac9d6484544", res)
	})
	t.Run("Success getting authID from mapping", func(t *testing.T) {
		p, err := subject.NewProcessor("[]", "Compass Clients", "")
		require.NoError(t, err)

		res := p.AuthIDFromSubjectFunc()(validSubject)
		require.Equal(t, "ed1f789b-1a85-4a63-b360-fac9d6484544", res)
	})
	t.Run("Success getting authID from OUs", func(t *testing.T) {
		p, err := subject.NewProcessor(fmt.Sprintf(configTpl, validConsumer, validAccessLvl, "OU=Random OU", validInternalConsumerID), "Compass Clients", "")
		require.NoError(t, err)

		res := p.AuthIDFromSubjectFunc()(validSubject)
		require.Equal(t, "ed1f789b-1a85-4a63-b360-fac9d6484544", res)
	})
}

func TestAuthSessionExtraFromSubjectFunc(t *testing.T) {
	ctx := context.Background()

	t.Run("Success getting auth session extra", func(t *testing.T) {
		p, err := subject.NewProcessor(validConfig, "", "")
		require.NoError(t, err)

		extra := p.AuthSessionExtraFromSubjectFunc()(ctx, validSubject)
		require.Equal(t, validConsumer, extra["consumer_type"])
		require.Equal(t, []string{validAccessLvl}, extra["tenant_access_levels"])
		require.Equal(t, validInternalConsumerID, extra["internal_consumer_id"])
	})
	t.Run("Returns nil when can't match subjects components", func(t *testing.T) {
		invalidSubject := "C=DE, OU=Compass Clients, OU=Random OU, L=validate, CN=test-compass-integration"
		p, err := subject.NewProcessor(validConfig, "", "")
		require.NoError(t, err)

		extra := p.AuthSessionExtraFromSubjectFunc()(ctx, invalidSubject)
		require.Nil(t, extra)
	})
	t.Run("Returns nil when can't match number of subjects components", func(t *testing.T) {
		invalidSubject := "C=DE, OU=Compass Clients, L=validate, CN=test-compass-integration"
		p, err := subject.NewProcessor(validConfig, "", "")
		require.NoError(t, err)

		extra := p.AuthSessionExtraFromSubjectFunc()(ctx, invalidSubject)
		require.Nil(t, extra)
	})
}
