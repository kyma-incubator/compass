package operation_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_ListAllByType(t *testing.T) {
	operationModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	operationEntity := fixEntityOperation(operationID, testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)

	suite := testdb.RepoListTestSuite{
		Name: "List operations by Type",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, op_type, status, data, error, error_severity, priority, created_at, updated_at FROM public.operation WHERE op_type = $1`),
				Args:  []driver.Value{model.OperationTypeOrdAggregation},
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(operationModel.ID, operationModel.OpType, operationModel.Status, operationModel.Data, operationModel.Error, operationModel.ErrorSeverity, operationModel.Priority, operationModel.CreatedAt, operationModel.UpdatedAt)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
				IsSelect: true,
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       operation.NewRepository,
		ExpectedModelEntities:     []interface{}{operationModel},
		ExpectedDBEntities:        []interface{}{operationEntity},
		MethodName:                "ListAllByType",
		MethodArgs:                []interface{}{model.OperationTypeOrdAggregation},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Create(t *testing.T) {
	var nilOperationModel *model.Operation
	operationModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	operationEntity := fixEntityOperation(operationID, testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Operation",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.operation \(.+\) VALUES \(.+\)$`,
				Args:        fixOperationCreateArgs(operationModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       operation.NewRepository,
		ModelEntity:               operationModel,
		DBEntity:                  operationEntity,
		NilModelEntity:            nilOperationModel,
		MethodName:                "Create",
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	var nilOperationModel *model.Operation
	operationModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	operationEntity := fixEntityOperation(operationID, testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Operation",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.operation SET status = ?, error = ?, error_severity = ?, priority = ?, updated_at = ? WHERE id = ?`),
				Args:          fixOperationUpdateArgs(operationModel),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       operation.NewRepository,
		ModelEntity:               operationModel,
		DBEntity:                  operationEntity,
		NilModelEntity:            nilOperationModel,
		DisableConverterErrorTest: true,
		UpdateMethodName:          "Update",
		IsGlobal:                  true,
	}

	suite.Run(t)
}

func TestPgRepository_Get(t *testing.T) {
	operationModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	operationEntity := fixEntityOperation(operationID, testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)

	suite := testdb.RepoGetTestSuite{
		Name: "Get Operation",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, op_type, status, data, error, error_severity, priority, created_at, updated_at FROM public.operation WHERE id = $1`),
				Args:  []driver.Value{operationID},
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(operationModel.ID, operationModel.OpType, operationModel.Status, operationModel.Data, operationModel.Error, operationModel.ErrorSeverity, operationModel.Priority, operationModel.CreatedAt, operationModel.UpdatedAt)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
				IsSelect: true,
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       operation.NewRepository,
		ExpectedModelEntity:       operationModel,
		ExpectedDBEntity:          operationEntity,
		MethodArgs:                []interface{}{operationID},
		DisableConverterErrorTest: true,
		MethodName:                "Get",
	}

	suite.Run(t)
}

func TestPgRepository_GetByDataAndType(t *testing.T) {
	operationModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	operationEntity := fixEntityOperation(operationID, testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)

	suite := testdb.RepoGetTestSuite{
		Name: "Get Operation by Data and Type",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, op_type, status, data, error, error_severity, priority, created_at, updated_at FROM public.operation WHERE data @> $1 AND op_type = $2`),
				Args:  []driver.Value{fixOperationDataAsString(applicationID, applicationTemplateID), model.OperationTypeOrdAggregation},
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(operationModel.ID, operationModel.OpType, operationModel.Status, operationModel.Data, operationModel.Error, operationModel.ErrorSeverity, operationModel.Priority, operationModel.CreatedAt, operationModel.UpdatedAt)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
				IsSelect: true,
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       operation.NewRepository,
		ExpectedModelEntity:       operationModel,
		ExpectedDBEntity:          operationEntity,
		MethodArgs:                []interface{}{fixOperationData(applicationID, applicationTemplateID), model.OperationTypeOrdAggregation},
		DisableConverterErrorTest: true,
		MethodName:                "GetByDataAndType",
	}

	suite.Run(t)
}
func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete Operation",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.operation WHERE id = $1`),
				Args:          []driver.Value{operationID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: operation.NewRepository,
		MethodName:          "Delete",
		MethodArgs:          []interface{}{operationID},
		IsDeleteMany:        false,
		IsGlobal:            true,
	}

	suite.Run(t)
}

