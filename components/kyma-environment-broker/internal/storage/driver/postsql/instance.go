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
	instance := internal.Instance{}
	var lastErr dberr.Error
	err := wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		instance, lastErr = sess.GetInstanceByID(instanceID)
		if lastErr != nil {
			if dberr.IsNotFound(lastErr) {
				return false, dberr.NotFound("Instance with id %s not exist", instanceID)
			}
			log.Warn(errors.Wrapf(lastErr, "while getting instance by ID %s", instanceID).Error())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, lastErr
	}
	return &instance, nil
}

func (s *Instance) Insert(instance internal.Instance) error {
	_, err := s.GetByID(instance.InstanceID)
	if err == nil {
		return dberr.AlreadyExists("instance with id %s already exist", instance.InstanceID)
	}

	sess := s.NewWriteSession()
	return wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
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
	var lastErr dberr.Error
	err := wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		lastErr = sess.UpdateInstance(instance)
		if lastErr != nil {
			if dberr.IsNotFound(lastErr) {
				return false, dberr.NotFound("Instance with id %s not exist", instance.InstanceID)
			}
			log.Warn(errors.Wrapf(lastErr, "while updating instance ID %s", instance.InstanceID).Error())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return lastErr
	}
	return nil
}

func (s *Instance) Delete(instanceID string) error {
	sess := s.NewWriteSession()
	return sess.DeleteInstance(instanceID)
}
