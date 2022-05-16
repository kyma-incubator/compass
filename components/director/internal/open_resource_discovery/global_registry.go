package ord

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// GlobalRegistryService processes global resources (products and vendors) provided via global registry.
//go:generate mockery --name=GlobalRegistryService --output=automock --outpkg=automock --case=underscore --disable-version-string
type GlobalRegistryService interface {
	SyncGlobalResources(ctx context.Context) (map[string]bool, error)
	ListGlobalResources(ctx context.Context) (map[string]bool, error)
}

// GlobalRegistryConfig contains configuration for GlobalRegistryService.
type GlobalRegistryConfig struct {
	URL string `envconfig:"APP_GLOBAL_REGISTRY_URL"`
}

type globalRegistryService struct {
	config GlobalRegistryConfig

	transact persistence.Transactioner

	vendorService  GlobalVendorService
	productService GlobalProductService

	ordClient Client
}

// NewGlobalRegistryService creates new instance of GlobalRegistryService.
func NewGlobalRegistryService(transact persistence.Transactioner, config GlobalRegistryConfig, vendorService GlobalVendorService, productService GlobalProductService, ordClient Client) *globalRegistryService {
	return &globalRegistryService{
		transact:       transact,
		config:         config,
		vendorService:  vendorService,
		productService: productService,
		ordClient:      ordClient,
	}
}

// SyncGlobalResources syncs global resources (products and vendors) provided via global registry.
func (s *globalRegistryService) SyncGlobalResources(ctx context.Context) (map[string]bool, error) {
	// dummy app used only for logging
	app := &model.Application{
		Name: "global-registry",
		Type: "global-registry",
		BaseEntity: &model.BaseEntity{
			ID: "global-registry",
		},
	}
	documents, _, err := s.ordClient.FetchOpenResourceDiscoveryDocuments(ctx, app, &model.Webhook{
		Type: model.WebhookTypeOpenResourceDiscovery,
		URL:  &s.config.URL,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching global registry documents from %s", s.config.URL)
	}

	if err := documents.Validate(s.config.URL, nil, nil, nil, nil, map[string]bool{}); err != nil {
		return nil, errors.Wrap(err, "while validating global registry documents")
	}

	vendorsInput := make([]*model.VendorInput, 0)
	productsInput := make([]*model.ProductInput, 0)
	packagesInput := make([]*model.PackageInput, 0)
	bundlesInput := make([]*model.BundleCreateInput, 0)
	apisInput := make([]*model.APIDefinitionInput, 0)
	eventsInput := make([]*model.EventDefinitionInput, 0)
	tombstonesInput := make([]*model.TombstoneInput, 0)
	for _, doc := range documents {
		vendorsInput = append(vendorsInput, doc.Vendors...)
		productsInput = append(productsInput, doc.Products...)
		packagesInput = append(packagesInput, doc.Packages...)
		bundlesInput = append(bundlesInput, doc.ConsumptionBundles...)
		apisInput = append(apisInput, doc.APIResources...)
		eventsInput = append(eventsInput, doc.EventResources...)
		tombstonesInput = append(tombstonesInput, doc.Tombstones...)
	}

	if len(packagesInput) > 0 || len(bundlesInput) > 0 || len(apisInput) > 0 || len(eventsInput) > 0 || len(tombstonesInput) > 0 {
		return nil, errors.New("global registry supports only vendors and products")
	}

	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	vendorsFromDB, err := s.processVendors(ctx, vendorsInput)
	if err != nil {
		return nil, err
	}

	productsFromDB, err := s.processProducts(ctx, productsInput)
	if err != nil {
		return nil, err
	}

	globalResourceOrdIDs := make(map[string]bool, len(vendorsFromDB)+len(productsFromDB))
	for _, vendor := range vendorsFromDB {
		globalResourceOrdIDs[vendor.OrdID] = true
	}
	for _, product := range productsFromDB {
		globalResourceOrdIDs[product.OrdID] = true
	}

	return globalResourceOrdIDs, tx.Commit()
}

func (s *globalRegistryService) ListGlobalResources(ctx context.Context) (map[string]bool, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	vendorsFromDB, err := s.vendorService.ListGlobal(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error while listing global vendors")
	}
	productsFromDB, err := s.productService.ListGlobal(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error while listing global products")
	}

	globalResourceOrdIDs := make(map[string]bool, len(vendorsFromDB)+len(productsFromDB))
	for _, vendor := range vendorsFromDB {
		globalResourceOrdIDs[vendor.OrdID] = true
	}
	for _, product := range productsFromDB {
		globalResourceOrdIDs[product.OrdID] = true
	}

	return globalResourceOrdIDs, tx.Commit()
}

func (s *globalRegistryService) processVendors(ctx context.Context, vendors []*model.VendorInput) ([]*model.Vendor, error) {
	vendorsFromDB, err := s.vendorService.ListGlobal(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error while listing global vendors")
	}

	for _, vendor := range vendors {
		if err := s.resyncVendor(ctx, vendorsFromDB, *vendor); err != nil {
			return nil, errors.Wrapf(err, "error while resyncing vendor with ORD ID %q", vendor.OrdID)
		}
	}

	for _, vendor := range vendorsFromDB {
		if _, found := searchInSlice(len(vendors), func(i int) bool {
			return vendors[i].OrdID == vendor.OrdID
		}); !found {
			if err := s.vendorService.DeleteGlobal(ctx, vendor.ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting vendor with ID %q", vendor.ID)
			}
		}
	}

	return s.vendorService.ListGlobal(ctx)
}

func (s *globalRegistryService) processProducts(ctx context.Context, products []*model.ProductInput) ([]*model.Product, error) {
	productsFromDB, err := s.productService.ListGlobal(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error while listing global products")
	}

	for _, product := range products {
		if err := s.resyncProduct(ctx, productsFromDB, *product); err != nil {
			return nil, errors.Wrapf(err, "error while resyncing product with ORD ID %q", product.OrdID)
		}
	}

	for _, product := range productsFromDB {
		if _, found := searchInSlice(len(products), func(i int) bool {
			return products[i].OrdID == product.OrdID
		}); !found {
			if err := s.productService.DeleteGlobal(ctx, product.ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting product with ID %q", product.ID)
			}
		}
	}

	return s.productService.ListGlobal(ctx)
}

func (s *globalRegistryService) resyncVendor(ctx context.Context, vendorsFromDB []*model.Vendor, vendor model.VendorInput) error {
	ctx = addFieldToLogger(ctx, "vendor_ord_id", vendor.OrdID)
	if i, found := searchInSlice(len(vendorsFromDB), func(i int) bool {
		return vendorsFromDB[i].OrdID == vendor.OrdID
	}); found {
		return s.vendorService.UpdateGlobal(ctx, vendorsFromDB[i].ID, vendor)
	}
	_, err := s.vendorService.CreateGlobal(ctx, vendor)
	return err
}

func (s *globalRegistryService) resyncProduct(ctx context.Context, productsFromDB []*model.Product, product model.ProductInput) error {
	ctx = addFieldToLogger(ctx, "product_ord_id", product.OrdID)
	if i, found := searchInSlice(len(productsFromDB), func(i int) bool {
		return productsFromDB[i].OrdID == product.OrdID
	}); found {
		return s.productService.UpdateGlobal(ctx, productsFromDB[i].ID, product)
	}
	_, err := s.productService.CreateGlobal(ctx, product)
	return err
}
