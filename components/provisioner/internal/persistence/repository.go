package persistence

import (
	"database/sql"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"time"
)

type Repository interface {
	BeginTransaction() (*sql.Tx, error)
	InsertCluster(runtimeID string, time time.Time, tx *sql.Tx) error
	InsertClusterConfig(runtimeID string, config model.ClusterConfig, tx *sql.Tx) error
	InsertKymaConfig(runtimeID string, config model.KymaConfig, tx *sql.Tx) error
	InsertOperation(operation model.Operation, tx *sql.Tx) error
	DeleteCluster(runtimeID string) error
	GetRuntimeStatus(runtimeID string) (model.RuntimeStatus, error)
	GetLastOperation(runtimeID string) (model.Operation, error)
}

type repository struct {
	connection *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return repository{
		connection: db,
	}
}

func (r repository) InsertCluster(runtimeID string, time time.Time, tx *sql.Tx) error {
	return nil
}

func (r repository) InsertClusterConfig(runtimeID string, config model.ClusterConfig, tx *sql.Tx) error {
	return nil
}

func (r repository) InsertKymaConfig(runtimeID string, config model.KymaConfig, tx *sql.Tx) error {
	return nil
}

func (r repository) InsertOperation(operation model.Operation, tx *sql.Tx) error {
	return nil
}

func (r repository) DeleteCluster(runtimeID string) error {
	return nil
}

func (r repository) GetRuntimeStatus(runtimeID string) (model.RuntimeStatus, error) {
	return model.RuntimeStatus{}, nil
}

func (r repository) BeginTransaction() (*sql.Tx, error) {
	return r.connection.Begin()
}

func (r repository) GetLastOperation(runtimeID string) (model.Operation, error) {
	return model.Operation{}, nil
}
