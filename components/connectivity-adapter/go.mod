module github.com/kyma-incubator/compass/components/connectivity-adapter

go 1.15

require (
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/kyma-incubator/compass/components/connector v0.0.0-20210526113340-87c6e3c6f049
	github.com/kyma-incubator/compass/components/director v0.0.0-20210607082003-d97c798f2482
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210622073645-42b139061644
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.3.0
	gopkg.in/asaskevich/govalidator.v9 v9.0.0-20180315120708-ccb8e960c48f
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
)
