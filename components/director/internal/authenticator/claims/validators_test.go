package claims_test

import (
	"context"
	"fmt"
	"testing"

	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/stretchr/testify/assert"
)

const (
	consumerID           = "consumerID"
	tenantID             = "tenantID"
	consumerTenantID     = "consumer-tnt"
	consumerExtTenantID  = "consumr-ext-tnt"
	providerTenantID     = "provider-tnt"
	providerExtTenantID  = "provider-ext-tnt"
	onBehalfOfConsumerID = "onBehalfOfConsumer"
	region               = "region"
	clientID             = "client_id"
	scopes               = "application:read application:write"
	subscriptionLabelKey = "subscription"

	runtimeID  = "rt-id"
	runtime2ID = "rt-id2"
)

func TestValidator_Validate(t *testing.T) {
	providerLabelKey := "providerName"
	consumerSubaccountLabelKey := "consumer_subaccount_id"
	tokenPrefix := "prefix-"
	testErr := errors.New("test")

	runtime := &model.Runtime{ID: runtimeID, Name: "rt"}

	expectedFilters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(providerLabelKey, fmt.Sprintf("\"%s\"", clientID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}

	rtmCtxWithConsumerSubaccountLabel := &model.RuntimeContextPage{
		Data: []*model.RuntimeContext{
			{
				ID:        "rtmCtxID",
				RuntimeID: runtime2ID,
				Key:       subscriptionLabelKey,
				Value:     tenantID,
			},
		},
	}

	rtmCtxFilter := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", consumerExtTenantID)),
	}

	applicationTemplate := &model.ApplicationTemplate{
		ID: "appTemplateID",
	}

	applicationFilter := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", consumerExtTenantID)),
	}

	applicationPage := &model.ApplicationPage{
		Data: []*model.Application{
			{
				ApplicationTemplateID: &applicationTemplate.ID,
			},
		},
	}

	emptyApplicationPage := &model.ApplicationPage{
		Data: []*model.Application{},
	}

	testCases := []struct {
		Name                         string
		RuntimeServiceFn             func() *automock.RuntimeService
		ApplicationTemplateServiceFn func() *automock.ApplicationTemplateService
		ApplicationServiceFn         func() *automock.ApplicationService
		RuntimeCtxSvcFn              func() *automock.RuntimeCtxService
		IntegrationSystemServiceFn   func() *automock.IntegrationSystemService
		Claims                       claims.Claims
		PersistenceFn                func() *persistenceautomock.PersistenceTx
		TransactionerFn              func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ExpectedErr                  string
	}{
		// common
		{
			Name:   "Succeeds when all claims properties are present",
			Claims: getClaims(consumerTenantID, consumerExtTenantID, scopes),
		},
		{
			Name:   "Succeeds when no scopes are present",
			Claims: getClaims(consumerTenantID, consumerExtTenantID, ""),
		},
		{
			Name:   "Succeeds when both internal and external tenant IDs are missing",
			Claims: getClaims("", "", scopes),
		},
		{
			Name:        "Fails when internal tenant ID is missing",
			Claims:      getClaims("", consumerExtTenantID, ""),
			ExpectedErr: "Tenant not found",
		},
		{
			Name: "Fails when inner validation fails",
			Claims: claims.Claims{
				Tenant: map[string]string{
					"consumerTenant": consumerTenantID,
					"externalTenant": consumerExtTenantID,
				},

				Scopes:       scopes,
				ConsumerID:   consumerID,
				ConsumerType: consumer.Runtime,
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: 1,
				},
			},
			ExpectedErr: "while validating claims",
		},
		{
			Name:        "consumer-provider flow: error when consumer type is not supported",
			Claims:      getClaimsForConsumerProviderFlow(consumer.Application, consumerTenantID, consumerExtTenantID, consumerID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			ExpectedErr: fmt.Sprintf("consumer with type %s is not supported", consumer.Application),
		},
		// runtime
		{
			Name:   "consumer-provider flow: error when token clientID missing",
			Claims: getClaimsForRuntimeConsumerProviderFlow(consumerTenantID, consumerExtTenantID, providerTenantID, providerExtTenantID, scopes, region, ""),
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := txtest.NoopTransactioner(persistTx)
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Twice()
				return transact
			},
			ExpectedErr: "could not find consumer token client ID",
		},
		{
			Name:   "consumer-provider flow: error when transaction cannot be opened",
			Claims: getClaimsForRuntimeConsumerProviderFlow(consumerTenantID, consumerExtTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return txtest.PersistenceContextThatDoesntExpectCommit()
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := txtest.NoopTransactioner(persistTx)
				transact.On("Begin").Return(persistTx, testErr).Twice()
				return transact
			},
			ExpectedErr: "An error has occurred while opening transaction",
		},
		{
			Name:        "consumer-provider flow: error when region missing",
			Claims:      getClaimsForRuntimeConsumerProviderFlow(consumerTenantID, consumerExtTenantID, providerTenantID, providerExtTenantID, scopes, "", clientID),
			ExpectedErr: "could not determine token's region",
		},
		{
			Name:   "Success for consumer-provider flow when runtime with labels and runtime context are found",
			Claims: getClaimsForRuntimeConsumerProviderFlow(consumerTenantID, consumerExtTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServiceFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("GetByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(runtime, nil).Once()
				return runtimeSvc
			},
			RuntimeCtxSvcFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", contextThatHasTenant(consumerTenantID), runtimeID, rtmCtxFilter, 100, "").Return(rtmCtxWithConsumerSubaccountLabel, nil).Once()
				return rtmCtxSvc
			},
		},
		{
			Name:   "Consumer-provider flow: Error when no runtimes nor applications are found",
			Claims: getClaimsForRuntimeConsumerProviderFlow(consumerTenantID, consumerExtTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServiceFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("GetByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(nil, testErr).Once()
				return runtimeSvc
			},
			ApplicationTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(applicationTemplate, nil).Once()
				return appTemplateSvc
			},
			ApplicationServiceFn: func() *automock.ApplicationService {
				applicationSvc := &automock.ApplicationService{}
				applicationSvc.On("List", contextThatHasTenant(consumerTenantID), applicationFilter, 100, "").Return(emptyApplicationPage, nil).Once()
				return applicationSvc
			},
			ExpectedErr: fmt.Sprintf("Consumer's external tenant %s was not found as subscription record in the runtime context table for any runtime in the provider tenant %s", consumerExtTenantID, providerTenantID),
		},
		{
			Name:   "Consumer-provider flow: Error when listing runtime context with filters",
			Claims: getClaimsForRuntimeConsumerProviderFlow(consumerTenantID, consumerExtTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServiceFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("GetByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(runtime, nil).Once()
				return runtimeSvc
			},
			ApplicationTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(applicationTemplate, nil).Once()
				return appTemplateSvc
			},
			ApplicationServiceFn: func() *automock.ApplicationService {
				applicationSvc := &automock.ApplicationService{}
				applicationSvc.On("List", contextThatHasTenant(consumerTenantID), applicationFilter, 100, "").Return(emptyApplicationPage, nil).Once()
				return applicationSvc
			RuntimeCtxSvcFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", contextThatHasTenant(consumerTenantID), runtimeID, mock.Anything, 100, "").Return(nil, testErr).Once()
				return rtmCtxSvc
			},
			ExpectedErr: testErr.Error(),
		},
		{
			Name:   "Success when no runtime, but there is an application subscribed",
			Claims: getClaimsForRuntimeConsumerProviderFlow(consumerTenantID, consumerExtTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServiceFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("ListByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(nil, testErr).Once()
				return runtimeSvc
			},
			ApplicationTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(applicationTemplate, nil).Once()
				return appTemplateSvc
			},
			ApplicationServiceFn: func() *automock.ApplicationService {
				applicationSvc := &automock.ApplicationService{}
				applicationSvc.On("List", contextThatHasTenant(consumerTenantID), applicationFilter, 100, "").Return(applicationPage, nil).Once()
				return applicationSvc
			},
		},
		{
			Name:   "Consumer-provider flow: Error when listing runtime context with filters",
			Claims: getClaimsForRuntimeConsumerProviderFlow(consumerTenantID, consumerExtTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServiceFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("GetByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(runtime, nil).Once()
				return runtimeSvc
			},
			RuntimeCtxSvcFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", contextThatHasTenant(consumerTenantID), runtimeID, mock.Anything, 100, "").Return(&model.RuntimeContextPage{}, nil).Once()
				return rtmCtxSvc
			},
			ApplicationTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(applicationTemplate, nil).Once()
				return appTemplateSvc
			},
			ApplicationServiceFn: func() *automock.ApplicationService {
				applicationSvc := &automock.ApplicationService{}
				applicationSvc.On("List", contextThatHasTenant(consumerTenantID), applicationFilter, 100, "").Return(emptyApplicationPage, nil).Once()
				return applicationSvc
			},
			ExpectedErr: testErr.Error(),
		},
		{
			Name:   "Consumer-provider flow: Error when no subscription was found",
			Claims: getClaimsForRuntimeConsumerProviderFlow(consumerTenantID, consumerExtTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServiceFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("GetByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(runtime, nil).Once()
				return runtimeSvc
			},
			RuntimeCtxSvcFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", contextThatHasTenant(consumerTenantID), runtimeID, mock.Anything, 100, "").Return(&model.RuntimeContextPage{}, nil).Once()
				return rtmCtxSvc
			},
			ExpectedErr: fmt.Sprintf("Consumer's external tenant %s was not found as subscription record in the runtime context table for the runtime in the provider tenant %s", consumerExtTenantID, providerTenantID),
		},
    {
			Name:   "Consumer-provider flow: Error when transaction cannot be committed",
			Claims: getClaimsForRuntimeConsumerProviderFlow(consumerTenantID, consumerExtTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServiceFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("GetByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(runtime, nil).Once()
				return runtimeSvc
			},
			RuntimeCtxSvcFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", contextThatHasTenant(consumerTenantID), runtimeID, rtmCtxFilter, 100, "").Return(rtmCtxWithConsumerSubaccountLabel, nil).Once()
				return rtmCtxSvc
			},
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := txtest.PersistenceContextThatDoesntExpectCommit()
				persistTx.On("Commit").Return(testErr).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				return txtest.TransactionerThatDoesARollbackTwice(persistTx)
			},
			ApplicationTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(applicationTemplate, nil).Once()
				return appTemplateSvc
			},
			ApplicationServiceFn: func() *automock.ApplicationService {
				applicationSvc := &automock.ApplicationService{}
				applicationSvc.On("List", contextThatHasTenant(consumerTenantID), applicationFilter, 100, "").Return(emptyApplicationPage, nil).Once()
				return applicationSvc
			},
			ExpectedErr: "test",
		},
		// integration system
		{
			Name:   "Success for integration system consumer-provider flow: when subaccount tenant ID is provided instead of integration system ID for consumer ID",
			Claims: getClaimsForIntegrationSystemConsumerProviderFlow(consumerTenantID, consumerExtTenantID, providerExtTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
		},
		{
			Name: "Success for integration system consumer-provider flow: when integration system with consumer ID exists",
			IntegrationSystemServiceFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Exists", context.TODO(), consumerID).Return(true, nil)
				return intSysSvc
			},
			Claims: getClaimsForIntegrationSystemConsumerProviderFlow(consumerTenantID, consumerExtTenantID, consumerID, providerTenantID, providerExtTenantID, scopes, region, clientID),
		},
		{
			Name: "integration system consumer-provider flow: error when check for integration system existence fails",
			IntegrationSystemServiceFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Exists", context.TODO(), consumerID).Return(false, testErr)
				return intSysSvc
			},
			Claims:      getClaimsForIntegrationSystemConsumerProviderFlow(consumerTenantID, consumerExtTenantID, consumerID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			ExpectedErr: testErr.Error(),
		},
		{
			Name: "integration system consumer-provider flow: error when integration system with consumer ID does not exist",
			IntegrationSystemServiceFn: func() *automock.IntegrationSystemService {
				intSysSvc := &automock.IntegrationSystemService{}
				intSysSvc.On("Exists", context.TODO(), consumerID).Return(false, nil)
				return intSysSvc
			},
			Claims:      getClaimsForIntegrationSystemConsumerProviderFlow(consumerTenantID, consumerExtTenantID, consumerID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			ExpectedErr: fmt.Sprintf("integration system with ID %s does not exist", consumerID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			runtimeSvc := &automock.RuntimeService{}
			if testCase.RuntimeServiceFn != nil {
				runtimeSvc = testCase.RuntimeServiceFn()
			}
			runtimeCtxSvc := &automock.RuntimeCtxService{}
			if testCase.RuntimeCtxSvcFn != nil {
				runtimeCtxSvc = testCase.RuntimeCtxSvcFn()
			}
			appTemplateSvc := &automock.ApplicationTemplateService{}
			if testCase.ApplicationTemplateServiceFn != nil {
				appTemplateSvc = testCase.ApplicationTemplateServiceFn()
			}
			applicationSvc := &automock.ApplicationService{}
			if testCase.ApplicationServiceFn != nil {
				applicationSvc = testCase.ApplicationServiceFn()
			}
			intSysSvc := &automock.IntegrationSystemService{}
			if testCase.IntegrationSystemServiceFn != nil {
				intSysSvc = testCase.IntegrationSystemServiceFn()
			}
			persistTxMock := txtest.PersistenceContextThatExpectsCommit()
			if testCase.PersistenceFn != nil {
				persistTxMock = testCase.PersistenceFn()
			}
			transactionerMock := txtest.TransactionerThatSucceedsTwice(persistTxMock)
			if testCase.TransactionerFn != nil {
				transactionerMock = testCase.TransactionerFn(persistTxMock)
			}

			validator := claims.NewValidator(transactionerMock, runtimeSvc, runtimeCtxSvc, appTemplateSvc, applicationSvc, intSysSvc, providerLabelKey, consumerSubaccountLabelKey, tokenPrefix)
			err := validator.Validate(context.TODO(), testCase.Claims)

			if len(testCase.ExpectedErr) > 0 {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, runtimeSvc, runtimeCtxSvc, intSysSvc)
		})
	}
}

