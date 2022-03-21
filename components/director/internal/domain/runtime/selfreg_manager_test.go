package runtime_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/rtmtest"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

const (
	selfRegisterDistinguishLabelKey = "test-distinguish-label-key"
	distinguishLblVal               = "test-value"
	testUUID                        = "b3ea1977-582e-4d61-ae12-b3a837a3858e"
	testRegion                      = "test-region-2"
	fakeRegion                      = "fake-region"
	regionLabelKey                  = "region"
)

var testConfig = runtime.SelfRegConfig{
	SelfRegisterDistinguishLabelKey: selfRegisterDistinguishLabelKey,
	SelfRegisterLabelKey:            "test-label-key",
	SelfRegisterLabelValuePrefix:    "test-prefix",
	SelfRegisterResponseKey:         rtmtest.ResponseLabelKey,
	SelfRegisterPath:                "test-path",
	SelfRegisterNameQueryParam:      "testNameQuery",
	SelfRegisterTenantQueryParam:    "testTenantQuery",
	SelfRegisterRequestBodyPattern:  `{"%s":"test"}`,
	RegionToInstanceConfig: map[string]runtime.InstanceConfig{
		"test-region": {
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			URL:          "https://test-url.com",
		},
		testRegion: {
			ClientID:     "test-client-id-2",
			ClientSecret: "test-client-secret-2",
			URL:          "https://test-url-second.com",
		},
	},

	ClientTimeout: 5 * time.Second,
}

func TestSelfRegisterManager_PrepareRuntimeForSelfRegistration(t *testing.T) {
	tokenConsumer := consumer.Consumer{
		ConsumerID: "test-consumer-id",
		Flow:       oathkeeper.OAuth2Flow,
	}
	certConsumer := consumer.Consumer{
		ConsumerID: "test-consumer-id",
		Flow:       oathkeeper.CertificateFlow,
	}
	lblInput := model.RuntimeInput{
		Labels: graphql.Labels{selfRegisterDistinguishLabelKey: distinguishLblVal, regionLabelKey: testRegion},
	}
	lblInputWithoutRegion := model.RuntimeInput{
		Labels: graphql.Labels{selfRegisterDistinguishLabelKey: distinguishLblVal},
	}
	lblInputAfterPrep := map[string]interface{}{
		testConfig.SelfRegisterLabelKey: rtmtest.ResponseLabelValue,
	}
	emptyLabels := make(map[string]interface{})

	fakeConfig := testConfig
	fakeConfig.RegionToInstanceConfig[fakeRegion] = runtime.InstanceConfig{URL: "https://test-url    .com"}

	ctxWithTokenConsumer := consumer.SaveToContext(context.TODO(), tokenConsumer)
	ctxWithCertConsumer := consumer.SaveToContext(context.TODO(), certConsumer)

	testCases := []struct {
		Name           string
		Config         runtime.SelfRegConfig
		CallerProvider func(*testing.T, runtime.SelfRegConfig, string) *automock.ExternalSvcCallerProvider
		Region         string
		Input          model.RuntimeInput
		Context        context.Context
		ExpectedErr    error
		ExpectedOutput map[string]interface{}
	}{
		{
			Name:           "Success",
			Config:         testConfig,
			Input:          lblInput,
			CallerProvider: rtmtest.CallerThatGetsCalledOnce(http.StatusCreated),
			Region:         testRegion,
			Context:        ctxWithCertConsumer,
			ExpectedErr:    nil,
			ExpectedOutput: lblInputAfterPrep,
		},
		{
			Name:           "Success for non-matching consumer",
			Config:         testConfig,
			CallerProvider: rtmtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			Input:          model.RuntimeInput{Labels: graphql.Labels{regionLabelKey: testRegion}},
			Context:        ctxWithTokenConsumer,
			ExpectedErr:    nil,
			ExpectedOutput: emptyLabels,
		},
		{
			Name:           "Error when region label is missing",
			Config:         testConfig,
			CallerProvider: rtmtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			Input:          lblInputWithoutRegion,
			Context:        ctxWithCertConsumer,
			ExpectedErr:    fmt.Errorf("missing %q label", regionLabelKey),
			ExpectedOutput: emptyLabels,
		},
		{
			Name:           "Error when region doesn't exist",
			Config:         testConfig,
			CallerProvider: rtmtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			Input:          model.RuntimeInput{Labels: graphql.Labels{selfRegisterDistinguishLabelKey: distinguishLblVal, regionLabelKey: "not-valid"}},
			Context:        ctxWithCertConsumer,
			ExpectedErr:    errors.New("missing configuration for region"),
			ExpectedOutput: emptyLabels,
		},
		{
			Name:           "Error when caller provider fails",
			Config:         testConfig,
			CallerProvider: rtmtest.CallerProviderThatFails,
			Region:         testRegion,
			Input:          lblInput,
			Context:        ctxWithCertConsumer,
			ExpectedErr:    errors.New("while getting caller"),
			ExpectedOutput: emptyLabels,
		},
		{
			Name:           "Error when context does not contain consumer",
			Config:         testConfig,
			CallerProvider: rtmtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			Input:          model.RuntimeInput{},
			Context:        context.TODO(),
			ExpectedErr:    consumer.NoConsumerError,
			ExpectedOutput: emptyLabels,
		},
		{
			Name:           "Error when can't create URL for preparation of self-registered runtime",
			Config:         fakeConfig,
			CallerProvider: rtmtest.CallerThatDoesNotGetCalled,
			Region:         fakeRegion,
			Input:          model.RuntimeInput{Labels: graphql.Labels{selfRegisterDistinguishLabelKey: "invalid value", regionLabelKey: fakeRegion}},
			Context:        ctxWithCertConsumer,
			ExpectedErr:    errors.New("while creating url for preparation of self-registered runtime"),
			ExpectedOutput: emptyLabels,
		},
		{
			Name:           "Error when Call doesn't succeed",
			Config:         testConfig,
			CallerProvider: rtmtest.CallerThatDoesNotSucceed,
			Region:         testRegion,
			Input:          lblInput,
			Context:        ctxWithCertConsumer,
			ExpectedErr:    rtmtest.TestError,
			ExpectedOutput: emptyLabels,
		},
		{
			Name:           "Error when status code is unexpected",
			Config:         testConfig,
			CallerProvider: rtmtest.CallerThatReturnsBadStatus,
			Region:         testRegion,
			Input:          lblInput,
			Context:        ctxWithCertConsumer,
			ExpectedErr:    errors.New("received unexpected status"),
			ExpectedOutput: emptyLabels,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svcCallerProvider := testCase.CallerProvider(t, testCase.Config, testCase.Region)
			manager := runtime.NewSelfRegisterManager(testCase.Config, svcCallerProvider)

			output, err := manager.PrepareRuntimeForSelfRegistration(testCase.Context, testCase.Input, testUUID)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, testCase.ExpectedOutput, output)

			svcCallerProvider.AssertExpectations(t)
		})
	}
}

