module github.com/kyma-incubator/compass/components/operations-controller

go 1.15

require (
	github.com/go-logr/logr v0.1.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20210216195113-0f23362ec238
	github.com/kyma-incubator/compass/components/system-broker v0.0.0-20210216195228-f2c108d93c69
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.4
	github.com/pkg/errors v0.9.1
	golang.org/x/oauth2 v0.0.0-20210201163806-010130855d6c
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)
