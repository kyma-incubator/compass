package provisioning

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
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
		opStorage:       os,
		accountProvider: accountProvider,
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
//
//		//type ERSContext struct {
//		//	TenantID        string                 `json:"tenant_id"`
//		//	SubAccountID    string                 `json:"subaccount_id"`
//		//	GlobalAccountID string                 `json:"globalaccount_id"`
//		//	ServiceManager  ServiceManagerEntryDTO `json:"sm_platform_credentials"`
//		//}
//		// update
//		//Błędy:
//		//1. jezeli chcemy ponowic step to zwracamy: operation, time.Minute, nil
//		//2. Jezeli chcemy zakonczyc operacje z bledem:
//		//    operation.State = Failed
//		//    operation = s.storage.Update(operation)
//		//    return operation, 0, error.New("")
//		//3. jezeli jest ok
//		//    operation = s.storage.Update(operation)
//		//	return operation, 0, nil
//
//
//
//	return operation, 0, nil
//}

/* Code from lms step

package lms
import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"time"
	"github.com/sirupsen/logrus"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"k8s.io/apimachinery/pkg/util/wait"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
)
type TenantProvider interface {
	ProvideLMSTenantID(name, region string) (string, error)
}
// provideTenantStep creates (if not exists) LMS tenant and provides its ID.
// The step does not breaks the provisioning flow.
type provideTenantStep struct {
	tenantProvider TenantProvider
	repo           storage.Operations
}
func NewProvideTenantStep(tp TenantProvider, repo storage.Operations) *provideTenantStep {
	return &provideTenantStep{
		tenantProvider: tp,
		repo:           repo,
	}
}
func (s *provideTenantStep) Name() string {
	return "Create LMS tenant"
}
func (s *provideTenantStep) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	if operation.Lms.TenantID != "" {
		return operation, 0, nil
	}
	region := "eu"                 //todo: take region from provisioning parameters (PP)
	tenant := "todo-global-acc-id" // todo: extract from PP
	lmsTenantID, err := s.tenantProvider.ProvideLMSTenantID(tenant, region)
	operation.Lms.TenantID = lmsTenantID
	if operation.Lms.RequestedAt.IsZero() {
		operation.Lms.RequestedAt = time.Now()
	}
	op, updateErr := s.repo.UpdateProvisioningOperation(operation)
	if updateErr != nil {
		return operation, time.Second, nil
	}
	operation = *op
	if err != nil {
		logger.Errorf("Unable to get LMS tenant ID: %s", err.Error())
		return operation, 5 * time.Second, nil
	}
	return operation, 0, nil
}


*/
