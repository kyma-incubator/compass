package selfregmanager_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager"
	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager/automock"
	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager/selfregmngrtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"github.com/kyma-incubator/compass/components/director/internal/model"

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
)

var testConfig = config.SelfRegConfig{
	SelfRegisterDistinguishLabelKey: selfRegisterDistinguishLabelKey,
	SelfRegisterLabelKey:            "test-label-key",
	SelfRegisterLabelValuePrefix:    "test-prefix",
	SelfRegisterResponseKey:         selfregmngrtest.ResponseLabelKey,
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
	consumerID := "test-consumer-id"
	tokenConsumer := consumer.Consumer{
		ConsumerID: consumerID,
		Flow:       oathkeeper.OAuth2Flow,
	}
	certConsumer := consumer.Consumer{
		ConsumerID: consumerID,
		Flow:       oathkeeper.CertificateFlow,
	}
	var emptyLabels map[string]interface{}

	lblInput := model.RuntimeInput{
		Labels: graphql.Labels{selfRegisterDistinguishLabelKey: distinguishLblVal, selfregmanager.RegionLabel: testRegion},
	}
	lblInputWithoutRegion := model.RuntimeInput{
		Labels: graphql.Labels{selfRegisterDistinguishLabelKey: distinguishLblVal},
	}
	lblInputAfterPrep := map[string]interface{}{
		testConfig.SelfRegisterLabelKey: selfregmngrtest.ResponseLabelValue,
		selfregmanager.RegionLabel:      testRegion,
		selfRegisterDistinguishLabelKey: distinguishLblVal,
	}
	lblInputAfterPrepWithSubaccount := map[string]interface{}{
		testConfig.SelfRegisterLabelKey:    selfregmngrtest.ResponseLabelValue,
		scenarioassignment.SubaccountIDKey: consumerID,
		selfregmanager.RegionLabel:         testRegion,
		selfRegisterDistinguishLabelKey:    distinguishLblVal,
	}
	lblWithDistinguish := map[string]interface{}{
		selfRegisterDistinguishLabelKey: distinguishLblVal,
	}
	lblWithDistinguishAndFakeRegion := map[string]interface{}{
		selfRegisterDistinguishLabelKey: "invalid value",
		selfregmanager.RegionLabel:      fakeRegion,
	}
	lblWithDistinguishAndStructRegion := map[string]interface{}{
		selfRegisterDistinguishLabelKey: distinguishLblVal,
		selfregmanager.RegionLabel:      struct{}{},
	}
	lblWithDistinguishAndInvalidRegion := map[string]interface{}{
		selfRegisterDistinguishLabelKey: distinguishLblVal,
		selfregmanager.RegionLabel:      "not-valid",
	}

	ctxWithTokenConsumer := consumer.SaveToContext(context.TODO(), tokenConsumer)
	ctxWithCertConsumer := consumer.SaveToContext(context.TODO(), certConsumer)

	testCases := []struct {
		Name           string
		Config         config.SelfRegConfig
		CallerProvider func(*testing.T, config.SelfRegConfig, string) *automock.ExternalSvcCallerProvider
		Region         string
		Input          model.RuntimeInput
		Context        context.Context
		ResourceType   resource.Type
		ExpectedErr    error
		ExpectedOutput map[string]interface{}
	}{
		{
			Name:           "Success",
			Config:         testConfig,
			Input:          lblInput,
			CallerProvider: selfregmngrtest.CallerThatGetsCalledOnce(http.StatusCreated),
			Region:         testRegion,
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			ExpectedErr:    nil,
			ExpectedOutput: lblInputAfterPrep,
		},
		{
			Name:           "Success with subaccount label for application templates",
			Config:         testConfig,
			Input:          lblInput,
			CallerProvider: selfregmngrtest.CallerThatGetsCalledOnce(http.StatusCreated),
			Region:         testRegion,
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.ApplicationTemplate,
			ExpectedErr:    nil,
			ExpectedOutput: lblInputAfterPrepWithSubaccount,
		},
		{
			Name:           "Success for non-matching consumer",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			Input:          lblInputWithoutRegion,
			Context:        ctxWithTokenConsumer,
			ResourceType:   resource.Runtime,
			ExpectedErr:    nil,
			ExpectedOutput: lblWithDistinguish,
		},
		{
			Name:           "Error when region label is missing",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			Input:          lblInputWithoutRegion,
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			ExpectedErr:    fmt.Errorf("missing %q label", selfregmanager.RegionLabel),
			ExpectedOutput: lblWithDistinguish,
		},
		{
			Name:           "Error when region label is not string",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			Input: model.RuntimeInput{
				Labels: graphql.Labels{selfRegisterDistinguishLabelKey: distinguishLblVal, selfregmanager.RegionLabel: struct{}{}},
			},
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			ExpectedErr:    fmt.Errorf("region value should be of type %q", "string"),
			ExpectedOutput: lblWithDistinguishAndStructRegion,
		},
		{
			Name:           "Error when region doesn't exist",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			Input:          model.RuntimeInput{Labels: graphql.Labels{selfRegisterDistinguishLabelKey: distinguishLblVal, selfregmanager.RegionLabel: "not-valid"}},
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			ExpectedErr:    errors.New("missing configuration for region"),
			ExpectedOutput: lblWithDistinguishAndInvalidRegion,
		},
		{
			Name:           "Error when caller provider fails",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerProviderThatFails,
			Region:         testRegion,
			Input:          lblInput,
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			ExpectedErr:    errors.New("while getting caller"),
			ExpectedOutput: lblInputAfterPrep,
		},
		{
			Name:           "Error when context does not contain consumer",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			Input:          model.RuntimeInput{},
			Context:        context.TODO(),
			ResourceType:   resource.Runtime,
			ExpectedErr:    consumer.NoConsumerError,
			ExpectedOutput: emptyLabels,
		},
		{
			Name:           "Error when can't create URL for preparation of self-registration",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         fakeRegion,
			Input:          model.RuntimeInput{Labels: graphql.Labels{selfRegisterDistinguishLabelKey: "invalid value", selfregmanager.RegionLabel: fakeRegion}},
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			ExpectedErr:    errors.New("while creating url for preparation of self-registered resource"),
			ExpectedOutput: lblWithDistinguishAndFakeRegion,
		},
		{
			Name:           "Error when Call doesn't succeed",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotSucceed,
			Region:         testRegion,
			Input:          lblInput,
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			ExpectedErr:    selfregmngrtest.TestError,
			ExpectedOutput: lblInputAfterPrep,
		},
		{
			Name:           "Error when status code is unexpected",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatReturnsBadStatus,
			Region:         testRegion,
			Input:          lblInput,
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			ExpectedErr:    errors.New("received unexpected status"),
			ExpectedOutput: lblInputAfterPrep,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svcCallerProvider := testCase.CallerProvider(t, testCase.Config, testCase.Region)
			manager, err := selfregmanager.NewSelfRegisterManager(testCase.Config, svcCallerProvider)
			require.NoError(t, err)

			output, err := manager.PrepareForSelfRegistration(testCase.Context, testCase.ResourceType, testCase.Input.Labels, testUUID)
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
			CallerProvider:                      selfregmngrtest.CallerThatGetsCalledOnce(http.StatusOK),
			Region:                              testRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctx,
			ExpectedErr:                         nil,
		},
		{
			Name:                                "Success when resoource is not self-registered",
			CallerProvider:                      selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:                              testRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: "",
			Context:                             ctx,
			ExpectedErr:                         nil,
		},
		{
			Name:                                "Error when can't create URL for cleanup of self-registered resource",
			CallerProvider:                      selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:                              fakeRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: "invalid value",
			Context:                             ctx,
			ExpectedErr:                         errors.New("while creating url for cleanup of self-registered resource"),
		},
		{
			Name:                                "Error when region doesn't exist",
			Config:                              testConfig,
			CallerProvider:                      selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:                              "not-valid",
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctx,
			ExpectedErr:                         errors.New("missing configuration for region"),
		},
		{
			Name:                                "Error when caller provider fails",
			Config:                              testConfig,
			CallerProvider:                      selfregmngrtest.CallerProviderThatFails,
			Region:                              testRegion,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctx,
			ExpectedErr:                         errors.New("while getting caller"),
		},
		{
			Name:                                "Error when Call doesn't succeed",
			CallerProvider:                      selfregmngrtest.CallerThatDoesNotSucceed,
			Region:                              testRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctx,
			ExpectedErr:                         selfregmngrtest.TestError,
		},
		{
			Name:                                "Error when Call doesn't succeed",
			CallerProvider:                      selfregmngrtest.CallerThatReturnsBadStatus,
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
			manager, err := selfregmanager.NewSelfRegisterManager(testCase.Config, svcCallerProvider)
			require.NoError(t, err)

			err = manager.CleanupSelfRegistration(testCase.Context, testCase.SelfRegisteredDistinguishLabelValue, testCase.Region)
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
		manager, err := selfregmanager.NewSelfRegisterManager(cfg, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "self registration secret path cannot be empty")
		require.Nil(t, manager)
	})
}
