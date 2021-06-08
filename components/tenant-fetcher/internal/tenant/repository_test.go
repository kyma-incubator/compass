package tenant_test

import (
	"context"
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
		TenantId:       testID,
		ID:             testID,
		Status:         tenantEntity.Active,
		Subdomain:      subdomain,
		CustomerId:     customerID,
		TenantProvider: testProviderName,
	}

	entity := tenantEntity.Entity{
		Name:           testID,
		ExternalTenant: testID,
		ID:             testID,
		CustomerId:     customerID,
		Subdomain:      subdomain,
		ProviderName:   testProviderName,
		Status:         tenantEntity.Active,
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
