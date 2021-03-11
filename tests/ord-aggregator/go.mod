module github.com/kyma-incubator/compass/tests/ord-aggregator

go 1.15

require (
	github.com/kyma-incubator/compass/components/director v0.0.0-20210310094754-d8dd471b9fdd
	github.com/kyma-incubator/compass/tests/director v0.0.0-20210304094417-deb791b48d6a
	github.com/kyma-incubator/compass/tests/ord-service v0.0.0-20210309203806-32c5f062ca73

	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.6.8
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/oauth2 v0.0.0-20210220000619-9bb904979d93
)
