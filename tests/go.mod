module github.com/kyma-incubator/compass/tests

go 1.15

require (
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible
	github.com/google/uuid v1.3.0
	github.com/jmoiron/sqlx v1.3.4
	github.com/kyma-incubator/compass/components/connectivity-adapter v0.0.0-20210816091934-c8f38c361ff5
	github.com/kyma-incubator/compass/components/connector v0.0.0-20210816091934-c8f38c361ff5
	github.com/kyma-incubator/compass/components/director v0.0.0-20210816091934-c8f38c361ff5
	github.com/kyma-incubator/compass/components/external-services-mock v0.0.0-20210816091934-c8f38c361ff5
	github.com/kyma-incubator/compass/components/gateway v0.0.0-20210816091934-c8f38c361ff5
	github.com/kyma-incubator/compass/components/operations-controller v0.0.0-20210816091934-c8f38c361ff5
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.18.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.8.0
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/net v0.0.0-20210410081132-afb366fc7cd1
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
)
