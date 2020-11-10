package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type RuntimeContext struct {
	ID        string
	RuntimeID string
	Tenant    string
	Key       string
	Value     string
}

type RuntimeContextInput struct {
	Key       string
	Value     string
	RuntimeID string
	Labels    map[string]interface{}
}

func (i *RuntimeContextInput) ToRuntimeContext(id, tenant string) *RuntimeContext {
	if i == nil {
		return nil
	}

	return &RuntimeContext{
		ID:        id,
		RuntimeID: i.RuntimeID,
		Tenant:    tenant,
		Key:       i.Key,
		Value:     i.Value,
	}
}

type RuntimeContextPage struct {
	Data       []*RuntimeContext
	PageInfo   *pagination.Page
	TotalCount int
}