func TestScopesValidator_Validate(t *testing.T) {
	t.Run("Succeeds when all claims properties are present", func(t *testing.T) {
		v := claims.NewScopesValidator([]string{"application:read"})
		c := getClaims(consumerTenantID, consumerExtTenantID, scopes)

		err := v.Validate(context.TODO(), c)
		assert.NoError(t, err)
	})
	t.Run("Fails when no scopes are present", func(t *testing.T) {
		requiredScopes := []string{"application:read"}
		v := claims.NewScopesValidator(requiredScopes)
		c := getClaims(consumerTenantID, consumerExtTenantID, "")

		err := v.Validate(context.TODO(), c)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("Not all required scopes %q were found in claim with scopes %q", requiredScopes, c.Scopes))
	})
	t.Run("Fails when inner validation fails", func(t *testing.T) {
		requiredScopes := []string{"application:read"}
		v := claims.NewScopesValidator(requiredScopes)
		c := getClaims(consumerTenantID, consumerExtTenantID, scopes)
		c.ExpiresAt = 1

		err := v.Validate(context.TODO(), c)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while validating claims")
	})
}

func getClaimsForRuntimeConsumerProviderFlow(consumerTenant, consumerExternalTenant, providerTenant, providerExtTenant, scopes, region, clientID string) claims.Claims {
	return getClaimsForConsumerProviderFlow(consumer.Runtime, consumerTenant, consumerExternalTenant, consumerID, providerTenant, providerExtTenant, scopes, region, clientID)
}

