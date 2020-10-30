module github.com/kyma-incubator/compass/tests/connectivity-adapter

go 1.13

require (
	github.com/99designs/gqlgen v0.10.2 // indirect
	github.com/avast/retry-go v2.5.0+incompatible
	github.com/kyma-incubator/compass/components/director v0.0.0-20201029155255-f6f02b451a2f
	github.com/kyma-incubator/compass/tests/director v0.0.0-20201029155255-f6f02b451a2f
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	github.com/vrischmann/envconfig v1.2.0
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
)

replace (
	github.com/kyma-incubator/compass/components/director => github.com/kyma-incubator/compass/components/director v0.0.0-20201029155255-f6f02b451a2f
	github.com/kyma-incubator/compass/components/gateway => github.com/kyma-incubator/compass/components/gateway v0.0.0-20201029161524-d83f1f7576b8
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
)
