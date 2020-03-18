module github.com/kyma-incubator/compass/tests/connectivity-adapter

go 1.13

require (
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/kyma-incubator/compass/components/director v0.0.0-20200309151613-712f462350c8
	github.com/kyma-incubator/compass/tests/director v0.0.0-20200204090111-0997ef97abfb
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/vrischmann/envconfig v1.2.0
)

replace github.com/kyma-incubator/compass/components/director => github.com/kfurgol/compass/components/director v0.0.0-20200318061650-e199a9977f64
