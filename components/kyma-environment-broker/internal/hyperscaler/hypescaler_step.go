package hyperscaler

//import (
//	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
//	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
//	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
//	"github.com/pkg/errors"
//	"github.com/sirupsen/logrus"
//	"time"
//)
//
//type ResolveTargetSecretStep struct {
//	operationManager *process.OperationManager
//	opStorage        storage.Operations
//	accountProvider  AccountProvider
//}
//
//func getHyperscalerTypeForPlanID(planID string) (HyperscalerType, error) {
//	switch planID {
//	case broker.gcpPlanID:
//		return GCP, nil
//	case broker.azurePlanID:
//		return Azure, nil
//	case broker.awsPlanID
//		return AWS, nil
//	default:
//		return "", errors.Errorf("Cannot determine the type of Hyperscaler to use for planID: %s", planID)
//	}
//}
//
//func NewResolveTargetSecretStep(opStorage storage.Operations, accountProvider *AccountProvider, globalAccountId string) *ResolveTargetSecretStep {
//
//	return &ResolveTargetSecretStep{
//		operationManager process.NewOperationManager(os),
//		opStorage:       opStorage,
//		accountProvider: accountProvider,
//		globalAccountId: globalAccountId,
//	}
//}
//
//func (s *ResolveTargetSecretStep) Name() string {
//	return "Resolve_Target_Secret"
//}
//
//func (s *ResolveTargetSecretStep) Run(operation internal.ProvisioningOperation, logger *logrus.Entry) (internal.ProvisioningOperation, time.Duration, error) {
//
//	if operation.TargetSecret == "" {
//
//		pp, err := operation.GetProvisioningParameters()
//
//		if err != nil {
//			return s.operationManager.OperationFailed(operation, "invalid operation provisioning parameters")
//		}
//
//		hypType, err := getHyperscalerTypeForPlanID(pp.PlanID)
//
//		var accountID string = pp.ErsContext.GlobalAccountID
//
//		logger.Infof(" HAP lookup for target secret name to provision cluster for global account ID %s using PlanId %s", accountID, pp.PlanID)
//
//		credentials, err := s.accountProvider.GardnerCredentials(hypType, accountID)
//
//		// TODO: think about retry of it ???
//		if err != nil {
//			operation.State = domain.Failed
//			return operation, 0, err
//		}
//
//		operation.TargetSecret = credentials.CredentialName
//
//		updated, err := s.opStorage.UpdateProvisioningOperation(operation)
//		if err != nil {
//			operation, 0, errors.Errorf("Cannot update operation", err)
//		}
//
//		operation.InputCreator.SetGardenerSecretName(operation.TargetSecret)
//
//		return operation, 0, nil
//	}
//
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
