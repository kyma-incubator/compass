package label_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func Test_FilterQuery(t *testing.T) {
	tenantID := uuid.New()

	fooQuery := `["foo-value"]`
	barQuery := `["bar-value"]`
	scenariosFooQuery := `$[*] ? (@ == "foo")`
	scenariosBarPongQuery := `$[*] ? (@ == "bar pong")`
	filterQueryWithJSON := `{"exists": false}`

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
	filterGlobalSubaacountIDValues := labelfilter.LabelFilter{
		Key:   "global_subaccount_id",
		Query: &filterQueryWithJSON,
	}

	stmtPrefix := `SELECT "runtime_id" FROM public.labels WHERE "runtime_id" IS NOT NULL AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = ?))`

	testCases := []struct {
		Name                 string
		ReturnSetCombination label.SetCombination
		FilterInput          []*labelfilter.LabelFilter
		ExpectedQueryFilter  string
		ExpectedArgs         []interface{}
		ExpectedError        error
	}{
		{
			Name:                 "Returns empty query filter when no label filters defined - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          nil,
			ExpectedQueryFilter:  "",
			ExpectedError:        nil,
		}, {
			Name:                 "Returns empty query filter when no label filters defined - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          nil,
			ExpectedQueryFilter:  "",
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for label assigned if label filter defined only with key - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllFoos.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for label assigned if label filter defined only with key - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllFoos.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for labels assigned if label filter defined only with keys (multiple) - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos, &filterAllBars},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? INTERSECT ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllFoos.Key, tenantID, filterAllBars.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for labels assigned if label filter defined only with keys (multiple) - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos, &filterAllBars},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? UNION ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllFoos.Key, tenantID, filterAllBars.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for label assigned with value - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for label assigned with value - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for labels assigned with values (multiple) - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues, &filterBarsWithValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" @> ? INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query, tenantID, filterBarsWithValues.Key, *filterBarsWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for labels assigned with values (multiple) - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues, &filterBarsWithValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" @> ? UNION ` + stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query, tenantID, filterBarsWithValues.Key, *filterBarsWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllScenarios},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllScenarios.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned with value",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterScenariosWithFooValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]`,
			ExpectedArgs:         []interface{}{tenantID, filterScenariosWithFooValues.Key, "foo"},
			ExpectedError:        nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned with values",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterScenariosWithFooValues, &filterScenariosWithbarPongValues},
			ExpectedQueryFilter: stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]` +
				` INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]`,
			ExpectedArgs:  []interface{}{tenantID, filterScenariosWithFooValues.Key, "foo", tenantID, filterScenariosWithbarPongValues.Key, "bar pong"},
			ExpectedError: nil,
		},
		{
			Name:                 "[Global_Subaccount_Id] Query for label assigned with values",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterGlobalSubaacountIDValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "app_id" NOT IN (SELECT "app_id" FROM labels WHERE key = 'global_subaccount_id')`,
			ExpectedArgs:         []interface{}{tenantID},
			ExpectedError:        nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			queryFilter, args, err := label.FilterQuery(model.RuntimeLabelableObject, testCase.ReturnSetCombination, tenantID, testCase.FilterInput)

			assert.Equal(t, testCase.ExpectedQueryFilter, removeWhitespace(queryFilter))
			assert.Equal(t, testCase.ExpectedArgs, args)
			assert.Equal(t, testCase.ExpectedError, err)
		})
	}
}

