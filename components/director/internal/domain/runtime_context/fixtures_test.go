package runtimectx_test

import (
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	id = "id"
	runtimeID = "runtimeID"
	key = "key"
	val = "val"

	tenantID = "tenantID"
)

func fixModelRuntimeCtx() *model.RuntimeContext {
	return &model.RuntimeContext{
		ID:        id,
		RuntimeID: runtimeID,
		Key:       key,
		Value:     val,
	}
}

func fixEntityRuntimeCtx() *runtimectx.RuntimeContext {
	return &runtimectx.RuntimeContext{
		ID:        id,
		RuntimeID: runtimeID,
		Key:       key,
		Value:     val,
	}
}