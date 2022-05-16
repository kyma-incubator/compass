package runtime_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/rtmtest"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

const (
	selfRegisterDistinguishLabelKey = "test-distinguish-label-key"
	distinguishLblVal               = "test-value"
	testUUID                        = "b3ea1977-582e-4d61-ae12-b3a837a3858e"
	testRegion                      = "test-region"
	fakeRegion                      = "fake-region"
	regionLabelKey                  = "region"
)

var testConfig = config.SelfRegConfig{
	SelfRegisterDistinguishLabelKey: selfRegisterDistinguishLabelKey,
	SelfRegisterLabelKey:            "test-label-key",
	SelfRegisterLabelValuePrefix:    "test-prefix",
	SelfRegisterResponseKey:         rtmtest.ResponseLabelKey,
	SelfRegisterPath:                "test-path",
	SelfRegisterNameQueryParam:      "testNameQuery",
	SelfRegisterTenantQueryParam:    "testTenantQuery",
	SelfRegisterRequestBodyPattern:  `{"%s":"test"}`,
	SelfRegisterSecretPath:          "testdata/TestSelfRegisterManager_PrepareRuntimeForSelfRegistration.golden",
	InstanceClientIDPath:            "clientId",
	InstanceClientSecretPath:        "clientSecret",
	InstanceURLPath:                 "url",
	InstanceTokenURLPath:            "tokenUrl",
	InstanceCertPath:                "clientCert",
	InstanceKeyPath:                 "clientKey",
	RegionToInstanceConfig: map[string]config.InstanceConfig{
		"test-region": {
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			URL:          "https://test-url-second.com",
			TokenURL:     "https://test-token-url-second.com",
			Cert:         "cert",
			Key:          "key",
		},
		"fake-region": {
			ClientID:     "client_id_2",
			ClientSecret: "client_secret_2",
			URL:          "https://test-url      -second.com",
			TokenURL:     "https://test-token-url-second.com",
			Cert:         "cert2",
			Key:          "key2",
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

	ctxWithTokenConsumer := consumer.SaveToContext(context.TODO(), tokenConsumer)
	ctxWithCertConsumer := consumer.SaveToContext(context.TODO(), certConsumer)

	testCases := []struct {
		Name           string
		Config         config.SelfRegConfig
		CallerProvider func(*testing.T, config.SelfRegConfig, string) *automock.ExternalSvcCallerProvider
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
			Input:          lblInputWithoutRegion,
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
			Name:           "Error when region label is not string",
			Config:         testConfig,
			CallerProvider: rtmtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			Input: model.RuntimeInput{
				Labels: graphql.Labels{selfRegisterDistinguishLabelKey: distinguishLblVal, regionLabelKey: struct{}{}},
			},
			Context:        ctxWithCertConsumer,
			ExpectedErr:    fmt.Errorf("region value should be of type %q", "string"),
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
			Config:         testConfig,
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
			manager, err := runtime.NewSelfRegisterManager(testCase.Config, svcCallerProvider)
			require.NoError(t, err)

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

	testCases := []struct {
		Name                                string
		Config                              config.SelfRegConfig
		CallerProvider                      func(*testing.T, config.SelfRegConfig, string) *automock.ExternalSvcCallerProvider
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
			Config:                              testConfig,
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
			manager, err := runtime.NewSelfRegisterManager(testCase.Config, svcCallerProvider)
			require.NoError(t, err)

			err = manager.CleanupSelfRegisteredRuntime(testCase.Context, testCase.SelfRegisteredDistinguishLabelValue, testCase.Region)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewSelfRegisterManager(t *testing.T) {
	t.Run("Error when creating self register manager fails", func(t *testing.T) {
		cfg := config.SelfRegConfig{}
		manager, err := runtime.NewSelfRegisterManager(cfg, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "self registration secret path cannot be empty")
		require.Nil(t, manager)
	})
}
