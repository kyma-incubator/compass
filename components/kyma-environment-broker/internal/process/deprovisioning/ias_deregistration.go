package deprovisioning

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ias"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/sirupsen/logrus"
)

type IASDeregistration struct {
	operationManager *process.DeprovisionOperationManager
	bundleBuilder    ias.BundleBuilder
}

func NewIASDeregistration(os storage.Operations, bundleBuilder ias.BundleBuilder) *IASDeregistration {
	return &IASDeregistration{
		operationManager: process.NewDeprovisionOperationManager(os),
		bundleBuilder:    bundleBuilder,
	}
}

func (s *IASDeregistration) Name() string {
	return "IAS_Deregistration"
}

func (s *IASDeregistration) Run(operation internal.DeprovisioningOperation, log logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error) {
	spb := s.bundleBuilder.NewBundle(operation.InstanceID)

	log.Info("Removing ServiceProvider from IAS")
	err := spb.DeleteServiceProvider()
	if err != nil {
		msg := fmt.Sprintf("cannot delete ServiceProvider %s", spb.ServiceProviderName())
		log.Errorf(msg)
		return s.operationManager.RetryOperation(operation, msg, 5*time.Second, 10*time.Minute, log)
	}

	return operation, 0, nil
}
