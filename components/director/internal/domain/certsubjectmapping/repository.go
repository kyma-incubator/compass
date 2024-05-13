package certsubjectmapping

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	tableName string = `public.cert_subject_mapping`
	idColumn  string = "id"
)

var (
	idTableColumns           = []string{"id"}
	updatableTableColumns    = []string{"subject", "consumer_type", "internal_consumer_id", "tenant_access_levels", "updated_at"}
	tableColumns             = []string{"id", "subject", "consumer_type", "internal_consumer_id", "tenant_access_levels", "created_at", "updated_at"}
	internalConsumerIDColumn = "internal_consumer_id"

	// Now is a function variable that returns the current time. It is used, so we could mock it in the tests.
	Now = time.Now
)

// entityConverter converts between the internal model and entity
//
//go:generate mockery --exported --name=entityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type entityConverter interface {
	ToEntity(in *model.CertSubjectMapping) (*Entity, error)
	FromEntity(entity *Entity) (*model.CertSubjectMapping, error)
}

type repository struct {
	creator               repo.CreatorGlobal
	existQuerierGlobal    repo.ExistQuerierGlobal
	singleGetterGlobal    repo.SingleGetterGlobal
	pageableQuerierGlobal repo.PageableQuerierGlobal
	updaterGlobal         repo.UpdaterGlobal
	deleterGlobal         repo.DeleterGlobal
	listerGlobal          repo.ListerGlobal
	conv                  entityConverter
}

// NewRepository creates a new CertSubjectMapping repository
func NewRepository(conv entityConverter) *repository {
	return &repository{
		creator:               repo.NewCreatorGlobal(resource.CertSubjectMapping, tableName, tableColumns),
		existQuerierGlobal:    repo.NewExistQuerierGlobal(resource.CertSubjectMapping, tableName),
		singleGetterGlobal:    repo.NewSingleGetterGlobal(resource.CertSubjectMapping, tableName, tableColumns),
		pageableQuerierGlobal: repo.NewPageableQuerierGlobal(resource.CertSubjectMapping, tableName, tableColumns),
		updaterGlobal:         repo.NewUpdaterGlobal(resource.CertSubjectMapping, tableName, updatableTableColumns, idTableColumns),
		deleterGlobal:         repo.NewDeleterGlobal(resource.CertSubjectMapping, tableName),
		listerGlobal:          repo.NewListerGlobal(resource.CertSubjectMapping, tableName, tableColumns),
		conv:                  conv,
	}
}

// Create creates a new certificate subject mapping in the database with the fields from the model
func (r *repository) Create(ctx context.Context, model *model.CertSubjectMapping) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting certificate subject mapping with ID: %s to entity", model.ID)
	entity, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrapf(err, "while converting certificate subject mapping with ID: %s", model.ID)
	}
	entity.CreatedAt = Now()

	log.C(ctx).Debugf("Persisting certificate mapping with ID: %s and subject: %s to DB", model.ID, model.Subject)
	return r.creator.Create(ctx, entity)
}

// Get queries for a single certificate subject mapping matching by a given ID
func (r *repository) Get(ctx context.Context, id string) (*model.CertSubjectMapping, error) {
	log.C(ctx).Debugf("Getting certificate mapping by ID: %s from DB", id)
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	result, err := r.conv.FromEntity(&entity)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting certificate subject mapping with ID: %s", id)
	}

	return result, nil
}

// Update updates the certificate subject mapping with the provided input model
func (r *repository) Update(ctx context.Context, model *model.CertSubjectMapping) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting certificate subject mapping with ID: %s to entity", model.ID)
	entity, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrapf(err, "while converting certificate subject mapping with ID: %s", model.ID)
	}
	currentTime := Now()
	entity.UpdatedAt = &currentTime

	log.C(ctx).Debugf("Updating certificate mapping with ID: %s and subject: %s", model.ID, model.Subject)
	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

// Delete deletes a certificate subject mapping with given ID
func (r *repository) Delete(ctx context.Context, id string) error {
	log.C(ctx).Debugf("Deleting certificate mapping with ID: %s from DB", id)
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Exists check if a certificate subject mapping with given ID exists
func (r *repository) Exists(ctx context.Context, id string) (bool, error) {
	log.C(ctx).Debugf("Check if certificate mapping with ID: %s exists", id)
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// List queries for all certificate subject mappings sorted by ID and paginated by the pageSize and cursor parameters
func (r *repository) List(ctx context.Context, pageSize int, cursor string) (*model.CertSubjectMappingPage, error) {
	log.C(ctx).Debug("Listing certificate subject mappings from DB")
	var entityCollection EntityCollection
	page, totalCount, err := r.pageableQuerierGlobal.ListGlobal(ctx, pageSize, cursor, idColumn, &entityCollection)
	if err != nil {
		return nil, err
	}

	items := make([]*model.CertSubjectMapping, 0, len(entityCollection))

	for _, entity := range entityCollection {
		result, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting certificate subject mapping with ID: %s", entity.ID)
		}

		items = append(items, result)
	}

	return &model.CertSubjectMappingPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

// ListAll lists all certificate subject mappings
func (r *repository) ListAll(ctx context.Context) ([]*model.CertSubjectMapping, error) {
	var entityCollection EntityCollection

	err := r.listerGlobal.ListGlobal(ctx, &entityCollection)

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entityCollection)
}

// ListByConsumerID queries all certificate subject mappings with given consumer id
func (r *repository) ListByConsumerID(ctx context.Context, consumerID string) ([]*model.CertSubjectMapping, error) {
	log.C(ctx).Debug("Listing certificate subject mappings from DB")
	var entityCollection EntityCollection

	condition := repo.NewEqualCondition(internalConsumerIDColumn, consumerID)

	if err := r.listerGlobal.ListGlobal(ctx, &entityCollection, condition); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entityCollection)
}

// DeleteByConsumerID deletes all certificate subject mappings for a specific consumer id
func (r *repository) DeleteByConsumerID(ctx context.Context, consumerID string) error {
	log.C(ctx).Debugf("Deleting all certificate subject mappings for consumer ID %q from DB", consumerID)
	return r.deleterGlobal.DeleteManyGlobal(ctx, repo.Conditions{repo.NewEqualCondition("internal_consumer_id", consumerID)})
}

func (r *repository) multipleFromEntities(entityCollection EntityCollection) ([]*model.CertSubjectMapping, error) {
	result := make([]*model.CertSubjectMapping, 0, len(entityCollection))

	for _, entity := range entityCollection {
		certSubjectMapping, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting certificate subject mapping with ID: %s", entity.ID)
		}

		result = append(result, certSubjectMapping)
	}

	return result, nil
}
