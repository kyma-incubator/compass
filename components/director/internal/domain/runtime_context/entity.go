package runtimectx

// RuntimeContext struct represents database entity for RuntimeContext
type RuntimeContext struct {
	ID        string `db:"id"`
	RuntimeID string `db:"runtime_id"`
	Key       string `db:"key"`
	Value     string `db:"value"`
}

func (e *RuntimeContext) GetID() string {
	return e.ID
}

func (e *RuntimeContext) GetParentID() string {
	return e.RuntimeID
}
