module github.com/kyma-incubator/compass/components/operations-controller

go 1.15

require (
	github.com/go-logr/logr v0.1.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20210216185717-b539fec3afce
	github.com/kyma-incubator/compass/components/system-broker v0.0.0-20210216190005-7c9faad763ec
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.4
	github.com/pkg/errors v0.9.1
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)
