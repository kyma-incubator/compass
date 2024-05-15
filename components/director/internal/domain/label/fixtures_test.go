package label_test

import (
	"database/sql"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

const (
	labelID                 = "lblId"
	refID                   = "refID"
	key                     = "test-label-key"
	value                   = "test-label-value"
	tenantID                = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	testFormationTemplateID = "test-formation-template-id"
)

var (
	fixColumns  = []string{"id", "tenant_id", "app_id", "runtime_id", "runtime_context_id", "app_template_id", "formation_template_id", "key", "value", "version", "webhook_id"}
	testErr     = errors.New("Test error")
	modelLabel  = fixModelLabel(model.FormationTemplateLabelableObject)
	modelLabels = map[string]*model.Label{key: modelLabel}
)

func fixModelLabel(objectType model.LabelableObject) *model.Label {
	return fixModelLabelWithID(labelID, key, objectType)
}

func fixModelLabelWithID(id, key string, objectType model.LabelableObject) *model.Label {
	return fixModelLabelWithRefID(id, key, objectType, refID)
}

func fixModelLabelWithRefID(id, key string, objectType model.LabelableObject, refID string) *model.Label {
	result := &model.Label{
		ID:         id,
		Key:        key,
		Value:      value,
		ObjectID:   refID,
		ObjectType: objectType,
		Version:    42,
	}
	if objectType == model.TenantLabelableObject {
		result.ObjectID = ""
		result.Tenant = str.Ptr(refID)
	}
	return result
}

func fixEntityLabel(objectType model.LabelableObject) *label.Entity {
	return fixEntityLabelWithID(labelID, key, objectType)
}

func fixEntityLabelWithID(id, key string, objectType model.LabelableObject) *label.Entity {
	return fixEntityLabelWithRefID(id, key, objectType, refID)
}

func fixEntityLabelWithRefID(id, key string, objectType model.LabelableObject, refID string) *label.Entity {
	var tenant sql.NullString
	var appID sql.NullString
	var runtimeCtxID sql.NullString
	var runtimeID sql.NullString
	var appTmplID sql.NullString
	var webhookID sql.NullString
	var formationTemplateID sql.NullString

	switch objectType {
	case model.RuntimeContextLabelableObject:
		runtimeCtxID = sql.NullString{String: refID, Valid: true}
	case model.RuntimeLabelableObject:
		runtimeID = sql.NullString{String: refID, Valid: true}
	case model.ApplicationLabelableObject:
		appID = sql.NullString{String: refID, Valid: true}
	case model.TenantLabelableObject:
		tenant = sql.NullString{String: tenantID, Valid: true}
	case model.AppTemplateLabelableObject:
		appTmplID = sql.NullString{String: refID, Valid: true}
	case model.WebhookLabelableObject:
		webhookID = sql.NullString{String: refID, Valid: true}
	case model.FormationTemplateLabelableObject:
		formationTemplateID = sql.NullString{String: refID, Valid: true}
	}

	return &label.Entity{
		Key:                 key,
		Value:               value,
		ID:                  id,
		TenantID:            tenant,
		AppID:               appID,
		AppTemplateID:       appTmplID,
		RuntimeContextID:    runtimeCtxID,
		RuntimeID:           runtimeID,
		WebhookID:           webhookID,
		FormationTemplateID: formationTemplateID,
		Version:             42,
	}
}
