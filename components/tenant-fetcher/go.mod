module github.com/kyma-incubator/compass/components/tenant-fetcher

go 1.15

//replace github.com/kyma-incubator/compass/components/director => ../director

require (
	github.com/google/uuid v1.1.5
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20210202150643-acb481d32fce
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/klog v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff v0.0.0-20190525122527-15d366b2352e // indirect
)
