package formationtemplate_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestRepository_Get(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Formation Template By ID",
		MethodName: "Get",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, application_types, runtime_types, missing_artifact_info_message, missing_artifact_warning_message FROM public.formation_templates WHERE id = $1`),
				Args:     []driver.Value{testID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(formationTemplateEntity.ID, formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.MissingArtifactInfoMessage, formationTemplateEntity.MissingArtifactWarningMessage)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationtemplate.NewRepository,
		ExpectedModelEntity:       &formationTemplateModel,
		ExpectedDBEntity:          &formationTemplateEntity,
		MethodArgs:                []interface{}{testID},
		DisableConverterErrorTest: false,
	}

	suite.Run(t)
}

func TestRepository_Create(t *testing.T) {
	suite := testdb.RepoCreateTestSuite{
		Name:       "Create Formation Template",
		MethodName: "Create",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.formation_templates \(.+\) VALUES \(.+\)$`,
				Args:        []driver.Value{formationTemplateEntity.ID, formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.MissingArtifactInfoMessage, formationTemplateEntity.MissingArtifactWarningMessage},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationtemplate.NewRepository,
		ModelEntity:               &formationTemplateModel,
		DBEntity:                  &formationTemplateEntity,
		NilModelEntity:            nilModelEntity,
		IsGlobal:                  true,
		DisableConverterErrorTest: false,
	}

	suite.Run(t)
}

func TestRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Exists Formation Template By ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.formation_templates WHERE id = $1`),
				Args:     []driver.Value{testID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		RepoConstructorFunc: formationtemplate.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		TargetID: testID,
		IsGlobal: true,
	}

	suite.Run(t)
}

func TestRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete Formation Template By ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.formation_templates WHERE id = $1`),
				Args:          []driver.Value{testID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		RepoConstructorFunc: formationtemplate.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		IsGlobal:   true,
		MethodArgs: []interface{}{testID},
	}

	suite.Run(t)
}

func TestRepository_List(t *testing.T) {
	suite := testdb.RepoListPageableTestSuite{
		Name:       "Get Formation Template By ID",
		MethodName: "List",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, application_types, runtime_types, missing_artifact_info_message, missing_artifact_warning_message FROM public.formation_templates ORDER BY id LIMIT 3 OFFSET 0`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(formationTemplateEntity.ID, formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.MissingArtifactInfoMessage, formationTemplateEntity.MissingArtifactWarningMessage)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT COUNT(*) FROM public.formation_templates`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"count"}).AddRow(1)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: []interface{}{&formationTemplateModel},
				ExpectedDBEntities:    []interface{}{&formationTemplateEntity},
				ExpectedPage: &model.FormationTemplatePage{
					Data: []*model.FormationTemplate{&formationTemplateModel},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationtemplate.NewRepository,
		MethodArgs:                []interface{}{3, ""},
		DisableConverterErrorTest: false,
	}

	suite.Run(t)
}

func TestRepository_Update(t *testing.T) {
	updateStmt := regexp.QuoteMeta(`UPDATE public.formation_templates SET name = ?, application_types = ?, runtime_types = ?, missing_artifact_info_message = ?, missing_artifact_warning_message = ? WHERE id = ?`)
	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Formation Template By ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateStmt,
				Args:          []driver.Value{formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.MissingArtifactInfoMessage, formationTemplateEntity.MissingArtifactWarningMessage, formationTemplateEntity.ID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		RepoConstructorFunc: formationtemplate.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		ModelEntity:    &formationTemplateModel,
		DBEntity:       &formationTemplateEntity,
		NilModelEntity: nilModelEntity,
		IsGlobal:       true,
	}

	suite.Run(t)
}
