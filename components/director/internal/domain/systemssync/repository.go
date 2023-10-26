package systemssync

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const tableName string = `public.systems_sync_timestamps`

var (
	tableColumns       = []string{"id", "tenant_id", "product_id", "last_sync_timestamp"}
	conflictingColumns = []string{"id"}
	updateColumns      = []string{"tenant_id", "product_id", "last_sync_timestamp"}
)

// EntityConverter converts between the service model and entity
//
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.SystemSynchronizationTimestamp) *Entity
	FromEntity(entity *Entity) *model.SystemSynchronizationTimestamp
}

type repository struct {
	upserter     repo.UpserterGlobal
	listerGlobal repo.ListerGlobal
	conv         EntityConverter
}

// NewRepository creates a new SystemsSync repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		upserter:     repo.NewUpserterGlobal(resource.SystemsSync, tableName, tableColumns, conflictingColumns, updateColumns),
		listerGlobal: repo.NewListerGlobal(resource.SystemsSync, tableName, tableColumns),
		conv:         conv,
	}
}

// List returns all system sync timestamps from database
func (r *repository) List(ctx context.Context) ([]*model.SystemSynchronizationTimestamp, error) {
	var entityCollection EntityCollection

	err := r.listerGlobal.ListGlobal(ctx, &entityCollection)

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entityCollection), nil
}

// Upsert updates sync timestamp or creates new one if it doesn't exist
func (r *repository) Upsert(ctx context.Context, in *model.SystemSynchronizationTimestamp) error {
	sync := r.conv.ToEntity(in)

	return r.upserter.UpsertGlobal(ctx, sync)
}

func (r *repository) multipleFromEntities(entities EntityCollection) []*model.SystemSynchronizationTimestamp {
	items := make([]*model.SystemSynchronizationTimestamp, 0, len(entities))

	for _, entity := range entities {
		sModel := r.conv.FromEntity(entity)
		items = append(items, sModel)
	}

	return items
}
