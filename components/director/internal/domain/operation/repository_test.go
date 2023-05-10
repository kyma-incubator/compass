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
	"time"
)

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
	var nilOperationModel *model.Operation
	operationModel := fixOperationModel(ordOpType, scheduledOpStatus)
	operationEntity := fixEntityOperation(operationID, ordOpType, scheduledOpStatus)

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
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
	}

	suite.Run(t)
}

func TestPgRepository_DeleteOlderThan(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "DeleteOlderThan Operation",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.operation WHERE finished_at IS NOT NULL AND op_type = $1 AND status = $2 AND finished_at < $3`),
				Args:          []driver.Value{ordOpType, scheduledOpStatus, time.Time{}},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: operation.NewRepository,
		MethodName:          "DeleteOlderThan",
		MethodArgs:          []interface{}{ordOpType, scheduledOpStatus, time.Time{}},
		IsDeleteMany:        true,
		IsGlobal:            true,
	}

	suite.Run(t)
}
