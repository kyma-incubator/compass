package ord

import (
	"context"
	"sync"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// WebhookService is responsible for the service-layer Webhook operations.
//
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookService interface {
	GetByIDAndWebhookTypeGlobal(ctx context.Context, objectID string, objectType model.WebhookReferenceObjectType, webhookType model.WebhookType) (*model.Webhook, error)
	ListByWebhookType(ctx context.Context, webhookType model.WebhookType) ([]*model.Webhook, error)
	ListForApplication(ctx context.Context, applicationID string) ([]*model.Webhook, error)
	ListForApplicationGlobal(ctx context.Context, applicationID string) ([]*model.Webhook, error)
	ListForApplicationTemplate(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error)
	EnrichWebhooksWithTenantMappingWebhooks(in []*graphql.WebhookInput) ([]*graphql.WebhookInput, error)
	Create(ctx context.Context, owningResourceID string, in model.WebhookInput, objectType model.WebhookReferenceObjectType) (string, error)
	Delete(ctx context.Context, id string, objectType model.WebhookReferenceObjectType) error
}

// ApplicationService is responsible for the service-layer Application operations.
//
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	Get(ctx context.Context, id string) (*model.Application, error)
	GetGlobalByID(ctx context.Context, id string) (*model.Application, error)
	ListAllByApplicationTemplateID(ctx context.Context, applicationTemplateID string) ([]*model.Application, error)
	Update(ctx context.Context, id string, in model.ApplicationUpdateInput) error
}

// BundleService is responsible for the service-layer Bundle operations.
//
//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleService interface {
	CreateBundle(ctx context.Context, resourceType resource.Type, resourceID string, in model.BundleCreateInput, bndlHash uint64) (string, error)
	UpdateBundle(ctx context.Context, resourceType resource.Type, id string, in model.BundleUpdateInput, bndlHash uint64) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationIDNoPaging(ctx context.Context, appID string) ([]*model.Bundle, error)
	ListByApplicationTemplateVersionIDNoPaging(ctx context.Context, appTemplateVersionID string) ([]*model.Bundle, error)
}

// BundleReferenceService is responsible for the service-layer BundleReference operations.
//
//go:generate mockery --name=BundleReferenceService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleReferenceService interface {
	GetBundleIDsForObject(ctx context.Context, objectType model.BundleReferenceObjectType, objectID *string) ([]string, error)
}

// APIService is responsible for the service-layer API operations.
//
//go:generate mockery --name=APIService --output=automock --outpkg=automock --case=underscore --disable-version-string
type APIService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, bundleID, packageID *string, in model.APIDefinitionInput, spec []*model.SpecInput, targetURLsPerBundle map[string]string, apiHash uint64, defaultBundleID string) (string, error)
	UpdateInManyBundles(ctx context.Context, resourceType resource.Type, id string, in model.APIDefinitionInput, specIn *model.SpecInput, defaultTargetURLPerBundle map[string]string, defaultTargetURLPerBundleToBeCreated map[string]string, bundleIDsToBeDeleted []string, apiHash uint64, defaultBundleID string) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.APIDefinition, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.APIDefinition, error)
}

// EventService is responsible for the service-layer Event operations.
//
//go:generate mockery --name=EventService --output=automock --outpkg=automock --case=underscore --disable-version-string
type EventService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, bundleID, packageID *string, in model.EventDefinitionInput, specs []*model.SpecInput, bundleIDs []string, eventHash uint64, defaultBundleID string) (string, error)
	UpdateInManyBundles(ctx context.Context, resourceType resource.Type, id string, in model.EventDefinitionInput, specIn *model.SpecInput, bundleIDsFromBundleReference, bundleIDsForCreation, bundleIDsForDeletion []string, eventHash uint64, defaultBundleID string) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.EventDefinition, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.EventDefinition, error)
}

// CapabilityService is responsible for the service-layer Capability operations.
//
//go:generate mockery --name=CapabilityService --output=automock --outpkg=automock --case=underscore --disable-version-string
type CapabilityService interface {
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Capability, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.Capability, error)
	Create(ctx context.Context, resourceType resource.Type, resourceID string, packageID *string, in model.CapabilityInput, spec []*model.SpecInput, capabilityHash uint64) (string, error)
	Update(ctx context.Context, resourceType resource.Type, id string, in model.CapabilityInput, capabilityHash uint64) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
}

// IntegrationDependencyService is responsible for the service-layer IntegrationDependency operations.
//
//go:generate mockery --name=IntegrationDependencyService --output=automock --outpkg=automock --case=underscore --disable-version-string
type IntegrationDependencyService interface {
	ListByApplicationID(ctx context.Context, appID string) ([]*model.IntegrationDependency, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.IntegrationDependency, error)
	Delete(ctx context.Context, resourceType resource.Type, id string) error
}

// IntegrationDependencyProcessor is responsible for processing of integration dependency entities.
//
//go:generate mockery --name=IntegrationDependencyProcessor --output=automock --outpkg=automock --case=underscore --disable-version-string
type IntegrationDependencyProcessor interface {
	Process(ctx context.Context, resourceType resource.Type, resourceID string, packagesFromDB []*model.Package, integrationDependencies []*model.IntegrationDependencyInput, resourceHashes map[string]uint64) ([]*model.IntegrationDependency, error)
}

