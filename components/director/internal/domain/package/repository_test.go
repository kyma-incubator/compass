package mp_package_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/domain/package/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	pkgModel := fixPackageModel()
	pkgEntity := fixEntityPackage()
	insertQuery := `^INSERT INTO public.packages \(.+\) VALUES \(.+\)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixPackageRow()...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.EntityConverter{}
		convMock.On("ToEntity", pkgModel).Return(pkgEntity, nil).Once()
		pgRepository := mp_package.NewRepository(&convMock)
		//WHEN
		err := pgRepository.Create(ctx, pkgModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.EntityConverter{}
		pgRepository := mp_package.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "model can not be nil")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE public.packages SET vendor = ?, title = ?, short_description = ?, description = ?, version = ?, package_links = ?, links = ?,
		licence_type = ?, tags = ?, countries = ?, labels = ?, policy_level = ?, custom_policy_level = ?, part_of_products = ?, line_of_business = ?, industry = ? WHERE tenant_id = ? AND id = ?`)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		pkg := fixPackageModel()
		entity := fixEntityPackage()

		convMock := &automock.EntityConverter{}
		convMock.On("ToEntity", pkg).Return(entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(append(fixPackageUpdateArgs(), tenantID, entity.ID)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := mp_package.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, pkg)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when model is nil", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		pgRepository := mp_package.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, nil)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "model can not be nil")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Delete(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := `^DELETE FROM public.packages WHERE tenant_id = \$1 AND id = \$2$`

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, packageID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.EntityConverter{}
	pgRepository := mp_package.NewRepository(convMock)
	//WHEN
	err := pgRepository.Delete(ctx, tenantID, packageID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_Exists(t *testing.T) {
	//GIVEN
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	existQuery := regexp.QuoteMeta(`SELECT 1 FROM public.packages WHERE tenant_id = $1 AND id = $2`)

	sqlMock.ExpectQuery(existQuery).WithArgs(tenantID, packageID).WillReturnRows(testdb.RowWhenObjectExist())
	convMock := &automock.EntityConverter{}
	pgRepository := mp_package.NewRepository(convMock)
	//WHEN
	found, err := pgRepository.Exists(ctx, tenantID, packageID)
	//THEN
	require.NoError(t, err)
	assert.True(t, found)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	// given
	pkgEntity := fixEntityPackage()

	selectQuery := `^SELECT (.+) FROM public.packages WHERE tenant_id = \$1 AND id = \$2$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixPackageColumns()).
			AddRow(fixPackageRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, packageID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", pkgEntity).Return(&model.Package{ID: packageID, TenantID: tenantID}, nil).Once()
		pgRepository := mp_package.NewRepository(convMock)
		// WHEN
		modelBndl, err := pgRepository.GetByID(ctx, tenantID, packageID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, packageID, modelBndl.ID)
		assert.Equal(t, tenantID, modelBndl.TenantID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := mp_package.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, packageID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelBndl, err := repo.GetByID(ctx, tenantID, packageID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelBndl)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")
		rows := sqlmock.NewRows(fixPackageColumns()).
			AddRow(fixPackageRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, packageID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", pkgEntity).Return(&model.Package{}, testError).Once()
		pgRepository := mp_package.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetByID(ctx, tenantID, packageID)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}
