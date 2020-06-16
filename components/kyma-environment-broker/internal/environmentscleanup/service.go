package environmentscleanup

import (
	"time"

	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/environmentscleanup/broker"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	labelSelector             = "owner.do-not-delete!=true"
	shootAnnotationRuntimeId  = "compass.provisioner.kyma-project.io/runtime-id"
	shootLabelGlobalAccountId = "account"
)

type GardenerClient interface {
	List(opts v1.ListOptions) (*v1beta1.ShootList, error)
}

type DirectorClient interface {
	GetInstanceId(accountID, runtimeID string) (string, error)
}

type BrokerClient interface {
	Deprovision(details broker.DeprovisionDetails) (string, error)
}

type Service struct {
	gardenerService GardenerClient
	directorService DirectorClient
	brokerService   BrokerClient

	MaxShootAge time.Duration
}

func NewService(gardenerClient GardenerClient, directorClient DirectorClient, brokerClient BrokerClient, maxShootAge time.Duration) *Service {
	return &Service{
		gardenerService: gardenerClient,
		directorService: directorClient,
		brokerService:   brokerClient,
		MaxShootAge:     maxShootAge,
	}
}

func (s *Service) PerformCleanup() {
	shootsToDelete, err := s.getShootsToDelete(labelSelector)
	if err != nil {
		log.Error(errors.Wrap(err, "while getting shoots to delete"))
	}

	for _, shoot := range shootsToDelete {
		log.Infof("Triggering environment deprovisioning for shoot %q, runtimeID: %q", shoot.Name, shoot.Annotations[shootAnnotationRuntimeId])
		err = s.triggerEnvironmentDeprovisioning(shoot)
		if err != nil {
			log.Error(errors.Wrapf(err, "while triggering deprovisioning for shoot %q", shoot.Name))
		}
	}
	log.Info("Kyma Environments cleanup performed successfully")
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
	globalAccountId, ok := shoot.Labels[shootLabelGlobalAccountId]
	if !ok {
		return "", errors.New("shoot's globalAccountID should not be nil")
	}
	runtimeId, ok := shoot.Annotations[shootAnnotationRuntimeId]
	if !ok {
		return "", errors.New("shoot's runtimeID should not be nil")
	}

	result, err := s.directorService.GetInstanceId(globalAccountId, runtimeId)
	if err != nil {
		return "", errors.Wrapf(err, "while getting Instance ID from director for shoot %q", shoot.Name)
	}

	return result, nil
}
