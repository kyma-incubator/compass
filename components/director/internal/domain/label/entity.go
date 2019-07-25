package label

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

type Entity struct {
	ID        string         `db:"id"`
	TenantID  string         `db:"tenant_id"`
	Key       string         `db:"key"`
	AppID     sql.NullString `db:"app_id"`
	RuntimeID sql.NullString `db:"runtime_id"`
	Value     string         `db:"value"`
}

// EntityFromRModel converts Label model to Label entity
func EntityFromModel(in *model.Label) (*Entity, error) {
	if in == nil {
		return nil, nil
	}

	var valueMarshalled []byte
	var err error

	if in.Value != nil {
		valueMarshalled, err = json.Marshal(in.Value)
		if err != nil {
			return nil, errors.Wrap(err, "while marshalling Value")
		}
	}

	var appID sql.NullString
	var rtmID sql.NullString
	switch in.ObjectType {
	case model.ApplicationLabelableObject:
		appID = sql.NullString{
			Valid:  true,
			String: in.ObjectID,
		}
	case model.RuntimeLabelableObject:
		rtmID = sql.NullString{
			Valid:  true,
			String: in.ObjectID,
		}
	}

	return &Entity{
		ID:        in.ID,
		TenantID:  in.Tenant,
		AppID:     appID,
		RuntimeID: rtmID,
		Key:       in.Key,
		Value:     string(valueMarshalled),
	}, nil
}

// ToModel converts Entity entity to Runtime model
func (e *Entity) ToModel() (model.Label, error) {
	if e == nil {
		return model.Label{}, nil
	}

	var valueUnmarshalled interface{}
	if e.Value != "" {
		err := json.Unmarshal([]byte(e.Value), &valueUnmarshalled)
		if err != nil {
			return model.Label{}, errors.Wrap(err, "while unmarshalling Value")
		}
	}

	var objectType model.LabelableObject
	var objectID string

	if e.AppID.Valid {
		objectID = e.AppID.String
		objectType = model.ApplicationLabelableObject
	} else if e.RuntimeID.Valid {
		objectID = e.RuntimeID.String
		objectType = model.RuntimeLabelableObject
	}

	return model.Label{
		ID:         e.ID,
		Tenant:     e.TenantID,
		ObjectID:   objectID,
		ObjectType: objectType,
		Key:        e.Key,
		Value:      valueUnmarshalled,
	}, nil
}
