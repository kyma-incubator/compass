package provisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	kebError "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/error"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ias"

	"github.com/sirupsen/logrus"
)

const (
	setIASTypeTimeout = 10 * time.Minute
)

type IASType struct {
	bundleBuilder ias.BundleBuilder
	disabled      bool
}

func NewIASType(builder ias.BundleBuilder, disabled bool) *IASType {
	return &IASType{
		bundleBuilder: builder,
		disabled:      disabled,
	}
}

func (s *IASType) Disabled() bool {
	return s.disabled
}

func (s *IASType) ConfigureType(operation internal.ProvisioningOperation, runtimeURL string, log logrus.FieldLogger) (time.Duration, error) {
	if s.disabled {
		return 0, nil
	}

	for spID := range ias.ServiceProviderInputs {
		spb, err := s.bundleBuilder.NewBundle(operation.InstanceID, spID)
		if err != nil {
			return s.handleError(operation, err, log, "failed to create ServiceProvider Bundle")
		}
		err = spb.FetchServiceProviderData()
		if err != nil {
			return s.handleError(operation, err, log, "fetching ServiceProvider data failed")
		}

		log.Infof("Configure SSO Type for ServiceProvider %q with RuntimeURL: %s", spb.ServiceProviderName(), runtimeURL)
		err = spb.ConfigureServiceProviderType(runtimeURL)
		if err != nil {
			return s.handleError(operation, err, log, "setting SSO Type failed")
		}
	}

	return 0, nil
}

func (s *IASType) handleError(operation internal.ProvisioningOperation, err error, log logrus.FieldLogger, msg string) (time.Duration, error) {
	log.Errorf("%s: %s", msg, err)
	switch {
	case kebError.IsTemporaryError(err):
		if time.Since(operation.UpdatedAt) > setIASTypeTimeout {
			log.Errorf("setting IAS type has reached timeout: %s", err)
			// operation will be marked as a success, RuntimeURL will not be set in IAS ServiceProvider application
			return 0, nil
		}
		log.Errorf("setting IAS type cannot be realized", err)
		return 10 * time.Second, nil
	default:
		log.Errorf("setting IAS type failed: %s", err)
		// operation will be marked as a success, RuntimeURL will not be set in IAS ServiceProvider application
		return 0, nil
	}
}
