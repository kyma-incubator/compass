package formationtemplate

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	tableName      string = `public.formation_templates`
	tenantIDColumn string = "tenant_id"
	idColumn       string = "id"
	nameColumn     string = "name"
)

var (
	idTableColumns            = []string{idColumn}
	updatableTableColumns     = []string{"name", "application_types", "runtime_types", "runtime_type_display_name", "runtime_artifact_kind", "leading_product_ids", "supports_reset", "discovery_consumers", "updated_at"}
	tenantTableColumn         = []string{tenantIDColumn}
	tableColumnsWithoutTenant = []string{idColumn, "name", "application_types", "runtime_types", "runtime_type_display_name", "runtime_artifact_kind", "leading_product_ids", "supports_reset", "discovery_consumers", "created_at", "updated_at"}
	tableColumns              = append(tableColumnsWithoutTenant, tenantTableColumn...)

	// Now is a function variable that returns the current time. It is used, so we could mock it in the tests.
	Now = time.Now
)

// EntityConverter converts between the internal model and entity
//
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.FormationTemplate) (*Entity, error)
	FromEntity(entity *Entity) (*model.FormationTemplate, error)
}

type repository struct {
	creator                        repo.CreatorGlobal
	existQuerierGlobal             repo.ExistQuerierGlobal
	singleGetterGlobal             repo.SingleGetterGlobal
	singleGetterWithEmbeddedTenant repo.SingleGetter
	pageableQuerierGlobal          repo.PageableQuerierGlobal
	updaterGlobal                  repo.UpdaterGlobal
	updaterWithEmbeddedTenant      repo.UpdaterGlobal
	deleterGlobal                  repo.DeleterGlobal
	deleterWithEmbeddedTenant      repo.Deleter
	conv                           EntityConverter
}

// NewRepository creates a new FormationTemplate repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:                        repo.NewCreatorGlobal(resource.FormationTemplate, tableName, tableColumns),
		existQuerierGlobal:             repo.NewExistQuerierGlobal(resource.FormationTemplate, tableName),
		singleGetterGlobal:             repo.NewSingleGetterGlobal(resource.FormationTemplate, tableName, tableColumns),
		singleGetterWithEmbeddedTenant: repo.NewSingleGetterWithEmbeddedTenant(tableName, tenantIDColumn, tableColumns),
		pageableQuerierGlobal:          repo.NewPageableQuerierGlobal(resource.FormationTemplate, tableName, tableColumns),
		updaterGlobal:                  repo.NewUpdaterGlobal(resource.FormationTemplate, tableName, updatableTableColumns, idTableColumns),
		updaterWithEmbeddedTenant:      repo.NewUpdaterWithEmbeddedTenant(resource.FormationTemplate, tableName, updatableTableColumns, tenantIDColumn, idTableColumns),
		deleterGlobal:                  repo.NewDeleterGlobal(resource.FormationTemplate, tableName),
		deleterWithEmbeddedTenant:      repo.NewDeleterWithEmbeddedTenant(tableName, tenantIDColumn),
		conv:                           conv,
	}
}

// Create creates a new FormationTemplate in the database with the fields in model
func (r *repository) Create(ctx context.Context, item *model.FormationTemplate) error {
	if item == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting formation template with ID: %s to entity", item.ID)
	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrapf(err, "while converting formation template with ID: %s", item.ID)
	}

	entity.CreatedAt = Now()
	log.C(ctx).Debugf("Persisting Formation Template entity with ID: %s to DB", item.ID)
	return r.creator.Create(ctx, entity)
}

// Get queries for a single FormationTemplate matching the given id
func (r *repository) Get(ctx context.Context, id string) (*model.FormationTemplate, error) {
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	result, err := r.conv.FromEntity(&entity)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Formation Template with ID %s", id)
	}

	return result, nil
}

