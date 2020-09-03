package mp_bundle_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/stretchr/testify/assert"

	mp_bundle "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	name := "foo"
	desc := "bar"

	bundleModel := fixBundleModel(t, name, desc)
	bundleEntity := fixEntityBundle(bundleID, name, desc)
	insertQuery := `^INSERT INTO public.bundles \(.+\) VALUES \(.+\)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defAuth, err := json.Marshal(bundleModel.DefaultInstanceAuth)
		require.NoError(t, err)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixBundleCreateArgs(string(defAuth), *bundleModel.InstanceAuthRequestInputSchema, bundleModel)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.EntityConverter{}
		convMock.On("ToEntity", bundleModel).Return(bundleEntity, nil).Once()
		pgRepository := mp_bundle.NewRepository(&convMock)
		//WHEN
		err = pgRepository.Create(ctx, bundleModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when conversion from model to entity failed", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.EntityConverter{}
		convMock.On("ToEntity", bundleModel).Return(&mp_bundle.Entity{}, errors.New("test error"))
		pgRepository := mp_bundle.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, bundleModel)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test error")
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.EntityConverter{}
		pgRepository := mp_bundle.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "model can not be nil")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE public.bundles SET name = ?, description = ?, instance_auth_request_json_schema = ?,
		default_instance_auth = ? WHERE tenant_id = ? AND id = ?`)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		bundle := fixBundleModel(t, "foo", "update")
		entity := fixEntityBundle(bundleID, "foo", "update")

		convMock := &automock.EntityConverter{}
		convMock.On("ToEntity", bundle).Return(entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(entity.Title, entity.Description, entity.InstanceAuthRequestJSONSchema, entity.DefaultInstanceAuth, tenantID, entity.ID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := mp_bundle.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, bundle)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when conversion from model to entity failed", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		bundleModel := &model.Bundle{}
		convMock := &automock.EntityConverter{}
		convMock.On("ToEntity", bundleModel).Return(&mp_bundle.Entity{}, errors.New("test error")).Once()
		pgRepository := mp_bundle.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, bundleModel)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test error")
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when model is nil", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		pgRepository := mp_bundle.NewRepository(convMock)
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
	deleteQuery := `^DELETE FROM public.bundles WHERE tenant_id = \$1 AND id = \$2$`

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, bundleID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.EntityConverter{}
	pgRepository := mp_bundle.NewRepository(convMock)
	//WHEN
	err := pgRepository.Delete(ctx, tenantID, bundleID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_Exists(t *testing.T) {
	//GIVEN
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	existQuery := regexp.QuoteMeta(`SELECT 1 FROM public.bundles WHERE tenant_id = $1 AND id = $2`)

	sqlMock.ExpectQuery(existQuery).WithArgs(tenantID, bundleID).WillReturnRows(testdb.RowWhenObjectExist())
	convMock := &automock.EntityConverter{}
	pgRepository := mp_bundle.NewRepository(convMock)
	//WHEN
	found, err := pgRepository.Exists(ctx, tenantID, bundleID)
	//THEN
	require.NoError(t, err)
	assert.True(t, found)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	// given
	bundleEntity := fixEntityBundle(bundleID, "foo", "bar")

	selectQuery := `^SELECT (.+) FROM public.bundles WHERE tenant_id = \$1 AND id = \$2$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(bundleID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, bundleID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", bundleEntity).Return(&model.Bundle{ID: bundleID, TenantID: tenantID}, nil).Once()
		pgRepository := mp_bundle.NewRepository(convMock)
		// WHEN
		modelBundle, err := pgRepository.GetByID(ctx, tenantID, bundleID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, bundleID, modelBundle.ID)
		assert.Equal(t, tenantID, modelBundle.TenantID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := mp_bundle.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, bundleID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelBundle, err := repo.GetByID(ctx, tenantID, bundleID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelBundle)
		require.EqualError(t, err, "Internal Server Error: while getting object from table public.bundles: test error")
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")
		rows := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(bundleID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, bundleID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", bundleEntity).Return(&model.Bundle{}, testError).Once()
		pgRepository := mp_bundle.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetByID(ctx, tenantID, bundleID)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_GetByInstanceAuthID(t *testing.T) {
	// given
	instanceAuthID := "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	bundleEntity := fixEntityBundle(bundleID, "foo", "bar")

	selectQuery := `^SELECT p.id, p.tenant_id, p.app_id, p.name, p.description, p.instance_auth_request_json_schema, p.default_instance_auth FROM public.bundles AS p JOIN public.bundle_instance_auths AS a on a.bundle_id=p.id where a.tenant_id=\$1 AND a.id=\$2`

	t.Run("Success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(bundleID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, instanceAuthID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", bundleEntity).Return(&model.Bundle{ID: bundleID, TenantID: tenantID, ApplicationID: appID}, nil).Once()
		pgRepository := mp_bundle.NewRepository(convMock)
		// WHEN
		modelBundle, err := pgRepository.GetByInstanceAuthID(ctx, tenantID, instanceAuthID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, bundleID, modelBundle.ID)
		assert.Equal(t, tenantID, modelBundle.TenantID)
		assert.Equal(t, appID, modelBundle.ApplicationID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := mp_bundle.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, instanceAuthID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelBundle, err := repo.GetByInstanceAuthID(ctx, tenantID, instanceAuthID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelBundle)
		require.EqualError(t, err, fmt.Sprintf("while getting Bundle by Instance Auth ID: %s", testError.Error()))
	})

	t.Run("Returns error when conversion failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")
		rows := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(bundleID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, instanceAuthID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", bundleEntity).Return(&model.Bundle{}, testError).Once()
		pgRepository := mp_bundle.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetByInstanceAuthID(ctx, tenantID, instanceAuthID)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_GetForApplication(t *testing.T) {
	// given
	bundleEntity := fixEntityBundle(bundleID, "foo", "bar")

	selectQuery := `^SELECT (.+) FROM public.bundles WHERE tenant_id = \$1 AND id = \$2 AND app_id = \$3`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(bundleID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, bundleID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", bundleEntity).Return(&model.Bundle{ID: bundleID, TenantID: tenantID, ApplicationID: appID}, nil).Once()
		pgRepository := mp_bundle.NewRepository(convMock)
		// WHEN
		modelBundle, err := pgRepository.GetForApplication(ctx, tenantID, bundleID, appID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, bundleID, modelBundle.ID)
		assert.Equal(t, tenantID, modelBundle.TenantID)
		assert.Equal(t, appID, modelBundle.ApplicationID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := mp_bundle.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, bundleID, appID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelBundle, err := repo.GetForApplication(ctx, tenantID, bundleID, appID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelBundle)
		require.EqualError(t, err, "Internal Server Error: while getting object from table public.bundles: test error")
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")
		rows := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(bundleID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, bundleID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", bundleEntity).Return(&model.Bundle{}, testError).Once()
		pgRepository := mp_bundle.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetForApplication(ctx, tenantID, bundleID, appID)
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
	firstBundleID := "111111111-1111-1111-1111-111111111111"
	firstBundleEntity := fixEntityBundle(firstBundleID, "foo", "bar")
	secondBundleID := "222222222-2222-2222-2222-222222222222"
	secondBundleEntity := fixEntityBundle(secondBundleID, "foo", "bar")

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.bundles
		WHERE tenant_id = \$1 AND app_id = \$2
		ORDER BY id LIMIT %d OFFSET %d`, ExpectedLimit, ExpectedOffset)

	rawCountQuery := `SELECT COUNT(*) FROM public.bundles
		WHERE tenant_id = $1 AND app_id = $2`
	countQuery := regexp.QuoteMeta(rawCountQuery)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(firstBundleID, "placeholder")...).
			AddRow(fixBundleRow(secondBundleID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(testdb.RowCount(2))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstBundleEntity).Return(&model.Bundle{ID: firstBundleID}, nil)
		convMock.On("FromEntity", secondBundleEntity).Return(&model.Bundle{ID: secondBundleID}, nil)
		pgRepository := mp_bundle.NewRepository(convMock)
		// WHEN
		modelBundle, err := pgRepository.ListByApplicationID(ctx, tenantID, appID, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelBundle.Data, 2)
		assert.Equal(t, firstBundleID, modelBundle.Data[0].ID)
		assert.Equal(t, secondBundleID, modelBundle.Data[1].ID)
		assert.Equal(t, "", modelBundle.PageInfo.StartCursor)
		assert.Equal(t, totalCount, modelBundle.TotalCount)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := mp_bundle.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelBundle, err := repo.ListByApplicationID(ctx, tenantID, appID, inputPageSize, inputCursor)

		// then
		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelBundle)
		require.EqualError(t, err, fmt.Sprintf("while fetching list of objects from DB: %s", testError.Error()))
	})

	t.Run("returns error when conversion from entity to model failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testErr := errors.New("test error")
		rows := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(firstBundleID, "foo")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(testdb.RowCount(1))
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstBundleEntity).Return(&model.Bundle{}, testErr).Once()
		pgRepository := mp_bundle.NewRepository(convMock)
		//WHEN
		_, err := pgRepository.ListByApplicationID(ctx, tenantID, appID, inputPageSize, inputCursor)
		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}
