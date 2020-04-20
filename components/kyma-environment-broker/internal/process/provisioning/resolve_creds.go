package provisioning

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ResolveCredentialsStep struct {
	operationManager *process.ProvisionOperationManager
	accountProvider  hyperscaler.AccountProvider
	opStorage        storage.Operations
	tenant           string
}

func getHyperscalerTypeForPlanID(planID string) (hyperscaler.HyperscalerType, error) {
	switch planID {
	case broker.GCPPlanID:
		return hyperscaler.GCP, nil
	case broker.AzurePlanID:
		return hyperscaler.Azure, nil
	default:
		return "", errors.Errorf("Cannot determine the type of Hyperscaler to use for planID: %s", planID)
	}
}

func NewResolveCredentialsStep(os storage.Operations, accountProvider hyperscaler.AccountProvider) *ResolveCredentialsStep {

	return &ResolveCredentialsStep{
		operationManager: process.NewProvisionOperationManager(os),
		opStorage:        os,
		accountProvider:  accountProvider,
	}
}

func (s *ResolveCredentialsStep) Name() string {
	return "Resolve_Target_Secret"
}

func (s *ResolveCredentialsStep) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {

	pp, err := operation.GetProvisioningParameters()

	if err != nil {
		logger.Error("Aborting after failing to get valid operation provisioning parameters")
		return s.operationManager.OperationFailed(operation, "invalid operation provisioning parameters")
	}

	if pp.Parameters.TargetSecret != nil {
		return operation, 0, nil
	}

	hypType, err := getHyperscalerTypeForPlanID(pp.PlanID)

	if err != nil {
		logger.Error("Aborting after failing to determine the type of Hyperscaler to use for planID: %s", pp.PlanID)
		return s.operationManager.OperationFailed(operation, err.Error())
	}

	logger.Infof("HAP lookup for credentials to provision cluster for global account ID %s on Hyperscaler %s", pp.ErsContext.GlobalAccountID, hypType)

	credentials, err := s.accountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)

	if err != nil {
		errMsg := fmt.Sprintf("HAP lookup for credentials to provision cluster for global account ID %s on Hyperscaler %s has failed: %s", pp.ErsContext.GlobalAccountID, hypType, err)
		logger.Info(errMsg)

		// if failed retry step every 10s by next 10min
		dur := time.Since(operation.UpdatedAt).Round(time.Minute)

		if dur < 10*time.Minute {
			return operation, 10 * time.Second, nil
		}

		logger.Errorf("Aborting after 10 minutes of failing to resolve provisioning credentials for global account ID %s on Hyperscaler %s", pp.ErsContext.GlobalAccountID, hypType)
		return s.operationManager.OperationFailed(operation, errMsg)
	}

	pp.Parameters.TargetSecret = &credentials.CredentialName
	err = operation.SetProvisioningParameters(pp)

	if err != nil {
		logger.Error("Aborting after failing to save provisioning parameters for operation")
		return s.operationManager.OperationFailed(operation, err.Error())
	}

	updatedOperation, err := s.opStorage.UpdateProvisioningOperation(operation)

	if err != nil {
		return operation, 1 * time.Minute, nil
	}

	logger.Infof("Resolved %s as target secret name to use for cluster provisioning for global account ID %s on Hyperscaler %s", *pp.Parameters.TargetSecret, pp.ErsContext.GlobalAccountID, hypType)

	return *updatedOperation, 0, nil
}
