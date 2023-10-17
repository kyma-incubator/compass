package processor

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// TombstoneService is responsible for the service-layer Tombstone operations.
//
//go:generate mockery --name=TombstoneService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TombstoneService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.TombstoneInput) (string, error)
	Update(ctx context.Context, resourceType resource.Type, id string, in model.TombstoneInput) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Tombstone, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appID string) ([]*model.Tombstone, error)
}

// TombstoneProcessor defines tombstone processor
type TombstoneProcessor struct {
	transact     persistence.Transactioner
	tombstoneSvc TombstoneService
}

// NewTombstoneProcessor creates new instance of TombstoneProcessor
func NewTombstoneProcessor(transact persistence.Transactioner, tombstoneSvc TombstoneService) *TombstoneProcessor {
	return &TombstoneProcessor{
		transact:     transact,
		tombstoneSvc: tombstoneSvc,
	}
}

// Process re-syncs the tombstones passed as an argument.
func (tp *TombstoneProcessor) Process(ctx context.Context, resourceType resource.Type, resourceID string, tombstones []*model.TombstoneInput) ([]*model.Tombstone, error) {
	tombstonesFromDB, err := tp.listTombstonesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	for _, tombstone := range tombstones {
		if err := tp.resyncTombstoneInTx(ctx, resourceType, resourceID, tombstonesFromDB, tombstone); err != nil {
			return nil, err
		}
	}

	tombstonesFromDB, err = tp.listTombstonesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return tombstonesFromDB, nil
}

func (tp *TombstoneProcessor) listTombstonesInTx(ctx context.Context, resourceType resource.Type, resourceID string) ([]*model.Tombstone, error) {
	tx, err := tp.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer tp.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var tombstonesFromDB []*model.Tombstone
	switch resourceType {
	case resource.Application:
		tombstonesFromDB, err = tp.tombstoneSvc.ListByApplicationID(ctx, resourceID)
	case resource.ApplicationTemplateVersion:
		tombstonesFromDB, err = tp.tombstoneSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing tombstones for %s with id %q", resourceType, resourceID)
	}

	return tombstonesFromDB, tx.Commit()
}

func (tp *TombstoneProcessor) resyncTombstoneInTx(ctx context.Context, resourceType resource.Type, resourceID string, tombstonesFromDB []*model.Tombstone, tombstone *model.TombstoneInput) error {
	tx, err := tp.transact.Begin()
	if err != nil {
		return err
	}
	defer tp.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := tp.resyncTombstone(ctx, resourceType, resourceID, tombstonesFromDB, *tombstone); err != nil {
		return errors.Wrapf(err, "error while resyncing tombstone for resource with ORD ID %q", tombstone.OrdID)
	}
	return tx.Commit()
}

func (tp *TombstoneProcessor) resyncTombstone(ctx context.Context, resourceType resource.Type, resourceID string, tombstonesFromDB []*model.Tombstone, tombstone model.TombstoneInput) error {
	if i, found := searchInSlice(len(tombstonesFromDB), func(i int) bool {
		return tombstonesFromDB[i].OrdID == tombstone.OrdID
	}); found {
		return tp.tombstoneSvc.Update(ctx, resourceType, tombstonesFromDB[i].ID, tombstone)
	}

	_, err := tp.tombstoneSvc.Create(ctx, resourceType, resourceID, tombstone)
	return err
}
