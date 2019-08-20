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

var webhookColumns = []string{"id", "tenant_id", "app_id", "type", "url", "auth"}
var missingInputModelError = errors.New("model has to be provided")

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	FromEntity(in Entity) (model.Webhook, error)
	ToEntity(in model.Webhook) (Entity, error)
}

type repository struct {
	singleGetter *repo.SingleGetter
	updater      *repo.Updater
	creator      *repo.Creator
	deleter      *repo.Deleter
	lister       *repo.Lister
	conv         EntityConverter
}

func NewRepository(conv EntityConverter) *repository {
	return &repository{
		singleGetter: repo.NewSingleGetter(tableName, "tenant_id", webhookColumns),
		creator:      repo.NewCreator(tableName, webhookColumns),
		updater:      repo.NewUpdater(tableName, []string{"type", "url", "auth"}, "tenant_id", []string{"id", "app_id"}),
		deleter:      repo.NewDeleter(tableName, "tenant_id"),
		lister:       repo.NewLister(tableName, "tenant_id", webhookColumns),
		conv:         conv,
	}
}

func (r *repository) GetByID(ctx context.Context, tenant, id string) (*model.Webhook, error) {
	var entity Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{{Field: "id", Val: id}}, &entity); err != nil {
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
	if err := r.lister.List(ctx, tenant, &entities, fmt.Sprintf("app_id = %s ", pq.QuoteLiteral(applicationID))); err != nil {
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

	return r.creator.Create(ctx, entity)
}

func (r *repository) CreateMany(ctx context.Context, items []*model.Webhook) error {
	for _, item := range items {
		if err := r.Create(ctx, item); err != nil {
			return errors.Wrapf(err, "while creating %s", item.PrettyString())
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
	return r.updater.UpdateSingle(ctx, entity)
}

func (r *repository) Delete(ctx context.Context, tenant, id string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{{Field: "id", Val: id}})
}

func (r *repository) DeleteAllByApplicationID(ctx context.Context, tenant, applicationID string) error {
	return r.deleter.DeleteMany(ctx, tenant, repo.Conditions{{Field: "app_id", Val: applicationID}})
}
