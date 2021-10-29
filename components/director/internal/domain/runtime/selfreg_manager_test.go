package runtime_test

import (
	"context"
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
		Input       *graphql.RuntimeInput
		Context     context.Context
		ExpectedErr error
	}{
		{
			Name:        "Success for non-matching consumer",
			Config:      testConfig,
			Input:       &graphql.RuntimeInput{},
			Context:     ctxWithTokenConsumer,
			ExpectedErr: nil,
		},
		//{
		//	Name:        "Success",
		//	Config:      ctxWithCertConsumer,
		//	ExpectedErr: nil,
		//},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			manager := runtime.NewSelfRegisterManager(testCase.Config)

			err := manager.PrepareRuntimeForSelfRegistration(testCase.Context, testCase.Input)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
