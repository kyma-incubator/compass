module github.com/kyma-incubator/compass/components/system-broker

go 1.14

require (
	code.cloudfoundry.org/lager v2.0.0+incompatible
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/gavv/httpexpect/v2 v2.2.0
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/klauspost/compress v1.11.9 // indirect
	github.com/kyma-incubator/compass/components/director v0.0.0-20210330110957-2e17deac2600
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/maxbrunsfeld/counterfeiter/v6 v6.3.0
	github.com/pivotal-cf/brokerapi v6.4.2+incompatible
	github.com/pivotal-cf/brokerapi/v7 v7.5.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.8.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.6.8
	github.com/valyala/fasthttp v1.22.0 // indirect
	github.com/vektah/gqlparser v1.3.1
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)
