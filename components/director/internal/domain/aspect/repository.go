package aspect

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	aspectsTable                  string = `public.aspects`
	integrationDependencyTable    string = `integration_dependencies`
	visibilityColumn              string = "visibility"
	internalVisibilityScope       string = "internal_visibility:read"
	publicVisibilityValue         string = "public"
	integrationDependencyIDColumn string = "integration_dependency_id"
	appIDColumn                   string = "app_id"
)

var (
	aspectColumns = []string{"id", "app_id", "app_template_version_id", "integration_dependency_id", "title", "description", "mandatory", "support_multiple_providers",
		"api_resources", "event_resources", "ready", "created_at", "updated_at", "deleted_at", "error"}
)

// AspectConverter converts Aspects between the model.Aspect service-layer representation and the repo-layer representation Entity.
//
//go:generate mockery --name=AspectConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type AspectConverter interface {
	FromEntity(entity *Entity) *model.Aspect
	ToEntity(aspectModel *model.Aspect) *Entity
}

type pgRepository struct {
	creator                             repo.Creator
	deleter                             repo.Deleter
	unionLister                         repo.UnionListerGlobal
	lister                              repo.Lister
	queryBuilderIntegrationDependencies repo.QueryBuilderGlobal

	conv AspectConverter
}

// NewRepository returns a new entity responsible for repo-layer Aspects operations.
func NewRepository(conv AspectConverter) *pgRepository {
	return &pgRepository{
		creator:                             repo.NewCreator(aspectsTable, aspectColumns),
		deleter:                             repo.NewDeleter(aspectsTable),
		unionLister:                         repo.NewUnionListerGlobal(resource.Aspect, aspectsTable, []string{}),
		lister:                              repo.NewLister(aspectsTable, aspectColumns),
		queryBuilderIntegrationDependencies: repo.NewQueryBuilderGlobal(resource.IntegrationDependency, integrationDependencyTable, []string{"id"}),

		conv: conv,
	}
}

// AspectCollection is an array of Entities
type AspectCollection []Entity

// Len returns the length of the collection
func (r AspectCollection) Len() int {
	return len(r)
}

// Create creates an Aspect for Integration Dependency.
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.Aspect) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	err := r.creator.Create(ctx, resource.Aspect, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

// DeleteByIntegrationDependencyID deletes Aspects for an Integration Dependency with given ID.
func (r *pgRepository) DeleteByIntegrationDependencyID(ctx context.Context, tenant string, integrationDependencyID string) error {
	return r.deleter.DeleteMany(ctx, resource.Aspect, tenant, repo.Conditions{repo.NewEqualCondition(integrationDependencyIDColumn, integrationDependencyID)})
}

// ListByIntegrationDependencyID lists Aspects by Integration Dependency id
func (r *pgRepository) ListByIntegrationDependencyID(ctx context.Context, tenant string, integrationDependencyID string) ([]*model.Aspect, error) {
	aspectCollection := AspectCollection{}

	condition := repo.NewEqualCondition(integrationDependencyIDColumn, integrationDependencyID)

	if err := r.lister.List(ctx, resource.IntegrationDependency, tenant, &aspectCollection, condition); err != nil {
		return nil, err
	}

	aspects := make([]*model.Aspect, 0, aspectCollection.Len())
	for _, aspect := range aspectCollection {
		apiModel := r.conv.FromEntity(&aspect)
		aspects = append(aspects, apiModel)
	}

	return aspects, nil
}

// ListByApplicationIDs retrieves all Aspects matching an array of applicationIDs from the Compass storage.
func (r *pgRepository) ListByApplicationIDs(ctx context.Context, applicationIDs []string, pageSize int, cursor string) ([]*model.Aspect, map[string]int, error) {
	unionLister := r.unionLister.Clone()
	unionLister.SetSelectedColumns(aspectColumns)

	isInternalVisibilityScopePresent, err := scope.Contains(ctx, internalVisibilityScope)
	if err != nil {
		log.C(ctx).Info("No scopes are present in the context meaning the flow is not user-initiated. Processing Integration Dependencies without visibility check...")
		isInternalVisibilityScopePresent = true
	}

	queryBuilder := r.queryBuilderIntegrationDependencies

	var conditions repo.Conditions
	if !isInternalVisibilityScopePresent {
		log.C(ctx).Info("No internal visibility scope is present in the context. Processing only public Integration Dependencies")

		query, args, err := queryBuilder.BuildQueryGlobal(false, repo.NewEqualCondition(visibilityColumn, publicVisibilityValue))
		if err != nil {
			return nil, nil, err
		}
		conditions = append(conditions, repo.NewInConditionForSubQuery(integrationDependencyIDColumn, query, args))
	}

	log.C(ctx).Infof("Internal visibility scope is present in the context. Processing Integration Dependencies without visibility check...")
	conditions = append(conditions, repo.NewNotNullCondition(integrationDependencyIDColumn))

	orderByColumns := repo.OrderByParams{repo.NewAscOrderBy(integrationDependencyIDColumn), repo.NewAscOrderBy(appIDColumn)}

	var objectApplicationIDs AspectCollection
	counts, err := unionLister.ListGlobal(ctx, applicationIDs, appIDColumn, pageSize, cursor, orderByColumns, &objectApplicationIDs, conditions...)

	if err != nil {
		return nil, nil, err
	}

	aspects := make([]*model.Aspect, 0, len(objectApplicationIDs))
	for _, d := range objectApplicationIDs {
		entity := r.conv.FromEntity(&d)

		aspects = append(aspects, entity)
	}

	return aspects, counts, nil
}
