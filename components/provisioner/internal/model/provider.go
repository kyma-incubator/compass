package model

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/hydroform/types"
)

type ProviderConfiguration interface {
	ToHydroformConfiguration(credentialsFileName string) (*types.Cluster, *types.Provider, error)
	ToShootTemplate(namespace string, subAccountId string) (*gardener_types.Shoot, error)
}
