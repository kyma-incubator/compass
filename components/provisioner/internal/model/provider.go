package model

import "github.com/kyma-incubator/hydroform/types"

type ProviderConfiguration interface {
	ToHydroformConfiguration(credentialsFileName string) (*types.Cluster, *types.Provider)
}
