module github.com/kyma-incubator/compass/components/operations-controller

go 1.16

require (
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/go-logr/logr v0.4.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20210825143858-d12ac209b4fe
	github.com/kyma-incubator/compass/components/system-broker v0.0.0-20210823104705-748ca0a6e5be
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/vektra/mockery/v2 v2.9.0 // indirect
	go.uber.org/zap v1.17.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)
