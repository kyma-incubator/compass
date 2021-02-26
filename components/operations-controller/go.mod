module github.com/kyma-incubator/compass/components/operations-controller

go 1.15

require (
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/go-logr/logr v0.1.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20210226012137-cc67531aad37
	github.com/kyma-incubator/compass/components/system-broker v0.0.0-20210226012225-860e8d345c4c
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.4
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/oauth2 v0.0.0-20210201163806-010130855d6c
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)
