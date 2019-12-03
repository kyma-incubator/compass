module github.com/kyma-incubator/compass/components/provisioner

go 1.13

require (
	github.com/99designs/gqlgen v0.9.0
	github.com/evanphx/json-patch v4.2.0+incompatible // indirect
	github.com/gocraft/dbr/v2 v2.6.3
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/terraform v0.12.16
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/kyma-incubator/hydroform v0.0.0-20191128070310-d7996cb46e38
	github.com/kyma-incubator/hydroform/install v0.0.0-20191128070310-d7996cb46e38
	github.com/lib/pq v1.2.0
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/vektah/gqlparser v1.2.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20191114100237-2cd11237263f
	k8s.io/apimachinery v0.0.0-20191004115701-31ade1b30762 // tag kubernetes-1.15.6
	k8s.io/client-go v0.0.0-20191114101336-8cba805ad12d // tag kubernetes-1.15.6
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a // indirect
)

replace github.com/kyma-incubator/hydroform => github.com/Szymongib/hydroform v0.0.0-20191203093804-e9fcbe88dcd5

replace github.com/kyma-incubator/hydroform/install => github.com/Szymongib/hydroform/install v0.0.0-20191203093804-e9fcbe88dcd5
