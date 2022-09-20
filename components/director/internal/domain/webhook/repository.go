package webhook

import (
	"context"
	"time"

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
	webhookColumns         = []string{"id", "app_id", "app_template_id", "type", "url", "auth", "runtime_id", "integration_system_id", "mode", "correlation_id_key", "retry_interval", "timeout", "url_template", "input_template", "header_template", "output_template", "status_template", "created_at"}
	updatableColumns       = []string{"type", "url", "auth", "mode", "retry_interval", "timeout", "url_template", "input_template", "header_template", "output_template", "status_template"}
	missingInputModelError = apperrors.NewInternalError("model has to be provided")
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	FromEntity(in *Entity) (*model.Webhook, error)
	ToEntity(in *model.Webhook) (*Entity, error)
}

type repository struct {
	singleGetter                   repo.SingleGetter
	singleGetterGlobal             repo.SingleGetterGlobal
	webhookUpdater                 repo.Updater
	updaterGlobal                  repo.UpdaterGlobal
	creator                        repo.Creator
	globalCreator                  repo.CreatorGlobal
	deleterGlobal                  repo.DeleterGlobal
	deleter                        repo.Deleter
	lister                         repo.Lister
	listerGlobal                   repo.ListerGlobal
	listerGlobalOrderedByCreatedAt repo.ListerGlobal
	conv                           EntityConverter
}

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		singleGetter:       repo.NewSingleGetter(tableName, webhookColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.Webhook, tableName, webhookColumns),
		creator:            repo.NewCreator(tableName, webhookColumns),
		globalCreator:      repo.NewCreatorGlobal(resource.Webhook, tableName, webhookColumns),
		webhookUpdater:     repo.NewUpdater(tableName, updatableColumns, []string{"id"}),
		updaterGlobal:      repo.NewUpdaterGlobal(resource.Webhook, tableName, updatableColumns, []string{"id", "app_template_id"}),
		deleterGlobal:      repo.NewDeleterGlobal(resource.Webhook, tableName),
		deleter:            repo.NewDeleter(tableName),
		lister:             repo.NewLister(tableName, webhookColumns),
		listerGlobal:       repo.NewListerGlobal(resource.Webhook, tableName, webhookColumns),
		listerGlobalOrderedByCreatedAt: repo.NewListerGlobalWithOrderBy(resource.Webhook, tableName, webhookColumns, repo.OrderByParams{
			{
				Field: "created_at",
				Dir:   repo.DescOrderBy,
			},
		}),
		conv: conv,
	}
}

// GetByID missing godoc
func (r *repository) GetByID(ctx context.Context, tenant, id string, objectType model.WebhookReferenceObjectType) (*model.Webhook, error) {
	var entity Entity
	if err := r.singleGetter.Get(ctx, objectType.GetResourceType(), tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}
	m, err := r.conv.FromEntity(&entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting from entity to model")
	}
	return m, nil
}

// GetByIDGlobal missing godoc
func (r *repository) GetByIDGlobal(ctx context.Context, id string) (*model.Webhook, error) {
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}
	m, err := r.conv.FromEntity(&entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting from entity to model")
	}
	return m, nil
}

// ListByReferenceObjectID missing godoc
func (r *repository) ListByReferenceObjectID(ctx context.Context, tenant, objID string, objType model.WebhookReferenceObjectType) ([]*model.Webhook, error) {
	var entities Collection

	refColumn, err := getReferenceColumnForListByReferenceObjectType(objType)
	if err != nil {
		return nil, err
	}

	conditions := repo.Conditions{
		repo.NewEqualCondition(refColumn, objID),
	}

	if err := r.lister.List(ctx, objType.GetResourceType(), tenant, &entities, conditions...); err != nil {
		return nil, err
	}

	return convertToWebhooks(entities, r)
}

// ListByReferenceObjectTypeAndWebhookType lists all webhooks of a given type for a given object type
func (r *repository) ListByReferenceObjectTypeAndWebhookType(ctx context.Context, tenant string, whType model.WebhookType, objType model.WebhookReferenceObjectType) ([]*model.Webhook, error) {
	var entities Collection

	refColumn, err := getReferenceColumnForListByReferenceObjectType(objType)
	if err != nil {
		return nil, err
	}

	conditions := repo.Conditions{
		repo.NewNotNullCondition(refColumn),
		repo.NewEqualCondition("type", whType),
	}

	if err := r.lister.List(ctx, objType.GetResourceType(), tenant, &entities, conditions...); err != nil {
		return nil, err
	}

	return convertToWebhooks(entities, r)
}

