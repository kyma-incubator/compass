package gqlclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/avast/retry-go/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var tenantHeader = "tenant"

// Client is responsible for making internal graphql calls to the director
type Client interface {
	GetApplicationBundles(ctx context.Context, appID, tenant string) ([]*graphql.BundleExt, error)
	CreateBasicBundleInstanceAuth(ctx context.Context, tenant, bndlID, rtmID, username, password string) error
	CreateOauthBundleInstanceAuth(ctx context.Context, tenant, bndlID, rtmID, tokenServiceURL, clientID, clientSecret string) error
	UpdateBasicBundleInstanceAuth(ctx context.Context, tenant, authID, bndlID, username, password string) error
	UpdateOauthBundleInstanceAuth(ctx context.Context, tenant, authID, bndlID, tokenServiceURL, clientID, clientSecret string) error
	DeleteBundleInstanceAuth(ctx context.Context, tenant, authID string) error
}

// NewClient constructs a client
func NewClient(gqlClient *gcli.Client) *client {
	return &client{Client: gqlClient}
}

// Client is extended graphql.Client with custom requests
type client struct {
	*gcli.Client
}

type gqlResult struct {
	Result interface{} `json:"result"`
}

// ApplicationBundles contains Bundle page for an application
type ApplicationBundles struct {
	Bundles graphql.BundlePageExt `json:"bundles"`
}

// runWithTenant executes gql request with tenant header and with retry on connectivity problems
func (c *client) runWithTenant(ctx context.Context, gqlReq *gcli.Request, tenant string, resp interface{}) error {
	gqlReq.Header.Set(tenantHeader, tenant)

	return withRetryOnTemporaryConnectionProblems(ctx, func() error {
		return c.Client.Run(ctx, gqlReq, resp)
	})
}

// GetApplicationBundles gets all bundles for an application with appID and tenant using the internal gql client
func (c *client) GetApplicationBundles(ctx context.Context, appID, tenant string) ([]*graphql.BundleExt, error) {
	strReq := `query {
  			result: application(id: "%s") {
  			  bundles(first:%d, after:"%s") {
  			    data {
  			      id
				  instanceAuths {
				    id
				    runtimeID
				  }
  			    }
  			    pageInfo {
  			      hasNextPage
  			      endCursor
  			    }
  			    totalCount
  			  }
  			}
		}`
	pageSize := 200
	initialGqlReq := gcli.NewRequest(fmt.Sprintf(strReq, appID, pageSize, ""))

	appBundles := ApplicationBundles{}
	gqlRes := gqlResult{Result: &appBundles}

	if err := c.runWithTenant(ctx, initialGqlReq, tenant, &gqlRes); err != nil {
		return nil, errors.Wrapf(err, "Error while getting bundles for application with id %q", appID)
	}

	result := make([]*graphql.BundleExt, 0, appBundles.Bundles.TotalCount)
	result = append(result, appBundles.Bundles.Data...)

	for appBundles.GetBundlesEndCursor() != "" {
		gqlReq := gcli.NewRequest(fmt.Sprintf(strReq, appID, pageSize, appBundles.GetBundlesEndCursor()))

		appBundles = ApplicationBundles{}
		if err := c.runWithTenant(ctx, gqlReq, tenant, &appBundles); err != nil {
			return nil, errors.Wrapf(err, "Error while getting bundles for application with id %q", appID)
		}

		result = append(result, appBundles.Bundles.Data...)
	}

	return result, nil
}

// CreateBasicBundleInstanceAuth creates bundle instance auth with basic credentials for bundle with bndlID and runtime with rtmID using the internal gql client
func (c *client) CreateBasicBundleInstanceAuth(ctx context.Context, tenant, bndlID, rtmID, username, password string) error {
	gqlReq := gcli.NewRequest(fmt.Sprintf(`mutation {
  		result: createBundleInstanceAuth(
  		  bundleID: "%s"
  		  in: {
  		    auth: {
  		      credential: { basic: { username: "%s", password: "%s" } }
  		    }
  		    runtimeID: "%s"
  		  }
		) {
    		id
		  }
		}`, bndlID, username, password, rtmID))

	if err := c.runWithTenant(ctx, gqlReq, tenant, nil); err != nil {
		return errors.Wrapf(err, "Error while creating Basic bundle instance auth for bundle with id %q and runtime with id %q", bndlID, rtmID)
	}

	return nil
}

