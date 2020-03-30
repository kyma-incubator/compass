package provisioning

import (
	"time"

	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
)

type LmsTenantProvider interface {
	ProvideLMSTenantID(name, region string) (string, error)
}

// provideLmsTenantStep creates (if not exists) LMS tenant and provides its ID.
// The step does not breaks the provisioning flow.
type provideLmsTenantStep struct {
	tenantProvider   LmsTenantProvider
	operationManager *OperationManager
}

func NewProvideLmsTenantStep(tp LmsTenantProvider, repo storage.Operations) *provideLmsTenantStep {
	return &provideLmsTenantStep{
		tenantProvider:   tp,
		operationManager: NewOperationManager(repo),
	}
}

func (s *provideLmsTenantStep) Name() string {
	return "Create_LMS_Tenant"
}

func (s *provideLmsTenantStep) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	if operation.Lms.TenantID != "" {
		return operation, 0, nil
	}

	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		msg := fmt.Sprintf("Unable to get provisioning parameters: %s", err.Error())
		logger.Errorf(msg)
		return s.operationManager.OperationFailed(operation, msg)
	}
	region := s.provideRegion(pp.Parameters.Region)

	name := pp.ErsContext.GlobalAccountID
	if len(name) > 8 {
		name = name[:8]
	}
	lmsTenantID, err := s.tenantProvider.ProvideLMSTenantID(name, region)
	if err != nil {
		errorMessage := fmt.Sprintf("Unable to request for LMS tenant ID: %s", err.Error())
		return s.operationManager.RetryOperation(operation, errorMessage, 30*time.Second, 2*time.Minute, logger)
	}

	operation.Lms.TenantID = lmsTenantID
	if operation.Lms.RequestedAt.IsZero() {
		operation.Lms.RequestedAt = time.Now()
	}

	op, repeat := s.operationManager.UpdateOperation(operation)
	if repeat != 0 {
		logger.Errorf("cannot save LMS tenant ID")
		return operation, time.Second, nil
	}

	return op, 0, nil
}

var lmsRegionsMap = map[string]string{
	"westeurope":    "eu",
	"eastus":        "us",
	"eastus2":       "us",
	"centralus":     "us",
	"northeurope":   "eu",
	"southeastasia": "aus",
	"japaneast":     "aus",
	"westus2":       "eu",
	"uksouth":       "eu",
	"FranceCentral": "eu",
	"EastUS2EUAP":   "us",
	"uaenorth":      "eu",
}

func (s *provideLmsTenantStep) provideRegion(r *string) string {
	if r == nil {
		return "eu"
	}
	region, found := lmsRegionsMap[*r]
	if !found {
		return "eu"
	}
	return region
}
