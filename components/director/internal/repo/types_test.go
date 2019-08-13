package repo_test

type User struct {
	ID        string `db:"id_col"`
	Tenant    string `db:"tenant_col"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Age       int
}

type UserCollection []User

func (u UserCollection) Len() int {
	return len(u)
}
