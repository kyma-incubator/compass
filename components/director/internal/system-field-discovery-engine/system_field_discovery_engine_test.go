package systemfielddiscoveryengine_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	systemfielddiscoveryengine "github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine"
	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	internalTnt = "internalTenant"
	externalTnt = "externalTnt"
)

var (
	testConfig = config.SystemFieldDiscoveryEngineConfig{
		SaasRegSecretPath:       "testdata/TestSystemFieldDiscoveryEngine_data.golden",
		OauthTokenPath:          "/oauth/token",
		SaasRegClientIDPath:     "clientId",
		SaasRegClientSecretPath: "clientSecret",
		SaasRegTokenURLPath:     "url",
		SaasRegURLPath:          "saas_registry_url",
		RegionToSaasRegConfig: map[string]config.SaasRegConfig{
			"test-region": {
				ClientID:        "client_id",
				ClientSecret:    "client_secret",
				TokenURL:        "https://test-url-second.com",
				SaasRegistryURL: "https://saas_registry_url",
			},
		},
	}
)

func Test_EnrichApplicationWebhookIfNeeded(t *testing.T) {
	appInput := model.ApplicationRegisterInput{
		Webhooks: []*model.WebhookInput{},
	}

	testCases := []struct {
		Name                           string
		Config                         config.SystemFieldDiscoveryEngineConfig
		Input                          model.ApplicationRegisterInput
		systemFieldDiscoveryLabelValue bool
		ExpectedWebhooks               []*model.WebhookInput
		ExpectedLabelValue             bool
	}{
		{
			Name:                           "Label is true",
			Config:                         testConfig,
			systemFieldDiscoveryLabelValue: true,
			Input:                          appInput,
			ExpectedWebhooks: []*model.WebhookInput{{
				Type: model.WebhookTypeSystemFieldDiscovery,
				URL:  str.Ptr("https://saas_registry_url/saas-manager/v1/service/subscriptions?includeIndirectSubscriptions=true&tenantId=subaccountID"),
				Auth: &model.AuthInput{
					Credential: &model.CredentialDataInput{
						Oauth: &model.OAuthCredentialDataInput{
							ClientID:     "client_id",
							ClientSecret: "client_secret",
							URL:          "https://test-url-second.com/oauth/token",
						},
					},
				},
			}},
			ExpectedLabelValue: true,
		},
		{
			Name:                           "Label is false",
			Config:                         testConfig,
			Input:                          appInput,
			systemFieldDiscoveryLabelValue: false,
			ExpectedWebhooks:               []*model.WebhookInput{},
			ExpectedLabelValue:             false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {

			engine, err := systemfielddiscoveryengine.NewSystemFieldDiscoveryEngine(testCase.Config, nil, nil, nil)
			require.NoError(t, err)

			webhooks, labelValue := engine.EnrichApplicationWebhookIfNeeded(context.TODO(), testCase.Input, testCase.systemFieldDiscoveryLabelValue, "test-region", "subaccountID", "appTemplateName", "appName")

			require.Equal(t, testCase.ExpectedWebhooks, webhooks)
			require.Equal(t, testCase.ExpectedLabelValue, labelValue)
		})
	}
}

func Test_CreateLabelForApplicationWebhook(t *testing.T) {
	testErr := errors.New("Test error")
	appID := "testAppID"
	webhookID := "testWebhookID"
	uuid := "647af599-7f2d-485c-a63b-615b5ff6daf1"
	ctx := fixCtxWithTenant()
	labelInput := &model.LabelInput{
		Key:        systemfielddiscoveryengine.RegistryLabelKey,
		ObjectID:   webhookID,
		ObjectType: model.WebhookLabelableObject,
		Value:      systemfielddiscoveryengine.RegistryLabelValue,
	}
	modelWebhook := &model.Webhook{ID: webhookID}

	testCases := []struct {
		Name             string
		Config           config.SystemFieldDiscoveryEngineConfig
		LabelServiceFn   func() *automock.LabelService
		WebhookServiceFn func() *automock.WebhookService
		UIDServiceFn     func() *automock.UidService
		ExpectedError    error
	}{
		{
			Name:   "Success",
			Config: testConfig,
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("CreateLabel", ctx, internalTnt, uuid, labelInput).Return(nil)
				return labelSvc
			},
			UIDServiceFn: func() *automock.UidService {
				UidSvc := &automock.UidService{}
				UidSvc.On("Generate").Return(uuid).Once()
				return UidSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("GetByIDAndWebhookTypeGlobal", ctx, appID, model.ApplicationWebhookReference, model.WebhookTypeSystemFieldDiscovery).Return(modelWebhook, nil).Once()
				return webhookSvc
			},
			ExpectedError: nil,
		},
		{
			Name:   "Returns error when get webhook fails",
			Config: testConfig,
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.AssertNotCalled(t, "CreateLabel")
				return labelSvc
			},
			UIDServiceFn: func() *automock.UidService {
				UidSvc := &automock.UidService{}
				UidSvc.On("Generate").Return(uuid).Once()
				return UidSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("GetByIDAndWebhookTypeGlobal", ctx, appID, model.ApplicationWebhookReference, model.WebhookTypeSystemFieldDiscovery).Return(nil, testErr).Once()
				return webhookSvc
			},
			ExpectedError: testErr,
		},
		{
			Name:   "Returns error create label fails",
			Config: testConfig,
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("CreateLabel", ctx, internalTnt, uuid, labelInput).Return(testErr)
				return labelSvc
			},
			UIDServiceFn: func() *automock.UidService {
				UidSvc := &automock.UidService{}
				UidSvc.On("Generate").Return(uuid).Once()
				return UidSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("GetByIDAndWebhookTypeGlobal", ctx, appID, model.ApplicationWebhookReference, model.WebhookTypeSystemFieldDiscovery).Return(modelWebhook, nil).Once()
				return webhookSvc
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {

			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			webhookSvc := testCase.WebhookServiceFn()

			engine, err := systemfielddiscoveryengine.NewSystemFieldDiscoveryEngine(testCase.Config, labelSvc, webhookSvc, uidSvc)
			require.NoError(t, err)

			err = engine.CreateLabelForApplicationWebhook(ctx, appID)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}
		})
	}
	t.Run("returns error on missing tenant in context", func(t *testing.T) {
		// GIVEN
		engine, err := systemfielddiscoveryengine.NewSystemFieldDiscoveryEngine(testConfig, nil, nil, nil)
		require.NoError(t, err)

		// WHEN
		err = engine.CreateLabelForApplicationWebhook(context.TODO(), appID)

		// THEN
		assert.EqualError(t, err, "cannot read tenant from context")
	})
}

func fixCtxWithTenant() context.Context {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, internalTnt, externalTnt)

	return ctx
}
