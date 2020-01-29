package postsql

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Instance struct {
	dbsession.Factory
}

func NewInstance(sess dbsession.Factory) *Instance {
	return &Instance{
		Factory: sess,
	}
}

// TODO: Wrap retries in single method WithRetries
func (s *Instance) GetByID(instanceID string) (*internal.Instance, error) {
	sess := s.NewReadSession()
	instance := &internal.Instance{}
	err := wait.Poll(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		inst, err := sess.GetInstanceByID(instanceID)
		if err != nil {
			if err.Code() == dberr.CodeNotFound {
				return false, dberr.NotFound("Instance with id %s not exist", instanceID)
			}
			log.Warn(errors.Wrapf(err, "while getting instance by ID %s", instanceID).Error())
			return false, nil
		}
		instance = &inst
		return true, nil
	})
	return instance, err
}

func (s *Instance) Insert(instance internal.Instance) error {
	_, err := s.GetByID(instance.InstanceID)
	if err == nil {
		return dberr.AlreadyExists("instance with id %s already exist", instance.InstanceID)
	}

	sess := s.NewWriteSession()
	return wait.Poll(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		err := sess.InsertInstance(instance)
		if err != nil {
			log.Warn(errors.Wrapf(err, "while saving instance ID %s", instance.InstanceID).Error())
			return false, nil
		}
		return true, nil
	})
}

func (s *Instance) Update(instance internal.Instance) error {
	sess := s.NewWriteSession()
	return wait.Poll(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		err := sess.UpdateInstance(instance)
		if err != nil {
			if err.Code() == dberr.CodeNotFound {
				return false, dberr.NotFound("Instance with id %s not exist", instance.InstanceID)
			}
			log.Warn(errors.Wrapf(err, "while updating instance ID %s", instance.InstanceID).Error())
			return false, nil
		}
		return true, nil
	})
}
