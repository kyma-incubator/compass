package repo

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// UpdaterGlobal missing godoc
type UpdaterGlobal interface {
	UpdateSingleGlobal(ctx context.Context, dbEntity interface{}) error
	UpdateSingleWithVersionGlobal(ctx context.Context, dbEntity interface{}) error
	SetIDColumns(idColumns []string)
	SetUpdatableColumns(updatableColumns []string)
	TechnicalUpdate(ctx context.Context, dbEntity interface{}) error
	Clone() UpdaterGlobal
}

type Updater interface {
	UpdateSingle(ctx context.Context, tenant string, dbEntity interface{}) error
	UpdateSingleWithVersion(ctx context.Context, tenant string, dbEntity interface{}) error
}

type updater struct {
	tableName        string
	resourceType     resource.Type
	updatableColumns []string
	tenantColumn     *string
	idColumns        []string
}

// NewUpdater missing godoc
func NewUpdater(resourceType resource.Type, tableName string, updatableColumns []string, idColumns []string) Updater {
	return &updater{
		resourceType:     resourceType,
		tableName:        tableName,
		updatableColumns: updatableColumns,
		idColumns:        idColumns,
	}
}

// NewUpdaterGlobal missing godoc
func NewUpdaterGlobal(resourceType resource.Type, tableName string, updatableColumns []string, idColumns []string) UpdaterGlobal {
	return &updater{
		resourceType:     resourceType,
		tableName:        tableName,
		updatableColumns: updatableColumns,
		idColumns:        idColumns,
	}
}

// NewUpdaterWithEmbeddedTenant missing godoc
func NewUpdaterWithEmbeddedTenant(resourceType resource.Type, tableName string, updatableColumns []string, tenantColumn string, idColumns []string) UpdaterGlobal {
	return &updater{
		resourceType:     resourceType,
		tableName:        tableName,
		updatableColumns: updatableColumns,
		tenantColumn:     &tenantColumn,
		idColumns:        idColumns,
	}
}

// UpdateSingleWithVersion missing godoc
func (u *updater) UpdateSingleWithVersion(ctx context.Context, tenant string, dbEntity interface{}) error {
	fieldsToSet := u.buildFieldsToSet()
	fieldsToSet = append(fieldsToSet, "version = version+1")

	if err := u.unsafeUpdateSingleWithFields(ctx, dbEntity, tenant, fieldsToSet); err != nil {
		if apperrors.IsConcurrentUpdate(err) {
			return apperrors.NewConcurrentUpdate()
		}
		return err
	}
	return nil
}

func (u *updater) UpdateSingle(ctx context.Context, tenant string, dbEntity interface{}) error {
	return u.unsafeUpdateSingleWithFields(ctx, dbEntity, tenant, u.buildFieldsToSet())
}

func (u *updater) UpdateSingleGlobal(ctx context.Context, dbEntity interface{}) error {
	return u.unsafeUpdateSingleWithFields(ctx, dbEntity, "", u.buildFieldsToSet())
}

func (u *updater) UpdateSingleWithVersionGlobal(ctx context.Context, dbEntity interface{}) error {
	fieldsToSet := u.buildFieldsToSet()
	fieldsToSet = append(fieldsToSet, "version = version+1")

	if err := u.unsafeUpdateSingleWithFields(ctx, dbEntity, "", fieldsToSet); err != nil {
		if apperrors.IsConcurrentUpdate(err) {
			return apperrors.NewConcurrentUpdate()
		}
		return err
	}
	return nil
}

// SetIDColumns missing godoc
func (u *updater) SetIDColumns(idColumns []string) {
	u.idColumns = idColumns
}

// SetUpdatableColumns missing godoc
func (u *updater) SetUpdatableColumns(updatableColumns []string) {
	u.updatableColumns = updatableColumns
}

// TechnicalUpdate missing godoc
func (u *updater) TechnicalUpdate(ctx context.Context, dbEntity interface{}) error {
	entity, ok := dbEntity.(Entity)
	if ok && entity.GetDeletedAt().IsZero() {
		entity.SetUpdatedAt(time.Now())
		dbEntity = entity
	}
	return u.unsafeUpdateSingleWithFields(ctx, dbEntity, "", u.buildFieldsToSet())
}

// Clone missing godoc
func (u *updater) Clone() UpdaterGlobal {
	var clonedUpdater updater

	clonedUpdater.tableName = u.tableName
	clonedUpdater.resourceType = u.resourceType
	clonedUpdater.updatableColumns = append(clonedUpdater.updatableColumns, u.updatableColumns...)
	clonedUpdater.tenantColumn = u.tenantColumn
	clonedUpdater.idColumns = append(clonedUpdater.idColumns, u.idColumns...)

	return &clonedUpdater
}

func (u *updater) unsafeUpdateSingleWithFields(ctx context.Context, dbEntity interface{}, tenant string, fieldsToSet []string) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	resourceType := u.resourceType
	if multiRefEntity, ok := dbEntity.(MultiRefEntity); ok {
		resourceType = multiRefEntity.GetRefSpecificResourceType()
	}

	query, err := u.buildQuery(fieldsToSet, tenant, resourceType)
	if err != nil {
		return err
	}

	log.C(ctx).Debugf("Executing DB query: %s", query)
	res, err := persist.NamedExecContext(ctx, query, dbEntity)
	if err = persistence.MapSQLError(ctx, err, resourceType, resource.Update, "while updating single entity from '%s' table", u.tableName); err != nil {
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

func (u *updater) buildQuery(fieldsToSet []string, tenant string, resourceType resource.Type) (string, error) {
	var stmtBuilder strings.Builder
	stmtBuilder.WriteString(fmt.Sprintf("UPDATE %s SET %s WHERE", u.tableName, strings.Join(fieldsToSet, ", ")))
	if len(u.idColumns) > 0 {
		var preparedIDColumns []string
		for _, idCol := range u.idColumns {
			preparedIDColumns = append(preparedIDColumns, fmt.Sprintf("%s = :%s", idCol, idCol))
		}
		stmtBuilder.WriteString(fmt.Sprintf(" %s", strings.Join(preparedIDColumns, " AND ")))
	}

	if len(tenant) > 0 { // if not global
		stmtBuilder.WriteString(" AND")
		if u.tenantColumn != nil { // if embedded tenant
			stmtBuilder.WriteString(fmt.Sprintf(" %s = :%s", *u.tenantColumn, *u.tenantColumn))
		} else { // tenant in m2m table
			accessTable, ok := resourceType.TenantAccessTable()
			if !ok {
				return "", errors.Errorf("entity %s does not have access table", resourceType)
			}

			if _, err := uuid.Parse(tenant); err != nil { // SQL Injection protection
				return "", errors.Wrapf(err, "tenant_id %s should be UUID", tenant)
			}

			stmtBuilder.WriteString(fmt.Sprintf("(id IN (SELECT %s FROM %s WHERE %s = '%s' AND %s = true)", M2MResourceIDColumn, accessTable, M2MTenantIDColumn, tenant, M2MOwnerColumn))

			if resourceType == resource.BundleInstanceAuth {
				stmtBuilder.WriteString(" OR owner_id = :owner_id") // TODO: <storage-redesign> externalize
			}

			stmtBuilder.WriteString(")")
		}
	}

	return stmtBuilder.String(), nil
}

func (u *updater) buildFieldsToSet() []string {
	fieldsToSet := make([]string, 0, len(u.updatableColumns)+1)
	for _, c := range u.updatableColumns {
		fieldsToSet = append(fieldsToSet, fmt.Sprintf("%s = :%s", c, c))
	}
	return fieldsToSet
}
