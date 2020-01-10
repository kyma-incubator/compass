package storage

import "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"

//go:generate mockery -name=Instances -output=automock -outpkg=automock -case=underscore
type Instances interface {
	GetByID(instanceID string) (*internal.Instance, error)
	Insert(instance internal.Instance) error
	Update(instance internal.Instance) error
}
