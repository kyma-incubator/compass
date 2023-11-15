package processor_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTombstonedResourcesDeleter_Delete(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	tombstones := []*model.Tombstone{
		{
			OrdID: "tombstone-ord-id",
		},
	}

	packages := []*model.Package{
		{
			ID:    "package-id",
			OrdID: "tombstone-ord-id",
		},
	}

	apis := []*model.APIDefinition{
		{
			BaseEntity: &model.BaseEntity{
				ID: "api-id",
			},
			OrdID: str.Ptr("tombstone-ord-id"),
		},
	}

	events := []*model.EventDefinition{
		{
			BaseEntity: &model.BaseEntity{
				ID: "event-id",
			},
			OrdID: str.Ptr("tombstone-ord-id"),
		},
	}

	entityTypes := []*model.EntityType{
		{
			BaseEntity: &model.BaseEntity{
				ID: "entity-type-id",
			},
			OrdID: "tombstone-ord-id",
		},
	}

	capabilities := []*model.Capability{
		{
			BaseEntity: &model.BaseEntity{
				ID: "capability-id",
			},
			OrdID: str.Ptr("tombstone-ord-id"),
		},
	}

	integrationDependencies := []*model.IntegrationDependency{
		{
			BaseEntity: &model.BaseEntity{
				ID: "integration-dependency-id",
			},
			OrdID: str.Ptr("tombstone-ord-id"),
		},
	}

	bundles := []*model.Bundle{
		{
			BaseEntity: &model.BaseEntity{
				ID: "bundle-id",
			},
			OrdID: str.Ptr("tombstone-ord-id"),
		},
	}

	vendors := []*model.Vendor{
		{
			ID:    "vendor-id",
			OrdID: "tombstone-ord-id",
		},
	}

	productModels := []*model.Product{
		{
			ID:    "product-id",
			OrdID: "tombstone-ord-id",
		},
	}

	successfulPackageDelete := func() *automock.PackageService {
		packageSvc := &automock.PackageService{}
		packageSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "package-id").Return(nil).Once()
		return packageSvc
	}

	successfulAPIDelete := func() *automock.APIService {
		apiSvc := &automock.APIService{}
		apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "api-id").Return(nil).Once()
		return apiSvc
	}

	successfulEventDelete := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "event-id").Return(nil).Once()
		return eventSvc
	}

	successfulEntityTypeDelete := func() *automock.EntityTypeService {
		entityTypeSvc := &automock.EntityTypeService{}
		entityTypeSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "entity-type-id").Return(nil).Once()
		return entityTypeSvc
	}

	successfulCapabilityDelete := func() *automock.CapabilityService {
		capabilitySvc := &automock.CapabilityService{}
		capabilitySvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "capability-id").Return(nil).Once()
		return capabilitySvc
	}

	successfulIntegrationDependencyDelete := func() *automock.IntegrationDependencyService {
		integrationDependencySvc := &automock.IntegrationDependencyService{}
		integrationDependencySvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "integration-dependency-id").Return(nil).Once()
		return integrationDependencySvc
	}

	successfulBundleDelete := func() *automock.BundleService {
		bundleSvc := &automock.BundleService{}
		bundleSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "bundle-id").Return(nil).Once()
		return bundleSvc
	}

	successfulVendorDelete := func() *automock.VendorService {
		vendorSvc := &automock.VendorService{}
		vendorSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "vendor-id").Return(nil).Once()
		return vendorSvc
	}

	successfulProductDelete := func() *automock.ProductService {
		productSvc := &automock.ProductService{}
		productSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "product-id").Return(nil).Once()
		return productSvc
	}

	testCases := []struct {
		Name                         string
		TransactionerFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		PackageSvcFn                 func() *automock.PackageService
		APISvcFn                     func() *automock.APIService
		EventSvcFn                   func() *automock.EventService
		EntityTypeSvcFn              func() *automock.EntityTypeService
		CapabilitySvcFn              func() *automock.CapabilityService
		IntegrationDependencySvcFn   func() *automock.IntegrationDependencyService
		VendorSvcFn                  func() *automock.VendorService
		ProductSvcFn                 func() *automock.ProductService
		BundleSvcFn                  func() *automock.BundleService
		InputResource                resource.Type
		InputVendors                 []*model.Vendor
		InputProducts                []*model.Product
		InputPackages                []*model.Package
		InputBundles                 []*model.Bundle
		InputAPIs                    []*model.APIDefinition
		InputEvents                  []*model.EventDefinition
		InputEntityTypes             []*model.EntityType
		InputCapabilities            []*model.Capability
		InputIntegrationDependencies []*model.IntegrationDependency
		InputTombstones              []*model.Tombstone
		InputFetchRequests           []*processor.OrdFetchRequest
		ExpectedOutput               []*processor.OrdFetchRequest
		ExpectedErr                  error
	}{
		{
			Name: "Success",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceeds()
			},
			PackageSvcFn:                 successfulPackageDelete,
			APISvcFn:                     successfulAPIDelete,
			EventSvcFn:                   successfulEventDelete,
			EntityTypeSvcFn:              successfulEntityTypeDelete,
			CapabilitySvcFn:              successfulCapabilityDelete,
			IntegrationDependencySvcFn:   successfulIntegrationDependencyDelete,
			BundleSvcFn:                  successfulBundleDelete,
			VendorSvcFn:                  successfulVendorDelete,
			ProductSvcFn:                 successfulProductDelete,
			InputTombstones:              tombstones,
			InputPackages:                packages,
			InputAPIs:                    apis,
			InputEvents:                  events,
			InputEntityTypes:             entityTypes,
			InputCapabilities:            capabilities,
			InputIntegrationDependencies: integrationDependencies,
			InputBundles:                 bundles,
			InputVendors:                 vendors,
			InputProducts:                productModels,
			InputResource:                resource.Application,
			ExpectedOutput:               []*processor.OrdFetchRequest{},
		},
		{
			Name: "Success with fetch requests",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceeds()
			},
			PackageSvcFn:                 successfulPackageDelete,
			APISvcFn:                     successfulAPIDelete,
			EventSvcFn:                   successfulEventDelete,
			EntityTypeSvcFn:              successfulEntityTypeDelete,
			CapabilitySvcFn:              successfulCapabilityDelete,
			IntegrationDependencySvcFn:   successfulIntegrationDependencyDelete,
			BundleSvcFn:                  successfulBundleDelete,
			VendorSvcFn:                  successfulVendorDelete,
			ProductSvcFn:                 successfulProductDelete,
			InputTombstones:              tombstones,
			InputPackages:                packages,
			InputAPIs:                    apis,
			InputEvents:                  events,
			InputEntityTypes:             entityTypes,
			InputCapabilities:            capabilities,
			InputIntegrationDependencies: integrationDependencies,
			InputBundles:                 bundles,
			InputVendors:                 vendors,
			InputProducts:                productModels,
			InputResource:                resource.Application,
			InputFetchRequests: []*processor.OrdFetchRequest{
				{
					FetchRequest:   nil,
					RefObjectOrdID: "tombstone-ord-id",
				},
				{
					FetchRequest:   nil,
					RefObjectOrdID: "random-id",
				},
			},
			ExpectedOutput: []*processor.OrdFetchRequest{
				{
					FetchRequest:   nil,
					RefObjectOrdID: "random-id",
				},
			},
		},
		{
			Name: "Fail beginning a transaction",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Fail while deleting package",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			PackageSvcFn: func() *automock.PackageService {
				packageSvc := &automock.PackageService{}
				packageSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "package-id").Return(testErr).Once()
				return packageSvc
			},
			InputTombstones: tombstones,
			InputPackages:   packages,
			InputResource:   resource.Application,
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while deleting api",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			PackageSvcFn: successfulPackageDelete,
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "api-id").Return(testErr).Once()
				return apiSvc
			},
			InputTombstones: tombstones,
			InputPackages:   packages,
			InputAPIs:       apis,
			InputResource:   resource.Application,
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while deleting event",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			PackageSvcFn: successfulPackageDelete,
			APISvcFn:     successfulAPIDelete,
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "event-id").Return(testErr).Once()
				return eventSvc
			},
			InputTombstones: tombstones,
			InputPackages:   packages,
			InputAPIs:       apis,
			InputEvents:     events,
			InputResource:   resource.Application,
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while deleting entity type",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			PackageSvcFn: successfulPackageDelete,
			APISvcFn:     successfulAPIDelete,
			EventSvcFn:   successfulEventDelete,
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "entity-type-id").Return(testErr).Once()
				return entityTypeSvc
			},
			InputTombstones:  tombstones,
			InputPackages:    packages,
			InputAPIs:        apis,
			InputEvents:      events,
			InputEntityTypes: entityTypes,
			InputResource:    resource.Application,
			ExpectedErr:      testErr,
		},
		{
			Name: "Fail while deleting capability",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			PackageSvcFn:    successfulPackageDelete,
			APISvcFn:        successfulAPIDelete,
			EventSvcFn:      successfulEventDelete,
			EntityTypeSvcFn: successfulEntityTypeDelete,
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "capability-id").Return(testErr).Once()
				return capabilitySvc
			},
			InputTombstones:   tombstones,
			InputPackages:     packages,
			InputAPIs:         apis,
			InputEvents:       events,
			InputEntityTypes:  entityTypes,
			InputCapabilities: capabilities,
			InputResource:     resource.Application,
			ExpectedErr:       testErr,
		},
		{
			Name: "Fail while deleting integration dependency",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			PackageSvcFn:    successfulPackageDelete,
			APISvcFn:        successfulAPIDelete,
			EventSvcFn:      successfulEventDelete,
			EntityTypeSvcFn: successfulEntityTypeDelete,
			CapabilitySvcFn: successfulCapabilityDelete,
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "integration-dependency-id").Return(testErr).Once()
				return integrationDependencySvc
			},
			InputTombstones:              tombstones,
			InputPackages:                packages,
			InputAPIs:                    apis,
			InputEvents:                  events,
			InputEntityTypes:             entityTypes,
			InputCapabilities:            capabilities,
			InputIntegrationDependencies: integrationDependencies,
			InputResource:                resource.Application,
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail while deleting bundle",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			PackageSvcFn:               successfulPackageDelete,
			APISvcFn:                   successfulAPIDelete,
			EventSvcFn:                 successfulEventDelete,
			EntityTypeSvcFn:            successfulEntityTypeDelete,
			CapabilitySvcFn:            successfulCapabilityDelete,
			IntegrationDependencySvcFn: successfulIntegrationDependencyDelete,
			BundleSvcFn: func() *automock.BundleService {
				bundleSvc := &automock.BundleService{}
				bundleSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "bundle-id").Return(testErr).Once()
				return bundleSvc
			},
			InputTombstones:              tombstones,
			InputPackages:                packages,
			InputAPIs:                    apis,
			InputEvents:                  events,
			InputEntityTypes:             entityTypes,
			InputCapabilities:            capabilities,
			InputIntegrationDependencies: integrationDependencies,
			InputBundles:                 bundles,
			InputResource:                resource.Application,
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail while deleting vendor",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			PackageSvcFn:               successfulPackageDelete,
			APISvcFn:                   successfulAPIDelete,
			EventSvcFn:                 successfulEventDelete,
			EntityTypeSvcFn:            successfulEntityTypeDelete,
			CapabilitySvcFn:            successfulCapabilityDelete,
			IntegrationDependencySvcFn: successfulIntegrationDependencyDelete,
			BundleSvcFn:                successfulBundleDelete,
			VendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "vendor-id").Return(testErr).Once()
				return vendorSvc
			},
			InputTombstones:              tombstones,
			InputPackages:                packages,
			InputAPIs:                    apis,
			InputEvents:                  events,
			InputEntityTypes:             entityTypes,
			InputCapabilities:            capabilities,
			InputIntegrationDependencies: integrationDependencies,
			InputBundles:                 bundles,
			InputVendors:                 vendors,
			InputResource:                resource.Application,
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail while deleting product",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			PackageSvcFn:               successfulPackageDelete,
			APISvcFn:                   successfulAPIDelete,
			EventSvcFn:                 successfulEventDelete,
			EntityTypeSvcFn:            successfulEntityTypeDelete,
			CapabilitySvcFn:            successfulCapabilityDelete,
			IntegrationDependencySvcFn: successfulIntegrationDependencyDelete,
			BundleSvcFn:                successfulBundleDelete,
			VendorSvcFn:                successfulVendorDelete,
			ProductSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, "product-id").Return(testErr).Once()
				return productSvc
			},
			InputTombstones:              tombstones,
			InputPackages:                packages,
			InputAPIs:                    apis,
			InputEvents:                  events,
			InputEntityTypes:             entityTypes,
			InputCapabilities:            capabilities,
			InputIntegrationDependencies: integrationDependencies,
			InputBundles:                 bundles,
			InputVendors:                 vendors,
			InputProducts:                productModels,
			InputResource:                resource.Application,
			ExpectedErr:                  testErr,
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			packageSvc := &automock.PackageService{}
			if test.PackageSvcFn != nil {
				packageSvc = test.PackageSvcFn()
			}

			apiSvc := &automock.APIService{}
			if test.APISvcFn != nil {
				apiSvc = test.APISvcFn()
			}

			eventSvc := &automock.EventService{}
			if test.EventSvcFn != nil {
				eventSvc = test.EventSvcFn()
			}

			entityTypeSvc := &automock.EntityTypeService{}
			if test.EntityTypeSvcFn != nil {
				entityTypeSvc = test.EntityTypeSvcFn()
			}

			capabilitySvc := &automock.CapabilityService{}
			if test.CapabilitySvcFn != nil {
				capabilitySvc = test.CapabilitySvcFn()
			}

			integrationDependencySvc := &automock.IntegrationDependencyService{}
			if test.IntegrationDependencySvcFn != nil {
				integrationDependencySvc = test.IntegrationDependencySvcFn()
			}

			vendorSvc := &automock.VendorService{}
			if test.VendorSvcFn != nil {
				vendorSvc = test.VendorSvcFn()
			}

			productSvc := &automock.ProductService{}
			if test.ProductSvcFn != nil {
				productSvc = test.ProductSvcFn()
			}

			bundleSvc := &automock.BundleService{}
			if test.BundleSvcFn != nil {
				bundleSvc = test.BundleSvcFn()
			}

			tombstonedResourcesDeleter := processor.NewTombstonedResourcesDeleter(tx, packageSvc, apiSvc, eventSvc, entityTypeSvc, capabilitySvc, integrationDependencySvc, vendorSvc, productSvc, bundleSvc)
			result, err := tombstonedResourcesDeleter.Delete(context.TODO(), test.InputResource, test.InputVendors, test.InputProducts, test.InputPackages, test.InputBundles, test.InputAPIs, test.InputEvents, test.InputEntityTypes, test.InputCapabilities, test.InputIntegrationDependencies, test.InputTombstones, test.InputFetchRequests)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.ExpectedOutput, result)
			}

			mock.AssertExpectationsForObjects(t, tx, packageSvc, apiSvc, eventSvc, entityTypeSvc, capabilitySvc, integrationDependencySvc, vendorSvc, productSvc, bundleSvc)
		})
	}
}
