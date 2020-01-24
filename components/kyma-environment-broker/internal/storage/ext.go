package storage

import "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"

type Instances interface {
	GetByID(instanceID string) (*internal.Instance, error)
	Insert(instance internal.Instance) error
	Update(instance internal.Instance) error
}
