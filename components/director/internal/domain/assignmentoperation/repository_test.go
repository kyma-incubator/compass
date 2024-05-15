package assignmentoperation_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/assignmentoperation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/assignmentoperation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

func TestRepository_Create(t *testing.T) {
	suite := testdb.RepoCreateTestSuite{
		Name:       "Create Formation",
		MethodName: "Create",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.assignment_operations \(.+\) VALUES \(.+\)$`,
				Args:        []driver.Value{operationID, operationType, assignmentID, formationID, operationTrigger, &defaultTime, &defaultTime},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       assignmentoperation.NewRepository,
		ModelEntity:               fixAssignmentOperationModel(),
		DBEntity:                  fixAssignmentOperationEntity(),
		NilModelEntity:            nilAssignmentOperationModel,
		IsGlobal:                  true,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_GetLatestOperation(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Latest Assignment Operation for the Formation Assignment in  the context of the Formation with the specified type",
		MethodName: "GetLatestOperation",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, type, formation_assignment_id, formation_id, triggered_by, started_at_timestamp, finished_at_timestamp FROM public.assignment_operations WHERE formation_assignment_id = $1 AND formation_id = $2 ORDER BY started_at_timestamp DESC`),
				Args:     []driver.Value{assignmentID, formationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(operationID, operationType, assignmentID, formationID, operationTrigger, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       assignmentoperation.NewRepository,
		ExpectedModelEntity:       fixAssignmentOperationModel(),
		ExpectedDBEntity:          fixAssignmentOperationEntity(),
		MethodArgs:                []interface{}{assignmentID, formationID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_ListByFormationIDs(t *testing.T) {
	emptyPageFAID := "empty-fa-id"
	onePageFAID := "one-fa-id"
	multiplePageFAID := "multiple-fa-id"

	assignmentOpModel1 := fixAssignmentOperationModelWithAssignmentID(onePageFAID)
	assignmentOpEntity1 := fixAssignmentOperationEntityWithAssignmentID(onePageFAID)

	assignmentOpModel2 := fixAssignmentOperationModelWithAssignmentID(multiplePageFAID)
	assignmentOpEntity2 := fixAssignmentOperationEntityWithAssignmentID(multiplePageFAID)

	pageSize := 1
	cursor := ""

	suite := testdb.RepoListPageableTestSuite{
		Name: "List Assignments Operations by Formation Assignment IDs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`(SELECT id, type, formation_assignment_id, formation_id, triggered_by, started_at_timestamp, finished_at_timestamp FROM public.assignment_operations WHERE formation_assignment_id = $1 ORDER BY formation_assignment_id ASC, started_at_timestamp DESC LIMIT $2 OFFSET $3)
				UNION 
				(SELECT id, type, formation_assignment_id, formation_id, triggered_by, started_at_timestamp, finished_at_timestamp FROM public.assignment_operations WHERE formation_assignment_id = $4 ORDER BY formation_assignment_id ASC, started_at_timestamp DESC LIMIT $5 OFFSET $6)
				UNION
				(SELECT id, type, formation_assignment_id, formation_id, triggered_by, started_at_timestamp, finished_at_timestamp FROM public.assignment_operations WHERE formation_assignment_id = $7 ORDER BY formation_assignment_id ASC, started_at_timestamp DESC LIMIT $8 OFFSET $9)`),
				Args:     []driver.Value{emptyPageFAID, pageSize, 0, onePageFAID, pageSize, 0, multiplePageFAID, pageSize, 0},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(assignmentOpEntity1.ID, assignmentOpEntity1.Type, assignmentOpEntity1.FormationAssignmentID, assignmentOpEntity1.FormationID, assignmentOpEntity1.TriggeredBy, assignmentOpEntity1.StartedAtTimestamp, assignmentOpEntity1.FinishedAtTimestamp).
						AddRow(assignmentOpEntity2.ID, assignmentOpEntity2.Type, assignmentOpEntity2.FormationAssignmentID, assignmentOpEntity2.FormationID, assignmentOpEntity2.TriggeredBy, assignmentOpEntity2.StartedAtTimestamp, assignmentOpEntity2.FinishedAtTimestamp),
					}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT formation_assignment_id AS id, COUNT(*) AS total_count FROM public.assignment_operations WHERE formation_assignment_id IN ($1, $2, $3) GROUP BY formation_assignment_id ORDER BY formation_assignment_id ASC`),
				Args:     []driver.Value{emptyPageFAID, onePageFAID, multiplePageFAID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"id", "total_count"}).AddRow(emptyPageFAID, 0).AddRow(onePageFAID, 1).AddRow(multiplePageFAID, 2)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: nil,
				ExpectedDBEntities:    nil,
				ExpectedPage: &model.AssignmentOperationPage{
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
				ExpectedModelEntities: []interface{}{assignmentOpModel1},
				ExpectedDBEntities:    []interface{}{assignmentOpEntity1},
				ExpectedPage: &model.AssignmentOperationPage{
					Data: []*model.AssignmentOperation{assignmentOpModel1},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
			{
				ExpectedModelEntities: []interface{}{assignmentOpModel2},
				ExpectedDBEntities:    []interface{}{assignmentOpEntity2},
				ExpectedPage: &model.AssignmentOperationPage{
					Data: []*model.AssignmentOperation{assignmentOpModel2},
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
		RepoConstructorFunc:       assignmentoperation.NewRepository,
		MethodArgs:                []interface{}{[]string{emptyPageFAID, onePageFAID, multiplePageFAID}, pageSize, cursor},
		MethodName:                "ListForFormationAssignmentIDs",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_Update(t *testing.T) {
	updateStmt := regexp.QuoteMeta(`UPDATE public.assignment_operations SET triggered_by = ?, started_at_timestamp = ?, finished_at_timestamp = ? WHERE id = ?`)
	suite := testdb.RepoUpdateTestSuite{
		IsGlobal: true,
		Name:     "Update Formation Assignment by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateStmt,
				Args:          []driver.Value{operationTrigger, &defaultTime, &defaultTime, operationID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       assignmentoperation.NewRepository,
		ModelEntity:               fixAssignmentOperationModel(),
		DBEntity:                  fixAssignmentOperationEntity(),
		NilModelEntity:            nilAssignmentOperationModel,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_DeleteByIDs(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name:         "Delete Assignment Operations by IDs",
		IsGlobal:     true,
		IsDeleteMany: true,
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.assignment_operations WHERE id IN ($1, $2)`),
				Args:          []driver.Value{operationID, operationID2},
				ValidResult:   sqlmock.NewResult(-1, 2),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		RepoConstructorFunc: assignmentoperation.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		MethodName: "DeleteByIDs",
		MethodArgs: []interface{}{[]string{operationID, operationID2}},
	}

	suite.Run(t)
}
