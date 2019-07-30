package labeldef_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCreateLabelDefinition(t *testing.T) {
	// GIVEN
	labelDefID := "d048f47b-b700-49ed-913d-180c3748164b"
	tenantID := "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
	someString := "any"
	var someSchema interface{}
	someSchema = someString

	in := model.LabelDefinition{
		ID:     labelDefID,
		Tenant: tenantID,
		Key:    "some-key",
		Schema: &someSchema,
	}

	escapedQuery := regexp.QuoteMeta("insert into public.label_definitions (id, tenant_id, key, schema) values(?, ?, ?, ?)")

	t.Run("success", func(t *testing.T) {
		db, dbMock := mockDatabase(t)
		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: "any"}, nil)

		dbMock.ExpectExec(escapedQuery).WithArgs(labelDefID, tenantID, "some-key", "any").WillReturnResult(sqlmock.NewResult(1, 1))
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()
		sut := labeldef.NewRepository(mockConverter)
		// WHEN
		err := sut.Create(ctx, in)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		sut := labeldef.NewRepository(nil)
		ctx := context.TODO()
		err := sut.Create(ctx, model.LabelDefinition{})
		require.EqualError(t, err, "unable to fetch database from context")
	})

	t.Run("returns error if insert fails", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: "any"}, nil)

		sut := labeldef.NewRepository(mockConverter)
		db, dbMock := mockDatabase(t)
		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		dbMock.ExpectExec(escapedQuery).WillReturnError(errors.New("some error"))
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()
		// WHEN
		err := sut.Create(ctx, in)
		// THEN
		require.EqualError(t, err, "while inserting Label Definition: some error")
	})

}

func mockDatabase(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {

	sqlDB, sqlMock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(sqlDB, "sqlmock")

	return sqlxDB, sqlMock
}
