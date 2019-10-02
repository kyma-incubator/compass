package webhook

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const (
	tableName = "public.webhooks"
)

var (
	webhookColumns         = []string{"id", "tenant_id", "app_id", "type", "url", "auth"}
	missingInputModelError = errors.New("model has to be provided")
	tenantColumn           = "tenant_id"
)

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	FromEntity(in Entity) (model.Webhook, error)
	ToEntity(in model.Webhook) (Entity, error)
}

type repository struct {
	repo.SingleGetter
	repo.Updater
	repo.Creator
	repo.Deleter
	repo.Lister
	conv EntityConverter
}

func NewRepository(conv EntityConverter) *repository {
	return &repository{
		SingleGetter: repo.NewSingleGetter(tableName, tenantColumn, webhookColumns),
		Creator:      repo.NewCreator(tableName, webhookColumns),
		Updater:      repo.NewUpdater(tableName, []string{"type", "url", "auth"}, tenantColumn, []string{"id", "app_id"}),
		Deleter:      repo.NewDeleter(tableName, tenantColumn),
		Lister:       repo.NewLister(tableName, tenantColumn, webhookColumns),
		conv:         conv,
	}
}

func (r *repository) GetByID(ctx context.Context, tenant, id string) (*model.Webhook, error) {
	var entity Entity
	if err := r.SingleGetter.Get(ctx, tenant, repo.Conditions{{Field: "id", Val: id}}, &entity); err != nil {
		return nil, err
	}
	m, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting from entity to model")
	}
	return &m, nil
}

func (r *repository) ListByApplicationID(ctx context.Context, tenant, applicationID string) ([]*model.Webhook, error) {
	var entities Collection
	if err := r.Lister.List(ctx, tenant, &entities, fmt.Sprintf("app_id = %s ", pq.QuoteLiteral(applicationID))); err != nil {
		return nil, err
	}

	var out []*model.Webhook
	for _, ent := range entities {
		w, err := r.conv.FromEntity(ent)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Webhook to model")
		}
		out = append(out, &w)
	}

	return out, nil
}

func (r *repository) Create(ctx context.Context, item *model.Webhook) error {
	if item == nil {
		return missingInputModelError
	}
	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}

	return r.Creator.Create(ctx, entity)
}

func (r *repository) CreateMany(ctx context.Context, items []*model.Webhook) error {
	for _, item := range items {
		if err := r.Create(ctx, item); err != nil {
			return errors.Wrap(err, "while creating many webhooks")
		}
	}
	return nil
}

func (r *repository) Update(ctx context.Context, item *model.Webhook) error {
	if item == nil {
		return missingInputModelError
	}
	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}
	return r.Updater.UpdateSingle(ctx, entity)
}

func (r *repository) Delete(ctx context.Context, tenant, id string) error {
	return r.Deleter.DeleteOne(ctx, tenant, repo.Conditions{{Field: "id", Val: id}})
}

func (r *repository) DeleteAllByApplicationID(ctx context.Context, tenant, applicationID string) error {
	return r.Deleter.DeleteMany(ctx, tenant, repo.Conditions{{Field: "app_id", Val: applicationID}})
}
