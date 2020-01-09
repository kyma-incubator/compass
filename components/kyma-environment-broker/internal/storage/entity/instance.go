package entity

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession"
	"github.com/pkg/errors"
)

type Instance struct {
	dbsession.Factory
}

func NewInstance(sess dbsession.Factory) *Instance {
	return &Instance{
		Factory: sess,
	}
}

func (s *Instance) GetByID(instanceID string) (*internal.Instance, error) {
	sess := s.NewReadSession()
	instance, err := sess.GetInstanceByID(instanceID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting instance by ID %s", instanceID)
	}

	return &instance, nil
}

func (s *Instance) Insert(instance internal.Instance) error {
	sess := s.NewWriteSession()
	err := sess.InsertInstance(instance)
	if err != nil {
		return errors.Wrapf(err, "while saving instance ID %s", instance.InstanceID)
	}

	return nil
}

func (s *Instance) Update(instance internal.Instance) error {
	sess := s.NewWriteSession()
	err := sess.UpdateInstance(instance)
	if err != nil {
		return errors.Wrapf(err, "while updating instance ID %s", instance.InstanceID)
	}

	return nil
}
