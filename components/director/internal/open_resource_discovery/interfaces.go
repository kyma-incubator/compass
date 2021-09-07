package ord

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// WebhookService missing godoc
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore
type WebhookService interface {
	ListForApplication(ctx context.Context, applicationID string) ([]*model.Webhook, error)
}

// ApplicationService missing godoc
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore
type ApplicationService interface {
	ListGlobal(ctx context.Context, pageSize int, cursor string) (*model.ApplicationPage, error)
}

// BundleService missing godoc
//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore
type BundleService interface {
	Create(ctx context.Context, applicationID string, in model.BundleCreateInput) (string, error)
	Update(ctx context.Context, id string, in model.BundleUpdateInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationIDNoPaging(ctx context.Context, appID string) ([]*model.Bundle, error)
}

// BundleReferenceService missing godoc
//go:generate mockery --name=BundleReferenceService --output=automock --outpkg=automock --case=underscore
type BundleReferenceService interface {
	GetBundleIDsForObject(ctx context.Context, objectType model.BundleReferenceObjectType, objectID *string) ([]string, error)
}

// APIService missing godoc
//go:generate mockery --name=APIService --output=automock --outpkg=automock --case=underscore
type APIService interface {
	Create(ctx context.Context, appID string, bundleID, packageID *string, in model.APIDefinitionInput, spec []*model.SpecInput, targetURLsPerBundle map[string]string, apiHash uint64) (string, error)
	UpdateInManyBundles(ctx context.Context, id string, in model.APIDefinitionInput, specIn *model.SpecInput, defaultTargetURLPerBundle map[string]string, defaultTargetURLPerBundleToBeCreated map[string]string, bundleIDsToBeDeleted []string, apiHash uint64) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.APIDefinition, error)
}

// EventService missing godoc
//go:generate mockery --name=EventService --output=automock --outpkg=automock --case=underscore
type EventService interface {
	Create(ctx context.Context, appID string, bundleID, packageID *string, in model.EventDefinitionInput, specs []*model.SpecInput, bundleIDs []string, eventHash uint64) (string, error)
	UpdateInManyBundles(ctx context.Context, id string, in model.EventDefinitionInput, specIn *model.SpecInput, bundleIDsForCreation []string, bundleIDsForDeletion []string, eventHash uint64) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.EventDefinition, error)
}

// SpecService missing godoc
//go:generate mockery --name=SpecService --output=automock --outpkg=automock --case=underscore
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	DeleteByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) error
	GetFetchRequest(ctx context.Context, specID string) (*model.FetchRequest, error)
	ListByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) ([]*model.Spec, error)
	RefetchSpec(ctx context.Context, id string) (*model.Spec, error)
}

// PackageService missing godoc
//go:generate mockery --name=PackageService --output=automock --outpkg=automock --case=underscore
type PackageService interface {
	Create(ctx context.Context, applicationID string, in model.PackageInput, pkgHash uint64) (string, error)
	Update(ctx context.Context, id string, in model.PackageInput, pkgHash uint64) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Package, error)
}

// ProductService missing godoc
//go:generate mockery --name=ProductService --output=automock --outpkg=automock --case=underscore
type ProductService interface {
	Create(ctx context.Context, applicationID string, in model.ProductInput) (string, error)
	Update(ctx context.Context, id string, in model.ProductInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Product, error)
}

// VendorService missing godoc
//go:generate mockery --name=VendorService --output=automock --outpkg=automock --case=underscore
type VendorService interface {
	Create(ctx context.Context, applicationID string, in model.VendorInput) (string, error)
	Update(ctx context.Context, id string, in model.VendorInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Vendor, error)
}

// TombstoneService missing godoc
//go:generate mockery --name=TombstoneService --output=automock --outpkg=automock --case=underscore
type TombstoneService interface {
	Create(ctx context.Context, applicationID string, in model.TombstoneInput) (string, error)
	Update(ctx context.Context, id string, in model.TombstoneInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Tombstone, error)
}
