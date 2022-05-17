package gqlcli

import (
	"context"
	"net/http"
	"time"

	gcli "github.com/machinebox/graphql"
)

//go:generate mockery --name=GraphQLClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type GraphQLClient interface {
	Run(ctx context.Context, req *gcli.Request, resp interface{}) error
}

//go:generate mockery --name=Provider --output=automock --outpkg=automock --case=underscore --disable-version-string
type Provider interface {
	GQLClient(rq *http.Request) GraphQLClient
}

type graphQLClientProvider struct {
	url     string
	timeout time.Duration
}

func NewProvider(url string, timeout time.Duration) *graphQLClientProvider {
	return &graphQLClientProvider{
		url:     url,
		timeout: timeout,
	}
}

func (p *graphQLClientProvider) GQLClient(rq *http.Request) GraphQLClient {
	return NewAuthorizedGraphQLClient(p.url, p.timeout, rq)
}
