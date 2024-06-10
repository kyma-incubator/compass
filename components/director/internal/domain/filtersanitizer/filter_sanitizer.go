package filtersanitizer

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// ObjectIDListerFuncGlobal is a function that returns object IDs for given formation names and object type globally
type ObjectIDListerFuncGlobal func(ctx context.Context, formationNames []string, objectType model.FormationAssignmentType) ([]string, error)

// ObjectIDListerFunc is a function that returns object IDs for given tenant, formation names and object type
type ObjectIDListerFunc func(ctx context.Context, tenantID string, formationNames []string, objectType model.FormationAssignmentType) ([]string, error)

// FilterSanitizer is responsible for removing scenarios filter from filters and returning object IDs in scenarios
type FilterSanitizer struct {
}

// RemoveScenarioFilter removes scenarios filter from filters and returns object IDs in scenarios
func (f *FilterSanitizer) RemoveScenarioFilter(ctx context.Context, tenant string, filters []*labelfilter.LabelFilter, objectType model.FormationAssignmentType, isGlobal bool, listerFunc ObjectIDListerFunc, globalLiserFunc ObjectIDListerFuncGlobal) (bool, []string, []*labelfilter.LabelFilter, error) {
	var hasScenarioFilter bool
	var objectIDsInScenarios = make([]string, 0)
	filtersWithoutScenarioFilter := make([]*labelfilter.LabelFilter, 0, len(filters))

	log.C(ctx).Debug("Checking for scenarios filter in filters")
	for i, labelFilter := range filters {
		if labelFilter.Key == model.ScenariosKey {
			log.C(ctx).Debug("Found scenarios filter")
			hasScenarioFilter = true
			formationNamesInterface, err := label.ExtractValueFromJSONPath(*labelFilter.Query)
			if err != nil {
				return hasScenarioFilter, nil, nil, errors.Wrap(err, "while extracting formation names from JSON path")
			}

			formationNames, err := label.ValueToStringsSlice(formationNamesInterface)
			if err != nil {
				return hasScenarioFilter, nil, nil, errors.Wrap(err, "while converting formation names to strings")
			}

			log.C(ctx).Debugf("Scenarios found in filter: %v", formationNamesInterface)
			if isGlobal {
				objectIDsInScenarios, err = globalLiserFunc(ctx, formationNames, objectType)
				if err != nil {
					return hasScenarioFilter, nil, nil, errors.Wrapf(err, "while getting object IDs for formations %v", formationNames)
				}
			} else {
				objectIDsInScenarios, err = listerFunc(ctx, tenant, formationNames, objectType)
				if err != nil {
					return hasScenarioFilter, nil, nil, errors.Wrapf(err, "while getting object IDs for formations %v", formationNames)
				}
			}
			log.C(ctx).Debug("Removed scenarios filter from initial filters")
		} else {
			filtersWithoutScenarioFilter = append(filtersWithoutScenarioFilter, filters[i])
		}
	}
	return hasScenarioFilter, objectIDsInScenarios, filtersWithoutScenarioFilter, nil
}
