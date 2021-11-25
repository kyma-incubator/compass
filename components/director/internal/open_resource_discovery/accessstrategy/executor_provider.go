package accessstrategy

import "github.com/kyma-incubator/compass/components/director/pkg/certloader"

// ExecutorProvider defines an interface for access strategy executor provider
//go:generate mockery --name=ExecutorProvider --output=automock --outpkg=automock --case=underscore
type ExecutorProvider interface {
	Provide(accessStrategyType Type) (Executor, error)
	// GetSupported(accessStrategies AccessStrategies) (Type, bool)
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
			CMPmTLSAccessStrategy: NewCMPmTLSAccessStrategyExecutor(certCache),
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

// GetSupported returns the first AccessStrategy in the slice that is supported by CMP
// func (p *Provider) GetSupported(accessStrategies AccessStrategies) (Type, bool){
// 	for _, as := range accessStrategies {
// 		if p.isSupported(as.Type) {
// 			return as.Type, true
// 		}
// 		if as.Type == CustomAccessStrategy && p.isSupported(as.CustomType) {
// 			return as.CustomType, true
// 		}
// 	}
// 	return "", false
// }
//
// func (p *Provider) isSupported(t Type) bool{
// 	_, ok := p.executors[t]
// 	return ok
// }
