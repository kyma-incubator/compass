module github.com/kyma-incubator/compass/components/system-broker

go 1.14

require (
	code.cloudfoundry.org/lager v2.0.0+incompatible
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/fatih/structs v1.1.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gavv/httpexpect/v2 v2.2.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20210823151554-3f81b3ffaa47
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/maxbrunsfeld/counterfeiter/v6 v6.4.1
	github.com/pivotal-cf/brokerapi/v7 v7.5.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.8.0
	github.com/vektah/gqlparser/v2 v2.1.0
	github.com/vektra/mockery/v2 v2.9.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)
