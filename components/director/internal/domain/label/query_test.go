package label

import (
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// TODO: Split the test for global and tenant-scoped version
func Test_FilterQuery_Intersection(t *testing.T) {
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

	stmtPrefix := `SELECT "runtime_id" FROM public.labels ` +
		`WHERE "runtime_id" IS NOT NULL AND "tenant_id" = ?`

	stmtPrefixGlobal := `SELECT "runtime_id" FROM public.labels ` +
		`WHERE "runtime_id" IS NOT NULL`

	testCases := []struct {
		Name                      string
		ReturnSetCombination      SetCombination
		FilterInput               []*labelfilter.LabelFilter
		ExpectedQueryFilter       string
		ExpectedQueryFilterGlobal string
		ExpectedArgs              []interface{}
		ExpectedArgsGlobal        []interface{}
		ExpectedError             error
	}{
		{
			Name:                      "Returns empty query filter when no label filters defined - intersect set",
			ReturnSetCombination:      IntersectSet,
			FilterInput:               nil,
			ExpectedQueryFilter:       "",
			ExpectedQueryFilterGlobal: "",
			ExpectedError:             nil,
		}, {
			Name:                      "Returns empty query filter when no label filters defined - union set",
			ReturnSetCombination:      UnionSet,
			FilterInput:               nil,
			ExpectedQueryFilter:       "",
			ExpectedQueryFilterGlobal: "",
			ExpectedError:             nil,
		}, {
			Name:                      "Query only for label assigned if label filter defined only with key - intersect set",
			ReturnSetCombination:      IntersectSet,
			FilterInput:               []*labelfilter.LabelFilter{&filterAllFoos},
			ExpectedQueryFilter:       stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:              []interface{}{tenantID, filterAllFoos.Key},
			ExpectedQueryFilterGlobal: stmtPrefixGlobal + ` AND "key" = ?`,
			ExpectedArgsGlobal:        []interface{}{filterAllFoos.Key},
			ExpectedError:             nil,
		}, {
			Name:                      "Query only for label assigned if label filter defined only with key - union set",
			ReturnSetCombination:      UnionSet,
			FilterInput:               []*labelfilter.LabelFilter{&filterAllFoos},
			ExpectedQueryFilter:       stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:              []interface{}{tenantID, filterAllFoos.Key},
			ExpectedQueryFilterGlobal: stmtPrefixGlobal + ` AND "key" = ?`,
			ExpectedArgsGlobal:        []interface{}{filterAllFoos.Key},
			ExpectedError:             nil,
		}, {
			Name:                      "Query only for labels assigned if label filter defined only with keys (multiple) - intersect set",
			ReturnSetCombination:      IntersectSet,
			FilterInput:               []*labelfilter.LabelFilter{&filterAllFoos, &filterAllBars},
			ExpectedQueryFilter:       stmtPrefix + ` AND "key" = ? INTERSECT ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:              []interface{}{tenantID, filterAllFoos.Key, tenantID, filterAllBars.Key},
			ExpectedQueryFilterGlobal: stmtPrefixGlobal + ` AND "key" = ? INTERSECT ` + stmtPrefixGlobal + ` AND "key" = ?`,
			ExpectedArgsGlobal:        []interface{}{filterAllFoos.Key, filterAllBars.Key},
			ExpectedError:             nil,
		}, {
			Name:                      "Query only for labels assigned if label filter defined only with keys (multiple) - union set",
			ReturnSetCombination:      UnionSet,
			FilterInput:               []*labelfilter.LabelFilter{&filterAllFoos, &filterAllBars},
			ExpectedQueryFilter:       stmtPrefix + ` AND "key" = ? UNION ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:              []interface{}{tenantID, filterAllFoos.Key, tenantID, filterAllBars.Key},
			ExpectedQueryFilterGlobal: stmtPrefixGlobal + ` AND "key" = ? UNION ` + stmtPrefixGlobal + ` AND "key" = ?`,
			ExpectedArgsGlobal:        []interface{}{filterAllFoos.Key, filterAllBars.Key},
			ExpectedError:             nil,
		}, {
			Name:                      "Query for label assigned with value - intersect set",
			ReturnSetCombination:      IntersectSet,
			FilterInput:               []*labelfilter.LabelFilter{&filterFoosWithValues},
			ExpectedQueryFilter:       stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:              []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query},
			ExpectedQueryFilterGlobal: stmtPrefixGlobal + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgsGlobal:        []interface{}{filterFoosWithValues.Key, *filterFoosWithValues.Query},
			ExpectedError:             nil,
		}, {
			Name:                      "Query for label assigned with value - union set",
			ReturnSetCombination:      UnionSet,
			FilterInput:               []*labelfilter.LabelFilter{&filterFoosWithValues},
			ExpectedQueryFilter:       stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:              []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query},
			ExpectedQueryFilterGlobal: stmtPrefixGlobal + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgsGlobal:        []interface{}{filterFoosWithValues.Key, *filterFoosWithValues.Query},
			ExpectedError:             nil,
		}, {
			Name:                      "Query for labels assigned with values (multiple) - intersect set",
			ReturnSetCombination:      IntersectSet,
			FilterInput:               []*labelfilter.LabelFilter{&filterFoosWithValues, &filterBarsWithValues},
			ExpectedQueryFilter:       stmtPrefix + ` AND "key" = ? AND "value" @> ? INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:              []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query, tenantID, filterBarsWithValues.Key, *filterBarsWithValues.Query},
			ExpectedQueryFilterGlobal: stmtPrefixGlobal + ` AND "key" = ? AND "value" @> ? INTERSECT ` + stmtPrefixGlobal + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgsGlobal:        []interface{}{filterFoosWithValues.Key, *filterFoosWithValues.Query, filterBarsWithValues.Key, *filterBarsWithValues.Query},
			ExpectedError:             nil,
		}, {
			Name:                      "Query for labels assigned with values (multiple) - union set",
			ReturnSetCombination:      UnionSet,
			FilterInput:               []*labelfilter.LabelFilter{&filterFoosWithValues, &filterBarsWithValues},
			ExpectedQueryFilter:       stmtPrefix + ` AND "key" = ? AND "value" @> ? UNION ` + stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:              []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query, tenantID, filterBarsWithValues.Key, *filterBarsWithValues.Query},
			ExpectedQueryFilterGlobal: stmtPrefixGlobal + ` AND "key" = ? AND "value" @> ? UNION ` + stmtPrefixGlobal + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgsGlobal:        []interface{}{filterFoosWithValues.Key, *filterFoosWithValues.Query, filterBarsWithValues.Key, *filterBarsWithValues.Query},
			ExpectedError:             nil,
		}, {
			Name:                      "[Scenarios] Query for label assigned",
			ReturnSetCombination:      IntersectSet,
			FilterInput:               []*labelfilter.LabelFilter{&filterAllScenarios},
			ExpectedQueryFilter:       stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:              []interface{}{tenantID, filterAllScenarios.Key},
			ExpectedQueryFilterGlobal: stmtPrefixGlobal + ` AND "key" = ?`,
			ExpectedArgsGlobal:        []interface{}{filterAllScenarios.Key},
			ExpectedError:             nil,
		}, {
			Name:                      "[Scenarios] Query for label assigned with value",
			ReturnSetCombination:      IntersectSet,
			FilterInput:               []*labelfilter.LabelFilter{&filterScenariosWithFooValues},
			ExpectedQueryFilter:       stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]`,
			ExpectedArgs:              []interface{}{tenantID, filterScenariosWithFooValues.Key, "foo"},
			ExpectedQueryFilterGlobal: stmtPrefixGlobal + ` AND "key" = ? AND "value" ?| array[?]`,
			ExpectedArgsGlobal:        []interface{}{filterScenariosWithFooValues.Key, "foo"},
			ExpectedError:             nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned with values",
			ReturnSetCombination: IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterScenariosWithFooValues, &filterScenariosWithbarPongValues},
			ExpectedQueryFilter: stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]` +
				` INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]`,
			ExpectedArgs: []interface{}{tenantID, filterScenariosWithFooValues.Key, "foo", tenantID, filterScenariosWithbarPongValues.Key, "bar pong"},
			ExpectedQueryFilterGlobal: stmtPrefixGlobal + ` AND "key" = ? AND "value" ?| array[?]` +
				` INTERSECT ` + stmtPrefixGlobal + ` AND "key" = ? AND "value" ?| array[?]`,
			ExpectedArgsGlobal: []interface{}{filterScenariosWithFooValues.Key, "foo", filterScenariosWithbarPongValues.Key, "bar pong"},
			ExpectedError:      nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			queryFilter, args, err := FilterQuery(model.RuntimeLabelableObject, testCase.ReturnSetCombination, tenantID, testCase.FilterInput)

			assert.Equal(t, testCase.ExpectedQueryFilter, queryFilter)
			assert.Equal(t, testCase.ExpectedArgs, args)
			assert.Equal(t, testCase.ExpectedError, err)

			queryFilterGlobal, args, err := FilterQueryGlobal(model.RuntimeLabelableObject, testCase.ReturnSetCombination, testCase.FilterInput)

			assert.Equal(t, testCase.ExpectedQueryFilterGlobal, queryFilterGlobal)
			assert.Equal(t, testCase.ExpectedArgsGlobal, args)
			assert.Equal(t, testCase.ExpectedError, err)
		})
	}
}
