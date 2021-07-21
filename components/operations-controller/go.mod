module github.com/kyma-incubator/compass/components/operations-controller

go 1.15

require (
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible
	github.com/go-logr/logr v0.4.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20210706102506-68f29bff5d41
	github.com/kyma-incubator/compass/components/system-broker v0.0.0-20210623072749-c159dfdd6510
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)
