package aspect_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/aspect"
	"github.com/kyma-incubator/compass/components/director/internal/domain/aspect/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	var nilAspectModel *model.Aspect
	aspectModel := fixAspectModel(aspectID)
	aspectEntity := fixEntityAspect(aspectID, appID, integrationDependencyID)

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

func TestPgRepository_ListByApplicationIDs(t *testing.T) {
	// GIVEN
	inputCursor := ""
	firstAppID := "111111111-1111-1111-1111-111111111111"
	secondAppID := "222222222-2222-2222-2222-222222222222"
	firstAspectID := "333333333-3333-3333-3333-333333333333"
	secondAspectID := "444444444-4444-4444-4444-444444444444"
	firstIntDepID := "555555555-5555-5555-5555-555555555555"
	secondIntDepID := "666666666-6666-6666-6666-666666666666"

	firstAspectEntity := fixEntityAspect(firstAspectID, firstAppID, firstIntDepID)
	secondAspectEntity := fixEntityAspect(secondAspectID, secondAppID, secondIntDepID)
	appIDs := []string{firstAppID, secondAppID}

	selectQueryForAspects := `^\(SELECT id, app_id, app_template_version_id, integration_dependency_id, title, description, mandatory, support_multiple_providers, api_resources, event_resources, ready, created_at, updated_at, deleted_at, error FROM public.aspects WHERE \(id IN \(SELECT id FROM aspects_tenants WHERE tenant_id = \$1\)\) AND app_id = \$2 ORDER BY integration_dependency_id ASC, app_id ASC LIMIT \$3 OFFSET \$4\) UNION \(SELECT id, app_id, app_template_version_id, integration_dependency_id, title, description, mandatory, support_multiple_providers, api_resources, event_resources, ready, created_at, updated_at, deleted_at, error FROM public.aspects WHERE \(id IN \(SELECT id FROM aspects_tenants WHERE tenant_id = \$5\)\) AND app_id = \$6 ORDER BY integration_dependency_id ASC, app_id ASC LIMIT \$7 OFFSET \$8\)`

	countQueryForAspects := `SELECT app_id AS id, COUNT\(\*\) AS total_count FROM public.aspects WHERE \(id IN \(SELECT id FROM aspects_tenants WHERE tenant_id = \$1\)\) GROUP BY app_id ORDER BY app_id ASC`

	t.Run("success when everything is returned for aspects", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1

		totalCountForFirstAspect := 1
		totalCountForSecondAspect := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixAspectColumns()).
			AddRow(fixAspectRowWithArgs(firstAspectID, firstAppID, firstIntDepID)...).
			AddRow(fixAspectRowWithArgs(secondAspectID, secondAppID, secondIntDepID)...)

		sqlMock.ExpectQuery(selectQueryForAspects).
			WithArgs(tenantID, firstAppID, ExpectedLimit, ExpectedOffset, tenantID, secondAppID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryForAspects).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstAppID, totalCountForFirstAspect).
				AddRow(secondAppID, totalCountForSecondAspect))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.AspectConverter{}
		convMock.On("FromEntity", firstAspectEntity).Return(&model.Aspect{
			ApplicationID:                str.Ptr(firstAppID),
			IntegrationDependencyID:      firstIntDepID,
			ApplicationTemplateVersionID: &appTemplateVersionID,
		}, nil)
		convMock.On("FromEntity", secondAspectEntity).Return(&model.Aspect{
			ApplicationID:                str.Ptr(secondAppID),
			IntegrationDependencyID:      secondIntDepID,
			ApplicationTemplateVersionID: &appTemplateVersionID,
		}, nil)
		pgRepository := aspect.NewRepository(convMock)
		// WHEN
		modelAspects, totalCounts, err := pgRepository.ListByApplicationIDs(ctx, tenantID, appIDs, inputPageSize, inputCursor)
		// THEN
		require.NoError(t, err)
		require.Len(t, modelAspects, 2)
		assert.Equal(t, firstAppID, *modelAspects[0].ApplicationID)
		assert.Equal(t, secondAppID, *modelAspects[1].ApplicationID)
		assert.Equal(t, firstIntDepID, modelAspects[0].IntegrationDependencyID)
		assert.Equal(t, secondIntDepID, modelAspects[1].IntegrationDependencyID)
		assert.Equal(t, totalCountForFirstAspect, totalCounts[firstAppID])
		assert.Equal(t, totalCountForSecondAspect, totalCounts[secondAppID])
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		inputPageSize := 1
		ExpectedLimit := 1
		ExpectedOffset := 0

		pgRepository := aspect.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQueryForAspects).
			WithArgs(tenantID, firstAppID, ExpectedLimit, ExpectedOffset, tenantID, secondAppID, ExpectedLimit, ExpectedOffset).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// WHEN
		modelAspects, totalCounts, err := pgRepository.ListByApplicationIDs(ctx, tenantID, appIDs, inputPageSize, inputCursor)

		// THEN
		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelAspects)
		assert.Nil(t, totalCounts)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_ListByIntegrationDependencyID(t *testing.T) {
	entity1Aspect := fixEntityAspect("aspectID1", appID, integrationDependencyID)
	aspectModel1 := fixAspectModel("aspectID1")
	entity2Aspect := fixEntityAspect("aspectID2", appID, integrationDependencyID)
	aspectModel2 := fixAspectModel("aspectID2")

	suite := testdb.RepoListTestSuite{
		Name: "List Aspects",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, integration_dependency_id, title, description, mandatory, support_multiple_providers, api_resources, event_resources, ready, created_at, updated_at, deleted_at, error FROM public.aspects WHERE integration_dependency_id = $1 AND (id IN (SELECT id FROM aspects_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{integrationDependencyID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAspectColumns()).AddRow(fixAspectRowWithArgs("aspectID1", appID, integrationDependencyID)...).AddRow(fixAspectRowWithArgs("aspectID2", appID, integrationDependencyID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAspectColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.AspectConverter{}
		},
		RepoConstructorFunc:       aspect.NewRepository,
		ExpectedModelEntities:     []interface{}{aspectModel1, aspectModel2},
		ExpectedDBEntities:        []interface{}{entity1Aspect, entity2Aspect},
		MethodArgs:                []interface{}{tenantID, integrationDependencyID},
		MethodName:                "ListByIntegrationDependencyID",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}
