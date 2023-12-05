package tenant_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var selfRegDistinguishLabel = "selfRegDistinguishLabel"

func TestPgRepository_Upsert(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? ) ON CONFLICT ( external_tenant ) DO UPDATE SET external_name=EXCLUDED.external_name`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingrepo := tenant.NewRepository(mockConverter)

		// WHEN
		_, err := tenantMappingrepo.Upsert(ctx, *tenantMappingModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when upserting", func(t *testing.T) {
		// GIVEN
		tenantModel := newModelBusinessTenantMapping(testID, testName, nil)
		tenantEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", tenantModel).Return(tenantEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? ) ON CONFLICT ( external_tenant ) DO UPDATE SET external_name=EXCLUDED.external_name`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantEntity)...).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		_, err := tenantMappingRepo.Upsert(ctx, *tenantModel)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "while upserting business tenant mapping: Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_UnsafeCreate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )  ON CONFLICT ( external_tenant ) DO NOTHING`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingrepo := tenant.NewRepository(mockConverter)

		// WHEN
		_, err := tenantMappingrepo.UnsafeCreate(ctx, *tenantMappingModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when creating", func(t *testing.T) {
		// GIVEN
		tenantModel := newModelBusinessTenantMapping(testID, testName, nil)
		tenantEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", tenantModel).Return(tenantEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )  ON CONFLICT ( external_tenant ) DO NOTHING`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantEntity)...).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		_, err := tenantMappingRepo.UnsafeCreate(ctx, *tenantModel)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "while creating business tenant mapping: Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
		tenantMappingModelWithParents := newModelBusinessTenantMapping(testID, testName, []string{testParentID, testParentID2})
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: testID, parentID: testParentID},
			{tenantID: testID, parentID: testParentID2},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs(testID).
			WillReturnRows(parentRowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.Get(ctx, testID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, tenantMappingModelWithParents, result)
	})

	t.Run("Error when get", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenantEntity.Inactive).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.Get(ctx, testID)

		// THEN
		require.Error(t, err)
		require.Nil(t, result)
	})
	t.Run("Error when get parents", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs(testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.Get(ctx, testID)

		// THEN
		require.Error(t, err)
		require.Nil(t, result)
	})
}

func TestPgRepository_GetByExternalTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
		tenantMappingModelWithParents := newModelBusinessTenantMapping(testID, testName, []string{testParentID, testParentID2})
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: testID, parentID: testParentID},
			{tenantID: testID, parentID: testParentID2},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
			WithArgs(testExternal, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs(testID).
			WillReturnRows(parentRowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.GetByExternalTenant(ctx, testExternal)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, tenantMappingModelWithParents, result)
	})

	t.Run("Error when getting", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $ `)).
			WithArgs(testExternal, tenantEntity.Inactive).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.GetByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, result)
	})

	t.Run("Error when getting parents", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $ `)).
			WithArgs(testExternal, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs(testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.GetByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, result)
	})
}

func TestPgRepository_Exists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.business_tenant_mappings WHERE id = $1`)).
			WithArgs(testID).
			WillReturnRows(testdb.RowWhenObjectExist())

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.Exists(ctx, testID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result)
	})

	t.Run("Error when checking existence", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.business_tenant_mappings WHERE id = $1`)).
			WithArgs(testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.Exists(ctx, testID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		assert.False(t, result)
	})
}

func TestPgRepository_ExistsByExternalTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.business_tenant_mappings WHERE external_tenant = $1`)).
			WithArgs(testExternal).
			WillReturnRows(testdb.RowWhenObjectExist())

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.ExistsByExternalTenant(ctx, testExternal)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result)
	})

	t.Run("Error when checking existence", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.business_tenant_mappings WHERE external_tenant = $1`)).
			WithArgs(testExternal).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.ExistsByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		assert.False(t, result)
	})
}

func TestPgRepository_ExistsSubscribed(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.business_tenant_mappings WHERE (type = $1 AND (id IN (SELECT tenant_id FROM tenant_runtime_contexts ) OR id IN (SELECT tenant_id FROM tenant_applications WHERE id IN (SELECT id FROM applications WHERE app_template_id IN (SELECT app_template_id FROM labels WHERE key = $2 AND app_template_id IS NOT NULL)))) AND id = $3)`)).
			WithArgs(tenantEntity.Subaccount, selfRegDistinguishLabel, testID).
			WillReturnRows(testdb.RowWhenObjectExist())

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.ExistsSubscribed(ctx, testID, selfRegDistinguishLabel)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result)
	})

	t.Run("Error when checking existence", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.business_tenant_mappings WHERE (type = $1 AND (id IN (SELECT tenant_id FROM tenant_runtime_contexts ) OR id IN (SELECT tenant_id FROM tenant_applications WHERE id IN (SELECT id FROM applications WHERE app_template_id IN (SELECT app_template_id FROM labels WHERE key = $2 AND app_template_id IS NOT NULL)))) AND id = $3)`)).
			WithArgs(tenantEntity.Subaccount, selfRegDistinguishLabel, testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.ExistsSubscribed(ctx, testID, selfRegDistinguishLabel)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		assert.False(t, result)
	})
}

