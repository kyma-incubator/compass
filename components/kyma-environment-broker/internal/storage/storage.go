package storage

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/driver/memory"
	postgres "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/driver/postsql"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"
)

//go:generate mockery -name=Instances -output=automock -outpkg=automock -case=underscore
type BrokerStorage interface {
	Instances() Instances
	Operations() Operations
}

func New(connectionURL string) (BrokerStorage, error) {
	connection, err := postsql.InitializeDatabase(connectionURL)
	if err != nil {
		return nil, err
	}
	fact := dbsession.NewFactory(connection)

	return storage{
		instance:  postgres.NewInstance(fact),
		operation: postgres.NewOperation(fact),
	}, nil
}

func NewMemoryStorage() BrokerStorage {
	return storage{
		instance:  memory.NewInstance(),
		operation: memory.NewOperation(),
	}
}

type storage struct {
	instance  Instances
	operation Operations
}

func (s storage) Instances() Instances {
	return s.instance
}

func (s storage) Operations() Operations {
	return s.operation
}
