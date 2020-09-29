module github.com/kyma-incubator/compass/components/system-broker

go 1.14

require (
	code.cloudfoundry.org/lager v2.0.0+incompatible
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/fatih/structs v1.1.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/kyma-incubator/compass/components/director v0.0.0-20200807110214-4424c47ecd1c
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/onrik/logrus v0.7.0
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/pivotal-cf/brokerapi v6.4.2+incompatible
	github.com/pivotal-cf/brokerapi/v7 v7.3.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.6.0
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cast v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/gjson v1.6.0 // indirect
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
)
