package puller

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	mp_bundle "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/open_discovery"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

type Service struct {
	transact persistence.Transactioner

	appSvc     application.ApplicationService
	webhookSvc webhook.WebhookService
	bundleSvc  mp_bundle.BundleService
	packageSvc mp_package.PackageService
}

func NewService(transact persistence.Transactioner, appSvc application.ApplicationService, webhookSvc webhook.WebhookService, bundleSvc mp_bundle.BundleService, packageSvc mp_package.PackageService) *Service {
	return &Service{
		transact: transact,

		appSvc:     appSvc,
		webhookSvc: webhookSvc,
		bundleSvc:  bundleSvc,
		packageSvc: packageSvc,
	}
}

func (s *Service) SyncODDocuments(ctx context.Context) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	pageCount := 1
	pageSize := 100
	page, err := s.appSvc.ListGlobal(ctx, pageSize, "")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error fetching page number %d", pageCount))
	}
	pageCount++
	if err := s.processAppPage(ctx, page.Data); err != nil {
		return errors.Wrap(err, fmt.Sprintf("error processing page number %d", pageCount))
	}

	for page.PageInfo.HasNextPage {
		page, err = s.appSvc.ListGlobal(ctx, pageSize, page.PageInfo.EndCursor)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error fetching page number %d", pageCount))
		}
		if err := s.processAppPage(ctx, page.Data); err != nil {
			return errors.Wrap(err, fmt.Sprintf("error processing page number %d", pageCount))
		}
		pageCount++
	}
	return tx.Commit()
}

func (s *Service) processAppPage(ctx context.Context, page []*model.Application) error {
	for _, app := range page {
		ctx = tenant.SaveToContext(ctx, app.Tenant, "")
		webhooks, err := s.webhookSvc.List(ctx, app.ID)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error fetching webhooks for app with id: %s", app.ID))
		}
		for _, wh := range webhooks {
			if wh.Type == model.WebhookTypeOpenDiscovery {
				document, err := NewClient().FetchOpenDiscoveryDocument(wh.URL)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("error fetching OD document for webhook with id %s for app with id %s", wh.ID, app.ID))
				}
				if err := s.processDocument(ctx, app.ID, document); err != nil {
					return errors.Wrap(err, fmt.Sprintf("error processing OD document for webhook with id %s for app with id %s", wh.ID, app.ID))
				}
			}
		}
	}
	return nil
}

func (s *Service) processDocument(ctx context.Context, appID string, document *open_discovery.Document) error {
	packages, bundlesWithAssociatedPackages, err := document.ToModelInputs()
	if err != nil {
		return err
	}
	for _, pkgInput := range packages {
		if err := s.packageSvc.CreateOrUpdate(ctx, appID, pkgInput.ID, *pkgInput); err != nil {
			return errors.Wrap(err, fmt.Sprintf("error create/update package with id %s", pkgInput.ID))
		}
	}
	for _, bundleInput := range bundlesWithAssociatedPackages {
		if err := s.bundleSvc.CreateOrUpdate(ctx, appID, bundleInput.In.ID, *bundleInput.In); err != nil {
			return errors.Wrap(err, fmt.Sprintf("error create/update bundle with id %s", bundleInput.In.ID))
		}
	}
	for _, bundleInput := range bundlesWithAssociatedPackages {
		for _, pkgID := range bundleInput.AssociatedPackages {
			if err := s.packageSvc.AssociateBundle(ctx, pkgID, bundleInput.In.ID); err != nil {
				return errors.Wrap(err, fmt.Sprintf("error associating bundle with id %s with package with id %s", bundleInput.In.ID, pkgID))
			}
		}
	}
	return nil
}
