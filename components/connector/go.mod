module github.com/kyma-incubator/compass/components/connector

go 1.17

require (
	github.com/99designs/gqlgen v0.11.0
	github.com/gorilla/mux v1.8.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20211005135133-e6f63cd73aa6
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vektah/gqlparser/v2 v2.1.0
	github.com/vrischmann/envconfig v1.3.0
	k8s.io/api v0.20.2 //DO NOT BUMP
	k8s.io/apimachinery v0.20.2 //DO NOT BUMP
	k8s.io/client-go v0.20.2 //DO NOT BUMP
)
