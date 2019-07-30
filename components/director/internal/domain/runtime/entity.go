package runtime

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

type runtimeStatusCondition string

// Runtime struct represents database entity for Runtime
type Runtime struct {
	ID              string         `db:"id"`
	TenantID        string         `db:"tenant_id"`
	Name            string         `db:"name"`
	Description     sql.NullString `db:"description"`
	StatusCondition string         `db:"status_condition"`
	StatusTimestamp time.Time      `db:"status_timestamp"`
	AgentAuth       string         `db:"auth"`
}

// EntityFromRuntimeModel converts Runtime model to Runtime entity
func EntityFromRuntimeModel(model *model.Runtime) (*Runtime, error) {
	var nullDescription sql.NullString
	if model.Description != nil && len(*model.Description) > 0 {
		nullDescription = sql.NullString{
			String: *model.Description,
			Valid:  true,
		}
	}

	agentAuthMarshalled, err := json.Marshal(model.AgentAuth)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling AgentAuth")
	}

	return &Runtime{
		ID:              model.ID,
		TenantID:        model.Tenant,
		Name:            model.Name,
		Description:     nullDescription,
		StatusCondition: string(model.Status.Condition),
		StatusTimestamp: model.Status.Timestamp,
		AgentAuth:       string(agentAuthMarshalled),
	}, nil
}

// ToModel converts Runtime entity to Runtime model
func (e Runtime) ToModel() (*model.Runtime, error) {
	var description *string
	if e.Description.Valid {
		description = new(string)
		*description = e.Description.String
	}

	var agentAuth model.Auth
	err := json.Unmarshal([]byte(e.AgentAuth), &agentAuth)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling AgentAuth")
	}

	return &model.Runtime{
		ID:          e.ID,
		Tenant:      e.TenantID,
		Name:        e.Name,
		Description: description,
		Status: &model.RuntimeStatus{
			Condition: model.RuntimeStatusCondition(e.StatusCondition),
			Timestamp: e.StatusTimestamp,
		},
		AgentAuth: &agentAuth,
	}, nil
}
