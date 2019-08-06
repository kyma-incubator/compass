package application

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//TODO: This tests only calling SQL functions and
// Cannot test paging, because we have mixed resository
func TestPgRepository_ListByRuntimeScenarios(t *testing.T) {
	tenantID := uuid.New()
	pageSize := 5
	cursor := ""

	runtimeScenarios := []string{"Java", "Go", "Elixir"}
	//Create Filters for scenarios, because we cannot mock  filter query generator
	var scenarioFilter []*labelfilter.LabelFilter
	for _, scenario := range runtimeScenarios {
		query := fmt.Sprintf(`$[*] ? (@ == "%s")`, scenario)
		scenarioFilter = append(scenarioFilter, &labelfilter.LabelFilter{Key: model.ScenariosKey, Query: &query})
	}

	scenarioQuery, err := label.FilterQuery(model.ApplicationLabelableObject, label.UnionSet, tenantID, scenarioFilter)
	require.NoError(t, err)
	applicationScenarioQuery := regexp.QuoteMeta(scenarioQuery)

	testCases := []struct {
		Name                        string
		ExpectedRuntimeScenarioRows *sqlmock.Rows
		ExpectedApplicationRows     *sqlmock.Rows
		ExpectedError               error
	}{
		{
			Name:                        "Success",
			ExpectedApplicationRows:     sqlmock.NewRows([]string{"app_id"}).AddRow(uuid.New()).AddRow(uuid.New()),
			ExpectedError:               nil,
		},
		{
			Name:                        "Return empty page when no application match",
			ExpectedApplicationRows:     sqlmock.NewRows([]string{"App_id"}),
			ExpectedError:               nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sqlxDB, sqlMock := mockDatabase(t)
			if testCase.ExpectedApplicationRows != nil {
				sqlMock.ExpectQuery(applicationScenarioQuery).
					WithArgs().
					WillReturnRows(testCase.ExpectedApplicationRows)
			}
			repository := NewRepository()

			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			//WHEN
			page, err := repository.ListByScenarios(ctx, tenantID, runtimeScenarios, &pageSize, &cursor)

			//THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, page)
			}
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func mockDatabase(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	sqlDB, sqlMock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(sqlDB, "sqlmock")

	return sqlxDB, sqlMock
}
