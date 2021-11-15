package spec_test

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestRepository_GetByID(t *testing.T) {
	apiSpecModel := fixModelAPISpec()
	apiSpecEntity := fixAPISpecEntity()
	eventSpecModel := fixModelEventSpec()
	eventSpecEntity := fixEventSpecEntity()

	apiSpecSuite := testdb.RepoGetTestSuite{
		Name: "Get API Spec By ID",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type, event_spec_format, event_spec_type, custom_type FROM public.specifications WHERE id = $1 AND (id IN (SELECT id FROM api_specifications_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{specID, tenant},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSpecColumns()).AddRow(fixAPISpecRow()...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSpecColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: spec.NewRepository,
		ExpectedModelEntity: apiSpecModel,
		ExpectedDBEntity:    apiSpecEntity,
		MethodArgs:          []interface{}{tenant, specID, model.APISpecReference},
	}

	eventSpecSuite := testdb.RepoGetTestSuite{
		Name: "Get Event Spec By ID",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type, event_spec_format, event_spec_type, custom_type FROM public.specifications WHERE id = $1 AND (id IN (SELECT id FROM event_specifications_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{specID, tenant},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSpecColumns()).AddRow(fixEventSpecRow()...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSpecColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: spec.NewRepository,
		ExpectedModelEntity: eventSpecModel,
		ExpectedDBEntity:    eventSpecEntity,
		MethodArgs:          []interface{}{tenant, specID, model.EventSpecReference},
	}

	apiSpecSuite.Run(t)
	eventSpecSuite.Run(t)
}

func TestRepository_Create(t *testing.T) {
	var nilSpecModel *model.Spec
	apiSpecModel := fixModelAPISpec()
	apiSpecEntity := fixAPISpecEntity()
	eventSpecModel := fixModelEventSpec()
	eventSpecEntity := fixEventSpecEntity()

	apiSpecSuite := testdb.RepoCreateTestSuite{
		Name: "Create API Specification",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM api_definitions_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenant, apiID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO public.specifications \(.+\) VALUES \(.+\)$`,
				Args:        fixAPISpecCreateArgs(apiSpecModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:       spec.NewRepository,
		ModelEntity:               apiSpecModel,
		DBEntity:                  apiSpecEntity,
		NilModelEntity:            nilSpecModel,
		TenantID:                  tenant,
		DisableConverterErrorTest: true,
	}

	eventSpecSuite := testdb.RepoCreateTestSuite{
		Name: "Create Event Specification",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM event_api_definitions_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenant, eventID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO public.specifications \(.+\) VALUES \(.+\)$`,
				Args:        fixEventSpecCreateArgs(eventSpecModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:       spec.NewRepository,
		ModelEntity:               eventSpecModel,
		DBEntity:                  eventSpecEntity,
		NilModelEntity:            nilSpecModel,
		TenantID:                  tenant,
		DisableConverterErrorTest: true,
	}

	apiSpecSuite.Run(t)
	eventSpecSuite.Run(t)
}

/*
func TestRepository_ListByReferenceObjectID(t *testing.T) {
	// GIVEN

	t.Run("Success for API", func(t *testing.T) {
		firstSpecID := "111111111-1111-1111-1111-111111111111"
		firstSpecEntity := fixAPISpecEntityWithID(firstSpecID)
		secondSpecID := "222222222-2222-2222-2222-222222222222"
		secondAPIDefEntity := fixAPISpecEntityWithID(secondSpecID)

		selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.specifications
		WHERE %s AND api_def_id = \$2
		ORDER BY created_at`, fixTenantIsolationSubquery())

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixSpecColumns()).
			AddRow(fixAPISpecRowWithID(firstSpecID)...).
			AddRow(fixAPISpecRowWithID(secondSpecID)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenant, apiID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", firstSpecEntity).Return(*fixModelAPISpecWithID(firstSpecID), nil)
		convMock.On("FromEntity", secondAPIDefEntity).Return(*fixModelAPISpecWithID(secondSpecID), nil)
		pgRepository := spec.NewRepository(convMock)
		// WHEN
		modelSpec, err := pgRepository.ListByReferenceObjectID(ctx, tenant, model.APISpecReference, apiID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelSpec, 2)
		assert.Equal(t, firstSpecID, modelSpec[0].ID)
		assert.Equal(t, secondSpecID, modelSpec[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("Success for Event", func(t *testing.T) {
		firstSpecID := "111111111-1111-1111-1111-111111111111"
		firstSpecEntity := fixEventSpecEntityWithID(firstSpecID)
		secondSpecID := "222222222-2222-2222-2222-222222222222"
		secondAPIDefEntity := fixEventSpecEntityWithID(secondSpecID)

		selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.specifications
		WHERE %s AND event_def_id = \$2
		ORDER BY created_at`, fixTenantIsolationSubquery())

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixSpecColumns()).
			AddRow(fixEventSpecRowWithID(firstSpecID)...).
			AddRow(fixEventSpecRowWithID(secondSpecID)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenant, eventID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", firstSpecEntity).Return(*fixModelEventSpecWithID(firstSpecID), nil)
		convMock.On("FromEntity", secondAPIDefEntity).Return(*fixModelEventSpecWithID(secondSpecID), nil)
		pgRepository := spec.NewRepository(convMock)
		// WHEN
		modelSpec, err := pgRepository.ListByReferenceObjectID(ctx, tenant, model.EventSpecReference, eventID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelSpec, 2)
		assert.Equal(t, firstSpecID, modelSpec[0].ID)
		assert.Equal(t, secondSpecID, modelSpec[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestRepository_ListByReferenceObjectIDs(t *testing.T) {
	ExpectedLimit := 1
	ExpectedOffset := 0

	totalCountForFirstAPI := 1
	totalCountForSecondAPI := 1
	totalCountForFirstEvent := 1
	totalCountForSecondEvent := 1
	testErr := errors.New("test err")

	firstSpecID := "111111111-1111-1111-1111-111111111111"
	secondSpecID := "222222222-2222-2222-2222-222222222222"
	firstAPIID := "333333333-3333-3333-3333-333333333333"
	secondAPIID := "444444444-4444-4444-4444-444444444444"
	firstEventID := "333333333-3333-3333-3333-333333333333"
	secondEventID := "444444444-4444-4444-4444-444444444444"
	apiIDs := []string{firstAPIID, secondAPIID}
	eventIDs := []string{firstEventID, secondEventID}

	firstAPISpecEntity := fixAPISpecEntityWithIDs(firstSpecID, firstAPIID)
	secondAPISpecEntity := fixAPISpecEntityWithIDs(secondSpecID, secondAPIID)
	firstEventSpecEntity := fixEventSpecEntityWithIDs(firstSpecID, firstEventID)
	secondEventSpecEntity := fixEventSpecEntityWithIDs(secondSpecID, secondEventID)

	selectQueryAPIs := fmt.Sprintf(`^\(SELECT (.+) FROM public\.specifications
		WHERE %s AND api_def_id IS NOT NULL AND api_def_id = \$2 ORDER BY created_at ASC, id ASC LIMIT \$3 OFFSET \$4\) UNION
		\(SELECT (.+) FROM public\.specifications WHERE %s AND api_def_id IS NOT NULL AND api_def_id = \$6 ORDER BY created_at ASC, id ASC LIMIT \$7 OFFSET \$8\)`, fixTenantIsolationSubqueryWithArg(1), fixTenantIsolationSubqueryWithArg(5))

	countQueryAPIs := fmt.Sprintf(`SELECT api_def_id AS id, COUNT\(\*\) AS total_count FROM public.specifications WHERE %s AND api_def_id IS NOT NULL GROUP BY api_def_id ORDER BY api_def_id ASC`, fixTenantIsolationSubquery())

	selectQueryEvents := fmt.Sprintf(`^\(SELECT (.+) FROM public\.specifications
		WHERE %s AND event_def_id IS NOT NULL AND event_def_id = \$2 ORDER BY created_at ASC, id ASC LIMIT \$3 OFFSET \$4\) UNION
		\(SELECT (.+) FROM public\.specifications WHERE %s AND event_def_id IS NOT NULL AND event_def_id = \$6 ORDER BY created_at ASC, id ASC LIMIT \$7 OFFSET \$8\)`, fixTenantIsolationSubqueryWithArg(1), fixTenantIsolationSubqueryWithArg(5))

	countQueryEvents := fmt.Sprintf(`SELECT event_def_id AS id, COUNT\(\*\) AS total_count FROM public.specifications WHERE %s AND event_def_id IS NOT NULL GROUP BY event_def_id ORDER BY event_def_id ASC`, fixTenantIsolationSubquery())

	t.Run("Success for API", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixSpecColumns()).
			AddRow(fixAPISpecRowWithIDs(firstSpecID, firstAPIID)...).
			AddRow(fixAPISpecRowWithIDs(secondSpecID, secondAPIID)...)

		sqlMock.ExpectQuery(selectQueryAPIs).
			WithArgs(tenant, firstAPIID, ExpectedLimit, ExpectedOffset, tenant, secondAPIID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryAPIs).
			WithArgs(tenant).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstAPIID, totalCountForFirstAPI).
				AddRow(secondAPIID, totalCountForSecondAPI))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", firstAPISpecEntity).Return(*fixModelAPISpecWithIDs(firstSpecID, firstAPIID), nil)
		convMock.On("FromEntity", secondAPISpecEntity).Return(*fixModelAPISpecWithIDs(secondSpecID, secondAPIID), nil)
		pgRepository := spec.NewRepository(convMock)
		// WHEN
		modelSpec, err := pgRepository.ListByReferenceObjectIDs(ctx, tenant, model.APISpecReference, apiIDs)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelSpec, 2)
		assert.Equal(t, firstSpecID, modelSpec[0].ID)
		assert.Equal(t, secondSpecID, modelSpec[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("Success for Event", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixSpecColumns()).
			AddRow(fixEventSpecRowWithIDs(firstSpecID, firstEventID)...).
			AddRow(fixEventSpecRowWithIDs(secondSpecID, secondEventID)...)

		sqlMock.ExpectQuery(selectQueryEvents).
			WithArgs(tenant, firstEventID, ExpectedLimit, ExpectedOffset, tenant, secondEventID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryEvents).
			WithArgs(tenant).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstEventID, totalCountForFirstEvent).
				AddRow(secondEventID, totalCountForSecondEvent))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", firstEventSpecEntity).Return(*fixModelEventSpecWithIDs(firstSpecID, firstEventID), nil)
		convMock.On("FromEntity", secondEventSpecEntity).Return(*fixModelEventSpecWithIDs(secondSpecID, secondEventID), nil)
		pgRepository := spec.NewRepository(convMock)
		// WHEN
		modelSpec, err := pgRepository.ListByReferenceObjectIDs(ctx, tenant, model.EventSpecReference, eventIDs)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelSpec, 2)
		assert.Equal(t, firstSpecID, modelSpec[0].ID)
		assert.Equal(t, secondSpecID, modelSpec[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("Returns error when conversion from entity fails", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixSpecColumns()).
			AddRow(fixEventSpecRowWithIDs(firstSpecID, firstEventID)...).
			AddRow(fixEventSpecRowWithIDs(secondSpecID, secondEventID)...)

		sqlMock.ExpectQuery(selectQueryEvents).
			WithArgs(tenant, firstEventID, ExpectedLimit, ExpectedOffset, tenant, secondEventID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryEvents).
			WithArgs(tenant).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstEventID, totalCountForFirstEvent).
				AddRow(secondEventID, totalCountForSecondEvent))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", firstEventSpecEntity).Return(model.Spec{}, testErr)
		pgRepository := spec.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.ListByReferenceObjectIDs(ctx, tenant, model.EventSpecReference, eventIDs)
		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		sqlMock.ExpectQuery(selectQueryEvents).
			WithArgs(tenant, firstEventID, ExpectedLimit, ExpectedOffset, tenant, secondEventID, ExpectedLimit, ExpectedOffset).
			WillReturnError(testErr)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		pgRepository := spec.NewRepository(convMock)
		// WHEN
		modelSpecs, err := pgRepository.ListByReferenceObjectIDs(ctx, tenant, model.EventSpecReference, eventIDs)
		//THEN
		assert.Nil(t, modelSpecs)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}
*/

func TestRepository_Delete(t *testing.T) {
	apiSpecSuite := testdb.RepoDeleteTestSuite{
		Name: "API Spec Delete",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.specifications WHERE id = $1 AND (id IN (SELECT id FROM api_specifications_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{specID, tenant},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: spec.NewRepository,
		MethodArgs:          []interface{}{tenant, specID, model.APISpecReference},
	}

	eventSpecSuite := testdb.RepoDeleteTestSuite{
		Name: "Event Spec Delete",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.specifications WHERE id = $1 AND (id IN (SELECT id FROM event_specifications_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{specID, tenant},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: spec.NewRepository,
		MethodArgs:          []interface{}{tenant, specID, model.EventSpecReference},
	}

	apiSpecSuite.Run(t)
	eventSpecSuite.Run(t)
}

func TestRepository_DeleteByReferenceObjectID(t *testing.T) {
	apiSpecSuite := testdb.RepoDeleteTestSuite{
		Name: "API Spec DeleteByReferenceObjectID",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.specifications WHERE api_def_id = $1 AND (id IN (SELECT id FROM api_specifications_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{apiID, tenant},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: spec.NewRepository,
		MethodArgs:          []interface{}{tenant, model.APISpecReference, apiID},
		MethodName:          "DeleteByReferenceObjectID",
		IsDeleteMany:        true,
	}

	eventSpecSuite := testdb.RepoDeleteTestSuite{
		Name: "Event Spec DeleteByReferenceObjectID",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.specifications WHERE event_def_id = $1 AND (id IN (SELECT id FROM event_specifications_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{eventID, tenant},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: spec.NewRepository,
		MethodArgs:          []interface{}{tenant, model.EventSpecReference, eventID},
		MethodName:          "DeleteByReferenceObjectID",
		IsDeleteMany:        true,
	}

	apiSpecSuite.Run(t)
	eventSpecSuite.Run(t)
}

func TestRepository_Update(t *testing.T) {
	var nilSpecModel *model.Spec
	apiSpecModel := fixModelAPISpec()
	apiSpecEntity := fixAPISpecEntity()
	eventSpecModel := fixModelEventSpec()
	eventSpecEntity := fixEventSpecEntity()

	apiSpecSuite := testdb.RepoUpdateTestSuite{
		Name: "Update API Spec",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.specifications SET spec_data = ?, api_spec_format = ?, api_spec_type = ?, event_spec_format = ?, event_spec_type = ? WHERE id = ? AND (id IN (SELECT id FROM api_specifications_tenants WHERE tenant_id = '%s' AND owner = true))`, tenant)),
				Args:          []driver.Value{apiSpecEntity.SpecData, apiSpecEntity.APISpecFormat, apiSpecEntity.APISpecType, apiSpecEntity.EventSpecFormat, apiSpecEntity.EventSpecType, apiSpecEntity.ID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:       spec.NewRepository,
		ModelEntity:               apiSpecModel,
		DBEntity:                  apiSpecEntity,
		NilModelEntity:            nilSpecModel,
		TenantID:                  tenant,
		DisableConverterErrorTest: true,
	}

	eventSpecSuite := testdb.RepoUpdateTestSuite{
		Name: "Update Event Spec",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.specifications SET spec_data = ?, api_spec_format = ?, api_spec_type = ?, event_spec_format = ?, event_spec_type = ? WHERE id = ? AND (id IN (SELECT id FROM event_specifications_tenants WHERE tenant_id = '%s' AND owner = true))`, tenant)),
				Args:          []driver.Value{eventSpecEntity.SpecData, eventSpecEntity.APISpecFormat, eventSpecEntity.APISpecType, eventSpecEntity.EventSpecFormat, eventSpecEntity.EventSpecType, eventSpecEntity.ID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:       spec.NewRepository,
		ModelEntity:               eventSpecModel,
		DBEntity:                  eventSpecEntity,
		NilModelEntity:            nilSpecModel,
		TenantID:                  tenant,
		DisableConverterErrorTest: true,
	}

	apiSpecSuite.Run(t)
	eventSpecSuite.Run(t)
}

func TestRepository_Exists(t *testing.T) {
	apiSpecSuite := testdb.RepoExistTestSuite{
		Name: "API Specification Exists",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.specifications WHERE id = $1 AND (id IN (SELECT id FROM api_specifications_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{specID, tenant},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: spec.NewRepository,
		TargetID:            specID,
		TenantID:            tenant,
		RefEntity:           model.APISpecReference,
	}

	eventSpecSuite := testdb.RepoExistTestSuite{
		Name: "Event Specification Exists",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.specifications WHERE id = $1 AND (id IN (SELECT id FROM event_specifications_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{specID, tenant},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: spec.NewRepository,
		TargetID:            specID,
		TenantID:            tenant,
		RefEntity:           model.EventSpecReference,
	}

	apiSpecSuite.Run(t)
	eventSpecSuite.Run(t)
}