func TestFilterQueryGlobal(t *testing.T) {
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
		`WHERE "runtime_id" IS NOT NULL`

	testCases := []struct {
		Name                 string
		ReturnSetCombination label.SetCombination
		FilterInput          []*labelfilter.LabelFilter
		ExpectedQueryFilter  string
		ExpectedArgs         []interface{}
		ExpectedError        error
	}{
		{
			Name:                 "Returns empty query filter when no label filters defined - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          nil,
			ExpectedError:        nil,
		}, {
			Name:                 "Returns empty query filter when no label filters defined - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          nil,
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for label assigned if label filter defined only with key - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{filterAllFoos.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for label assigned if label filter defined only with key - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{filterAllFoos.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for labels assigned if label filter defined only with keys (multiple) - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos, &filterAllBars},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? INTERSECT ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{filterAllFoos.Key, filterAllBars.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for labels assigned if label filter defined only with keys (multiple) - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos, &filterAllBars},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? UNION ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{filterAllFoos.Key, filterAllBars.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for label assigned with value - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{filterFoosWithValues.Key, *filterFoosWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for label assigned with value - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{filterFoosWithValues.Key, *filterFoosWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for labels assigned with values (multiple) - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues, &filterBarsWithValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" @> ? INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{filterFoosWithValues.Key, *filterFoosWithValues.Query, filterBarsWithValues.Key, *filterBarsWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for labels assigned with values (multiple) - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues, &filterBarsWithValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" @> ? UNION ` + stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{filterFoosWithValues.Key, *filterFoosWithValues.Query, filterBarsWithValues.Key, *filterBarsWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllScenarios},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{filterAllScenarios.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned with value",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterScenariosWithFooValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]`,
			ExpectedArgs:         []interface{}{filterScenariosWithFooValues.Key, "foo"},
			ExpectedError:        nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned with values",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterScenariosWithFooValues, &filterScenariosWithbarPongValues},
			ExpectedQueryFilter: stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]` +
				` INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]`,
			ExpectedArgs:  []interface{}{filterScenariosWithFooValues.Key, "foo", filterScenariosWithbarPongValues.Key, "bar pong"},
			ExpectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			queryFilter, args, err := label.FilterQueryGlobal(model.RuntimeLabelableObject, testCase.ReturnSetCombination, testCase.FilterInput)

			assert.Equal(t, testCase.ExpectedQueryFilter, queryFilter)
			assert.Equal(t, testCase.ExpectedArgs, args)
			assert.Equal(t, testCase.ExpectedError, err)
		})
	}
}

func TestFilterQueryWithEmbeddedTenant(t *testing.T) {
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

	stmtPrefix := `SELECT "tenant_id" FROM public.labels ` +
		`WHERE "tenant_id" IS NOT NULL AND tenant_id = ?`

	testCases := []struct {
		Name                 string
		ReturnSetCombination label.SetCombination
		FilterInput          []*labelfilter.LabelFilter
		ExpectedQueryFilter  string
		ExpectedArgs         []interface{}
		ExpectedError        error
	}{
		{
			Name:                 "Returns empty query filter when no label filters defined - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          nil,
			ExpectedError:        nil,
		}, {
			Name:                 "Returns empty query filter when no label filters defined - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          nil,
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for label assigned if label filter defined only with key - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllFoos.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for label assigned if label filter defined only with key - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllFoos.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for labels assigned if label filter defined only with keys (multiple) - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos, &filterAllBars},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? INTERSECT ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllFoos.Key, tenantID, filterAllBars.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for labels assigned if label filter defined only with keys (multiple) - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos, &filterAllBars},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? UNION ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllFoos.Key, tenantID, filterAllBars.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for label assigned with value - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for label assigned with value - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for labels assigned with values (multiple) - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues, &filterBarsWithValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" @> ? INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query, tenantID, filterBarsWithValues.Key, *filterBarsWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for labels assigned with values (multiple) - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues, &filterBarsWithValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" @> ? UNION ` + stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query, tenantID, filterBarsWithValues.Key, *filterBarsWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllScenarios},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllScenarios.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned with value",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterScenariosWithFooValues},
			ExpectedQueryFilter:  stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]`,
			ExpectedArgs:         []interface{}{tenantID, filterScenariosWithFooValues.Key, "foo"},
			ExpectedError:        nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned with values",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterScenariosWithFooValues, &filterScenariosWithbarPongValues},
			ExpectedQueryFilter: stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]` +
				` INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]`,
			ExpectedArgs:  []interface{}{tenantID, filterScenariosWithFooValues.Key, "foo", tenantID, filterScenariosWithbarPongValues.Key, "bar pong"},
			ExpectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			queryFilter, args, err := label.FilterQuery(model.TenantLabelableObject, testCase.ReturnSetCombination, tenantID, testCase.FilterInput)

			assert.Equal(t, testCase.ExpectedQueryFilter, queryFilter)
			assert.Equal(t, testCase.ExpectedArgs, args)
			assert.Equal(t, testCase.ExpectedError, err)
		})
	}
}

