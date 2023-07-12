package gqlclient

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/types/credentials"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/avast/retry-go/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var tenantHeader = "tenant"

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
		gqlRes = gqlResult{Result: &appBundles}
		if err := c.runWithTenant(ctx, gqlReq, tenant, &gqlRes); err != nil {
			return nil, errors.Wrapf(err, "Error while getting bundles for application with id %q", appID)
		}

		result = append(result, appBundles.Bundles.Data...)
	}

	return result, nil
}

// CreateBundleInstanceAuth creates bundle instance auth with credentials for bundle with bndlID and runtime with rtmID using the internal gql client
func (c *client) CreateBundleInstanceAuth(ctx context.Context, tenant, bndlID, rtmID string, credentials credentials.Credentials) error {
	gqlReq := gcli.NewRequest(fmt.Sprintf(`mutation {
  		result: createBundleInstanceAuth(
  		  bundleID: "%s"
  		  in: {
  		    auth: {
  		      credential: { %s }
  		    }
  		    runtimeID: "%s"
  		  }
		) {
    		id
		  }
		}`, bndlID, credentials.ToString(), rtmID))

	if err := c.runWithTenant(ctx, gqlReq, tenant, nil); err != nil {
		return errors.Wrapf(err, "Error while creating bundle instance auth for bundle with id %q and runtime with id %q", bndlID, rtmID)
	}

	return nil
}

// UpdateBundleInstanceAuth updates bundle instance auth with credentials using the internal gql client
func (c *client) UpdateBundleInstanceAuth(ctx context.Context, tenant, authID, bndlID string, credentials credentials.Credentials) error {
	gqlReq := gcli.NewRequest(fmt.Sprintf(`mutation {
  		result: updateBundleInstanceAuth(
		  id: "%s"
  		  bundleID: "%s"
  		  in: {
  		    auth: {
  		      credential: { %s }
  		    }
  		  }
		) {
    		id
		  }
		}`, authID, bndlID, credentials.ToString()))

	if err := c.runWithTenant(ctx, gqlReq, tenant, nil); err != nil {
		return errors.Wrapf(err, "Error while updating bundle instance auth with id %q for bundle with id %q", authID, bndlID)
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
