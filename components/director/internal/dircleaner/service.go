package dircleaner

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// CleanerService missing godoc
type CleanerService interface {
	Clean(ctx context.Context) error
}

// TenantService missing godoc
type TenantService interface {
	ListSubaccounts(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	Update(ctx context.Context, id string, tenantInput model.BusinessTenantMappingInput) error
}

// CisService missing godoc
type CisService interface {
	GetGlobalAccount(ctx context.Context, subaccountID string) (string, error)
}

type service struct {
	transact  persistence.Transactioner
	tenantSvc TenantService
	cisSvc    CisService
}

// NewCleaner missing godoc
func NewCleaner(transact persistence.Transactioner, tenantSvc TenantService, cisSvc CisService) CleanerService {
	return &service{
		transact:  transact,
		tenantSvc: tenantSvc,
		cisSvc:    cisSvc,
	}
}

// Clean missing godoc
func (s *service) Clean(ctx context.Context) error {
	log.C(ctx).Info("Listing all subaccounts provided by the event-service")
	allSubaccounts, err := s.tenantSvc.ListSubaccounts(ctx)
	if err != nil {
		return errors.Wrap(err, "while getting all subaccounts")
	}
	log.C(ctx).Info("Done listing subaccounts")
	for _, subaccount := range allSubaccounts {
		globalAccountGUIDFromCis, err := s.cisSvc.GetGlobalAccount(ctx, subaccount.ID)
		if err != nil {
			return errors.Wrapf(err, "while getting global account guid for subaacount with ID %s", subaccount.ID)
		}
		err = func() error {
			tx, err := s.transact.Begin()
			if err != nil {
				return errors.Wrap(err, "failed to begin transaction")
			}
			ctx = persistence.SaveToContext(ctx, tx)
			defer s.transact.RollbackUnlessCommitted(ctx, tx)

			parentFromDB, err := s.tenantSvc.GetTenantByID(ctx, subaccount.Parent)
			if err != nil {
				return err
			}
			if parentFromDB.ExternalTenant != globalAccountGUIDFromCis { // the record is directory and not GA
				log.C(ctx).Infof("Updating record with id %s", parentFromDB.ID)
				update := model.BusinessTenantMappingInput{
					Name:           globalAccountGUIDFromCis, // set new name
					ExternalTenant: globalAccountGUIDFromCis, // set new external tenant
					Parent:         parentFromDB.Parent,
					Type:           tenant.TypeToStr(parentFromDB.Type),
					Provider:       parentFromDB.Provider,
				}
				if err := s.tenantSvc.Update(ctx, parentFromDB.ID, update); err != nil {
					return err
				}
				log.C(ctx).Info("Record updated")
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}
