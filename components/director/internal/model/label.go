package model

import "github.com/google/uuid"

type Label struct {
	ID         string
	Tenant     string
	Key        string
	Value      interface{}
	ObjectID   string
	ObjectType LabelableObject
}

type LabelInput struct {
	Key        string
	Value      interface{}
	ObjectID   string
	ObjectType LabelableObject
}

func (i *LabelInput) ToLabel(id, tenant string) *Label {
	return &Label{
		ID:         id,
		Tenant:     tenant,
		ObjectType: i.ObjectType,
		ObjectID:   i.ObjectID,
		Key:        i.Key,
		Value:      i.Value,
	}
}

type LabelableObject string

const (
	RuntimeLabelableObject        LabelableObject = "Runtime"
	RuntimeContextLabelableObject LabelableObject = "Runtime Context"
	ApplicationLabelableObject    LabelableObject = "Application"
)

func NewLabelForRuntimeContext(runtimeCtx RuntimeContext, key string, value interface{}) *Label {
	return &Label{
		ID:         uuid.New().String(),
		Tenant:     runtimeCtx.Tenant,
		ObjectType: RuntimeContextLabelableObject,
		ObjectID:   runtimeCtx.ID,
		Key:        key,
		Value:      value,
	}
}

func NewLabelForRuntime(runtime Runtime, key string, value interface{}) *Label {
	return &Label{
		ID:         uuid.New().String(),
		Tenant:     runtime.Tenant,
		ObjectType: RuntimeLabelableObject,
		ObjectID:   runtime.ID,
		Key:        key,
		Value:      value,
	}
}

func NewLabelForApplication(app Application, key string, value interface{}) *Label {
	return &Label{
		ID:         uuid.New().String(),
		Tenant:     app.Tenant,
		ObjectType: ApplicationLabelableObject,
		ObjectID:   app.ID,
		Key:        key,
		Value:      value,
	}
}
