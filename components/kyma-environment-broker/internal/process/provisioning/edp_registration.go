package provisioning

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/edp"
	kebError "github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/error"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/process"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage"

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
	subAccountID := parameters.ErsContext.SubAccountID

	log.Infof("Create DataTenant for %s subaccount", subAccountID)
	err = s.client.CreateDataTenant(edp.DataTenantPayload{
		Name:        subAccountID,
		Environment: s.config.Environment,
		Secret:      s.generateSecret(subAccountID, s.config.Environment),
	})
	if err != nil {
		return s.handleError(operation, err, log, "cannot create DataTenant")
	}

	log.Infof("Create DataTenant metadata for %s subaccount", subAccountID)
	for key, value := range map[string]string{
		edp.MaasConsumerEnvironmentKey: s.selectEnvironmentKey(parameters.PlatformRegion, log),
		edp.MaasConsumerRegionKey:      parameters.PlatformRegion,
		edp.MaasConsumerSubAccountKey:  subAccountID,
	} {
		err = s.client.CreateMetadataTenant(subAccountID, s.config.Environment, edp.MetadataTenantPayload{
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

func (s *EDPRegistrationStep) selectEnvironmentKey(region string, log logrus.FieldLogger) string {
	parts := strings.Split(region, "-")
	switch parts[0] {
	case "cf":
		return "CF"
	case "k8s":
		return "KUBERNETES"
	case "neo":
		return "NEO"
	default:
		log.Warnf("region %s does not fit any of the options, default CF is used", region)
		return "CF"
	}
}

// generateSecret generates secret during dataTenant creation, at this moment the secret is not needed
// except required parameter
func (s *EDPRegistrationStep) generateSecret(name, env string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s", name, env)))
}
