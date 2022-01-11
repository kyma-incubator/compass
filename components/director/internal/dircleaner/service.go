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
	allSubaccounts, err := s.getSubaccounts(ctx)
	if err != nil {
		return err
	}
	log.C(ctx).Infof("Total number of listed subaccounts: %d", len(allSubaccounts))
	succsessfullyProcessed := 0
	dirsToDelete := make(map[string]bool)

	for _, subaccount := range allSubaccounts {
		log.C(ctx).Infof("Processing subaccount with ID %s", subaccount.ID)
		globalAccountGUIDFromCis, err := s.cisSvc.GetGlobalAccount(ctx, subaccount.ExternalTenant)
		if err != nil {
			log.C(ctx).Errorf("Could not get globalAccountGuid for subaacount with ID %s from CIS %v", subaccount.ID, err)
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
				log.C(ctx).Errorf("Could not take parent for subaccout with id %s %v", subaccount.ID, err)
				return err
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
					} else { // the update was successful
						// mark the directory for deletion later because it still may have other child subaccounts which will be deleted by the cascade
						dirsToDelete[parentFromDB.ExternalTenant] = true
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

			if err = tx.Commit(); err != nil {
				log.C(ctx).Error(err)
			}
			return nil
		}()
		if err != nil {
			log.C(ctx).Error(err)
		}
	}

	if err = s.deleteDirectories(ctx, dirsToDelete); err != nil {
		log.C(ctx).Error(err)
	}
	log.C(ctx).Infof("Successfully processed %d records from %d", succsessfullyProcessed, len(allSubaccounts))
	return nil
}

func (s *service) deleteDirectories(ctx context.Context, dirs map[string]bool) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("%d directories are to be deleted", len(dirs))
	successfullyDeleted := 0
	for extTenant := range dirs {
		if err = s.tenantSvc.DeleteByExternalTenant(ctx, extTenant); err != nil {
			log.C(ctx).Error(err)
		} else {
			successfullyDeleted++
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	log.C(ctx).Infof("%d from %d directories were deleted", successfullyDeleted, len(dirs))

	return nil
}

func (s *service) getSubaccounts(ctx context.Context) ([]*model.BusinessTenantMapping, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Info("Listing all subaccounts provided by the event-service")
	allSubaccounts, err := s.tenantSvc.ListSubaccounts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while getting all subaccounts")
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return allSubaccounts, nil
}
