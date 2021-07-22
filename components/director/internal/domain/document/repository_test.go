package document_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const insertQuery = "INSERT INTO public.documents ( id, tenant_id, bundle_id, title, display_name, description, format, kind, data, ready, created_at, updated_at, deleted_at, error ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"

var columns = []string{"id", "tenant_id", "bundle_id", "title", "display_name", "description", "format", "kind", "data", "ready", "created_at", "updated_at", "deleted_at", "error"}

func TestRepository_Create(t *testing.T) {
	refID := bndlID()
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		docModel := fixModelDocument(givenID(), refID)
		docEntity := fixEntityDocument(givenID(), refID)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", *docModel).Return(docEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(insertQuery)).
			WithArgs(givenID(), givenTenant(), refID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data,
				docEntity.Ready, docEntity.CreatedAt, docEntity.UpdatedAt, docEntity.DeletedAt, docEntity.Error).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(mockConverter)
		// WHEN
		err := repo.Create(ctx, docModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		docModel := fixModelDocument(givenID(), refID)
		docEntity := fixEntityDocument(givenID(), refID)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", *docModel).Return(docEntity, nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("INSERT INTO .*").WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(mockConverter)
		// WHEN
		err := repo.Create(ctx, docModel)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Converter Error", func(t *testing.T) {
		// GIVEN
		docModel := fixModelDocument(givenID(), refID)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", *docModel).Return(nil, givenError())

		repo := document.NewRepository(mockConverter)
		// WHEN
		err := repo.Create(context.TODO(), docModel)
		// THEN
		require.EqualError(t, err, "while creating Document entity from model: some error")
	})
}

func TestRepository_CreateMany(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		conv := &automock.Converter{}
		defer conv.AssertExpectations(t)

		given := []*model.Document{
			fixModelDocument("1", bndlID()),
			fixModelDocument("2", bndlID()),
			fixModelDocument("3", bndlID()),
		}
		expected := []*document.Entity{
			fixEntityDocument("1", bndlID()),
			fixEntityDocument("2", bndlID()),
			fixEntityDocument("3", bndlID()),
		}

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		for i, givenModel := range given {
			expectedEntity := expected[i]
			conv.On("ToEntity", *givenModel).Return(expectedEntity, nil).Once()
			dbMock.ExpectExec(regexp.QuoteMeta(insertQuery)).WithArgs(
				expectedEntity.ID, expectedEntity.TenantID, expectedEntity.BndlID, expectedEntity.Title, expectedEntity.DisplayName, expectedEntity.Description, expectedEntity.Format, expectedEntity.Kind, expectedEntity.Data,
				expectedEntity.Ready, expectedEntity.CreatedAt, expectedEntity.UpdatedAt, expectedEntity.DeletedAt, expectedEntity.Error).WillReturnResult(sqlmock.NewResult(-1, 1))
		}

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(conv)
		// WHEN
		err := repo.CreateMany(ctx, given)
		// THEN
		require.NoError(t, err)
	})

	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		conv := &automock.Converter{}
		defer conv.AssertExpectations(t)

		given := []*model.Document{
			fixModelDocument("1", bndlID()),
			fixModelDocument("2", bndlID()),
			fixModelDocument("3", bndlID()),
		}
		expected := []*document.Entity{
			fixEntityDocument("1", bndlID()),
			fixEntityDocument("2", bndlID()),
		}

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		conv.On("ToEntity", *given[0]).Return(expected[0], nil).Once()
		conv.On("ToEntity", *given[1]).Return(expected[1], nil).Once()
		dbMock.ExpectExec(regexp.QuoteMeta(insertQuery)).WithArgs(
			expected[0].ID, expected[0].TenantID, expected[0].BndlID, expected[0].Title, expected[0].DisplayName, expected[0].Description, expected[0].Format, expected[0].Kind, expected[0].Data,
			expected[0].Ready, expected[0].CreatedAt, expected[0].UpdatedAt, expected[0].DeletedAt, expected[0].Error).WillReturnResult(sqlmock.NewResult(-1, 1))
		dbMock.ExpectExec(regexp.QuoteMeta(insertQuery)).WithArgs(
			expected[1].ID, expected[1].TenantID, expected[1].BndlID, expected[1].Title, expected[1].DisplayName, expected[1].Description, expected[1].Format, expected[1].Kind, expected[1].Data,
			expected[1].Ready, expected[1].CreatedAt, expected[1].UpdatedAt, expected[1].DeletedAt, expected[1].Error).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(conv)

		// WHEN
		err := repo.CreateMany(ctx, given)
		// THEN
		require.EqualError(t, err, "while creating Document with ID 2: Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Converter Error", func(t *testing.T) {
		// GIVEN
		conv := &automock.Converter{}
		defer conv.AssertExpectations(t)

		given := []*model.Document{
			fixModelDocument("1", bndlID()),
			fixModelDocument("2", bndlID()),
			fixModelDocument("3", bndlID()),
		}
		expected := []*document.Entity{
			fixEntityDocument("1", bndlID()),
			fixEntityDocument("2", bndlID()),
		}

		db, dbMock := testdb.MockDatabase(t)
		//defer dbMock.AssertExpectations(t)

		conv.On("ToEntity", *given[0]).Return(expected[0], nil).Once()
		conv.On("ToEntity", *given[1]).Return(nil, givenError()).Once()
		dbMock.ExpectExec(regexp.QuoteMeta(insertQuery)).WithArgs(
			expected[0].ID, expected[0].TenantID, expected[0].BndlID, expected[0].Title, expected[0].DisplayName, expected[0].Description, expected[0].Format, expected[0].Kind, expected[0].Data,
			expected[0].Ready, expected[0].CreatedAt, expected[0].UpdatedAt, expected[0].DeletedAt, expected[0].Error).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(conv)

		// WHEN
		err := repo.CreateMany(ctx, given)
		// THEN
		require.EqualError(t, err, "while creating Document with ID 2: while creating Document entity from model: some error")
	})
}


