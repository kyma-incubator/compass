package formationassignment_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

func TestRepository_Create(t *testing.T) {
	suite := testdb.RepoCreateTestSuite{
		Name:       "Create Formation Assignment",
		MethodName: "Create",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.formation_assignments \(.+\) VALUES \(.+\)$`,
				Args:        []driver.Value{TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestState, TestConfigValueStr},
				ValidResult: sqlmock.NewResult(-1, 1),
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationassignment.NewRepository,
		ModelEntity:               faModel,
		DBEntity:                  faEntity,
		NilModelEntity:            nilFormationAssignmentModel,
		IsGlobal:                  true,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_Get(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Formation Assignment by ID",
		MethodName: "Get",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 AND id = $2`),
				Args:     []driver.Value{TestTenantID, TestID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestState, TestConfigValueStr)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationassignment.NewRepository,
		ExpectedModelEntity:       fixFormationAssignmentModel(TestConfigValueRawJSON),
		ExpectedDBEntity:          fixFormationAssignmentEntity(TestConfigValueStr),
		MethodArgs:                []interface{}{TestID, TestTenantID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_GetForFormation(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Formation Assignment For Formation",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 AND id = $2 AND formation_id = $3`),
				Args:     []driver.Value{TestTenantID, TestID, TestFormationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestState, TestConfigValueStr),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationassignment.NewRepository,
		ExpectedModelEntity:       faModel,
		ExpectedDBEntity:          faEntity,
		MethodArgs:                []interface{}{TestTenantID, TestID, TestFormationID},
		MethodName:                "GetForFormation",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_List(t *testing.T) {
	suite := testdb.RepoListPageableTestSuite{
		Name:       "List Formations Assignments",
		MethodName: "List",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 ORDER BY id LIMIT 4 OFFSET 0`),
				Args:     []driver.Value{TestTenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestState, TestConfigValueStr)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT COUNT(*) FROM public.formation_assignments`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"count"}).AddRow(1)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: formationassignment.NewRepository,
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: []interface{}{faModel},
				ExpectedDBEntities:    []interface{}{faEntity},
				ExpectedPage: &model.FormationAssignmentPage{
					Data: []*model.FormationAssignment{faModel},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
		},
		MethodArgs:                []interface{}{4, "", TestTenantID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_ListByFormationIDs(t *testing.T) {
	emptyPageFormationID := "empty-formation-id"
	onePageFormationID := "one-formation-id"
	multiplePageFormationID := "multiple-formation-id"

	faModel1 := fixFormationAssignmentModelWithFormationID(onePageFormationID)
	faEntity1 := fixFormationAssignmentEntityWithFormationID(onePageFormationID)

	faModel2 := fixFormationAssignmentModelWithFormationID(multiplePageFormationID)
	faEntity2 := fixFormationAssignmentEntityWithFormationID(multiplePageFormationID)

	pageSize := 1
	cursor := ""

	suite := testdb.RepoListPageableTestSuite{
		Name: "List Formation Assignments by Formation IDs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`(SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id = $2 ORDER BY formation_id ASC, id ASC LIMIT $3 OFFSET $4)
												UNION
												(SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value FROM public.formation_assignments WHERE tenant_id = $5 AND formation_id = $6 ORDER BY formation_id ASC, id ASC LIMIT $7 OFFSET $8)
												UNION
												(SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value FROM public.formation_assignments WHERE tenant_id = $9 AND formation_id = $10 ORDER BY formation_id ASC, id ASC LIMIT $11 OFFSET $12)`),
				Args:     []driver.Value{TestTenantID, emptyPageFormationID, pageSize, 0, TestTenantID, onePageFormationID, pageSize, 0, TestTenantID, multiplePageFormationID, pageSize, 0},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(faEntity1.ID, faEntity1.FormationID, faEntity1.TenantID, faEntity1.Source, faEntity1.SourceType, faEntity1.Target, faEntity1.TargetType, faEntity1.State, faEntity1.Value).
						AddRow(faEntity2.ID, faEntity2.FormationID, faEntity2.TenantID, faEntity2.Source, faEntity2.SourceType, faEntity2.Target, faEntity2.TargetType, faEntity2.State, faEntity2.Value),
					}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT formation_id AS id, COUNT(*) AS total_count FROM public.formation_assignments WHERE tenant_id = $1 GROUP BY formation_id ORDER BY formation_id ASC`),
				Args:     []driver.Value{TestTenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"id", "total_count"}).AddRow(emptyPageFormationID, 0).AddRow(onePageFormationID, 1).AddRow(multiplePageFormationID, 2)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: nil,
				ExpectedDBEntities:    nil,
				ExpectedPage: &model.FormationAssignmentPage{
					Data: nil,
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 0,
				},
			},
			{
				ExpectedModelEntities: []interface{}{faModel1},
				ExpectedDBEntities:    []interface{}{faEntity1},
				ExpectedPage: &model.FormationAssignmentPage{
					Data: []*model.FormationAssignment{faModel1},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
			{
				ExpectedModelEntities: []interface{}{faModel2},
				ExpectedDBEntities:    []interface{}{faEntity2},
				ExpectedPage: &model.FormationAssignmentPage{
					Data: []*model.FormationAssignment{faModel2},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   pagination.EncodeNextOffsetCursor(0, pageSize),
						HasNextPage: true,
					},
					TotalCount: 2,
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationassignment.NewRepository,
		MethodArgs:                []interface{}{TestTenantID, []string{emptyPageFormationID, onePageFormationID, multiplePageFormationID}, pageSize, cursor},
		MethodName:                "ListByFormationIDs",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_ListAllForObject(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name:       "ListAllForObject Formations Assignments",
		MethodName: "ListAllForObject",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value FROM public.formation_assignments WHERE (tenant_id = $1 AND (formation_id = $2 AND (source = $3 OR target = $4)))`),
				Args:     []driver.Value{TestTenantID, TestFormationID, TestSource, TestSource},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestState, TestConfigValueStr)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		ExpectedModelEntities:     []interface{}{faModel},
		ExpectedDBEntities:        []interface{}{faEntity},
		RepoConstructorFunc:       formationassignment.NewRepository,
		MethodArgs:                []interface{}{TestTenantID, TestFormationID, TestSource},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_Update(t *testing.T) {
	updateStmt := regexp.QuoteMeta(`UPDATE public.formation_assignments SET state = ?, value = ? WHERE id = ? AND tenant_id = ?`)
	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Formation Assignment by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateStmt,
				Args:          []driver.Value{TestState, TestConfigValueStr, TestID, TestTenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationassignment.NewRepository,
		ModelEntity:               faModel,
		DBEntity:                  faEntity,
		NilModelEntity:            nilFormationAssignmentModel,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete Formation Assignment by id",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.formation_assignments WHERE tenant_id = $1 AND id = $2`),
				Args:          []driver.Value{TestTenantID, TestID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		RepoConstructorFunc: formationassignment.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		IsGlobal:   true,
		MethodName: "Delete",
		MethodArgs: []interface{}{TestID, TestTenantID},
	}

	suite.Run(t)
}

func TestRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Exists Formation Assignment by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.formation_assignments WHERE tenant_id = $1 AND id = $2`),
				Args:     []driver.Value{TestTenantID, TestID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		RepoConstructorFunc: formationassignment.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		TargetID:   TestID,
		TenantID:   TestTenantID,
		MethodName: "Exists",
		MethodArgs: []interface{}{TestID, TestTenantID},
	}

	suite.Run(t)
}
