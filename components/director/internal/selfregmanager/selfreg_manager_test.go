package selfregmanager_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager"
	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager/automock"
	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager/selfregmngrtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/stretchr/testify/require"
)

const (
	selfRegisterDistinguishLabelKey = "test-distinguish-label-key"
	distinguishLblVal               = "test-value"
	testUUID                        = "b3ea1977-582e-4d61-ae12-b3a837a3858e"
	testRegion                      = "test-region"
	fakeRegion                      = "fake-region"
	missingSaaSAppRegion            = "missing-saas-app-region"
	emptySaaSAppRegion              = "empty-saas-app-region"
	consumerID                      = "test-consumer-id"
	testSaaSAppName                 = "testSaaSAppName-1"
	secondTestSaaSAppName           = "testSaaSAppName-2"
)

var (
	testConfig = config.SelfRegConfig{
		SelfRegisterDistinguishLabelKey: selfRegisterDistinguishLabelKey,
		SelfRegisterLabelKey:            "test-label-key",
		SelfRegisterLabelValuePrefix:    "test-prefix",
		SelfRegisterResponseKey:         selfregmngrtest.ResponseLabelKey,
		SaaSAppNameLabelKey:             "test-CMPSaaSAppName",
		SelfRegisterPath:                "test-path",
		SelfRegisterNameQueryParam:      "testNameQuery",
		SelfRegisterTenantQueryParam:    "testTenantQuery",
		SelfRegisterRequestBodyPattern:  `{"%s":"test"}`,
		SelfRegisterSecretPath:          "testdata/TestSelfRegisterManager_PrepareRuntimeForSelfRegistration.golden",
		SelfRegSaaSAppSecretPath:        "testdata/TestSelfRegisterManager_SaaSAppName.golden",
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
			"missing-saas-app-region": {
				ClientID:     "client_id",
				ClientSecret: "client_secret",
				URL:          "https://test-url-second.com",
				TokenURL:     "https://test-token-url-second.com",
				Cert:         "cert",
				Key:          "key",
			},
			"empty-saas-app-region": {
				ClientID:     "client_id",
				ClientSecret: "client_secret",
				URL:          "https://test-url-second.com",
				TokenURL:     "https://test-token-url-second.com",
				Cert:         "cert",
				Key:          "key",
			},
		},
		SaaSAppNamePath: "localSaaSAppNamePath",
		RegionToSaaSAppName: map[string]string{
			"test-region":           testSaaSAppName,
			"second-region":         secondTestSaaSAppName,
			"empty-saas-app-region": "",
		},
		ClientTimeout: 5 * time.Second,
	}

	tokenConsumer = consumer.Consumer{
		ConsumerID: consumerID,
		Flow:       oathkeeper.OAuth2Flow,
		Region:     testRegion,
	}

	certConsumer = consumer.Consumer{
		ConsumerID: consumerID,
		Flow:       oathkeeper.CertificateFlow,
		Region:     testRegion,
	}

	certConsumerWithoutRegion = consumer.Consumer{
		ConsumerID: consumerID,
		Flow:       oathkeeper.CertificateFlow,
	}

	certConsumerWithFakeRegion = consumer.Consumer{
		ConsumerID: consumerID,
		Flow:       oathkeeper.CertificateFlow,
		Region:     fakeRegion,
	}

	certConsumerWithMissingSaaSAppRegion = consumer.Consumer{
		ConsumerID: consumerID,
		Flow:       oathkeeper.CertificateFlow,
		Region:     missingSaaSAppRegion,
	}

	certConsumerWithEmptySaaSAppRegion = consumer.Consumer{
		ConsumerID: consumerID,
		Flow:       oathkeeper.CertificateFlow,
		Region:     emptySaaSAppRegion,
	}
)

