package operation_test

import (
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"regexp"
	"testing"
)

func TestPgRepository_Create(t *testing.T) {
	var nilOperationModel *model.Operation
	operationModel := fixOperationModel(ordOpType, model.OperationStatusScheduled)
	operationEntity := fixEntityOperation(operationID, ordOpType, model.OperationStatusScheduled)

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
	operationModel := fixOperationModel(ordOpType, model.OperationStatusScheduled)
	operationEntity := fixEntityOperation(operationID, ordOpType, model.OperationStatusScheduled)

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
	operationModel := fixOperationModel(ordOpType, model.OperationStatusScheduled)
	operationEntity := fixEntityOperation(operationID, ordOpType, model.OperationStatusScheduled)

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
