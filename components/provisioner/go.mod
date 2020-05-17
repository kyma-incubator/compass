module github.com/kyma-incubator/compass/components/provisioner

go 1.13

require (
	github.com/99designs/gqlgen v0.9.3
	github.com/Microsoft/hcsshim v0.8.7 // indirect
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/gardener/gardener v0.33.7
	github.com/gocraft/dbr/v2 v2.6.3
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform v0.12.16
	github.com/kubernetes-incubator/service-catalog v0.2.2
	github.com/kubernetes-sigs/service-catalog v0.2.2
	github.com/kubernetes/client-go v11.0.0+incompatible
	github.com/kyma-incubator/compass/components/director v0.0.0-20200506060219-a2a2a07e1283
	github.com/kyma-incubator/hydroform v0.0.0-20191217171037-affe7099c3b9
	github.com/kyma-incubator/hydroform/install v0.0.0-20200414071650-35d4d6f8c53e
	github.com/kyma-project/kyma v0.5.1-0.20200323195746-ee2b142b8442
	github.com/kyma-project/kyma/components/compass-runtime-agent v0.0.0-20200422062252-6074323197a6
	github.com/lib/pq v1.2.0
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/testcontainers/testcontainers-go v0.3.1
	github.com/vektah/gqlparser v1.2.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	google.golang.org/genproto v0.0.0-20190927181202-20e1ac93f88c // indirect
	google.golang.org/grpc v1.24.0 // indirect
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible // tag kubernetes-1.15.6
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.5.0
)

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

replace k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible => k8s.io/client-go v0.15.8-beta.1

replace k8s.io/apimachinery v0.17.2 => k8s.io/apimachinery v0.15.8-beta.1

replace k8s.io/apiextensions-apiserver v0.17.2 => k8s.io/apiextensions-apiserver v0.15.8-beta.1

replace k8s.io/api v0.17.2 => k8s.io/api v0.15.8-beta.1

replace sigs.k8s.io/controller-runtime v0.5.0 => sigs.k8s.io/controller-runtime v0.3.0

replace github.com/Azure/go-autorest v11.5.0+incompatible => github.com/Azure/go-autorest v13.3.2+incompatible

replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200317142112-1b76d66859c6

replace github.com/gophercloud/gophercloud => github.com/gophercloud/gophercloud v0.0.0-20190125124242-bb1ef8ce758c

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
