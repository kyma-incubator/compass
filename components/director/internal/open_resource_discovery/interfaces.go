package ord

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// WebhookService is responsible for the service-layer Webhook operations.
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookService interface {
	ListByWebhookType(ctx context.Context, webhookType model.WebhookType) ([]*model.Webhook, error)
}

// ApplicationService is responsible for the service-layer Application operations.
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	Get(ctx context.Context, id string) (*model.Application, error)
	ListAllByApplicationTemplateID(ctx context.Context, applicationTemplateID string) ([]*model.Application, error)
}

// BundleService is responsible for the service-layer Bundle operations.
//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleService interface {
	Create(ctx context.Context, applicationID string, in model.BundleCreateInput) (string, error)
	Update(ctx context.Context, id string, in model.BundleUpdateInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationIDNoPaging(ctx context.Context, appID string) ([]*model.Bundle, error)
}

// BundleReferenceService is responsible for the service-layer BundleReference operations.
//go:generate mockery --name=BundleReferenceService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleReferenceService interface {
	GetBundleIDsForObject(ctx context.Context, objectType model.BundleReferenceObjectType, objectID *string) ([]string, error)
}

// APIService is responsible for the service-layer API operations.
//go:generate mockery --name=APIService --output=automock --outpkg=automock --case=underscore --disable-version-string
type APIService interface {
	Create(ctx context.Context, appID string, bundleID, packageID *string, in model.APIDefinitionInput, spec []*model.SpecInput, targetURLsPerBundle map[string]string, apiHash uint64, defaultBundleID string) (string, error)
	UpdateInManyBundles(ctx context.Context, id string, in model.APIDefinitionInput, specIn *model.SpecInput, defaultTargetURLPerBundle map[string]string, defaultTargetURLPerBundleToBeCreated map[string]string, bundleIDsToBeDeleted []string, apiHash uint64, defaultBundleID string) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.APIDefinition, error)
}

// EventService is responsible for the service-layer Event operations.
//go:generate mockery --name=EventService --output=automock --outpkg=automock --case=underscore --disable-version-string
type EventService interface {
	Create(ctx context.Context, appID string, bundleID, packageID *string, in model.EventDefinitionInput, specs []*model.SpecInput, bundleIDs []string, eventHash uint64, defaultBundleID string) (string, error)
	UpdateInManyBundles(ctx context.Context, id string, in model.EventDefinitionInput, specIn *model.SpecInput, bundleIDsFromBundleReference, bundleIDsForCreation, bundleIDsForDeletion []string, eventHash uint64, defaultBundleID string) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.EventDefinition, error)
}

// SpecService is responsible for the service-layer Specification operations.
//go:generate mockery --name=SpecService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	CreateByReferenceObjectIDWithDelayedFetchRequest(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, *model.FetchRequest, error)
	DeleteByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) error
	GetByID(ctx context.Context, id string, objectType model.SpecReferenceObjectType) (*model.Spec, error)
	GetFetchRequest(ctx context.Context, specID string, objectType model.SpecReferenceObjectType) (*model.FetchRequest, error)
	UpdateSpecOnly(ctx context.Context, spec model.Spec) error
	ListIDByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) ([]string, error)
	RefetchSpec(ctx context.Context, id string, objectType model.SpecReferenceObjectType) (*model.Spec, error)
}

// FetchRequestService is responsible for executing specification fetch requests.
//go:generate mockery --name=FetchRequestService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FetchRequestService interface {
	FetchSpec(ctx context.Context, fr *model.FetchRequest) (*string, *model.FetchRequestStatus)
	Update(ctx context.Context, fr *model.FetchRequest) error
}

// PackageService is responsible for the service-layer Package operations.
//go:generate mockery --name=PackageService --output=automock --outpkg=automock --case=underscore --disable-version-string
type PackageService interface {
	Create(ctx context.Context, applicationID string, in model.PackageInput, pkgHash uint64) (string, error)
	Update(ctx context.Context, id string, in model.PackageInput, pkgHash uint64) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Package, error)
}

// ProductService is responsible for the service-layer Product operations.
//go:generate mockery --name=ProductService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ProductService interface {
	Create(ctx context.Context, applicationID string, in model.ProductInput) (string, error)
	Update(ctx context.Context, id string, in model.ProductInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Product, error)
}

// GlobalProductService is responsible for the service-layer operations for Global Product (with NULL app_id) without tenant isolation.
//go:generate mockery --name=GlobalProductService --output=automock --outpkg=automock --case=underscore --disable-version-string
type GlobalProductService interface {
	CreateGlobal(ctx context.Context, in model.ProductInput) (string, error)
	UpdateGlobal(ctx context.Context, id string, in model.ProductInput) error
	DeleteGlobal(ctx context.Context, id string) error
	ListGlobal(ctx context.Context) ([]*model.Product, error)
}

// VendorService is responsible for the service-layer Vendor operations.
//go:generate mockery --name=VendorService --output=automock --outpkg=automock --case=underscore --disable-version-string
type VendorService interface {
	Create(ctx context.Context, applicationID string, in model.VendorInput) (string, error)
	Update(ctx context.Context, id string, in model.VendorInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Vendor, error)
}

// GlobalVendorService is responsible for the service-layer operations for Global Vendors (with NULL app_id) without tenant isolation.
//go:generate mockery --name=GlobalVendorService --output=automock --outpkg=automock --case=underscore --disable-version-string
type GlobalVendorService interface {
	CreateGlobal(ctx context.Context, in model.VendorInput) (string, error)
	UpdateGlobal(ctx context.Context, id string, in model.VendorInput) error
	DeleteGlobal(ctx context.Context, id string) error
	ListGlobal(ctx context.Context) ([]*model.Vendor, error)
}

// TombstoneService is responsible for the service-layer Tombstone operations.
//go:generate mockery --name=TombstoneService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TombstoneService interface {
	Create(ctx context.Context, applicationID string, in model.TombstoneInput) (string, error)
	Update(ctx context.Context, id string, in model.TombstoneInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Tombstone, error)
}

// TenantService missing godoc
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantService interface {
	GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error)
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}
