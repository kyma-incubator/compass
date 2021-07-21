package bundlereferences_test

import (
	"context"

	"errors"

	"fmt"

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

	selectQueryForAPIWithoutBundleID := fmt.Sprintf(`SELECT (.+) FROM public\.bundle_references WHERE %s AND api_def_id = \$2`, fixTenantIsolationSubquery())
	selectQueryForAPIWithBundleID := fmt.Sprintf(`SELECT (.+) FROM public\.bundle_references WHERE %s AND api_def_id = \$2 AND bundle_id = \$3`, fixTenantIsolationSubquery())

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
	selectQuery := fmt.Sprintf(`SELECT (.+) FROM public\.bundle_references WHERE %s AND api_def_id = \$2`, fixTenantIsolationSubquery())

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
	updateQueryWithoutBundleID := regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.bundle_references SET api_def_id = ?, event_def_id = ?, api_def_url = ? WHERE %s AND api_def_id = ?`, fixUpdateTenantIsolationSubquery()))
	updateQueryWithBundleID := regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.bundle_references SET api_def_id = ?, event_def_id = ?, bundle_id = ?, api_def_url = ? WHERE %s AND api_def_id = ? AND bundle_id = ?`, fixUpdateTenantIsolationSubquery()))

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
		deleteQuery := fmt.Sprintf(`DELETE FROM public\.bundle_references WHERE %s AND api_def_id = \$2 AND bundle_id = \$3`, fixTenantIsolationSubquery())

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

