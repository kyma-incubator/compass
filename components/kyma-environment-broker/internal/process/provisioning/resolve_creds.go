package provisioning

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ResolveCredentialsStep struct {
	operationManager *process.OperationManager
	accountProvider  hyperscaler.AccountProvider
	opStorage        storage.Operations
	tenant           string
}

func getHyperscalerTypeForPlanID(planID string) (hyperscaler.HyperscalerType, error) {
	switch planID {
	case broker.GcpPlanID:
		return hyperscaler.GCP, nil
	case broker.AzurePlanID:
		return hyperscaler.Azure, nil
	case broker.AwsPlanID:
		return hyperscaler.AWS, nil
	default:
		return "", errors.Errorf("Cannot determine the type of Hyperscaler to use for planID: %s", planID)
	}
}

func NewResolveCredentialsStep(os storage.Operations, accountProvider hyperscaler.AccountProvider) *ResolveCredentialsStep {

	return &ResolveCredentialsStep{
		operationManager: process.NewOperationManager(os),
		opStorage:        os,
		accountProvider:  accountProvider,
	}
}

func (s *ResolveCredentialsStep) Name() string {
	return "Resolve_Target_Secret"
}

func (s *ResolveCredentialsStep) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {

	if operation.TargetSecret != "" {
		return operation, 0, nil
	}

	pp, err := operation.GetProvisioningParameters()

	if err != nil {
		return s.operationManager.OperationFailed(operation, "invalid operation provisioning parameters")
	}

	hypType, err := getHyperscalerTypeForPlanID(pp.PlanID)

	if err != nil {
		return s.operationManager.OperationFailed(operation, err.Error())
	}

	logger.Infof("HAP lookup for credentials to provision cluster for global account ID %s on hyperscaler %s", pp.ErsContext.GlobalAccountID, hypType)

	credentials, err := s.accountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)

	if err != nil {
		errMsg := fmt.Sprintf("HAP lookup for credentials to provision cluster for global account ID %s for hyperscaler %s has failed: %s", pp.ErsContext.GlobalAccountID, hypType, err)
		return s.operationManager.OperationFailed(operation, errMsg)
	}

	operation.TargetSecret = credentials.CredentialName

	return s.operationManager.OperationSucceeded(operation, "Resolved provisioning credentials secret name with HAP")
}
