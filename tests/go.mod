module github.com/kyma-incubator/compass/tests

go 1.15

require (
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/google/uuid v1.2.0
	github.com/kyma-incubator/compass/components/connectivity-adapter v0.0.0-20210211152841-1e543bc032f6
	github.com/kyma-incubator/compass/components/connector v0.0.0-20210203135116-086a057e4d3c
	github.com/kyma-incubator/compass/components/director v0.0.0-20210211152841-1e543bc032f6
	github.com/kyma-incubator/compass/components/gateway v0.0.0-20200429083609-7d80a85180c6
	github.com/kyma-incubator/compass/tests/director v0.0.0-20210212135956-7783dc62b2fc
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.6.8
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	golang.org/x/oauth2 v0.0.0-20210113205817-d3ed898aa8a3
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v11.0.0+incompatible
)
