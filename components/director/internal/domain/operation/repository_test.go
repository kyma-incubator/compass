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

func TestPgRepository_Create(t *testing.T) {
	var nilOperationModel *model.Operation
	operationModel := fixOperationModel(testOpType, model.OperationStatusScheduled)
	operationEntity := fixEntityOperation(operationID, testOpType, model.OperationStatusScheduled)

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
	operationModel := fixOperationModel(testOpType, model.OperationStatusScheduled)
	operationEntity := fixEntityOperation(operationID, testOpType, model.OperationStatusScheduled)

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Operation",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.operation SET status = ?, error = ?, priority = ?, updated_at = ? WHERE id = ?`),
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
	operationModel := fixOperationModel(testOpType, model.OperationStatusScheduled)
	operationEntity := fixEntityOperation(operationID, testOpType, model.OperationStatusScheduled)

	suite := testdb.RepoGetTestSuite{
		Name: "Get Operation",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, op_type, status, data, error, priority, created_at, updated_at FROM public.operation WHERE id = $1`),
				Args:  []driver.Value{operationID},
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(operationModel.ID, operationModel.OpType, operationModel.Status, operationModel.Data, operationModel.Error, operationModel.Priority, operationModel.CreatedAt, operationModel.UpdatedAt)}
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
	operationModel := fixOperationModel(model.OperationTypeOrdAggregation, model.OperationStatusScheduled)
	operationEntity := fixEntityOperation(operationID, model.OperationTypeOrdAggregation, model.OperationStatusScheduled)

	suite := testdb.RepoListTestSuite{
		Name: "PriorityQueue ListByType",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, op_type, status, data, error, priority, created_at, updated_at FROM public.scheduled_operations WHERE op_type = $1 LIMIT $2`),
				Args:     []driver.Value{string(model.OperationTypeOrdAggregation), 10},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(operationModel.ID, operationModel.OpType, operationModel.Status, operationModel.Data, operationModel.Error, operationModel.Priority, operationModel.CreatedAt, operationModel.UpdatedAt)}
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
		expectedQuery := regexp.QuoteMeta("UPDATE public.operation SET status = $1, updated_at = $2 WHERE status IN ($3, $4) AND updated_at < $5")
		dbMock.ExpectExec(expectedQuery).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)

		// WHEN
		err := operationRepo.RescheduleOperations(ctx, time.Second)

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
		expectedQuery := regexp.QuoteMeta("UPDATE public.operation SET status = $1, updated_at = $2 WHERE status = $3 AND updated_at < $4")
		dbMock.ExpectExec(expectedQuery).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)

		// WHEN
		err := operationRepo.RescheduleHangedOperations(ctx, time.Second)

		// THEN
		require.NoError(t, err)
	})
}
