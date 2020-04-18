module github.com/kyma-incubator/compass/components/metris

go 1.14

require (
	github.com/Azure/azure-sdk-for-go v41.1.0+incompatible
	github.com/Azure/go-autorest/autorest v0.10.0
	github.com/Azure/go-autorest/autorest/adal v0.8.3 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/alecthomas/kong v0.2.4
	github.com/gardener/gardener v1.1.6
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.2.2
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.5.1
	github.com/stretchr/testify v1.5.1
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20200406173513-056763e48d71 // indirect
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	k8s.io/api v0.18.0
	k8s.io/apimachinery v0.18.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

replace (
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655 // kubernetes-1.16.0
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90 // kubernetes-1.16.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190918161219-8c8f079fddc3 // kubernetes-1.16.0
)