func TestRepository_ListAllForBundle(t *testing.T) {
	// GIVEN
	inputCursor := ""

	firstBndlID := "111111111-1111-1111-1111-111111111111"
	secondBndlID := "222222222-2222-2222-2222-222222222222"
	bundleIDs := []string{firstBndlID, secondBndlID}

	firstDocID := "111111111-1111-1111-1111-111111111111"
	firstDocEntity := fixEntityDocument(firstDocID, firstBndlID)
	secondDocID := "222222222-2222-2222-2222-222222222222"
	secondDocEntity := fixEntityDocument(secondDocID, secondBndlID)

	selectQuery := fmt.Sprintf(`\(SELECT (.+) FROM public\.documents
		WHERE %s AND bundle_id = \$2 ORDER BY bundle_id ASC, id ASC LIMIT \$3 OFFSET \$4\) UNION
		\(SELECT (.+) FROM public\.documents WHERE %s AND bundle_id = \$6 ORDER BY bundle_id ASC, id ASC LIMIT \$7 OFFSET \$8\)`, fixTenantIsolationSubqueryWithArg(1), fixTenantIsolationSubqueryWithArg(5))

	countQuery := fmt.Sprintf(`SELECT bundle_id AS id, COUNT\(\*\) AS total_count FROM public.documents WHERE %s GROUP BY bundle_id ORDER BY bundle_id ASC`, fixTenantIsolationSubquery())

	t.Run("success when there are no more pages", func(t *testing.T) {
		ExpectedLimit := 3
		ExpectedOffset := 0
		inputPageSize := 3
		totalCountForFirstBundle := 1
		totalCountForSecondBundle := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(columns).
			AddRow(firstDocEntity.ID, firstDocEntity.TenantID, firstDocEntity.BndlID, firstDocEntity.Title, firstDocEntity.DisplayName, firstDocEntity.Description, firstDocEntity.Format, firstDocEntity.Kind, firstDocEntity.Data,
				firstDocEntity.Ready, firstDocEntity.CreatedAt, firstDocEntity.UpdatedAt, firstDocEntity.DeletedAt, firstDocEntity.Error).
			AddRow(secondDocEntity.ID, secondDocEntity.TenantID, secondDocEntity.BndlID, secondDocEntity.Title, secondDocEntity.DisplayName, secondDocEntity.Description, secondDocEntity.Format, secondDocEntity.Kind, secondDocEntity.Data,
				secondDocEntity.Ready, secondDocEntity.CreatedAt, secondDocEntity.UpdatedAt, secondDocEntity.DeletedAt, secondDocEntity.Error)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(givenTenant(), firstBndlID, ExpectedLimit, ExpectedOffset, givenTenant(), secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(givenTenant()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", *firstDocEntity).Return(model.Document{BaseEntity: &model.BaseEntity{ID: firstDocEntity.ID}}, nil).Once()
		convMock.On("FromEntity", *secondDocEntity).Return(model.Document{BaseEntity: &model.BaseEntity{ID: secondDocEntity.ID}}, nil).Once()
		pgRepository := document.NewRepository(convMock)
		// WHEN
		modelDocs, err := pgRepository.ListAllForBundle(ctx, givenTenant(), bundleIDs, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelDocs, 2)
		assert.Equal(t, firstDocID, modelDocs[0].Data[0].ID)
		assert.Equal(t, secondDocID, modelDocs[1].Data[0].ID)
		assert.Equal(t, "", modelDocs[0].PageInfo.StartCursor)
		assert.Equal(t, totalCountForFirstBundle, modelDocs[0].TotalCount)
		assert.False(t, modelDocs[0].PageInfo.HasNextPage)
		assert.Equal(t, "", modelDocs[1].PageInfo.StartCursor)
		assert.Equal(t, totalCountForSecondBundle, modelDocs[1].TotalCount)
		assert.False(t, modelDocs[1].PageInfo.HasNextPage)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("success when there is next page", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1
		totalCountForFirstBundle := 10
		totalCountForSecondBundle := 10

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(columns).
			AddRow(firstDocEntity.ID, firstDocEntity.TenantID, firstDocEntity.BndlID, firstDocEntity.Title, firstDocEntity.DisplayName, firstDocEntity.Description, firstDocEntity.Format, firstDocEntity.Kind, firstDocEntity.Data,
				firstDocEntity.Ready, firstDocEntity.CreatedAt, firstDocEntity.UpdatedAt, firstDocEntity.DeletedAt, firstDocEntity.Error).
			AddRow(secondDocEntity.ID, secondDocEntity.TenantID, secondDocEntity.BndlID, secondDocEntity.Title, secondDocEntity.DisplayName, secondDocEntity.Description, secondDocEntity.Format, secondDocEntity.Kind, secondDocEntity.Data,
				secondDocEntity.Ready, secondDocEntity.CreatedAt, secondDocEntity.UpdatedAt, secondDocEntity.DeletedAt, secondDocEntity.Error)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(givenTenant(), firstBndlID, ExpectedLimit, ExpectedOffset, givenTenant(), secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(givenTenant()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", *firstDocEntity).Return(model.Document{BaseEntity: &model.BaseEntity{ID: firstDocEntity.ID}}, nil).Once()
		convMock.On("FromEntity", *secondDocEntity).Return(model.Document{BaseEntity: &model.BaseEntity{ID: secondDocEntity.ID}}, nil).Once()
		pgRepository := document.NewRepository(convMock)
		// WHEN
		modelDocs, err := pgRepository.ListAllForBundle(ctx, givenTenant(), bundleIDs, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelDocs, 2)
		assert.Equal(t, firstDocID, modelDocs[0].Data[0].ID)
		assert.Equal(t, secondDocID, modelDocs[1].Data[0].ID)
		assert.Equal(t, "", modelDocs[0].PageInfo.StartCursor)
		assert.Equal(t, totalCountForFirstBundle, modelDocs[0].TotalCount)
		assert.True(t, modelDocs[0].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelDocs[0].PageInfo.EndCursor)
		assert.Equal(t, "", modelDocs[1].PageInfo.StartCursor)
		assert.Equal(t, totalCountForSecondBundle, modelDocs[1].TotalCount)
		assert.True(t, modelDocs[1].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelDocs[1].PageInfo.EndCursor)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("success when there is next page and it can be traversed", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		ExpectedSecondOffset := 1
		inputPageSize := 1
		totalCountForFirstBundle := 2
		totalCountForSecondBundle := 2

		thirdDocID := "333333333-3333-3333-3333-333333333333"
		thirdDocEntity := fixEntityDocument(thirdDocID, firstBndlID)
		fourthDocID := "444444444-4444-4444-4444-444444444444"
		fourthDocEntity := fixEntityDocument(fourthDocID, secondBndlID)

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(columns).
			AddRow(firstDocEntity.ID, firstDocEntity.TenantID, firstDocEntity.BndlID, firstDocEntity.Title, firstDocEntity.DisplayName, firstDocEntity.Description, firstDocEntity.Format, firstDocEntity.Kind, firstDocEntity.Data,
				firstDocEntity.Ready, firstDocEntity.CreatedAt, firstDocEntity.UpdatedAt, firstDocEntity.DeletedAt, firstDocEntity.Error).
			AddRow(secondDocEntity.ID, secondDocEntity.TenantID, secondDocEntity.BndlID, secondDocEntity.Title, secondDocEntity.DisplayName, secondDocEntity.Description, secondDocEntity.Format, secondDocEntity.Kind, secondDocEntity.Data,
				secondDocEntity.Ready, secondDocEntity.CreatedAt, secondDocEntity.UpdatedAt, secondDocEntity.DeletedAt, secondDocEntity.Error)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(givenTenant(), firstBndlID, ExpectedLimit, ExpectedOffset, givenTenant(), secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(givenTenant()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		rowsSecondPage := sqlmock.NewRows(columns).
			AddRow(thirdDocEntity.ID, thirdDocEntity.TenantID, thirdDocEntity.BndlID, thirdDocEntity.Title, thirdDocEntity.DisplayName, thirdDocEntity.Description, thirdDocEntity.Format, thirdDocEntity.Kind, thirdDocEntity.Data,
				thirdDocEntity.Ready, thirdDocEntity.CreatedAt, thirdDocEntity.UpdatedAt, thirdDocEntity.DeletedAt, thirdDocEntity.Error).
			AddRow(fourthDocEntity.ID, fourthDocEntity.TenantID, fourthDocEntity.BndlID, fourthDocEntity.Title, fourthDocEntity.DisplayName, fourthDocEntity.Description, fourthDocEntity.Format, fourthDocEntity.Kind, fourthDocEntity.Data,
				fourthDocEntity.Ready, fourthDocEntity.CreatedAt, fourthDocEntity.UpdatedAt, fourthDocEntity.DeletedAt, fourthDocEntity.Error)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(givenTenant(), firstBndlID, ExpectedLimit, ExpectedSecondOffset, givenTenant(), secondBndlID, ExpectedLimit, ExpectedSecondOffset).
			WillReturnRows(rowsSecondPage)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(givenTenant()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", *firstDocEntity).Return(model.Document{BaseEntity: &model.BaseEntity{ID: firstDocEntity.ID}}, nil).Once()
		convMock.On("FromEntity", *secondDocEntity).Return(model.Document{BaseEntity: &model.BaseEntity{ID: secondDocEntity.ID}}, nil).Once()
		convMock.On("FromEntity", *thirdDocEntity).Return(model.Document{BaseEntity: &model.BaseEntity{ID: thirdDocEntity.ID}}, nil).Once()
		convMock.On("FromEntity", *fourthDocEntity).Return(model.Document{BaseEntity: &model.BaseEntity{ID: fourthDocEntity.ID}}, nil).Once()
		pgRepository := document.NewRepository(convMock)
		// WHEN
		modelDocs, err := pgRepository.ListAllForBundle(ctx, givenTenant(), bundleIDs, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelDocs, 2)
		assert.Equal(t, firstDocID, modelDocs[0].Data[0].ID)
		assert.Equal(t, secondDocID, modelDocs[1].Data[0].ID)
		assert.Equal(t, "", modelDocs[0].PageInfo.StartCursor)
		assert.Equal(t, totalCountForFirstBundle, modelDocs[0].TotalCount)
		assert.True(t, modelDocs[0].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelDocs[0].PageInfo.EndCursor)
		assert.Equal(t, "", modelDocs[1].PageInfo.StartCursor)
		assert.Equal(t, totalCountForSecondBundle, modelDocs[1].TotalCount)
		assert.True(t, modelDocs[1].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelDocs[1].PageInfo.EndCursor)
		endCursor := modelDocs[0].PageInfo.EndCursor

		modelDocsSecondPage, err := pgRepository.ListAllForBundle(ctx, givenTenant(), bundleIDs, inputPageSize, endCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelDocsSecondPage, 2)
		assert.Equal(t, thirdDocID, modelDocsSecondPage[0].Data[0].ID)
		assert.Equal(t, fourthDocID, modelDocsSecondPage[1].Data[0].ID)
		assert.Equal(t, totalCountForFirstBundle, modelDocsSecondPage[0].TotalCount)
		assert.False(t, modelDocsSecondPage[0].PageInfo.HasNextPage)
		assert.Equal(t, totalCountForSecondBundle, modelDocsSecondPage[1].TotalCount)
		assert.False(t, modelDocsSecondPage[1].PageInfo.HasNextPage)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns empty page", func(t *testing.T) {
		inputPageSize := 1
		totalCountForFirstBundle := 0
		totalCountForSecondBundle := 0

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(columns)

		sqlMock.ExpectQuery(selectQuery).WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(givenTenant()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		pgRepository := document.NewRepository(convMock)
		// WHEN
		modelDocs, err := pgRepository.ListAllForBundle(ctx, givenTenant(), bundleIDs, inputPageSize, inputCursor)
		//THEN

		require.NoError(t, err)
		require.Len(t, modelDocs[0].Data, 0)
		require.Len(t, modelDocs[1].Data, 0)
		assert.Equal(t, totalCountForFirstBundle, modelDocs[0].TotalCount)
		assert.False(t, modelDocs[0].PageInfo.HasNextPage)
		assert.Equal(t, totalCountForSecondBundle, modelDocs[1].TotalCount)
		assert.False(t, modelDocs[1].PageInfo.HasNextPage)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when conversion from entity to model failed", func(t *testing.T) {
		ExpectedLimit := 3
		ExpectedOffset := 0
		inputPageSize := 3
		totalCountForFirstBundle := 1
		totalCountForSecondBundle := 1
		testErr := errors.New("test error")

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(columns).
			AddRow(firstDocEntity.ID, firstDocEntity.TenantID, firstDocEntity.BndlID, firstDocEntity.Title, firstDocEntity.DisplayName, firstDocEntity.Description, firstDocEntity.Format, firstDocEntity.Kind, firstDocEntity.Data,
				firstDocEntity.Ready, firstDocEntity.CreatedAt, firstDocEntity.UpdatedAt, firstDocEntity.DeletedAt, firstDocEntity.Error).
			AddRow(secondDocEntity.ID, secondDocEntity.TenantID, secondDocEntity.BndlID, secondDocEntity.Title, secondDocEntity.DisplayName, secondDocEntity.Description, secondDocEntity.Format, secondDocEntity.Kind, secondDocEntity.Data,
				secondDocEntity.Ready, secondDocEntity.CreatedAt, secondDocEntity.UpdatedAt, secondDocEntity.DeletedAt, secondDocEntity.Error)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(givenTenant(), firstBndlID, ExpectedLimit, ExpectedOffset, givenTenant(), secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(givenTenant()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstBndlID, totalCountForFirstBundle).
				AddRow(secondBndlID, totalCountForSecondBundle))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", *firstDocEntity).Return(model.Document{}, testErr).Once()
		pgRepository := document.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.ListAllForBundle(ctx, givenTenant(), bundleIDs, inputPageSize, inputCursor)
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

		pgRepository := document.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(givenTenant(), firstBndlID, ExpectedLimit, ExpectedOffset, givenTenant(), secondBndlID, ExpectedLimit, ExpectedOffset).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelDocs, err := pgRepository.ListAllForBundle(ctx, givenTenant(), bundleIDs, inputPageSize, inputCursor)

		// then
		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelDocs)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_Exists(t *testing.T) {
	// given
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	sqlMock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT 1 FROM public.documents WHERE %s AND id = $2", fixUnescapedTenantIsolationSubquery()))).WithArgs(
		givenTenant(), givenID()).
		WillReturnRows(testdb.RowWhenObjectExist())

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	repo := document.NewRepository(nil)

	// when
	ex, err := repo.Exists(ctx, givenTenant(), givenID())

	// then
	require.NoError(t, err)
	assert.True(t, ex)
}

func TestRepository_GetByID(t *testing.T) {
	refID := bndlID()

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		docModel := fixModelDocument(givenID(), refID)
		docEntity := fixEntityDocument(givenID(), refID)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", *docEntity).Return(*docModel, nil).Once()

		repo := document.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(columns).
			AddRow(givenID(), givenTenant(), refID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data,
				docEntity.Ready, docEntity.CreatedAt, docEntity.UpdatedAt, docEntity.DeletedAt, docEntity.Error)

		query := fmt.Sprintf("SELECT id, tenant_id, bundle_id, title, display_name, description, format, kind, data, ready, created_at, updated_at, deleted_at, error FROM public.documents WHERE %s AND id = $2", fixUnescapedTenantIsolationSubquery())
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(givenTenant(), givenID()).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		actual, err := repo.GetByID(ctx, givenTenant(), givenID())
		// THEN
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, docModel, actual)

		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Converter Error", func(t *testing.T) {
		// GIVEN
		docEntity := fixEntityDocument(givenID(), refID)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", *docEntity).Return(model.Document{}, givenError())

		repo := document.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows(columns).
			AddRow(givenID(), givenTenant(), refID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data,
				docEntity.Ready, docEntity.CreatedAt, docEntity.UpdatedAt, docEntity.DeletedAt, docEntity.Error)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID()).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := repo.GetByID(ctx, givenTenant(), givenID())
		// THEN
		require.EqualError(t, err, "while converting Document entity to model: some error")
	})

	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		repo := document.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := repo.GetByID(ctx, givenTenant(), givenID())
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_GetForBundle(t *testing.T) {
	refID := bndlID()

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		docModel := fixModelDocument(givenID(), refID)
		docEntity := fixEntityDocument(givenID(), refID)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", *docEntity).Return(*docModel, nil).Once()

		repo := document.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(columns).
			AddRow(givenID(), givenTenant(), refID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data,
				docEntity.Ready, docEntity.CreatedAt, docEntity.UpdatedAt, docEntity.DeletedAt, docEntity.Error)

		query := fmt.Sprintf("SELECT id, tenant_id, bundle_id, title, display_name, description, format, kind, data, ready, created_at, updated_at, deleted_at, error FROM public.documents WHERE %s AND id = $2 AND bundle_id = $3", fixUnescapedTenantIsolationSubquery())
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(givenTenant(), givenID(), bndlID()).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		actual, err := repo.GetForBundle(ctx, givenTenant(), givenID(), bndlID())
		// THEN
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, docModel, actual)

		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Converter Error", func(t *testing.T) {
		// GIVEN
		docEntity := fixEntityDocument(givenID(), refID)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", *docEntity).Return(model.Document{}, givenError())

		repo := document.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows(columns).
			AddRow(givenID(), givenTenant(), refID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data,
				docEntity.Ready, docEntity.CreatedAt, docEntity.UpdatedAt, docEntity.DeletedAt, docEntity.Error)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID(), bndlID()).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := repo.GetForBundle(ctx, givenTenant(), givenID(), bndlID())
		// THEN
		require.EqualError(t, err, "while converting Document entity to model: some error")
	})

	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		repo := document.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID(), bndlID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := repo.GetForBundle(ctx, givenTenant(), givenID(), bndlID())
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM public.documents WHERE %s AND id = $2", fixUnescapedTenantIsolationSubquery()))).WithArgs(
			givenTenant(), givenID()).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(nil)
		// WHEN
		err := repo.Delete(ctx, givenTenant(), givenID())
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("DELETE FROM .*").WithArgs(
			givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(nil)
		// WHEN
		err := repo.Delete(ctx, givenTenant(), givenID())
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func givenID() string {
	return "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}

func bndlID() string {
	return "ppppppppp-pppp-pppp-pppp-pppppppppppp"
}

func givenTenant() string {
	return "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
}

func givenError() error {
	return errors.New("some error")
}
