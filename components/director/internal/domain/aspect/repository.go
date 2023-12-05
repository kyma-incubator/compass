package aspect

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	aspectsTable                  string = `public.aspects`
	integrationDependencyTable    string = `integration_dependencies`
	integrationDependencyIDColumn string = "integration_dependency_id"
	appIDColumn                   string = "app_id"
)

var (
	aspectColumns = []string{"id", "app_id", "app_template_version_id", "integration_dependency_id", "title", "description", "mandatory", "support_multiple_providers",
		"api_resources", "ready", "created_at", "updated_at", "deleted_at", "error"}
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
	unionLister                         repo.UnionLister
	lister                              repo.Lister
	queryBuilderIntegrationDependencies repo.QueryBuilderGlobal

	conv AspectConverter
}

// NewRepository returns a new entity responsible for repo-layer Aspects operations.
func NewRepository(conv AspectConverter) *pgRepository {
	return &pgRepository{
		creator:                             repo.NewCreator(aspectsTable, aspectColumns),
		deleter:                             repo.NewDeleter(aspectsTable),
		unionLister:                         repo.NewUnionLister(aspectsTable, aspectColumns),
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

	if err := r.lister.List(ctx, resource.Aspect, tenant, &aspectCollection, condition); err != nil {
		return nil, err
	}

	aspects := make([]*model.Aspect, 0, aspectCollection.Len())
	for _, aspect := range aspectCollection {
		aspectModel := r.conv.FromEntity(&aspect)
		aspects = append(aspects, aspectModel)
	}

	return aspects, nil
}

// ListByApplicationIDs retrieves all Aspects matching an array of applicationIDs from the Compass storage.
func (r *pgRepository) ListByApplicationIDs(ctx context.Context, tenantID string, applicationIDs []string, pageSize int, cursor string) ([]*model.Aspect, map[string]int, error) {
	var aspectCollection AspectCollection
	orderByColumns := repo.OrderByParams{repo.NewAscOrderBy(integrationDependencyIDColumn), repo.NewAscOrderBy(appIDColumn)}

	counts, err := r.unionLister.List(ctx, resource.Aspect, tenantID, applicationIDs, appIDColumn, pageSize, cursor, orderByColumns, &aspectCollection)
	if err != nil {
		return nil, nil, err
	}

	aspects := make([]*model.Aspect, 0, len(aspectCollection))
	for _, d := range aspectCollection {
		entity := r.conv.FromEntity(&d)

		aspects = append(aspects, entity)
	}

	return aspects, counts, nil
}
