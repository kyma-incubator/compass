package runtime_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/rtmtest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
)

var testConfig = runtime.SelfRegConfig{
	SelfRegisterDistinguishLabelKey: "test-distinguish-label-key",
	SelfRegisterLabelKey:            "test-label-key",
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
	//certConsumer := consumer.Consumer{
	//	ConsumerID: "test-consumer-id",
	//	Flow:       oathkeeper.CertificateFlow,
	//}

	ctxWithTokenConsumer := consumer.SaveToContext(context.TODO(), tokenConsumer)
	//ctxWithCertConsumer := consumer.SaveToContext(context.TODO(), certConsumer)

	testCases := []struct {
		Name        string
		Config      runtime.SelfRegConfig
		Caller      func(*testing.T) *automock.ExternalSvcCaller
		Input       *graphql.RuntimeInput
		Context     context.Context
		ExpectedErr error
	}{
		//{
		//	Name:        "Success",
		//	Config:      ctxWithCertConsumer,
		//	ExpectedErr: nil,
		//},
		{
			Name:        "Success for non-matching consumer",
			Config:      testConfig,
			Caller:      rtmtest.CallerThatDoesNotGetCalled,
			Input:       &graphql.RuntimeInput{},
			Context:     ctxWithTokenConsumer,
			ExpectedErr: nil,
		},
		{
			Name:        "Error when context does not contain consumer",
			Config:      testConfig,
			Caller:      rtmtest.CallerThatDoesNotGetCalled,
			Input:       &graphql.RuntimeInput{},
			Context:     context.TODO(),
			ExpectedErr: consumer.NoConsumerError,
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

			svcCaller.AssertExpectations(t)
		})
	}
}

//func TestSelfRegisterManager_CleanupSelfRegisteredRuntime(t *testing.T) {
//	//distinguishedLabelVal := "self-registered-runtime"
//	ctx := context.TODO()
//	//testErr := errors.New("test error")
//
//	testCases := []struct {
//		Name        string
//		Config      runtime.SelfRegConfig
//		SelfRegisteredDistinguishLabelValue      string
//		Context     context.Context
//		ExpectedErr error
//	}{
//		{
//			Name:        "Success when runtime is not self-registered",
//			Config:      testConfig,
//			SelfRegisteredDistinguishLabelValue: "",
//			Context:     ctx,
//			ExpectedErr: nil,
//		},
//		{
//			Name:        "Error when can't create URL for cleanup of self-registered runtime",
//			Config:      testConfig,
//			SelfRegisteredDistinguishLabelValue: "invalid value",
//			Context:     ctx,
//			ExpectedErr: errors.New("while creating url for cleanup of self-registered runtime"),
//		},
//		{
//			Name:        "Error when can't create http request for cleanup of self-registered runtime",
//			Config:      testConfig,
//			SelfRegisteredDistinguishLabelValue: "invalid value",
//			Context:     ctx,
//			ExpectedErr: errors.New("while creating url for cleanup of self-registered runtime"),
//		},
//		//{
//		//	Name:        "Success",
//		//	Config:      ctxWithCertConsumer,
//		//	ExpectedErr: nil,
//		//},
//	}
//
//	for _, testCase := range testCases {
//		t.Run(testCase.Name, func(t *testing.T) {
//			manager := runtime.NewSelfRegisterManager(testCase.Config)
//
//			err := manager.CleanupSelfRegisteredRuntime(testCase.Context, testCase.SelfRegisteredDistinguishLabelValue)
//			if testCase.ExpectedErr != nil {
//				require.Error(t, err)
//				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
//			} else {
//				require.NoError(t, err)
//			}
//		})
//	}
//}
