package bundlereferences_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_GetByID(t *testing.T) {
	// GIVEN
	apiBundleReferenceEntity := fixAPIBundleReferenceEntity()

	selectQueryForAPIWithoutBundleID := `SELECT (.+) FROM public\.bundle_references WHERE tenant_id = \$1 AND api_def_id = \$2`
	selectQueryForAPIWithBundleID := `SELECT (.+) FROM public\.bundle_references WHERE tenant_id = \$1 AND api_def_id = \$2 AND bundle_id = \$3`

	t.Run("success when no bundleID", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutEventID()...)

		sqlMock.ExpectQuery(selectQueryForAPIWithoutBundleID).
			WithArgs(tenantID, apiDefID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", apiBundleReferenceEntity).Return(fixAPIBundleReferenceModel(), nil).Once()
		pgRepository := bundlereferences.NewRepository(convMock)

		// WHEN
		modelAPIBundleRef, err := pgRepository.GetByID(ctx, model.BundleAPIReference, tenantID, str.Ptr(apiDefID), nil)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, apiDefID, *modelAPIBundleRef.ObjectID)
		assert.Equal(t, tenantID, modelAPIBundleRef.Tenant)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
	t.Run("success with bundleID", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutEventID()...)

		sqlMock.ExpectQuery(selectQueryForAPIWithBundleID).
			WithArgs(tenantID, apiDefID, bundleID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", apiBundleReferenceEntity).Return(fixAPIBundleReferenceModel(), nil).Once()
		pgRepository := bundlereferences.NewRepository(convMock)

		// WHEN
		modelAPIBundleRef, err := pgRepository.GetByID(ctx, model.BundleAPIReference, tenantID, str.Ptr(apiDefID), str.Ptr(bundleID))

		//THEN
		require.NoError(t, err)
		assert.Equal(t, apiDefID, *modelAPIBundleRef.ObjectID)
		assert.Equal(t, tenantID, modelAPIBundleRef.Tenant)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_GetBundleIDsForObject(t *testing.T) {
	selectQuery := `SELECT (.+) FROM public\.bundle_references WHERE tenant_id = \$1 AND api_def_id = \$2`

	t.Run("success", func(t *testing.T) {
		// GIVEN
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows([]string{"bundle_id"}).
			AddRow(fixBundleIDs(bundleID)...).
			AddRow(fixBundleIDs(secondBundleID)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, apiDefID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.BundleReferenceConverter{}
		pgRepository := bundlereferences.NewRepository(convMock)

		// WHEN
		bundleIDs, err := pgRepository.GetBundleIDsForObject(ctx, tenantID, model.BundleAPIReference, str.Ptr(apiDefID))

		//THEN
		require.NoError(t, err)
		assert.Equal(t, bundleID, bundleIDs[0])
		assert.Equal(t, secondBundleID, bundleIDs[1])
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	bundleRefModel := fixAPIBundleReferenceModel()
	bundleRefEntity := fixAPIBundleReferenceEntity()
	insertQuery := `INSERT INTO public\.bundle_references (.+) VALUES (.+)`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixBundleReferenceCreateArgs(&bundleRefModel)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.BundleReferenceConverter{}
		convMock.On("ToEntity", bundleRefModel).Return(bundleRefEntity, nil).Once()
		pgRepository := bundlereferences.NewRepository(&convMock)

		//WHEN
		err := pgRepository.Create(ctx, &bundleRefModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
	t.Run("error when item is nil", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.BundleReferenceConverter{}
		pgRepository := bundlereferences.NewRepository(&convMock)

		// WHEN
		err := pgRepository.Create(ctx, nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item can not be empty")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Update(t *testing.T) {
	updateQueryWithoutBundleID := regexp.QuoteMeta(`UPDATE public.bundle_references SET api_def_id = ?, event_def_id = ?, api_def_url = ? WHERE tenant_id = ? AND api_def_id = ?`)
	updateQueryWithBundleID := regexp.QuoteMeta(`UPDATE public.bundle_references SET api_def_id = ?, event_def_id = ?, bundle_id = ?, api_def_url = ? WHERE tenant_id = ? AND api_def_id = ? AND bundle_id = ?`)

	t.Run("success without bundleID", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		apiBundleReferenceModel := fixAPIBundleReferenceModel()
		apiBundleReferenceModel.BundleID = nil
		apiBundleReferenceEntity := fixAPIBundleReferenceEntity()

		convMock := &automock.BundleReferenceConverter{}
		convMock.On("ToEntity", apiBundleReferenceModel).Return(apiBundleReferenceEntity, nil)
		sqlMock.ExpectExec(updateQueryWithoutBundleID).
			WithArgs(apiBundleReferenceEntity.APIDefID, apiBundleReferenceEntity.EventDefID, apiBundleReferenceEntity.APIDefaultTargetURL, tenantID, apiBundleReferenceEntity.APIDefID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := bundlereferences.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, &apiBundleReferenceModel)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
	t.Run("success with bundleID", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		apiBundleReferenceModel := fixAPIBundleReferenceModel()
		apiBundleReferenceEntity := fixAPIBundleReferenceEntity()

		convMock := &automock.BundleReferenceConverter{}
		convMock.On("ToEntity", apiBundleReferenceModel).Return(apiBundleReferenceEntity, nil)
		sqlMock.ExpectExec(updateQueryWithBundleID).
			WithArgs(apiBundleReferenceEntity.APIDefID, apiBundleReferenceEntity.EventDefID, apiBundleReferenceEntity.BundleID, apiBundleReferenceEntity.APIDefaultTargetURL, tenantID, apiBundleReferenceEntity.APIDefID, apiBundleReferenceEntity.BundleID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := bundlereferences.NewRepository(convMock)

		//WHEN
		err := pgRepository.Update(ctx, &apiBundleReferenceModel)

		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
	t.Run("error when item is nil", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.BundleReferenceConverter{}
		pgRepository := bundlereferences.NewRepository(&convMock)

		// WHEN
		err := pgRepository.Update(ctx, nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item cannot be nil")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_DeleteByReferenceObjectID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		deleteQuery := `DELETE FROM public\.bundle_references WHERE tenant_id = \$1 AND api_def_id = \$2 AND bundle_id = \$3`

		sqlMock.ExpectExec(deleteQuery).
			WithArgs(tenantID, apiDefID, bundleID).
			WillReturnResult(sqlmock.NewResult(-1, 1))
		convMock := &automock.BundleReferenceConverter{}
		pgRepository := bundlereferences.NewRepository(convMock)

		//WHEN
		err := pgRepository.DeleteByReferenceObjectID(ctx, tenantID, bundleID, model.BundleAPIReference, apiDefID)

		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}
