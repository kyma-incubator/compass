package fetchrequest_test

import (
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestRepository_Create(t *testing.T) {
	timestamp := time.Now()
	var nilFrModel *model.FetchRequest
	apiFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.APISpecFetchRequestReference)
	apiFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.APISpecFetchRequestReference)
	eventFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.EventSpecFetchRequestReference)
	eventFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.EventSpecFetchRequestReference)
	docFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.DocumentFetchRequestReference)
	docFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.DocumentFetchRequestReference)

	apiFRSuite := testdb.RepoCreateTestSuite{
		Name: "Create API FR",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM api_specifications_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, refID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       regexp.QuoteMeta("INSERT INTO public.fetch_requests ( id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{givenID(), sql.NullString{}, "foo.bar", apiFREntity.Auth, apiFREntity.Mode, apiFREntity.Filter, apiFREntity.StatusCondition, apiFREntity.StatusMessage, apiFREntity.StatusTimestamp, refID},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		ModelEntity:         &apiFRModel,
		DBEntity:            &apiFREntity,
		NilModelEntity:      nilFrModel,
		TenantID:            tenantID,
	}

	eventFRSuite := testdb.RepoCreateTestSuite{
		Name: "Create Event FR",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM event_specifications_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, refID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       regexp.QuoteMeta("INSERT INTO public.fetch_requests ( id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{givenID(), sql.NullString{}, "foo.bar", eventFREntity.Auth, eventFREntity.Mode, eventFREntity.Filter, eventFREntity.StatusCondition, eventFREntity.StatusMessage, eventFREntity.StatusTimestamp, refID},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		ModelEntity:         &eventFRModel,
		DBEntity:            &eventFREntity,
		NilModelEntity:      nilFrModel,
		TenantID:            tenantID,
	}

	docFRSuite := testdb.RepoCreateTestSuite{
		Name: "Create Doc FR",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM documents_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, refID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       regexp.QuoteMeta("INSERT INTO public.fetch_requests ( id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{givenID(), refID, "foo.bar", docFREntity.Auth, docFREntity.Mode, docFREntity.Filter, docFREntity.StatusCondition, docFREntity.StatusMessage, docFREntity.StatusTimestamp, sql.NullString{}},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		ModelEntity:         &docFRModel,
		DBEntity:            &docFREntity,
		NilModelEntity:      nilFrModel,
		TenantID:            tenantID,
	}

	apiFRSuite.Run(t)
	eventFRSuite.Run(t)
	docFRSuite.Run(t)
}
/*
func TestRepository_GetByReferenceObjectID(t *testing.T) {
	refID := "foo"
	testCases := []struct {
		Name       string
		FieldName  string
		ObjectType model.FetchRequestReferenceObjectType
		DocumentID sql.NullString
		SpecID     sql.NullString
	}{
		{Name: "Document", FieldName: "document_id", ObjectType: model.DocumentFetchRequestReference, DocumentID: sql.NullString{String: refID, Valid: true}},
		{Name: "Spec", FieldName: "spec_id", ObjectType: model.SpecFetchRequestReference, SpecID: sql.NullString{String: refID, Valid: true}},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Success - %s", testCase.Name), func(t *testing.T) {
			// GIVEN
			timestamp := time.Now()
			frModel := fixFetchRequestModelWithReference(givenID(), timestamp, testCase.ObjectType, refID)
			frEntity := fixFetchRequestEntityWithReferences(givenID(), timestamp, testCase.SpecID, testCase.DocumentID)

			mockConverter := &automock.Converter{}
			mockConverter.On("FromEntity", frEntity).Return(frModel, nil).Once()

			repo := fetchrequest.NewRepository(mockConverter)
			db, dbMock := testdb.MockDatabase(t)

			rows := sqlmock.NewRows([]string{"id", "tenant_id", "document_id", "url", "auth", "mode", "filter", "status_condition", "status_message", "status_timestamp", "spec_id"}).
				AddRow(givenID(), givenTenant(), testCase.DocumentID, "foo.bar", frEntity.Auth, frEntity.Mode, frEntity.Filter, frEntity.StatusCondition, frEntity.StatusMessage, frEntity.StatusTimestamp, testCase.SpecID)

			query := fmt.Sprintf("SELECT id, tenant_id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id FROM public.fetch_requests WHERE %s AND %s = $2", fixUnescapedTenantIsolationSubquery(), testCase.FieldName)
			dbMock.ExpectQuery(regexp.QuoteMeta(query)).
				WithArgs(givenTenant(), givenID()).WillReturnRows(rows)

			ctx := persistence.SaveToContext(context.TODO(), db)
			// WHEN
			actual, err := repo.GetByReferenceObjectID(ctx, givenTenant(), testCase.ObjectType, givenID())
			// THEN
			require.NoError(t, err)
			require.NotNil(t, actual)
			assert.Equal(t, frModel, *actual)

			mockConverter.AssertExpectations(t)
			dbMock.AssertExpectations(t)
		})
	}

	t.Run("Error - Converter", func(t *testing.T) {
		// GIVEN
		timestamp := time.Now()
		frEntity := fixFullFetchRequestEntity(t, givenID(), timestamp)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", frEntity).Return(model.FetchRequest{}, givenError())

		repo := fetchrequest.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "document_id", "url", "auth", "mode", "filter", "status_condition", "status_message", "status_timestamp", "spec_id"}).
			AddRow(givenID(), givenTenant(), "documentID", "foo.bar", frEntity.Auth, frEntity.Mode, frEntity.Filter, frEntity.StatusCondition, frEntity.StatusMessage, frEntity.StatusTimestamp, sql.NullString{})

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID()).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := repo.GetByReferenceObjectID(ctx, givenTenant(), model.SpecFetchRequestReference, givenID())
		// THEN
		require.EqualError(t, err, "while getting FetchRequest model from entity: some error")
	})

	t.Run("Error - DB", func(t *testing.T) {
		// GIVEN
		repo := fetchrequest.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := repo.GetByReferenceObjectID(ctx, givenTenant(), model.DocumentFetchRequestReference, givenID())
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Error - Invalid Object Reference Type", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := fetchrequest.NewRepository(nil)
		// WHEN
		_, err := repo.GetByReferenceObjectID(ctx, givenTenant(), "test", givenID())
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("Invalid type of the Fetch Request reference object").Error())
	})
}

func TestRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM public.fetch_requests WHERE %s AND id = $2", fixUnescapedTenantIsolationSubquery()))).WithArgs(
			givenTenant(), givenID()).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := fetchrequest.NewRepository(nil)
		// WHEN
		err := repo.Delete(ctx, givenTenant(), givenID())
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error - DB", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("DELETE FROM .*").WithArgs(
			givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := fetchrequest.NewRepository(nil)
		// WHEN
		err := repo.Delete(ctx, givenTenant(), givenID())
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_DeleteByReferenceObjectID(t *testing.T) {
	refID := "foo"
	testCases := []struct {
		Name       string
		FieldName  string
		ObjectType model.FetchRequestReferenceObjectType
		DocumentID sql.NullString
		SpecID     sql.NullString
	}{
		{Name: "Document", FieldName: "document_id", ObjectType: model.DocumentFetchRequestReference, DocumentID: sql.NullString{String: refID, Valid: true}},
		{Name: "Spec", FieldName: "spec_id", ObjectType: model.SpecFetchRequestReference, SpecID: sql.NullString{String: refID, Valid: true}},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Success - %s", testCase.Name), func(t *testing.T) {
			// GIVEN
			db, dbMock := testdb.MockDatabase(t)
			defer dbMock.AssertExpectations(t)

			dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM public.fetch_requests WHERE %s AND %s = $2", fixUnescapedTenantIsolationSubquery(), testCase.FieldName))).WithArgs(
				givenTenant(), givenID()).WillReturnResult(sqlmock.NewResult(-1, 1))

			ctx := persistence.SaveToContext(context.TODO(), db)
			repo := fetchrequest.NewRepository(nil)
			// WHEN
			err := repo.DeleteByReferenceObjectID(ctx, givenTenant(), testCase.ObjectType, givenID())
			// THEN
			require.NoError(t, err)
		})
	}

	t.Run("Error - Invalid Object Reference Type", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := fetchrequest.NewRepository(nil)
		// WHEN
		err := repo.DeleteByReferenceObjectID(ctx, givenTenant(), "test", givenID())
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("Invalid type of the Fetch Request reference object").Error())
	})

	t.Run("Error - DB", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("DELETE FROM .*").WithArgs(
			givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := fetchrequest.NewRepository(nil)
		// WHEN
		err := repo.DeleteByReferenceObjectID(ctx, givenTenant(), model.SpecFetchRequestReference, givenID())
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_ListByReferenceObjectIDs(t *testing.T) {
	timestamp := time.Now()
	firstSpecID := "111111111-1111-1111-1111-111111111111"
	secondSpecID := "222222222-2222-2222-2222-222222222222"
	firstFRID := "333333333-3333-3333-3333-333333333333"
	secondFRID := "444444444-4444-4444-4444-444444444444"
	firstDocID := "333333333-3333-3333-3333-333333333333"
	secondDocID := "444444444-4444-4444-4444-444444444444"
	specIDs := []string{firstSpecID, secondSpecID}
	docIDs := []string{firstDocID, secondDocID}

	firstSpecFREntity := fixFetchRequestEntityWithReferences(firstFRID, timestamp, sql.NullString{
		String: firstSpecID,
		Valid:  true,
	}, sql.NullString{})

	secondSpecFREntity := fixFetchRequestEntityWithReferences(secondFRID, timestamp, sql.NullString{
		String: secondSpecID,
		Valid:  true,
	}, sql.NullString{})

	firstDocFREntity := fixFetchRequestEntityWithReferences(firstFRID, timestamp, sql.NullString{}, sql.NullString{
		String: firstDocID,
		Valid:  true,
	})

	secondDocFREntity := fixFetchRequestEntityWithReferences(secondFRID, timestamp, sql.NullString{}, sql.NullString{
		String: secondDocID,
		Valid:  true,
	})

	selectQuerySpecs := fmt.Sprintf(`^SELECT (.+) FROM public\.fetch_requests WHERE %s AND spec_id IN \(\$2, \$3\)`, fixTenantIsolationSubquery())

	selectQueryDocs := fmt.Sprintf(`^SELECT (.+) FROM public\.fetch_requests WHERE %s AND document_id IN \(\$2, \$3\)`, fixTenantIsolationSubquery())

	t.Run("Success for Specs", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", firstSpecFREntity).Return(fixFetchRequestModelWithReference(firstFRID, timestamp, model.SpecFetchRequestReference, firstSpecID), nil).Once()
		mockConverter.On("FromEntity", secondSpecFREntity).Return(fixFetchRequestModelWithReference(secondFRID, timestamp, model.SpecFetchRequestReference, secondSpecID), nil).Once()

		repo := fetchrequest.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "document_id", "url", "auth", "mode", "filter", "status_condition", "status_message", "status_timestamp", "spec_id"}).
			AddRow(firstFRID, firstSpecFREntity.TenantID, firstSpecFREntity.DocumentID, firstSpecFREntity.URL, firstSpecFREntity.Auth, firstSpecFREntity.Mode, firstSpecFREntity.Filter, firstSpecFREntity.StatusCondition, firstSpecFREntity.StatusMessage, firstSpecFREntity.StatusTimestamp, firstSpecFREntity.SpecID).
			AddRow(secondFRID, secondSpecFREntity.TenantID, secondSpecFREntity.DocumentID, secondSpecFREntity.URL, secondSpecFREntity.Auth, secondSpecFREntity.Mode, secondSpecFREntity.Filter, secondSpecFREntity.StatusCondition, secondSpecFREntity.StatusMessage, secondSpecFREntity.StatusTimestamp, secondSpecFREntity.SpecID)

		dbMock.ExpectQuery(selectQuerySpecs).
			WithArgs(givenTenant(), firstSpecID, secondSpecID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		frModels, err := repo.ListByReferenceObjectIDs(ctx, givenTenant(), model.SpecFetchRequestReference, specIDs)
		require.NoError(t, err)
		require.Len(t, frModels, 2)
		assert.Equal(t, firstFRID, frModels[0].ID)
		assert.Equal(t, secondFRID, frModels[1].ID)
		assert.Equal(t, firstSpecID, frModels[0].ObjectID)
		assert.Equal(t, secondSpecID, frModels[1].ObjectID)
		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Success for Docs", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", firstDocFREntity).Return(fixFetchRequestModelWithReference(firstFRID, timestamp, model.DocumentFetchRequestReference, firstDocID), nil).Once()
		mockConverter.On("FromEntity", secondDocFREntity).Return(fixFetchRequestModelWithReference(secondFRID, timestamp, model.DocumentFetchRequestReference, secondDocID), nil).Once()

		repo := fetchrequest.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "document_id", "url", "auth", "mode", "filter", "status_condition", "status_message", "status_timestamp", "spec_id"}).
			AddRow(firstFRID, firstDocFREntity.TenantID, firstDocFREntity.DocumentID, firstDocFREntity.URL, firstDocFREntity.Auth, firstDocFREntity.Mode, firstDocFREntity.Filter, firstDocFREntity.StatusCondition, firstDocFREntity.StatusMessage, firstDocFREntity.StatusTimestamp, firstDocFREntity.SpecID).
			AddRow(secondFRID, secondDocFREntity.TenantID, secondDocFREntity.DocumentID, secondDocFREntity.URL, secondDocFREntity.Auth, secondDocFREntity.Mode, secondDocFREntity.Filter, secondDocFREntity.StatusCondition, secondDocFREntity.StatusMessage, secondDocFREntity.StatusTimestamp, secondDocFREntity.SpecID)

		dbMock.ExpectQuery(selectQueryDocs).
			WithArgs(givenTenant(), firstDocID, secondDocID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		frModels, err := repo.ListByReferenceObjectIDs(ctx, givenTenant(), model.DocumentFetchRequestReference, docIDs)
		require.NoError(t, err)
		require.Len(t, frModels, 2)
		assert.Equal(t, firstFRID, frModels[0].ID)
		assert.Equal(t, secondFRID, frModels[1].ID)
		assert.Equal(t, firstDocID, frModels[0].ObjectID)
		assert.Equal(t, secondDocID, frModels[1].ObjectID)
		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Returns error when conversion from entity fails", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", firstDocFREntity).Return(model.FetchRequest{}, givenError()).Once()

		repo := fetchrequest.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "document_id", "url", "auth", "mode", "filter", "status_condition", "status_message", "status_timestamp", "spec_id"}).
			AddRow(firstFRID, firstDocFREntity.TenantID, firstDocFREntity.DocumentID, firstDocFREntity.URL, firstDocFREntity.Auth, firstDocFREntity.Mode, firstDocFREntity.Filter, firstDocFREntity.StatusCondition, firstDocFREntity.StatusMessage, firstDocFREntity.StatusTimestamp, firstDocFREntity.SpecID).
			AddRow(secondFRID, secondDocFREntity.TenantID, secondDocFREntity.DocumentID, secondDocFREntity.URL, secondDocFREntity.Auth, secondDocFREntity.Mode, secondDocFREntity.Filter, secondDocFREntity.StatusCondition, secondDocFREntity.StatusMessage, secondDocFREntity.StatusTimestamp, secondDocFREntity.SpecID)

		dbMock.ExpectQuery(selectQueryDocs).
			WithArgs(givenTenant(), firstDocID, secondDocID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := repo.ListByReferenceObjectIDs(ctx, givenTenant(), model.DocumentFetchRequestReference, docIDs)
		require.Error(t, err)
		require.Contains(t, err.Error(), givenError().Error())
		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error - Invalid Object Reference Type", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := fetchrequest.NewRepository(nil)
		// WHEN
		_, err := repo.ListByReferenceObjectIDs(ctx, givenTenant(), "invalidObjectType", docIDs)
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("Invalid type of the Fetch Request reference object").Error())
	})

	t.Run("Error - DB", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery(selectQueryDocs).
			WithArgs(givenTenant(), firstDocID, secondDocID).
			WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := fetchrequest.NewRepository(nil)
		// WHEN
		_, err := repo.ListByReferenceObjectIDs(ctx, givenTenant(), model.DocumentFetchRequestReference, docIDs)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
*/

func givenID() string {
	return "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}