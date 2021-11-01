package model

import "github.com/google/uuid"

// Label missing godoc
type Label struct {
	ID         string
	Tenant     string
	Key        string
	Value      interface{}
	ObjectID   string
	ObjectType LabelableObject
	Version    int
}

// LabelInput missing godoc
type LabelInput struct {
	Key        string
	Value      interface{}
	ObjectID   string
	ObjectType LabelableObject
	Version    int
}

// ToLabel missing godoc
func (i *LabelInput) ToLabel(id, tenant string) *Label {
	var tenantID string
	if i.Key == ScenariosKey || i.ObjectType == TenantLabelableObject {
		tenantID = tenant
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

// LabelableObject missing godoc
type LabelableObject string

const (
	// RuntimeLabelableObject missing godoc
	RuntimeLabelableObject LabelableObject = "Runtime"
	// RuntimeContextLabelableObject missing godoc
	RuntimeContextLabelableObject LabelableObject = "Runtime Context"
	// ApplicationLabelableObject missing godoc
	ApplicationLabelableObject LabelableObject = "Application"
	// TenantLabelableObject missing godoc
	TenantLabelableObject LabelableObject = "Tenant"
)

// NewLabelForRuntimeContext missing godoc
func NewLabelForRuntimeContext(runtimeCtxID, tenant, key string, value interface{}) *Label {
	var tenantID string
	if key == ScenariosKey {
		tenantID = tenant
	}
	return &Label{
		ID:         uuid.New().String(),
		Tenant:     tenantID,
		ObjectType: RuntimeContextLabelableObject,
		ObjectID:   runtimeCtxID,
		Key:        key,
		Value:      value,
		Version:    0,
	}
}

// NewLabelForRuntime missing godoc
func NewLabelForRuntime(runtimeID, tenant, key string, value interface{}) *Label {
	var tenantID string
	if key == ScenariosKey {
		tenantID = tenant
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

// NewLabelForApplication missing godoc
func NewLabelForApplication(appID, tenant, key string, value interface{}) *Label {
	var tenantID string
	if key == ScenariosKey {
		tenantID = tenant
	}
	return &Label{
		ID:         uuid.New().String(),
		Tenant:     tenantID,
		ObjectType: ApplicationLabelableObject,
		ObjectID:   appID,
		Key:        key,
		Value:      value,
		Version:    0,
	}
}
