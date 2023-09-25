package ord_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOperationMaintainer_Maintain(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testErr)

	webhooks := []*model.Webhook{
		{
			ID:         "wh-id-1",
			ObjectID:   "app-id",
			ObjectType: model.ApplicationWebhookReference,
			Type:       model.WebhookTypeOpenResourceDiscovery,
			URL:        str.Ptr("https://test.com"),
			Auth:       nil,
		},
		{
			ID:         "wh-id-2",
			ObjectID:   "app-template-id",
			ObjectType: model.ApplicationTemplateWebhookReference,
			Type:       model.WebhookTypeOpenResourceDiscovery,
			URL:        str.Ptr("https://test.com"),
			Auth:       nil,
		},
	}
	staticWebhooks := []*model.Webhook{
		{
			ID:         "wh-static-id-1",
			ObjectID:   "app-template-id",
			ObjectType: model.ApplicationTemplateWebhookReference,
			Type:       model.WebhookTypeOpenResourceDiscoveryStatic,
			URL:        str.Ptr("https://test-static.com"),
			Auth:       nil,
		},
	}
	operation := &model.Operation{
		ID:     "op-id",
		OpType: "",
		Status: "",
		Data:   json.RawMessage("{}"),
	}
	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		OpSvcFn         func() *automock.OperationService
		WebhookSvcFn    func() *automock.WebhookService
		AppSvcFn        func() *automock.ApplicationService
		ExpectedErr     error
	}{
		{
			Name: "Success",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceeds()
			},
			OpSvcFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("ListAllByType", txtest.CtxWithDBMatcher(), model.OperationTypeOrdAggregation).Return([]*model.Operation{operation}, nil).Once()
				svc.On("CreateMultiple", txtest.CtxWithDBMatcher(), mock.AnythingOfType("[]*model.OperationInput")).Run(func(args mock.Arguments) {
					arg := args.Get(1)
					res, ok := arg.([]*model.OperationInput)
					if !ok {
						return
					}
					assert.Equal(t, 3, len(res))
				}).Return(nil).Once()
				svc.On("DeleteMultiple", txtest.CtxWithDBMatcher(), mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(1)
					res, ok := arg.([]string)
					if !ok {
						return
					}
					assert.Equal(t, 1, len(res))
				}).Return(nil).Once()
				return svc
			},
			WebhookSvcFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(webhooks, nil).Once()
				svc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscoveryStatic).Return(staticWebhooks, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				apps := []*model.Application{
					{
						BaseEntity: &model.BaseEntity{
							ID: "app-id",
						},
					},
				}
				svc := &automock.ApplicationService{}
				svc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), "app-template-id").Return(apps, nil).Once()
				return svc
			},
		},
		{
			Name: "Error while beginning transaction",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			OpSvcFn: func() *automock.OperationService {
				return &automock.OperationService{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			AppSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while listing webhooks by type",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			OpSvcFn: func() *automock.OperationService {
				return &automock.OperationService{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(nil, testErr).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while listing all application from app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			OpSvcFn: func() *automock.OperationService {
				return &automock.OperationService{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(webhooks, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), "app-template-id").Return(nil, testErr).Once()
				return svc
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			_, tx := testCase.TransactionerFn()
			opSvc := testCase.OpSvcFn()
			webhookSvc := testCase.WebhookSvcFn()
			appSvc := testCase.AppSvcFn()

			opMaintainer := ord.NewOperationMaintainer(model.OperationTypeOrdAggregation, tx, opSvc, webhookSvc, appSvc)

			// WHEN
			err := opMaintainer.Maintain(ctx)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, opSvc, webhookSvc, appSvc)
		})
	}
}
