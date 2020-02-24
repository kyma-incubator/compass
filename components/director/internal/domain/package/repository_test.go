package mp_package_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/stretchr/testify/assert"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/domain/package/automock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	name := "foo"
	desc := "bar"

	pkgModel := fixPackageModel(t, name, desc)
	pkgEntity := fixEntityPackage(packageID, name, desc)
	insertQuery := `^INSERT INTO public.packages \(.+\) VALUES \(.+\)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defAuth, err := json.Marshal(pkgModel.DefaultInstanceAuth)
		require.NoError(t, err)
		schema, err := json.Marshal(pkgModel.InstanceAuthRequestInputSchema)
		require.NoError(t, err)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixPackageCreateArgs(string(defAuth), string(schema), pkgModel)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.EntityConverter{}
		convMock.On("ToEntity", pkgModel).Return(pkgEntity, nil).Once()
		pgRepository := mp_package.NewRepository(&convMock)
		//WHEN
		err = pgRepository.Create(ctx, pkgModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when conversion from model to entity failed", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.EntityConverter{}
		convMock.On("ToEntity", pkgModel).Return(&mp_package.Entity{}, errors.New("test error"))
		pgRepository := mp_package.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, pkgModel)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test error")
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
	updateQuery := regexp.QuoteMeta(`UPDATE public.packages SET name = ?, description = ?, instance_auth_request_json_schema = ?,
		default_instance_auth = ? WHERE tenant_id = ? AND id = ?`)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		pkg := fixPackageModel(t, "foo", "update")
		entity := fixEntityPackage(packageID, "foo", "update")

		convMock := &automock.EntityConverter{}
		convMock.On("ToEntity", pkg).Return(entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(entity.Name, entity.Description, entity.InstanceAuthRequestJSONSchema, entity.DefaultInstanceAuth, tenantID, entity.ID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := mp_package.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, pkg)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when conversion from model to entity failed", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		pkgModel := &model.Package{}
		convMock := &automock.EntityConverter{}
		convMock.On("ToEntity", pkgModel).Return(&mp_package.Entity{}, errors.New("test error")).Once()
		pgRepository := mp_package.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, pkgModel)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test error")
		convMock.AssertExpectations(t)
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
	pkgEntity := fixEntityPackage(packageID, "foo", "bar")

	selectQuery := `^SELECT (.+) FROM public.packages WHERE tenant_id = \$1 AND id = \$2$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixPackageColumns()).
			AddRow(fixPackageRow(packageID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, packageID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", pkgEntity).Return(&model.Package{ID: packageID, TenantID: tenantID}, nil).Once()
		pgRepository := mp_package.NewRepository(convMock)
		// WHEN
		modelPkg, err := pgRepository.GetByID(ctx, tenantID, packageID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, packageID, modelPkg.ID)
		assert.Equal(t, tenantID, modelPkg.TenantID)
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
		modelPkg, err := repo.GetByID(ctx, tenantID, packageID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelPkg)
		require.EqualError(t, err, fmt.Sprintf("while getting object from DB: %s", testError.Error()))
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")
		rows := sqlmock.NewRows(fixPackageColumns()).
			AddRow(fixPackageRow(packageID, "placeholder")...)

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

func TestPgRepository_GetForApplication(t *testing.T) {
	// given
	pkgEntity := fixEntityPackage(packageID, "foo", "bar")

	selectQuery := `^SELECT (.+) FROM public.packages WHERE tenant_id = \$1 AND id = \$2 AND app_id = \$3`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixPackageColumns()).
			AddRow(fixPackageRow(packageID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, packageID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", pkgEntity).Return(&model.Package{ID: packageID, TenantID: tenantID, ApplicationID: appID}, nil).Once()
		pgRepository := mp_package.NewRepository(convMock)
		// WHEN
		modelPkg, err := pgRepository.GetForApplication(ctx, tenantID, packageID, appID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, packageID, modelPkg.ID)
		assert.Equal(t, tenantID, modelPkg.TenantID)
		assert.Equal(t, appID, modelPkg.ApplicationID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := mp_package.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, packageID, appID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelPkg, err := repo.GetForApplication(ctx, tenantID, packageID, appID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelPkg)
		require.EqualError(t, err, fmt.Sprintf("while getting object from DB: %s", testError.Error()))
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")
		rows := sqlmock.NewRows(fixPackageColumns()).
			AddRow(fixPackageRow(packageID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, packageID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", pkgEntity).Return(&model.Package{}, testError).Once()
		pgRepository := mp_package.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetForApplication(ctx, tenantID, packageID, appID)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_ListByApplicationID(t *testing.T) {
	// GIVEN
	ExpectedLimit := 3
	ExpectedOffset := 0

	inputPageSize := 3
	inputCursor := ""
	totalCount := 2
	firstPkgID := "111111111-1111-1111-1111-111111111111"
	firstPkgEntity := fixEntityPackage(firstPkgID, "foo", "bar")
	secondPkgID := "222222222-2222-2222-2222-222222222222"
	secondPkgEntity := fixEntityPackage(secondPkgID, "foo", "bar")

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.packages
		WHERE tenant_id=\$1 AND app_id = '%s'
		ORDER BY id LIMIT %d OFFSET %d`, appID, ExpectedLimit, ExpectedOffset)

	rawCountQuery := fmt.Sprintf(`SELECT COUNT(*) FROM public.packages
		WHERE tenant_id=$1 AND app_id = '%s'`, appID)
	countQuery := regexp.QuoteMeta(rawCountQuery)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixPackageColumns()).
			AddRow(fixPackageRow(firstPkgID, "placeholder")...).
			AddRow(fixPackageRow(secondPkgID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(testdb.RowCount(2))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstPkgEntity).Return(&model.Package{ID: firstPkgID}, nil)
		convMock.On("FromEntity", secondPkgEntity).Return(&model.Package{ID: secondPkgID}, nil)
		pgRepository := mp_package.NewRepository(convMock)
		// WHEN
		modelPkg, err := pgRepository.ListByApplicationID(ctx, tenantID, appID, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelPkg.Data, 2)
		assert.Equal(t, firstPkgID, modelPkg.Data[0].ID)
		assert.Equal(t, secondPkgID, modelPkg.Data[1].ID)
		assert.Equal(t, "", modelPkg.PageInfo.StartCursor)
		assert.Equal(t, totalCount, modelPkg.TotalCount)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := mp_package.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelPkg, err := repo.ListByApplicationID(ctx, tenantID, appID, inputPageSize, inputCursor)

		// then
		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelPkg)
		require.EqualError(t, err, fmt.Sprintf("while fetching list of objects from DB: %s", testError.Error()))
	})

	t.Run("returns error when conversion from entity to model failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testErr := errors.New("test error")
		rows := sqlmock.NewRows(fixPackageColumns()).
			AddRow(fixPackageRow(firstPkgID, "foo")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(testdb.RowCount(1))
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstPkgEntity).Return(&model.Package{}, testErr).Once()
		pgRepository := mp_package.NewRepository(convMock)
		//WHEN
		_, err := pgRepository.ListByApplicationID(ctx, tenantID, appID, inputPageSize, inputCursor)
		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}
