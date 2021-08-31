module github.com/kyma-incubator/compass/components/operations-controller

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20210831105733-c49b027c21d3
	github.com/kyma-incubator/compass/components/system-broker v0.0.0-20210831105733-c49b027c21d3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.17.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)
