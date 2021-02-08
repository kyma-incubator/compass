package aggregator

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

type service struct {
	transact persistence.Transactioner

	appSvc       ApplicationService
	webhookSvc   WebhookService
	bundleSvc    BundleService
	packageSvc   PackageService
	productSvc   ProductService
	vendorSvc    VendorService
	tombstoneSvc TombstoneService

	ordClient *client
}

func NewService(transact persistence.Transactioner, appSvc ApplicationService, webhookSvc WebhookService, bundleSvc BundleService, packageSvc PackageService, productSvc ProductService, vendorSvc VendorService, tombstoneSvc TombstoneService, client *client) *service {
	return &service{
		transact:     transact,
		appSvc:       appSvc,
		webhookSvc:   webhookSvc,
		bundleSvc:    bundleSvc,
		packageSvc:   packageSvc,
		productSvc:   productSvc,
		vendorSvc:    vendorSvc,
		tombstoneSvc: tombstoneSvc,
		ordClient:    client,
	}
}

func (s *service) SyncORDDocuments(ctx context.Context) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	pageCount := 1
	pageSize := 200
	page, err := s.appSvc.ListGlobal(ctx, pageSize, "")
	if err != nil {
		return errors.Wrapf(err, "error while fetching application page number %d", pageCount)
	}
	pageCount++
	if err := s.processAppPage(ctx, page.Data); err != nil {
		return errors.Wrapf(err, "error while processing application page number %d", pageCount)
	}

	for page.PageInfo.HasNextPage {
		page, err = s.appSvc.ListGlobal(ctx, pageSize, page.PageInfo.EndCursor)
		if err != nil {
			return errors.Wrapf(err, "error while fetching page number %d", pageCount)
		}
		if err := s.processAppPage(ctx, page.Data); err != nil {
			return errors.Wrapf(err, "error while processing page number %d", pageCount)
		}
		pageCount++
	}

	return tx.Commit()
}

func (s *service) processAppPage(ctx context.Context, page []*model.Application) error {
	for _, app := range page {
		ctx = tenant.SaveToContext(ctx, app.Tenant, "")
		webhooks, err := s.webhookSvc.List(ctx, app.ID)
		if err != nil {
			return errors.Wrapf(err, "error fetching webhooks for app with id %q", app.ID)
		}
		documents := make([]*open_resource_discovery.Document, 0, 0)
		for _, wh := range webhooks {
			if wh.Type == model.WebhookTypeOpenResourceDiscovery {
				docs, err := s.ordClient.FetchOpenDiscoveryDocuments(ctx, wh.URL)
				if err != nil {
					return errors.Wrapf(err, "error fetching ORD document for webhook with id %q for app with id %q", wh.ID, app.ID)
				}
				documents = append(documents, docs...)
			}
		}
		if err := s.processDocuments(ctx, app.ID, documents); err != nil {
			return errors.Wrapf(err, "error processing ORD documents for app with id %q", app.ID)
		}
	}
	return nil
}

func (s *service) processDocuments(ctx context.Context, appID string, documents []*open_resource_discovery.Document) error {
	return nil
}
