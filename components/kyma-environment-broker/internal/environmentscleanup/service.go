package environmentscleanup

import (
	"fmt"
	"strings"
	"time"

	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/environmentscleanup/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	shootAnnotationRuntimeId = "compass.provisioner.kyma-project.io/runtime-id"
)

//go:generate mockery -name=GardenerClient -output=automock
type GardenerClient interface {
	List(opts v1.ListOptions) (*v1beta1.ShootList, error)
}

//go:generate mockery -name=BrokerClient -output=automock
type BrokerClient interface {
	Deprovision(details broker.DeprovisionDetails) (string, error)
}

type Service struct {
	gardenerService GardenerClient
	brokerService   BrokerClient
	instanceStorage storage.Instances
	MaxShootAge     time.Duration
	LabelSelector   string
}

func NewService(gardenerClient GardenerClient, brokerClient BrokerClient, instanceStorage storage.Instances, maxShootAge time.Duration, labelSelector string) *Service {
	return &Service{
		gardenerService: gardenerClient,
		brokerService:   brokerClient,
		instanceStorage: instanceStorage,
		MaxShootAge:     maxShootAge,
		LabelSelector:   labelSelector,
	}
}

func (s *Service) PerformCleanup() error {
	var result *multierror.Error
	shootsToDelete, err := s.getShootsToDelete(s.LabelSelector)
	if err != nil {
		log.Error(errors.Wrap(err, "while getting shoots to delete"))
		result = multierror.Append(result, err)
	}

	var instancesDetails []instanceDetailsDTO
	for _, shoot := range shootsToDelete {
		runtimeId, ok := shoot.Annotations[shootAnnotationRuntimeId]
		if !ok {
			err = errors.New(fmt.Sprintf("shoot %q has no runtime-id annotation", shoot.Name))
			log.Error(err)
			result = multierror.Append(result, err)
			continue
		}
		instancesDetails = append(instancesDetails, instanceDetailsDTO{
			RuntimeID:        runtimeId,
			CloudProfileName: shoot.Spec.CloudProfileName,
		})
	}
	log.Infof("Instances Details to process: %v", instancesDetails)

	instancesToDelete, err := s.getInstanceIds(instancesDetails)
	if err != nil {
		log.Error(errors.Wrap(err, "while getting instance IDs for runtimes"))
		result = multierror.Append(result, err)
	}

	for _, payload := range instancesToDelete {
		log.Infof("Triggering environment deprovisioning for instance ID %q", payload.InstanceID)
		currentErr := s.triggerEnvironmentDeprovisioning(payload)
		if currentErr != nil {
			log.Error(errors.Wrapf(currentErr, "while triggering deprovisioning for instance ID %q", payload.InstanceID))
			result = multierror.Append(result, currentErr)
		}
	}
	if result != nil {
		result.ErrorFormat = func(i []error) string {
			var s []string
			for _, v := range i {
				s = append(s, v.Error())
			}
			return strings.Join(s, ", ")
		}
	}
	return result.ErrorOrNil()
}

func (s *Service) getShootsToDelete(labelSelector string) ([]v1beta1.Shoot, error) {
	opts := v1.ListOptions{
		LabelSelector: labelSelector,
	}
	shootList, err := s.gardenerService.List(opts)
	if err != nil {
		return []v1beta1.Shoot{}, errors.Wrap(err, "while listing Gardener shoots")
	}

	var shoots []v1beta1.Shoot
	for _, shoot := range shootList.Items {
		shootCreationTimestamp := shoot.GetCreationTimestamp()
		shootAge := time.Since(shootCreationTimestamp.Time)

		if shootAge.Hours() >= s.MaxShootAge.Hours() {
			log.Infof("Shoot %q is older than %f hours with age: %f hours", shoot.Name, s.MaxShootAge.Hours(), shootAge.Hours())
			shoots = append(shoots, shoot)
		}
	}

	return shoots, nil
}

type instanceDetailsDTO struct {
	InstanceID       string
	RuntimeID        string
	CloudProfileName string
}

func (s *Service) getInstanceIds(instancesToDelete []instanceDetailsDTO) ([]instanceDetailsDTO, error) {
	var runtimeIdList []string

	for _, instanceDetails := range instancesToDelete {
		runtimeIdList = append(runtimeIdList, instanceDetails.RuntimeID)
	}

	instances, err := s.instanceStorage.FindAllInstancesForRuntimes(runtimeIdList)
	if err != nil {
		return []instanceDetailsDTO{}, err
	}

	var toDelete []instanceDetailsDTO
	for _, instance := range instances {
		for _, instanceToDeleteDetails := range instancesToDelete {
			if instance.RuntimeID == instanceToDeleteDetails.RuntimeID {
				instanceToDeleteDetails.InstanceID = instance.InstanceID
				toDelete = append(toDelete, instanceToDeleteDetails)
			}
		}
	}

	return toDelete, nil
}

func (s *Service) triggerEnvironmentDeprovisioning(instanceDetails instanceDetailsDTO) error {
	payload := broker.DeprovisionDetails{
		InstanceID:       instanceDetails.InstanceID,
		CloudProfileName: instanceDetails.CloudProfileName,
	}
	opID, err := s.brokerService.Deprovision(payload)
	if err != nil {
		return err
	}
	log.Infof("Successfully send deprovision request, got operation ID %q", opID)
	return nil
}

func (s *Service) getInstanceIDForShoots(shoot v1beta1.Shoot) (string, error) {
	runtimeId, ok := shoot.Annotations[shootAnnotationRuntimeId]
	if !ok {
		return "", errors.New("shoot's runtimeID should not be nil")
	}
	instance, err := s.instanceStorage.GetByRuntimeID(runtimeId)
	if err != nil {
		return "", err
	}

	return instance.InstanceID, nil
}
