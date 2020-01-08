package storage

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/entity"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/session"
	"github.com/pkg/errors"
)

type BrokerStorage interface {
	Instances() Instances
}

func New(cfg Config) (BrokerStorage, error) {
	connection, err := InitializeDatabase(cfg.ConnectionURL())
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize database")
	}
	fact := session.NewFactory(connection)

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
