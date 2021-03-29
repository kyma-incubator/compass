module github.com/kyma-incubator/compass/tests

go 1.15

require (
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/google/uuid v1.2.0
	github.com/kyma-incubator/compass/components/connectivity-adapter v0.0.0-20210222141445-9f0329ae5b8d
	github.com/kyma-incubator/compass/components/connector v0.0.0-20210315172259-e186b4cac80b
	github.com/kyma-incubator/compass/components/director v0.0.0-20210329081516-c889c4bdd3ba
	github.com/kyma-incubator/compass/components/external-services-mock v0.0.0-20210309084252-cb1359ea9c14
	github.com/kyma-incubator/compass/components/gateway v0.0.0-20210224145945-7c0650085504
	github.com/kyma-incubator/compass/components/operations-controller v0.0.0-20210309084252-cb1359ea9c14
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.6.8
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	golang.org/x/oauth2 v0.0.0-20210220000619-9bb904979d93
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
)

replace (
	k8s.io/api => k8s.io/api v0.17.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.2
	k8s.io/client-go => k8s.io/client-go v0.17.2
)