func getClaimsForIntegrationSystemConsumerProviderFlow(consumerTenant, consumerExternalTenant, consumerID, providerTenant, providerExtTenant, scopes, region, clientID string) claims.Claims {
	return getClaimsForConsumerProviderFlow(consumer.IntegrationSystem, consumerTenant, consumerExternalTenant, consumerID, providerTenant, providerExtTenant, scopes, region, clientID)
}

func getClaimsForConsumerProviderFlow(consumerType consumer.ConsumerType, consumerTenant, consumerExternalTenant, consumerID, providerTenant, providerExtTenant, scopes, region, clientID string) claims.Claims {
	return claims.Claims{
		Tenant: map[string]string{
			tenantmapping.ConsumerTenantKey:         consumerTenant,
			tenantmapping.ExternalTenantKey:         consumerExternalTenant,
			tenantmapping.ProviderTenantKey:         providerTenant,
			tenantmapping.ProviderExternalTenantKey: providerExtTenant,
		},
		Scopes:        scopes,
		ConsumerID:    consumerID,
		ConsumerType:  consumerType,
		OnBehalfOf:    onBehalfOfConsumerID,
		Region:        region,
		TokenClientID: clientID,
	}
}

func getClaims(intTenantID, extTenantID, scopes string) claims.Claims {
	return claims.Claims{
		Tenant: map[string]string{
			tenantmapping.ConsumerTenantKey: intTenantID,
			tenantmapping.ExternalTenantKey: extTenantID,
		},

		Scopes:       scopes,
		ConsumerID:   consumerID,
		ConsumerType: consumer.Runtime,
	}
}

func contextThatHasTenant(expectedTenant string) interface{} {
	return mock.MatchedBy(func(actual context.Context) bool {
		actualTenant, err := tenant.LoadFromContext(actual)
		if err != nil {
			return false
		}
		return actualTenant == expectedTenant
	})
}
