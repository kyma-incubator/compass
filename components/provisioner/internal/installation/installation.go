package installation

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util/k8s"
	"time"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"

	pkgErrors "github.com/pkg/errors"

	"github.com/kyma-incubator/hydroform/install/installation"
	"k8s.io/client-go/rest"
)

const (
	tillerWaitTime = 5 * time.Minute
)

type InstallationHandler func(*rest.Config, ...installation.InstallationOption) (installation.Installer, error)

//go:generate mockery -name=Service
type Service interface {
	InstallKyma(runtimeId, kubeconfigRaw string, release model.Release, globalConfig model.Configuration, componentsConfig []model.KymaComponentConfig) error
	CheckInstallationState(kubeconfig *rest.Config) (installation.InstallationState, error)
	// TODO: this will block for quite a while, consider running it in gorutine or split it to more steps (install tillert -> check periodicaly -> deploy installer -> trigger installation)
	TriggerInstallation(kubeconfigRaw *rest.Config, release model.Release, globalConfig model.Configuration, componentsConfig []model.KymaComponentConfig) error
	TriggerUpgrade(kubeconfigRaw *rest.Config, release model.Release, globalConfig model.Configuration, componentsConfig []model.KymaComponentConfig) error
	TriggerUninstall(kubeconfig *rest.Config) error
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
	installationHandler                InstallationHandler
}

func (s *installationService) TriggerInstallation(kubeconfig *rest.Config, release model.Release, globalConfig model.Configuration, componentsConfig []model.KymaComponentConfig) error {
	kymaInstaller, err := s.createKymaInstaller(kubeconfig, componentsConfig)
	if err != nil {
		return fmt.Errorf("failed to trigger installation: %s", err.Error())
	}

	installationConfig := installation.Installation{
		TillerYaml:    release.TillerYAML,
		InstallerYaml: release.InstallerYAML,
		Configuration: NewInstallationConfiguration(globalConfig, componentsConfig),
	}

	err = kymaInstaller.PrepareInstallation(installationConfig)
	if err != nil {
		return pkgErrors.Wrap(err, "Failed to prepare installation")
	}

	installationCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// We are not waiting for events, just triggering installation
	_, _, err = kymaInstaller.StartInstallation(installationCtx)
	if err != nil {
		return pkgErrors.Wrap(err, "Failed to start Kyma installation")
	}

	return nil
}

// TODO: consider refactoring this function
func (s *installationService) TriggerUpgrade(kubeconfig *rest.Config, release model.Release, globalConfig model.Configuration, componentsConfig []model.KymaComponentConfig) error {
	kymaInstaller, err := s.createKymaInstaller(kubeconfig, componentsConfig)
	if err != nil {
		return fmt.Errorf("failed to trigger installation: %s", err.Error())
	}

	installationConfig := installation.Installation{
		TillerYaml:    release.TillerYAML,
		InstallerYaml: release.InstallerYAML,
		Configuration: NewInstallationConfiguration(globalConfig, componentsConfig),
	}

	err = kymaInstaller.PrepareUpgrade(installationConfig)
	if err != nil {
		return pkgErrors.Wrap(err, "Failed to prepare upgrade")
	}

	installationCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// We are not waiting for events, just triggering installation
	_, _, err = kymaInstaller.StartInstallation(installationCtx)
	if err != nil {
		return pkgErrors.Wrap(err, "Failed to start Kyma installation")
	}

	return nil
}

func (s *installationService) createKymaInstaller(kubeconfig *rest.Config, componentsConfig []model.KymaComponentConfig) (installation.Installer, error) {
	kymaInstaller, err := s.installationHandler(
		kubeconfig,
		installation.WithTillerWaitTime(tillerWaitTime),
		installation.WithInstallationCRModification(GetInstallationCRModificationFunc(componentsConfig)),
	)
	if err != nil {
		return nil, pkgErrors.Wrap(err, "Failed to create Kyma installer")
	}

	return kymaInstaller, nil
}