func TestPgRepository_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN

		initializedVal := true
		notInitializedVal := false

		tenantModels := []*model.BusinessTenantMapping{
			newModelBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal, nil),
			newModelBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal, nil),
			newModelBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal, nil),
		}

		resultingTenantModels := []*model.BusinessTenantMapping{
			newModelBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal, []string{testParentID, testParentID2}),
			newModelBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal, []string{testParentID2}),
			newModelBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal, nil),
		}

		tenantEntities := []*tenantEntity.Entity{
			newEntityBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal),
			newEntityBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal),
			newEntityBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal),
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantEntities[0]).Return(tenantModels[0]).Once()
		mockConverter.On("FromEntity", tenantEntities[1]).Return(tenantModels[1]).Once()
		mockConverter.On("FromEntity", tenantEntities[2]).Return(tenantModels[2]).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
			{sqlRow: sqlRow{id: "id1", name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &initializedVal},
			{sqlRow: sqlRow{id: "id2", name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
			{sqlRow: sqlRow{id: "id3", name: "name3", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
		})
		parentRowsForTenant1ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: "id1", parentID: testParentID},
			{tenantID: "id1", parentID: testParentID2},
		})
		parentRowsForTenant2ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: "id2", parentID: testParentID2},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT t.id, t.external_name, t.external_tenant, t.type, t.provider_name, t.status, ld.tenant_id IS NOT NULL AS initialized FROM public.business_tenant_mappings t LEFT JOIN public.label_definitions ld ON t.id=ld.tenant_id WHERE t.status = $1 ORDER BY initialized DESC, t.external_name ASC`)).
			WithArgs(tenantEntity.Active).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs("id1").
			WillReturnRows(parentRowsForTenant1ToReturn)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs("id2").
			WillReturnRows(parentRowsForTenant2ToReturn)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs("id3").
			WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.List(ctx)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, resultingTenantModels, result)
	})

	t.Run("Error when listing", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT t.id, t.external_name, t.external_tenant, t.type, t.provider_name, t.status, ld.tenant_id IS NOT NULL AS initialized FROM public.business_tenant_mappings t LEFT JOIN public.label_definitions ld ON t.id=ld.tenant_id WHERE t.status = $1 ORDER BY initialized DESC, t.external_name ASC`)).
			WithArgs(tenantEntity.Active).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.List(ctx)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		require.Nil(t, result)
	})

	t.Run("Error when missing persistence context", func(t *testing.T) {
		// GIVEN
		repo := tenant.NewRepository(nil)
		ctx := context.TODO()

		// WHEN
		_, err := repo.List(ctx)

		// THEN
		require.EqualError(t, err, "while fetching persistence from context: Internal Server Error: unable to fetch database from context")
	})
}

