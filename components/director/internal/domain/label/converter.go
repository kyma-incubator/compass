package label

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// NewConverter missing godoc
func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

// ToEntity missing godoc
func (c *converter) ToEntity(in *model.Label) (*Entity, error) {
	value, err := in.GetValue()
	if err != nil {
		return nil, err
	}

	var appID sql.NullString
	var rtmID sql.NullString
	var rtmCtxID sql.NullString
	var appTmplID sql.NullString
	var webhookID sql.NullString
	var formationTmplID sql.NullString
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
	case model.WebhookLabelableObject:
		webhookID = sql.NullString{
			Valid:  true,
			String: in.ObjectID,
		}
	case model.FormationTemplateLabelableObject:
		formationTmplID = sql.NullString{
			Valid:  true,
			String: in.ObjectID,
		}
	}

	return &Entity{
		ID:                  in.ID,
		TenantID:            repo.NewNullableString(in.Tenant),
		AppID:               appID,
		RuntimeID:           rtmID,
		RuntimeContextID:    rtmCtxID,
		AppTemplateID:       appTmplID,
		WebhookID:           webhookID,
		FormationTemplateID: formationTmplID,
		Key:                 in.Key,
		Value:               value,
		Version:             in.Version,
	}, nil
}

// FromEntity missing godoc
func (c *converter) FromEntity(in *Entity) (*model.Label, error) {
	value, err := in.GetValue()
	if err != nil {
		return nil, err
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
	} else if in.AppTemplateID.Valid {
		objectID = in.AppTemplateID.String
		objectType = model.AppTemplateLabelableObject
	} else if in.WebhookID.Valid {
		objectID = in.WebhookID.String
		objectType = model.WebhookLabelableObject
	} else if in.FormationTemplateID.Valid {
		objectID = in.FormationTemplateID.String
		objectType = model.FormationTemplateLabelableObject
	}

	return &model.Label{
		ID:         in.ID,
		Tenant:     repo.StringPtrFromNullableString(in.TenantID),
		ObjectID:   objectID,
		ObjectType: objectType,
		Key:        in.Key,
		Value:      value,
		Version:    in.Version,
	}, nil
}
