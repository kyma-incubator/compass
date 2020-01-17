package gqlcli

import (
	"context"
	"net/http"

	gcli "github.com/machinebox/graphql"
)

//go:generate mockery -name=GraphQLClient -output=automock -outpkg=automock -case=underscore
type GraphQLClient interface {
	Run(ctx context.Context, req *gcli.Request, resp interface{}) error
}

//go:generate mockery -name=Provider -output=automock -outpkg=automock -case=underscore
type Provider interface {
	GQLClient(rq *http.Request) GraphQLClient
}

type graphQLClientProvider struct {
	url string
}

func NewProvider(url string) *graphQLClientProvider {
	return &graphQLClientProvider{url: url}
}

func (p *graphQLClientProvider) GQLClient(rq *http.Request) GraphQLClient {
	return NewAuthorizedGraphQLClient(p.url, rq)
}