func TestPgRepository_ListAllForBundle(t *testing.T) {
	// GIVEN
	inputCursor := ""
	firstTargetURL := "https://test.com"
	secondTargetURL := "https://test2.com"
	firstBndlID := "111111111-1111-1111-1111-111111111111"
	secondBndlID := "222222222-2222-2222-2222-222222222222"
	firstAPIID := "333333333-3333-3333-3333-333333333333"
	secondAPIID := "444444444-4444-4444-4444-444444444444"
	firstEventID := "555555555-5555-5555-5555-555555555555"
	secondEventID := "666666666-6666-6666-6666-666666666666"

	firstAPIBndlRefEntity := fixAPIBundleReferenceEntityWithArgs(firstBndlID, firstAPIID, firstTargetURL)
	secondAPIBndlRefEntity := fixAPIBundleReferenceEntityWithArgs(secondBndlID, secondAPIID, secondTargetURL)
	firstEventBndlRefEntity := fixEventBundleReferenceEntityWithArgs(firstBndlID, firstEventID)
	secondEventBndlRefEntity := fixEventBundleReferenceEntityWithArgs(secondBndlID, secondEventID)
	bundleIDs := []string{firstBndlID, secondBndlID}

	selectQuery := fmt.Sprintf(`\(SELECT (.+) FROM public\.bundle_references 
		WHERE %s AND api_def_id IS NOT NULL AND bundle_id = \$2 ORDER BY api_def_id ASC, bundle_id ASC, api_def_url ASC LIMIT \$3 OFFSET \$4\) UNION 
		\(SELECT (.+) FROM public\.bundle_references WHERE %s AND api_def_id IS NOT NULL AND bundle_id = \$6 ORDER BY api_def_id ASC, bundle_id ASC, api_def_url ASC LIMIT \$7 OFFSET \$8\)`, fixTenantIsolationSubqueryWithArg(1), fixTenantIsolationSubqueryWithArg(5))

	countQuery := fmt.Sprintf(`SELECT bundle_id AS id, COUNT\(\*\) AS total_count FROM public.bundle_references WHERE %s AND api_def_id IS NOT NULL GROUP BY bundle_id ORDER BY bundle_id ASC`, fixTenantIsolationSubquery())

	t.Run("success when everything is returned for APIs", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1

		totalCountForFirstBundle := 1
		totalCountForSecondBundle := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutEventIDWithArgs(firstBndlID, firstAPIID, firstTargetURL)...).
			AddRow(fixBundleReferenceRowWithoutEventIDWithArgs(secondBndlID, secondAPIID, secondTargetURL)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstBndlID, ExpectedLimit, ExpectedOffset, tenantID, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", firstAPIBndlRefEntity).Return(model.BundleReference{
			Tenant:              tenantID,
			BundleID:            str.Ptr(firstBndlID),
			ObjectType:          model.BundleAPIReference,
			ObjectID:            str.Ptr(firstAPIID),
			APIDefaultTargetURL: str.Ptr(firstTargetURL),
		}, nil)
		convMock.On("FromEntity", secondAPIBndlRefEntity).Return(model.BundleReference{
			Tenant:              tenantID,
			BundleID:            str.Ptr(secondBndlID),
			ObjectType:          model.BundleAPIReference,
			ObjectID:            str.Ptr(secondAPIID),
			APIDefaultTargetURL: str.Ptr(secondTargetURL),
		}, nil)
		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		modelBndlRefs, totalCounts, err := pgRepository.ListAllForBundle(ctx, model.BundleAPIReference, tenantID, bundleIDs, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelBndlRefs, 2)
		assert.Equal(t, firstBndlID, *modelBndlRefs[0].BundleID)
		assert.Equal(t, secondBndlID, *modelBndlRefs[1].BundleID)
		assert.Equal(t, firstAPIID, *modelBndlRefs[0].ObjectID)
		assert.Equal(t, secondAPIID, *modelBndlRefs[1].ObjectID)
		assert.Equal(t, firstTargetURL, *modelBndlRefs[0].APIDefaultTargetURL)
		assert.Equal(t, secondTargetURL, *modelBndlRefs[1].APIDefaultTargetURL)
		assert.Equal(t, totalCountForFirstBundle, totalCounts[firstBndlID])
		assert.Equal(t, totalCountForSecondBundle, totalCounts[secondBndlID])
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("success when everything is returned for Events", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1

		totalCountForFirstBundle := 1
		totalCountForSecondBundle := 1

		selectQueryForEvents := fmt.Sprintf(`\(SELECT (.+) FROM public\.bundle_references 
		WHERE %s AND event_def_id IS NOT NULL AND bundle_id = \$2 ORDER BY event_def_id ASC, bundle_id ASC LIMIT \$3 OFFSET \$4\) UNION 
		\(SELECT (.+) FROM public\.bundle_references WHERE %s AND event_def_id IS NOT NULL AND bundle_id = \$6 ORDER BY event_def_id ASC, bundle_id ASC LIMIT \$7 OFFSET \$8\)`, fixTenantIsolationSubqueryWithArg(1), fixTenantIsolationSubqueryWithArg(5))

		countQueryForEvents := fmt.Sprintf(`SELECT bundle_id AS id, COUNT\(\*\) AS total_count FROM public.bundle_references WHERE %s AND event_def_id IS NOT NULL GROUP BY bundle_id ORDER BY bundle_id ASC`, fixTenantIsolationSubquery())

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutAPIIDWithArgs(firstBndlID, firstEventID)...).
			AddRow(fixBundleReferenceRowWithoutAPIIDWithArgs(secondBndlID, secondEventID)...)

		sqlMock.ExpectQuery(selectQueryForEvents).
			WithArgs(tenantID, firstBndlID, ExpectedLimit, ExpectedOffset, tenantID, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryForEvents).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", firstEventBndlRefEntity).Return(model.BundleReference{
			Tenant:              tenantID,
			BundleID:            str.Ptr(firstBndlID),
			ObjectType:          model.BundleEventReference,
			ObjectID:            str.Ptr(firstEventID),
		}, nil)
		convMock.On("FromEntity", secondEventBndlRefEntity).Return(model.BundleReference{
			Tenant:              tenantID,
			BundleID:            str.Ptr(secondBndlID),
			ObjectType:          model.BundleEventReference,
			ObjectID:            str.Ptr(secondEventID),
		}, nil)
		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		modelBndlRefs, totalCounts, err := pgRepository.ListAllForBundle(ctx, model.BundleEventReference, tenantID, bundleIDs, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelBndlRefs, 2)
		assert.Equal(t, firstBndlID, *modelBndlRefs[0].BundleID)
		assert.Equal(t, secondBndlID, *modelBndlRefs[1].BundleID)
		assert.Equal(t, firstEventID, *modelBndlRefs[0].ObjectID)
		assert.Equal(t, secondEventID, *modelBndlRefs[1].ObjectID)
		assert.Equal(t, totalCountForFirstBundle, totalCounts[firstBndlID])
		assert.Equal(t, totalCountForSecondBundle, totalCounts[secondBndlID])
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("success when there are more records", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1

		totalCountForFirstBundle := 10
		totalCountForSecondBundle := 10

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutEventIDWithArgs(firstBndlID, firstAPIID, firstTargetURL)...).
			AddRow(fixBundleReferenceRowWithoutEventIDWithArgs(secondBndlID, secondAPIID, secondTargetURL)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstBndlID, ExpectedLimit, ExpectedOffset, tenantID, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", firstAPIBndlRefEntity).Return(model.BundleReference{
			Tenant:              tenantID,
			BundleID:            str.Ptr(firstBndlID),
			ObjectType:          model.BundleAPIReference,
			ObjectID:            str.Ptr(firstAPIID),
			APIDefaultTargetURL: str.Ptr(firstTargetURL),
		}, nil)
		convMock.On("FromEntity", secondAPIBndlRefEntity).Return(model.BundleReference{
			Tenant:              tenantID,
			BundleID:            str.Ptr(secondBndlID),
			ObjectType:          model.BundleAPIReference,
			ObjectID:            str.Ptr(secondAPIID),
			APIDefaultTargetURL: str.Ptr(secondTargetURL),
		}, nil)
		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		modelBndlRefs, totalCounts, err := pgRepository.ListAllForBundle(ctx, model.BundleAPIReference, tenantID, bundleIDs, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelBndlRefs, 2)
		assert.Equal(t, firstBndlID, *modelBndlRefs[0].BundleID)
		assert.Equal(t, secondBndlID, *modelBndlRefs[1].BundleID)
		assert.Equal(t, firstAPIID, *modelBndlRefs[0].ObjectID)
		assert.Equal(t, secondAPIID, *modelBndlRefs[1].ObjectID)
		assert.Equal(t, firstTargetURL, *modelBndlRefs[0].APIDefaultTargetURL)
		assert.Equal(t, secondTargetURL, *modelBndlRefs[1].APIDefaultTargetURL)
		assert.Equal(t, totalCountForFirstBundle, totalCounts[firstBndlID])
		assert.Equal(t, totalCountForSecondBundle, totalCounts[secondBndlID])
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when conversion from entity fails", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1
		totalCountForFirstBundle := 1
		totalCountForSecondBundle := 1

		testErr := errors.New("test error")

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutEventIDWithArgs(firstBndlID, firstAPIID, firstTargetURL)...).
			AddRow(fixBundleReferenceRowWithoutEventIDWithArgs(secondBndlID, secondAPIID, secondTargetURL)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstBndlID, ExpectedLimit, ExpectedOffset, tenantID, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", firstAPIBndlRefEntity).Return(model.BundleReference{}, testErr)
		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		_, _, err := pgRepository.ListAllForBundle(ctx, model.BundleAPIReference, tenantID, bundleIDs, inputPageSize, inputCursor)
		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		inputPageSize := 1
		ExpectedLimit := 1
		ExpectedOffset := 0

		pgRepository := bundlereferences.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstBndlID, ExpectedLimit, ExpectedOffset, tenantID, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelBndlRefs, totalCounts, err := pgRepository.ListAllForBundle(ctx, model.BundleAPIReference, tenantID, bundleIDs, inputPageSize, inputCursor)

		// then
		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelBndlRefs)
		assert.Nil(t, totalCounts)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
