package installation

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"k8s.io/client-go/tools/clientcmd"

	pkgErrors "github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/artifacts"

	"github.com/kyma-incubator/hydroform/install/installation"
	"k8s.io/client-go/rest"
)

type InstallationHandler func(*rest.Config, ...installation.InstallationOption) (installation.Installer, error)

//go:generate mockery -name=ArtifactsProvider
type ArtifactsProvider interface {
	GetArtifacts(version string) (artifacts.ReleaseArtifacts, error)
}

//go:generate mockery -name=Service
type Service interface {
	InstallKyma(kubeconfigRaw, kymaVersion string) error
}

func NewInstallationService(installationTimeout time.Duration, artifactsProvider ArtifactsProvider, installationHandler InstallationHandler, installErrFailureThreshold int) Service {
	return &installationService{
		installationErrorsFailureThreshold: installErrFailureThreshold,
		kymaInstallationTimeout:            installationTimeout,
		artifactsProvider:                  artifactsProvider,
		installationHandler:                installationHandler,
	}
}

type installationService struct {
	installationErrorsFailureThreshold int
	kymaInstallationTimeout            time.Duration
	artifactsProvider                  ArtifactsProvider
	installationHandler                InstallationHandler
}

func (s *installationService) InstallKyma(kubeconfigRaw, kymaVersion string) error {
	kubeconfig, err := clientcmd.NewClientConfigFromBytes([]byte(kubeconfigRaw))
	if err != nil {
		return fmt.Errorf("error constructing kubeconfig from raw config: %s", err.Error())
	}

	clientConfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to get client kubeconfig from parsed config: %s", err.Error())
	}

	releaseArtifacts, err := s.artifactsProvider.GetArtifacts(kymaVersion)
	if err != nil {
		return pkgErrors.Wrap(err, "Failed to get release Artifacts")
	}

	kymaInstaller, err := s.installationHandler(clientConfig)
	if err != nil {
		return pkgErrors.Wrap(err, "Failed to create Kyma installer")
	}

	installationConfig := installation.Installation{
		TillerYaml:    releaseArtifacts.TillerYAML,
		InstallerYaml: releaseArtifacts.InstallerYAML,
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

	err = s.waitForInstallation(stateChannel, errChannel)
	if err != nil {
		return pkgErrors.Wrap(err, "Error while waiting for Kyma to install")
	}

	return nil
}

func (s *installationService) waitForInstallation(stateChannel <-chan installation.InstallationState, errorChannel <-chan error) error {
	for {
		select {
		case state, ok := <-stateChannel:
			if !ok {
				return nil
			}
			log.Printf("Description: %s, State: %s", state.Description, state.State)
		case err, ok := <-errorChannel:
			if !ok {
				continue
			}
			log.Printf("An error occurred: %v", err) // TODO = log with context or remove

			installationError := installation.InstallationError{}
			if ok := errors.As(err, &installationError); ok {
				log.Printf("Installation errors occured:")
				for _, e := range installationError.ErrorEntries {
					log.Printf("Component: %s", e.Component)
					log.Printf(e.Log)
				}

				if len(installationError.ErrorEntries) > s.installationErrorsFailureThreshold {
					return fmt.Errorf("installation errors exceeded occured: %s", installationError.Details())
				}
				continue
			}

			return err
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
