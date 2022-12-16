package formationassignment_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"

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
				Args:        []driver.Value{TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr},
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 AND id = $2`),
				Args:     []driver.Value{TestTenantID, TestID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr)}
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

func TestRepository_GetGlobalByID(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Formation Assignment Globally by ID",
		MethodName: "GetGlobalByID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE id = $1`),
				Args:     []driver.Value{TestID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr)}
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
		MethodArgs:                []interface{}{TestID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_GetGlobalByIDAndFormationID(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Formation Assignment Globally by ID and Formation ID",
		MethodName: "GetGlobalByIDAndFormationID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE id = $1 AND formation_id = $2`),
				Args:     []driver.Value{TestID, TestFormationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr)}
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
		MethodArgs:                []interface{}{TestID, TestFormationID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_GetForFormation(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Formation Assignment For Formation",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 AND id = $2 AND formation_id = $3`),
				Args:     []driver.Value{TestTenantID, TestID, TestFormationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr),
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

func TestRepository_GetBySourceAndTarget(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Formation Assignment by Source and Target",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id = $2 AND source = $3 AND target = $4`),
				Args:     []driver.Value{TestTenantID, TestFormationID, TestSource, TestTarget},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr),
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
		MethodArgs:                []interface{}{TestTenantID, TestFormationID, TestSource, TestTarget},
		MethodName:                "GetBySourceAndTarget",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_GetReverseBySourceAndTarget(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Reverse Formation Assignment by Source and Target",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id = $2 AND source = $3 AND target = $4`),
				Args:     []driver.Value{TestTenantID, TestFormationID, TestTarget, TestSource},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr),
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
		MethodArgs:                []interface{}{TestTenantID, TestFormationID, TestSource, TestTarget},
		MethodName:                "GetReverseBySourceAndTarget",
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 ORDER BY id LIMIT 4 OFFSET 0`),
				Args:     []driver.Value{TestTenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr)}
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
				Query: regexp.QuoteMeta(`(SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id = $2 ORDER BY formation_id ASC, id ASC LIMIT $3 OFFSET $4)
												UNION
												(SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE tenant_id = $5 AND formation_id = $6 ORDER BY formation_id ASC, id ASC LIMIT $7 OFFSET $8)
												UNION
												(SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE tenant_id = $9 AND formation_id = $10 ORDER BY formation_id ASC, id ASC LIMIT $11 OFFSET $12)`),
				Args:     []driver.Value{TestTenantID, emptyPageFormationID, pageSize, 0, TestTenantID, onePageFormationID, pageSize, 0, TestTenantID, multiplePageFormationID, pageSize, 0},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(faEntity1.ID, faEntity1.FormationID, faEntity1.TenantID, faEntity1.Source, faEntity1.SourceType, faEntity1.Target, faEntity1.TargetType, faEntity1.LastOperation, faEntity1.LastOperationInitiator, faEntity1.LastOperationInitiatorType, faEntity1.State, faEntity1.Value).
						AddRow(faEntity2.ID, faEntity2.FormationID, faEntity2.TenantID, faEntity2.Source, faEntity2.SourceType, faEntity2.Target, faEntity2.TargetType, faEntity2.LastOperation, faEntity2.LastOperationInitiator, faEntity2.LastOperationInitiatorType, faEntity2.State, faEntity2.Value),
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE (tenant_id = $1 AND (formation_id = $2 AND (source = $3 OR target = $4)))`),
				Args:     []driver.Value{TestTenantID, TestFormationID, TestSource, TestSource},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr)}
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

func TestRepository_ListAllForObjectIDs(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name:       "ListAllForObjectIDs Formations Assignments",
		MethodName: "ListAllForObjectIDs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE (tenant_id = $1 AND (formation_id = $2 AND (source IN ($3, $4) OR target IN ($5, $6)))`),
				Args:     []driver.Value{TestTenantID, TestFormationID, TestSource, TestTarget, TestSource, TestTarget},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr)}
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
		MethodArgs:                []interface{}{TestTenantID, TestFormationID, []string{TestSource, TestTarget}},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_Update(t *testing.T) {
	updateStmt := regexp.QuoteMeta(`UPDATE public.formation_assignments SET last_operation = ?, last_operation_initiator = ?, last_operation_initiator_type = ?, state = ?, value = ? WHERE id = ? AND tenant_id = ?`)
	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Formation Assignment by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateStmt,
				Args:          []driver.Value{model.AssignFormation, TestSource, TestSourceType, TestState, TestConfigValueStr, TestID, TestTenantID},
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
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		RepoConstructorFunc: formationassignment.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
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

func TestRepository_GetByTargetAndSource(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Formation Assignment by Target and Source",
		MethodName: "GetByTargetAndSource",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 AND target = $2 AND source = $3`),
				Args:     []driver.Value{TestTenantID, TestTarget, TestSource},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr)}
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
		MethodArgs:                []interface{}{TestTarget, TestSource, TestTenantID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_ListForIDs(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name:       "ListForIDs Formations Assignments",
		MethodName: "ListForIDs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 AND id IN ($2)`),
				Args:     []driver.Value{TestTenantID, TestSource},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr)}
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
		MethodArgs:                []interface{}{TestTenantID, []string{TestSource}},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)

	// Additional test - empty slice because test suite returns empty result given valid query
	t.Run("returns empty slice given no scenarios", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		repository := formationassignment.NewRepository(nil)

		// WHEN
		actual, err := repository.ListForIDs(ctx, TestTenantID, []string{})

		// THEN
		assert.NoError(t, err)
		assert.Nil(t, actual)
	})
}

func TestRepository_ListByFormationIDsNoPaging(t *testing.T) {
	testErr := errors.New("test error")

	t.Run("success", func(t *testing.T) {
		converterMock := &automock.EntityConverter{}
		defer converterMock.AssertExpectations(t)
		converterMock.On("FromEntity", faEntity).Return(faModel).Once()

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id IN ($2)`)).
			WithArgs(TestTenantID, TestFormationID).WillReturnRows(sqlmock.NewRows(fixColumns).
			AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, "assign", TestSource, TestSourceType, TestState, TestConfigValueStr))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		expected := [][]*model.FormationAssignment{{faModel}}
		repository := formationassignment.NewRepository(converterMock)

		// WHEN
		actual, err := repository.ListByFormationIDsNoPaging(ctx, TestTenantID, []string{TestFormationID})

		// THEN
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("success - returns empty slice given no scenarios", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		repository := formationassignment.NewRepository(nil)

		// WHEN
		actual, err := repository.ListByFormationIDsNoPaging(ctx, TestTenantID, []string{})

		// THEN
		assert.NoError(t, err)
		assert.Nil(t, actual)
	})

	t.Run("returns error when listing fails", func(t *testing.T) {
		converterMock := &automock.EntityConverter{}
		defer converterMock.AssertExpectations(t)

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		sqlMock.AssertExpectations(t)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, last_operation, last_operation_initiator, last_operation_initiator_type, state, value FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id IN ($2)`)).
			WithArgs(TestTenantID, TestFormationID).WillReturnError(testErr)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		repository := formationassignment.NewRepository(converterMock)

		// WHEN
		actual, err := repository.ListByFormationIDsNoPaging(ctx, TestTenantID, []string{TestFormationID})

		// THEN
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		assert.Nil(t, actual)
	})
}
