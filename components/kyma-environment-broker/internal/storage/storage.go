package storage

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/service"
	"github.com/pkg/errors"
)

type BrokerStorage interface {
	Instances() Instances
}

func New(connectionURL string) (BrokerStorage, error) {
	connection, err := InitializeDatabase(connectionURL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize database")
	}
	fact := dbsession.NewFactory(connection)

	return storage{
		instance: service.NewInstance(fact),
	}, nil
}

type storage struct {
	instance Instances
}

func (s storage) Instances() Instances {
	return s.instance
}
