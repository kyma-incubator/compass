package formation_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

var (
	formationEntity = fixFormationEntity()
	formationModel  = fixFormationModel()
)

func TestRepository_Create(t *testing.T) {
	suite := testdb.RepoCreateTestSuite{
		Name:       "Create Formation",
		MethodName: "Create",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.formations \(.+\) VALUES \(.+\)$`,
				Args:        []driver.Value{FormationID, Tnt, FormationTemplateID, testFormationName},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formation.NewRepository,
		ModelEntity:               formationModel,
		DBEntity:                  formationEntity,
		NilModelEntity:            nilFormationModel,
		IsGlobal:                  true,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_Get(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Formation by ID",
		MethodName: "Get",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, formation_template_id, name FROM public.formations WHERE tenant_id = $1 AND id = $2`),
				Args:     []driver.Value{Tnt, FormationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(FormationID, Tnt, FormationTemplateID, testFormationName)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formation.NewRepository,
		ExpectedModelEntity:       formationModel,
		ExpectedDBEntity:          formationEntity,
		MethodArgs:                []interface{}{FormationID, Tnt},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_GetByName(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Formation By Name",
		MethodName: "GetByName",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, formation_template_id, name FROM public.formations WHERE tenant_id = $1 AND name = $2`),
				Args:     []driver.Value{Tnt, testFormationName},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(FormationID, Tnt, FormationTemplateID, testFormationName)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formation.NewRepository,
		ExpectedModelEntity:       formationModel,
		ExpectedDBEntity:          formationEntity,
		MethodArgs:                []interface{}{testFormationName, Tnt},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_List(t *testing.T) {
	suite := testdb.RepoListPageableTestSuite{
		Name:       "List Formations ",
		MethodName: "List",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, formation_template_id, name FROM public.formations WHERE tenant_id = $1 ORDER BY id LIMIT 4 OFFSET 0`),
				Args:     []driver.Value{Tnt},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(FormationID, Tnt, FormationTemplateID, testFormationName)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT COUNT(*) FROM public.formations`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"count"}).AddRow(1)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: formation.NewRepository,
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: []interface{}{formationModel},
				ExpectedDBEntities:    []interface{}{formationEntity},
				ExpectedPage: &model.FormationPage{
					Data: []*model.Formation{formationModel},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
		},
		MethodArgs:                []interface{}{Tnt, 4, ""},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_Update(t *testing.T) {
	updateStmt := regexp.QuoteMeta(`UPDATE public.formations SET name = ? WHERE id = ? AND tenant_id = ?`)
	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Formation by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateStmt,
				Args:          []driver.Value{testFormationName, FormationID, Tnt},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formation.NewRepository,
		ModelEntity:               formationModel,
		DBEntity:                  formationEntity,
		NilModelEntity:            nilFormationModel,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete Formation by name",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.formations WHERE tenant_id = $1 AND name = $2`),
				Args:          []driver.Value{Tnt, testFormationName},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		RepoConstructorFunc: formation.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		IsGlobal:   true,
		MethodName: "DeleteByName",
		MethodArgs: []interface{}{Tnt, testFormationName},
	}

	suite.Run(t)
}

func TestRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Exists Formation by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.formations WHERE tenant_id = $1 AND id = $2`),
				Args:     []driver.Value{Tnt, FormationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		RepoConstructorFunc: formation.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		TargetID:   FormationID,
		TenantID:   Tnt,
		MethodName: "Exists",
		MethodArgs: []interface{}{FormationID, Tnt},
	}

	suite.Run(t)
}
