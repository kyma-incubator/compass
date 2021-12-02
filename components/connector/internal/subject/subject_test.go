package subject_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/subject"
	"github.com/stretchr/testify/require"
)

const (
	configTpl = `[{"consumer_type": "%s", "tenant_access_level": "%s", "subject": "%s"}]`

	validConsumer  = "Integration System"
	validAccessLvl = "account"
	validSubject   = "C=DE, OU=Compass Clients, OU=ed1f789b-1a85-4a63-b360-fac9d6484544, L=validate, CN=test-compass-integration"
	invalidValue   = "test"
)

var validConfig = fmt.Sprintf(configTpl, validConsumer, validAccessLvl, validSubject)

func TestNewProcessor(t *testing.T) {

	t.Run("returns error when configuration with invalid format is provided", func(t *testing.T) {
		_, err := subject.NewProcessor("invalid-config", "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "while unmarshalling mappings")
	})
	t.Run("returns error when configuration contains invalid consumer type", func(t *testing.T) {
		config := `[{"consumer_type": "Test", "tenant_access_level": "account", "subject": "C=DE, OU=Compass Clients, OU=ed1f789b-1a85-4a63-b360-fac9d6484544, L=validate, CN=test-compass-integration"}]`
		_, err := subject.NewProcessor(config, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "consumer type Test is not valid")
	})
	t.Run("returns error when configuration contains invalid consumer type", func(t *testing.T) {
		config := `[{"consumer_type": "Integration System", "tenant_access_level": "test", "subject": "C=DE, OU=Compass Clients, OU=ed1f789b-1a85-4a63-b360-fac9d6484544, L=validate, CN=test-compass-integration"}]`
		_, err := subject.NewProcessor(config, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "tenant access level is not valid")
	})
	t.Run("returns error when configuration contains invalid consumer type", func(t *testing.T) {
		config := `[{"consumer_type": "Integration System", "tenant_access_level": "account", "subject": ""}]`
		_, err := subject.NewProcessor(config, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "subject is not provided")
	})

	testCases := []struct {
		name             string
		config           string
		ouPattern        string
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
			name:            "Returns error when configuration contains invalid consumer type",
			config:           fmt.Sprintf(configTpl, invalidValue, validAccessLvl, validSubject),
			expectedErrorMsg: "consumer type test is not valid",
		},
		{
			name:            "Returns error when configuration contains invalid consumer type",
			config:           fmt.Sprintf(configTpl, invalidValue, validAccessLvl, validSubject),
			expectedErrorMsg: "while unmarshalling mappings",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			_, err := subject.NewProcessor(test.config, test.ouPattern)
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

}

func TestAuthSessionExtraFromSubjectFunc(t *testing.T) {

}
