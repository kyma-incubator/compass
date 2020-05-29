package repo_test

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

// User is a exemplary type to test generic Repositories
type User struct {
	ID        string `db:"id_col"`
	Tenant    string `db:"tenant_id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Age       int
}

const UserType = resource.Type("UserType")

type UserCollection []User

func (u UserCollection) Len() int {
	return len(u)
}
