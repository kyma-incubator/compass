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

	associatedPackages map[string][]string
	pkgIDs             map[string][]string
}

func NewService(transact persistence.Transactioner, appSvc application.ApplicationService, webhookSvc webhook.WebhookService, bundleSvc mp_bundle.BundleService, packageSvc mp_package.PackageService) *Service {
	return &Service{
		transact: transact,

		appSvc:             appSvc,
		webhookSvc:         webhookSvc,
		bundleSvc:          bundleSvc,
		packageSvc:         packageSvc,
		associatedPackages: make(map[string][]string, 0),
		pkgIDs:             make(map[string][]string, 0),
	}
}

func (s *Service) SyncODDocuments(ctx context.Context) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	defer func() { // Clean up cache
		s.associatedPackages = make(map[string][]string, 0)
		s.pkgIDs = make(map[string][]string, 0)
	}()

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

	if err := s.resyncPackageBundleAssociations(ctx); err != nil {
		return err
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
		documents := make(open_discovery.Documents, 0, 0)
		for _, wh := range webhooks {
			if wh.Type == model.WebhookTypeOpenDiscovery {
				docs, err := NewClient().FetchOpenDiscoveryDocuments(wh.URL)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("error fetching OD document for webhook with id %s for app with id %s", wh.ID, app.ID))
				}
				documents = append(documents, docs...)
			}
		}
		if err := s.processDocuments(ctx, app.ID, documents); err != nil {
			return errors.Wrap(err, fmt.Sprintf("error processing OD documents for app with id %s", app.ID))
		}
	}
	return nil
}

func (s *Service) processDocuments(ctx context.Context, appID string, documents open_discovery.Documents) error {
	packages, bundlesWithAssociatedPackages, err := documents.ToModelInputs()
	if err != nil {
		return err
	}
	for _, pkgInput := range packages {
		id, err := s.packageSvc.CreateOrUpdate(ctx, appID, pkgInput.OpenDiscoveryID, *pkgInput)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error create/update package with id %s", pkgInput.ID))
		}
		s.pkgIDs[pkgInput.OpenDiscoveryID] = append(s.pkgIDs[pkgInput.OpenDiscoveryID], id)
		if err := s.packageSvc.DeleteAllBundleAssociations(ctx, id); err != nil {
			return err
		}
	}
	for _, bundleInput := range bundlesWithAssociatedPackages {
		id, err := s.bundleSvc.CreateOrUpdate(ctx, appID, bundleInput.In.OpenDiscoveryID, *bundleInput.In)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error create/update bundle with id %s", bundleInput.In.ID))
		}
		s.associatedPackages[id] = bundleInput.AssociatedPackages
	}
	return nil
}

func (s *Service) resyncPackageBundleAssociations(ctx context.Context) error {
	for bundleID, associatedPkgs := range s.associatedPackages {
		for _, pkgODID := range associatedPkgs {
			pkgIDs, ok := s.pkgIDs[pkgODID]
			if !ok {
				continue
			}
			for _, pkgID := range pkgIDs {
				if err := s.packageSvc.AssociateBundle(ctx, pkgID, bundleID); err != nil {
					return errors.Wrap(err, fmt.Sprintf("error associating bundle with id %s with package with id %s", bundleID, pkgID))
				}
			}
		}
	}
	return nil
}
