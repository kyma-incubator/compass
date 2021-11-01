package runtime_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/rtmtest"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	selfRegisterDistinguishLabelKey = "test-distinguish-label-key"
	distinguishLblVal               = "test-value"
)

var testConfig = runtime.SelfRegConfig{
	SelfRegisterDistinguishLabelKey: selfRegisterDistinguishLabelKey,
	SelfRegisterLabelKey:            rtmtest.ResponseLabelKey,
	SelfRegisterLabelValuePrefix:    "test-prefix",
	SelfRegisterResponseKey:         "test-response-key",
	SelfRegisterPath:                "test-path",
	SelfRegisterNameQueryParam:      "testNameQuery",
	SelfRegisterTenantQueryParam:    "testTenantQuery",
	SelfRegisterRequestBodyPattern:  `{"%s":"test"}`,
	ClientID:                        "test-client-id",
	ClientSecret:                    "test-client-secret",
	URL:                             "https://test-url.com",
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
	lblInput := &graphql.RuntimeInput{
		Labels: graphql.Labels{selfRegisterDistinguishLabelKey: distinguishLblVal},
	}

	fakeConfig := testConfig
	fakeConfig.SelfRegisterPath = "fake path"

	ctxWithTokenConsumer := consumer.SaveToContext(context.TODO(), tokenConsumer)
	ctxWithCertConsumer := consumer.SaveToContext(context.TODO(), certConsumer)

	testCases := []struct {
		Name        string
		Config      runtime.SelfRegConfig
		Caller      func(*testing.T) *automock.ExternalSvcCaller
		Input       *graphql.RuntimeInput
		InputAssert func(*testing.T, *graphql.RuntimeInput)
		Context     context.Context
		ExpectedErr error
	}{
		{
			Name:   "Success",
			Config: testConfig,
			Input:  lblInput,
			Caller: rtmtest.CallerThatGetsCalledOnce(http.StatusCreated),
			InputAssert: func(t *testing.T, in *graphql.RuntimeInput) {
				assert.Equal(t, distinguishLblVal, in.Labels[selfRegisterDistinguishLabelKey])
			},
			Context:     ctxWithCertConsumer,
			ExpectedErr: nil,
		},
		{
			Name:        "Success for non-matching consumer",
			Config:      testConfig,
			Caller:      rtmtest.CallerThatDoesNotGetCalled,
			InputAssert: noopAssert,
			Input:       &graphql.RuntimeInput{},
			Context:     ctxWithTokenConsumer,
			ExpectedErr: nil,
		},
		{
			Name:        "Error when context does not contain consumer",
			Config:      testConfig,
			Caller:      rtmtest.CallerThatDoesNotGetCalled,
			InputAssert: noopAssert,
			Input:       &graphql.RuntimeInput{},
			Context:     context.TODO(),
			ExpectedErr: consumer.NoConsumerError,
		},
		{
			Name:        "Error when can't create URL for preparation of self-registered runtime",
			Config:      fakeConfig,
			Caller:      rtmtest.CallerThatDoesNotGetCalled,
			InputAssert: noopAssert,
			Input:       &graphql.RuntimeInput{Labels: graphql.Labels{selfRegisterDistinguishLabelKey: "invalid value"}},
			Context:     ctxWithCertConsumer,
			ExpectedErr: errors.New("while creating url for preparation of self-registered runtime"),
		},
		{
			Name:        "Error when Call doesn't succeed",
			Config:      testConfig,
			Caller:      rtmtest.CallerThatDoesNotSucceed,
			InputAssert: noopAssert,
			Input:       lblInput,
			Context:     ctxWithCertConsumer,
			ExpectedErr: rtmtest.TestError,
		},
		{
			Name:        "Error when status code is unexpected",
			Config:      testConfig,
			Caller:      rtmtest.CallerThatReturnsBadStatus,
			InputAssert: noopAssert,
			Input:       lblInput,
			Context:     ctxWithCertConsumer,
			ExpectedErr: errors.New("recieved unexpected status"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svcCaller := testCase.Caller(t)
			manager := runtime.NewSelfRegisterManager(testCase.Config, svcCaller)

			err := manager.PrepareRuntimeForSelfRegistration(testCase.Context, testCase.Input)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			testCase.InputAssert(t, testCase.Input)
			svcCaller.AssertExpectations(t)
		})
	}
}

func TestSelfRegisterManager_CleanupSelfRegisteredRuntime(t *testing.T) {
	ctx := context.TODO()

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
			Config:                              testConfig,
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
			svcCaller := testCase.Caller(t)
			manager := runtime.NewSelfRegisterManager(testCase.Config, svcCaller)

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

func noopAssert(*testing.T, *graphql.RuntimeInput) {}
