package ord_test

import (
	"context"
	resource2 "github.com/kyma-incubator/compass/components/director/pkg/resource"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_SyncGlobalResources(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	resource := ord.Resource{
		Name:          "global-registry",
		Type:          resource2.Application,
		ID:            "global-registry",
		LocalTenantID: nil,
	}
	var emptyConf map[string]interface{}

	testWebhook := &model.Webhook{
		Type: model.WebhookTypeOpenResourceDiscovery,
		URL:  str.Ptr(baseURL),
	}

	doc := fixGlobalRegistryORDDocument()

	successfulVendorUpdate := func() *automock.GlobalVendorService {
		vendorSvc := &automock.GlobalVendorService{}
		vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalVendors(), nil).Once()
		vendorSvc.On("UpdateGlobal", txtest.CtxWithDBMatcher(), vendorID, *doc.Vendors[0]).Return(nil).Once()
		vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalVendors(), nil).Once()
		return vendorSvc
	}

	successfulVendorCreate := func() *automock.GlobalVendorService {
		vendorSvc := &automock.GlobalVendorService{}
		vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
		vendorSvc.On("CreateGlobal", txtest.CtxWithDBMatcher(), *doc.Vendors[0]).Return("", nil).Once()
		vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalVendors(), nil).Once()
		return vendorSvc
	}

	successfulProductUpdate := func() *automock.GlobalProductService {
		productSvc := &automock.GlobalProductService{}
		productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalProducts(), nil).Once()
		productSvc.On("UpdateGlobal", txtest.CtxWithDBMatcher(), productID, *doc.Products[0]).Return(nil).Once()
		productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalProducts(), nil).Once()
		return productSvc
	}

	successfulProductCreate := func() *automock.GlobalProductService {
		productSvc := &automock.GlobalProductService{}
		productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
		productSvc.On("CreateGlobal", txtest.CtxWithDBMatcher(), *doc.Products[0]).Return("", nil).Once()
		productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalProducts(), nil).Once()
		return productSvc
	}

	successfulClientFn := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", context.TODO(), resource, testWebhook, emptyConf).Return(ord.Documents{fixGlobalRegistryORDDocument()}, baseURL, nil)
		return client
	}

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		productSvcFn    func() *automock.GlobalProductService
		vendorSvcFn     func() *automock.GlobalVendorService
		clientFn        func() *automock.Client
		ExpectedErr     error
	}{
		{
			Name:            "Success when resources are not in db should Create them",
			TransactionerFn: txGen.ThatSucceeds,
			productSvcFn:    successfulProductCreate,
			vendorSvcFn:     successfulVendorCreate,
			clientFn:        successfulClientFn,
		},
		{
			Name:            "Success when resources are in db should Update them",
			TransactionerFn: txGen.ThatSucceeds,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			clientFn:        successfulClientFn,
		},
		{
			Name:            "Success when resources are in db should Update them and delete all global resources that are not returned anymore",
			TransactionerFn: txGen.ThatSucceeds,
			productSvcFn: func() *automock.GlobalProductService {
				productSvc := &automock.GlobalProductService{}
				products := fixGlobalProducts()
				products = append(products, &model.Product{ID: "product-id-2", OrdID: "test:product:TEST:"})
				productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(products, nil).Once()
				productSvc.On("UpdateGlobal", txtest.CtxWithDBMatcher(), productID, *doc.Products[0]).Return(nil).Once()
				productSvc.On("DeleteGlobal", txtest.CtxWithDBMatcher(), "product-id-2").Return(nil).Once()
				productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalProducts(), nil).Once()
				return productSvc
			},
			vendorSvcFn: func() *automock.GlobalVendorService {
				vendorSvc := &automock.GlobalVendorService{}
				vendors := fixGlobalVendors()
				vendors = append(vendors, &model.Vendor{ID: "vendor-id-2", OrdID: "test:vendor:TEST:"})
				vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(vendors, nil).Once()
				vendorSvc.On("UpdateGlobal", txtest.CtxWithDBMatcher(), vendorID, *doc.Vendors[0]).Return(nil).Once()
				vendorSvc.On("DeleteGlobal", txtest.CtxWithDBMatcher(), "vendor-id-2").Return(nil).Once()
				vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalVendors(), nil).Once()
				return vendorSvc
			},
			clientFn: successfulClientFn,
		},
		{
			Name:            "Error when fetch ord docs fail",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("FetchOpenResourceDiscoveryDocuments", context.TODO(), resource, testWebhook, emptyConf).Return(nil, "", testErr)
				return client
			},
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when ord docs are invalid",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixGlobalRegistryORDDocument()
				doc.Vendors[0].OrdID = "invalid-ord-id"
				client.On("FetchOpenResourceDiscoveryDocuments", context.TODO(), resource, testWebhook, emptyConf).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			ExpectedErr: errors.New("ordId: must be in a valid format."),
		},
		{
			Name:            "Error when ord docs contains resource that is not vendor or product",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixGlobalRegistryORDDocument()
				doc.ConsumptionBundles = fixORDDocument().ConsumptionBundles
				client.On("FetchOpenResourceDiscoveryDocuments", context.TODO(), resource, testWebhook, emptyConf).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			ExpectedErr: errors.New("global registry supports only vendors and products"),
		},
		{
			Name:            "Error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			clientFn:        successfulClientFn,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Error when vendor list fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			vendorSvcFn: func() *automock.GlobalVendorService {
				vendorSvc := &automock.GlobalVendorService{}
				vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(nil, testErr).Once()
				return vendorSvc
			},
			clientFn:    successfulClientFn,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when vendor create fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			vendorSvcFn: func() *automock.GlobalVendorService {
				vendorSvc := &automock.GlobalVendorService{}
				vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				vendorSvc.On("CreateGlobal", txtest.CtxWithDBMatcher(), *doc.Vendors[0]).Return("", testErr).Once()
				return vendorSvc
			},
			clientFn:    successfulClientFn,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when vendor update fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			vendorSvcFn: func() *automock.GlobalVendorService {
				vendorSvc := &automock.GlobalVendorService{}
				vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalVendors(), nil).Once()
				vendorSvc.On("UpdateGlobal", txtest.CtxWithDBMatcher(), vendorID, *doc.Vendors[0]).Return(testErr).Once()
				return vendorSvc
			},
			clientFn:    successfulClientFn,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when vendor delete fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			vendorSvcFn: func() *automock.GlobalVendorService {
				vendorSvc := &automock.GlobalVendorService{}
				vendors := fixGlobalVendors()
				vendors = append(vendors, &model.Vendor{ID: "vendor-id-2", OrdID: "test:vendor:TEST:"})
				vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(vendors, nil).Once()
				vendorSvc.On("UpdateGlobal", txtest.CtxWithDBMatcher(), vendorID, *doc.Vendors[0]).Return(nil).Once()
				vendorSvc.On("DeleteGlobal", txtest.CtxWithDBMatcher(), "vendor-id-2").Return(testErr).Once()
				return vendorSvc
			},
			clientFn:    successfulClientFn,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when second vendor list fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			vendorSvcFn: func() *automock.GlobalVendorService {
				vendorSvc := &automock.GlobalVendorService{}
				vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				vendorSvc.On("CreateGlobal", txtest.CtxWithDBMatcher(), *doc.Vendors[0]).Return("", nil).Once()
				vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(nil, testErr).Once()
				return vendorSvc
			},
			clientFn:    successfulClientFn,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when product list fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			vendorSvcFn:     successfulVendorCreate,
			productSvcFn: func() *automock.GlobalProductService {
				productSvc := &automock.GlobalProductService{}
				productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(nil, testErr).Once()
				return productSvc
			},
			clientFn:    successfulClientFn,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when product create fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			vendorSvcFn:     successfulVendorCreate,
			productSvcFn: func() *automock.GlobalProductService {
				productSvc := &automock.GlobalProductService{}
				productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				productSvc.On("CreateGlobal", txtest.CtxWithDBMatcher(), *doc.Products[0]).Return("", testErr).Once()
				return productSvc
			},
			clientFn:    successfulClientFn,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when product update fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			vendorSvcFn:     successfulVendorCreate,
			productSvcFn: func() *automock.GlobalProductService {
				productSvc := &automock.GlobalProductService{}
				productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalProducts(), nil).Once()
				productSvc.On("UpdateGlobal", txtest.CtxWithDBMatcher(), productID, *doc.Products[0]).Return(testErr).Once()
				return productSvc
			},
			clientFn:    successfulClientFn,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when product delete fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			vendorSvcFn:     successfulVendorCreate,
			productSvcFn: func() *automock.GlobalProductService {
				productSvc := &automock.GlobalProductService{}
				products := fixGlobalProducts()
				products = append(products, &model.Product{ID: "product-id-2", OrdID: "test:product:TEST:"})
				productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(products, nil).Once()
				productSvc.On("UpdateGlobal", txtest.CtxWithDBMatcher(), productID, *doc.Products[0]).Return(nil).Once()
				productSvc.On("DeleteGlobal", txtest.CtxWithDBMatcher(), "product-id-2").Return(testErr).Once()
				return productSvc
			},
			clientFn:    successfulClientFn,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when second product list fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			vendorSvcFn:     successfulVendorCreate,
			productSvcFn: func() *automock.GlobalProductService {
				productSvc := &automock.GlobalProductService{}
				productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				productSvc.On("CreateGlobal", txtest.CtxWithDBMatcher(), *doc.Products[0]).Return("", nil).Once()
				productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalProducts(), testErr).Once()
				return productSvc
			},
			clientFn:    successfulClientFn,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when transaction commit fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			productSvcFn:    successfulProductCreate,
			vendorSvcFn:     successfulVendorCreate,
			clientFn:        successfulClientFn,
			ExpectedErr:     testErr,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()
			productSvc := &automock.GlobalProductService{}
			if test.productSvcFn != nil {
				productSvc = test.productSvcFn()
			}
			vendorSvc := &automock.GlobalVendorService{}
			if test.vendorSvcFn != nil {
				vendorSvc = test.vendorSvcFn()
			}
			client := &automock.Client{}
			if test.clientFn != nil {
				client = test.clientFn()
			}

			svc := ord.NewGlobalRegistryService(tx, ord.GlobalRegistryConfig{URL: baseURL}, vendorSvc, productSvc, client, credentialExchangeStrategyTenantMappings)
			globalIDs, err := svc.SyncGlobalResources(context.TODO())
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Len(t, globalIDs, 2)
				require.True(t, globalIDs[vendorORDID])
				require.True(t, globalIDs[globalProductORDID])
			}

			mock.AssertExpectationsForObjects(t, tx, productSvc, vendorSvc, client)
		})
	}
}

