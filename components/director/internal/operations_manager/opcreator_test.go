package operationsmanager_test

//func TestORDOperationCreator_Create(t *testing.T) {
//	// GIVEN
//	testErr := errors.New("Test error")
//	ctx := context.TODO()
//	txGen := txtest.NewTransactionContextGenerator(testErr)
//
//	webhooks := []*model.Webhook{
//		{
//			ID:         "wh-id-1",
//			ObjectID:   "app-id",
//			ObjectType: model.ApplicationWebhookReference,
//			Type:       model.WebhookTypeOpenResourceDiscovery,
//			URL:        str.Ptr("https://test.com"),
//			Auth:       nil,
//		},
//		{
//			ID:         "wh-id-2",
//			ObjectID:   "app-template-id",
//			ObjectType: model.ApplicationTemplateWebhookReference,
//			Type:       model.WebhookTypeOpenResourceDiscovery,
//			URL:        str.Ptr("https://test.com"),
//			Auth:       nil,
//		},
//	}
//	testCases := []struct {
//		Name            string
//		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
//		OpSvcFn         func() *automock.OperationService
//		WebhookSvcFn    func() *automock.WebhookService
//		AppSvcFn        func() *automock.ApplicationService
//		ExpectedErr     error
//	}{
//		{
//			Name: "Success",
//			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				return txGen.ThatSucceedsMultipleTimes(3)
//			},
//			OpSvcFn: func() *automock.OperationService {
//				svc := &automock.OperationService{}
//				svc.On("CreateMultiple", txtest.CtxWithDBMatcher(), mock.AnythingOfType("[]*model.OperationInput")).Return(nil).Once()
//				return svc
//			},
//			WebhookSvcFn: func() *automock.WebhookService {
//				svc := &automock.WebhookService{}
//				svc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(webhooks, nil).Once()
//				return svc
//			},
//			AppSvcFn: func() *automock.ApplicationService {
//				apps := []*model.Application{
//					{
//						BaseEntity: &model.BaseEntity{
//							ID: "app-id",
//						},
//					},
//				}
//				svc := &automock.ApplicationService{}
//				svc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), "app-template-id").Return(apps, nil).Once()
//				return svc
//			},
//		},
//		{
//			Name: "Error while getting webhooks with ord type",
//			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				return txGen.ThatDoesntExpectCommit()
//			},
//			OpSvcFn: func() *automock.OperationService {
//				return &automock.OperationService{}
//			},
//			WebhookSvcFn: func() *automock.WebhookService {
//				svc := &automock.WebhookService{}
//				svc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(nil, testErr).Once()
//				return svc
//			},
//			AppSvcFn: func() *automock.ApplicationService {
//				return &automock.ApplicationService{}
//			},
//			ExpectedErr: testErr,
//		},
//		{
//			Name: "Error while beginning transaction on getting webhooks with ord type",
//			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				return txGen.ThatFailsOnBegin()
//			},
//			OpSvcFn: func() *automock.OperationService {
//				return &automock.OperationService{}
//			},
//			WebhookSvcFn: func() *automock.WebhookService {
//				return &automock.WebhookService{}
//			},
//			AppSvcFn: func() *automock.ApplicationService {
//				return &automock.ApplicationService{}
//			},
//			ExpectedErr: testErr,
//		},
//		{
//			Name: "Error while committing transaction on getting webhooks with ord type",
//			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				return txGen.ThatFailsOnCommit()
//			},
//			OpSvcFn: func() *automock.OperationService {
//				return &automock.OperationService{}
//			},
//			WebhookSvcFn: func() *automock.WebhookService {
//				svc := &automock.WebhookService{}
//				svc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(webhooks, nil).Once()
//				return svc
//			},
//			AppSvcFn: func() *automock.ApplicationService {
//				return &automock.ApplicationService{}
//			},
//			ExpectedErr: testErr,
//		},
//		{
//			Name: "Error while creating multiple operations",
//			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				return txGen.ThatSucceedsMultipleTimes(3)
//			},
//			OpSvcFn: func() *automock.OperationService {
//				svc := &automock.OperationService{}
//				svc.On("CreateMultiple", txtest.CtxWithDBMatcher(), mock.AnythingOfType("[]*model.OperationInput")).Return(testErr).Once()
//				return svc
//			},
//			WebhookSvcFn: func() *automock.WebhookService {
//				svc := &automock.WebhookService{}
//				svc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(webhooks, nil).Once()
//				return svc
//			},
//			AppSvcFn: func() *automock.ApplicationService {
//				apps := []*model.Application{
//					{
//						BaseEntity: &model.BaseEntity{
//							ID: "app-id",
//						},
//					},
//				}
//				svc := &automock.ApplicationService{}
//				svc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), "app-template-id").Return(apps, nil).Once()
//				return svc
//			},
//			ExpectedErr: testErr,
//		},
//		{
//			Name: "Error while listing applications by app template id",
//			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
//				return txGen.ThatSucceedsMultipleTimes(2)
//			},
//			OpSvcFn: func() *automock.OperationService {
//				return &automock.OperationService{}
//			},
//			WebhookSvcFn: func() *automock.WebhookService {
//				svc := &automock.WebhookService{}
//				svc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(webhooks, nil).Once()
//				return svc
//			},
//			AppSvcFn: func() *automock.ApplicationService {
//				svc := &automock.ApplicationService{}
//				svc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), "app-template-id").Return(nil, testErr).Once()
//				return svc
//			},
//			ExpectedErr: testErr,
//		},
//	}
//
//	for _, testCase := range testCases {
//		t.Run(testCase.Name, func(t *testing.T) {
//			// GIVEN
//			_, tx := testCase.TransactionerFn()
//			opSvc := testCase.OpSvcFn()
//			webhookSvc := testCase.WebhookSvcFn()
//			appSvc := testCase.AppSvcFn()
//
//			opCreator := operationsmanager.NewOperationCreator(operationsmanager.OrdCreatorType, tx, opSvc, webhookSvc, appSvc)
//
//			// WHEN
//			err := opCreator.Create(ctx)
//
//			// THEN
//			if testCase.ExpectedErr != nil {
//				require.Error(t, err)
//				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
//			} else {
//				assert.Nil(t, err)
//			}
//
//			mock.AssertExpectationsForObjects(t, tx, opSvc, webhookSvc, appSvc)
//		})
//	}
//}
