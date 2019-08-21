package fetchrequest

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type pgRepository struct {}

func NewRepository() *pgRepository {
	return &pgRepository{}
}

func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.FetchRequest) error {
	panic("not implemeneted")
}

func (r *pgRepository) Update(ctx context.Context, tenant string, id string, in *model.FetchRequest) error {
	panic("not implemeneted")
}

func (r *pgRepository) GetByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) (*model.FetchRequest, error) {
	panic("not implemeneted")
}

func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	panic("not implemeneted")
}
