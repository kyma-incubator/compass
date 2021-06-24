package tenant_test

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/testdb"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/tenant"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/tenant/automock"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	tenantModel := model.TenantModel{
		ID:               testID,
		Name:             testID,
		TenantId:         testID,
		ParentExternalId: testID,
		ParentInternalId: testID,
		Type:             tenantEntity.Account,
		Provider:         testProviderName,
		Status:           tenantEntity.Active,
	}

	entity := tenantEntity.Entity{
		ID:             testID,
		Name:           testID,
		ExternalTenant: testID,
		Parent: sql.NullString{
			String: testID,
			Valid:  true,
		},
		Type:         tenantEntity.Account,
		ProviderName: testProviderName,
		Status:       tenantEntity.Active,
	}

	t.Run("Success", func(t *testing.T) {
		//GIVEN
		db, dbMock := testdb.MockDatabase(t)
		ctx := context.Background()
		ctx = persistence.SaveToContext(ctx, db)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", tenantModel).Return(entity)
		defer mockConverter.AssertExpectations(t)

		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(createQuery)).
			WithArgs(createQueryArgs...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		repo := tenant.NewRepository(mockConverter)

		//WHEN
		err := repo.Create(ctx, tenantModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when there is no db instance in the context", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()

		mockConverter := &automock.Converter{}
		mockConverter.AssertNotCalled(t, "ToEntity")
		defer mockConverter.AssertExpectations(t)

		repo := tenant.NewRepository(mockConverter)

		//WHEN
		err := repo.Create(ctx, tenantModel)

		// THEN
		require.Error(t, err)
	})
}

func TestRepository_GetByExternalID(t *testing.T) {
	tenantModel := model.TenantModel{
		ID:               testID,
		Name:             testID,
		TenantId:         testID,
		ParentExternalId: testID,
		ParentInternalId: testID,
		Type:             tenantEntity.Account,
		Provider:         testProviderName,
		Status:           tenantEntity.Active,
	}

	entity := tenantEntity.Entity{
		ID:             testID,
		Name:           testID,
		ExternalTenant: testID,
		Parent: sql.NullString{
			String: testID,
			Valid:  true,
		},
		Type:         tenantEntity.Account,
		ProviderName: testProviderName,
		Status:       tenantEntity.Active,
	}

	t.Run("Success", func(t *testing.T) {
		//GIVEN
		db, dbMock := testdb.MockDatabase(t)

		ctx := context.Background()
		ctx = persistence.SaveToContext(ctx, db)

		rowsToReturn := fixSQLRows([]sqlRow{
			{id: entity.ID, name: entity.Name, externalTenant: entity.ExternalTenant, parent: entity.Parent.String, tenantType: entity.Type, provider: entity.ProviderName, status: entity.Status},
		})
		dbMock.ExpectQuery(fmt.Sprintf(getByExternalIDQueryFormat, entity.ExternalTenant)).
			WillReturnRows(rowsToReturn)
		defer dbMock.AssertExpectations(t)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", entity).Return(tenantModel)
		defer mockConverter.AssertExpectations(t)

		repo := tenant.NewRepository(mockConverter)

		//WHEN
		tnt, err := repo.GetByExternalID(ctx, tenantModel.ParentExternalId)

		// THEN
		require.NoError(t, err)
		require.Equal(t, tenantModel, tnt)
	})

	t.Run("Error when there is no db instance in the context", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()

		mockConverter := &automock.Converter{}
		mockConverter.AssertNotCalled(t, "ToEntity")
		defer mockConverter.AssertExpectations(t)

		repo := tenant.NewRepository(mockConverter)

		//WHEN
		_, err := repo.GetByExternalID(ctx, tenantModel.ParentExternalId)

		// THEN
		require.Error(t, err)
	})
}

func TestRepository_Update(t *testing.T) {
	tenantModel := model.TenantModel{
		ID:               testID,
		Name:             testID,
		TenantId:         testID,
		ParentExternalId: testID,
		ParentInternalId: testID,
		Type:             tenantEntity.Account,
		Provider:         testProviderName,
		Status:           tenantEntity.Active,
	}

	entity := tenantEntity.Entity{
		ID:             testID,
		Name:           testID,
		ExternalTenant: testID,
		Parent: sql.NullString{
			String: testID,
			Valid:  true,
		},
		Type:         tenantEntity.Account,
		ProviderName: testProviderName,
		Status:       tenantEntity.Active,
	}

	t.Run("Success", func(t *testing.T) {
		//GIVEN
		db, dbMock := testdb.MockDatabase(t)
		ctx := context.Background()
		ctx = persistence.SaveToContext(ctx, db)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", tenantModel).Return(entity)
		defer mockConverter.AssertExpectations(t)

		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(updateQueryFormat, entity.ExternalTenant))).
			WithArgs(entity.Name, entity.Parent, entity.Type, entity.ProviderName, entity.Status).
			WillReturnResult(sqlmock.NewResult(0, 1))

		repo := tenant.NewRepository(mockConverter)

		//WHEN
		err := repo.Update(ctx, tenantModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when there is no db instance in the context", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()

		mockConverter := &automock.Converter{}
		mockConverter.AssertNotCalled(t, "ToEntity")
		defer mockConverter.AssertExpectations(t)

		repo := tenant.NewRepository(mockConverter)

		//WHEN
		err := repo.Update(ctx, tenantModel)

		// THEN
		require.Error(t, err)
	})
}

func TestRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		//GIVEN
		db, dbMock := testdb.MockDatabase(t)
		ctx := context.Background()
		ctx = persistence.SaveToContext(ctx, db)

		mockConverter := &automock.Converter{}

		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(deleteQuery)).
			WithArgs(testID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		repo := tenant.NewRepository(mockConverter)

		//WHEN
		err := repo.DeleteByExternalID(ctx, testID)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when there is no db instance in the context", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		repo := tenant.NewRepository(mockConverter)

		//WHEN
		err := repo.DeleteByExternalID(ctx, testID)

		// THEN
		require.Error(t, err)
	})
}
