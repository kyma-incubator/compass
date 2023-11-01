package aspect_test

import (
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/aspect"
	"github.com/kyma-incubator/compass/components/director/internal/domain/aspect/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"regexp"
	"testing"
)

func TestPgRepository_Create(t *testing.T) {
	var nilAspectModel *model.Aspect
	aspectModel := fixAspectModel(aspectID)
	aspectEntity := fixEntityAspect(aspectID)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Aspect",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM integration_dependencies_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, integrationDependencyID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO public.aspects \(.+\) VALUES \(.+\)$`,
				Args:        fixAspectCreateArgs(aspectID, aspectModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.AspectConverter{}
		},
		RepoConstructorFunc:       aspect.NewRepository,
		ModelEntity:               aspectModel,
		DBEntity:                  aspectEntity,
		NilModelEntity:            nilAspectModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_DeleteByIntegrationDependencyID(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Aspect Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.aspects WHERE integration_dependency_id = $1 AND (id IN (SELECT id FROM aspects_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{integrationDependencyID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.AspectConverter{}
		},
		RepoConstructorFunc: aspect.NewRepository,
		MethodArgs:          []interface{}{tenantID, integrationDependencyID},
		MethodName:          "DeleteByIntegrationDependencyID",
		IsDeleteMany:        true,
	}

	suite.Run(t)
}
