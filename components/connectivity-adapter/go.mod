module github.com/kyma-incubator/compass/components/connectivity-adapter

go 1.13

require (
	github.com/asaskevich/govalidator v0.0.0-20180315120708-ccb8e960c48f
	github.com/gorilla/mux v1.7.3
	github.com/kyma-incubator/compass v0.0.0-20200123101435-9cd00b2924b8
	github.com/kyma-incubator/compass/components/director v0.0.0-20200120072209-565610bd185a
	github.com/kyma-incubator/compass/tests/director v0.0.0-20200120072209-565610bd185a
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/vektah/gqlparser v1.2.1 // indirect
	github.com/vrischmann/envconfig v1.2.0
	k8s.io/api v0.17.2 // indirect
	k8s.io/client-go v11.0.0+incompatible // indirect
)
