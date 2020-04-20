package deprovisioning

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/edp"
	kebError "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/error"

	"github.com/sirupsen/logrus"
)

//go:generate mockery -name=EDPClient -output=automock -outpkg=automock -case=underscore
type EDPClient interface {
	DeleteDataTenant(name, env string) error
	DeleteMetadataTenant(name, env, key string) error
}

type EDPDeregistration struct {
	client EDPClient
	config edp.Config
}

func NewEDPDeregistration(client EDPClient, config edp.Config) *EDPDeregistration {
	return &EDPDeregistration{
		client: client,
		config: config,
	}
}

func (s *EDPDeregistration) Name() string {
	return "EDP_Deregistration"
}

func (s *EDPDeregistration) Run(operation internal.DeprovisioningOperation, log logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error) {
	log.Info("Delete DataTenant metadata")
	for _, key := range []string{
		edp.MaasConsumerEnvironmentKey,
		edp.MaasConsumerRegionKey,
		edp.MaasConsumerSubAccountKey,
	} {
		err := s.client.DeleteMetadataTenant(operation.SubAccountID, s.config.Environment, key)
		if err != nil {
			return s.handleError(operation, err, log, fmt.Sprintf("cannot remove DataTenant metadata with key: %s", key))
		}
	}

	log.Info("Delete DataTenant")
	err := s.client.DeleteDataTenant(operation.SubAccountID, s.config.Environment)
	if err != nil {
		return s.handleError(operation, err, log, "cannot remove DataTenant")
	}

	return operation, 0, nil
}

func (s *EDPDeregistration) handleError(operation internal.DeprovisioningOperation, err error, log logrus.FieldLogger, msg string) (internal.DeprovisioningOperation, time.Duration, error) {
	log.Errorf("%s: %s", msg, err)

	if kebError.IsTemporaryError(err) {
		since := time.Since(operation.UpdatedAt)
		if since < time.Minute*30 {
			log.Errorf("request to EDP failed: %s. Retry...", err)
			return operation, 10 * time.Second, nil
		}
	}

	log.Errorf("Step %s failed. EDP data have not been deleted.", s.Name())
	return operation, 0, nil
}