func TestFilterSubquery(t *testing.T) {
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

	stmtPrefix := `SELECT "runtime_id" FROM public.labels WHERE "runtime_id" IS NOT NULL AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = ?))`

	testCases := []struct {
		Name                 string
		ReturnSetCombination label.SetCombination
		FilterInput          []*labelfilter.LabelFilter
		ExpectedQueryFilter  string
		ExpectedArgs         []interface{}
		ExpectedError        error
	}{
		{
			Name:                 "Returns empty query filter when no label filters defined - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          nil,
			ExpectedQueryFilter:  "",
			ExpectedError:        nil,
		}, {
			Name:                 "Returns empty query filter when no label filters defined - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          nil,
			ExpectedQueryFilter:  "",
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for label assigned if label filter defined only with key - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos},
			ExpectedQueryFilter:  `INTERSECT ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllFoos.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for label assigned if label filter defined only with key - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos},
			ExpectedQueryFilter:  `UNION ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllFoos.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for labels assigned if label filter defined only with keys (multiple) - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos, &filterAllBars},
			ExpectedQueryFilter:  `INTERSECT ` + stmtPrefix + ` AND "key" = ? INTERSECT ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllFoos.Key, tenantID, filterAllBars.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query only for labels assigned if label filter defined only with keys (multiple) - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllFoos, &filterAllBars},
			ExpectedQueryFilter:  `UNION ` + stmtPrefix + ` AND "key" = ? UNION ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllFoos.Key, tenantID, filterAllBars.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for label assigned with value - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues},
			ExpectedQueryFilter:  `INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for label assigned with value - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues},
			ExpectedQueryFilter:  `UNION ` + stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for labels assigned with values (multiple) - intersect set",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues, &filterBarsWithValues},
			ExpectedQueryFilter:  `INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" @> ? INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query, tenantID, filterBarsWithValues.Key, *filterBarsWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "Query for labels assigned with values (multiple) - union set",
			ReturnSetCombination: label.UnionSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterFoosWithValues, &filterBarsWithValues},
			ExpectedQueryFilter:  `UNION ` + stmtPrefix + ` AND "key" = ? AND "value" @> ? UNION ` + stmtPrefix + ` AND "key" = ? AND "value" @> ?`,
			ExpectedArgs:         []interface{}{tenantID, filterFoosWithValues.Key, *filterFoosWithValues.Query, tenantID, filterBarsWithValues.Key, *filterBarsWithValues.Query},
			ExpectedError:        nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterAllScenarios},
			ExpectedQueryFilter:  `INTERSECT ` + stmtPrefix + ` AND "key" = ?`,
			ExpectedArgs:         []interface{}{tenantID, filterAllScenarios.Key},
			ExpectedError:        nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned with value",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterScenariosWithFooValues},
			ExpectedQueryFilter:  `INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]`,
			ExpectedArgs:         []interface{}{tenantID, filterScenariosWithFooValues.Key, "foo"},
			ExpectedError:        nil,
		}, {
			Name:                 "[Scenarios] Query for label assigned with values",
			ReturnSetCombination: label.IntersectSet,
			FilterInput:          []*labelfilter.LabelFilter{&filterScenariosWithFooValues, &filterScenariosWithbarPongValues},
			ExpectedQueryFilter: `INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]` +
				` INTERSECT ` + stmtPrefix + ` AND "key" = ? AND "value" ?| array[?]`,
			ExpectedArgs:  []interface{}{tenantID, filterScenariosWithFooValues.Key, "foo", tenantID, filterScenariosWithbarPongValues.Key, "bar pong"},
			ExpectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			queryFilter, args, err := label.FilterSubquery(model.RuntimeLabelableObject, testCase.ReturnSetCombination, tenantID, testCase.FilterInput)

			assert.Equal(t, testCase.ExpectedQueryFilter, removeWhitespace(queryFilter))
			assert.Equal(t, testCase.ExpectedArgs, args)
			assert.Equal(t, testCase.ExpectedError, err)
		})
	}
}

func removeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
