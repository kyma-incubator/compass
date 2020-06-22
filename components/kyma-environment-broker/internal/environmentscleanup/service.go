package environmentscleanup

import (
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
	shootsToDelete, err := s.getShootsToDelete(s.LabelSelector)
	if err != nil {
		return errors.Wrap(err, "while getting shoots to delete")
	}

	var result *multierror.Error
	for _, shoot := range shootsToDelete {
		log.Infof("Triggering environment deprovisioning for shoot %q, runtimeID: %q", shoot.Name, shoot.Annotations[shootAnnotationRuntimeId])
		currentErr := s.triggerEnvironmentDeprovisioning(shoot)
		if currentErr != nil {
			result = multierror.Append(result, currentErr)
			log.Error(errors.Wrapf(currentErr, "while triggering deprovisioning for shoot %q", shoot.Name))
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

	var result []v1beta1.Shoot
	for _, shoot := range shootList.Items {
		log.Infof("Processing shoot %q created on %q", shoot.GetName(), shoot.GetCreationTimestamp())

		shootCreationTimestamp := shoot.GetCreationTimestamp()

		shootAge := time.Since(shootCreationTimestamp.Time)

		if shootAge.Hours() >= s.MaxShootAge.Hours() {
			log.Infof("Shoot %q is older than %f hours with age: %f hours", shoot.Name, s.MaxShootAge.Hours(), shootAge.Hours())
			result = append(result, shoot)
		}
	}

	return result, nil
}

func (s *Service) triggerEnvironmentDeprovisioning(shoot v1beta1.Shoot) error {
	instanceID, err := s.getInstanceId(shoot)
	if err != nil {
		return err
	}

	payload := broker.DeprovisionDetails{InstanceID: instanceID, CloudProfileName: shoot.Spec.CloudProfileName}

	opID, err := s.brokerService.Deprovision(payload)
	if err != nil {
		return err
	}

	log.Infof("Successfully send deprovision request, got operation ID %q", opID)
	return nil
}

func (s *Service) getInstanceId(shoot v1beta1.Shoot) (string, error) {
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
