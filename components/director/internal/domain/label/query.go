package label

import (
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// SetCombination type defines possible result set combination for quering
type SetCombination string

const (
	IntersectSet      SetCombination = "INTERSECT"
	UnionSet          SetCombination = "UNION"
	scenariosLabelKey string         = "SCENARIOS"
	stmtPrefixFormat  string         = `SELECT "%s" FROM %s WHERE "%s" IS NOT NULL AND "tenant_id" = '%s'`
)

// FilterQuery builds select query for given filters
//
// It supports quering defined by `queryFor` parameter. All queries are created
// in the context of given tenant
func FilterQuery(queryFor model.LabelableObject, setCombination SetCombination, tenant uuid.UUID, filter []*labelfilter.LabelFilter) string {
	if filter == nil {
		return ""
	}

	objectField := labelableObjectField(queryFor)

	stmtPrefix := fmt.Sprintf(stmtPrefixFormat, objectField, tableName, objectField, tenant)

	var queryBuilder strings.Builder
	for idx, lblFilter := range filter {
		if idx > 0 {
			queryBuilder.WriteString(fmt.Sprintf(` %s `, setCombination))
		}

		queryBuilder.WriteString(fmt.Sprintf(stmtPrefix))

		// TODO: for optimization it can be detected if the given Key was already added to the query
		// if so, it can be ommited
		queryBuilder.WriteString(fmt.Sprintf(` AND "key" = %s`, pq.QuoteLiteral(lblFilter.Key)))

		if lblFilter.Query != nil {
			queryValue := *lblFilter.Query
			// Handling the Scenarios label case - we assume that Query is
			// in SQL/JSON path format supported by PostgreSQL 12. Till it
			// is not production ready, we need to transform the Query from
			// SQL/JSON path to old JSON queries.
			if strings.ToUpper(lblFilter.Key) == scenariosLabelKey {
				queryValue = `["` + *ExtractValueFromJSONPath(queryValue) + `"]`
			}

			queryBuilder.WriteString(fmt.Sprintf(` AND "value" @> %s`, pq.QuoteLiteral(queryValue)))
		}
	}

	return queryBuilder.String()
}