func TestPgRepository_DeleteMultiple(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete multiple Operations",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.operation WHERE id IN ($1)`),
				Args:          []driver.Value{operationID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: operation.NewRepository,
		MethodName:          "DeleteMultiple",
		MethodArgs:          []interface{}{[]string{operationID}},
		IsDeleteMany:        true,
		IsGlobal:            true,
	}

	suite.Run(t)
}

func TestRepository_PriorityQueueListByType(t *testing.T) {
	operationModel := fixOperationModel(model.OperationTypeOrdAggregation, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	operationEntity := fixEntityOperation(operationID, model.OperationTypeOrdAggregation, model.OperationStatusScheduled, model.OperationErrorSeverityNone)

	suite := testdb.RepoListTestSuite{
		Name: "PriorityQueue ListByType",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, op_type, status, data, error, error_severity, priority, created_at, updated_at FROM public.scheduled_operations WHERE op_type = $1 LIMIT $2`),
				Args:     []driver.Value{string(model.OperationTypeOrdAggregation), 10},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(operationModel.ID, operationModel.OpType, operationModel.Status, operationModel.Data, operationModel.Error, operationModel.ErrorSeverity, operationModel.Priority, operationModel.CreatedAt, operationModel.UpdatedAt)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       operation.NewRepository,
		ExpectedModelEntities:     []interface{}{operationModel},
		ExpectedDBEntities:        []interface{}{operationEntity},
		MethodArgs:                []interface{}{10, model.OperationTypeOrdAggregation},
		MethodName:                "PriorityQueueListByType",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_LockOperation(t *testing.T) {
	lockID := int64(491666746554389322)
	opID := "3a31599c-7a86-455d-83db-0014a7d459e8"

	t.Run("Success", func(t *testing.T) {
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		operationRepo := operation.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		expectedQuery := regexp.QuoteMeta("SELECT pg_try_advisory_xact_lock($1)")
		rows := sqlmock.NewRows([]string{"pg_try_advisory_xact_lock"}).AddRow(true)
		dbMock.ExpectQuery(expectedQuery).
			WithArgs(lockID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		// WHEN
		isLocked, err := operationRepo.LockOperation(ctx, opID)

		// THEN
		require.Equal(t, true, isLocked)
		require.NoError(t, err)
	})

	t.Run("Failed - could not acquire logs", func(t *testing.T) {
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		operationRepo := operation.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		expectedQuery := regexp.QuoteMeta("SELECT pg_try_advisory_xact_lock($1)")
		rows := sqlmock.NewRows([]string{"pg_try_advisory_xact_lock"}).AddRow(false)
		dbMock.ExpectQuery(expectedQuery).
			WithArgs(lockID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		// WHEN
		isLocked, err := operationRepo.LockOperation(ctx, opID)

		// THEN
		require.Equal(t, false, isLocked)
		require.NoError(t, err)
	})

	t.Run("Failed - empty operation id", func(t *testing.T) {
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		operationRepo := operation.NewRepository(mockConverter)

		// WHEN
		isLocked, err := operationRepo.LockOperation(context.TODO(), "")

		// THEN
		require.Equal(t, false, isLocked)
		require.NotNil(t, err)
	})
}

func TestRepository_RescheduleOperations(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		operationRepo := operation.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		expectedQuery := regexp.QuoteMeta("UPDATE public.operation SET status = $1, updated_at = $2 WHERE status IN ($3, $4) AND op_type = $5 AND updated_at < $6")
		dbMock.ExpectExec(expectedQuery).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)

		// WHEN
		err := operationRepo.RescheduleOperations(ctx, model.OperationTypeOrdAggregation, time.Second, []string{model.OperationStatusCompleted.ToString(), model.OperationStatusFailed.ToString()})

		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_DeleteOperations(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		operationRepo := operation.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		expectedQuery := regexp.QuoteMeta("DELETE FROM public.operation WHERE status = $1 AND op_type = $2 AND updated_at < $3")
		dbMock.ExpectExec(expectedQuery).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)

		// WHEN
		err := operationRepo.DeleteOperations(ctx, model.OperationTypeOrdAggregation, time.Second)

		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_RescheduleHangedOperations(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		operationRepo := operation.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		expectedQuery := regexp.QuoteMeta("UPDATE public.operation SET status = $1, updated_at = $2 WHERE status = $3 AND op_type = $4 AND updated_at < $5")
		dbMock.ExpectExec(expectedQuery).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)

		// WHEN
		err := operationRepo.RescheduleHangedOperations(ctx, model.OperationTypeOrdAggregation, time.Second)

		// THEN
		require.NoError(t, err)
	})
}