func (s *installationService) InstallKyma(runtimeId, kubeconfigRaw string, release model.Release, globalConfig model.Configuration, componentsConfig []model.KymaComponentConfig) error {
	kubeconfig, err := k8s.ParseToK8sConfig([]byte(kubeconfigRaw))
	if err != nil {
		return fmt.Errorf("error parsing kubeconfig from raw config: %s", err.Error())
	}

	kymaInstaller, err := s.createKymaInstaller(kubeconfig, componentsConfig)
	if err != nil {
		return fmt.Errorf("failed to deploy Kyma installer: %s", err.Error())
	}

	installationConfig := installation.Installation{
		TillerYaml:    release.TillerYAML,
		InstallerYaml: release.InstallerYAML,
		Configuration: NewInstallationConfiguration(globalConfig, componentsConfig),
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

func (s *installationService) CheckInstallationState(kubeconfig *rest.Config) (installation.InstallationState, error) {
	return installation.CheckInstallationState(kubeconfig)
}

func (s *installationService) TriggerUninstall(kubeconfig *rest.Config) error {
	return installation.TriggerUninstall(kubeconfig)
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
				logrus.Warnf("Warning: installation error occurred while installing Kyma for %s Runtime: %s. Details: %s", runtimeId, installationError.Error(), installationError.Details())

				if len(installationError.ErrorEntries) > s.installationErrorsFailureThreshold {
					return fmt.Errorf("installation errors exceeded threshold, errors details: %s", installationError.Details())
				}
				continue
			}

			return fmt.Errorf("an error occurred while installing Kyma for %s Runtime: %s.", runtimeId, err.Error())
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func GetInstallationCRModificationFunc(componentsConfig []model.KymaComponentConfig) func(*v1alpha1.Installation) {
	return func(installation *v1alpha1.Installation) {
		components := make([]v1alpha1.KymaComponent, 0, len(componentsConfig))

		for _, cc := range componentsConfig {
			components = append(components, v1alpha1.KymaComponent{
				Name:      string(cc.Component),
				Namespace: cc.Namespace,
				Source:    toKymaComponentSource(cc.SourceURL),
			})
		}

		installation.Spec.Components = components
	}
}

func toKymaComponentSource(sourceURL string) *v1alpha1.ComponentSource {
	if sourceURL == "" {
		return nil
	}
	return &v1alpha1.ComponentSource{URL: sourceURL}
}

func NewInstallationConfiguration(globalConfg model.Configuration, componentsConfig []model.KymaComponentConfig) installation.Configuration {
	installationConfig := installation.Configuration{
		Configuration:          make([]installation.ConfigEntry, 0, len(globalConfg.ConfigEntries)),
		ComponentConfiguration: make([]installation.ComponentConfiguration, 0, len(componentsConfig)),
	}

	installationConfig.Configuration = toInstallationConfigEntries(globalConfg.ConfigEntries)

	for _, componentCfg := range componentsConfig {
		installationComponentConfig := installation.ComponentConfiguration{
			Component:     string(componentCfg.Component),
			Configuration: toInstallationConfigEntries(componentCfg.Configuration.ConfigEntries),
		}

		installationConfig.ComponentConfiguration = append(installationConfig.ComponentConfiguration, installationComponentConfig)
	}

	return installationConfig
}

func toInstallationConfigEntries(entries []model.ConfigEntry) []installation.ConfigEntry {
	installationCfgEntries := make([]installation.ConfigEntry, 0, len(entries))

	for _, e := range entries {
		installationCfgEntries = append(installationCfgEntries, toInstallationConfigEntry(e))
	}

	return installationCfgEntries
}

func toInstallationConfigEntry(entry model.ConfigEntry) installation.ConfigEntry {
	return installation.ConfigEntry{
		Key:    entry.Key,
		Value:  entry.Value,
		Secret: entry.Secret,
	}
}
