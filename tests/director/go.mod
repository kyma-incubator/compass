module github.com/kyma-incubator/compass/tests/director-tests

go 1.13

require (
	github.com/99designs/gqlgen v0.10.2 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/kyma-incubator/compass/components/director v0.0.0-00010101000000-000000000000
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/matryer/is v1.2.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/vrischmann/envconfig v1.2.0
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20191227163750-53104e6ec876 // indirect
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553
	golang.org/x/sys v0.0.0-20191228213918-04cbcbbfeed8 // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect
	k8s.io/apimachinery v0.17.0 // indirect
)

replace github.com/kyma-incubator/compass/components/director => github.com/aszecowka/compass/components/director v0.0.0-20200110114546-cc2fc148edd2
