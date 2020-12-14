module github.com/kyma-incubator/compass/components/gateway

go 1.13

require (
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/kyma-incubator/compass/components/director v0.0.0-20201109133626-4876e6d3caae
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.6.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