func TestSelfRegisterManager_CleanupSelfRegisteredRuntime(t *testing.T) {
	ctx := context.TODO()
	fakeConfig := testConfig
	fakeConfig.RegionToInstanceConfig[fakeRegion] = runtime.InstanceConfig{URL: "https://test-url    .com"}

	testCases := []struct {
		Name                                string
		Config                              runtime.SelfRegConfig
		CallerProvider                      func(*testing.T, runtime.SelfRegConfig, string) *automock.ExternalSvcCallerProvider
		Region                              string
		SelfRegisteredDistinguishLabelValue string
		Context                             context.Context
		ExpectedErr                         error
	}{
		{
			Name:                                "Success",
			CallerProvider:                      rtmtest.CallerThatGetsCalledOnce(http.StatusOK),
			Region:                              testRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctx,
			ExpectedErr:                         nil,
		},
		{
			Name:                                "Success when runtime is not self-registered",
			CallerProvider:                      rtmtest.CallerThatDoesNotGetCalled,
			Region:                              testRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: "",
			Context:                             ctx,
			ExpectedErr:                         nil,
		},
		{
			Name:                                "Error when can't create URL for cleanup of self-registered runtime",
			CallerProvider:                      rtmtest.CallerThatDoesNotGetCalled,
			Region:                              fakeRegion,
			Config:                              fakeConfig,
			SelfRegisteredDistinguishLabelValue: "invalid value",
			Context:                             ctx,
			ExpectedErr:                         errors.New("while creating url for cleanup of self-registered runtime"),
		},
		{
			Name:                                "Error when region doesn't exist",
			Config:                              testConfig,
			CallerProvider:                      rtmtest.CallerThatDoesNotGetCalled,
			Region:                              "not-valid",
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctx,
			ExpectedErr:                         errors.New("missing configuration for region"),
		},
		{
			Name:                                "Error when caller provider fails",
			Config:                              testConfig,
			CallerProvider:                      rtmtest.CallerProviderThatFails,
			Region:                              testRegion,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctx,
			ExpectedErr:                         errors.New("while getting caller"),
		},
		{
			Name:                                "Error when Call doesn't succeed",
			CallerProvider:                      rtmtest.CallerThatDoesNotSucceed,
			Region:                              testRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctx,
			ExpectedErr:                         rtmtest.TestError,
		},
		{
			Name:                                "Error when Call doesn't succeed",
			CallerProvider:                      rtmtest.CallerThatReturnsBadStatus,
			Region:                              testRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctx,
			ExpectedErr:                         errors.New("received unexpected status code"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svcCallerProvider := testCase.CallerProvider(t, testCase.Config, testCase.Region)
			manager := runtime.NewSelfRegisterManager(testCase.Config, svcCallerProvider)

			err := manager.CleanupSelfRegisteredRuntime(testCase.Context, testCase.SelfRegisteredDistinguishLabelValue, testCase.Region)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
