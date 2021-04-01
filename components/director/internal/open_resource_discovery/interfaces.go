package open_resource_discovery

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore
type WebhookService interface {
	ListForApplication(ctx context.Context, applicationID string) ([]*model.Webhook, error)
}

//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore
type ApplicationService interface {
	ListGlobal(ctx context.Context, pageSize int, cursor string) (*model.ApplicationPage, error)
}

//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore
type BundleService interface {
	Create(ctx context.Context, applicationID string, in model.BundleCreateInput) (string, error)
	Update(ctx context.Context, id string, in model.BundleUpdateInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationIDNoPaging(ctx context.Context, appID string) ([]*model.Bundle, error)
}

//go:generate mockery --name=APIService --output=automock --outpkg=automock --case=underscore
type APIService interface {
	Create(ctx context.Context, appId string, bundleID, packageID *string, in model.APIDefinitionInput, spec []*model.SpecInput) (string, error)
	Update(ctx context.Context, id string, in model.APIDefinitionInput, specIn *model.SpecInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.APIDefinition, error)
}

//go:generate mockery --name=EventService --output=automock --outpkg=automock --case=underscore
type EventService interface {
	Create(ctx context.Context, appId string, bundleID, packageID *string, in model.EventDefinitionInput, spec []*model.SpecInput) (string, error)
	Update(ctx context.Context, id string, in model.EventDefinitionInput, specIn *model.SpecInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.EventDefinition, error)
}

//go:generate mockery --name=SpecService --output=automock --outpkg=automock --case=underscore
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	DeleteByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) error
}

//go:generate mockery --name=PackageService --output=automock --outpkg=automock --case=underscore
type PackageService interface {
	Create(ctx context.Context, applicationID string, in model.PackageInput) (string, error)
	Update(ctx context.Context, id string, in model.PackageInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Package, error)
}

//go:generate mockery --name=ProductService --output=automock --outpkg=automock --case=underscore
type ProductService interface {
	Create(ctx context.Context, applicationID string, in model.ProductInput) (string, error)
	Update(ctx context.Context, id string, in model.ProductInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Product, error)
}

//go:generate mockery --name=VendorService --output=automock --outpkg=automock --case=underscore
type VendorService interface {
	Create(ctx context.Context, applicationID string, in model.VendorInput) (string, error)
	Update(ctx context.Context, id string, in model.VendorInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Vendor, error)
}

//go:generate mockery --name=TombstoneService --output=automock --outpkg=automock --case=underscore
type TombstoneService interface {
	Create(ctx context.Context, applicationID string, in model.TombstoneInput) (string, error)
	Update(ctx context.Context, id string, in model.TombstoneInput) error
	Delete(ctx context.Context, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Tombstone, error)
}
