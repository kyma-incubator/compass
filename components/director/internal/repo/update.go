package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// Updater missing godoc
type Updater interface {
	UpdateSingle(ctx context.Context, dbEntity interface{}) error
	UpdateSingleWithVersion(ctx context.Context, dbEntity interface{}) error
	SetIDColumns(idColumns []string)
	SetUpdatableColumns(updatableColumns []string)
	Clone() Updater
	TechnicalUpdate(ctx context.Context, dbEntity interface{}) error
}

// UpdaterGlobal missing godoc
type UpdaterGlobal interface {
	UpdateSingleGlobal(ctx context.Context, dbEntity interface{}) error
}

type universalUpdater struct {
	tableName        string
	resourceType     resource.Type
	updatableColumns []string
	tenantColumn     *string
	idColumns        []string
}

// NewUpdater missing godoc
func NewUpdater(resourceType resource.Type, tableName string, updatableColumns []string, tenantColumn string, idColumns []string) Updater {
	return &universalUpdater{
		resourceType:     resourceType,
		tableName:        tableName,
		updatableColumns: updatableColumns,
		tenantColumn:     &tenantColumn,
		idColumns:        idColumns,
	}
}

// NewUpdaterGlobal missing godoc
func NewUpdaterGlobal(resourceType resource.Type, tableName string, updatableColumns []string, idColumns []string) UpdaterGlobal {
	return &universalUpdater{
		resourceType:     resourceType,
		tableName:        tableName,
		updatableColumns: updatableColumns,
		idColumns:        idColumns,
	}
}

// SetIDColumns missing godoc
func (u *universalUpdater) SetIDColumns(idColumns []string) {
	u.idColumns = idColumns
}

// SetUpdatableColumns missing godoc
func (u *universalUpdater) SetUpdatableColumns(updatableColumns []string) {
	u.updatableColumns = updatableColumns
}

// UpdateSingle missing godoc
func (u *universalUpdater) UpdateSingle(ctx context.Context, dbEntity interface{}) error {
	return u.unsafeUpdateSingleWithFields(ctx, dbEntity, u.buildFieldsToSet(), false)
}

// UpdateSingleGlobal missing godoc
func (u *universalUpdater) UpdateSingleGlobal(ctx context.Context, dbEntity interface{}) error {
	return u.unsafeUpdateSingleWithFields(ctx, dbEntity, u.buildFieldsToSet(), true)
}

// UpdateSingle missing godoc
func (u *universalUpdater) UpdateSingleWithVersion(ctx context.Context, dbEntity interface{}) error {
	fieldsToSet := u.buildFieldsToSet()
	fieldsToSet = append(fieldsToSet, "version = version+1")

	if err := u.unsafeUpdateSingleWithFields(ctx, dbEntity, fieldsToSet, false); err != nil {
		if apperrors.IsConcurrentUpdate(err) {
			return apperrors.NewConcurrentUpdate()
		}
		return err
	}
	return nil
}

// TechnicalUpdate missing godoc
func (u *universalUpdater) TechnicalUpdate(ctx context.Context, dbEntity interface{}) error {
	entity, ok := dbEntity.(Entity)
	if ok && entity.GetDeletedAt().IsZero() {
		entity.SetUpdatedAt(time.Now())
		dbEntity = entity
	}
	return u.unsafeUpdateSingleWithFields(ctx, dbEntity, u.buildFieldsToSet(), false)
}

// Clone missing godoc
func (u *universalUpdater) Clone() Updater {
	var clonedUpdater universalUpdater

	clonedUpdater.tableName = u.tableName
	clonedUpdater.resourceType = u.resourceType
	clonedUpdater.updatableColumns = append(clonedUpdater.updatableColumns, u.updatableColumns...)
	clonedUpdater.tenantColumn = u.tenantColumn
	clonedUpdater.idColumns = append(clonedUpdater.idColumns, u.idColumns...)

	return &clonedUpdater
}

func (u *universalUpdater) unsafeUpdateSingleWithFields(ctx context.Context, dbEntity interface{}, fieldsToSet []string, isGlobal bool) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query, err := u.buildQuery(fieldsToSet, isGlobal)
	if err != nil {
		return err
	}

	log.C(ctx).Debugf("Executing DB query: %s", query)
	res, err := persist.NamedExecContext(ctx, query, dbEntity)
	if err = persistence.MapSQLError(ctx, err, u.resourceType, resource.Update, "while updating single entity from '%s' table", u.tableName); err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "while checking affected rows")
	}
	if affected != 1 {
		if u.resourceType == resource.BundleReference {
			return apperrors.NewCannotUpdateObjectInManyBundles()
		}
		return apperrors.NewInternalError(apperrors.ShouldUpdateSingleRowButUpdatedMsgF, affected)
	}

	return nil
}

func (u *universalUpdater) buildFieldsToSet() []string {
	fieldsToSet := make([]string, 0, len(u.updatableColumns)+1)
	for _, c := range u.updatableColumns {
		fieldsToSet = append(fieldsToSet, fmt.Sprintf("%s = :%s", c, c))
	}
	return fieldsToSet
}

func (u *universalUpdater) buildQuery(fieldsToSet []string, isGlobal bool) (string, error) {
	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("UPDATE %s SET %s", u.tableName, strings.Join(fieldsToSet, ", ")))
	if !isGlobal || len(u.idColumns) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}
	if !isGlobal {
		if err := writeEnumeratedConditions(&stmtBuilder, Conditions{/*NewTenantIsolationConditionWithPlaceholder(*u.tenantColumn, fmt.Sprintf(":%s", *u.tenantColumn), nil)*/}); err != nil {
			return "", errors.Wrap(err, "while writing enumerated conditions")
		}
		if len(u.idColumns) > 0 {
			stmtBuilder.WriteString(" AND")
		}
	}
	if len(u.idColumns) > 0 {
		var preparedIDColumns []string
		for _, idCol := range u.idColumns {
			preparedIDColumns = append(preparedIDColumns, fmt.Sprintf("%s = :%s", idCol, idCol))
		}
		stmtBuilder.WriteString(fmt.Sprintf(" %s", strings.Join(preparedIDColumns, " AND ")))
	}

	return stmtBuilder.String(), nil
}
