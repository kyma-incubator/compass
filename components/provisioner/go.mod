module github.com/kyma-incubator/compass/components/provisioner

go 1.13

require (
	contrib.go.opencensus.io/exporter/ocagent v0.4.12 // indirect
	github.com/99designs/gqlgen v0.9.3
	github.com/Azure/azure-pipeline-go v0.2.2 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Microsoft/hcsshim v0.8.7 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/gardener/external-dns-management v0.0.0-20190722114702-f6b12f6e4b43 // indirect
	github.com/gardener/gardener v0.33.7
	github.com/go-ini/ini v1.46.0 // indirect
	github.com/gobuffalo/logger v1.0.1 // indirect
	github.com/gobuffalo/packd v0.3.0 // indirect
	github.com/gocraft/dbr/v2 v2.6.3
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform v0.12.16
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/karrick/godirwalk v1.10.12 // indirect
	github.com/kubernetes/client-go v11.0.0+incompatible
	github.com/kyma-incubator/compass/components/director v0.0.0-20200123101435-9cd00b2924b8
	github.com/kyma-incubator/hydroform v0.0.0-20191217171037-affe7099c3b9
	github.com/kyma-incubator/hydroform/install v0.0.0-20200302090055-3f3b00d9799c
	github.com/kyma-project/kyma v0.5.1-0.20191106070956-5aa08d114ca0
	github.com/lib/pq v1.2.0
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/matryer/is v1.2.0 // indirect
	github.com/mattn/go-ieproxy v0.0.0-20190805055040-f9202b1cfdeb // indirect
	github.com/munnerz/goautoneg v0.0.0-20190414153302-2ae31c8b6b30 // indirect
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/terraform-providers/terraform-provider-openstack v1.20.0 // indirect
	github.com/testcontainers/testcontainers-go v0.3.1
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/ulikunitz/xz v0.5.6 // indirect
	github.com/vektah/gqlparser v1.2.0
	github.com/vrischmann/envconfig v1.2.0
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	google.golang.org/appengine v1.6.4 // indirect
	google.golang.org/genproto v0.0.0-20190927181202-20e1ac93f88c // indirect
	google.golang.org/grpc v1.24.0 // indirect
	gopkg.in/ini.v1 v1.44.0 // indirect
	k8s.io/api v0.15.8-beta.1
	k8s.io/apimachinery v0.15.8-beta.1
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible // tag kubernetes-1.15.6
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a // indirect
	k8s.io/utils v0.0.0-20190712204705-3dccf664f023 // indirect
	sigs.k8s.io/controller-runtime v0.3.0
)

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

replace k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible => k8s.io/client-go v0.15.8-beta.1

replace github.com/Azure/go-autorest v11.5.0+incompatible => github.com/Azure/go-autorest v13.3.2+incompatible
