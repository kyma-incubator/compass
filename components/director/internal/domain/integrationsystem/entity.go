package integrationsystem

type Entity struct {
	ID          string  `db:"id"`
	Name        string  `db:"name"`
	Description *string `db:"description"`
}

type Collection []Entity

func (c Collection) Len() int {
	return len(c)
}
