module github.com/kyma-incubator/compass/components/kubeconfig-service

go 1.13

require (
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/kyma-incubator/compass/components/provisioner v0.0.0-20200610112737-6086a0421bda
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	google.golang.org/appengine v1.6.5
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8
