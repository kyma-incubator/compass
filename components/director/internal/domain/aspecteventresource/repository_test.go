package aspecteventresource_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/aspecteventresource"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/aspecteventresource/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	var nilAspectEventResourceModel *model.AspectEventResource
	aspectEventResourceModel := fixAspectEventResourceModel(aspectEventResourceID)
	aspectEventResourceEntity := fixEntityAspectEventResource(aspectEventResourceID, appID, aspectID)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Aspect",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM aspects_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, aspectID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO public.aspect_event_resources \(.+\) VALUES \(.+\)$`,
				Args:        fixAspectEventResourceCreateArgs(aspectEventResourceID, aspectEventResourceModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.AspectEventResourceConverter{}
		},
		RepoConstructorFunc:       aspecteventresource.NewRepository,
		ModelEntity:               aspectEventResourceModel,
		DBEntity:                  aspectEventResourceEntity,
		NilModelEntity:            nilAspectEventResourceModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_ListByApplicationIDs(t *testing.T) {
	// GIVEN
	inputCursor := ""
	firstAppID := "111111111-1111-1111-1111-111111111111"
	secondAppID := "222222222-2222-2222-2222-222222222222"
	firstAspectEventResourceID := "333333333-3333-3333-3333-333333333333"
	secondAspectEventResourceID := "444444444-4444-4444-4444-444444444444"
	firstAspectID := "555555555-5555-5555-5555-555555555555"
	secondAspectID := "666666666-6666-6666-6666-666666666666"

	firstAspectEventResourceEntity := fixEntityAspectEventResource(firstAspectEventResourceID, firstAppID, firstAspectID)
	secondAspectEventResourceEntity := fixEntityAspectEventResource(secondAspectEventResourceID, secondAppID, secondAspectID)
	appIDs := []string{firstAppID, secondAppID}

	selectQueryForAspectEventResources := `^\(SELECT id, app_id, app_template_version_id, aspect_id, ord_id, min_version, subset, ready, created_at, updated_at, deleted_at, error FROM public.aspect_event_resources WHERE \(id IN \(SELECT id FROM aspect_event_resources_tenants WHERE tenant_id = \$1\)\) AND app_id = \$2 ORDER BY aspect_id ASC, app_id ASC LIMIT \$3 OFFSET \$4\) UNION \(SELECT id, app_id, app_template_version_id, aspect_id, ord_id, min_version, subset, ready, created_at, updated_at, deleted_at, error FROM public.aspect_event_resources WHERE \(id IN \(SELECT id FROM aspect_event_resources_tenants WHERE tenant_id = \$5\)\) AND app_id = \$6 ORDER BY aspect_id ASC, app_id ASC LIMIT \$7 OFFSET \$8\)`

	countQueryForAspectsEventResources := `SELECT app_id AS id, COUNT\(\*\) AS total_count FROM public.aspect_event_resources WHERE \(id IN \(SELECT id FROM aspect_event_resources_tenants WHERE tenant_id = \$1\)\) AND app_id IN \(\$2, \$3\) GROUP BY app_id ORDER BY app_id ASC`

	t.Run("success when everything is returned for aspect event resources", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1

		totalCountForFirstAspectEventResource := 1
		totalCountForSecondAspectEventResource := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixAspectEventResourceColumns()).
			AddRow(fixAspectEventResourceRowWithArgs(firstAspectEventResourceID, firstAppID, firstAspectID)...).
			AddRow(fixAspectEventResourceRowWithArgs(secondAspectEventResourceID, secondAppID, secondAspectID)...)

		sqlMock.ExpectQuery(selectQueryForAspectEventResources).
			WithArgs(tenantID, firstAppID, ExpectedLimit, ExpectedOffset, tenantID, secondAppID, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQueryForAspectsEventResources).
			WithArgs(tenantID, firstAppID, secondAppID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(firstAppID, totalCountForFirstAspectEventResource).
				AddRow(secondAppID, totalCountForSecondAspectEventResource))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.AspectEventResourceConverter{}
		convMock.On("FromEntity", firstAspectEventResourceEntity).Return(&model.AspectEventResource{
			ApplicationID:                str.Ptr(firstAppID),
			AspectID:                     firstAspectID,
			ApplicationTemplateVersionID: &appTemplateVersionID,
		}, nil)
		convMock.On("FromEntity", secondAspectEventResourceEntity).Return(&model.AspectEventResource{
			ApplicationID:                str.Ptr(secondAppID),
			AspectID:                     secondAspectID,
			ApplicationTemplateVersionID: &appTemplateVersionID,
		}, nil)
		pgRepository := aspecteventresource.NewRepository(convMock)
		// WHEN
		modelAspectEventResources, totalCounts, err := pgRepository.ListByApplicationIDs(ctx, tenantID, appIDs, inputPageSize, inputCursor)
		// THEN
		require.NoError(t, err)
		require.Len(t, modelAspectEventResources, 2)
		assert.Equal(t, firstAppID, *modelAspectEventResources[0].ApplicationID)
		assert.Equal(t, secondAppID, *modelAspectEventResources[1].ApplicationID)
		assert.Equal(t, firstAspectID, modelAspectEventResources[0].AspectID)
		assert.Equal(t, secondAspectID, modelAspectEventResources[1].AspectID)
		assert.Equal(t, totalCountForFirstAspectEventResource, totalCounts[firstAppID])
		assert.Equal(t, totalCountForSecondAspectEventResource, totalCounts[secondAppID])
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		inputPageSize := 1
		ExpectedLimit := 1
		ExpectedOffset := 0

		pgRepository := aspecteventresource.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQueryForAspectEventResources).
			WithArgs(tenantID, firstAppID, ExpectedLimit, ExpectedOffset, tenantID, secondAppID, ExpectedLimit, ExpectedOffset).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// WHEN
		modelAspectEventResources, totalCounts, err := pgRepository.ListByApplicationIDs(ctx, tenantID, appIDs, inputPageSize, inputCursor)

		// THEN
		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelAspectEventResources)
		assert.Nil(t, totalCounts)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_ListByAspectID(t *testing.T) {
	entity1AspectEventResource := fixEntityAspectEventResource("aspectEventResourceID1", appID, aspectID)
	aspectEventResourceModel1 := fixAspectEventResourceModel("aspectEventResourceID1")
	entity2AspectEventResource := fixEntityAspectEventResource("aspectEventResourceID2", appID, aspectID)
	aspectEventResourceModel2 := fixAspectEventResourceModel("aspectEventResourceID2")

	suite := testdb.RepoListTestSuite{
		Name: "List Aspect Event Resources",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, aspect_id, ord_id, min_version, subset, ready, created_at, updated_at, deleted_at, error FROM public.aspect_event_resources WHERE aspect_id = $1 AND (id IN (SELECT id FROM aspect_event_resources_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{aspectID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAspectEventResourceColumns()).AddRow(fixAspectEventResourceRowWithArgs("aspectEventResourceID1", appID, aspectID)...).AddRow(fixAspectEventResourceRowWithArgs("aspectEventResourceID2", appID, aspectID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAspectEventResourceColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.AspectEventResourceConverter{}
		},
		RepoConstructorFunc:       aspecteventresource.NewRepository,
		ExpectedModelEntities:     []interface{}{aspectEventResourceModel1, aspectEventResourceModel2},
		ExpectedDBEntities:        []interface{}{entity1AspectEventResource, entity2AspectEventResource},
		MethodArgs:                []interface{}{tenantID, aspectID},
		MethodName:                "ListByAspectID",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}
