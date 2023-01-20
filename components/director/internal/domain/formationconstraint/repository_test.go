package formationconstraint_test

import (
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"regexp"
	"testing"
)

func TestRepository_Get(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Formation Constraint By ID",
		MethodName: "Get",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, constraint_type, target_operation, operator, resource_type, resource_subtype, operator_scope, input_template, constraint_scope FROM public.formation_constraints WHERE id = $1`),
				Args:     []driver.Value{testID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(entity.ID, entity.Name, entity.ConstraintType, entity.TargetOperation, entity.Operator, entity.ResourceType, entity.ResourceSubtype, entity.InputTemplate, entity.ConstraintScope)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationconstraint.NewRepository,
		ExpectedModelEntity:       formationConstraintModel,
		ExpectedDBEntity:          &entity,
		MethodArgs:                []interface{}{testID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_ListAll(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name:       "List Formation Constraints",
		MethodName: "ListAll",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, constraint_type, target_operation, operator, resource_type, resource_subtype, operator_scope, input_template, constraint_scope FROM public.formation_constraints`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(entity.ID, entity.Name, entity.ConstraintType, entity.TargetOperation, entity.Operator, entity.ResourceType, entity.ResourceSubtype, entity.InputTemplate, entity.ConstraintScope)}
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
		RepoConstructorFunc:       formationconstraint.NewRepository,
		MethodArgs:                []interface{}{},
		ExpectedDBEntities:        []interface{}{&entity},
		ExpectedModelEntities:     []interface{}{formationConstraintModel},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_ListByIDs(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name:       "List Formation Constraints",
		MethodName: "ListByIDs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, constraint_type, target_operation, operator, resource_type, resource_subtype, operator_scope, input_template, constraint_scope FROM public.formation_constraints WHERE id IN ($1)`),
				IsSelect: true,
				Args:     []driver.Value{testID},
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(entity.ID, entity.Name, entity.ConstraintType, entity.TargetOperation, entity.Operator, entity.ResourceType, entity.ResourceSubtype, entity.InputTemplate, entity.ConstraintScope)}
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
		RepoConstructorFunc:       formationconstraint.NewRepository,
		MethodArgs:                []interface{}{[]string{testID}},
		ExpectedDBEntities:        []interface{}{&entity},
		ExpectedModelEntities:     []interface{}{formationConstraintModel},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_ListMatchingFormationConstraints(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name:       "List Formation Constraints",
		MethodName: "ListMatchingFormationConstraints",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, constraint_type, target_operation, operator, resource_type, resource_subtype, operator_scope, input_template, constraint_scope FROM public.formation_constraints WHERE (target_operation = $1 AND constraint_type = $2 AND resource_type = $3 AND resource_subtype = $4 AND (constraint_scope = $5 OR id IN ($6)))`),
				IsSelect: true,
				Args:     []driver.Value{location.OperationName, location.ConstraintType, details.ResourceType, details.ResourceSubtype, "GLOBAL", testID},
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(entity.ID, entity.Name, entity.ConstraintType, entity.TargetOperation, entity.Operator, entity.ResourceType, entity.ResourceSubtype, entity.InputTemplate, entity.ConstraintScope)}
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
		RepoConstructorFunc:       formationconstraint.NewRepository,
		MethodArgs:                []interface{}{[]string{testID}, location, matchingDetails},
		ExpectedDBEntities:        []interface{}{&entity},
		ExpectedModelEntities:     []interface{}{formationConstraintModel},
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
				Query:       `^INSERT INTO public.formation_constraints \(.+\) VALUES \(.+\)$`,
				Args:        []driver.Value{entity.ID, entity.Name, entity.ConstraintType, entity.TargetOperation, entity.Operator, entity.ResourceType, entity.ResourceSubtype, entity.InputTemplate, entity.ConstraintScope},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationconstraint.NewRepository,
		ModelEntity:               formationConstraintModel,
		DBEntity:                  &entity,
		NilModelEntity:            nilModelEntity,
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
				Query:         regexp.QuoteMeta(`DELETE FROM public.formation_constraints WHERE id = $1`),
				Args:          []driver.Value{testID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		RepoConstructorFunc: formationconstraint.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		IsGlobal:   true,
		MethodArgs: []interface{}{testID},
	}

	suite.Run(t)
}
