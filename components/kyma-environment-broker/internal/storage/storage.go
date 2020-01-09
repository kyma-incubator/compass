package storage

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/entity"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"
	"github.com/pkg/errors"
)

type BrokerStorage interface {
	Instances() Instances
}

func New(connectionURL string) (BrokerStorage, error) {
	connection, err := postsql.InitializeDatabase(connectionURL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize database")
	}
	fact := dbsession.NewFactory(connection)

	return storage{
		instance: entity.NewInstance(fact),
	}, nil
}

type storage struct {
	instance Instances
}

func (s storage) Instances() Instances {
	return s.instance
}
