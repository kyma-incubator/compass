package storage

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/driver/memory"
	postgres "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/driver/postsql"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"
	"github.com/sirupsen/logrus"
)

type BrokerStorage interface {
	Instances() Instances
	Operations() Operations
	LMSTenants() LMSTenants
}

func NewFromConfig(cfg Config, log logrus.FieldLogger) (BrokerStorage, error) {
	connection, err := postsql.InitializeDatabase(cfg.ConnectionURL())
	if err != nil {
		return nil, err
	}

	log.Infof("Setting DB connection pool params: connectionMaxLifetime=%s "+
		"maxIdleConnections=%d maxOpenConnections=%d", cfg.ConnMaxLifetime, cfg.MaxIdleConns, cfg.MaxOpenConns)
	connection.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	connection.SetMaxIdleConns(cfg.MaxIdleConns)
	connection.SetMaxOpenConns(cfg.MaxOpenConns)

	fact := dbsession.NewFactory(connection)

	return storage{
		instance:   postgres.NewInstance(fact),
		operation:  postgres.NewOperation(fact),
		lmsTenants: postgres.NewLMSTenants(fact),
	}, nil
}

func NewMemoryStorage() BrokerStorage {
	return storage{
		instance:   memory.NewInstance(),
		operation:  memory.NewOperation(),
		lmsTenants: memory.NewLMSTenants(),
	}
}

type storage struct {
	instance   Instances
	operation  Operations
	lmsTenants LMSTenants
}

func (s storage) Instances() Instances {
	return s.instance
}

func (s storage) Operations() Operations {
	return s.operation
}

func (s storage) LMSTenants() LMSTenants {
	return s.lmsTenants
}
