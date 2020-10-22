module github.com/kyma-incubator/compass/components/pairing-adapter

go 1.13

require (
	github.com/kyma-incubator/compass/components/director v0.0.0-20201022103343-c480a02149cd
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9 // indirect
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
)
