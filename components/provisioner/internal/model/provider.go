package model

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/kyma-incubator/hydroform/types"
)

type ProviderConfiguration interface {
	ToHydroformConfiguration(credentialsFileName string) (*types.Cluster, *types.Provider, error)
	ToShootTemplate(namespace string) *gardener_types.Shoot
}
