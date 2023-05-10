package tenantbusinesstype_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenantbusinesstype"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenantbusinesstype/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestRepository_Create(t *testing.T) {
	var nilTbtModel *model.TenantBusinessType
	tbtModel := fixModelTenantBusinessType(tbtID, tbtCode, tbtName)
	tbtEntity := fixEntityTenantBusinessType(tbtID, tbtCode, tbtName)

	suite := testdb.RepoCreateTestSuite{
		Name: "Generic Create Tenant Business Type",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.tenant_business_types \(.+\) VALUES \(.+\)$`,
				Args:        []driver.Value{tbtModel.ID, tbtModel.Code, tbtModel.Name},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       tenantbusinesstype.NewRepository,
		ModelEntity:               tbtModel,
		DBEntity:                  tbtEntity,
		NilModelEntity:            nilTbtModel,
		MethodName:                "Create",
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
	}

	suite.Run(t)
}

func TestRepository_GetByID(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get by id Tenant Business Type",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, code, name FROM public.tenant_business_types WHERE id = $1`),
				Args:     []driver.Value{tbtID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTbtColumns()).AddRow([]driver.Value{tbtID, tbtCode, tbtName}...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTbtColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       tenantbusinesstype.NewRepository,
		ExpectedModelEntity:       fixModelTenantBusinessType(tbtID, tbtCode, tbtName),
		ExpectedDBEntity:          fixEntityTenantBusinessType(tbtID, tbtCode, tbtName),
		DisableConverterErrorTest: true,
		MethodName:                "GetByID",
		MethodArgs:                []interface{}{tbtID},
	}

	suite.Run(t)
}

func TestPgRepository_ListAll(t *testing.T) {
	tbt1ID := "aec0e9c5-06da-4625-9f8a-bda17ab8c3b9"
	tbt2ID := "ccdbef8f-b97a-490c-86e2-2bab2862a6e4"
	tbtEntity1 := fixEntityTenantBusinessType(tbt1ID, tbtCode, tbtName)
	tbtEntity2 := fixEntityTenantBusinessType(tbt2ID, "test-code-2", "test-name-2")

	appModel1 := fixModelTenantBusinessType(tbt1ID, tbtCode, tbtName)
	appModel2 := fixModelTenantBusinessType(tbt2ID, "test-code-2", "test-name-2")

	suite := testdb.RepoListTestSuite{
		Name: "List all Tenant Business Types",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, code, name FROM public.tenant_business_types`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTbtColumns()).
						AddRow(tbtEntity1.ID, tbtEntity1.Code, tbtEntity1.Name).
						AddRow(tbtEntity2.ID, tbtEntity2.Code, tbtEntity2.Name),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTbtColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       tenantbusinesstype.NewRepository,
		ExpectedModelEntities:     []interface{}{appModel1, appModel2},
		ExpectedDBEntities:        []interface{}{tbtEntity1, tbtEntity2},
		MethodName:                "ListAll",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}
