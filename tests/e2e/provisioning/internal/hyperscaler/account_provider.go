package hyperscaler

import (
	"fmt"
)

type AccountProvider interface {
	GardenerCredentials(hyperscalerType Type, tenantName string) (Credentials, error)
}

type accountProvider struct {
	compassPool  AccountPool
	gardenerPool AccountPool
}

// NewAccountProvider returns a new AccountProvider
func NewAccountProvider(compassPool AccountPool, gardenerPool AccountPool) AccountProvider {
	return &accountProvider{
		compassPool:  compassPool,
		gardenerPool: gardenerPool,
	}
}

// GardenerCredentials returns credentials for Gardener account
func (p *accountProvider) GardenerCredentials(hyperscalerType Type, tenantName string) (Credentials, error) {
	if p.gardenerPool == nil {
		return Credentials{}, fmt.Errorf("failed to get Gardener Credentials. Gardener Account pool is not configured for tenant: %s", tenantName)
	}

	return p.gardenerPool.Credentials(hyperscalerType, tenantName)
}
