package formationassignment_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/stretchr/testify/require"

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
				Args:        []driver.Value{TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime},
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
		ModelEntity:               faModelWithConfigAndError,
		DBEntity:                  faEntityWithConfigAndError,
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $1 AND id = $2`),
				Args:     []driver.Value{TestTenantID, TestID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)}
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
		ExpectedModelEntity:       faModelWithConfigAndError,
		ExpectedDBEntity:          faEntityWithConfigAndError,
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE id = $1`),
				Args:     []driver.Value{TestID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)}
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
		ExpectedModelEntity:       faModelWithConfigAndError,
		ExpectedDBEntity:          faEntityWithConfigAndError,
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE id = $1 AND formation_id = $2`),
				Args:     []driver.Value{TestID, TestFormationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)}
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
		ExpectedModelEntity:       faModelWithConfigAndError,
		ExpectedDBEntity:          faEntityWithConfigAndError,
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $1 AND id = $2 AND formation_id = $3`),
				Args:     []driver.Value{TestTenantID, TestID, TestFormationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime),
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
		ExpectedModelEntity:       faModelWithConfigAndError,
		ExpectedDBEntity:          faEntityWithConfigAndError,
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id = $2 AND source = $3 AND target = $4`),
				Args:     []driver.Value{TestTenantID, TestFormationID, TestSource, TestTarget},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime),
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
		ExpectedModelEntity:       faModelWithConfigAndError,
		ExpectedDBEntity:          faEntityWithConfigAndError,
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id = $2 AND source = $3 AND target = $4`),
				Args:     []driver.Value{TestTenantID, TestFormationID, TestTarget, TestSource},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime),
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
		ExpectedModelEntity:       faModelWithConfigAndError,
		ExpectedDBEntity:          faEntityWithConfigAndError,
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $1 ORDER BY id LIMIT 4 OFFSET 0`),
				Args:     []driver.Value{TestTenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)}
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
				ExpectedModelEntities: []interface{}{faModelWithConfigAndError},
				ExpectedDBEntities:    []interface{}{faEntityWithConfigAndError},
				ExpectedPage: &model.FormationAssignmentPage{
					Data: []*model.FormationAssignment{faModelWithConfigAndError},
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
	faModel1.Error = TestErrorValueRawJSON
	faEntity1 := fixFormationAssignmentEntityWithFormationID(onePageFormationID)
	faEntity1.Error = repo.NewValidNullableString(TestErrorValueStr)

	faModel2 := fixFormationAssignmentModelWithFormationID(multiplePageFormationID)
	faModel2.Error = TestErrorValueRawJSON
	faEntity2 := fixFormationAssignmentEntityWithFormationID(multiplePageFormationID)
	faEntity2.Error = repo.NewValidNullableString(TestErrorValueStr)

	pageSize := 1
	cursor := ""

	suite := testdb.RepoListPageableTestSuite{
		Name: "List Formation Assignments by Formation IDs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`(SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id = $2 ORDER BY formation_id ASC, id ASC LIMIT $3 OFFSET $4)
												UNION
												(SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $5 AND formation_id = $6 ORDER BY formation_id ASC, id ASC LIMIT $7 OFFSET $8)
												UNION
												(SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $9 AND formation_id = $10 ORDER BY formation_id ASC, id ASC LIMIT $11 OFFSET $12)`),
				Args:     []driver.Value{TestTenantID, emptyPageFormationID, pageSize, 0, TestTenantID, onePageFormationID, pageSize, 0, TestTenantID, multiplePageFormationID, pageSize, 0},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(faEntity1.ID, faEntity1.FormationID, faEntity1.TenantID, faEntity1.Source, faEntity1.SourceType, faEntity1.Target, faEntity1.TargetType, faEntity1.State, faEntity1.Value, faEntity1.Error, faEntity1.LastStateChangeTimestamp, faEntity1.LastNotificationSentTimestamp).
						AddRow(faEntity2.ID, faEntity2.FormationID, faEntity2.TenantID, faEntity2.Source, faEntity2.SourceType, faEntity2.Target, faEntity2.TargetType, faEntity2.State, faEntity2.Value, faEntity2.Error, faEntity2.LastStateChangeTimestamp, faEntity2.LastNotificationSentTimestamp),
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE (tenant_id = $1 AND (formation_id = $2 AND (source = $3 OR target = $4)))`),
				Args:     []driver.Value{TestTenantID, TestFormationID, TestSource, TestSource},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		ExpectedModelEntities:     []interface{}{faModelWithConfigAndError},
		ExpectedDBEntities:        []interface{}{faEntityWithConfigAndError},
		RepoConstructorFunc:       formationassignment.NewRepository,
		MethodArgs:                []interface{}{TestTenantID, TestFormationID, TestSource},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_ListAllForObjectGlobal(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name:       "List All Formations Assignments for object globally",
		MethodName: "ListAllForObjectGlobal",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE (source = $1 OR target = $2)`),
				Args:     []driver.Value{TestSource, TestSource},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		ExpectedModelEntities:     []interface{}{faModelWithConfigAndError},
		ExpectedDBEntities:        []interface{}{faEntityWithConfigAndError},
		RepoConstructorFunc:       formationassignment.NewRepository,
		MethodArgs:                []interface{}{TestSource},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_DeleteAssignmentsForObjectID(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name:       "DeleteAssignmentsForObjectID Formations Assignments",
		MethodName: "DeleteAssignmentsForObjectID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.formation_assignments WHERE (tenant_id = $1 AND (formation_id = $2 AND (source = $3 OR target = $4)))`),
				Args:          []driver.Value{TestTenantID, TestFormationID, TestSource, TestSource},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		IsDeleteMany:        true,
		RepoConstructorFunc: formationassignment.NewRepository,
		MethodArgs:          []interface{}{TestTenantID, TestFormationID, TestSource},
	}

	suite.Run(t)
}

