package label_test

import (
	"database/sql"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

const (
	labelId  = "lblId"
	refID    = "refID"
	key      = "test"
	value    = "test"
	tenantID = "tenant"
)

func fixModelLabel(objectType model.LabelableObject) *model.Label {
	result := &model.Label{
		ID:         labelId,
		Key:        key,
		Value:      value,
		ObjectID:   refID,
		ObjectType: objectType,
	}
	if objectType == model.TenantLabelableObject {
		result.Tenant = str.Ptr(tenantID)
	}
	return result
}

func fixEntityLabel(objectType model.LabelableObject) *label.Entity {
	var tenant sql.NullString
	var appID sql.NullString
	var runtimeCtxID sql.NullString
	var runtimeID sql.NullString
	switch objectType {
	case model.RuntimeContextLabelableObject:
		runtimeCtxID = sql.NullString{String: refID, Valid: true}
	case model.RuntimeLabelableObject:
		runtimeID = sql.NullString{String: refID, Valid: true}
	case model.ApplicationLabelableObject:
		appID = sql.NullString{String: refID, Valid: true}
	case model.TenantLabelableObject:
		tenant = sql.NullString{String: tenantID, Valid: true}
	}

	return &label.Entity{
		Key:              key,
		Value:            value,
		ID:               labelId,
		TenantID:         tenant,
		AppID:            appID,
		RuntimeContextID: runtimeCtxID,
		RuntimeID:        runtimeID,
	}
}
