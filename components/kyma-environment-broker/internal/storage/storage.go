package storage

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession"
	memory "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/driver/memory"
	postgres "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/driver/postsql"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"
)

//go:generate mockery -name=Instances -output=automock -outpkg=automock -case=underscore
type BrokerStorage interface {
	Instances() Instances
}

func New(connectionURL string) (BrokerStorage, error) {
	connection, err := postsql.InitializeDatabase(connectionURL)
	if err != nil {
		return nil, err
	}
	fact := dbsession.NewFactory(connection)

	return storage{
		instance: postgres.NewInstance(fact),
	}, nil
}

func NewMemoryStorage() BrokerStorage {
	return storage{
		instance: memory.NewInstance(),
	}
}

type storage struct {
	instance Instances
}

func (s storage) Instances() Instances {
	return s.instance
}