func TestService_ListGlobalResources(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	successfulVendorList := func() *automock.GlobalVendorService {
		vendorSvc := &automock.GlobalVendorService{}
		vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalVendors(), nil).Once()
		return vendorSvc
	}

	successfulProductList := func() *automock.GlobalProductService {
		productSvc := &automock.GlobalProductService{}
		productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(fixGlobalProducts(), nil).Once()
		return productSvc
	}

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		productSvcFn    func() *automock.GlobalProductService
		vendorSvcFn     func() *automock.GlobalVendorService
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			productSvcFn:    successfulProductList,
			vendorSvcFn:     successfulVendorList,
		},
		{
			Name:            "Error when vendor list fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			vendorSvcFn: func() *automock.GlobalVendorService {
				vendorSvc := &automock.GlobalVendorService{}
				vendorSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(nil, testErr).Once()
				return vendorSvc
			},
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when product list fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			vendorSvcFn:     successfulVendorList,
			productSvcFn: func() *automock.GlobalProductService {
				productSvc := &automock.GlobalProductService{}
				productSvc.On("ListGlobal", txtest.CtxWithDBMatcher()).Return(nil, testErr).Once()
				return productSvc
			},
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when transaction commit fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			productSvcFn:    successfulProductList,
			vendorSvcFn:     successfulVendorList,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Error when transaction begin fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ExpectedErr:     testErr,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()
			productSvc := &automock.GlobalProductService{}
			if test.productSvcFn != nil {
				productSvc = test.productSvcFn()
			}
			vendorSvc := &automock.GlobalVendorService{}
			if test.vendorSvcFn != nil {
				vendorSvc = test.vendorSvcFn()
			}
			client := &automock.Client{}

			svc := ord.NewGlobalRegistryService(tx, ord.GlobalRegistryConfig{URL: baseURL}, vendorSvc, productSvc, client, credentialExchangeStrategyTenantMappings)
			globalIDs, err := svc.ListGlobalResources(context.TODO())
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Len(t, globalIDs, 2)
				require.True(t, globalIDs[vendorORDID])
				require.True(t, globalIDs[globalProductORDID])
			}

			mock.AssertExpectationsForObjects(t, tx, productSvc, vendorSvc, client)
		})
	}
}
