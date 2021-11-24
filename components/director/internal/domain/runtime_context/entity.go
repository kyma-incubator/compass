package runtimectx

// RuntimeContext struct represents database entity for RuntimeContext
type RuntimeContext struct {
	ID        string `db:"id"`
	RuntimeID string `db:"runtime_id"`
	Key       string `db:"key"`
	Value     string `db:"value"`
}

// GetID returns ID of RuntimeContext
func (e *RuntimeContext) GetID() string {
	return e.ID
}

// GetParentID returns ID of parent entity
func (e *RuntimeContext) GetParentID() string {
	return e.RuntimeID
}
