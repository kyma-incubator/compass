package bundlereferences_test

import (
	"context"
	"errors"

	"github.com/kyma-incubator/compass/components/director/pkg/scope"

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

	selectQueryForAPIWithoutBundleID := `SELECT (.+) FROM public\.bundle_references WHERE api_def_id = \$1`
	selectQueryForAPIWithBundleID := `SELECT (.+) FROM public\.bundle_references WHERE api_def_id = \$1 AND bundle_id = \$2`

	t.Run("success when no bundleID", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutEventID()...)

		sqlMock.ExpectQuery(selectQueryForAPIWithoutBundleID).
			WithArgs(apiDefID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", apiBundleReferenceEntity).Return(fixAPIBundleReferenceModel(), nil).Once()
		pgRepository := bundlereferences.NewRepository(convMock)

		// WHEN
		modelAPIBundleRef, err := pgRepository.GetByID(ctx, model.BundleAPIReference, str.Ptr(apiDefID), nil)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, apiDefID, *modelAPIBundleRef.ObjectID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
	t.Run("success with bundleID", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutEventID()...)

		sqlMock.ExpectQuery(selectQueryForAPIWithBundleID).
			WithArgs(apiDefID, bundleID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", apiBundleReferenceEntity).Return(fixAPIBundleReferenceModel(), nil).Once()
		pgRepository := bundlereferences.NewRepository(convMock)

		// WHEN
		modelAPIBundleRef, err := pgRepository.GetByID(ctx, model.BundleAPIReference, str.Ptr(apiDefID), str.Ptr(bundleID))

		// THEN
		require.NoError(t, err)
		assert.Equal(t, apiDefID, *modelAPIBundleRef.ObjectID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_GetBundleIDsForObject(t *testing.T) {
	selectQuery := `SELECT (.+) FROM public\.bundle_references WHERE api_def_id = \$1`

	t.Run("success", func(t *testing.T) {
		// GIVEN
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows([]string{"bundle_id"}).
			AddRow(fixBundleIDs(bundleID)...).
			AddRow(fixBundleIDs(secondBundleID)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(apiDefID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.BundleReferenceConverter{}
		pgRepository := bundlereferences.NewRepository(convMock)

		// WHEN
		bundleIDs, err := pgRepository.GetBundleIDsForObject(ctx, model.BundleAPIReference, str.Ptr(apiDefID))

		// THEN
		require.NoError(t, err)
		assert.Equal(t, bundleID, bundleIDs[0])
		assert.Equal(t, secondBundleID, bundleIDs[1])
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
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

		// WHEN
		err := pgRepository.Create(ctx, &bundleRefModel)
		// THEN
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
	updateQueryWithoutBundleID := regexp.QuoteMeta(`UPDATE public.bundle_references SET api_def_id = ?, event_def_id = ?, api_def_url = ? WHERE api_def_id = ?`)
	updateQueryWithBundleID := regexp.QuoteMeta(`UPDATE public.bundle_references SET api_def_id = ?, event_def_id = ?, bundle_id = ?, api_def_url = ?, is_default_bundle = ? WHERE api_def_id = ? AND bundle_id = ?`)

	t.Run("success without bundleID", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		apiBundleReferenceModel := fixAPIBundleReferenceModel()
		apiBundleReferenceModel.BundleID = nil
		apiBundleReferenceEntity := fixAPIBundleReferenceEntity()

		convMock := &automock.BundleReferenceConverter{}
		convMock.On("ToEntity", apiBundleReferenceModel).Return(apiBundleReferenceEntity, nil)
		sqlMock.ExpectExec(updateQueryWithoutBundleID).
			WithArgs(apiBundleReferenceEntity.APIDefID, apiBundleReferenceEntity.EventDefID, apiBundleReferenceEntity.APIDefaultTargetURL, apiBundleReferenceEntity.APIDefID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		err := pgRepository.Update(ctx, &apiBundleReferenceModel)
		// THEN
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
			WithArgs(apiBundleReferenceEntity.APIDefID, apiBundleReferenceEntity.EventDefID, apiBundleReferenceEntity.BundleID, apiBundleReferenceEntity.APIDefaultTargetURL, apiBundleReferenceEntity.IsDefaultBundle, apiBundleReferenceEntity.APIDefID, apiBundleReferenceEntity.BundleID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := bundlereferences.NewRepository(convMock)

		// WHEN
		err := pgRepository.Update(ctx, &apiBundleReferenceModel)

		// THEN
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
		deleteQuery := `DELETE FROM public\.bundle_references WHERE api_def_id = \$1 AND bundle_id = \$2`

		sqlMock.ExpectExec(deleteQuery).
			WithArgs(apiDefID, bundleID).
			WillReturnResult(sqlmock.NewResult(-1, 1))
		convMock := &automock.BundleReferenceConverter{}
		pgRepository := bundlereferences.NewRepository(convMock)

		// WHEN
		err := pgRepository.DeleteByReferenceObjectID(ctx, bundleID, model.BundleAPIReference, apiDefID)

		// THEN
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
	scopesWithInternalVisibility := []string{"internal_visibility:read"}
	scopesWithoutInternalVisibility := []string{"test:test"}
	publicVisibility := "public"

	firstAPIBndlRefEntity := fixAPIBundleReferenceEntityWithArgs(firstBndlID, firstAPIID, firstTargetURL)
	secondAPIBndlRefEntity := fixAPIBundleReferenceEntityWithArgs(secondBndlID, secondAPIID, secondTargetURL)
	firstEventBndlRefEntity := fixEventBundleReferenceEntityWithArgs(firstBndlID, firstEventID)
	secondEventBndlRefEntity := fixEventBundleReferenceEntityWithArgs(secondBndlID, secondEventID)
	bundleIDs := []string{firstBndlID, secondBndlID}

	selectQueryForAPIs := `^\(SELECT (.+) FROM public\.bundle_references 
		WHERE api_def_id IS NOT NULL AND bundle_id = \$1 ORDER BY api_def_id ASC, bundle_id ASC, api_def_url ASC LIMIT \$2 OFFSET \$3\) UNION 
		\(SELECT (.+) FROM public\.bundle_references WHERE api_def_id IS NOT NULL AND bundle_id = \$4 ORDER BY api_def_id ASC, bundle_id ASC, api_def_url ASC LIMIT \$5 OFFSET \$6\)`

	countQueryForAPIs := `SELECT bundle_id AS id, COUNT\(\*\) AS total_count FROM public.bundle_references WHERE api_def_id IS NOT NULL GROUP BY bundle_id ORDER BY bundle_id ASC`

	selectQueryWithVisibilityCheckForAPIs := `^\(SELECT (.+) FROM public\.bundle_references 
		WHERE api_def_id IN \(SELECT id FROM api_definitions WHERE visibility = \$1\) AND api_def_id IS NOT NULL AND bundle_id = \$2 ORDER BY api_def_id ASC, bundle_id ASC, api_def_url ASC LIMIT \$3 OFFSET \$4\) UNION 
		\(SELECT (.+) FROM public\.bundle_references WHERE api_def_id IN \(SELECT id FROM api_definitions WHERE visibility = \$5\) AND api_def_id IS NOT NULL AND bundle_id = \$6 ORDER BY api_def_id ASC, bundle_id ASC, api_def_url ASC LIMIT \$7 OFFSET \$8\)`

	countQueryWithVisibilityCheckForAPIs := `SELECT bundle_id AS id, COUNT\(\*\) AS total_count FROM public.bundle_references WHERE api_def_id IN \(SELECT id FROM api_definitions WHERE visibility = \$1\) AND api_def_id IS NOT NULL GROUP BY bundle_id ORDER BY bundle_id ASC`

	// queries for Events

	selectQueryForEvents := `^\(SELECT (.+) FROM public\.bundle_references 
		WHERE event_def_id IS NOT NULL AND bundle_id = \$1 ORDER BY event_def_id ASC, bundle_id ASC LIMIT \$2 OFFSET \$3\) UNION 
		\(SELECT (.+) FROM public\.bundle_references WHERE event_def_id IS NOT NULL AND bundle_id = \$4 ORDER BY event_def_id ASC, bundle_id ASC LIMIT \$5 OFFSET \$6\)`

	countQueryForEvents := `SELECT bundle_id AS id, COUNT\(\*\) AS total_count FROM public.bundle_references WHERE event_def_id IS NOT NULL GROUP BY bundle_id ORDER BY bundle_id ASC`

	selectQueryWithVisibilityCheckForEvents := `^\(SELECT (.+) FROM public\.bundle_references 
		WHERE api_def_id IN \(SELECT id FROM api_definitions WHERE visibility = \$1\) AND event_def_id IS NOT NULL AND bundle_id = \$2 ORDER BY event_def_id ASC, bundle_id ASC LIMIT \$3 OFFSET \$4\) UNION 
		\(SELECT (.+) FROM public\.bundle_references WHERE api_def_id IN \(SELECT id FROM api_definitions WHERE visibility = \$5\) AND event_def_id IS NOT NULL AND bundle_id = \$6 ORDER BY event_def_id ASC, bundle_id ASC LIMIT \$7 OFFSET \$8\)`

	countQueryWithVisibilityCheckForEvents := `SELECT bundle_id AS id, COUNT\(\*\) AS total_count FROM public.bundle_references WHERE api_def_id IN \(SELECT id FROM api_definitions WHERE visibility = \$1\) AND event_def_id IS NOT NULL GROUP BY bundle_id ORDER BY bundle_id ASC`

	t.Run("success when everything is returned for APIs when there is internal_visibility scope", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1

		totalCountForFirstBundle := 1
		totalCountForSecondBundle := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutEventIDWithArgs(firstBndlID, firstAPIID, firstTargetURL)...).
			AddRow(fixBundleReferenceRowWithoutEventIDWithArgs(secondBndlID, secondAPIID, secondTargetURL)...)

		sqlMock.ExpectQuery(selectQueryForAPIs).
			WithArgs(firstBndlID, ExpectedLimit, ExpectedOffset, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryForAPIs).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		ctx = scope.SaveToContext(ctx, scopesWithInternalVisibility)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", firstAPIBndlRefEntity).Return(model.BundleReference{
			BundleID:            str.Ptr(firstBndlID),
			ObjectType:          model.BundleAPIReference,
			ObjectID:            str.Ptr(firstAPIID),
			APIDefaultTargetURL: str.Ptr(firstTargetURL),
		}, nil)
		convMock.On("FromEntity", secondAPIBndlRefEntity).Return(model.BundleReference{
			BundleID:            str.Ptr(secondBndlID),
			ObjectType:          model.BundleAPIReference,
			ObjectID:            str.Ptr(secondAPIID),
			APIDefaultTargetURL: str.Ptr(secondTargetURL),
		}, nil)
		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		modelBndlRefs, totalCounts, err := pgRepository.ListByBundleIDs(ctx, model.BundleAPIReference, bundleIDs, inputPageSize, inputCursor)
		// THEN
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

	t.Run("success when there is no internal_visibility scope and result for APIs is filtered", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1

		totalCountForFirstBundle := 0
		totalCountForSecondBundle := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutEventIDWithArgs(secondBndlID, secondAPIID, secondTargetURL)...)

		sqlMock.ExpectQuery(selectQueryWithVisibilityCheckForAPIs).
			WithArgs(publicVisibility, firstBndlID, ExpectedLimit, ExpectedOffset, publicVisibility, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryWithVisibilityCheckForAPIs).
			WithArgs(publicVisibility).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		ctx = scope.SaveToContext(ctx, scopesWithoutInternalVisibility)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", secondAPIBndlRefEntity).Return(model.BundleReference{
			BundleID:            str.Ptr(secondBndlID),
			ObjectType:          model.BundleAPIReference,
			ObjectID:            str.Ptr(secondAPIID),
			APIDefaultTargetURL: str.Ptr(secondTargetURL),
		}, nil)
		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		modelBndlRefs, totalCounts, err := pgRepository.ListByBundleIDs(ctx, model.BundleAPIReference, bundleIDs, inputPageSize, inputCursor)
		// THEN
		require.NoError(t, err)
		require.Len(t, modelBndlRefs, 1)
		assert.Equal(t, secondBndlID, *modelBndlRefs[0].BundleID)
		assert.Equal(t, secondAPIID, *modelBndlRefs[0].ObjectID)
		assert.Equal(t, secondTargetURL, *modelBndlRefs[0].APIDefaultTargetURL)
		assert.Equal(t, totalCountForFirstBundle, totalCounts[firstBndlID])
		assert.Equal(t, totalCountForSecondBundle, totalCounts[secondBndlID])
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("success when everything is returned for Events when there is internal_visibility scope", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1

		totalCountForFirstBundle := 1
		totalCountForSecondBundle := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutAPIIDWithArgs(firstBndlID, firstEventID)...).
			AddRow(fixBundleReferenceRowWithoutAPIIDWithArgs(secondBndlID, secondEventID)...)

		sqlMock.ExpectQuery(selectQueryForEvents).
			WithArgs(firstBndlID, ExpectedLimit, ExpectedOffset, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryForEvents).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		ctx = scope.SaveToContext(ctx, scopesWithInternalVisibility)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", firstEventBndlRefEntity).Return(model.BundleReference{
			BundleID:   str.Ptr(firstBndlID),
			ObjectType: model.BundleEventReference,
			ObjectID:   str.Ptr(firstEventID),
		}, nil)
		convMock.On("FromEntity", secondEventBndlRefEntity).Return(model.BundleReference{
			BundleID:   str.Ptr(secondBndlID),
			ObjectType: model.BundleEventReference,
			ObjectID:   str.Ptr(secondEventID),
		}, nil)
		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		modelBndlRefs, totalCounts, err := pgRepository.ListByBundleIDs(ctx, model.BundleEventReference, bundleIDs, inputPageSize, inputCursor)
		// THEN
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

	t.Run("success when there is no internal_visibility scope and result for Events is filtered", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1

		totalCountForFirstBundle := 0
		totalCountForSecondBundle := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutAPIIDWithArgs(secondBndlID, secondEventID)...)

		sqlMock.ExpectQuery(selectQueryWithVisibilityCheckForEvents).
			WithArgs(publicVisibility, firstBndlID, ExpectedLimit, ExpectedOffset, publicVisibility, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryWithVisibilityCheckForEvents).
			WithArgs(publicVisibility).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		ctx = scope.SaveToContext(ctx, scopesWithoutInternalVisibility)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", secondEventBndlRefEntity).Return(model.BundleReference{
			BundleID:   str.Ptr(secondBndlID),
			ObjectType: model.BundleEventReference,
			ObjectID:   str.Ptr(secondEventID),
		}, nil)
		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		modelBndlRefs, totalCounts, err := pgRepository.ListByBundleIDs(ctx, model.BundleEventReference, bundleIDs, inputPageSize, inputCursor)
		// THEN
		require.NoError(t, err)
		require.Len(t, modelBndlRefs, 1)
		assert.Equal(t, secondBndlID, *modelBndlRefs[0].BundleID)
		assert.Equal(t, secondEventID, *modelBndlRefs[0].ObjectID)
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

		sqlMock.ExpectQuery(selectQueryForAPIs).
			WithArgs(firstBndlID, ExpectedLimit, ExpectedOffset, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryForAPIs).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		ctx = scope.SaveToContext(ctx, scopesWithInternalVisibility)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", firstAPIBndlRefEntity).Return(model.BundleReference{
			BundleID:            str.Ptr(firstBndlID),
			ObjectType:          model.BundleAPIReference,
			ObjectID:            str.Ptr(firstAPIID),
			APIDefaultTargetURL: str.Ptr(firstTargetURL),
		}, nil)
		convMock.On("FromEntity", secondAPIBndlRefEntity).Return(model.BundleReference{
			BundleID:            str.Ptr(secondBndlID),
			ObjectType:          model.BundleAPIReference,
			ObjectID:            str.Ptr(secondAPIID),
			APIDefaultTargetURL: str.Ptr(secondTargetURL),
		}, nil)
		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		modelBndlRefs, totalCounts, err := pgRepository.ListByBundleIDs(ctx, model.BundleAPIReference, bundleIDs, inputPageSize, inputCursor)
		// THEN
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

	t.Run("returns both public and internal/private Events when check for internal scope fails", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1

		totalCountForFirstBundle := 1
		totalCountForSecondBundle := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutAPIIDWithArgs(firstBndlID, firstEventID)...).
			AddRow(fixBundleReferenceRowWithoutAPIIDWithArgs(secondBndlID, secondEventID)...)

		sqlMock.ExpectQuery(selectQueryForEvents).
			WithArgs(firstBndlID, ExpectedLimit, ExpectedOffset, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryForEvents).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", firstEventBndlRefEntity).Return(model.BundleReference{
			BundleID:   str.Ptr(firstBndlID),
			ObjectType: model.BundleEventReference,
			ObjectID:   str.Ptr(firstEventID),
		}, nil)
		convMock.On("FromEntity", secondEventBndlRefEntity).Return(model.BundleReference{
			BundleID:   str.Ptr(secondBndlID),
			ObjectType: model.BundleEventReference,
			ObjectID:   str.Ptr(secondEventID),
		}, nil)
		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		modelBndlRefs, totalCounts, err := pgRepository.ListByBundleIDs(ctx, model.BundleEventReference, bundleIDs, inputPageSize, inputCursor)
		// THEN
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

	t.Run("returns both public and internal/private APIs when check for internal scope fails", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1

		totalCountForFirstBundle := 1
		totalCountForSecondBundle := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleReferenceColumns()).
			AddRow(fixBundleReferenceRowWithoutEventIDWithArgs(firstBndlID, firstAPIID, firstTargetURL)...).
			AddRow(fixBundleReferenceRowWithoutEventIDWithArgs(secondBndlID, secondAPIID, secondTargetURL)...)

		sqlMock.ExpectQuery(selectQueryForAPIs).
			WithArgs(firstBndlID, ExpectedLimit, ExpectedOffset, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryForAPIs).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", firstAPIBndlRefEntity).Return(model.BundleReference{
			BundleID:            str.Ptr(firstBndlID),
			ObjectType:          model.BundleAPIReference,
			ObjectID:            str.Ptr(firstAPIID),
			APIDefaultTargetURL: str.Ptr(firstTargetURL),
		}, nil)
		convMock.On("FromEntity", secondAPIBndlRefEntity).Return(model.BundleReference{
			BundleID:            str.Ptr(secondBndlID),
			ObjectType:          model.BundleAPIReference,
			ObjectID:            str.Ptr(secondAPIID),
			APIDefaultTargetURL: str.Ptr(secondTargetURL),
		}, nil)
		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		modelBndlRefs, totalCounts, err := pgRepository.ListByBundleIDs(ctx, model.BundleAPIReference, bundleIDs, inputPageSize, inputCursor)
		// THEN
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

		sqlMock.ExpectQuery(selectQueryForAPIs).
			WithArgs(firstBndlID, ExpectedLimit, ExpectedOffset, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryForAPIs).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		ctx = scope.SaveToContext(ctx, scopesWithInternalVisibility)
		convMock := &automock.BundleReferenceConverter{}
		convMock.On("FromEntity", firstAPIBndlRefEntity).Return(model.BundleReference{}, testErr)
		pgRepository := bundlereferences.NewRepository(convMock)
		// WHEN
		_, _, err := pgRepository.ListByBundleIDs(ctx, model.BundleAPIReference, bundleIDs, inputPageSize, inputCursor)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		inputPageSize := 1
		ExpectedLimit := 1
		ExpectedOffset := 0

		pgRepository := bundlereferences.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQueryForAPIs).
			WithArgs(firstBndlID, ExpectedLimit, ExpectedOffset, secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		ctx = scope.SaveToContext(ctx, scopesWithInternalVisibility)

		// WHEN
		modelBndlRefs, totalCounts, err := pgRepository.ListByBundleIDs(ctx, model.BundleAPIReference, bundleIDs, inputPageSize, inputCursor)

		// THEN
		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelBndlRefs)
		assert.Nil(t, totalCounts)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
