package label

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// NewConverter missing godoc
func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

// ToEntity missing godoc
func (c *converter) ToEntity(in *model.Label) (*Entity, error) {
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
	var rtmCtxID sql.NullString
	var appTmplID sql.NullString
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
	case model.AppTemplateLabelableObject:
		appTmplID = sql.NullString{
			Valid:  true,
			String: in.ObjectID,
		}
	}

	return &Entity{
		ID:               in.ID,
		TenantID:         repo.NewNullableString(in.Tenant),
		AppID:            appID,
		RuntimeID:        rtmID,
		RuntimeContextID: rtmCtxID,
		AppTemplateID:    appTmplID,
		Key:              in.Key,
		Value:            string(valueMarshalled),
		Version:          in.Version,
	}, nil
}

// FromEntity missing godoc
func (c *converter) FromEntity(in *Entity) (*model.Label, error) {
	var valueUnmarshalled interface{}
	if in.Value != "" {
		err := json.Unmarshal([]byte(in.Value), &valueUnmarshalled)
		if err != nil {
			return nil, errors.Wrap(err, "while unmarshalling Value")
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

	return &model.Label{
		ID:         in.ID,
		Tenant:     repo.StringPtrFromNullableString(in.TenantID),
		ObjectID:   objectID,
		ObjectType: objectType,
		Key:        in.Key,
		Value:      valueUnmarshalled,
		Version:    in.Version,
	}, nil
}
