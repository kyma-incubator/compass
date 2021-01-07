module github.com/kyma-incubator/compass/tests/director

go 1.13

require (
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
	github.com/avast/retry-go v2.5.0+incompatible
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/kyma-incubator/compass/components/director v0.0.0-20210107130916-b97cabae65e4
	github.com/kyma-incubator/compass/components/gateway v0.0.0-20200429083609-7d80a85180c6
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/matryer/is v1.3.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.6.1
	github.com/vektah/gqlparser v1.3.1 // indirect
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
	k8s.io/apimachinery v0.17.3 // indirect
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
)