func TestRepository_ListAllForObjectIDs(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name:       "ListAllForObjectIDs Formations Assignments",
		MethodName: "ListAllForObjectIDs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE (tenant_id = $1 AND (formation_id = $2 AND (source IN ($3, $4) OR target IN ($5, $6)))`),
				Args:     []driver.Value{TestTenantID, TestFormationID, TestSource, TestTarget, TestSource, TestTarget},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		ExpectedModelEntities:     []interface{}{faModelWithConfigAndError},
		ExpectedDBEntities:        []interface{}{faEntityWithConfigAndError},
		RepoConstructorFunc:       formationassignment.NewRepository,
		MethodArgs:                []interface{}{TestTenantID, TestFormationID, []string{TestSource, TestTarget}},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)

	t.Run("returns empty slice given no object IDs", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		repository := formationassignment.NewRepository(nil)

		// WHEN
		actual, err := repository.ListAllForObjectIDs(ctx, "", "", []string{})

		// THEN
		assert.NoError(t, err)
		assert.Nil(t, actual)
	})
}

func TestRepository_Update(t *testing.T) {
	updateStmt := regexp.QuoteMeta(`UPDATE public.formation_assignments SET state = ?, value = ?, error = ?, last_state_change_timestamp = ?, last_notification_sent_timestamp = ? WHERE id = ? AND tenant_id = ?`)
	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Formation Assignment by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE id = $1`),
				Args:     []driver.Value{TestID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
			{
				Query:         updateStmt,
				Args:          []driver.Value{TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime, TestID, TestTenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationassignment.NewRepository,
		ModelEntity:               faModelWithConfigAndError,
		DBEntity:                  faEntityWithConfigAndError,
		NilModelEntity:            nilFormationAssignmentModel,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)

	t.Run("Success when the formation assignment state is changed and timestamp is updated", func(t *testing.T) {
		// GIVEN
		faModelWithReadyState := fixFormationAssignmentModelWithConfigAndError(TestConfigValueRawJSON, TestErrorValueRawJSON)
		faModelWithReadyState.State = readyAssignmentState

		faEntityWithReadyState := fixFormationAssignmentEntityWithConfigurationAndError(TestConfigValueStr, TestErrorValueStr)
		faEntityWithReadyState.State = readyAssignmentState

		slqxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)
		ctx := persistence.SaveToContext(emptyCtx, slqxDB)

		rows := sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE id = $1`)).
			WithArgs(TestID).WillReturnRows(rows)

		sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.formation_assignments SET state = ?, value = ?, error = ?, last_state_change_timestamp = ?, last_notification_sent_timestamp = ? WHERE id = ? AND tenant_id = ?`)).
			WithArgs(readyAssignmentState, TestConfigValueStr, TestErrorValueStr, sqlmock.AnyArg(), &defaultTime, TestID, TestTenantID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", faModelWithReadyState).Return(faEntityWithReadyState)

		r := formationassignment.NewRepository(mockConverter)

		// WHEN
		err := r.Update(ctx, faModelWithReadyState)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Success when the formation assignment state is CONFIG_PENDING but the configuration is changed, last state change timestamp should be updated", func(t *testing.T) {
		//t.Run("test", func(t *testing.T) {
		// GIVEN
		faModelWithConfigPendingState := fixFormationAssignmentModelWithConfigAndError(TestConfigValueRawJSON, TestErrorValueRawJSON)
		faModelWithConfigPendingState.State = configPendingAssignmentState

		faEntityWithConfigPendingState := fixFormationAssignmentEntityWithConfigurationAndError(TestConfigValueStr, TestErrorValueStr)
		faEntityWithConfigPendingState.State = configPendingAssignmentState

		slqxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)
		ctx := persistence.SaveToContext(emptyCtx, slqxDB)

		rows := sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, configPendingAssignmentState, TestNewConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE id = $1`)).
			WithArgs(TestID).WillReturnRows(rows)

		sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.formation_assignments SET state = ?, value = ?, error = ?, last_state_change_timestamp = ?, last_notification_sent_timestamp = ? WHERE id = ? AND tenant_id = ?`)).
			WithArgs(configPendingAssignmentState, TestConfigValueStr, TestErrorValueStr, sqlmock.AnyArg(), &defaultTime, TestID, TestTenantID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", faModelWithConfigPendingState).Return(faEntityWithConfigPendingState)

		r := formationassignment.NewRepository(mockConverter)

		// WHEN
		err := r.Update(ctx, faModelWithConfigPendingState)

		// THEN
		require.NoError(t, err)
	})
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id = $2 AND target = $3 AND source = $4`),
				Args:     []driver.Value{TestTenantID, TestFormationID, TestTarget, TestSource},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)}
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
		ExpectedModelEntity:       faModelWithConfigAndError,
		ExpectedDBEntity:          faEntityWithConfigAndError,
		MethodArgs:                []interface{}{TestTarget, TestSource, TestTenantID, TestFormationID},
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
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $1 AND id IN ($2)`),
				Args:     []driver.Value{TestTenantID, TestSource},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		ExpectedModelEntities:     []interface{}{faModelWithConfigAndError},
		ExpectedDBEntities:        []interface{}{faEntityWithConfigAndError},
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
		converterMock.On("FromEntity", faEntityWithConfigAndError).Return(faModelWithConfigAndError).Once()

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id IN ($2)`)).
			WithArgs(TestTenantID, TestFormationID).WillReturnRows(sqlmock.NewRows(fixColumns).
			AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		expected := [][]*model.FormationAssignment{{faModelWithConfigAndError}}
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
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id IN ($2)`)).
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

func TestRepository_GetAssignmentsForFormationWithStates(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name: "GetAssignmentsForFormationWithStates",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id = $2 AND state IN ($3, $4)`),
				Args:     []driver.Value{TestTenantID, TestFormationID, TestStateInitial, readyAssignmentState},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		ExpectedModelEntities:     []interface{}{faModelWithConfigAndError},
		ExpectedDBEntities:        []interface{}{faEntityWithConfigAndError},
		RepoConstructorFunc:       formationassignment.NewRepository,
		MethodArgs:                []interface{}{TestTenantID, TestFormationID, []string{TestStateInitial, readyAssignmentState}},
		MethodName:                "GetAssignmentsForFormationWithStates",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)

	// Additional test - empty slice because test suite returns empty result given valid query
	t.Run("returns empty slice given no scenarios", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		repository := formationassignment.NewRepository(nil)

		// WHEN
		actual, err := repository.GetAssignmentsForFormationWithStates(ctx, TestTenantID, TestFormationID, []string{})

		// THEN
		assert.NoError(t, err)
		assert.Nil(t, actual)
	})
}

func TestRepository_GetAssignmentsForFormation(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name: "GetAssignmentsForFormation",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, formation_id, tenant_id, source, source_type, target, target_type, state, value, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formation_assignments WHERE tenant_id = $1 AND formation_id = $2`),
				Args:     []driver.Value{TestTenantID, TestFormationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(TestID, TestFormationID, TestTenantID, TestSource, TestSourceType, TestTarget, TestTargetType, TestStateInitial, TestConfigValueStr, TestErrorValueStr, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		ExpectedModelEntities:     []interface{}{faModelWithConfigAndError},
		ExpectedDBEntities:        []interface{}{faEntityWithConfigAndError},
		RepoConstructorFunc:       formationassignment.NewRepository,
		MethodArgs:                []interface{}{TestTenantID, TestFormationID},
		MethodName:                "GetAssignmentsForFormation",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)

	// Additional test - empty slice because test suite returns empty result given valid query
	t.Run("returns empty slice given no scenarios", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		repository := formationassignment.NewRepository(nil)

		// WHEN
		actual, err := repository.GetAssignmentsForFormationWithStates(ctx, TestTenantID, TestFormationID, []string{})

		// THEN
		assert.NoError(t, err)
		assert.Nil(t, actual)
	})
}
