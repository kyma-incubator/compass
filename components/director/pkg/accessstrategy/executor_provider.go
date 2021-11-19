package accessstrategy

type executorProvider struct {
	executors map[Type]Executor
}

// NewExecutorProvider returns a new access strategy executor provider based on type <-> executor mapping
func NewExecutorProvider(executors map[Type]Executor) *executorProvider {
	return &executorProvider{
		executors: executors,
	}
}

// NewDefaultExecutorProvider returns a new access strategy executor provider with the default static type <-> executor mapping
func NewDefaultExecutorProvider() *executorProvider {
	return &executorProvider{
		executors: supportedAccessStrategies,
	}
}

// Provide provides an executor for a given access strategy if supported, UnsupportedErr otherwise
func (ep *executorProvider) Provide(accessStrategyType Type) (Executor, error) {
	executor, ok := ep.executors[accessStrategyType]
	if !ok {
		return nil, UnsupportedErr
	}
	return executor, nil
}
