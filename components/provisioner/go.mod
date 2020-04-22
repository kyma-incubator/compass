module github.com/kyma-incubator/compass/components/provisioner

go 1.13

require (
	github.com/99designs/gqlgen v0.9.3
	github.com/Microsoft/hcsshim v0.8.7 // indirect
	github.com/gardener/gardener v0.33.7
	github.com/gocraft/dbr/v2 v2.6.3
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform v0.12.16
	github.com/kubernetes/client-go v11.0.0+incompatible
	github.com/kyma-incubator/compass/components/director v0.0.0-20200417141145-3a08dc5142a0
	github.com/kyma-incubator/hydroform v0.0.0-20191217171037-affe7099c3b9
	github.com/kyma-incubator/hydroform/install v0.0.0-20200302090055-3f3b00d9799c
	github.com/kyma-project/kyma v0.5.1-0.20200323195746-ee2b142b8442
	github.com/lib/pq v1.2.0
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/testcontainers/testcontainers-go v0.3.1
	github.com/vektah/gqlparser v1.2.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	google.golang.org/genproto v0.0.0-20190927181202-20e1ac93f88c // indirect
	google.golang.org/grpc v1.24.0 // indirect
	k8s.io/api v0.15.8-beta.1
	k8s.io/apimachinery v0.15.8-beta.1
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible // tag kubernetes-1.15.6
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a // indirect
	sigs.k8s.io/controller-runtime v0.3.0
)

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

replace k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible => k8s.io/client-go v0.15.8-beta.1

replace github.com/Azure/go-autorest v11.5.0+incompatible => github.com/Azure/go-autorest v13.3.2+incompatible

replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200317142112-1b76d66859c6

replace github.com/gophercloud/gophercloud => github.com/gophercloud/gophercloud v0.0.0-20190125124242-bb1ef8ce758c

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
