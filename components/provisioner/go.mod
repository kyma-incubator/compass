module github.com/kyma-incubator/compass/components/provisioner

go 1.13

require (
	github.com/99designs/gqlgen v0.9.3
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Microsoft/hcsshim v0.8.7 // indirect
	github.com/evanphx/json-patch v4.2.0+incompatible // indirect
	github.com/gocraft/dbr/v2 v2.6.3
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/terraform v0.12.16
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/kubernetes/client-go v11.0.0+incompatible
	github.com/kyma-incubator/compass/components/director v0.0.0-20200124100441-afb7816716c1
	github.com/kyma-incubator/hydroform v0.0.0-20191217171037-affe7099c3b9
	github.com/kyma-incubator/hydroform/install v0.0.0-20191217171037-affe7099c3b9
	github.com/kyma-project/kyma v0.5.1-0.20191106070956-5aa08d114ca0
	github.com/lib/pq v1.2.0
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/matryer/is v1.2.0 // indirect
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/testcontainers/testcontainers-go v0.0.8
	github.com/vektah/gqlparser v1.2.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/sys v0.0.0-20191220142924-d4481acd189f // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect
	k8s.io/api v0.0.0-20191114100237-2cd11237263f
	k8s.io/apimachinery v0.0.0-20191004115701-31ade1b30762
	k8s.io/client-go v0.0.0-20191114101336-8cba805ad12d // tag kubernetes-1.15.6
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a // indirect
)

replace github.com/kyma-incubator/compass => github.com/rafalpotempa/compass v0.0.0-20200124094425-387a1474f08c
