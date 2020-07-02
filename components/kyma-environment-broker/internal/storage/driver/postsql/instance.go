package postsql

import (
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage/dbsession"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage/predicate"
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

func (s *Instance) FindAllJoinedWithOperations(prct ...predicate.Predicate) ([]internal.InstanceWithOperation, error) {
	sess := s.NewReadSession()
	var (
		instances []internal.InstanceWithOperation
		lastErr   dberr.Error
	)
	err := wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		instances, lastErr = sess.FindAllInstancesJoinedWithOperation(prct...)
		if lastErr != nil {
			log.Warn(errors.Wrapf(lastErr, "while fetching all instances").Error())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, lastErr
	}

	return instances, nil
}

func (s *Instance) FindAllInstancesForRuntimes(runtimeIdList []string) ([]internal.Instance, error) {
	sess := s.NewReadSession()
	var instances []internal.Instance
	var lastErr dberr.Error
	err := wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		instances, lastErr = sess.FindAllInstancesForRuntimes(runtimeIdList)
		if lastErr != nil {
			if dberr.IsNotFound(lastErr) {
				return false, dberr.NotFound("Instances with runtime IDs from list '%+q' not exist", runtimeIdList)
			}
			log.Warn(errors.Wrapf(lastErr, "while getting instances from runtime ID list '%+q'", runtimeIdList).Error())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, lastErr
	}
	return instances, nil
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

func (s *Instance) GetInstanceStats() (internal.InstanceStats, error) {
	entries, err := s.NewReadSession().GetInstanceStats()
	if err != nil {
		return internal.InstanceStats{}, err
	}

	result := internal.InstanceStats{
		PerGlobalAccountID: make(map[string]int),
	}
	for _, e := range entries {
		result.PerGlobalAccountID[e.GlobalAccountID] = e.Total
		result.TotalNumberOfInstances = result.TotalNumberOfInstances + e.Total
	}
	return result, nil
}