func TestSelfRegisterManager_IsSelfRegistrationFlow(t *testing.T) {
	ctxWithTokenConsumer := consumer.SaveToContext(context.TODO(), tokenConsumer)
	ctxWithCertConsumer := consumer.SaveToContext(context.TODO(), certConsumer)

	testCases := []struct {
		Name           string
		Config         config.SelfRegConfig
		Region         string
		InputLabels    map[string]interface{}
		Context        context.Context
		ExpectedErr    error
		ExpectedOutput bool
	}{
		{
			Name:           "Success",
			Config:         testConfig,
			InputLabels:    fixLblInput(),
			Region:         testRegion,
			Context:        ctxWithCertConsumer,
			ExpectedErr:    nil,
			ExpectedOutput: true,
		},
		{
			Name:           "Success for non-matching consumer",
			Config:         testConfig,
			Region:         testRegion,
			InputLabels:    fixLblWithoutRegion(),
			Context:        ctxWithTokenConsumer,
			ExpectedErr:    nil,
			ExpectedOutput: false,
		},
		{
			Name:           "Error for missing distinguished label",
			Config:         testConfig,
			Region:         testRegion,
			InputLabels:    map[string]interface{}{},
			Context:        ctxWithCertConsumer,
			ExpectedErr:    fmt.Errorf("missing %q label", selfRegisterDistinguishLabelKey),
			ExpectedOutput: false,
		},
		{
			Name:           "Error when context does not contain consumer",
			Config:         testConfig,
			Region:         testRegion,
			InputLabels:    map[string]interface{}{},
			Context:        context.TODO(),
			ExpectedErr:    consumer.NoConsumerError,
			ExpectedOutput: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			manager, err := selfregmanager.NewSelfRegisterManager(testCase.Config, nil)
			require.NoError(t, err)

			output, err := manager.IsSelfRegistrationFlow(testCase.Context, testCase.InputLabels)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, testCase.ExpectedOutput, output)
		})
	}
}

