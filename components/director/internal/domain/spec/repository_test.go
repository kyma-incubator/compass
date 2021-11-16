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

func TestRepository_ListByReferenceObjectID(t *testing.T) {
	apiSpecModel1 := fixModelAPISpecWithID("1")
	apiSpecModel2 := fixModelAPISpecWithID("2")
	apiSpecEntity1 := fixAPISpecEntityWithID("1")
	apiSpecEntity2 := fixAPISpecEntityWithID("2")

	apiSpecSuite := testdb.RepoListTestSuite{
		Name: "List API Specs By Ref Object ID",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type, event_spec_format, event_spec_type, custom_type FROM public.specifications WHERE api_def_id = $1 AND (id IN (SELECT id FROM api_specifications_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{apiID, tenant},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSpecColumns()).AddRow(fixAPISpecRowWithID("1")...).AddRow(fixAPISpecRowWithID("2")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSpecColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:   spec.NewRepository,
		ExpectedModelEntities: []interface{}{apiSpecModel1, apiSpecModel2},
		ExpectedDBEntities:    []interface{}{&apiSpecEntity1, &apiSpecEntity2},
		MethodArgs:            []interface{}{tenant, model.APISpecReference, apiID},
		MethodName:            "ListByReferenceObjectID",
	}

	eventSpecModel1 := fixModelEventSpecWithID("1")
	eventSpecModel2 := fixModelEventSpecWithID("2")
	eventSpecEntity1 := fixEventSpecEntityWithID("1")
	eventSpecEntity2 := fixEventSpecEntityWithID("2")

	eventSpecSuite := testdb.RepoListTestSuite{
		Name: "List Event Specs By Ref Object ID",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type, event_spec_format, event_spec_type, custom_type FROM public.specifications WHERE event_def_id = $1 AND (id IN (SELECT id FROM event_specifications_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{apiID, tenant},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSpecColumns()).AddRow(fixEventSpecRowWithID("1")...).AddRow(fixEventSpecRowWithID("2")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSpecColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:   spec.NewRepository,
		ExpectedModelEntities: []interface{}{eventSpecModel1, eventSpecModel2},
		ExpectedDBEntities:    []interface{}{&eventSpecEntity1, &eventSpecEntity2},
		MethodArgs:            []interface{}{tenant, model.EventSpecReference, apiID},
		MethodName:            "ListByReferenceObjectID",
	}

	apiSpecSuite.Run(t)
	eventSpecSuite.Run(t)
}

func TestRepository_ListByReferenceObjectIDs(t *testing.T) {
	firstFrID := "111111111-1111-1111-1111-111111111111"
	firstRefID := "refID1"
	secondFrID := "222222222-2222-2222-2222-222222222222"
	secondRefID := "refID2"

	apiSpecModel1 := fixModelAPISpecWithIDs(firstFrID, firstRefID)
	apiSpecModel2 := fixModelAPISpecWithIDs(secondFrID, secondRefID)
	apiSpecEntity1 := fixAPISpecEntityWithIDs(firstFrID, firstRefID)
	apiSpecEntity2 := fixAPISpecEntityWithIDs(secondFrID, secondRefID)

	apiSpecSuite := testdb.RepoListTestSuite{
		Name: "List API Specifications by Object IDs",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query: regexp.QuoteMeta(`(SELECT id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type, event_spec_format, event_spec_type, custom_type FROM public.specifications 
												WHERE api_def_id IS NOT NULL AND (id IN (SELECT id FROM api_specifications_tenants WHERE tenant_id = $1)) AND api_def_id = $2 ORDER BY created_at ASC, id ASC LIMIT $3 OFFSET $4)
 											   UNION
												(SELECT id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type, event_spec_format, event_spec_type, custom_type FROM public.specifications 
												WHERE api_def_id IS NOT NULL AND (id IN (SELECT id FROM api_specifications_tenants WHERE tenant_id = $5)) AND api_def_id = $6 ORDER BY created_at ASC, id ASC LIMIT $7 OFFSET $8)`),
				Args:     []driver.Value{tenant, firstRefID, 1, 0, tenant, secondRefID, 1, 0},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSpecColumns()).AddRow(fixAPISpecRowWithIDs(firstFrID, firstRefID)...).AddRow(fixAPISpecRowWithIDs(secondFrID, secondRefID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSpecColumns())}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT api_def_id AS id, COUNT(*) AS total_count FROM public.specifications WHERE api_def_id IS NOT NULL AND (id IN (SELECT id FROM api_specifications_tenants WHERE tenant_id = $1)) GROUP BY api_def_id ORDER BY api_def_id ASC`),
				Args:     []driver.Value{tenant},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"id", "total_count"}).AddRow(firstRefID, 1).AddRow(secondRefID, 1)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"id", "total_count"}).AddRow(firstRefID, 0).AddRow(secondRefID, 0)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:   spec.NewRepository,
		ExpectedModelEntities: []interface{}{apiSpecModel1, apiSpecModel2},
		ExpectedDBEntities:    []interface{}{&apiSpecEntity1, &apiSpecEntity2},
		MethodArgs:            []interface{}{tenant, model.APISpecReference, []string{firstRefID, secondRefID}},
		MethodName:            "ListByReferenceObjectIDs",
	}

	eventSpecModel1 := fixModelEventSpecWithIDs(firstFrID, firstRefID)
	eventSpecModel2 := fixModelEventSpecWithIDs(secondFrID, secondRefID)
	eventSpecEntity1 := fixEventSpecEntityWithIDs(firstFrID, firstRefID)
	eventSpecEntity2 := fixEventSpecEntityWithIDs(secondFrID, secondRefID)

	eventSpecSuite := testdb.RepoListTestSuite{
		Name: "List Event Specifications by Object IDs",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query: regexp.QuoteMeta(`(SELECT id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type, event_spec_format, event_spec_type, custom_type FROM public.specifications 
												WHERE event_def_id IS NOT NULL AND (id IN (SELECT id FROM event_specifications_tenants WHERE tenant_id = $1)) AND event_def_id = $2 ORDER BY created_at ASC, id ASC LIMIT $3 OFFSET $4)
 											   UNION
												(SELECT id, api_def_id, event_def_id, spec_data, api_spec_format, api_spec_type, event_spec_format, event_spec_type, custom_type FROM public.specifications 
												WHERE event_def_id IS NOT NULL AND (id IN (SELECT id FROM event_specifications_tenants WHERE tenant_id = $5)) AND event_def_id = $6 ORDER BY created_at ASC, id ASC LIMIT $7 OFFSET $8)`),
				Args:     []driver.Value{tenant, firstRefID, 1, 0, tenant, secondRefID, 1, 0},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSpecColumns()).AddRow(fixEventSpecRowWithIDs(firstFrID, firstRefID)...).AddRow(fixEventSpecRowWithIDs(secondFrID, secondRefID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSpecColumns())}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT event_def_id AS id, COUNT(*) AS total_count FROM public.specifications WHERE event_def_id IS NOT NULL AND (id IN (SELECT id FROM event_specifications_tenants WHERE tenant_id = $1)) GROUP BY event_def_id ORDER BY event_def_id ASC`),
				Args:     []driver.Value{tenant},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"id", "total_count"}).AddRow(firstRefID, 1).AddRow(secondRefID, 1)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"id", "total_count"}).AddRow(firstRefID, 0).AddRow(secondRefID, 0)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:   spec.NewRepository,
		ExpectedModelEntities: []interface{}{eventSpecModel1, eventSpecModel2},
		ExpectedDBEntities:    []interface{}{&eventSpecEntity1, &eventSpecEntity2},
		MethodArgs:            []interface{}{tenant, model.EventSpecReference, []string{firstRefID, secondRefID}},
		MethodName:            "ListByReferenceObjectIDs",
	}

	apiSpecSuite.Run(t)
	eventSpecSuite.Run(t)
}

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
