package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

// RuntimeContext missing godoc
type RuntimeContext struct {
	ID        string
	RuntimeID string
	Key       string
	Value     string
}

// GetID missing godoc
func (r *RuntimeContext) GetID() string {
	return r.ID
}

// RuntimeContextInput missing godoc
type RuntimeContextInput struct {
	Key       string
	Value     string
	RuntimeID string
}

// ToRuntimeContext missing godoc
func (i *RuntimeContextInput) ToRuntimeContext(id string) *RuntimeContext {
	if i == nil {
		return nil
	}

	return &RuntimeContext{
		ID:        id,
		RuntimeID: i.RuntimeID,
		Key:       i.Key,
		Value:     i.Value,
	}
}

// RuntimeContextPage missing godoc
type RuntimeContextPage struct {
	Data       []*RuntimeContext
	PageInfo   *pagination.Page
	TotalCount int
}
