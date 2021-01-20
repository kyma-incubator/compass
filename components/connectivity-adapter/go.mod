module github.com/kyma-incubator/compass/components/connectivity-adapter

go 1.15

require (
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.7.3
	github.com/kyma-incubator/compass v0.0.0-20200703104319-1c4490318bfd
	github.com/kyma-incubator/compass/components/director v0.0.0-20210120182142-72278004f5e6 // indirect
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.6.1
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/asaskevich/govalidator.v9 v9.0.0-20180315120708-ccb8e960c48f
	k8s.io/client-go v11.0.0+incompatible // indirect
	k8s.io/klog v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff v0.0.0-20190525122527-15d366b2352e // indirect
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9 // indirect
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
)
