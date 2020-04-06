package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

type engine struct {
}

func NewEngine() *engine {
	return &engine{}
}

const updateQuery = `UPDATE labels AS l SET value=SCENARIOS.SCENARIOS 
		FROM (SELECT array_to_json(array_agg(scenario)) AS SCENARIOS FROM automatic_scenario_assignments 
					WHERE selector_key=$1 AND selector_value=$2 AND tenant_id=$3) AS SCENARIOS
		WHERE l.runtime_id IN (SELECT runtime_id FROM labels  
									WHERE key =$1 AND value ?| array[$2] AND runtime_id IS NOT NULL AND tenant_ID=$3) 
			AND l.key ='scenarios'
			AND l.tenant_id=$3`

func (e *engine) EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment, tenantID string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while getting persitance from context")
	}

	_, err = persist.Exec(updateQuery, in.Selector.Key, in.Selector.Value, tenantID)
	if err != nil {
		return errors.Wrap(err, "while updating scenarios")
	}

	return nil
}

func (engine) RemoveAssignedScenario(in model.AutomaticScenarioAssignment) error {
	// TODO: Implement it

	// remove scenario from runtimes, which have label matching selector
	return nil
}

func (engine) RemoveAssignedScenarios(in []*model.AutomaticScenarioAssignment) error {
	// TODO: Implement it

	// remove scenarios from runtimes, which have label matching selector
	return nil
}
