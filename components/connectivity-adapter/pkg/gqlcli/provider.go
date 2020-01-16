package gqlcli

import (
	gcli "github.com/machinebox/graphql"
	"net/http"
)

//go:generate mockery -name=Provider -output=automock -outpkg=automock -case=underscore
type Provider interface {
	GQLClient(rq *http.Request) *gcli.Client
}

type graphQLClientProvider struct {
	url string
}

func NewProvider(url string) *graphQLClientProvider {
	return &graphQLClientProvider{url: url}
}

func (p *graphQLClientProvider) GQLClient(rq *http.Request) *gcli.Client {
	return NewAuthorizedGraphQLClient(p.url, rq)
}

