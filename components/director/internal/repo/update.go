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
	UpdateSingle(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error
	UpdateSingleWithVersion(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error
}

type universalUpdater struct {
	tableName        string
	resourceType     resource.Type
	updatableColumns []string
	tenantColumn     *string
	idColumns        []string
}

// NewUpdater missing godoc
func NewUpdater(tableName string, updatableColumns []string, idColumns []string) Updater {
	return &universalUpdater{
		tableName:        tableName,
		updatableColumns: updatableColumns,
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

// NewUpdaterWithEmbeddedTenant missing godoc
func NewUpdaterWithEmbeddedTenant(resourceType resource.Type, tableName string, updatableColumns []string, tenantColumn string, idColumns []string) UpdaterGlobal {
	return &universalUpdater{
		resourceType:     resourceType,
		tableName:        tableName,
		updatableColumns: updatableColumns,
		tenantColumn:     &tenantColumn,
		idColumns:        idColumns,
	}
}

// UpdateSingleWithVersion performs get of the resource with owner check before updating the entity with version.
// This is needed in order to distinguish the generic Unauthorized error due to the tenant has no owner access to the entity
// and the case of concurrent modification where the version differs. In both cases the affected rows would be 0.
func (u *universalUpdater) UpdateSingleWithVersion(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	var id string
	if identifiable, ok := dbEntity.(Identifiable); ok {
		id = identifiable.GetID()
	}

	if len(id) == 0 {
		return apperrors.NewInternalError("id cannot be empty, check if the entity implements Identifiable")
	}

	exister := NewExistQuerierWithOwnerCheck(u.tableName)
	found, err := exister.Exists(ctx, resourceType, tenant, Conditions{NewEqualCondition("id", id)})
	if err != nil {
		return err
	}
	if !found {
		return apperrors.NewInvalidOperationError("entity does not exist or caller tenant does not have owner access")
	}

	return u.updateSingleWithVersion(ctx, tenant, dbEntity, resourceType)
}

func (u *universalUpdater) UpdateSingle(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error {
	return u.unsafeUpdateSingleWithFields(ctx, dbEntity, tenant, u.buildFieldsToSet(), resourceType)
}

func (u *universalUpdater) UpdateSingleGlobal(ctx context.Context, dbEntity interface{}) error {
	return u.unsafeUpdateSingleWithFields(ctx, dbEntity, "", u.buildFieldsToSet(), u.resourceType)
}

func (u *universalUpdater) UpdateSingleWithVersionGlobal(ctx context.Context, dbEntity interface{}) error {
	return u.updateSingleWithVersion(ctx, "", dbEntity, u.resourceType)
}

// SetIDColumns missing godoc
func (u *universalUpdater) SetIDColumns(idColumns []string) {
	u.idColumns = idColumns
}

// SetUpdatableColumns missing godoc
func (u *universalUpdater) SetUpdatableColumns(updatableColumns []string) {
	u.updatableColumns = updatableColumns
}

// TechnicalUpdate missing godoc
func (u *universalUpdater) TechnicalUpdate(ctx context.Context, dbEntity interface{}) error {
	entity, ok := dbEntity.(Entity)
	if ok && entity.GetDeletedAt().IsZero() {
		entity.SetUpdatedAt(time.Now())
		dbEntity = entity
	}
	return u.unsafeUpdateSingleWithFields(ctx, dbEntity, "", u.buildFieldsToSet(), u.resourceType)
}

// Clone missing godoc
func (u *universalUpdater) Clone() UpdaterGlobal {
	var clonedUpdater universalUpdater

	clonedUpdater.tableName = u.tableName
	clonedUpdater.resourceType = u.resourceType
	clonedUpdater.updatableColumns = append(clonedUpdater.updatableColumns, u.updatableColumns...)
	clonedUpdater.tenantColumn = u.tenantColumn
	clonedUpdater.idColumns = append(clonedUpdater.idColumns, u.idColumns...)

	return &clonedUpdater
}

func (u *universalUpdater) updateSingleWithVersion(ctx context.Context, tenant string, dbEntity interface{}, resourceType resource.Type) error {
	fieldsToSet := u.buildFieldsToSet()
	fieldsToSet = append(fieldsToSet, "version = version+1")

	if err := u.unsafeUpdateSingleWithFields(ctx, dbEntity, tenant, fieldsToSet, resourceType); err != nil {
		if apperrors.IsConcurrentUpdate(err) {
			return apperrors.NewConcurrentUpdate()
		}
		return err
	}
	return nil
}

func (u *universalUpdater) unsafeUpdateSingleWithFields(ctx context.Context, dbEntity interface{}, tenant string, fieldsToSet []string, resourceType resource.Type) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
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
	if affected == 0 && (len(tenant) > 0 || u.tenantColumn != nil) {
		return apperrors.NewUnauthorizedError(apperrors.ShouldBeOwnerMsg)
	}
	if affected != 1 {
		if u.resourceType == resource.BundleReference {
			return apperrors.NewCannotUpdateObjectInManyBundles()
		}
		return apperrors.NewInternalError(apperrors.ShouldUpdateSingleRowButUpdatedMsgF, affected)
	}

	return nil
}

func (u *universalUpdater) buildQuery(fieldsToSet []string, tenant string, resourceType resource.Type) (string, error) {
	var stmtBuilder strings.Builder
	stmtBuilder.WriteString(fmt.Sprintf("UPDATE %s SET %s WHERE", u.tableName, strings.Join(fieldsToSet, ", ")))
	if len(u.idColumns) > 0 {
		var preparedIDColumns []string
		for _, idCol := range u.idColumns {
			preparedIDColumns = append(preparedIDColumns, fmt.Sprintf("%s = :%s", idCol, idCol))
		}
		stmtBuilder.WriteString(fmt.Sprintf(" %s", strings.Join(preparedIDColumns, " AND ")))
		if u.tenantColumn != nil || len(tenant) > 0 {
			stmtBuilder.WriteString(" AND")
		}
	}

	if u.tenantColumn != nil { // if embedded tenant
		stmtBuilder.WriteString(fmt.Sprintf(" %s = :%s", *u.tenantColumn, *u.tenantColumn))
	} else if len(tenant) > 0 { // if not global
		accessTable, ok := resourceType.TenantAccessTable()
		if !ok {
			return "", errors.Errorf("entity %s does not have access table", resourceType)
		}

		if _, err := uuid.Parse(tenant); err != nil { // SQL Injection protection
			return "", errors.Wrapf(err, "tenant_id %s should be UUID", tenant)
		}

		stmtBuilder.WriteString(fmt.Sprintf(" (id IN (SELECT %s FROM %s WHERE %s = '%s' AND %s = true)", M2MResourceIDColumn, accessTable, M2MTenantIDColumn, tenant, M2MOwnerColumn))

		if resourceType == resource.BundleInstanceAuth {
			stmtBuilder.WriteString(" OR owner_id = :owner_id") // TODO: <storage-redesign> externalize
		}

		stmtBuilder.WriteString(")")
	}

	return stmtBuilder.String(), nil
}

func (u *universalUpdater) buildFieldsToSet() []string {
	fieldsToSet := make([]string, 0, len(u.updatableColumns)+1)
	for _, c := range u.updatableColumns {
		fieldsToSet = append(fieldsToSet, fmt.Sprintf("%s = :%s", c, c))
	}
	return fieldsToSet
}
