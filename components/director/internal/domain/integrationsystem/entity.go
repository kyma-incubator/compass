package integrationsystem

// Entity missing godoc
type Entity struct {
	ID          string  `db:"id"`
	Name        string  `db:"name"`
	Description *string `db:"description"`
}

// Collection missing godoc
type Collection []Entity

// Len missing godoc
func (c Collection) Len() int {
	return len(c)
}
