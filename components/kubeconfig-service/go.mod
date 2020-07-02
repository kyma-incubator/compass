module github.com/kyma-project/control-plane/components/kubeconfig-service

go 1.13

require (
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/howeyc/fsnotify v0.9.0
	github.com/kyma-project/control-plane/components/provisioner v0.0.0-20200702121812-3236dcca50b1
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.5.1
	github.com/vrischmann/envconfig v1.2.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/apiserver v0.17.2
)

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8
