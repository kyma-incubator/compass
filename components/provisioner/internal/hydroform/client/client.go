package client

import (
	"github.com/kyma-incubator/hydroform"
	"github.com/kyma-incubator/hydroform/types"
)

//go:generate mockery -name=Client
type Client interface {
	Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error)
	Status(cluster *types.Cluster, provider *types.Provider) (*types.ClusterStatus, error)
	Credentials(cluster *types.Cluster, provider *types.Provider) ([]byte, error)
	Deprovision(cluster *types.Cluster, provider *types.Provider) error
}

type client struct{}

func NewHydroformClient() Client {
	return &client{}
}

func (c client) Provision(cluster *types.Cluster, provider *types.Provider) (*types.Cluster, error) {
	return hydroform.Provision(cluster, provider)
}
func (c client) Status(cluster *types.Cluster, provider *types.Provider) (*types.ClusterStatus, error) {
	return hydroform.Status(cluster, provider)
}
func (c client) Credentials(cluster *types.Cluster, provider *types.Provider) ([]byte, error) {
	return hydroform.Credentials(cluster, provider)
}
func (c client) Deprovision(cluster *types.Cluster, provider *types.Provider) error {
	return hydroform.Deprovision(cluster, provider)
}