// GetByIDAndWebhookType returns a webhook given an objectID, objectType and webhookType
func (r *repository) GetByIDAndWebhookType(ctx context.Context, tenant, objectID string, objectType model.WebhookReferenceObjectType, webhookType model.WebhookType) (*model.Webhook, error) {
	var entity Entity
	refColumn, err := getReferenceColumnForListByReferenceObjectType(objectType)
	if err != nil {
		return nil, err
	}

	conditions := repo.Conditions{
		repo.NewEqualCondition(refColumn, objectID),
		repo.NewEqualCondition("type", webhookType),
	}
	if err := r.singleGetter.Get(ctx, objectType.GetResourceType(), tenant, conditions, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}
	m, err := r.conv.FromEntity(&entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting from entity to model")
	}
	return m, nil
}

func (r *repository) ListByApplicationIDWithSelectForUpdate(ctx context.Context, tenant, applicationID string) ([]*model.Webhook, error) {
	var entities Collection

	conditions := repo.Conditions{
		repo.NewEqualCondition("app_id", applicationID),
	}

	if err := r.lister.ListWithSelectForUpdate(ctx, resource.AppWebhook, tenant, &entities, conditions...); err != nil {
		return nil, err
	}

	return convertToWebhooks(entities, r)
}

// ListByWebhookType retrieves all webhooks which have the given webhook type in descending order
func (r *repository) ListByWebhookType(ctx context.Context, webhookType model.WebhookType) ([]*model.Webhook, error) {
	var entities Collection

	conditions := repo.Conditions{
		repo.NewEqualCondition("type", webhookType),
	}

	if err := r.listerGlobalOrderedByCreatedAt.ListGlobal(ctx, &entities, conditions...); err != nil {
		return nil, err
	}

	return convertToWebhooks(entities, r)
}

// ListByApplicationTemplateID missing godoc
func (r *repository) ListByApplicationTemplateID(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error) {
	var entities Collection

	conditions := repo.Conditions{
		repo.NewEqualCondition("app_template_id", applicationTemplateID),
	}

	if err := r.listerGlobal.ListGlobal(ctx, &entities, conditions...); err != nil {
		return nil, err
	}

	return convertToWebhooks(entities, r)
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, tenant string, item *model.Webhook) error {
	if item == nil {
		return missingInputModelError
	}
	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}

	if entity.CreatedAt == nil || entity.CreatedAt.IsZero() {
		now := time.Now()
		entity.CreatedAt = &now
	}

	log.C(ctx).Debugf("Persisting Webhook entity with type %s and id %s for %s to db", item.Type, item.ID, item.ObjectType)
	if len(tenant) == 0 {
		return r.globalCreator.Create(ctx, entity)
	}
	return r.creator.Create(ctx, item.ObjectType.GetResourceType(), tenant, entity)
}

// CreateMany missing godoc
func (r *repository) CreateMany(ctx context.Context, tenant string, items []*model.Webhook) error {
	for _, item := range items {
		if err := r.Create(ctx, tenant, item); err != nil {
			return errors.Wrapf(err, "while creating Webhook with type %s and id %s for %s", item.Type, item.ID, item.ObjectType)
		}
		log.C(ctx).Infof("Successfully created Webhook with type %s and id %s for %s", item.Type, item.ID, item.ObjectType)
	}
	return nil
}

// Update missing godoc
func (r *repository) Update(ctx context.Context, tenant string, item *model.Webhook) error {
	if item == nil {
		return missingInputModelError
	}
	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}
	if item.ObjectType.GetResourceType() == resource.Webhook { // Global resource webhook
		return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
	}
	return r.webhookUpdater.UpdateSingle(ctx, item.ObjectType.GetResourceType(), tenant, entity)
}

// Delete missing godoc
func (r *repository) Delete(ctx context.Context, id string) error {
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteAllByApplicationID missing godoc
func (r *repository) DeleteAllByApplicationID(ctx context.Context, tenant, applicationID string) error {
	return r.deleter.DeleteMany(ctx, resource.AppWebhook, tenant, repo.Conditions{repo.NewEqualCondition("app_id", applicationID)})
}

func convertToWebhooks(entities Collection, r *repository) ([]*model.Webhook, error) {
	out := make([]*model.Webhook, 0, len(entities))
	for _, ent := range entities {
		w, err := r.conv.FromEntity(&ent)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Webhook to model")
		}
		out = append(out, w)
	}

	return out, nil
}

func getReferenceColumnForListByReferenceObjectType(objType model.WebhookReferenceObjectType) (string, error) {
	switch objType {
	case model.ApplicationWebhookReference:
		return "app_id", nil
	case model.RuntimeWebhookReference:
		return "runtime_id", nil
	default:
		return "", errors.New("referenced object should be one of application and runtime")
	}
}
