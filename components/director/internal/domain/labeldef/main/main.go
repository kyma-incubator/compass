package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/sirupsen/logrus"
)

func main() {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", "127.0.0.1", "5432", "usr",
		"pwd", "compass", "disable")
	transact, closeFunc, err := persistence.Configure(logrus.StandardLogger(), connString)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := closeFunc()
		if err != nil {
			panic(err)
		}
	}()

	repo := labeldef.NewRepository(labeldef.NewConverter())

	tx, err := transact.Begin()
	if err != nil {
		panic(err)
	}

	defer transact.RollbackUnlessCommited(tx)
	ctx := persistence.SaveToContext(context.TODO(), tx)

	u := User{}
	var ptr interface{}
	ptr = u

	err = repo.Create(ctx, model.LabelDefinition{
		ID:     uuid.New().String(),
		Tenant: uuid.New().String(),
		Key:    "aaa-with-schema",
		Schema: &ptr,
	})

	if err != nil {
		panic(err)
	}

	err = repo.Create(ctx, model.LabelDefinition{
		ID:     uuid.New().String(),
		Tenant: uuid.New().String(),
		Key:    "aaa-without-schema",
		Schema: nil,
	})

	if err != nil {
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

}

type User struct {
	ID        string `db:"id_col"`
	Tenant    string `db:"tenant_col"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Age       int
}
