package runtime

import (
	"database/sql"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type runtimeStatusCondition string

// Runtime struct represents database entity for Runtime
type Runtime struct {
	ID                string         `db:"id"`
	TenantID          string         `db:"tenant_id"`
	Name              string         `db:"name"`
	Description       sql.NullString `db:"description"`
	StatusCondition   string         `db:"status_condition"`
	StatusTimestamp   time.Time      `db:"status_timestamp"`
	CreationTimestamp time.Time      `db:"creation_timestamp"`
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

	return &Runtime{
		ID:                model.ID,
		TenantID:          model.Tenant,
		Name:              model.Name,
		Description:       nullDescription,
		StatusCondition:   string(model.Status.Condition),
		StatusTimestamp:   model.Status.Timestamp,
		CreationTimestamp: model.CreationTimestamp,
	}, nil
}

// GraphQLToModel converts Runtime entity to Runtime model
func (e Runtime) ToModel() (*model.Runtime, error) {
	var description *string
	if e.Description.Valid {
		description = new(string)
		*description = e.Description.String
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
		CreationTimestamp: e.CreationTimestamp,
	}, nil
}