// SpecService is responsible for the service-layer Specification operations.
//
//go:generate mockery --name=SpecService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	CreateByReferenceObjectIDWithDelayedFetchRequest(ctx context.Context, in model.SpecInput, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) (string, *model.FetchRequest, error)
	DeleteByReferenceObjectID(ctx context.Context, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) error
	GetByID(ctx context.Context, id string, objectType model.SpecReferenceObjectType) (*model.Spec, error)
	ListFetchRequestsByReferenceObjectIDs(ctx context.Context, tenant string, objectIDs []string, objectType model.SpecReferenceObjectType) ([]*model.FetchRequest, error)
	ListFetchRequestsByReferenceObjectIDsGlobal(ctx context.Context, objectIDs []string, objectType model.SpecReferenceObjectType) ([]*model.FetchRequest, error)
	UpdateSpecOnly(ctx context.Context, spec model.Spec) error
	UpdateSpecOnlyGlobal(ctx context.Context, spec model.Spec) error
	ListIDByReferenceObjectID(ctx context.Context, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) ([]string, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.Spec, error)
}

// FetchRequestService is responsible for executing specification fetch requests.
//
//go:generate mockery --name=FetchRequestService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FetchRequestService interface {
	FetchSpec(ctx context.Context, fr *model.FetchRequest, headers *sync.Map) (*string, *model.FetchRequestStatus)
	Update(ctx context.Context, fr *model.FetchRequest) error
	UpdateGlobal(ctx context.Context, fr *model.FetchRequest) error
}

// PackageService is responsible for the service-layer Package operations.
//
//go:generate mockery --name=PackageService --output=automock --outpkg=automock --case=underscore --disable-version-string
type PackageService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.PackageInput, pkgHash uint64) (string, error)
	Update(ctx context.Context, resourceType resource.Type, id string, in model.PackageInput, pkgHash uint64) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Package, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.Package, error)
}

// ProductService is responsible for the service-layer Product operations.
//
//go:generate mockery --name=ProductService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ProductService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.ProductInput) (string, error)
	Update(ctx context.Context, resourceType resource.Type, id string, in model.ProductInput) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Product, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appID string) ([]*model.Product, error)
}

// GlobalProductService is responsible for the service-layer operations for Global Product (with NULL app_id) without tenant isolation.
//
//go:generate mockery --name=GlobalProductService --output=automock --outpkg=automock --case=underscore --disable-version-string
type GlobalProductService interface {
	CreateGlobal(ctx context.Context, in model.ProductInput) (string, error)
	UpdateGlobal(ctx context.Context, id string, in model.ProductInput) error
	DeleteGlobal(ctx context.Context, id string) error
	ListGlobal(ctx context.Context) ([]*model.Product, error)
}

// VendorService is responsible for the service-layer Vendor operations.
//
//go:generate mockery --name=VendorService --output=automock --outpkg=automock --case=underscore --disable-version-string
type VendorService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.VendorInput) (string, error)
	Update(ctx context.Context, resourceType resource.Type, id string, in model.VendorInput) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Vendor, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.Vendor, error)
}

// GlobalVendorService is responsible for the service-layer operations for Global Vendors (with NULL app_id) without tenant isolation.
//
//go:generate mockery --name=GlobalVendorService --output=automock --outpkg=automock --case=underscore --disable-version-string
type GlobalVendorService interface {
	CreateGlobal(ctx context.Context, in model.VendorInput) (string, error)
	UpdateGlobal(ctx context.Context, id string, in model.VendorInput) error
	DeleteGlobal(ctx context.Context, id string) error
	ListGlobal(ctx context.Context) ([]*model.Vendor, error)
}

// TombstoneProcessor is responsible for processing of tombstone entities.
//
//go:generate mockery --name=TombstoneProcessor --output=automock --outpkg=automock --case=underscore --disable-version-string
type TombstoneProcessor interface {
	Process(ctx context.Context, resourceType resource.Type, resourceID string, tombstones []*model.TombstoneInput) ([]*model.Tombstone, error)
}

// TenantService missing godoc
//
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantService interface {
	GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error)
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// ApplicationTemplateVersionService is responsible for the service-layer Application Template Version operations
//
//go:generate mockery --name=ApplicationTemplateVersionService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateVersionService interface {
	GetByAppTemplateIDAndVersion(ctx context.Context, id, version string) (*model.ApplicationTemplateVersion, error)
	ListByAppTemplateID(ctx context.Context, appTemplateID string) ([]*model.ApplicationTemplateVersion, error)
	Create(ctx context.Context, appTemplateID string, item *model.ApplicationTemplateVersionInput) (string, error)
	Update(ctx context.Context, id, appTemplateID string, in model.ApplicationTemplateVersionInput) error
}

// ApplicationTemplateService is responsible for the service-layer Application Template operations
//
//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateService interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
}

// WebhookConverter is responsible for converting webhook structs
//
//go:generate mockery --name=WebhookConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookConverter interface {
	InputFromGraphQL(in *graphql.WebhookInput) (*model.WebhookInput, error)
}

// LabelService is responsible for the service-layer Label operations
//
//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelService interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}
