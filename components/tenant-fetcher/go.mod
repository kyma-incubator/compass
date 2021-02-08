module github.com/kyma-incubator/compass/components/tenant-fetcher

go 1.15

//replace github.com/kyma-incubator/compass/components/director => ../director

require (
	github.com/99designs/gqlgen v0.9.3
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/google/uuid v1.2.0
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20210208130910-3537b35e9350
	github.com/lestrrat-go/jwx v0.9.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.3.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/klog v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff v0.0.0-20190525122527-15d366b2352e // indirect
)
