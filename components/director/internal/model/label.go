package model

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
	RuntimeLabelableObject     LabelableObject = "Runtime"
	ApplicationLabelableObject LabelableObject = "Application"
)
