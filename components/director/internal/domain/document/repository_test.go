package document_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCreateDocument(t *testing.T) {
	// GIVEN
	docID := "d048f47b-b700-49ed-913d-180c3748164b"
	tenantID := "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
	appID := "7c0a2239-1787-473e-856e-0653d02c49ee"
	//someString := "any"
	//var someSchema interface{}
	//someSchema = someString

	title := "title"
	displayName := "name"
	description := "desc"
	format := "MARKDOWN"
	kind := "kind"
	data := "abc"

	in := model.Document{
		ID:            docID,
		Tenant:        tenantID,
		ApplicationID: appID,
		Title:         title,
		DisplayName:   displayName,
		Description:   description,
		Format:        model.DocumentFormat(format),
		Kind:          &kind,
		Data:          &data,
		FetchRequest:  nil,
	}

	t.Run("successfully created document", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", in).Return(document.Entity{
			ID:          docID,
			TenantID:    tenantID,
			AppID:       appID,
			Title:       title,
			DisplayName: displayName,
			Description: description,
			Format:      format,
			Kind: sql.NullString{
				String: kind,
				Valid:  true,
			},
			Data: sql.NullString{
				String: data,
				Valid:  true,
			},
			FetchRequest: sql.NullString{},
		}, nil)
		escapedQuery := regexp.QuoteMeta("insert into public.documents (id, tenant_id, app_id, title, display_name, description, format, kind, data, fetch_request) values(?,?,?,?,?,?,?,?,?,?)")
		dbMock.ExpectExec(escapedQuery).WithArgs(docID, tenantID, appID, title, displayName, description, format, kind, data, nil).WillReturnResult(sqlmock.NewResult(1, 1))

		defer dbMock.AssertExpectations(t)
		sut := document.NewPostgresRepository()
		// WHEN
		err := sut.Create(ctx, &in)
		// THEN
		require.NoError(t, err)
	})
	//
	//t.Run("successfully created definition without schema", func(t *testing.T) {
	//	db, dbMock := testdb.MockDatabase(t)
	//	ctx := context.TODO()
	//	ctx = persistence.SaveToContext(ctx, db)
	//	mockConverter := &automock.Converter{}
	//	defer mockConverter.AssertExpectations(t)
	//	mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID}, nil)
	//	escapedQuery := regexp.QuoteMeta("insert into public.label_definitions (id,tenant_id,key,schema) values(?,?,?,?)")
	//	dbMock.ExpectExec(escapedQuery).WithArgs(labelDefID, tenantID, "some-key", nil).WillReturnResult(sqlmock.NewResult(1, 1))
	//	defer dbMock.AssertExpectations(t)
	//	sut := labeldef.NewRepository(mockConverter)
	//	// WHEN
	//	err := sut.Create(ctx, in)
	//	// THEN
	//	require.NoError(t, err)
	//})
	//
	//t.Run("returns error if missing persistence context", func(t *testing.T) {
	//	sut := labeldef.NewRepository(nil)
	//	ctx := context.TODO()
	//	err := sut.Create(ctx, model.LabelDefinition{})
	//	require.EqualError(t, err, "unable to fetch database from context")
	//})
	//
	//t.Run("returns error if insert fails", func(t *testing.T) {
	//	mockConverter := &automock.Converter{}
	//	defer mockConverter.AssertExpectations(t)
	//	mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}}, nil)
	//
	//	sut := labeldef.NewRepository(mockConverter)
	//	db, dbMock := testdb.MockDatabase(t)
	//	ctx := context.TODO()
	//	ctx = persistence.SaveToContext(ctx, db)
	//	escapedQuery := regexp.QuoteMeta("insert into public.label_definitions (id,tenant_id,key,schema) values(?,?,?,?)")
	//	dbMock.ExpectExec(escapedQuery).WillReturnError(errors.New("some error"))
	//	defer dbMock.AssertExpectations(t)
	//	// WHEN
	//	err := sut.Create(ctx, in)
	//	// THEN
	//	require.EqualError(t, err, "while inserting Label Definition: some error")
	//})
}

