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
	return &Label{
		ID:         id,
		Tenant:     tenant,
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
func NewLabelForRuntimeContext(runtimeCtx RuntimeContext, key string, value interface{}) *Label {
	return &Label{
		ID:         uuid.New().String(),
		Tenant:     runtimeCtx.Tenant,
		ObjectType: RuntimeContextLabelableObject,
		ObjectID:   runtimeCtx.ID,
		Key:        key,
		Value:      value,
		Version:    0,
	}
}

// NewLabelForRuntime missing godoc
func NewLabelForRuntime(runtime Runtime, key string, value interface{}) *Label {
	return &Label{
		ID:         uuid.New().String(),
		Tenant:     runtime.Tenant,
		ObjectType: RuntimeLabelableObject,
		ObjectID:   runtime.ID,
		Key:        key,
		Value:      value,
		Version:    0,
	}
}

// NewLabelForApplication missing godoc
func NewLabelForApplication(app Application, key string, value interface{}) *Label {
	return &Label{
		ID:         uuid.New().String(),
		Tenant:     app.Tenant,
		ObjectType: ApplicationLabelableObject,
		ObjectID:   app.ID,
		Key:        key,
		Value:      value,
		Version:    0,
	}
}
