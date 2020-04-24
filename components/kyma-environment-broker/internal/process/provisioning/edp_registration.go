package provisioning

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/edp"
	kebError "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/error"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/sirupsen/logrus"
)

//go:generate mockery -name=EDPClient -output=automock -outpkg=automock -case=underscore
type EDPClient interface {
	CreateDataTenant(data edp.DataTenantPayload) error
	CreateMetadataTenant(name, env string, data edp.MetadataTenantPayload) error
}

type EDPRegistrationStep struct {
	operationManager *process.ProvisionOperationManager
	client           EDPClient
	config           edp.Config
}

func NewEDPRegistrationStep(os storage.Operations, client EDPClient, config edp.Config) *EDPRegistrationStep {
	return &EDPRegistrationStep{
		operationManager: process.NewProvisionOperationManager(os),
		client:           client,
		config:           config,
	}
}

func (s *EDPRegistrationStep) Name() string {
	return "EDP_Registration"
}

func (s *EDPRegistrationStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	parameters, err := operation.GetProvisioningParameters()
	if err != nil {
		return s.handleError(operation, err, log, "invalid operation provisioning parameters")
	}

	log.Info("Create DataTenant")
	err = s.client.CreateDataTenant(edp.DataTenantPayload{
		Name:        parameters.ErsContext.SubAccountID,
		Environment: s.config.Environment,
		Secret:      s.generateSecret(parameters.ErsContext.SubAccountID, s.config.Environment),
	})
	if err != nil {
		return s.handleError(operation, err, log, "cannot create DataTenant")
	}

	log.Info("Create DataTenant metadata")
	for key, value := range map[string]string{
		edp.MaasConsumerEnvironmentKey: "KUBERNETES",
		edp.MaasConsumerRegionKey:      parameters.PlatformRegion,
		edp.MaasConsumerSubAccountKey:  parameters.ErsContext.SubAccountID,
	} {
		err = s.client.CreateMetadataTenant(parameters.ErsContext.SubAccountID, s.config.Environment, edp.MetadataTenantPayload{
			Key:   key,
			Value: value,
		})
		if err != nil {
			return s.handleError(operation, err, log, fmt.Sprintf("cannot create DataTenant metadata %s", key))
		}
	}

	return operation, 0, nil
}

func (s *EDPRegistrationStep) handleError(operation internal.ProvisioningOperation, err error, log logrus.FieldLogger, msg string) (internal.ProvisioningOperation, time.Duration, error) {
	log.Errorf("%s: %s", msg, err)

	if kebError.IsTemporaryError(err) {
		since := time.Since(operation.UpdatedAt)
		if since < time.Minute*30 {
			log.Errorf("request to EDP failed: %s. Retry...", err)
			return operation, 10 * time.Second, nil
		}
	}

	if !s.config.Required {
		log.Errorf("Step %s failed. Step is not required. Skip step.", s.Name())
		return operation, 0, nil
	}

	return s.operationManager.OperationFailed(operation, msg)
}

// generateSecret generates secret during dataTenant creation, at this moment the secret is not needed
// except required parameter
func (s *EDPRegistrationStep) generateSecret(name, env string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s", name, env)))
}
