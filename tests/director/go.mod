module github.com/kyma-incubator/compass/tests/director

go 1.13

require (
	github.com/avast/retry-go v2.5.0+incompatible
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/google/uuid v1.1.2
	github.com/kyma-incubator/compass/components/director v0.0.0-20210119151733-586ca2a3ee31 // indirect
	github.com/kyma-incubator/compass/components/gateway v0.0.0-20200429083609-7d80a85180c6
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.6.1
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
)
