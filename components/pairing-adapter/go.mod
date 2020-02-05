module github.com/kyma-incubator/compass/components/pairing-adapter

go 1.13

require (
	github.com/kyma-incubator/compass/components/director v0.0.0-20200205080921-fcbe3d2c0f3a
	github.com/motemen/go-loghttp v0.0.0-20170804080138-974ac5ceac27 // indirect
	github.com/motemen/go-nuts v0.0.0-20190725124253-1d2432db96b0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
)

replace github.com/kyma-incubator/compass/components/director => github.com/aszecowka/compass/components/director v0.0.0-20200217154136-2b31626a6aac
