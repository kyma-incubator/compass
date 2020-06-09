module github.com/kyma-incubator/compass/components/connectivity-adapter

go 1.13

require (
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/kyma-incubator/compass v0.0.0-20200528095257-9103d98058ce
	github.com/kyma-incubator/compass/components/director v0.0.0-20200528095257-9103d98058ce
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/matryer/is v1.3.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/vektah/gqlparser v1.3.1 // indirect
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9 // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2 // indirect
	gopkg.in/asaskevich/govalidator.v9 v9.0.0-20180315120708-ccb8e960c48f
	k8s.io/api v0.17.2 // indirect
	k8s.io/apimachinery v0.17.3 // indirect
	k8s.io/client-go v11.0.0+incompatible // indirect
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9 // indirect
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
)
