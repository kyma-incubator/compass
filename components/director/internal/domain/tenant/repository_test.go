package tenant_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := fixModelTenantMapping(testID, testName)
		tenantMappingEntity := fixEntityTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.tenant_mapping ( id, name, external_tenant, internal_tenant, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingrepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingrepo.Create(ctx, *tenantMappingModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when creating", func(t *testing.T) {
		// GIVEN
		intSysModel := fixModelTenantMapping(testID, testName)
		intSysEntity := fixEntityTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", intSysModel).Return(intSysEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.tenant_mapping ( id, name, external_tenant, internal_tenant, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixTenantMappingCreateArgs(*intSysEntity)...).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingRepo.Create(ctx, *intSysModel)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})
}

func TestPgRepository_CreateMany(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModels := []model.TenantMapping{*fixModelTenantMapping("id1", "name1"),
			*fixModelTenantMapping("id2", "name2"),
			*fixModelTenantMapping("id3", "name3")}
		tenantMappingEntities := []*tenant.Entity{fixEntityTenantMapping("id1", "name1"),
			fixEntityTenantMapping("id2", "name2"),
			fixEntityTenantMapping("id3", "name3")}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", &tenantMappingModels[0]).Return(tenantMappingEntities[0]).Once()
		mockConverter.On("ToEntity", &tenantMappingModels[1]).Return(tenantMappingEntities[1]).Once()
		mockConverter.On("ToEntity", &tenantMappingModels[2]).Return(tenantMappingEntities[2]).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.tenant_mapping ( id, name, external_tenant, internal_tenant, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntities[0])...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.tenant_mapping ( id, name, external_tenant, internal_tenant, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntities[1])...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.tenant_mapping ( id, name, external_tenant, internal_tenant, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntities[2])...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingrepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingrepo.CreateMany(ctx, tenantMappingModels)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when creating", func(t *testing.T) {
		// GIVEN
		tenantMappingModels := []model.TenantMapping{*fixModelTenantMapping("id1", "name1"),
			*fixModelTenantMapping("id2", "name2"),
			*fixModelTenantMapping("id3", "name3")}
		tenantMappingEntity := fixEntityTenantMapping("id1", "name")

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", &tenantMappingModels[0]).Return(tenantMappingEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.tenant_mapping ( id, name, external_tenant, internal_tenant, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		err := tenantMappingRepo.CreateMany(ctx, tenantMappingModels)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})
}

func TestPgRepository_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := fixModelTenantMapping(testID, testName)
		tenantMappingEntity := fixEntityTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, internalTenant: testInternal, provider: "Compass", status: tenant.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, external_tenant, internal_tenant, provider_name, status FROM public.tenant_mapping WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenant.Inactive).
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

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := fixModelTenantMapping(testID, testName)
		tenantMappingEntity := fixEntityTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, internalTenant: testInternal, provider: "Compass", status: tenant.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, external_tenant, internal_tenant, provider_name, status FROM public.tenant_mapping WHERE id = $1 AND status != $2 `)).
			WithArgs(testID, tenant.Inactive).
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
		tenantMappingModel := fixModelTenantMapping(testID, testName)
		tenantMappingEntity := fixEntityTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, internalTenant: testInternal, provider: "Compass", status: tenant.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, external_tenant, internal_tenant, provider_name, status FROM public.tenant_mapping WHERE external_tenant = $1 AND status != $2 `)).
			WithArgs(testExternal, tenant.Inactive).
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
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, external_tenant, internal_tenant, provider_name, status FROM public.tenant_mapping WHERE external_tenant = $1 AND status != $2 `)).
			WithArgs(testExternal, tenant.Inactive).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.GetByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		require.Nil(t, result)
	})
}

func TestPgRepository_GetByInternalTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		tenantMappingModel := fixModelTenantMapping(testID, testName)
		tenantMappingEntity := fixEntityTenantMapping(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, externalTenant: testExternal, internalTenant: testInternal, provider: "Compass", status: tenant.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, external_tenant, internal_tenant, provider_name, status FROM public.tenant_mapping WHERE internal_tenant = $1 AND status != $2 `)).
			WithArgs(testInternal, tenant.Inactive).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.GetByInternalTenant(ctx, testInternal)

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
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, external_tenant, internal_tenant, provider_name, status FROM public.tenant_mapping WHERE internal_tenant = $1 AND status != $2 `)).
			WithArgs(testInternal, tenant.Inactive).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.GetByInternalTenant(ctx, testInternal)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		require.Nil(t, result)
	})
}

func TestPgRepository_Exists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.tenant_mapping WHERE id = $1`)).
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
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.tenant_mapping WHERE id = $1`)).
			WithArgs(testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.Exists(ctx, testID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		assert.False(t, result)
	})
}

func TestPgRepository_ExistsByInternalTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.tenant_mapping WHERE internal_tenant = $1`)).
			WithArgs(testInternal).
			WillReturnRows(testdb.RowWhenObjectExist())

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.ExistsByInternalTenant(ctx, testInternal)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result)
	})

	t.Run("Error when checking existence", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.tenant_mapping WHERE internal_tenant = $1`)).
			WithArgs(testInternal).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.ExistsByInternalTenant(ctx, testInternal)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		assert.False(t, result)
	})
}

func TestPgRepository_ExistsByExternalTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.tenant_mapping WHERE external_tenant = $1`)).
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
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.tenant_mapping WHERE external_tenant = $1`)).
			WithArgs(testExternal).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.ExistsByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		assert.False(t, result)
	})
}

func TestPgRepository_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		intSysModels := []*model.TenantMapping{
			fixModelTenantMapping("id1", "name1"),
			fixModelTenantMapping("id2", "name2"),
			fixModelTenantMapping("id3", "name3"),
		}

		intSysEntities := []*tenant.Entity{
			fixEntityTenantMapping("id1", "name1"),
			fixEntityTenantMapping("id2", "name2"),
			fixEntityTenantMapping("id3", "name3"),
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", intSysEntities[0]).Return(intSysModels[0]).Once()
		mockConverter.On("FromEntity", intSysEntities[1]).Return(intSysModels[1]).Once()
		mockConverter.On("FromEntity", intSysEntities[2]).Return(intSysModels[2]).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: "id1", name: "name1", externalTenant: testExternal, internalTenant: testInternal, provider: "Compass", status: tenant.Active},
			{id: "id2", name: "name2", externalTenant: testExternal, internalTenant: testInternal, provider: "Compass", status: tenant.Active},
			{id: "id3", name: "name3", externalTenant: testExternal, internalTenant: testInternal, provider: "Compass", status: tenant.Active},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, external_tenant, internal_tenant, provider_name, status FROM public.tenant_mapping ORDER BY id LIMIT 3 OFFSET 0`)).
			WillReturnRows(rowsToReturn)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM public.tenant_mapping`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.List(ctx, testPageSize, testCursor)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, intSysModels, result.Data)
	})

	t.Run("Error when listing", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, external_tenant, internal_tenant, provider_name, status FROM public.tenant_mapping ORDER BY id LIMIT 3 OFFSET 0`)).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.List(ctx, testPageSize, testCursor)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		require.Nil(t, result.Data)
	})
}
