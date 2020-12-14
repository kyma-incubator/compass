package label

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

func (c *converter) ToEntity(in model.Label) (Entity, error) {
	var valueMarshalled []byte
	var err error

	if in.Value != nil {
		valueMarshalled, err = json.Marshal(in.Value)
		if err != nil {
			return Entity{}, errors.Wrap(err, "while marshalling Value")
		}
	}

	var appID sql.NullString
	var rtmID sql.NullString
	var rtmCtxID sql.NullString
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
	case model.RuntimeContextLabelableObject:
		rtmCtxID = sql.NullString{
			Valid:  true,
			String: in.ObjectID,
		}
	}

	return Entity{
		ID:               in.ID,
		TenantID:         in.Tenant,
		AppID:            appID,
		RuntimeID:        rtmID,
		RuntimeContextID: rtmCtxID,
		Key:              in.Key,
		Value:            string(valueMarshalled),
	}, nil
}

func (c *converter) FromEntity(in Entity) (model.Label, error) {
	var valueUnmarshalled interface{}
	if in.Value != "" {
		err := json.Unmarshal([]byte(in.Value), &valueUnmarshalled)
		if err != nil {
			return model.Label{}, errors.Wrap(err, "while unmarshalling Value")
		}
	}

	var objectType model.LabelableObject
	var objectID string

	if in.AppID.Valid {
		objectID = in.AppID.String
		objectType = model.ApplicationLabelableObject
	} else if in.RuntimeID.Valid {
		objectID = in.RuntimeID.String
		objectType = model.RuntimeLabelableObject
	} else if in.RuntimeContextID.Valid {
		objectID = in.RuntimeContextID.String
		objectType = model.RuntimeContextLabelableObject
	}

	return model.Label{
		ID:         in.ID,
		Tenant:     in.TenantID,
		ObjectID:   objectID,
		ObjectType: objectType,
		Key:        in.Key,
		Value:      valueUnmarshalled,
	}, nil
}
