package tenant_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

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

func TestPgRepository_Upsert(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, parent, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ?, ? ) ON CONFLICT ( external_tenant ) DO UPDATE SET external_name=EXCLUDED.external_name`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingrepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingrepo.Upsert(ctx, *tenantMappingModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when upserting", func(t *testing.T) {
		// GIVEN
		tenantModel := newModelBusinessTenantMapping(testID, testName)
		tenantEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", tenantModel).Return(tenantEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, parent, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ?, ? ) ON CONFLICT ( external_tenant ) DO UPDATE SET external_name=EXCLUDED.external_name`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantEntity)...).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingRepo.Upsert(ctx, *tenantModel)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_UnsafeCreate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, parent, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ?, ? )  ON CONFLICT ( external_tenant ) DO NOTHING`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingrepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingrepo.UnsafeCreate(ctx, *tenantMappingModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when creating", func(t *testing.T) {
		// GIVEN
		tenantModel := newModelBusinessTenantMapping(testID, testName)
		tenantEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", tenantModel).Return(tenantEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, parent, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ?, ? )  ON CONFLICT ( external_tenant ) DO NOTHING`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantEntity)...).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingRepo.UnsafeCreate(ctx, *tenantModel)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, parent: sql.NullString{}, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, parent, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.Get(ctx, testID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, tenantMappingModel, result)
	})

	t.Run("Error when get", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, parent: sql.NullString{}, typeRow: tenantEntity.TypeToStr(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, parent, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.Get(ctx, testID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, tenantMappingModel, result)
	})
}

func TestPgRepository_GetByExternalTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, parent: sql.NullString{}, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, parent, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
			WithArgs(testExternal, tenantEntity.Inactive).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.GetByExternalTenant(ctx, testExternal)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, tenantMappingModel, result)
	})

	t.Run("Error when getting", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, parent, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $ `)).
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

func TestPgRepository_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN

		initializedVal := true
		notInitializedVal := false

		tenantModels := []*model.BusinessTenantMapping{
			newModelBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal),
			newModelBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal),
			newModelBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal),
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
			{sqlRow: sqlRow{id: "id1", name: "name1", externalTenant: testExternal, parent: sql.NullString{}, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &initializedVal},
			{sqlRow: sqlRow{id: "id2", name: "name2", externalTenant: testExternal, parent: sql.NullString{}, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
			{sqlRow: sqlRow{id: "id3", name: "name3", externalTenant: testExternal, parent: sql.NullString{}, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT t.id, t.external_name, t.external_tenant, t.parent, t.type, t.provider_name, t.status, ld.tenant_id IS NOT NULL AS initialized FROM public.business_tenant_mappings t LEFT JOIN public.label_definitions ld ON t.id=ld.tenant_id WHERE t.status = $1 ORDER BY initialized DESC, t.external_name ASC`)).
			WithArgs(tenantEntity.Active).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.List(ctx)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, tenantModels, result)
	})

	t.Run("Error when listing", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT t.id, t.external_name, t.external_tenant, t.parent, t.type, t.provider_name, t.status, ld.tenant_id IS NOT NULL AS initialized FROM public.business_tenant_mappings t LEFT JOIN public.label_definitions ld ON t.id=ld.tenant_id WHERE t.status = $1 ORDER BY initialized DESC, t.external_name ASC`)).
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

func TestPgRepository_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName).WithStatus(tenantEntity.Inactive)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName).WithStatus(tenantEntity.Inactive)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", &tenantMappingModel).Return(&tenantMappingEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, parent = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
			WithArgs(testName, testExternal, sql.NullString{}, "account", "Compass", tenantEntity.Inactive, testID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingRepo.Update(ctx, &tenantMappingModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when updating", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := newModelBusinessTenantMapping(testID, testName).WithStatus(tenantEntity.Inactive)
		tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName).WithStatus(tenantEntity.Inactive)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", &tenantMappingModel).Return(&tenantMappingEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, parent = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
			WithArgs(testName, testExternal, sql.NullString{}, "account", "Compass", tenantEntity.Inactive, testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingRepo.Update(ctx, &tenantMappingModel)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_DeleteByExternalTenant(t *testing.T) {
	deleteStatement := regexp.QuoteMeta(`DELETE FROM public.business_tenant_mappings WHERE external_tenant = $1`)

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(deleteStatement).
			WithArgs(testExternal).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := tenant.NewRepository(nil)

		// WHEN
		err := repo.DeleteByExternalTenant(ctx, testExternal)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Database error", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(deleteStatement).
			WithArgs(testExternal).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := tenant.NewRepository(nil)

		// WHEN
		err := repo.DeleteByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
