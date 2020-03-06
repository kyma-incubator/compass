package provisioning

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:generate mockery -name=ProvisionInputCreator -dir=../../ -output=automock -outpkg=automock -case=underscore

func TestServiceManagerOverridesStepSuccess(t *testing.T) {
	ts := SMOverrideTestSuite{}

	tests := map[string]struct {
		requestParams        string
		overrideParams       ServiceManagerOverrideConfig
		expCredentialsValues []*gqlschema.ConfigEntryInput
	}{
		"always apply override for Service Manager credentials": {
			requestParams:  ts.SMRequestParameters("req-url", "req-user", "req-pass"),
			overrideParams: ts.SMOverrideConfig(SMOverrideModeAlways, "over-url", "over-user", "over-pass"),

			expCredentialsValues: []*gqlschema.ConfigEntryInput{
				{Key: "config.sm.url", Value: "over-url"},
				{Key: "sm.user", Value: "over-user"},
				{Key: "sm.password", Value: "over-pass", Secret: ptr.Bool(true)},
			},
		},
		"never apply override for Service Manager credentials": {
			requestParams:  ts.SMRequestParameters("req-url", "req-user", "req-pass"),
			overrideParams: ts.SMOverrideConfig(SMOverrideModeNever, "over-url", "over-user", "over-pass"),

			expCredentialsValues: []*gqlschema.ConfigEntryInput{
				{Key: "config.sm.url", Value: "req-url"},
				{Key: "sm.user", Value: "req-user"},
				{Key: "sm.password", Value: "req-pass", Secret: ptr.Bool(true)},
			},
		},
		"apply override for Service Manager credentials because they are not present in request": {
			requestParams:  "{}",
			overrideParams: ts.SMOverrideConfig(SMOverrideModeWhenNotSendInRequest, "over-url", "over-user", "over-pass"),

			expCredentialsValues: []*gqlschema.ConfigEntryInput{
				{Key: "config.sm.url", Value: "over-url"},
				{Key: "sm.user", Value: "over-user"},
				{Key: "sm.password", Value: "over-pass", Secret: ptr.Bool(true)},
			},
		},
		"do not apply override for Service Manager credentials because they are present in request": {
			requestParams:  ts.SMRequestParameters("req-url", "req-user", "req-pass"),
			overrideParams: ts.SMOverrideConfig(SMOverrideModeWhenNotSendInRequest, "over-url", "over-user", "over-pass"),

			expCredentialsValues: []*gqlschema.ConfigEntryInput{
				{Key: "config.sm.url", Value: "req-url"},
				{Key: "sm.user", Value: "req-user"},
				{Key: "sm.password", Value: "req-pass", Secret: ptr.Bool(true)},
			},
		},
	}
	for tN, tC := range tests {
		t.Run(tN, func(t *testing.T) {
			// given
			inputCreatorMock := &automock.ProvisionInputCreator{}
			inputCreatorMock.On("SetOverrides", "service-manager-proxy", tC.expCredentialsValues).
				Return(nil).Once()

			operation := internal.ProvisioningOperation{
				ProvisioningParameters: tC.requestParams,
				InputCreator:           inputCreatorMock,
			}

			memoryStorage := storage.NewMemoryStorage()
			smStep := NewServiceManagerOverridesStep(memoryStorage.Operations(), tC.overrideParams)

			// when
			gotOperation, retryTime, err := smStep.Run(operation, NewLogDummy())

			// then
			require.NoError(t, err)

			assert.Zero(t, retryTime)
			assert.Equal(t, operation, gotOperation)
			inputCreatorMock.AssertExpectations(t)
		})
	}
}

func TestServiceManagerOverridesStepError(t *testing.T) {
	tests := map[string]struct {
		givenReqParams string
		expErr         string
	}{
		"return error when creds in request are not provided and overrides should not be applied": {
			givenReqParams: "{}",
			expErr:         "Service Manager Credentials are required to be send in provisioning request.",
		},
		"return retry type instead of error when not able to get provisioning parameters": {
			givenReqParams: "{malformed params..",

			expErr: "invalid operation provisioning parameters",
		},
	}
	for tN, tC := range tests {
		t.Run(tN, func(t *testing.T) {
			// given
			operation := internal.ProvisioningOperation{
				Operation:              internal.Operation{ID: "123"},
				ProvisioningParameters: tC.givenReqParams,
			}

			memoryStorage := storage.NewMemoryStorage()
			require.NoError(t, memoryStorage.Operations().InsertProvisioningOperation(operation))
			smStep := NewServiceManagerOverridesStep(memoryStorage.Operations(), ServiceManagerOverrideConfig{})

			// when
			gotOperation, retryTime, err := smStep.Run(operation, NewLogDummy())

			// then
			require.EqualError(t, err, tC.expErr)
			assert.Zero(t, retryTime)
			assert.Equal(t, domain.Failed, gotOperation.State)
		})
	}
}

type SMOverrideTestSuite struct{}

func (SMOverrideTestSuite) SMRequestParameters(smURL, smUser, smPass string) string {
	return fmt.Sprintf(`{
		"ers_context": {
		  "sm_platform_credentials": {
		    "url": "%s",
			"credentials": {
			  "basic": {
				"username": "%s",
				"password": "%s"
			  }
			}
		  }
		}}`, smURL, smUser, smPass)
}

func (s SMOverrideTestSuite) SMOverrideConfig(mode ServiceManagerOverrideMode, url string, user string, pass string) ServiceManagerOverrideConfig {
	return ServiceManagerOverrideConfig{
		OverrideMode: mode,
		URL:          url,
		Username:     user,
		Password:     pass,
	}
}

// NewLogDummy returns dummy logger which discards logged messages on the fly.
// Useful when logger is required as dependency in unit testing.
func NewLogDummy() *logrus.Entry {
	rawLgr := logrus.New()
	rawLgr.Out = ioutil.Discard
	lgr := rawLgr.WithField("testing", true)

	return lgr
}
