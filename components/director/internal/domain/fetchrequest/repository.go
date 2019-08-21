package fetchrequest

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type repository struct {}

func NewRepository() *repository {
	return &repository{}
}

func (r *repository) Create(ctx context.Context, tenant string, item *model.FetchRequest) error {
	panic("not implemeneted")
}

func (r *repository) GetByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) (*model.FetchRequest, error) {
	panic("not implemeneted")
}

func (r *repository) Delete(ctx context.Context, tenant, id string) error {
	panic("not implemeneted")
}
