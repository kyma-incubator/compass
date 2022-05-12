package labeldef_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCreateLabelDefinition(t *testing.T) {
	// GIVEN
	labelDefID := "d048f47b-b700-49ed-913d-180c3748164b"
	tenantID := "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
	someString := "any"
	version := 42
	var someSchema interface{} = someString

	in := model.LabelDefinition{
		ID:      labelDefID,
		Tenant:  tenantID,
		Key:     "some-key",
		Schema:  &someSchema,
		Version: version,
	}

	t.Run("successfully created definition with schema", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}, Version: version}, nil)
		escapedQuery := regexp.QuoteMeta("INSERT INTO public.label_definitions ( id, tenant_id, key, schema, version ) VALUES ( ?, ?, ?, ?, ? )")
		dbMock.ExpectExec(escapedQuery).WithArgs(labelDefID, tenantID, "some-key", "any", version).WillReturnResult(sqlmock.NewResult(1, 1))

		defer dbMock.AssertExpectations(t)
		sut := labeldef.NewRepository(mockConverter)
		// WHEN
		err := sut.Create(ctx, in)
		// THEN
		require.NoError(t, err)
	})

	t.Run("successfully created definition without schema", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{}, Version: version}, nil)
		escapedQuery := regexp.QuoteMeta("INSERT INTO public.label_definitions ( id, tenant_id, key, schema, version ) VALUES ( ?, ?, ?, ?, ? )")
		dbMock.ExpectExec(escapedQuery).WithArgs(labelDefID, tenantID, "some-key", nil, version).WillReturnResult(sqlmock.NewResult(1, 1))
		defer dbMock.AssertExpectations(t)
		sut := labeldef.NewRepository(mockConverter)
		// WHEN
		err := sut.Create(ctx, in)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error if insert fails", func(t *testing.T) {
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}, Version: version}, nil)

		sut := labeldef.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		escapedQuery := regexp.QuoteMeta("INSERT INTO public.label_definitions ( id, tenant_id, key, schema, version ) VALUES ( ?, ?, ?, ?, ? )")
		dbMock.ExpectExec(escapedQuery).WillReturnError(errors.New("some error"))
		defer dbMock.AssertExpectations(t)
		// WHEN
		err := sut.Create(ctx, in)
		// THEN
		require.EqualError(t, err, "while inserting Label Definition: Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepositoryUpdateLabelDefinition(t *testing.T) {
	// GIVEN
	labelDefID := "d048f47b-b700-49ed-913d-180c3748164b"
	tenantID := "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
	someString := "any"
	version := 42
	var someSchema interface{} = someString

	in := model.LabelDefinition{
		ID:     labelDefID,
		Tenant: tenantID,
		Key:    "some-key",
		Schema: &someSchema,
	}

	t.Run("successfully updated definition with schema", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}, Version: version}, nil)

		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET schema = ?, version = version+1 WHERE id = ? AND tenant_id = ?`)
		dbMock.ExpectExec(escapedQuery).WithArgs("any", labelDefID, tenantID).WillReturnResult(sqlmock.NewResult(1, 1))
		defer dbMock.AssertExpectations(t)

		sut := labeldef.NewRepository(mockConverter)

		// WHEN
		err := sut.Update(ctx, in)
		// THEN
		require.NoError(t, err)
	})

	t.Run("successfully updated definition without schema", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, Version: version}, nil)

		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET schema = ?, version = version+1 WHERE id = ? AND tenant_id = ?`)
		dbMock.ExpectExec(escapedQuery).WithArgs(nil, labelDefID, tenantID).WillReturnResult(sqlmock.NewResult(1, 1))
		defer dbMock.AssertExpectations(t)

		sut := labeldef.NewRepository(mockConverter)

		// WHEN
		err := sut.Update(ctx, in)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error if update fails", func(t *testing.T) {
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}, Version: version}, nil)

		sut := labeldef.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET schema = ?, version = version+1 WHERE id = ? AND tenant_id = ?`)
		dbMock.ExpectExec(escapedQuery).WillReturnError(errors.New("some error"))
		defer dbMock.AssertExpectations(t)

		// WHEN
		err := sut.Update(ctx, in)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns error if update fails", func(t *testing.T) {
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}, Version: version}, nil)

		sut := labeldef.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET schema = ?, version = version+1 WHERE id = ? AND tenant_id = ?`)
		dbMock.ExpectExec(escapedQuery).WillReturnError(errors.New("some error"))
		defer dbMock.AssertExpectations(t)

		// WHEN
		err := sut.Update(ctx, in)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns error if no row was affected by query", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}, Version: version}, nil)

		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET schema = ?, version = version+1 WHERE id = ? AND tenant_id = ?`)
		dbMock.ExpectExec(escapedQuery).WithArgs("any", labelDefID, tenantID).WillReturnResult(sqlmock.NewResult(1, 0))
		defer dbMock.AssertExpectations(t)

		sut := labeldef.NewRepository(mockConverter)

		// WHEN
		err := sut.Update(ctx, in)
		// THEN
		require.EqualError(t, err, "Could not update object due to concurrent update")
	})
}

func TestRepositoryUpdateLabelDefinitionWithVersion(t *testing.T) {
	// GIVEN
	labelDefID := "d048f47b-b700-49ed-913d-180c3748164b"
	tenantID := "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
	someString := "any"
	version := 42
	var someSchema interface{} = someString

	in := model.LabelDefinition{
		ID:     labelDefID,
		Tenant: tenantID,
		Key:    "some-key",
		Schema: &someSchema,
	}

	t.Run("successfully updated definition with schema", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}, Version: version}, nil)

		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET schema = ?, version = version+1 WHERE id = ? AND version = ? AND tenant_id = ?`)
		dbMock.ExpectExec(escapedQuery).WithArgs("any", labelDefID, version, tenantID).WillReturnResult(sqlmock.NewResult(1, 1))
		defer dbMock.AssertExpectations(t)

		sut := labeldef.NewRepository(mockConverter)

		// WHEN
		err := sut.UpdateWithVersion(ctx, in)
		// THEN
		require.NoError(t, err)
	})

	t.Run("successfully updated definition without schema", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, Version: version}, nil)

		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET schema = ?, version = version+1 WHERE id = ? AND version = ? AND tenant_id = ?`)
		dbMock.ExpectExec(escapedQuery).WithArgs(nil, labelDefID, version, tenantID).WillReturnResult(sqlmock.NewResult(1, 1))
		defer dbMock.AssertExpectations(t)

		sut := labeldef.NewRepository(mockConverter)

		// WHEN
		err := sut.UpdateWithVersion(ctx, in)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error if update fails", func(t *testing.T) {
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}, Version: version}, nil)

		sut := labeldef.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET schema = ?, version = version+1 WHERE id = ? AND version = ? AND tenant_id = ?`)
		dbMock.ExpectExec(escapedQuery).WillReturnError(errors.New("some error"))
		defer dbMock.AssertExpectations(t)

		// WHEN
		err := sut.UpdateWithVersion(ctx, in)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns error if no row was affected by query", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}, Version: version}, nil)

		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET schema = ?, version = version+1 WHERE id = ? AND version = ? AND tenant_id = ?`)
		dbMock.ExpectExec(escapedQuery).WithArgs("any", labelDefID, version, tenantID).WillReturnResult(sqlmock.NewResult(1, 0))
		defer dbMock.AssertExpectations(t)

		sut := labeldef.NewRepository(mockConverter)

		// WHEN
		err := sut.UpdateWithVersion(ctx, in)
		// THEN
		require.EqualError(t, err, "Could not update object due to concurrent update")
	})
}

func TestRepositoryGetByKey(t *testing.T) {
	t.Run("returns LabelDefinition", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		var someSchema interface{} = ExampleSchema{Title: "title"}
		mockConverter.On("FromEntity",
			labeldef.Entity{ID: "id", TenantID: "tenant", Key: "key", SchemaJSON: sql.NullString{Valid: true, String: `{"title":"title"}`}}).
			Return(
				model.LabelDefinition{
					ID:     "id",
					Tenant: "tenant",
					Key:    "key",
					Schema: &someSchema}, nil)
		sut := labeldef.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "schema"}).AddRow("id", "tenant", "key", `{"title":"title"}`)
		dbMock.ExpectQuery(`^SELECT (.+) FROM public.label_definitions WHERE tenant_id = \$1 AND key = \$2$`).
			WithArgs("tenant", "key").WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := sut.GetByKey(ctx, "tenant", "key")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "id", actual.ID)
		assert.Equal(t, "tenant", actual.Tenant)
		assert.Equal(t, "key", actual.Key)
		assert.Equal(t, &someSchema, actual.Schema)
	})
	t.Run("returns errorNotFound if LabelDefinition does not exist", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.label_definitions WHERE tenant_id = \$1 AND key = \$2$`).
			WithArgs("anything", "anything").WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "key", "schema"}))
		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		sut := labeldef.NewRepository(nil)
		// WHEN
		actual, err := sut.GetByKey(ctx, "anything", "anything")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Object not found")
		assert.Nil(t, actual)
	})
	t.Run("returns error when conversion fails", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", mock.Anything).Return(model.LabelDefinition{}, errors.New("conversion error"))
		sut := labeldef.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "schema"}).AddRow("id", "tenant", "key", `{"title":"title"}`)
		dbMock.ExpectQuery(`^SELECT (.+) FROM public.label_definitions WHERE tenant_id = \$1 AND key = \$2$`).
			WithArgs("tenant", "key").WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := sut.GetByKey(ctx, "tenant", "key")
		// THEN
		require.EqualError(t, err, "while converting Label Definition: conversion error")
	})
	t.Run("returns error if select fails", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.label_definitions WHERE tenant_id = \$1 AND key = \$2$`).
			WithArgs("tenant", "key").WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := sut.GetByKey(ctx, "tenant", "key")
		// THEN
		require.EqualError(t, err, "while getting Label Definition by key=key: Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepositoryList(t *testing.T) {
	t.Run("returns list of Label Definitions", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		mockConverter.On("FromEntity",
			labeldef.Entity{ID: "id1", TenantID: "tenant", Key: "key1", Version: 0}).
			Return(
				model.LabelDefinition{
					ID:      "id1",
					Tenant:  "tenant",
					Key:     "key1",
					Version: 0,
				}, nil)
		mockConverter.On("FromEntity",
			labeldef.Entity{ID: "id2", TenantID: "tenant", Key: "key2", SchemaJSON: sql.NullString{Valid: true, String: `{"title":"title"}`}, Version: 0}).
			Return(
				model.LabelDefinition{
					ID:      "id2",
					Tenant:  "tenant",
					Key:     "key2",
					Version: 0,
				}, nil)
		sut := labeldef.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "schema", "version"}).
			AddRow("id1", "tenant", "key1", nil, 0).
			AddRow("id2", "tenant", "key2", `{"title":"title"}`, 0)

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.label_definitions WHERE tenant_id = \$1$`).WithArgs("tenant").WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := sut.List(ctx, "tenant")
		// THEN
		require.NoError(t, err)
		require.Len(t, actual, 2)
		assert.Equal(t, "id1", actual[0].ID)
		assert.Equal(t, "key1", actual[0].Key)
		assert.Equal(t, 0, actual[0].Version)
		assert.Equal(t, "id2", actual[1].ID)
		assert.Equal(t, "key2", actual[1].Key)
		assert.Equal(t, 0, actual[1].Version)
	})
	t.Run("returns empty list of Label Definitions if given tenant has nothing defined", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewRepository(nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "schema", "version"})

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.label_definitions WHERE tenant_id = \$1$`).WithArgs("tenant").WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := sut.List(ctx, "tenant")
		// THEN
		require.NoError(t, err)
		require.Empty(t, actual)
	})

	t.Run("returns error when conversion fails", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		mockConverter.On("FromEntity",
			labeldef.Entity{ID: "id1", TenantID: "tenant", Key: "key1", Version: 0}).
			Return(
				model.LabelDefinition{}, errors.New("conversion error"))

		sut := labeldef.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "schema", "version"}).
			AddRow("id1", "tenant", "key1", nil, 0).
			AddRow("id2", "tenant", "key2", `{"title":"title"}`, 0)

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.label_definitions WHERE tenant_id = \$1$`).WithArgs("tenant").WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := sut.List(ctx, "tenant")
		// THEN
		require.EqualError(t, err, "while converting Label Definition [key=key1]: conversion error")
	})
	t.Run("returns error if if select fails", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewRepository(nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.label_definitions WHERE tenant_id = \$1$`).
			WithArgs("tenant").WillReturnError(errors.New("db error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := sut.List(ctx, "tenant")
		// THEN
		require.EqualError(t, err, "while listing Label Definitions: Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepositoryLabelDefExists(t *testing.T) {
	t.Run("returns true", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta("SELECT 1 FROM public.label_definitions WHERE tenant_id = $1 AND key = $2")
		dbMock.ExpectQuery(escapedQuery).WithArgs("tenant", "key").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow("1"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := sut.Exists(ctx, "tenant", "key")
		// THEN
		require.NoError(t, err)
		assert.True(t, actual)
	})

	t.Run("returns false", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta("SELECT 1 FROM public.label_definitions WHERE tenant_id = $1 AND key = $2")
		dbMock.ExpectQuery(escapedQuery).WithArgs("anything", "anything").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		sut := labeldef.NewRepository(nil)
		// WHEN
		actual, err := sut.Exists(ctx, "anything", "anything")
		// THEN
		require.NoError(t, err)
		assert.False(t, actual)
	})
	t.Run("returns error if select fails", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta("SELECT 1 FROM public.label_definitions WHERE tenant_id = $1 AND key = $2")
		dbMock.ExpectQuery(escapedQuery).WithArgs("tenant", "key").WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := sut.Exists(ctx, "tenant", "key")
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_Upsert(t *testing.T) {
	t.Run("Success - Label for Runtime", func(t *testing.T) {
		// GIVEN
		key := "test"
		tnt := "tenant"
		schema := "{}"
		schemaInterface := reflect.ValueOf(schema).Interface()
		version := 42

		labeldefModel := model.LabelDefinition{
			ID:      "foo",
			Tenant:  tnt,
			Key:     key,
			Schema:  &schemaInterface,
			Version: version,
		}
		labeldefEntity := labeldef.Entity{
			ID:       "foo",
			TenantID: tnt,
			Key:      key,
			SchemaJSON: sql.NullString{
				String: schema,
				Valid:  true,
			},
			Version: version,
		}

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", labeldefModel).Return(labeldefEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := labeldef.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`INSERT INTO public.label_definitions ( id, tenant_id, key, schema, version ) VALUES ( ?, ?, ?, ?, ? ) ON CONFLICT ( tenant_id, key ) DO UPDATE SET schema=EXCLUDED.schema WHERE  public.label_definitions.tenant_id = ?`)
		dbMock.ExpectExec(escapedQuery).WithArgs(labeldefEntity.ID, labeldefEntity.TenantID, labeldefEntity.Key, labeldefEntity.SchemaJSON, labeldefEntity.Version, labeldefEntity.TenantID).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, labeldefModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when entityConverting fails", func(t *testing.T) {
		labeldefModel := model.LabelDefinition{}
		labeldefEntity := labeldef.Entity{}
		testErr := errors.New("test-err")

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", labeldefModel).Return(labeldefEntity, testErr).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := labeldef.NewRepository(mockConverter)

		// WHEN
		ctx := context.TODO()
		err := labelRepo.Upsert(ctx, labeldefModel)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
	})
}
