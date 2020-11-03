module github.com/kyma-incubator/compass/components/connectivity-adapter

go 1.13

require (
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/google/go-cmp v0.5.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/kyma-incubator/compass v0.0.0-20200703104319-1c4490318bfd
	github.com/kyma-incubator/compass/components/director v0.0.0-20201102113127-98bfd9c66077
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/sys v0.0.0-20201020230747-6e5568b54d1a // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/asaskevich/govalidator.v9 v9.0.0-20180315120708-ccb8e960c48f
	k8s.io/api v0.17.2 // indirect
	k8s.io/client-go v11.0.0+incompatible // indirect
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9 // indirect
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
)
