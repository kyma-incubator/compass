package webhook

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const (
	tableName = "public.webhooks"
)

var (
	webhookColumns         = []string{"id", "tenant_id", "app_id", "type", "url", "auth", "runtime_id", "integration_system_id", "mode", "correlation_id_key", "retry_interval", "timeout", "url_template", "input_template", "header_template", "output_template", "status_template"}
	updatableColumns       = []string{"type", "url", "auth", "mode", "retry_interval", "timeout", "url_template", "input_template", "header_template", "output_template", "status_template"}
	missingInputModelError = apperrors.NewInternalError("model has to be provided")
	tenantColumn           = "tenant_id"
)

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	FromEntity(in Entity) (model.Webhook, error)
	ToEntity(in model.Webhook) (Entity, error)
}

type repository struct {
	singleGetter repo.SingleGetter
	updater      repo.Updater
	creator      repo.Creator
	deleter      repo.Deleter
	lister       repo.Lister
	conv         EntityConverter
}

func NewRepository(conv EntityConverter) *repository {
	return &repository{
		singleGetter: repo.NewSingleGetter(resource.Webhook, tableName, tenantColumn, webhookColumns),
		creator:      repo.NewCreator(resource.Webhook, tableName, webhookColumns),
		updater:      repo.NewUpdater(resource.Webhook, tableName, updatableColumns, tenantColumn, []string{"id", "app_id"}),
		deleter:      repo.NewDeleter(resource.Webhook, tableName, tenantColumn),
		lister:       repo.NewLister(resource.Webhook, tableName, tenantColumn, webhookColumns),
		conv:         conv,
	}
}

func (r *repository) GetByID(ctx context.Context, tenant, id string) (*model.Webhook, error) {
	var entity Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
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

	conditions := repo.Conditions{
		repo.NewEqualCondition("app_id", applicationID),
	}

	if err := r.lister.List(ctx, tenant, &entities, conditions...); err != nil {
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

	log.C(ctx).Debugf("Persisting Webhook entity with type %s and id %s for %s to db", item.Type, item.ID, PrintOwnerInfo(item))
	return r.creator.Create(ctx, entity)
}

func (r *repository) CreateMany(ctx context.Context, items []*model.Webhook) error {
	for _, item := range items {
		if err := r.Create(ctx, item); err != nil {
			return errors.Wrapf(err, "while creating Webhook with type %s and id %s for %s", item.Type, item.ID, PrintOwnerInfo(item))
		}
		log.C(ctx).Infof("Successfully created Webhook with type %s and id %s for %s", item.Type, item.ID, PrintOwnerInfo(item))
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
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *repository) DeleteAllByApplicationID(ctx context.Context, tenant, applicationID string) error {
	return r.deleter.DeleteMany(ctx, tenant, repo.Conditions{repo.NewEqualCondition("app_id", applicationID)})
}

func PrintOwnerInfo(item *model.Webhook) string {
	appID := ""
	runtimeID := ""
	integrationSystemID := ""

	if item.ApplicationID != nil {
		appID = *item.ApplicationID
	}

	if item.RuntimeID != nil {
		runtimeID = *item.RuntimeID
	}

	if item.IntegrationSystemID != nil {
		integrationSystemID = *item.IntegrationSystemID
	}

	return fmt.Sprintf("Application ID: %q, Runtime ID: %q, Integration System ID: %q", appID, runtimeID, integrationSystemID)
}
