package label

import (
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
)

func Test_FilterQuery(t *testing.T) {
	tenantID := uuid.New()

	fooQuery := `["foo-value"]`
	barQuery := `["bar-value"]`
	scenariosFooQuery := `$[*] ? (@ == "foo")`
	scenariosBarPongQuery := `$[*] ? (@ == "bar pong")`

	filterAllFoos := labelfilter.LabelFilter{
		Key:   "Foo",
		Query: nil,
	}
	filterAllBars := labelfilter.LabelFilter{
		Key:   "Bar",
		Query: nil,
	}
	filterFoosWithValues := labelfilter.LabelFilter{
		Key:   "Foo",
		Query: &fooQuery,
	}
	filterBarsWithValues := labelfilter.LabelFilter{
		Key:   "Bar",
		Query: &barQuery,
	}
	filterAllScenarios := labelfilter.LabelFilter{
		Key:   "Scenarios",
		Query: nil,
	}
	filterScenariosWithFooValues := labelfilter.LabelFilter{
		Key:   "Scenarios",
		Query: &scenariosFooQuery,
	}
	filterScenariosWithbarPongValues := labelfilter.LabelFilter{
		Key:   "Scenarios",
		Query: &scenariosBarPongQuery,
	}

	stmtPrefix := `SELECT "runtime_id" FROM "public"."labels" WHERE "tenant_id" = '` + tenantID.String() + `'`

	testCases := []struct {
		Name                string
		FilterInput         []*labelfilter.LabelFilter
		ExpectedQueryFilter string
	}{
		{
			Name:                "Returns empty query filter when no label filters defined",
			FilterInput:         nil,
			ExpectedQueryFilter: "",
		}, {
			Name:                "Query only for label assigned if label filter defined only with key",
			FilterInput:         []*labelfilter.LabelFilter{&filterAllFoos},
			ExpectedQueryFilter: stmtPrefix + ` AND "key" = '` + filterAllFoos.Key + `'`,
		}, {
			Name:        "Query only for labels assigned if label filter defined only with keys (multiple)",
			FilterInput: []*labelfilter.LabelFilter{&filterAllFoos, &filterAllBars},
			ExpectedQueryFilter: stmtPrefix + ` AND "key" = '` + filterAllFoos.Key + `'` +
				` INTERSECT ` + stmtPrefix + ` AND "key" = '` + filterAllBars.Key + `'`,
		}, {
			Name:                "Query for label assigned with value",
			FilterInput:         []*labelfilter.LabelFilter{&filterFoosWithValues},
			ExpectedQueryFilter: stmtPrefix + ` AND "key" = '` + filterFoosWithValues.Key + `' AND "value" @> '` + *filterFoosWithValues.Query + `'`,
		}, {
			Name:        "Query for labels assigned with values (multiple)",
			FilterInput: []*labelfilter.LabelFilter{&filterFoosWithValues, &filterBarsWithValues},
			ExpectedQueryFilter: stmtPrefix + ` AND "key" = '` + filterFoosWithValues.Key + `' AND "value" @> '` + *filterFoosWithValues.Query + `'` +
				` INTERSECT ` + stmtPrefix + ` AND "key" = '` + filterBarsWithValues.Key + `' AND "value" @> '` + *filterBarsWithValues.Query + `'`,
		}, {
			Name:                "[Scenarios] Query for label assigned",
			FilterInput:         []*labelfilter.LabelFilter{&filterAllScenarios},
			ExpectedQueryFilter: stmtPrefix + ` AND "key" = '` + filterAllScenarios.Key + `'`,
		}, {
			Name:                "[Scenarios] Query for label assigned with value",
			FilterInput:         []*labelfilter.LabelFilter{&filterScenariosWithFooValues},
			ExpectedQueryFilter: stmtPrefix + ` AND "key" = '` + filterScenariosWithFooValues.Key + `' AND "value" @> '["foo"]'`,
		}, {
			Name:        "[Scenarios] Query for label assigned with values",
			FilterInput: []*labelfilter.LabelFilter{&filterScenariosWithFooValues, &filterScenariosWithbarPongValues},
			ExpectedQueryFilter: stmtPrefix + ` AND "key" = '` + filterScenariosWithFooValues.Key + `' AND "value" @> '["foo"]'` +
				` INTERSECT ` + stmtPrefix + ` AND "key" = '` + filterScenariosWithbarPongValues.Key + `' AND "value" @> '["bar pong"]'`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			queryFilter := FilterQuery(QueryForRuntime, tenantID, testCase.FilterInput)

			assert.Equal(t, testCase.ExpectedQueryFilter, queryFilter)
		})
	}
}