// GetByNameAndTenant check if the provided tenant is not empty and tries to retrieve the formation template with tenant. If it fails with not found error then the formation template is retrieved globally.
// If the tenant is empty in the first place the formation template is retrieved globally.
func (r *repository) GetByNameAndTenant(ctx context.Context, templateName, tenantID string) (*model.FormationTemplate, error) {
	log.C(ctx).Debugf("Getting formation template by name: %q and tenant %q ...", templateName, tenantID)
	var entity Entity

	conditionsEqualName := repo.Conditions{repo.NewEqualCondition("name", templateName)}
	conditionsEqualNameAndNullTenant := repo.Conditions{
		repo.NewEqualCondition("name", templateName),
		repo.NewNullCondition(tenantIDColumn),
	}

	// If the call is with tenant but the query (select * from FT where name = ? and tenant_id = ?) returns NOT FOUND that means that there is no such tenant scoped FT and the call should get the global FT with that name.
	//
	// With this approach we allow the client to create a tenant scoped FT with the same name as a global FT - so when we get the FT first we will try to match it by name and tenant
	// and if there is no such FT, we will get the global one
	if tenantID == "" {
		if err := r.singleGetterGlobal.GetGlobal(ctx, conditionsEqualNameAndNullTenant, repo.NoOrderBy, &entity); err != nil {
			return nil, err
		}
	} else {
		if err := r.singleGetterWithEmbeddedTenant.Get(ctx, resource.FormationTemplate, tenantID, conditionsEqualName, repo.NoOrderBy, &entity); err != nil {
			if !apperrors.IsNotFoundError(err) {
				return nil, err
			}
			entity = Entity{}
			if err = r.singleGetterGlobal.GetGlobal(ctx, conditionsEqualNameAndNullTenant, repo.NoOrderBy, &entity); err != nil {
				return nil, err
			}
		}
	}

	result, err := r.conv.FromEntity(&entity)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Formation Template with name: %q", templateName)
	}

	return result, nil
}

// List queries for all FormationTemplate filtered by name, sorted by ID and paginated by the pageSize and cursor parameters
func (r *repository) List(ctx context.Context, filters []*labelfilter.LabelFilter, name *string, tenantID string, pageSize int, cursor string) (*model.FormationTemplatePage, error) {
	var entityCollection EntityCollection

	var conditions *repo.ConditionTree
	if tenantID == "" {
		conditions = &repo.ConditionTree{Operand: repo.NewNullCondition(tenantIDColumn)}
	} else {
		conditions = repo.Or(&repo.ConditionTree{Operand: repo.NewNullCondition(tenantIDColumn)}, &repo.ConditionTree{Operand: repo.NewEqualCondition(tenantIDColumn, tenantID)})
	}

	if name != nil {
		conditions = repo.And(conditions, &repo.ConditionTree{Operand: repo.NewEqualCondition(nameColumn, *name)})
	}

	// The tenant isolation for the formation template is handled with the 'tenant_id' where clause above.
	// That's why we use default UUID here, and in the label filter query we ignore it when the query is for formation template entity.
	var defaultUUIDTenant uuid.UUID

	filterSubquery, args, err := label.FilterQuery(model.FormationTemplateLabelableObject, label.IntersectSet, defaultUUIDTenant, filters)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query for formation template")
	}

	if filterSubquery != "" {
		conditions = repo.And(conditions, &repo.ConditionTree{Operand: repo.NewInConditionForSubQuery(idColumn, filterSubquery, args)})
	}

	page, totalCount, err := r.pageableQuerierGlobal.ListGlobalWithAdditionalConditions(ctx, pageSize, cursor, idColumn, &entityCollection, conditions)
	if err != nil {
		return nil, err
	}

	items := make([]*model.FormationTemplate, 0, len(entityCollection))

	for _, entity := range entityCollection {
		isModel, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting Formation Template entity with ID: %s", entity.ID)
		}

		items = append(items, isModel)
	}
	return &model.FormationTemplatePage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

// Update updates the FormationTemplate matching the ID of the input model
func (r *repository) Update(ctx context.Context, model *model.FormationTemplate) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	entity, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrapf(err, "while converting Formation Template with ID %s", model.ID)
	}
	currentTime := Now()
	entity.UpdatedAt = &currentTime

	if model.TenantID != nil {
		return r.updaterWithEmbeddedTenant.UpdateSingleGlobal(ctx, entity)
	}
	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

// Delete deletes a formation template with given ID and tenantID
func (r *repository) Delete(ctx context.Context, id, tenantID string) error {
	conditions := repo.Conditions{repo.NewEqualCondition(idColumn, id)}

	if tenantID == "" {
		conditions = append(conditions, repo.NewNullCondition(tenantIDColumn))
		if err := r.deleterGlobal.DeleteOneGlobal(ctx, conditions); apperrors.IsInternalServerError(err) {
			return apperrors.NewTenantRequiredError()
		}
	}

	return r.deleterWithEmbeddedTenant.DeleteOne(ctx, resource.FormationTemplate, tenantID, conditions)
}

// ExistsGlobal check if a formation template with given ID exists globally
func (r *repository) ExistsGlobal(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}