// CreateOauthBundleInstanceAuth creates bundle instance auth with oauth credentials for bundle with bndlID and runtime with rtmID using the internal gql client
func (c *client) CreateOauthBundleInstanceAuth(ctx context.Context, tenant, bndlID, rtmID, tokenServiceURL, clientID, clientSecret string) error {
	gqlReq := gcli.NewRequest(fmt.Sprintf(`mutation {
  		result: createBundleInstanceAuth(
  		  bundleID: "%s"
  		  in: {
  		    auth: {
  		      credential: { oauth: { clientId: "%s" clientSecret: "%s" url: "%s"} }
  		    }
  		    runtimeID: "%s"
  		  }
		) {
    		id
		  }
		}`, bndlID, clientID, clientSecret, tokenServiceURL, rtmID))

	if err := c.runWithTenant(ctx, gqlReq, tenant, nil); err != nil {
		return errors.Wrapf(err, "Error while creating Basic bundle instance auth for bundle with id %q and runtime with id %q", bndlID, rtmID)
	}

	return nil
}

// UpdateBasicBundleInstanceAuth updates bundle instance auth with basic credentials using the internal gql client
func (c *client) UpdateBasicBundleInstanceAuth(ctx context.Context, tenant, authID, bndlID, username, password string) error {
	gqlReq := gcli.NewRequest(fmt.Sprintf(`mutation {
  		result: updateBundleInstanceAuth(
		  id: "%s"
  		  bundleID: "%s"
  		  in: {
  		    auth: {
  		      credential: { basic: { username: "%s", password: "%s" } }
  		    }
  		  }
		) {
    		id
		  }
		}`, authID, bndlID, username, password))

	if err := c.runWithTenant(ctx, gqlReq, tenant, nil); err != nil {
		return errors.Wrapf(err, "Error while updating bundle instance auth with Basic credentials with id %q for bundle with id %q", authID, bndlID)
	}

	return nil
}

// UpdateOauthBundleInstanceAuth updates bundle instance auth with oauth credentials using the internal gql client
func (c *client) UpdateOauthBundleInstanceAuth(ctx context.Context, tenant, authID, bndlID, tokenServiceURL, clientID, clientSecret string) error {
	gqlReq := gcli.NewRequest(fmt.Sprintf(`mutation {
  		result: updateBundleInstanceAuth(
		  id: "%s"
  		  bundleID: "%s"
  		  in: {
  		    auth: {
  		      credential: { oauth: { clientId: "%s" clientSecret: "%s" url: "%s"} }
  		    }
  		  }
		) {
    		id
		  }
		}`, authID, bndlID, clientID, clientSecret, tokenServiceURL))

	if err := c.runWithTenant(ctx, gqlReq, tenant, nil); err != nil {
		return errors.Wrapf(err, "Error while updating bundle instance auth with Oauth credentials with id %q for bundle with id %q", authID, bndlID)
	}

	return nil
}

// DeleteBundleInstanceAuth deletes bundle instance auth with authID using the internal gql client
func (c *client) DeleteBundleInstanceAuth(ctx context.Context, tenant, authID string) error {
	gqlReq := gcli.NewRequest(fmt.Sprintf(`mutation {
		result: deleteBundleInstanceAuth(
			authID: "%s"
		) {
    		id
		  }
		}`, authID))

	if err := c.runWithTenant(ctx, gqlReq, tenant, nil); err != nil {
		return errors.Wrapf(err, "Error while deleting bundle instance auth with id %q", authID)
	}

	return nil
}

// GetBundlesEndCursor returns bundles page end cursor
func (a *ApplicationBundles) GetBundlesEndCursor() string {
	if a.Bundles.PageInfo == nil {
		return ""
	}

	return string(a.Bundles.PageInfo.EndCursor)
}

func withRetryOnTemporaryConnectionProblems(ctx context.Context, risky func() error) error {
	return retry.Do(risky, retry.Attempts(7), retry.Delay(time.Second), retry.OnRetry(func(n uint, err error) {
		log.C(ctx).Warnf("OnRetry: attempts: %d, error: %v", n, err)
	}), retry.LastErrorOnly(true), retry.RetryIf(func(err error) bool {
		return strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "connection reset by peer")
	}))
}
