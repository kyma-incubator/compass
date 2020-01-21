package repo_test

// User is a exemplary type to test generic Repositories
type User struct {
	ID        string `db:"id_col"`
	Tenant    string `db:"tenant_id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Age       int
}

type UserCollection []User

func (u UserCollection) Len() int {
	return len(u)
}
