package e2e

import (
	"fmt"
)

type playgroundURLBuilder struct {
	format             string
	domain             string
	graphQLExamplePath string
}

func newPlaygroundURLBuilder(cfg *playgroundTestConfig) *playgroundURLBuilder {
	return &playgroundURLBuilder{format: cfg.DirectorURLFormat, domain: cfg.Gateway.Domain, graphQLExamplePath: cfg.DirectorGraphQLExamplePath}
}

func (b *playgroundURLBuilder) getRedirectionStartURL(subdomain string) string {
	redirectionURL := fmt.Sprintf(b.format, subdomain, b.domain)
	return redirectionURL
}

func (b *playgroundURLBuilder) getFinalURL(subdomain string) string {
	finalURL := fmt.Sprintf("%s/", b.getRedirectionStartURL(subdomain))
	return finalURL
}

func (b *playgroundURLBuilder) getGraphQLExampleURL(subdomain string) string {
	return fmt.Sprintf("%s%s", b.getFinalURL(subdomain), b.graphQLExamplePath)
}
