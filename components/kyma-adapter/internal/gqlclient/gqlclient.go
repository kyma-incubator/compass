package gqlclient

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var tenantHeader = "tenant"

// Client is extended graphql.Client with custom requests
type Client struct {
	*gcli.Client
}

type gqlResult struct {
	Result interface{} `json:"result"`
}

// ApplicationBundles contains Bundle page for an application
type ApplicationBundles struct {
	Bundles graphql.BundlePageExt `json:"bundles"`
}

// GetApplicationBundles gets all bundles for an application with appId and tenant using the internal gql client
func (c *Client) GetApplicationBundles(ctx context.Context, appId, tenant string) ([]*graphql.BundleExt, error) {
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
	initialGqlReq := gcli.NewRequest(fmt.Sprintf(strReq, appId, pageSize, ""))
	initialGqlReq.Header.Set(tenantHeader, tenant)

	appBundles := ApplicationBundles{}
	gqlRes := gqlResult{Result: &appBundles}

	if err := c.Client.Run(ctx, initialGqlReq, &gqlRes); err != nil {
		return nil, errors.Wrapf(err, "Error while getting bundles for application with id %q", appId)
	}

	result := make([]*graphql.BundleExt, 0, appBundles.Bundles.TotalCount)
	result = append(result, appBundles.Bundles.Data...)

	for appBundles.Bundles.PageInfo.EndCursor != "" {
		gqlReq := gcli.NewRequest(fmt.Sprintf(strReq, appId, pageSize, appBundles.Bundles.PageInfo.EndCursor))
		gqlReq.Header.Set(tenantHeader, tenant)

		appBundles = ApplicationBundles{}
		if err := c.Client.Run(ctx, gqlReq, &appBundles); err != nil {
			return nil, errors.Wrapf(err, "Error while getting bundles for application with id %q", appId)
		}

		result = append(result, appBundles.Bundles.Data...)
	}

	return result, nil
}

// CreateBasicBundleInstanceAuth creates bundle instance auth with basic credentials for bundle with bndlId and runtime with rtmId using the internal gql client
func (c *Client) CreateBasicBundleInstanceAuth(ctx context.Context, tenant, bndlId, rtmId, username, password string) error {
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
		}`, bndlId, username, password, rtmId))

	gqlReq.Header.Set(tenantHeader, tenant)

	if err := c.Client.Run(ctx, gqlReq, nil); err != nil {
		return errors.Wrapf(err, "Error while creating Basic bundle instance auth for bundle with id %q and runtime with id %q", bndlId, rtmId)
	}

	return nil
}

// CreateOauthBundleInstanceAuth creates bundle instance auth with oauth credentials for bundle with bndlId and runtime with rtmId using the internal gql client
func (c *Client) CreateOauthBundleInstanceAuth(ctx context.Context, tenant, bndlId, rtmId, tokenServiceUrl, clientId, clientSecret string) error {
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
		}`, bndlId, clientId, clientSecret, tokenServiceUrl, rtmId))

	gqlReq.Header.Set(tenantHeader, tenant)

	if err := c.Client.Run(ctx, gqlReq, nil); err != nil {
		return errors.Wrapf(err, "Error while creating Basic bundle instance auth for bundle with id %q and runtime with id %q", bndlId, rtmId)
	}

	return nil
}

// UpdateBasicBundleInstanceAuth updates bundle instance auth with basic credentials using the internal gql client
func (c *Client) UpdateBasicBundleInstanceAuth(ctx context.Context, tenant, authId, bndlId, username, password string) error {
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
		}`, authId, bndlId, username, password))

	gqlReq.Header.Set(tenantHeader, tenant)

	if err := c.Client.Run(ctx, gqlReq, nil); err != nil {
		return errors.Wrapf(err, "Error while updating bundle instance auth with Basic credentials with id %q for bundle with id %q", authId, bndlId)
	}

	return nil
}

// UpdateOauthBundleInstanceAuth updates bundle instance auth with oauth credentials using the internal gql client
func (c *Client) UpdateOauthBundleInstanceAuth(ctx context.Context, tenant, authId, bndlId, tokenServiceUrl, clientId, clientSecret string) error {
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
		}`, authId, bndlId, clientId, clientSecret, tokenServiceUrl))

	gqlReq.Header.Set(tenantHeader, tenant)

	if err := c.Client.Run(ctx, gqlReq, nil); err != nil {
		return errors.Wrapf(err, "Error while updating bundle instance auth with Oauth credentials with id %q for bundle with id %q", authId, bndlId)
	}

	return nil
}

// DeleteBundleInstanceAuth deletes bundle instance auth with authId using the internal gql client
func (c *Client) DeleteBundleInstanceAuth(ctx context.Context, tenant, authId string) error {
	gqlReq := gcli.NewRequest(fmt.Sprintf(`mutation {
		result: deleteBundleInstanceAuth(
			authID: "%s"
		) {
    		id
		  }
		}`, authId))

	gqlReq.Header.Set(tenantHeader, tenant)

	if err := c.Client.Run(ctx, gqlReq, nil); err != nil {
		return errors.Wrapf(err, "Error while deleting bundle instance auth with id %q", authId)
	}

	return nil
}
