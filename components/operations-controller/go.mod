module github.com/kyma-incubator/compass/components/operations-controller

go 1.15

require (
	github.com/go-logr/logr v0.1.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20210215163040-bf34a18315ef
	github.com/kyma-incubator/compass/components/system-broker v0.0.0-20210215163343-ece06e94d3e4
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.4
	github.com/pkg/errors v0.9.1
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)
