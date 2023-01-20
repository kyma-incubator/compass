package formationtemplateconstraintreferences_test

import (
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"regexp"
	"testing"
)

func TestRepository_ListMatchingFormationConstraints(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name:       "List Formation Constraints",
		MethodName: "ListByFormationTemplateID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT formation_constraint_id, formation_template_id FROM public.formation_template_constraint_references WHERE formation_template = $1`),
				IsSelect: true,
				Args:     []driver.Value{templateID},
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(entity.FormationTemplateID, entity.ConstraintID)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			conv := &automock.EntityConverter{}
			return conv
		},
		RepoConstructorFunc:       formationtemplateconstraintreferences.NewRepository,
		MethodArgs:                []interface{}{templateID},
		ExpectedDBEntities:        []interface{}{entity},
		ExpectedModelEntities:     []interface{}{constraintReference},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_Create(t *testing.T) {
	suite := testdb.RepoCreateTestSuite{
		Name:       "Create Formation Constraint",
		MethodName: "Create",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.formation_template_constraint_references \(.+\) VALUES \(.+\)$`,
				Args:        []driver.Value{entity.ConstraintID, entity.FormationTemplateID},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationtemplateconstraintreferences.NewRepository,
		ModelEntity:               constraintReference,
		DBEntity:                  entity,
		NilModelEntity:            nilModel,
		IsGlobal:                  true,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete Formation Constraint by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.formation_template_constraint_references WHERE formation_template = $1 AND formation_constraint = $2`),
				Args:          []driver.Value{templateID, constraintID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		RepoConstructorFunc: formationtemplateconstraintreferences.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		IsGlobal:   true,
		MethodArgs: []interface{}{templateID, constraintID},
	}

	suite.Run(t)
}
