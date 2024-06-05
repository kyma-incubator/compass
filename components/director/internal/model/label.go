package model

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// Label represents a label with additional metadata for a given entity.
type Label struct {
	ID         string
	Tenant     *string
	Key        string
	Value      interface{}
	ObjectID   string
	ObjectType LabelableObject
	Version    int
}

// GetValue returns a Label value as string.
func (l *Label) GetValue() (string, error) {
	var valueMarshalled []byte
	var err error

	if l.Value != nil {
		valueMarshalled, err = json.Marshal(l.Value)
		if err != nil {
			return "", errors.Wrap(err, "while marshalling Value")
		}
		if len(valueMarshalled) > 0 {
			return string(valueMarshalled), nil
		}
	}
	return "", errors.Errorf("Value cannot be empty")
}

// LabelInput is an input for creating a new label.
type LabelInput struct {
	Key        string
	Value      interface{}
	ObjectID   string
	ObjectType LabelableObject
	Version    int
}

// ToLabel converts a LabelInput to a Label.
func (i *LabelInput) ToLabel(id, tenant string) *Label {
	var tenantID *string
	if i.Key == ScenariosKey || i.ObjectType == TenantLabelableObject {
		tenantID = &tenant
	}

	return &Label{
		ID:         id,
		Tenant:     tenantID,
		ObjectType: i.ObjectType,
		ObjectID:   i.ObjectID,
		Key:        i.Key,
		Value:      i.Value,
		Version:    i.Version,
	}
}

// LabelableObject represents the type of entity that can have labels.
type LabelableObject string

const (
	// RuntimeLabelableObject represents a runtime entity.
	RuntimeLabelableObject LabelableObject = "Runtime"
	// RuntimeContextLabelableObject represents a runtime context entity.
	RuntimeContextLabelableObject LabelableObject = "Runtime Context"
	// ApplicationLabelableObject represents an application entity.
	ApplicationLabelableObject LabelableObject = "Application"
	// TenantLabelableObject represents a tenant entity.
	TenantLabelableObject LabelableObject = "Tenant"
	// AppTemplateLabelableObject represents an application template entity.
	AppTemplateLabelableObject LabelableObject = "Application Template"
	// WebhookLabelableObject represents a webhook entity.
	WebhookLabelableObject LabelableObject = "Webhook"
	// FormationTemplateLabelableObject represents a formation template entity.
	FormationTemplateLabelableObject LabelableObject = "Formation Template"
)

// GetResourceType returns the resource type of the label based on the referenced entity.
func (obj LabelableObject) GetResourceType() resource.Type {
	switch obj {
	case RuntimeLabelableObject:
		return resource.RuntimeLabel
	case RuntimeContextLabelableObject:
		return resource.RuntimeContextLabel
	case ApplicationLabelableObject:
		return resource.ApplicationLabel
	case TenantLabelableObject:
		return resource.TenantLabel
	case WebhookLabelableObject:
		return resource.Webhook
	case FormationTemplateLabelableObject:
		return resource.FormationTemplate
	}
	return ""
}

// NewLabelForRuntime creates a new label for a runtime.
func NewLabelForRuntime(runtimeID, tenant, key string, value interface{}) *Label {
	var tenantID *string
	if key == ScenariosKey {
		tenantID = &tenant
	}
	return &Label{
		ID:         uuid.New().String(),
		Tenant:     tenantID,
		ObjectType: RuntimeLabelableObject,
		ObjectID:   runtimeID,
		Key:        key,
		Value:      value,
		Version:    0,
	}
}
