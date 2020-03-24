module github.com/kyma-incubator/compass/tests/director

go 1.13

require (
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
	github.com/avast/retry-go v2.5.0+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/kyma-incubator/compass/components/director v0.0.0-20200323155219-340bb3dd067f
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/matryer/is v1.3.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/vektah/gqlparser v1.3.1 // indirect
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/crypto v0.0.0-20200210222208-86ce3cb69678 // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/sys v0.0.0-20200212091648-12a6c2dcc1e4 // indirect
	k8s.io/apimachinery v0.17.3 // indirect
)

replace github.com/kyma-incubator/compass/components/director => github.com/crabtree/compass/components/director v0.0.0-20200324075036-31ccf2fe456e
