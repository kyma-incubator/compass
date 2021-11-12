package ordpackage_test

import (
	"database/sql/driver"
	"fmt"
	ordpackage "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/package/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestPgRepository_Create(t *testing.T) {
	suite := testdb.RepoCreateTestSuite{
		Name: "Create Package",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, appID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO public.packages \(.+\) VALUES \(.+\)$`,
				Args:        fixPackageRow(),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       ordpackage.NewRepository,
		ModelEntity:               fixPackageModel(),
		DBEntity:                  fixEntityPackage(),
		NilModelEntity:            fixNilModelPackage(),
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	entity := fixEntityPackage()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Package",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query: regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.packages SET vendor = ?, title = ?, short_description = ?, description = ?, version = ?, package_links = ?, links = ?,
		licence_type = ?, tags = ?, countries = ?, labels = ?, policy_level = ?, custom_policy_level = ?, part_of_products = ?, line_of_business = ?, industry = ?, resource_hash = ? WHERE id = ? AND (id IN (SELECT id FROM packages_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID)),
				Args:          append(fixPackageUpdateArgs(), entity.ID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       ordpackage.NewRepository,
		ModelEntity:               fixPackageModel(),
		DBEntity:                  entity,
		NilModelEntity:            fixNilModelPackage(),
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

/*
func TestPgRepository_Delete(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := fmt.Sprintf(`^DELETE FROM public.packages WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, packageID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.EntityConverter{}
	pgRepository := ordpackage.NewRepository(convMock)
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
	existQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT 1 FROM public.packages WHERE %s AND id = $2`, fixUnescapedTenantIsolationSubquery()))

	sqlMock.ExpectQuery(existQuery).WithArgs(tenantID, packageID).WillReturnRows(testdb.RowWhenObjectExist())
	convMock := &automock.EntityConverter{}
	pgRepository := ordpackage.NewRepository(convMock)
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

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.packages WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

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
		pgRepository := ordpackage.NewRepository(convMock)
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
		repo := ordpackage.NewRepository(nil)
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
		pgRepository := ordpackage.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetByID(ctx, tenantID, packageID)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_ListByApplicationID(t *testing.T) {
	// GIVEN
	totalCount := 2
	firstPkgEntity := fixEntityPackage()
	secondPkgEntity := fixEntityPackage()

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.packages WHERE %s AND app_id = \$2`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixPackageColumns()).
			AddRow(fixPackageRow()...).
			AddRow(fixPackageRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstPkgEntity).Return(&model.Package{ID: firstPkgEntity.ID}, nil)
		convMock.On("FromEntity", secondPkgEntity).Return(&model.Package{ID: secondPkgEntity.ID}, nil)
		pgRepository := ordpackage.NewRepository(convMock)
		// WHEN
		modelPkg, err := pgRepository.ListByApplicationID(ctx, tenantID, appID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelPkg, totalCount)
		assert.Equal(t, firstPkgEntity.ID, modelPkg[0].ID)
		assert.Equal(t, secondPkgEntity.ID, modelPkg[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}*/
