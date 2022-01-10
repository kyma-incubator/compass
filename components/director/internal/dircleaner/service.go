package dircleaner

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
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
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	DeleteByExternalTenant(ctx context.Context, externalTenantID string) error
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
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Info("Listing all subaccounts provided by the event-service")
	allSubaccounts, err := s.tenantSvc.ListSubaccounts(ctx)
	if err != nil {
		return errors.Wrap(err, "while getting all subaccounts")
	}
	log.C(ctx).Infof("Total number of listed subaccounts: %d", len(allSubaccounts))
	succsessfullyProcessed := 0
	for _, subaccount := range allSubaccounts {
		log.C(ctx).Infof("Processing subaccount with ID %s", subaccount.ID)
		globalAccountGUIDFromCis, err := s.cisSvc.GetGlobalAccount(ctx, subaccount.ExternalTenant)
		if err != nil {
			log.C(ctx).Errorf("Could not get globalAccountGuid for subaacount with ID %s from CIS", subaccount.ID)
			continue
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
				log.C(ctx).Error(err)
			}
			if parentFromDB.ExternalTenant != globalAccountGUIDFromCis { // the record is directory and not GA
				conflictingGA, err := s.tenantSvc.GetTenantByExternalID(ctx, globalAccountGUIDFromCis)

				if conflictingGA != nil && err == nil { // there is a record which conflicts by external tenant id
					updateSubaccount := model.BusinessTenantMappingInput{
						Name:           subaccount.Name,
						ExternalTenant: subaccount.ExternalTenant,
						Parent:         conflictingGA.ID,
						Type:           tenant.TypeToStr(subaccount.Type),
						Provider:       subaccount.Provider,
					}
					log.C(ctx).Infof("Updating subaccount with id %s to point to existing GA with id %s", subaccount.ID, conflictingGA.ID)
					if err = s.tenantSvc.Update(ctx, subaccount.ID, updateSubaccount); err != nil {
						log.C(ctx).Error(err)
					}
					if err == nil { // the update was successful
						// now delete the directory
						log.C(ctx).Infof("Deleting directory with external tenant id %s", parentFromDB.ExternalTenant)
						if err = s.tenantSvc.DeleteByExternalTenant(ctx, parentFromDB.ExternalTenant); err != nil {
							log.C(ctx).Error(err)
						}
					}

					if err == nil {
						succsessfullyProcessed++
					}
				} else if err != nil && !apperrors.IsNotFoundError(err) {
					log.C(ctx).Error(err)
				} else {
					update := model.BusinessTenantMappingInput{
						Name:           globalAccountGUIDFromCis, // set new name
						ExternalTenant: globalAccountGUIDFromCis, // set new external tenant
						Parent:         parentFromDB.Parent,
						Type:           tenant.TypeToStr(parentFromDB.Type),
						Provider:       parentFromDB.Provider,
					}
					log.C(ctx).Infof("Updating directory with id %s with new external id %s", parentFromDB.ID, globalAccountGUIDFromCis)
					if err := s.tenantSvc.Update(ctx, parentFromDB.ID, update); err != nil {
						log.C(ctx).Error(err)
					} else {
						succsessfullyProcessed++
					}
				}
			} else { // Nothing to do with this subaccount
				succsessfullyProcessed++
			}

			return nil
		}()
		if err != nil {
			log.C(ctx).Error(err)
		}
	}

	log.C(ctx).Infof("Successfully processed %d records from %d", succsessfullyProcessed, len(allSubaccounts))
	return nil
}
