module github.com/kyma-incubator/compass/tests/connector-tests

go 1.15

require (
	github.com/99designs/gqlgen v0.10.1 // indirect
	github.com/kyma-incubator/compass/components/connector v0.0.0-20201219152541-d77ebc00ac2d
	github.com/kyma-incubator/compass/tests/director v0.0.0-20210225132347-455a86db66a2
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
)
