package operation_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
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
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
	}

	suite.Run(t)
}
