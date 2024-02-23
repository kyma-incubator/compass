package tenantfetchersvc_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tenantID         = "testTenant"
	externalTenantID = "externalTestTenant"
)

func TestWebhookProcessor_ProcessWebhooks(t *testing.T) {
	testError = errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testError)

	timestamp := time.Now()
	whID := "whID"
	appID := "appID"
	saasRegURL := "https://saas-reg-url.com"
	modelWebhooks := []*model.Webhook{
		{
			ID:         whID,
			ObjectID:   appID,
			ObjectType: model.ApplicationWebhookReference,
			Type:       model.WebhookTypeSystemFieldDiscovery,
			CreatedAt:  &timestamp,
			URL:        str.Ptr(saasRegURL),
			Auth: &model.Auth{
				Credential: model.CredentialData{
					Oauth: &model.OAuthCredentialData{
						ClientID:     "clientid",
						ClientSecret: "clientsecret",
						URL:          "https://tokenUrl/oauth/token",
					},
				},
			},
		},
	}

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppSvcFn        func() *automock.ApplicationService
		WebhookSvcFn    func() *automock.WebhookService
		TenantSvcFn     func() *automock.TenantService
		WebhookClient   *http.Client
		ExpectedError   error
	}{
		{
			Name: "Success",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsTwice()
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("UpdateBaseURLAndReadyState", txtest.CtxWithDBMatcher(), appID, "https://new-url.com", true).Return(nil).Once()
				return appSvc
			},
			WebhookSvcFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListByTypeAndLabelFilter", txtest.CtxWithDBMatcher(), model.WebhookTypeSystemFieldDiscovery, labelfilter.NewForKeyWithQuery(tenantfetchersvc.RegistryLabelKey, fmt.Sprintf("\"%s\"", tenantfetchersvc.SaaSRegistryLabelValue))).Return(modelWebhooks, nil).Once()
				webhookSvc.On("Delete", txtest.CtxWithDBMatcher(), whID, model.ApplicationWebhookReference).Return(nil).Once()
				return webhookSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ID: tenantID, ExternalTenant: externalTenantID}, nil).Once()
				return tenantSvc
			},
			WebhookClient: createWebhookClient(t),
			ExpectedError: nil,
		},
		{
			Name: "Fails when getting tenant by id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsTwice()
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "UpdateBaseURLAndReadyState")
				return appSvc
			},
			WebhookSvcFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListByTypeAndLabelFilter", txtest.CtxWithDBMatcher(), model.WebhookTypeSystemFieldDiscovery, labelfilter.NewForKeyWithQuery(tenantfetchersvc.RegistryLabelKey, fmt.Sprintf("\"%s\"", tenantfetchersvc.SaaSRegistryLabelValue))).Return(modelWebhooks, nil).Once()
				webhookSvc.AssertNotCalled(t, "Delete")
				return webhookSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(nil, testError).Once()
				return tenantSvc
			},
			WebhookClient: createWebhookClient(t),
			ExpectedError: nil,
		},
		{
			Name: "Fails when getting lowest owner for resources",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(1)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Times(1)

				return persistTx, transact
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "UpdateBaseURLAndReadyState")
				return appSvc
			},
			WebhookSvcFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListByTypeAndLabelFilter", txtest.CtxWithDBMatcher(), model.WebhookTypeSystemFieldDiscovery, labelfilter.NewForKeyWithQuery(tenantfetchersvc.RegistryLabelKey, fmt.Sprintf("\"%s\"", tenantfetchersvc.SaaSRegistryLabelValue))).Return(modelWebhooks, nil).Once()
				webhookSvc.AssertNotCalled(t, "Delete")
				return webhookSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return("", testError).Once()
				tenantSvc.AssertNotCalled(t, "GetTenantByID")
				return tenantSvc
			},
			WebhookClient: createWebhookClient(t),
			ExpectedError: nil,
		},
		{
			Name: "Fails when updating application base url and ready state",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(1)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Times(1)

				return persistTx, transact
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("UpdateBaseURLAndReadyState", txtest.CtxWithDBMatcher(), appID, "https://new-url.com", true).Return(testError).Once()
				return appSvc
			},
			WebhookSvcFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListByTypeAndLabelFilter", txtest.CtxWithDBMatcher(), model.WebhookTypeSystemFieldDiscovery, labelfilter.NewForKeyWithQuery(tenantfetchersvc.RegistryLabelKey, fmt.Sprintf("\"%s\"", tenantfetchersvc.SaaSRegistryLabelValue))).Return(modelWebhooks, nil).Once()
				webhookSvc.AssertNotCalled(t, "Delete")
				return webhookSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ID: tenantID, ExternalTenant: externalTenantID}, nil).Once()
				return tenantSvc
			},
			WebhookClient: createWebhookClient(t),
			ExpectedError: nil,
		},
		{
			Name: "Fails when deleting webhook",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(1)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Times(1)

				return persistTx, transact
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("UpdateBaseURLAndReadyState", txtest.CtxWithDBMatcher(), appID, "https://new-url.com", true).Return(nil).Once()
				return appSvc
			},
			WebhookSvcFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListByTypeAndLabelFilter", txtest.CtxWithDBMatcher(), model.WebhookTypeSystemFieldDiscovery, labelfilter.NewForKeyWithQuery(tenantfetchersvc.RegistryLabelKey, fmt.Sprintf("\"%s\"", tenantfetchersvc.SaaSRegistryLabelValue))).Return(modelWebhooks, nil).Once()
				webhookSvc.On("Delete", txtest.CtxWithDBMatcher(), whID, model.ApplicationWebhookReference).Return(testError).Once()
				return webhookSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ID: tenantID, ExternalTenant: externalTenantID}, nil).Once()
				return tenantSvc
			},
			WebhookClient: createWebhookClient(t),
			ExpectedError: nil,
		},
		{
			Name: "Returns error when listing webhooks fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "UpdateBaseURLAndReadyState")
				return appSvc
			},
			WebhookSvcFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListByTypeAndLabelFilter", txtest.CtxWithDBMatcher(), model.WebhookTypeSystemFieldDiscovery, labelfilter.NewForKeyWithQuery(tenantfetchersvc.RegistryLabelKey, fmt.Sprintf("\"%s\"", tenantfetchersvc.SaaSRegistryLabelValue))).Return(nil, testError).Once()
				webhookSvc.AssertNotCalled(t, "Delete")
				return webhookSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.AssertNotCalled(t, "GetLowestOwnerForResource")
				tenantSvc.AssertNotCalled(t, "GetTenantByID")
				return tenantSvc
			},
			WebhookClient: createWebhookClient(t),
			ExpectedError: testError,
		},
		{
			Name: "Success when webhook is older, so it does not get processed",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceeds()
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "UpdateBaseURLAndReadyState")
				return appSvc
			},
			WebhookSvcFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				oldWebhooks := modelWebhooks
				ts := time.Now().AddDate(0, 0, -8)
				oldWebhooks[0].CreatedAt = &ts
				webhookSvc.On("ListByTypeAndLabelFilter", txtest.CtxWithDBMatcher(), model.WebhookTypeSystemFieldDiscovery, labelfilter.NewForKeyWithQuery(tenantfetchersvc.RegistryLabelKey, fmt.Sprintf("\"%s\"", tenantfetchersvc.SaaSRegistryLabelValue))).Return(oldWebhooks, nil).Once()
				webhookSvc.AssertNotCalled(t, "Delete")
				return webhookSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.AssertNotCalled(t, "GetLowestOwnerForResource")
				tenantSvc.AssertNotCalled(t, "GetTenantByID")
				return tenantSvc
			},
			WebhookClient: createWebhookClient(t),
			ExpectedError: nil,
		},

		{
			Name: "Success when webhook does not have credentials, so it does not get processed",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceeds()
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "UpdateBaseURLAndReadyState")
				return appSvc
			},
			WebhookSvcFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				noCredsWebhooks := modelWebhooks
				noCredsWebhooks[0].Auth = nil
				webhookSvc.On("ListByTypeAndLabelFilter", txtest.CtxWithDBMatcher(), model.WebhookTypeSystemFieldDiscovery, labelfilter.NewForKeyWithQuery(tenantfetchersvc.RegistryLabelKey, fmt.Sprintf("\"%s\"", tenantfetchersvc.SaaSRegistryLabelValue))).Return(noCredsWebhooks, nil).Once()
				webhookSvc.AssertNotCalled(t, "Delete")
				return webhookSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.AssertNotCalled(t, "GetLowestOwnerForResource")
				tenantSvc.AssertNotCalled(t, "GetTenantByID")
				return tenantSvc
			},
			WebhookClient: createWebhookClient(t),
			ExpectedError: nil,
		},
		{
			Name: "Fails when executing webhook",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceeds()
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "UpdateBaseURLAndReadyState")
				return appSvc
			},
			WebhookSvcFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListByTypeAndLabelFilter", txtest.CtxWithDBMatcher(), model.WebhookTypeSystemFieldDiscovery, labelfilter.NewForKeyWithQuery(tenantfetchersvc.RegistryLabelKey, fmt.Sprintf("\"%s\"", tenantfetchersvc.SaaSRegistryLabelValue))).Return(modelWebhooks, nil).Once()
				webhookSvc.AssertNotCalled(t, "Delete")
				return webhookSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.AssertNotCalled(t, "GetLowestOwnerForResource")
				tenantSvc.AssertNotCalled(t, "GetTenantByID")
				return tenantSvc
			},
			WebhookClient: createBadRequestWebhookClient(),
			ExpectedError: nil,
		},
		{
			Name: "Success when there is no app url in the response for webhook, so it does not get processed",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceeds()
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "UpdateBaseURLAndReadyState")
				return appSvc
			},
			WebhookSvcFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListByTypeAndLabelFilter", txtest.CtxWithDBMatcher(), model.WebhookTypeSystemFieldDiscovery, labelfilter.NewForKeyWithQuery(tenantfetchersvc.RegistryLabelKey, fmt.Sprintf("\"%s\"", tenantfetchersvc.SaaSRegistryLabelValue))).Return(modelWebhooks, nil).Once()
				webhookSvc.AssertNotCalled(t, "Delete")
				return webhookSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.AssertNotCalled(t, "GetLowestOwnerForResource")
				tenantSvc.AssertNotCalled(t, "GetTenantByID")
				return tenantSvc
			},
			WebhookClient: createNoAppURLWebhookClient(),
			ExpectedError: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, tx := testCase.TransactionerFn()
			appSvc := testCase.AppSvcFn()
			webhookSvc := testCase.WebhookSvcFn()
			tenantSvc := testCase.TenantSvcFn()
			webhookClient := testCase.WebhookClient

			webhookProcessor := tenantfetchersvc.NewWebhookProcessor(tx, webhookSvc, tenantSvc, appSvc, webhookClient, cronjob.ElectionConfig{}, 0, true, 7)
			err := webhookProcessor.ProcessWebhooks(context.TODO())

			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			tx.AssertExpectations(t)
			webhookSvc.AssertExpectations(t)
			appSvc.AssertExpectations(t)
			tenantSvc.AssertExpectations(t)
		})
	}
}

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func createWebhookClient(t *testing.T) *http.Client {
	return newTestClient(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, req.Method, http.MethodGet)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"subscriptions": [{"url": "https://new-url.com"}]}`)),
		}, nil
	})
}

func createBadRequestWebhookClient() *http.Client {
	return newTestClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader(`{"subscriptions": [{"url": "https://new-url.com"}]}`)),
		}, nil
	})
}

func createNoAppURLWebhookClient() *http.Client {
	return newTestClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"subscriptions": [{}]}`)),
		}, nil
	})
}
