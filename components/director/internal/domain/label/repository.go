package label

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

const (
	tableName    string = "public.labels"
	tenantColumn string = "tenant_id"
)

var tableColumns = []string{"id", tenantColumn, "app_id", "runtime_id", "runtime_context_id", "key", "value"}

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToEntity(in model.Label) (Entity, error)
	FromEntity(in Entity) (model.Label, error)
}

type repository struct {
	upserter repo.Upserter
	lister   repo.Lister
	deleter  repo.Deleter
	conv     Converter
}

func NewRepository(conv Converter) *repository {
	return &repository{
		upserter: repo.NewUpserter(resource.Label, tableName, tableColumns, []string{tenantColumn, "coalesce(app_id, '00000000-0000-0000-0000-000000000000')", "coalesce(runtime_id, '00000000-0000-0000-0000-000000000000')", "coalesce(runtime_context_id, '00000000-0000-0000-0000-000000000000')", "key"}, []string{"value"}),
		lister:   repo.NewLister(resource.Label, tableName, tenantColumn, tableColumns),
		deleter:  repo.NewDeleter(resource.Label, tableName, tenantColumn),
		conv:     conv,
	}
}

func (r *repository) Upsert(ctx context.Context, label *model.Label) error {
	if label == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	labelEntity, err := r.conv.ToEntity(*label)
	if err != nil {
		return errors.Wrap(err, "while creating label entity from model")
	}

	return r.upserter.Upsert(ctx, labelEntity)
}

func (r *repository) GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching DB from context")
	}

	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE key = $1 AND %s = $2 AND tenant_id = $3`,
		strings.Join(tableColumns, ", "), tableName, labelableObjectField(objectType))

	var entity Entity
	err = persist.Get(&entity, stmt, key, objectID, tenant)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewNotFoundError(resource.Label, key)
		}
		return nil, errors.Wrap(err, "while getting Entity from DB")
	}

	labelModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Label entity to model")
	}

	return &labelModel, nil
}

func (r *repository) ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching DB from context")
	}

	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE  %s = $1 AND tenant_id = $2`,
		strings.Join(tableColumns, ", "), tableName, labelableObjectField(objectType))

	var entities []Entity
	err = persist.Select(&entities, stmt, objectID, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Labels from DB")
	}

	labelsMap := make(map[string]*model.Label)

	for _, entity := range entities {
		m, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Label entity to model")
		}

		labelsMap[m.Key] = &m
	}

	return labelsMap, nil
}

func (r *repository) ListByKey(ctx context.Context, tenant, key string) ([]*model.Label, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching DB from context")
	}

	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE key = $1 AND tenant_id = $2`,
		strings.Join(tableColumns, ", "), tableName)

	var entities []Entity
	err = persist.Select(&entities, stmt, key, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Labels from DB")
	}

	var labels []*model.Label

	for _, entity := range entities {
		m, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Label entity to model")
		}

		labels = append(labels, &m)
	}

	return labels, nil
}

func (r *repository) Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	stmt := fmt.Sprintf(`DELETE FROM %s WHERE key = $1 AND %s = $2 AND tenant_id = $3`, tableName, labelableObjectField(objectType))
	_, err = persist.Exec(stmt, key, objectID, tenant)

	return errors.Wrap(err, "while deleting the Label entity from database")
}

func (r *repository) DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	stmt := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1 AND tenant_id = $2`, tableName, labelableObjectField(objectType))
	_, err = persist.Exec(stmt, objectID, tenant)

	return errors.Wrapf(err, "while deleting all Label entities from database for %s %s", objectType, objectID)
}

func (r *repository) DeleteByKeyNegationPattern(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labelKeyPattern string) error {
	return r.deleter.DeleteMany(ctx, tenant, repo.Conditions{
		repo.NewEqualCondition(labelableObjectField(objectType), objectID),
		repo.NewEqualCondition(tenantColumn, tenant),
		repo.NewNotRegexConditionString("key", labelKeyPattern),
	})
}

func (r *repository) DeleteByKey(ctx context.Context, tenant string, key string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	stmt := fmt.Sprintf(`DELETE FROM %s WHERE key = $1 AND tenant_id = $2`, tableName)
	_, err = persist.Exec(stmt, key, tenant)
	if err != nil {
		return errors.Wrapf(err, `while deleting all Label entities from database with key "%s"`, key)
	}

	return nil
}

func (r *repository) GetRuntimesIDsByStringLabel(ctx context.Context, tenantID, key, value string) ([]string, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching persistence from context")
	}
	query := `SELECT LA.runtime_id FROM LABELS AS LA WHERE LA."key"=$1 AND value ?| array[$2] AND LA.tenant_id=$3 AND LA.runtime_ID IS NOT NULL;`

	var matchedRtmsIDs []string
	err = persist.Select(&matchedRtmsIDs, query, key, value, tenantID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching runtimes id which match selector")
	}
	return matchedRtmsIDs, nil
}

func (r *repository) GetScenarioLabelsForRuntimes(ctx context.Context, tenantID string, runtimesIDs []string) ([]model.Label, error) {
	if len(runtimesIDs) == 0 {
		return nil, apperrors.NewInvalidDataError("cannot execute query without runtimeIDs")
	}

	conditions := repo.Conditions{
		repo.NewEqualCondition("key", model.ScenariosKey),
		repo.NewInConditionForStringValues("runtime_id", runtimesIDs),
	}

	var labels Collection
	err := r.lister.List(ctx, tenantID, &labels, conditions...)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching runtimes scenarios")
	}

	var labelModels []model.Label
	for _, label := range labels {

		labelModel, err := r.conv.FromEntity(label)
		if err != nil {
			return nil, errors.Wrap(err, "while converting label entity to model")
		}
		labelModels = append(labelModels, labelModel)
	}
	return labelModels, nil
}

func (r *repository) GetRuntimeScenariosWhereLabelsMatchSelector(ctx context.Context, tenantID, selectorKey, selectorValue string) ([]model.Label, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching persistence from context")
	}

	query := `SELECT * FROM LABELS AS L WHERE l."key"='scenarios' AND l.tenant_id=$3 AND l.runtime_id in 
					(
				SELECT LA.runtime_id FROM LABELS AS LA WHERE LA."key"=$1 AND value ?| array[$2] AND LA.tenant_id=$3 AND LA.runtime_ID IS NOT NULL
			);`

	var lables []Entity
	err = persist.Select(&lables, query, selectorKey, selectorValue, tenantID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching runtimes scenarios associated with given selector")
	}

	var labelModels []model.Label
	for _, label := range lables {

		labelModel, err := r.conv.FromEntity(label)
		if err != nil {
			return nil, errors.Wrap(err, "while converting label entity to model")
		}
		labelModels = append(labelModels, labelModel)
	}
	return labelModels, nil
}

func labelableObjectField(objectType model.LabelableObject) string {
	switch objectType {
	case model.ApplicationLabelableObject:
		return "app_id"
	case model.RuntimeLabelableObject:
		return "runtime_id"
	case model.RuntimeContextLabelableObject:
		return "runtime_context_id"
	}

	return ""
}
