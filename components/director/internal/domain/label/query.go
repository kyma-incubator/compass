package label

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// SetCombination type defines possible result set combination for querying
type SetCombination string

const (
	IntersectSet           SetCombination = "INTERSECT"
	ExceptSet              SetCombination = "EXCEPT"
	UnionSet               SetCombination = "UNION"
	scenariosLabelKey      string         = "SCENARIOS"
	stmtPrefixFormat       string         = `SELECT "%s" FROM %s WHERE "%s" IS NOT NULL AND "tenant_id" = ?`
	stmtPrefixGlobalFormat string         = `SELECT "%s" FROM %s WHERE "%s" IS NOT NULL`
)

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
func FilterQueryGlobal(queryFor model.LabelableObject, setCombination SetCombination, filter []*labelfilter.LabelFilter) (string, []interface{}, error) {
	if filter == nil {
		return "", nil, nil
	}

	objectField := labelableObjectField(queryFor)

	stmtPrefix := fmt.Sprintf(stmtPrefixGlobalFormat, objectField, tableName, objectField)

	return buildFilterQuery(stmtPrefix, nil, setCombination, filter, false)
}

func filterQuery(queryFor model.LabelableObject, setCombination SetCombination, tenant uuid.UUID, filter []*labelfilter.LabelFilter, isSubQuery bool) (string, []interface{}, error) {
	if filter == nil {
		return "", nil, nil
	}

	objectField := labelableObjectField(queryFor)

	stmtPrefix := fmt.Sprintf(stmtPrefixFormat, objectField, tableName, objectField)
	var stmtPrefixArgs []interface{}
	stmtPrefixArgs = append(stmtPrefixArgs, tenant)

	return buildFilterQuery(stmtPrefix, stmtPrefixArgs, setCombination, filter, isSubQuery)
}

func buildFilterQuery(stmtPrefix string, stmtPrefixArgs []interface{}, setCombination SetCombination, filter []*labelfilter.LabelFilter, isSubQuery bool) (string, []interface{}, error) {
	var queryBuilder strings.Builder

	var args []interface{}
	for idx, lblFilter := range filter {
		if idx > 0 || isSubQuery {
			queryBuilder.WriteString(fmt.Sprintf(` %s `, setCombination))
		}

		queryBuilder.WriteString(stmtPrefix)
		if stmtPrefixArgs != nil && len(stmtPrefixArgs) > 0 {
			args = append(args, stmtPrefixArgs...)
		}

		// TODO: for optimization it can be detected if the given Key was already added to the query
		// if so, it can be omitted
		queryBuilder.WriteString(` AND "key" = ?`)

		args = append(args, lblFilter.Key)

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
			} else {
				args = append(args, queryValue)
				queryBuilder.WriteString(` AND "value" @> ?`)
			}
		}
	}

	return queryBuilder.String(), args, nil
}