func TestPgRepository_ListPageBySearchTerm(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		searchTerm := "name"
		first := 10
		endCursor := ""
		initializedVal := true
		notInitializedVal := false

		tenantModels := []*model.BusinessTenantMapping{
			newModelBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal, nil),
			newModelBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal, nil),
			newModelBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal, nil),
		}

		resultingTenantModels := []*model.BusinessTenantMapping{
			newModelBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal, []string{testParentID, testParentID2}),
			newModelBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal, []string{testParentID2}),
			newModelBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal, nil),
		}

		tenantEntities := []*tenantEntity.Entity{
			newEntityBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal),
			newEntityBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal),
			newEntityBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal),
		}

		tenantPage := &model.BusinessTenantMappingPage{
			Data: tenantModels,
			PageInfo: &pagination.Page{
				StartCursor: "",
				EndCursor:   "",
				HasNextPage: false,
			},
			TotalCount: len(resultingTenantModels),
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantEntities[0]).Return(tenantModels[0]).Once()
		mockConverter.On("FromEntity", tenantEntities[1]).Return(tenantModels[1]).Once()
		mockConverter.On("FromEntity", tenantEntities[2]).Return(tenantModels[2]).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
			{sqlRow: sqlRow{id: "id1", name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &initializedVal},
			{sqlRow: sqlRow{id: "id2", name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
			{sqlRow: sqlRow{id: "id3", name: "name3", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
		})
		parentRowsForTenant1ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: "id1", parentID: testParentID},
			{tenantID: "id1", parentID: testParentID2},
		})
		parentRowsForTenant2ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: "id2", parentID: testParentID2},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE (status = $1 AND (id::text ILIKE $2 OR external_name ILIKE $3 OR external_tenant ILIKE $4)) ORDER BY external_name LIMIT 10 OFFSET 0`)).
			WithArgs(tenantEntity.Active, "%name%", "%name%", "%name%").
			WillReturnRows(rowsToReturn)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM public.business_tenant_mappings WHERE (status = $1 AND (id::text ILIKE $2 OR external_name ILIKE $3 OR external_tenant ILIKE $4))`)).
			WithArgs(tenantEntity.Active, "%name%", "%name%", "%name%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs("id1").
			WillReturnRows(parentRowsForTenant1ToReturn)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs("id2").
			WillReturnRows(parentRowsForTenant2ToReturn)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs("id3").
			WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListPageBySearchTerm(ctx, searchTerm, first, endCursor)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, tenantPage, result)
	})

	t.Run("Error when listing", func(t *testing.T) {
		// GIVEN
		searchTerm := "name"
		first := 10
		endCursor := ""

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE (status = $1 AND (id::text ILIKE $2 OR external_name ILIKE $3 OR external_tenant ILIKE $4)) ORDER BY external_name LIMIT 10 OFFSET 0`)).
			WithArgs(tenantEntity.Active, "%name%", "%name%", "%name%").
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListPageBySearchTerm(ctx, searchTerm, first, endCursor)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, result)
	})

	t.Run("Error when missing persistence context", func(t *testing.T) {
		// GIVEN
		searchTerm := "name"
		first := 10
		endCursor := ""
		repo := tenant.NewRepository(nil)
		ctx := context.TODO()

		// WHEN
		_, err := repo.ListPageBySearchTerm(ctx, searchTerm, first, endCursor)

		// THEN
		require.EqualError(t, err, "while listing tenants from DB: Internal Server Error: unable to fetch database from context")
	})
}

func TestPgRepository_ListByExternalTenants(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		initializedVal := true
		tntID := id()
		externalTntID := id()
		tntModel := &model.BusinessTenantMapping{ID: tntID, ExternalTenant: externalTntID, Initialized: &initializedVal}
		tntEntity := &tenantEntity.Entity{ID: tntModel.ID, ExternalTenant: tntModel.ExternalTenant, Initialized: &initializedVal}

		resultingTntModel := &model.BusinessTenantMapping{ID: tntID, ExternalTenant: externalTntID, Initialized: &initializedVal, Parents: []string{testParentID}}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tntEntity).Return(tntModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{{sqlRow: sqlRow{id: tntModel.ID, externalTenant: tntModel.ExternalTenant}, initialized: &initializedVal}})
		parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: tntID, parentID: testParentID},
		})

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1)`)).
			WithArgs(tntModel.ExternalTenant).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs(tntID).
			WillReturnRows(parentRowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListByExternalTenants(ctx, []string{tntModel.ExternalTenant})

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, []*model.BusinessTenantMapping{resultingTntModel}, result)
	})

	t.Run("Success when high load of tenants is requested", func(t *testing.T) {
		// GIVEN
		initializedVal := true
		tntID := id()
		externalTntID := id()
		tntModel := &model.BusinessTenantMapping{ID: tntID, ExternalTenant: externalTntID, Initialized: &initializedVal}
		tntEntity := &tenantEntity.Entity{ID: tntModel.ID, ExternalTenant: tntModel.ExternalTenant, Initialized: &initializedVal}

		resultingTntModel := &model.BusinessTenantMapping{ID: tntID, ExternalTenant: externalTntID, Initialized: &initializedVal, Parents: []string{testParentID}}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tntEntity).Return(tntModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{{sqlRow: sqlRow{id: tntModel.ID, externalTenant: tntModel.ExternalTenant}, initialized: &initializedVal}})
		parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: tntID, parentID: testParentID},
		})
		// get first chunk of tenant IDs
		firstChunkIDs := chunkSizedTenantIDs(49999)
		firstChunkIDs = append(firstChunkIDs, tntModel.ExternalTenant)
		firstChunkQuery, firstChunkQueryArgs := buildQueryWithTenantIDs(firstChunkIDs)
		dbMock.ExpectQuery(regexp.QuoteMeta(firstChunkQuery)).
			WithArgs(firstChunkQueryArgs...).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs(tntID).
			WillReturnRows(parentRowsToReturn)

		// get second chunk of tenant IDs
		secondChunkIDs := chunkSizedTenantIDs(100)
		secondChunkQuery, secondChunkIDsChunkQueryArgs := buildQueryWithTenantIDs(secondChunkIDs)
		dbMock.ExpectQuery(regexp.QuoteMeta(secondChunkQuery)).
			WithArgs(secondChunkIDsChunkQueryArgs...).
			WillReturnRows(fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{}))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListByExternalTenants(ctx, append(firstChunkIDs, secondChunkIDs...))

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, []*model.BusinessTenantMapping{resultingTntModel}, result)
	})

	t.Run("Error when listing", func(t *testing.T) {
		// GIVEN
		externalTenantID := id()

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1)`)).
			WithArgs(externalTenantID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListByExternalTenants(ctx, []string{externalTenantID})

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Unexpected error while executing SQL query")
		require.Nil(t, result)
	})

	t.Run("Error when missing persistence context", func(t *testing.T) {
		// GIVEN
		repo := tenant.NewRepository(nil)
		ctx := context.TODO()

		// WHEN
		_, err := repo.ListByExternalTenants(ctx, []string{id()})

		// THEN
		require.EqualError(t, err, "Internal Server Error: unable to fetch database from context")
	})
}

func TestPgRepository_ListByParentAndType(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN

		tntModels := []*model.BusinessTenantMapping{
			newModelBusinessTenantMappingWithParentAndType(testID, "name1", nil, nil, tenantEntity.Account), nil,
		}
		resultTntModels := []*model.BusinessTenantMapping{
			newModelBusinessTenantMappingWithParentAndType(testID, "name1", []string{testParentID, testParentID2}, nil, tenantEntity.Account), nil,
		}
		tntEntity := newEntityBusinessTenantMappingWithParentAndAccount(testID, "name1", tenantEntity.Account)
		tntEntity.Initialized = boolToPtr(true)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", tntEntity).Return(tntModels[0]).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		tenantByParentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: testID, parentID: testParentID},
		})
		rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
			{sqlRow: sqlRow{id: testID, name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
		})

		parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: testID, parentID: testParentID},
			{tenantID: testID, parentID: testParentID2},
		})

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
			WithArgs(testParentID).
			WillReturnRows(tenantByParentRowsToReturn)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id IN ($1) AND type = $2`)).
			WithArgs(testID, tenantEntity.Account).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs(testParentID).
			WillReturnRows(parentRowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListByParentAndType(ctx, testParentID, tenantEntity.Account)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, resultTntModels, result)
	})

	t.Run("Error when listing", func(t *testing.T) {
		// GIVEN
		parentID := "test"

		tntEntity := newEntityBusinessTenantMappingWithParentAndAccount("id1", "name1", tenantEntity.Account)
		tntEntity.Initialized = boolToPtr(true)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE parent = $1 AND type = $2`)).
			WithArgs(parentID, tenantEntity.Account).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListByParentAndType(ctx, parentID, tenantEntity.Account)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, result)
	})

	t.Run("Error when missing persistence context", func(t *testing.T) {
		// GIVEN
		parentID := "test"
		repo := tenant.NewRepository(nil)
		ctx := context.TODO()

		// WHEN
		result, err := repo.ListByParentAndType(ctx, parentID, tenantEntity.Account)

		// THEN
		require.EqualError(t, err, "Internal Server Error: unable to fetch database from context")
		assert.Nil(t, result)
	})
}

func TestPgRepository_ListByType(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		//parentID := "test"

		resultTntModel := []*model.BusinessTenantMapping{
			//newModelBusinessTenantMappingWithParentAndType("id1", "name1", parentID, nil, tenantEntity.Account), nil,
		}

		tntEntity := newEntityBusinessTenantMappingWithParentAndAccount("id1", "name1", tenantEntity.Account)
		tntEntity.Initialized = boolToPtr(true)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", tntEntity).Return(resultTntModel[0]).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
			{sqlRow: sqlRow{id: "id1", name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE type = $1`)).
			WithArgs(tenantEntity.Account).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListByType(ctx, tenantEntity.Account)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, resultTntModel, result)
	})

	t.Run("Error when listing", func(t *testing.T) {
		// GIVEN

		tntEntity := newEntityBusinessTenantMappingWithParentAndAccount("id1", "name1", tenantEntity.Account)
		tntEntity.Initialized = boolToPtr(true)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE type = $1`)).
			WithArgs(tenantEntity.Account).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListByType(ctx, tenantEntity.Account)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, result)
	})

	t.Run("Error when missing persistence context", func(t *testing.T) {
		// GIVEN
		repo := tenant.NewRepository(nil)
		ctx := context.TODO()

		// WHEN
		result, err := repo.ListByType(ctx, tenantEntity.Account)

		// THEN
		require.EqualError(t, err, "Internal Server Error: unable to fetch database from context")
		assert.Nil(t, result)
	})
}

func TestPgRepository_ListBySubscribedRuntimesAndApplicationTemplates(t *testing.T) {
	//parentID := "test"

	tntEntity := newEntityBusinessTenantMappingWithParentAndAccount("id1", "name1", tenantEntity.Account)
	tntEntity.Initialized = boolToPtr(true)

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		resultTntModel := []*model.BusinessTenantMapping{
			//newModelBusinessTenantMappingWithParentAndType("id1", "name1", parentID, nil, tenantEntity.Account), nil,
		}

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", tntEntity).Return(resultTntModel[0]).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
			{sqlRow: sqlRow{id: "id1", name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE (type = $1 AND (id IN (SELECT tenant_id FROM tenant_runtime_contexts ) OR id IN (SELECT tenant_id FROM tenant_applications WHERE id IN (SELECT id FROM applications WHERE app_template_id IN (SELECT app_template_id FROM labels WHERE key = $2 AND app_template_id IS NOT NULL)))`)).
			WithArgs(tenantEntity.Subaccount, selfRegDistinguishLabel).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListBySubscribedRuntimesAndApplicationTemplates(ctx, selfRegDistinguishLabel)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, resultTntModel, result)
	})

	t.Run("Error when listing", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE (type = $1 AND (id IN (SELECT tenant_id FROM tenant_runtime_contexts ) OR id IN (SELECT tenant_id FROM tenant_applications WHERE id IN (SELECT id FROM applications WHERE app_template_id IN (SELECT app_template_id FROM labels WHERE key = $2 AND app_template_id IS NOT NULL)))`)).
			WithArgs(tenantEntity.Subaccount, selfRegDistinguishLabel).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListBySubscribedRuntimesAndApplicationTemplates(ctx, selfRegDistinguishLabel)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, result)
	})

	t.Run("Error when missing persistence context", func(t *testing.T) {
		// GIVEN
		repo := tenant.NewRepository(nil)
		ctx := context.TODO()

		// WHEN
		result, err := repo.ListBySubscribedRuntimesAndApplicationTemplates(ctx, selfRegDistinguishLabel)

		// THEN
		require.EqualError(t, err, "Internal Server Error: unable to fetch database from context")
		assert.Nil(t, result)
	})
}

func buildQueryWithTenantIDs(ids []string) (string, []driver.Value) {
	argumentValues := make([]driver.Value, 0)
	var sb strings.Builder
	for i, id := range ids {
		argumentValues = append(argumentValues, id)
		sb.WriteString(fmt.Sprintf("$%d", i+1))
		if i < len(ids)-1 {
			sb.WriteString(", ")
		}
	}

	queryFormat := `SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN (%s)`
	query := fmt.Sprintf(queryFormat, sb.String())

	return query, argumentValues
}

func chunkSizedTenantIDs(chunkSize int) []string {
	ids := make([]string, chunkSize)
	for i := 0; i < chunkSize; i++ {
		ids[i] = id()
	}
	return ids
}

func TestPgRepository_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParentID}, nil, tenantEntity.Account)
		tenantMappingEntity := newEntityBusinessTenantMappingWithParent(testID, testName)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()

		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
			WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingRepo.Update(ctx, tenantMappingModel)

		// THEN
		require.NoError(t, err)
		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when getting", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParentID}, nil, tenantEntity.Account)

		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenantEntity.Inactive).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		err := tenantMappingRepo.Update(ctx, tenantMappingModel)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when updating", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParentID}, nil, tenantEntity.Account)
		tenantMappingEntity := newEntityBusinessTenantMappingWithParent(testID, testName)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()

		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
			WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingRepo.Update(ctx, tenantMappingModel)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Success when parent is updated", func(t *testing.T) {
		// GIVEN
		oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParentID}, nil, tenantEntity.Account)
		newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParentID2}, nil, tenantEntity.Account)
		oldTenantMappingEntity := newEntityBusinessTenantMappingWithParent(testID, testName)
		newTenantMappingEntity := newEntityBusinessTenantMappingWithParent(testID, testName)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
		mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()

		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
			WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		for topLvlEntity := range resource.TopLevelEntities {
			if _, ok := topLvlEntity.IgnoredTenantAccessTable(); ok {
				continue
			}
			tenantAccesses := fixTenantAccesses()

			dbMock.ExpectQuery(`SELECT tenant_id, id, owner FROM (.+) WHERE tenant_id = \$1 AND owner = \$2`).
				WithArgs(testID, true).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

			dbMock.ExpectExec(`WITH RECURSIVE parents AS \(SELECT t1\.id, t1\.parent FROM business_tenant_mappings t1 WHERE id = \? UNION ALL SELECT t2\.id, t2\.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2\.id = t\.parent\) INSERT INTO (.+) \( tenant_id, id, owner \) \(SELECT parents\.id AS tenant_id, \? as id, \? AS owner FROM parents\)`).
				WithArgs(testParentID2, tenantAccesses[0].ResourceID, true).WillReturnResult(sqlmock.NewResult(-1, 1))

			dbMock.ExpectExec(`WITH RECURSIVE parents AS \(SELECT t1\.id, t1\.parent FROM business_tenant_mappings t1 WHERE id = \$1 UNION ALL SELECT t2\.id, t2\.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2\.id = t\.parent\) DELETE FROM (.+) WHERE id IN \(\$2\) AND tenant_id IN \(SELECT id FROM parents\)`).
				WithArgs(testParentID, tenantAccesses[0].ResourceID).WillReturnResult(sqlmock.NewResult(-1, 1))
		}

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingRepo.Update(ctx, newTenantMappingModel)

		// THEN
		require.NoError(t, err)
		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when parent is updated and list tenant accesses fail", func(t *testing.T) {
		// GIVEN
		oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParentID}, nil, tenantEntity.Account)
		newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParentID2}, nil, tenantEntity.Account)
		oldTenantMappingEntity := newEntityBusinessTenantMappingWithParent(testID, testName)
		newTenantMappingEntity := newEntityBusinessTenantMappingWithParent(testID, testName)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
		mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()

		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
			WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		dbMock.ExpectQuery(`SELECT tenant_id, id, owner FROM (.+) WHERE tenant_id = \$1 AND owner = \$2`).
			WithArgs(testID, true).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingRepo.Update(ctx, newTenantMappingModel)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when parent is updated and create tenant access fail", func(t *testing.T) {
		// GIVEN
		oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParentID}, nil, tenantEntity.Account)
		newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParentID2}, nil, tenantEntity.Account)
		oldTenantMappingEntity := newEntityBusinessTenantMappingWithParent(testID, testName)
		newTenantMappingEntity := newEntityBusinessTenantMappingWithParent(testID, testName)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
		mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()

		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
			WithArgs(testName, testExternal, testParentID2, "account", "Compass", tenantEntity.Active, testID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		appTenantAccesses := fixTenantAccesses()
		dbMock.ExpectQuery(`SELECT tenant_id, id, owner FROM (.+) WHERE tenant_id = \$1 AND owner = \$2`).
			WithArgs(testID, true).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

		dbMock.ExpectExec(`WITH RECURSIVE parents AS \(SELECT t1\.id, t1\.parent FROM business_tenant_mappings t1 WHERE id = \? UNION ALL SELECT t2\.id, t2\.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2\.id = t\.parent\) INSERT INTO (.+) \( tenant_id, id, owner \) \(SELECT parents\.id AS tenant_id, \? as id, \? AS owner FROM parents\)`).
			WithArgs(testParentID2, appTenantAccesses[0].ResourceID, true).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingRepo.Update(ctx, newTenantMappingModel)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when parent is updated and tenant access delete fail", func(t *testing.T) {
		// GIVEN
		oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParentID}, nil, tenantEntity.Account)
		newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParentID2}, nil, tenantEntity.Account)
		oldTenantMappingEntity := newEntityBusinessTenantMappingWithParent(testID, testName)
		newTenantMappingEntity := newEntityBusinessTenantMappingWithParent(testID, testName)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
		mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()

		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
			WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		appTenantAccesses := fixTenantAccesses()
		dbMock.ExpectQuery(`SELECT tenant_id, id, owner FROM (.+) WHERE tenant_id = \$1 AND owner = \$2`).
			WithArgs(testID, true).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

		dbMock.ExpectExec(`WITH RECURSIVE parents AS \(SELECT t1\.id, t1\.parent FROM business_tenant_mappings t1 WHERE id = \? UNION ALL SELECT t2\.id, t2\.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2\.id = t\.parent\) INSERT INTO (.+) \( tenant_id, id, owner \) \(SELECT parents\.id AS tenant_id, \? as id, \? AS owner FROM parents\)`).
			WithArgs(testParentID2, appTenantAccesses[0].ResourceID, true).WillReturnResult(sqlmock.NewResult(-1, 1))

		dbMock.ExpectExec(`WITH RECURSIVE parents AS \(SELECT t1\.id, t1\.parent FROM business_tenant_mappings t1 WHERE id = \$1 UNION ALL SELECT t2\.id, t2\.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2\.id = t\.parent\) DELETE FROM (.+) WHERE id IN \(\$2\) AND tenant_id IN \(SELECT id FROM parents\)`).
			WithArgs(testParentID, appTenantAccesses[0].ResourceID).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingRepo.Update(ctx, newTenantMappingModel)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})
}

func TestPgRepository_DeleteByExternalTenant(t *testing.T) {
	deleteStatement := regexp.QuoteMeta(`DELETE FROM public.business_tenant_mappings WHERE external_tenant = $1`)

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()

		db, dbMock := testdb.MockDatabase(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
			WithArgs(testExternal, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		for topLvlEntity := range resource.TopLevelEntities {
			if _, ok := topLvlEntity.IgnoredTenantAccessTable(); ok {
				continue
			}
			tenantAccesses := fixTenantAccesses()

			dbMock.ExpectQuery(`SELECT tenant_id, id, owner FROM (.+) WHERE tenant_id = \$1 AND owner = \$2`).
				WithArgs(testID, true).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

			dbMock.ExpectExec(`DELETE FROM (.+) WHERE id IN \(\$1\)`).
				WithArgs(tenantAccesses[0].ResourceID).WillReturnResult(sqlmock.NewResult(-1, 1))
		}

		dbMock.ExpectExec(deleteStatement).
			WithArgs(testExternal).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := tenant.NewRepository(mockConverter)

		// WHEN
		err := repo.DeleteByExternalTenant(ctx, testExternal)

		// THEN
		require.NoError(t, err)
		dbMock.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
	})

	t.Run("Success when getting tenant before delete returns not found", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
			WithArgs(testExternal, tenantEntity.Inactive).
			WillReturnError(sql.ErrNoRows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := tenant.NewRepository(nil)

		// WHEN
		err := repo.DeleteByExternalTenant(ctx, testExternal)

		// THEN
		require.NoError(t, err)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when getting tenant before delete fail", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
			WithArgs(testExternal, tenantEntity.Inactive).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := tenant.NewRepository(nil)

		// WHEN
		err := repo.DeleteByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when List tenant access fail", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()

		db, dbMock := testdb.MockDatabase(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
			WithArgs(testExternal, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectQuery(`SELECT tenant_id, id, owner FROM (.+) WHERE tenant_id = \$1 AND owner = \$2`).
			WithArgs(testID, true).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := tenant.NewRepository(mockConverter)

		// WHEN
		err := repo.DeleteByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		dbMock.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
	})

	t.Run("Error when Delete tenant access fail", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()

		db, dbMock := testdb.MockDatabase(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
			WithArgs(testExternal, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		appTenantAccesses := fixTenantAccesses()
		dbMock.ExpectQuery(`SELECT tenant_id, id, owner FROM (.+) WHERE tenant_id = \$1 AND owner = \$2`).
			WithArgs(testID, true).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

		dbMock.ExpectExec(`DELETE FROM (.+) WHERE id IN \(\$1\)`).
			WithArgs(appTenantAccesses[0].ResourceID).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := tenant.NewRepository(mockConverter)

		// WHEN
		err := repo.DeleteByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		dbMock.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
	})

	t.Run("Error when delete fails", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()

		db, dbMock := testdb.MockDatabase(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
			WithArgs(testExternal, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		for topLvlEntity := range resource.TopLevelEntities {
			if _, ok := topLvlEntity.IgnoredTenantAccessTable(); ok {
				continue
			}
			tenantAccesses := fixTenantAccesses()

			dbMock.ExpectQuery(`SELECT tenant_id, id, owner FROM (.+) WHERE tenant_id = \$1 AND owner = \$2`).
				WithArgs(testID, true).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

			dbMock.ExpectExec(`DELETE FROM (.+) WHERE id IN \(\$1\)`).
				WithArgs(tenantAccesses[0].ResourceID).WillReturnResult(sqlmock.NewResult(-1, 1))
		}

		dbMock.ExpectExec(deleteStatement).
			WithArgs(testExternal).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := tenant.NewRepository(mockConverter)

		// WHEN
		err := repo.DeleteByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		dbMock.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
	})
}

func TestPgRepository_GetLowestOwnerForResource(t *testing.T) {
	runtimeID := "runtimeID"

	t.Run("Success", func(t *testing.T) {
		db, dbMock := mockDBSuccess(t, runtimeID)
		defer dbMock.AssertExpectations(t)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.GetLowestOwnerForResource(ctx, resource.Runtime, runtimeID)

		// THEN
		require.NoError(t, err)
		require.Equal(t, testID, result)
	})

	t.Run("Error when getting", func(t *testing.T) {
		db, dbMock := mockDBError(t, runtimeID)
		defer dbMock.AssertExpectations(t)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.GetLowestOwnerForResource(ctx, resource.Runtime, runtimeID)

		// THEN
		require.Error(t, err)
		require.Empty(t, result)
	})
}

func TestPgRepository_GetCustomerIDParentRecursively(t *testing.T) {
	dbQuery := `WITH RECURSIVE parents AS
                   (SELECT t1.id, t1.parent, t1.external_tenant, t1.type
                    FROM business_tenant_mappings t1
                    WHERE id = $1
                    UNION ALL
                    SELECT t2.id, t2.parent, t2.external_tenant, t2.type
                    FROM business_tenant_mappings t2
                             INNER JOIN parents p on p.parent = t2.id)
			SELECT external_tenant, type FROM parents WHERE parent is null`

	t.Run("Success when parent and type are returned", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := sqlmock.NewRows([]string{"external_tenant", "type"}).AddRow(testParentID, tenantEntity.TypeToStr(tenantEntity.Customer))
		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(testID).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		customerID, err := tenantMappingRepo.GetCustomerIDParentRecursively(ctx, testID)

		// THEN
		require.NoError(t, err)
		require.Equal(t, customerID, testParentID)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when executing db query", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(testID).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		customerID, err := tenantMappingRepo.GetCustomerIDParentRecursively(ctx, testID)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		require.Empty(t, customerID)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error if missing persistence context", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		tenantMappingRepo := tenant.NewRepository(nil)
		// WHEN
		_, err := tenantMappingRepo.GetCustomerIDParentRecursively(ctx, testID)
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("Return empty string when returned type is not customer", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := sqlmock.NewRows([]string{"external_tenant", "type"}).AddRow(testParentID, tenantEntity.TypeToStr(tenantEntity.Account))
		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(testID).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		customerID, err := tenantMappingRepo.GetCustomerIDParentRecursively(ctx, testID)

		// THEN
		require.NoError(t, err)
		require.Empty(t, customerID)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when empty parent is returned", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := sqlmock.NewRows([]string{"external_tenant", "type"}).AddRow("", tenantEntity.TypeToStr(tenantEntity.Customer))
		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(testID).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		customerID, err := tenantMappingRepo.GetCustomerIDParentRecursively(ctx, testID)

		// THEN
		expectedError := fmt.Sprintf("external parent customer ID for internal tenant ID: %s can not be empty", testID)
		require.Error(t, err)
		require.EqualError(t, err, expectedError)
		require.Empty(t, customerID)
		dbMock.AssertExpectations(t)
	})
}

func TestPgRepository_GetParentRecursivelyByExternalTenant(t *testing.T) {
	dbQuery := `WITH RECURSIVE parents AS
                   (SELECT t1.id, t1.external_name, t1.external_tenant, t1.provider_name, t1.status, t1.parent, t1.type
                    FROM business_tenant_mappings t1
                    WHERE external_tenant = $1
                    UNION ALL
                    SELECT t2.id, t2.external_name, t2.external_tenant, t2.provider_name, t2.status, t2.parent, t2.type
                    FROM business_tenant_mappings t2
                             INNER JOIN parents p on p.parent = t2.id)
			SELECT id, external_name, external_tenant, provider_name, status, parent, type FROM parents WHERE parent is null`

	tenantMappingModel := &model.BusinessTenantMapping{
		ID:             testID,
		Name:           testName,
		ExternalTenant: testExternal,
		Parents:        []string{},
		Type:           tenantEntity.Account,
		Provider:       testProvider,
		Status:         tenantEntity.Active,
	}
	tenantMappingEntity := &tenantEntity.Entity{
		ID:             testID,
		Name:           testName,
		ExternalTenant: testExternal,
		Type:           tenantEntity.Account,
		ProviderName:   testProvider,
		Status:         tenantEntity.Active,
	}

	t.Run("Success when parent and type are returned", func(t *testing.T) {
		// GIVEN

		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := sqlmock.NewRows([]string{"id", "external_name", "external_tenant", "provider_name", "status", "parent", "type"}).AddRow(testID, testName, testExternal, testProvider, tenantEntity.Active, "", tenantEntity.TypeToStr(tenantEntity.Account))
		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(testExternal).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		parentTenant, err := tenantMappingRepo.GetParentsRecursivelyByExternalTenant(ctx, testExternal)

		// THEN
		require.NoError(t, err)
		require.Equal(t, parentTenant, tenantMappingModel)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when executing db query", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(testExternal).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		parentTenant, err := tenantMappingRepo.GetParentsRecursivelyByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		require.Empty(t, parentTenant)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error if missing persistence context", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		tenantMappingRepo := tenant.NewRepository(nil)
		// WHEN
		_, err := tenantMappingRepo.GetParentsRecursivelyByExternalTenant(ctx, testExternal)
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})
}

const selectTenantsQuery = `(SELECT tenant_id FROM tenant_runtimes ta WHERE ta.id = $1 AND ta.owner = true AND (NOT EXISTS(SELECT 1 FROM public.business_tenant_mappings WHERE parent = ta.tenant_id) OR (NOT EXISTS(SELECT 1 FROM tenant_runtimes ta2 WHERE ta2.id = $2 AND ta2.owner = true AND ta2.tenant_id IN (SELECT id FROM public.business_tenant_mappings WHERE parent = ta.tenant_id)))))`

func mockDBSuccess(t *testing.T, runtimeID string) (*sqlx.DB, testdb.DBMock) {
	db, dbMock := testdb.MockDatabase(t)
	rowsToReturn := sqlmock.NewRows([]string{"tenant_id"}).AddRow(testID)
	dbMock.ExpectQuery(regexp.QuoteMeta(selectTenantsQuery)).
		WithArgs(runtimeID, runtimeID).
		WillReturnRows(rowsToReturn)
	return db, dbMock
}

func mockDBError(t *testing.T, runtimeID string) (*sqlx.DB, testdb.DBMock) {
	db, dbMock := testdb.MockDatabase(t)
	dbMock.ExpectQuery(regexp.QuoteMeta(selectTenantsQuery)).
		WithArgs(runtimeID, runtimeID).WillReturnError(testError)
	return db, dbMock
}

func id() string {
	return uuid.New().String()
}