func TestSelfRegisterManager_PrepareForSelfRegistration(t *testing.T) {
	ctxWithTokenConsumer := consumer.SaveToContext(context.TODO(), tokenConsumer)
	ctxWithCertConsumer := consumer.SaveToContext(context.TODO(), certConsumer)
	ctxWithCertConsumerWithoutRegion := consumer.SaveToContext(context.TODO(), certConsumerWithoutRegion)
	ctxWithCertConsumerWithFakeRegion := consumer.SaveToContext(context.TODO(), certConsumerWithFakeRegion)
	ctxWithCertConsumerWithMissingSaaSAppRegion := consumer.SaveToContext(context.TODO(), certConsumerWithMissingSaaSAppRegion)
	ctxWithCertConsumerWithEmptySaaSAppRegion := consumer.SaveToContext(context.TODO(), certConsumerWithEmptySaaSAppRegion)

	testCases := []struct {
		Name           string
		Config         config.SelfRegConfig
		CallerProvider func(*testing.T, config.SelfRegConfig, string) *automock.ExternalSvcCallerProvider
		Region         string
		InputLabels    map[string]interface{}
		Context        context.Context
		ResourceType   resource.Type
		Validation     func() error
		ExpectedErr    error
		ExpectedOutput map[string]interface{}
	}{
		{
			Name:           "Success",
			Config:         testConfig,
			InputLabels:    fixLblWithoutRegion(),
			CallerProvider: selfregmngrtest.CallerThatGetsCalledOnce(http.StatusCreated),
			Region:         testRegion,
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			Validation:     func() error { return nil },
			ExpectedErr:    nil,
			ExpectedOutput: fixLblInputAfterPrep(),
		},
		{
			Name:           "Success with subaccount label for application templates",
			Config:         testConfig,
			InputLabels:    fixLblWithoutRegion(),
			CallerProvider: selfregmngrtest.CallerThatGetsCalledOnce(http.StatusCreated),
			Region:         testRegion,
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.ApplicationTemplate,
			Validation:     func() error { return nil },
			ExpectedErr:    nil,
			ExpectedOutput: fixLblInputAfterPrepWithSubaccount(),
		},
		{
			Name:           "Success for non-matching consumer",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			InputLabels:    fixLblWithoutRegion(),
			Context:        ctxWithTokenConsumer,
			ResourceType:   resource.Runtime,
			Validation:     func() error { return nil },
			ExpectedErr:    nil,
			ExpectedOutput: fixLblWithDistinguish(),
		},
		{
			Name:           "Success for missing distinguished label but resource is Runtime",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			InputLabels:    map[string]interface{}{},
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			Validation:     func() error { return nil },
			ExpectedErr:    nil,
			ExpectedOutput: map[string]interface{}{},
		},
		{
			Name:           "Error validation failed",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			InputLabels:    fixLblInput(),
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.ApplicationTemplate,
			Validation:     func() error { return errors.New("validation failed") },
			ExpectedErr:    errors.New("validation failed"),
			ExpectedOutput: nil,
		},
		{
			Name:           "Error for missing distinguished label and resource is App Template",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			InputLabels:    map[string]interface{}{},
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.ApplicationTemplate,
			Validation:     func() error { return nil },
			ExpectedErr:    fmt.Errorf("missing %q label", selfRegisterDistinguishLabelKey),
			ExpectedOutput: nil,
		},
		{
			Name:           "Error during region check when tenant region is unable to be retrieved from context",
			Config:         testConfig,
			InputLabels:    fixLblWithoutRegion(),
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			Context:        ctxWithCertConsumerWithoutRegion,
			ResourceType:   resource.Runtime,
			Validation:     func() error { return nil },
			ExpectedErr:    fmt.Errorf("missing %s value in consumer context", selfregmanager.RegionLabel),
			ExpectedOutput: nil,
		},
		{
			Name:           "Error when region label is provided",
			Config:         testConfig,
			InputLabels:    fixLblInput(),
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			Validation:     func() error { return nil },
			ExpectedErr:    fmt.Errorf("providing %q label and value is forbidden", selfregmanager.RegionLabel),
			ExpectedOutput: nil,
		},
		{
			Name:           "Error when caller provider fails",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerProviderThatFails,
			Region:         testRegion,
			InputLabels:    fixLblWithoutRegion(),
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			Validation:     func() error { return nil },
			ExpectedErr:    errors.New("while getting caller"),
			ExpectedOutput: nil,
		},
		{
			Name:           "Error when context does not contain consumer",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         testRegion,
			InputLabels:    map[string]interface{}{},
			Context:        context.TODO(),
			ResourceType:   resource.Runtime,
			Validation:     func() error { return nil },
			ExpectedErr:    consumer.NoConsumerError,
			ExpectedOutput: nil,
		},
		{
			Name:           "Error when can't create URL for preparation of self-registration",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:         fakeRegion,
			InputLabels:    map[string]interface{}{selfRegisterDistinguishLabelKey: "invalid value"},
			Context:        ctxWithCertConsumerWithFakeRegion,
			ResourceType:   resource.Runtime,
			Validation:     func() error { return nil },
			ExpectedErr:    errors.New("while creating url for preparation of self-registered resource"),
			ExpectedOutput: nil,
		},
		{
			Name:           "Error when Call doesn't succeed",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatDoesNotSucceed,
			Region:         testRegion,
			InputLabels:    fixLblWithoutRegion(),
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			Validation:     func() error { return nil },
			ExpectedErr:    selfregmngrtest.TestError,
			ExpectedOutput: nil,
		},
		{
			Name:           "Error when status code is unexpected",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatReturnsBadStatus,
			Region:         testRegion,
			InputLabels:    fixLblWithoutRegion(),
			Context:        ctxWithCertConsumer,
			ResourceType:   resource.Runtime,
			Validation:     func() error { return nil },
			ExpectedErr:    errors.New("received unexpected status"),
			ExpectedOutput: nil,
		},
		{
			Name:           "Error when SaaS application name is not found for a given region",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatGetsCalledOnce(http.StatusCreated),
			Region:         missingSaaSAppRegion,
			InputLabels:    fixLblWithoutRegion(),
			Context:        ctxWithCertConsumerWithMissingSaaSAppRegion,
			ResourceType:   resource.Runtime,
			Validation:     func() error { return nil },
			ExpectedErr:    fmt.Errorf("missing SaaS application name for region: \"%s\"", missingSaaSAppRegion),
			ExpectedOutput: nil,
		},
		{
			Name:           "Error when SaaS application name is empty",
			Config:         testConfig,
			CallerProvider: selfregmngrtest.CallerThatGetsCalledOnce(http.StatusCreated),
			Region:         emptySaaSAppRegion,
			InputLabels:    fixLblWithoutRegion(),
			Context:        ctxWithCertConsumerWithEmptySaaSAppRegion,
			ResourceType:   resource.Runtime,
			Validation:     func() error { return nil },
			ExpectedErr:    fmt.Errorf("SaaS application name for region: \"%s\" could not be empty", emptySaaSAppRegion),
			ExpectedOutput: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svcCallerProvider := testCase.CallerProvider(t, testCase.Config, testCase.Region)
			manager, err := selfregmanager.NewSelfRegisterManager(testCase.Config, svcCallerProvider)
			require.NoError(t, err)

			output, err := manager.PrepareForSelfRegistration(testCase.Context, testCase.ResourceType, testCase.InputLabels, testUUID, testCase.Validation)
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

func TestSelfRegisterManager_CleanupSelfRegistration(t *testing.T) {
	tokenConsumer := consumer.Consumer{
		ConsumerID: consumerID,
		Flow:       oathkeeper.OAuth2Flow,
	}
	certConsumer := consumer.Consumer{
		ConsumerID: consumerID,
		Flow:       oathkeeper.CertificateFlow,
	}

	ctxWithTokenConsumer := consumer.SaveToContext(context.TODO(), tokenConsumer)
	ctxWithCertConsumer := consumer.SaveToContext(context.TODO(), certConsumer)

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
			Context:                             ctxWithCertConsumer,
			ExpectedErr:                         nil,
		},
		{
			Name:                                "Success when resoource is not self-registered",
			CallerProvider:                      selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:                              testRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: "",
			Context:                             ctxWithCertConsumer,
			ExpectedErr:                         nil,
		},
		{
			Name:                                "Error when can't create URL for cleanup of self-registered resource",
			CallerProvider:                      selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:                              fakeRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: "invalid value",
			Context:                             ctxWithCertConsumer,
			ExpectedErr:                         errors.New("while creating url for cleanup of self-registered resource"),
		},
		{
			Name:                                "Error when region doesn't exist",
			Config:                              testConfig,
			CallerProvider:                      selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:                              "not-valid",
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctxWithCertConsumer,
			ExpectedErr:                         errors.New("missing configuration for region"),
		},
		{
			Name:                                "Error when caller provider fails",
			Config:                              testConfig,
			CallerProvider:                      selfregmngrtest.CallerProviderThatFails,
			Region:                              testRegion,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctxWithCertConsumer,
			ExpectedErr:                         errors.New("while getting caller"),
		},
		{
			Name:                                "Error when Call doesn't succeed",
			CallerProvider:                      selfregmngrtest.CallerThatDoesNotSucceed,
			Region:                              testRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctxWithCertConsumer,
			ExpectedErr:                         selfregmngrtest.TestError,
		},
		{
			Name:                                "Error when Call doesn't succeed",
			CallerProvider:                      selfregmngrtest.CallerThatReturnsBadStatus,
			Region:                              testRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: distinguishLblVal,
			Context:                             ctxWithCertConsumer,
			ExpectedErr:                         errors.New("received unexpected status code"),
		},
		{
			Name:                                "Success when token consumer is used",
			CallerProvider:                      selfregmngrtest.CallerThatDoesNotGetCalled,
			Region:                              testRegion,
			Config:                              testConfig,
			SelfRegisteredDistinguishLabelValue: "",
			Context:                             ctxWithTokenConsumer,
			ExpectedErr:                         nil,
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
		require.Contains(t, err.Error(), "config path cannot be empty")
		require.Nil(t, manager)
	})
}

func fixLblInputAfterPrep() map[string]interface{} {
	return map[string]interface{}{
		testConfig.SelfRegisterLabelKey: selfregmngrtest.ResponseLabelValue,
		selfregmanager.RegionLabel:      testRegion,
		selfRegisterDistinguishLabelKey: distinguishLblVal,
		testConfig.SaaSAppNameLabelKey:  testSaaSAppName,
	}
}

func fixLblInputAfterPrepWithSubaccount() map[string]interface{} {
	return map[string]interface{}{
		testConfig.SelfRegisterLabelKey:    selfregmngrtest.ResponseLabelValue,
		scenarioassignment.SubaccountIDKey: consumerID,
		selfregmanager.RegionLabel:         testRegion,
		selfRegisterDistinguishLabelKey:    distinguishLblVal,
		testConfig.SaaSAppNameLabelKey:     testSaaSAppName,
	}
}

func fixLblWithDistinguish() map[string]interface{} {
	return map[string]interface{}{
		selfRegisterDistinguishLabelKey: distinguishLblVal,
	}
}

func fixLblInput() map[string]interface{} {
	return map[string]interface{}{
		selfRegisterDistinguishLabelKey: distinguishLblVal,
		selfregmanager.RegionLabel:      testRegion,
	}
}

func fixLblWithoutRegion() map[string]interface{} {
	return map[string]interface{}{
		selfRegisterDistinguishLabelKey: distinguishLblVal,
	}
}