//
//func TestRepositoryUpdateLabelDefinition(t *testing.T) {
//	// GIVEN
//	labelDefID := "d048f47b-b700-49ed-913d-180c3748164b"
//	tenantID := "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
//	someString := "any"
//	var someSchema interface{}
//	someSchema = someString
//
//	in := model.LabelDefinition{
//		ID:     labelDefID,
//		Tenant: tenantID,
//		Key:    "some-key",
//		Schema: &someSchema,
//	}
//
//	t.Run("successfully updated definition with schema", func(t *testing.T) {
//		db, dbMock := testdb.MockDatabase(t)
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//
//		mockConverter := &automock.Converter{}
//		defer mockConverter.AssertExpectations(t)
//		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}}, nil)
//
//		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET "schema" = ? WHERE "id" = ?`)
//		dbMock.ExpectExec(escapedQuery).WithArgs("any", labelDefID).WillReturnResult(sqlmock.NewResult(1, 1))
//		defer dbMock.AssertExpectations(t)
//
//		sut := labeldef.NewRepository(mockConverter)
//
//		// WHEN
//		err := sut.Update(ctx, in)
//		// THEN
//		require.NoError(t, err)
//	})
//
//	t.Run("successfully updated definition without schema", func(t *testing.T) {
//		db, dbMock := testdb.MockDatabase(t)
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//
//		mockConverter := &automock.Converter{}
//		defer mockConverter.AssertExpectations(t)
//		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID}, nil)
//
//		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET "schema" = ? WHERE "id" = ?`)
//		dbMock.ExpectExec(escapedQuery).WithArgs(nil, labelDefID).WillReturnResult(sqlmock.NewResult(1, 1))
//		defer dbMock.AssertExpectations(t)
//
//		sut := labeldef.NewRepository(mockConverter)
//
//		// WHEN
//		err := sut.Update(ctx, in)
//		// THEN
//		require.NoError(t, err)
//	})
//
//	t.Run("returns error if missing persistence context", func(t *testing.T) {
//		sut := labeldef.NewRepository(nil)
//		ctx := context.TODO()
//		err := sut.Update(ctx, model.LabelDefinition{})
//		require.EqualError(t, err, "while fetching persistence from context: unable to fetch database from context")
//	})
//
//	t.Run("returns error if update fails", func(t *testing.T) {
//		mockConverter := &automock.Converter{}
//		defer mockConverter.AssertExpectations(t)
//		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}}, nil)
//
//		sut := labeldef.NewRepository(mockConverter)
//		db, dbMock := testdb.MockDatabase(t)
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//
//		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET "schema" = ? WHERE "id" = ?`)
//		dbMock.ExpectExec(escapedQuery).WillReturnError(errors.New("some error"))
//		defer dbMock.AssertExpectations(t)
//
//		// WHEN
//		err := sut.Update(ctx, in)
//		// THEN
//		require.EqualError(t, err, "while updating Label Definition: some error")
//	})
//
//	t.Run("returns error if update fails", func(t *testing.T) {
//		mockConverter := &automock.Converter{}
//		defer mockConverter.AssertExpectations(t)
//		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}}, nil)
//
//		sut := labeldef.NewRepository(mockConverter)
//		db, dbMock := testdb.MockDatabase(t)
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//
//		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET "schema" = ? WHERE "id" = ?`)
//		dbMock.ExpectExec(escapedQuery).WillReturnError(errors.New("some error"))
//		defer dbMock.AssertExpectations(t)
//
//		// WHEN
//		err := sut.Update(ctx, in)
//		// THEN
//		require.EqualError(t, err, "while updating Label Definition: some error")
//	})
//
//	t.Run("returns error if no row was affected by query", func(t *testing.T) {
//		db, dbMock := testdb.MockDatabase(t)
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//
//		mockConverter := &automock.Converter{}
//		defer mockConverter.AssertExpectations(t)
//		mockConverter.On("ToEntity", in).Return(labeldef.Entity{ID: labelDefID, Key: "some-key", TenantID: tenantID, SchemaJSON: sql.NullString{String: "any", Valid: true}}, nil)
//
//		escapedQuery := regexp.QuoteMeta(`UPDATE public.label_definitions SET "schema" = ? WHERE "id" = ?`)
//		dbMock.ExpectExec(escapedQuery).WithArgs("any", labelDefID).WillReturnResult(sqlmock.NewResult(1, 0))
//		defer dbMock.AssertExpectations(t)
//
//		sut := labeldef.NewRepository(mockConverter)
//
//		// WHEN
//		err := sut.Update(ctx, in)
//		// THEN
//		require.EqualError(t, err, "no row was affected by query")
//	})
//}
//
//func TestRepositoryGetByKey(t *testing.T) {
//	t.Run("returns LabelDefinition", func(t *testing.T) {
//		// GIVEN
//		mockConverter := &automock.Converter{}
//		defer mockConverter.AssertExpectations(t)
//
//		var someSchema interface{} = ExampleSchema{Title: "title"}
//		mockConverter.On("FromEntity",
//			labeldef.Entity{ID: "id", TenantID: "tenant", Key: "key", SchemaJSON: sql.NullString{Valid: true, String: `{"title":"title"}`}}).
//			Return(
//				model.LabelDefinition{
//					ID:     "id",
//					Tenant: "tenant",
//					Key:    "key",
//					Schema: &someSchema}, nil)
//		sut := labeldef.NewRepository(mockConverter)
//
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta("select * from public.label_definitions where tenant_id=$1 and key=$2")
//		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "schema"}).AddRow("id", "tenant", "key", `{"title":"title"}`)
//		dbMock.ExpectQuery(escapedQuery).WithArgs("tenant", "key").WillReturnRows(mockedRows)
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//		// WHEN
//		actual, err := sut.GetByKey(ctx, "tenant", "key")
//		// THEN
//		require.NoError(t, err)
//		assert.Equal(t, "id", actual.ID)
//		assert.Equal(t, "tenant", actual.Tenant)
//		assert.Equal(t, "key", actual.Key)
//		assert.Equal(t, &someSchema, actual.Schema)
//	})
//	t.Run("returns nil if LabelDefinition does not exist", func(t *testing.T) {
//		// GIVEN
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta("select * from public.label_definitions where tenant_id=$1 and key=$2")
//		dbMock.ExpectQuery(escapedQuery).WithArgs("anything", "anything").WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "key", "schema"}))
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//
//		sut := labeldef.NewRepository(nil)
//		// WHEN
//		actual, err := sut.GetByKey(ctx, "anything", "anything")
//		// THEN
//		require.NoError(t, err)
//		assert.Nil(t, actual)
//
//	})
//	t.Run("returns error when conversion fails", func(t *testing.T) {
//		// GIVEN
//		mockConverter := &automock.Converter{}
//		defer mockConverter.AssertExpectations(t)
//		mockConverter.On("FromEntity", mock.Anything).Return(model.LabelDefinition{}, errors.New("conversion error"))
//		sut := labeldef.NewRepository(mockConverter)
//
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta("select * from public.label_definitions where tenant_id=$1 and key=$2")
//		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "schema"}).AddRow("id", "tenant", "key", `{"title":"title"}`)
//		dbMock.ExpectQuery(escapedQuery).WithArgs("tenant", "key").WillReturnRows(mockedRows)
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//		// WHEN
//		_, err := sut.GetByKey(ctx, "tenant", "key")
//		// THEN
//		require.EqualError(t, err, "while converting Label Definition: conversion error")
//	})
//	t.Run("returns error if missing persistence context", func(t *testing.T) {
//		// GIVEN
//		sut := labeldef.NewRepository(nil)
//		// WHEN
//		_, err := sut.GetByKey(context.TODO(), "tenant", "key")
//		// THEN
//		require.EqualError(t, err, "unable to fetch database from context")
//	})
//	t.Run("returns error if select fails", func(t *testing.T) {
//		// GIVEN
//		sut := labeldef.NewRepository(nil)
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta("select * from public.label_definitions where tenant_id=$1 and key=$2")
//		dbMock.ExpectQuery(escapedQuery).WithArgs("tenant", "key").WillReturnError(errors.New("persistence error"))
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//		// WHEN
//		_, err := sut.GetByKey(ctx, "tenant", "key")
//		// THEN
//		require.EqualError(t, err, "while querying Label Definition: persistence error")
//	})
//}
//
//func TestRepositoryList(t *testing.T) {
//	t.Run("returns list of Label Definitions", func(t *testing.T) {
//		// GIVEN
//		mockConverter := &automock.Converter{}
//		defer mockConverter.AssertExpectations(t)
//
//		mockConverter.On("FromEntity",
//			labeldef.Entity{ID: "id1", TenantID: "tenant", Key: "key1"}).
//			Return(
//				model.LabelDefinition{
//					ID:     "id1",
//					Tenant: "tenant",
//					Key:    "key1",
//				}, nil)
//		mockConverter.On("FromEntity",
//			labeldef.Entity{ID: "id2", TenantID: "tenant", Key: "key2", SchemaJSON: sql.NullString{Valid: true, String: `{"title":"title"}`}}).
//			Return(
//				model.LabelDefinition{
//					ID:     "id2",
//					Tenant: "tenant",
//					Key:    "key2",
//				}, nil)
//		sut := labeldef.NewRepository(mockConverter)
//
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta("select * from public.label_definitions where tenant_id=$1")
//		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "schema"}).
//			AddRow("id1", "tenant", "key1", nil).
//			AddRow("id2", "tenant", "key2", `{"title":"title"}`)
//
//		dbMock.ExpectQuery(escapedQuery).WithArgs("tenant").WillReturnRows(mockedRows)
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//		// WHEN
//		actual, err := sut.List(ctx, "tenant")
//		// THEN
//		require.NoError(t, err)
//		require.Len(t, actual, 2)
//		assert.Equal(t, "id1", actual[0].ID)
//		assert.Equal(t, "key1", actual[0].Key)
//		assert.Equal(t, "id2", actual[1].ID)
//		assert.Equal(t, "key2", actual[1].Key)
//	})
//	t.Run("returns empty list of Label Definitions if given tenant has nothing defined", func(t *testing.T) {
//		// GIVEN
//		sut := labeldef.NewRepository(nil)
//
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta("select * from public.label_definitions where tenant_id=$1")
//		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "schema"})
//
//		dbMock.ExpectQuery(escapedQuery).WithArgs("tenant").WillReturnRows(mockedRows)
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//		// WHEN
//		actual, err := sut.List(ctx, "tenant")
//		// THEN
//		require.NoError(t, err)
//		require.Empty(t, actual)
//	})
//
//	t.Run("returns error when conversion fails", func(t *testing.T) {
//		// GIVEN
//		mockConverter := &automock.Converter{}
//		defer mockConverter.AssertExpectations(t)
//
//		mockConverter.On("FromEntity",
//			labeldef.Entity{ID: "id1", TenantID: "tenant", Key: "key1"}).
//			Return(
//				model.LabelDefinition{}, errors.New("conversion error"))
//
//		sut := labeldef.NewRepository(mockConverter)
//
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta("select * from public.label_definitions where tenant_id=$1")
//		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "schema"}).
//			AddRow("id1", "tenant", "key1", nil).
//			AddRow("id2", "tenant", "key2", `{"title":"title"}`)
//
//		dbMock.ExpectQuery(escapedQuery).WithArgs("tenant").WillReturnRows(mockedRows)
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//		// WHEN
//		_, err := sut.List(ctx, "tenant")
//		// THEN
//		require.EqualError(t, err, "while converting Label Definition [key=key1]: conversion error")
//
//	})
//	t.Run("returns error if missing persistence context", func(t *testing.T) {
//		// GIVEN
//		sut := labeldef.NewRepository(nil)
//		// WHEN
//		_, err := sut.List(context.TODO(), "tenant")
//		// THEN
//		require.EqualError(t, err, "unable to fetch database from context")
//	})
//	t.Run("returns error if if select fails", func(t *testing.T) {
//		// GIVEN
//		sut := labeldef.NewRepository(nil)
//
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta("select * from public.label_definitions where tenant_id=$1")
//
//		dbMock.ExpectQuery(escapedQuery).WithArgs("tenant").WillReturnError(errors.New("db error"))
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//		// WHEN
//		_, err := sut.List(ctx, "tenant")
//		// THEN
//		require.EqualError(t, err, "while listing Label Definitions: db error")
//	})
//}
//
//func TestRepositoryLabelDefExists(t *testing.T) {
//	t.Run("returns true", func(t *testing.T) {
//		// GIVEN
//		sut := labeldef.NewRepository(nil)
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta("select 1 as exists from public.label_definitions where tenant_id=$1 and key=$2")
//		dbMock.ExpectQuery(escapedQuery).WithArgs("tenant", "key").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow("1"))
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//		// WHEN
//		actual, err := sut.Exists(ctx, "tenant", "key")
//		// THEN
//		require.NoError(t, err)
//		assert.True(t, actual)
//	})
//
//	t.Run("returns false", func(t *testing.T) {
//		// GIVEN
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta("select 1 as exists from public.label_definitions where tenant_id=$1 and key=$2")
//		dbMock.ExpectQuery(escapedQuery).WithArgs("anything", "anything").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//
//		sut := labeldef.NewRepository(nil)
//		// WHEN
//		actual, err := sut.Exists(ctx, "anything", "anything")
//		// THEN
//		require.NoError(t, err)
//		assert.False(t, actual)
//
//	})
//	t.Run("returns error if missing persistence context", func(t *testing.T) {
//		// GIVEN
//		sut := labeldef.NewRepository(nil)
//		// WHEN
//		_, err := sut.Exists(context.TODO(), "tenant", "key")
//		// THEN
//		require.EqualError(t, err, "unable to fetch database from context")
//	})
//	t.Run("returns error if select fails", func(t *testing.T) {
//		// GIVEN
//		sut := labeldef.NewRepository(nil)
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta("select 1 as exists from public.label_definitions where tenant_id=$1 and key=$2")
//		dbMock.ExpectQuery(escapedQuery).WithArgs("tenant", "key").WillReturnError(errors.New("persistence error"))
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//		// WHEN
//		_, err := sut.Exists(ctx, "tenant", "key")
//		// THEN
//		require.EqualError(t, err, "while querying Label Definition: persistence error")
//	})
//}
//
//func TestRepository_DeleteByKey(t *testing.T) {
//	t.Run("Success", func(t *testing.T) {
//		// GIVEN
//		key := "test"
//		tnt := "tenant"
//
//		mockConverter := &automock.Converter{}
//		defer mockConverter.AssertExpectations(t)
//
//		repo := labeldef.NewRepository(mockConverter)
//
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.label_definitions WHERE tenant_id=$1 AND key=$2`)
//		dbMock.ExpectExec(escapedQuery).WithArgs(tnt, key).WillReturnResult(sqlmock.NewResult(1, 1))
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//		// WHEN
//		err := repo.DeleteByKey(ctx, tnt, key)
//		// THEN
//		require.NoError(t, err)
//	})
//
//	t.Run("Error - Operation", func(t *testing.T) {
//		// GIVEN
//		key := "test"
//		tnt := "tenant"
//		testErr := errors.New("Test err")
//
//		mockConverter := &automock.Converter{}
//		defer mockConverter.AssertExpectations(t)
//
//		repo := labeldef.NewRepository(mockConverter)
//
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.label_definitions WHERE tenant_id=$1 AND key=$2`)
//		dbMock.ExpectExec(escapedQuery).WithArgs(tnt, key).WillReturnError(testErr)
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//		// WHEN
//		err := repo.DeleteByKey(ctx, tnt, key)
//		// THEN
//		require.Error(t, err)
//		assert.Contains(t, err.Error(), testErr.Error())
//	})
//
//	t.Run("Error - Missing persistence", func(t *testing.T) {
//		// GIVEN
//		repo := labeldef.NewRepository(nil)
//		key := "key"
//		tnt := "tenant"
//		ctx := context.TODO()
//
//		// WHEN
//		err := repo.DeleteByKey(ctx, tnt, key)
//		// THEN
//		require.Error(t, err)
//		require.Contains(t, err.Error(), "unable to fetch database from context")
//	})
//
//	t.Run("Error - No rows affected", func(t *testing.T) {
//		// GIVEN
//		key := "test"
//		tnt := "tenant"
//
//		mockConverter := &automock.Converter{}
//		defer mockConverter.AssertExpectations(t)
//
//		repo := labeldef.NewRepository(mockConverter)
//
//		db, dbMock := testdb.MockDatabase(t)
//		defer dbMock.AssertExpectations(t)
//
//		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.label_definitions WHERE tenant_id=$1 AND key=$2`)
//		dbMock.ExpectExec(escapedQuery).WithArgs(tnt, key).WillReturnResult(sqlmock.NewResult(1, 0))
//
//		ctx := context.TODO()
//		ctx = persistence.SaveToContext(ctx, db)
//		// WHEN
//		err := repo.DeleteByKey(ctx, tnt, key)
//		// THEN
//		require.EqualError(t, err, "no rows were affected by query")
//	})
//}
//
