package fetchrequest_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		timestamp := time.Now()
		frModel := fixFullFetchRequestModel(givenID(), timestamp)
		frEntity := fixFullFetchRequestEntity(t, givenID(), timestamp)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", frModel).Return(frEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta("INSERT INTO public.fetch_requests ( id, tenant_id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )")).
			WithArgs(givenID(), givenTenant(), "documentID", "foo.bar", frEntity.Auth, frEntity.Mode, frEntity.Filter, frEntity.StatusCondition, frEntity.StatusMessage, frEntity.StatusTimestamp, sql.NullString{}).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := fetchrequest.NewRepository(mockConverter)
		// WHEN
		err := repo.Create(ctx, &frModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error - DB", func(t *testing.T) {
		// GIVEN
		timestamp := time.Now()
		frModel := fixFullFetchRequestModel(givenID(), timestamp)
		frEntity := fixFullFetchRequestEntity(t, givenID(), timestamp)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", frModel).Return(frEntity, nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("INSERT INTO .*").WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := fetchrequest.NewRepository(mockConverter)
		// WHEN
		err := repo.Create(ctx, &frModel)
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("Unexpected error while executing SQL query").Error())
	})

	t.Run("Error - Converter", func(t *testing.T) {
		// GIVEN
		timestamp := time.Now()
		frModel := fixFullFetchRequestModel(givenID(), timestamp)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", frModel).Return(fetchrequest.Entity{}, givenError())

		repo := fetchrequest.NewRepository(mockConverter)
		// WHEN
		err := repo.Create(context.TODO(), &frModel)
		// THEN
		require.EqualError(t, err, "while creating FetchRequest entity from model: some error")
	})
}

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

			query := fmt.Sprintf("SELECT id, tenant_id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id FROM public.fetch_requests WHERE tenant_id = $1 AND %s = $2", testCase.FieldName)
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

		dbMock.ExpectExec(regexp.QuoteMeta("DELETE FROM public.fetch_requests WHERE tenant_id = $1 AND id = $2")).WithArgs(
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

			dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM public.fetch_requests WHERE tenant_id = $1 AND %s = $2", testCase.FieldName))).WithArgs(
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

func givenID() string {
	return "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}

func givenTenant() string {
	return "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
}

func givenError() error {
	return errors.New("some error")
}
