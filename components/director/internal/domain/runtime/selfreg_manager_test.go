package runtime_test

import (
	"context"
	"errors"
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
	RegionToConfig: map[string]runtime.InstanceConfig{
		"test-region": runtime.InstanceConfig{
			ClientID:                        "test-client-id",
			ClientSecret:                    "test-client-secret",
			URL:                             "https://test-url.com",
		},
	},

	ClientTimeout:                   5 * time.Second,
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
		Labels: graphql.Labels{selfRegisterDistinguishLabelKey: distinguishLblVal},
	}
	lblInputAfterPrep := map[string]interface{}{
		testConfig.SelfRegisterLabelKey: rtmtest.ResponseLabelValue,
	}
	emptyLabels := make(map[string]interface{})
	lblInvalidInput := model.RuntimeInput{Labels: graphql.Labels{selfRegisterDistinguishLabelKey: "invalid value"}}

	fakeConfig := testConfig
	fakeConfig.RegionToConfig["fake-region"] = runtime.InstanceConfig{URL: "https://test-url    .com"}

	ctxWithTokenConsumer := consumer.SaveToContext(context.TODO(), tokenConsumer)
	ctxWithCertConsumer := consumer.SaveToContext(context.TODO(), certConsumer)

	testCases := []struct {
		Name           string
		Config         runtime.SelfRegConfig
		Caller         func(*testing.T) *automock.ExternalSvcCaller
		Input          model.RuntimeInput
		Context        context.Context
		ExpectedErr    error
		ExpectedOutput map[string]interface{}
	}{
		{
			Name:           "Success",
			Config:         testConfig,
			Input:          lblInput,
			Caller:         rtmtest.CallerThatGetsCalledOnce(http.StatusCreated),
			Context:        ctxWithCertConsumer,
			ExpectedErr:    nil,
			ExpectedOutput: lblInputAfterPrep,
		},
		{
			Name:           "Success for non-matching consumer",
			Config:         testConfig,
			Caller:         rtmtest.CallerThatDoesNotGetCalled,
			Input:          model.RuntimeInput{},
			Context:        ctxWithTokenConsumer,
			ExpectedErr:    nil,
			ExpectedOutput: emptyLabels,
		},
		{
			Name:           "Error when context does not contain consumer",
			Config:         testConfig,
			Caller:         rtmtest.CallerThatDoesNotGetCalled,
			Input:          model.RuntimeInput{},
			Context:        context.TODO(),
			ExpectedErr:    consumer.NoConsumerError,
			ExpectedOutput: emptyLabels,
		},
		{
			Name:           "Error when can't create URL for preparation of self-registered runtime",
			Config:         fakeConfig,
			Caller:         rtmtest.CallerThatDoesNotGetCalled,
			Input:          lblInvalidInput,
			Context:        ctxWithCertConsumer,
			ExpectedErr:    errors.New("while creating url for preparation of self-registered runtime"),
			ExpectedOutput: emptyLabels,
		},
		{
			Name:           "Error when Call doesn't succeed",
			Config:         testConfig,
			Caller:         rtmtest.CallerThatDoesNotSucceed,
			Input:          lblInput,
			Context:        ctxWithCertConsumer,
			ExpectedErr:    rtmtest.TestError,
			ExpectedOutput: emptyLabels,
		},
		{
			Name:           "Error when status code is unexpected",
			Config:         testConfig,
			Caller:         rtmtest.CallerThatReturnsBadStatus,
			Input:          lblInput,
			Context:        ctxWithCertConsumer,
			ExpectedErr:    errors.New("received unexpected status"),
			ExpectedOutput: emptyLabels,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svcCaller := testCase.Caller(t)
			manager := runtime.NewSelfRegisterManager(testCase.Config, svcCaller)

			output, err := manager.PrepareRuntimeForSelfRegistration(testCase.Context, testCase.Input, testUUID)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, testCase.ExpectedOutput, output)

			svcCaller.AssertExpectations(t)
		})
	}
}

func TestSelfRegisterManager_CleanupSelfRegisteredRuntime(t *testing.T) {
	ctx := context.TODO()
	fakeConfig := testConfig
	fakeConfig.RegionToConfig["fake-region"] = runtime.InstanceConfig{URL: "https://test-url    .com"}

	testCases := []struct {
		Name                                string
		Config                              runtime.SelfRegConfig
		Caller                              func(*testing.T) *automock.ExternalSvcCaller
		SelfRegisteredDistinguishLabelValue string
		Context                             context.Context
		ExpectedErr                         error
	}{
		{
			Name:                                "Success",
			Caller:                              rtmtest.CallerThatGetsCalledOnce(http.StatusOK),
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctx,
			ExpectedErr:                         nil,
		},
		{
			Name:                                "Success when runtime is not self-registered",
			Caller:                              rtmtest.CallerThatDoesNotGetCalled,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: "",
			Context:                             ctx,
			ExpectedErr:                         nil,
		},
		{
			Name:                                "Error when can't create URL for cleanup of self-registered runtime",
			Caller:                              rtmtest.CallerThatDoesNotGetCalled,
			Config:                              fakeConfig,
			SelfRegisteredDistinguishLabelValue: "invalid value",
			Context:                             ctx,
			ExpectedErr:                         errors.New("while creating url for cleanup of self-registered runtime"),
		},
		{
			Name:                                "Error when Call doesn't succeed",
			Caller:                              rtmtest.CallerThatDoesNotSucceed,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctx,
			ExpectedErr:                         rtmtest.TestError,
		},
		{
			Name:                                "Error when Call doesn't succeed",
			Caller:                              rtmtest.CallerThatReturnsBadStatus,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctx,
			ExpectedErr:                         errors.New("received unexpected status code"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			manager := runtime.NewSelfRegisterManager(testCase.Config, )

			err := manager.CleanupSelfRegisteredRuntime(testCase.Context, testCase.SelfRegisteredDistinguishLabelValue)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
