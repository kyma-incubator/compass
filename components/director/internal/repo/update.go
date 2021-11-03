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
	UpdateSingle(ctx context.Context, dbEntity interface{}) error
	UpdateSingleWithVersion(ctx context.Context, dbEntity interface{}) error
	SetIDColumns(idColumns []string)
	SetUpdatableColumns(updatableColumns []string)
	TechnicalUpdate(ctx context.Context, dbEntity interface{}) error
	Clone() UpdaterGlobal
}

type Updater interface {
	UpdateSingle(ctx context.Context, tenant string, dbEntity interface{}) error
	UpdateSingleWithVersion(ctx context.Context, tenant string, dbEntity interface{}) error
}

// TODO <storage-redesign> Remove code duplication
type updater struct {
	tableName        string
	resourceType     resource.Type
	updatableColumns []string
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
	accessTable, ok := resourceType.TenantAccessTable()
	if !ok {
		return "", errors.Errorf("entity %s does not have access table", resourceType)
	}

	var stmtBuilder strings.Builder
	stmtBuilder.WriteString(fmt.Sprintf("UPDATE %s SET %s WHERE", u.tableName, strings.Join(fieldsToSet, ", ")))
	if len(u.idColumns) > 0 {
		var preparedIDColumns []string
		for _, idCol := range u.idColumns {
			preparedIDColumns = append(preparedIDColumns, fmt.Sprintf("%s = :%s", idCol, idCol))
		}
		stmtBuilder.WriteString(fmt.Sprintf(" %s AND ", strings.Join(preparedIDColumns, " AND ")))
	}

	if _, err := uuid.Parse(tenant); err != nil { // SQL Injection protection
		return "", errors.Wrapf(err, "tenant_id %s should be UUID", tenant)
	}

	stmtBuilder.WriteString(fmt.Sprintf("(id IN (SELECT %s FROM %s WHERE %s = '%s' AND %s = true)", M2MResourceIDColumn, accessTable, M2MTenantIDColumn, tenant, M2MOwnerColumn))

	if resourceType == resource.BundleInstanceAuth {
		stmtBuilder.WriteString(" OR owner_id = :owner_id") // TODO: <storage-redesign> externalize
	}

	stmtBuilder.WriteString(")")

	return stmtBuilder.String(), nil
}

func (u *updater) buildFieldsToSet() []string {
	fieldsToSet := make([]string, 0, len(u.updatableColumns)+1)
	for _, c := range u.updatableColumns {
		fieldsToSet = append(fieldsToSet, fmt.Sprintf("%s = :%s", c, c))
	}
	return fieldsToSet
}

type globalUpdater struct {
	tableName        string
	resourceType     resource.Type
	updatableColumns []string
	tenantColumn     *string
	idColumns        []string
}

type updaterBuilder struct {
	tableName        string
	resourceType     resource.Type
	updatableColumns []string
	tenantColumn     *string
	idColumns        []string
}

func NewGlobalUpdaterBuilderFor(resType resource.Type) *updaterBuilder {
	return &updaterBuilder{
		resourceType: resType,
	}
}

func (ub *updaterBuilder) WithTable(tableName string) *updaterBuilder {
	ub.tableName = tableName
	return ub
}

func (ub *updaterBuilder) WithUpdatableColumns(columns ...string) *updaterBuilder {
	ub.updatableColumns = columns
	return ub
}

func (ub *updaterBuilder) WithIDColumns(columns ...string) *updaterBuilder {
	ub.idColumns = columns
	return ub
}

// WithTenantColumn is there a tenant column in the table that needs to be checked as a tenant isolation condition prior updating.
// This is different from tenant scoped updater since it does not check any m2m table or view and relies on embedded tenant in the entity.
// This is especially useful for entities that are not inheritable between different tenants.
// NOTE: Omit the tenant column for entities that are top level - does not belong to any tenant.
func (ub *updaterBuilder) WithTenantColumn(column string) *updaterBuilder {
	ub.tenantColumn = &column
	return ub
}

func (ub *updaterBuilder) Build() UpdaterGlobal {
	return &globalUpdater{
		tableName:        ub.tableName,
		resourceType:     ub.resourceType,
		updatableColumns: ub.updatableColumns,
		tenantColumn:     ub.tenantColumn,
		idColumns:        ub.idColumns,
	}
}

// SetIDColumns missing godoc
func (u *globalUpdater) SetIDColumns(idColumns []string) {
	u.idColumns = idColumns
}

// SetUpdatableColumns missing godoc
func (u *globalUpdater) SetUpdatableColumns(updatableColumns []string) {
	u.updatableColumns = updatableColumns
}

// UpdateSingle missing godoc
func (u *globalUpdater) UpdateSingle(ctx context.Context, dbEntity interface{}) error {
	return u.unsafeUpdateSingleWithFields(ctx, dbEntity, u.buildFieldsToSet())
}

// UpdateSingleGlobal missing godoc
func (u *globalUpdater) UpdateSingleGlobal(ctx context.Context, dbEntity interface{}) error {
	return u.unsafeUpdateSingleWithFields(ctx, dbEntity, u.buildFieldsToSet())
}

// TechnicalUpdate missing godoc
func (u *globalUpdater) TechnicalUpdate(ctx context.Context, dbEntity interface{}) error {
	entity, ok := dbEntity.(Entity)
	if ok && entity.GetDeletedAt().IsZero() {
		entity.SetUpdatedAt(time.Now())
		dbEntity = entity
	}
	return u.unsafeUpdateSingleWithFields(ctx, dbEntity, u.buildFieldsToSet())
}

// UpdateSingleWithVersion missing godoc
func (u *globalUpdater) UpdateSingleWithVersion(ctx context.Context, dbEntity interface{}) error {
	fieldsToSet := u.buildFieldsToSet()
	fieldsToSet = append(fieldsToSet, "version = version+1")

	if err := u.unsafeUpdateSingleWithFields(ctx, dbEntity, fieldsToSet); err != nil {
		if apperrors.IsConcurrentUpdate(err) {
			return apperrors.NewConcurrentUpdate()
		}
		return err
	}
	return nil
}

// Clone missing godoc
func (u *globalUpdater) Clone() UpdaterGlobal {
	var clonedUpdater globalUpdater

	clonedUpdater.tableName = u.tableName
	clonedUpdater.resourceType = u.resourceType
	clonedUpdater.updatableColumns = append(clonedUpdater.updatableColumns, u.updatableColumns...)
	clonedUpdater.tenantColumn = u.tenantColumn
	clonedUpdater.idColumns = append(clonedUpdater.idColumns, u.idColumns...)

	return &clonedUpdater
}

func (u *globalUpdater) unsafeUpdateSingleWithFields(ctx context.Context, dbEntity interface{}, fieldsToSet []string) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query, err := u.buildQuery(fieldsToSet)
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

func (u *globalUpdater) buildFieldsToSet() []string {
	fieldsToSet := make([]string, 0, len(u.updatableColumns)+1)
	for _, c := range u.updatableColumns {
		fieldsToSet = append(fieldsToSet, fmt.Sprintf("%s = :%s", c, c))
	}
	return fieldsToSet
}

func (u *globalUpdater) buildQuery(fieldsToSet []string) (string, error) {
	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("UPDATE %s SET %s", u.tableName, strings.Join(fieldsToSet, ", ")))
	if u.tenantColumn != nil || len(u.idColumns) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}
	if u.tenantColumn != nil {
		stmtBuilder.WriteString(fmt.Sprintf(" %s = :%s", *u.tenantColumn, *u.tenantColumn))
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
