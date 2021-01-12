module github.com/kyma-incubator/compass/tests/system-broker

go 1.14

require (
	github.com/kyma-incubator/compass/components/connector v0.0.0-20201219152541-d77ebc00ac2d
	github.com/kyma-incubator/compass/components/director v0.0.0-20210107130916-b97cabae65e4
	github.com/kyma-incubator/compass/tests/connector-tests v0.0.0-20210111103459-2ce2dc25a7de
	github.com/kyma-incubator/compass/tests/director v0.0.0-20210111131231-96d45aba64e1
	github.com/pivotal-cf/brokerapi/v7 v7.5.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/vrischmann/envconfig v1.3.0
)
