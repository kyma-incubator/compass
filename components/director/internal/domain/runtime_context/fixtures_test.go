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

	tenantID = "b91b59f7-2563-40b2-aba9-fef726037aa3"
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