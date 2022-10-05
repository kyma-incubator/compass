package label

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// SetCombination type defines possible result set combination for querying
type SetCombination string

const (
	// IntersectSet missing godoc
	IntersectSet SetCombination = "INTERSECT"
	// ExceptSet missing godoc
	ExceptSet SetCombination = "EXCEPT"
	// UnionSet missing godoc
	UnionSet                   SetCombination = "UNION"
	scenariosLabelKey          string         = "SCENARIOS"
	globalSubaccountIDLabelKey string         = "global_subaccount_id"
	stmtPrefixFormat           string         = `SELECT "%s" FROM %s WHERE "%s" IS NOT NULL AND`
	stmtPrefixGlobalFormat     string         = `SELECT "%s" FROM %s WHERE "%s" IS NOT NULL`
)

type queryFilter struct {
	exists bool
}

// FilterQuery builds select query for given filters
//
// It supports querying defined by `queryFor` parameter. All queries are created
// in the context of given tenant
func FilterQuery(queryFor model.LabelableObject, setCombination SetCombination, tenant uuid.UUID, filter []*labelfilter.LabelFilter) (string, []interface{}, error) {
	return filterQuery(queryFor, setCombination, tenant, filter, false)
}

// FilterSubquery builds select sub query for given filters that can be appended to other query
//
// It supports querying defined by `queryFor` parameter. All queries are created
// in the context of given tenant
func FilterSubquery(queryFor model.LabelableObject, setCombination SetCombination, tenant uuid.UUID, filter []*labelfilter.LabelFilter) (string, []interface{}, error) {
	return filterQuery(queryFor, setCombination, tenant, filter, true)
}

// FilterQueryGlobal builds select query for given filters
//
// It supports querying defined by `queryFor` parameter. All queries are created
// in the global context
func FilterQueryGlobal(queryFor model.LabelableObject, setCombination SetCombination, filters []*labelfilter.LabelFilter) (string, []interface{}, error) {
	if filters == nil {
		return "", nil, nil
	}

	objectField := labelableObjectField(queryFor)

	stmtPrefix := fmt.Sprintf(stmtPrefixGlobalFormat, objectField, tableName, objectField)

	return buildFilterQuery(stmtPrefix, nil, setCombination, filters, false)
}

func filterQuery(queryFor model.LabelableObject, setCombination SetCombination, tenant uuid.UUID, filter []*labelfilter.LabelFilter, isSubQuery bool) (string, []interface{}, error) {
	if filter == nil {
		return "", nil, nil
	}

	objectField := labelableObjectField(queryFor)

	var cond repo.Condition
	if queryFor == model.TenantLabelableObject {
		cond = repo.NewEqualCondition("tenant_id", tenant)
	} else {
		var err error
		cond, err = repo.NewTenantIsolationCondition(queryFor.GetResourceType(), tenant.String(), false)
		if err != nil {
			return "", nil, err
		}
	}

	stmtPrefixFormatWithTenantIsolation := stmtPrefixFormat + " " + cond.GetQueryPart()
	stmtPrefix := fmt.Sprintf(stmtPrefixFormatWithTenantIsolation, objectField, tableName, objectField)
	var stmtPrefixArgs []interface{}
	stmtPrefixArgs = append(stmtPrefixArgs, tenant)

	return buildFilterQuery(stmtPrefix, stmtPrefixArgs, setCombination, filter, isSubQuery)
}

func buildFilterQuery(stmtPrefix string, stmtPrefixArgs []interface{}, setCombination SetCombination, filters []*labelfilter.LabelFilter, isSubQuery bool) (string, []interface{}, error) {
	var queryBuilder strings.Builder

	args := make([]interface{}, 0, len(filters))
	for idx, lblFilter := range filters {
		if idx > 0 || isSubQuery {
			queryBuilder.WriteString(fmt.Sprintf(` %s `, setCombination))
		}

		queryBuilder.WriteString(stmtPrefix)
		if len(stmtPrefixArgs) > 0 {
			args = append(args, stmtPrefixArgs...)
		}

		// TODO: for optimization it can be detected if the given Key was already added to the query
		// if so, it can be omitted

		shouldKeyExists := true
		var err error
		if lblFilter.Key == globalSubaccountIDLabelKey {
			shouldKeyExists, err = shouldGlobalSubaccountExists(lblFilter.Query)
			if err != nil {
				return "", nil, errors.Wrap(err, "while determining if global_subaccount_id exists")
			}
		}

		if shouldKeyExists {
			queryBuilder.WriteString(` AND "key" = ?`)
			args = append(args, lblFilter.Key)
		}

		if lblFilter.Query != nil {
			queryValue := *lblFilter.Query
			// Handling the Scenarios label case - we assume that Query is
			// in SQL/JSON path format supported by PostgreSQL 12. Till it
			// is not production ready, we need to transform the Query from
			// SQL/JSON path to old JSON queries.
			if strings.ToUpper(lblFilter.Key) == scenariosLabelKey {
				extractedValues, err := ExtractValueFromJSONPath(queryValue)
				if err != nil {
					return "", nil, errors.Wrap(err, "while extracting value from JSON path")
				}

				args = append(args, extractedValues...)

				queryValues := make([]string, len(extractedValues))
				for idx := range extractedValues {
					queryValues[idx] = "?"
				}
				queryValue = `array[` + strings.Join(queryValues, ",") + `]`

				queryBuilder.WriteString(fmt.Sprintf(` AND "value" ?| %s`, queryValue))
			} else if lblFilter.Key == globalSubaccountIDLabelKey && !shouldKeyExists {
				queryBuilder.WriteString(` AND "app_id" NOT IN (SELECT "app_id" FROM labels WHERE key = 'global_subaccount_id')`)
			} else {
				args = append(args, queryValue)
				queryBuilder.WriteString(` AND "value" @> ?`)
			}
		}
	}

	return queryBuilder.String(), args, nil
}

func shouldGlobalSubaccountExists(filter *string) (bool, error) {
	if filter == nil {
		return true, nil
	}

	// check if *filter is valid json
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(*filter), &js); err != nil {
		return true, nil
	}

	query := &queryFilter{}
	if err := json.Unmarshal([]byte(*filter), query); err != nil {
		return false, err
	}

	return query.exists, nil
}
