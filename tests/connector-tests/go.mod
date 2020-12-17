module github.com/kyma-incubator/compass/tests/connector-tests

go 1.15

require (
	github.com/99designs/gqlgen v0.10.1 // indirect
	github.com/kyma-incubator/compass/components/connector v0.0.0-20200608084054-64f737ad7e1d
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
)

replace github.com/kyma-incubator/compass/components/connector => ../../components/connector
