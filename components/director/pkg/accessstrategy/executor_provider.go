package accessstrategy

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
)

// ExecutorProvider defines an interface for access strategy executor provider
//go:generate mockery --name=ExecutorProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ExecutorProvider interface {
	Provide(accessStrategyType Type) (Executor, error)
}

// Provider is responsible to provides an access strategy executors
type Provider struct {
	executors map[Type]Executor
}

// NewExecutorProvider returns a new access strategy executor provider based on type <-> executor mapping
func NewExecutorProvider(executors map[Type]Executor) *Provider {
	return &Provider{
		executors: executors,
	}
}

// NewDefaultExecutorProvider returns a new access strategy executor provider with the default static type <-> executor mapping
func NewDefaultExecutorProvider(certCache certloader.Cache) *Provider {
	return &Provider{
		executors: map[Type]Executor{
			OpenAccessStrategy:    &openAccessStrategyExecutor{},
			CMPmTLSAccessStrategy: NewCMPmTLSAccessStrategyExecutor(certCache, nil),
		},
	}
}

// NewExecutorProviderWithTenant returns a new access strategy executor provider by given tenant provider function
func NewExecutorProviderWithTenant(certCache certloader.Cache, tenantProviderFunc func(ctx context.Context) (string, error)) *Provider {
	return &Provider{
		executors: map[Type]Executor{
			OpenAccessStrategy:    &openAccessStrategyExecutor{},
			CMPmTLSAccessStrategy: NewCMPmTLSAccessStrategyExecutor(certCache, tenantProviderFunc),
		},
	}
}

// Provide provides an executor for a given access strategy if supported, UnsupportedErr otherwise
func (p *Provider) Provide(accessStrategyType Type) (Executor, error) {
	executor, ok := p.executors[accessStrategyType]
	if !ok {
		return nil, UnsupportedErr
	}
	return executor, nil
}
