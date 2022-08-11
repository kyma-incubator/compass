package runtimectx_test

import (
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	id                     = "id"
	runtimeID              = "runtimeID"
	runtimeID2             = "runtimeID2"
	emptyPageRuntimeID     = "emtyPageRuntimeID"
	onePageRuntimeID       = "onePageRuntimeID"
	multiplePagesRuntimeID = "multiplePagesRuntimeID"
	key                    = "key"
	val                    = "val"

	tenantID     = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	runtimeCtxID = "runtimeCtxID"
)

var fixColumns = []string{"id", "runtime_id", "key", "value"}

func fixModelRuntimeCtx() *model.RuntimeContext {
	return fixModelRuntimeCtxWithID(runtimeCtxID)
}

func fixModelRuntimeCtxWithID(id string) *model.RuntimeContext {
	return &model.RuntimeContext{
		ID:        id,
		RuntimeID: runtimeID,
		Key:       key,
		Value:     val,
	}
}

func fixModelRuntimeCtxWithIDAndRuntimeID(id, rtmID string) *model.RuntimeContext {
	rtmCtx := fixModelRuntimeCtxWithID(id)
	rtmCtx.RuntimeID = rtmID
	return rtmCtx
}

func fixEntityRuntimeCtx() *runtimectx.RuntimeContext {
	return fixEntityRuntimeCtxWithID(runtimeCtxID)
}

func fixEntityRuntimeCtxWithID(id string) *runtimectx.RuntimeContext {
	return &runtimectx.RuntimeContext{
		ID:        id,
		RuntimeID: runtimeID,
		Key:       key,
		Value:     val,
	}
}

func fixEntityRuntimeCtxWithIDAndRuntimeID(id, rtmID string) *runtimectx.RuntimeContext {
	rtmCtx := fixEntityRuntimeCtxWithID(id)
	rtmCtx.RuntimeID = rtmID
	return rtmCtx
}
