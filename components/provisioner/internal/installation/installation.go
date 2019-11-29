package installation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/release"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"

	"k8s.io/client-go/tools/clientcmd"

	pkgErrors "github.com/pkg/errors"

	"github.com/kyma-incubator/hydroform/install/installation"
	"k8s.io/client-go/rest"
)

type InstallationHandler func(*rest.Config, ...installation.InstallationOption) (installation.Installer, error)

//go:generate mockery -name=Service
type Service interface {
	InstallKyma(runtimeId, kubeconfigRaw string, release model.Release) error
}

func NewInstallationService(installationTimeout time.Duration, installationHandler InstallationHandler, installErrFailureThreshold int) Service {
	return &installationService{
		installationErrorsFailureThreshold: installErrFailureThreshold,
		kymaInstallationTimeout:            installationTimeout,
		installationHandler:                installationHandler,
	}
}

type installationService struct {
	installationErrorsFailureThreshold int
	kymaInstallationTimeout            time.Duration
	releaseRepo                        release.ReadRepository
	installationHandler                InstallationHandler
}

func (s *installationService) InstallKyma(runtimeId, kubeconfigRaw string, release model.Release) error {
	kubeconfig, err := clientcmd.NewClientConfigFromBytes([]byte(kubeconfigRaw))
	if err != nil {
		return fmt.Errorf("error constructing kubeconfig from raw config: %s", err.Error())
	}

	clientConfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to get client kubeconfig from parsed config: %s", err.Error())
	}

	kymaInstaller, err := s.installationHandler(clientConfig)
	if err != nil {
		return pkgErrors.Wrap(err, "Failed to create Kyma installer")
	}

	installationConfig := installation.Installation{
		TillerYaml:    release.TillerYAML,
		InstallerYaml: release.InstallerYAML,
		Configuration: installation.Configuration{},
	}

	err = kymaInstaller.PrepareInstallation(installationConfig)
	if err != nil {
		return pkgErrors.Wrap(err, "Failed to prepare installation")
	}

	installationCtx, cancel := context.WithTimeout(context.Background(), s.kymaInstallationTimeout)
	defer cancel()

	stateChannel, errChannel, err := kymaInstaller.StartInstallation(installationCtx)
	if err != nil {
		return pkgErrors.Wrap(err, "Failed to start Kyma installation")
	}

	err = s.waitForInstallation(runtimeId, stateChannel, errChannel)
	if err != nil {
		return pkgErrors.Wrap(err, "Error while waiting for Kyma to install")
	}

	return nil
}

func (s *installationService) waitForInstallation(runtimeId string, stateChannel <-chan installation.InstallationState, errorChannel <-chan error) error {
	for {
		select {
		case state, ok := <-stateChannel:
			if !ok {
				return nil
			}
			logrus.Infof("Installing Kyma on Runtime %s. Description: %s, State: %s", runtimeId, state.Description, state.State)
		case err, ok := <-errorChannel:
			if !ok {
				continue
			}

			installationError := installation.InstallationError{}
			if ok := errors.As(err, &installationError); ok {
				logrus.Warnf("Warning: installation error occurred while installing kyma for %s Runtime: %s. Details: %s", runtimeId, installationError.Error(), installationError.Details())

				if len(installationError.ErrorEntries) > s.installationErrorsFailureThreshold {
					return fmt.Errorf("installation errors exceeded threshold, errors details: %s", installationError.Details())
				}
				continue
			}

			return fmt.Errorf("an error occurred while installing kyma for %s Runtime: %s.", runtimeId, err.Error())
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
