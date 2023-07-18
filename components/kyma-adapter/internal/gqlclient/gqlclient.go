package gqlclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/types/credentials"

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

// BundleInstanceAuthInput represents the input needed by the bundle instance auth modify functions
type BundleInstanceAuthInput interface{}

// CreateBundleInstanceAuthInput represents the input needed by the CreateBundleInstanceAuth function
type CreateBundleInstanceAuthInput struct {
	BundleID    string
	RuntimeID   string
	Credentials credentials.Credentials
}

// UpdateBundleInstanceAuthInput represents the input needed by the UpdateBundleInstanceAuth function
type UpdateBundleInstanceAuthInput struct {
	Bundle      *graphql.BundleExt
	RuntimeID   string
	Credentials credentials.Credentials
}

// DeleteBundleInstanceAuthInput represents the input needed by the DeleteBundleInstanceAuth function
type DeleteBundleInstanceAuthInput struct {
	Bundle    *graphql.BundleExt
	RuntimeID string
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

	appBundles := ApplicationBundles{}
	result := make([]*graphql.BundleExt, 0)

	for shouldFetch := true; shouldFetch; shouldFetch = (appBundles.GetBundlesEndCursor() != "") {
		gqlReq := gcli.NewRequest(fmt.Sprintf(strReq, appID, pageSize, appBundles.GetBundlesEndCursor()))

		appBundles = ApplicationBundles{}
		gqlRes := gqlResult{Result: &appBundles}
		if err := c.runWithTenant(ctx, gqlReq, tenant, &gqlRes); err != nil {
			errMsg := fmt.Sprintf("Error while getting bundles for application with id %q", appID)
			log.C(ctx).WithError(err).Error(errMsg)
			return nil, errors.New(errMsg)
		}

		result = append(result, appBundles.Bundles.Data...)
	}

	return result, nil
}

// CreateBundleInstanceAuth creates bundle instance auth with a given input
func (c *client) CreateBundleInstanceAuth(ctx context.Context, tenant string, input BundleInstanceAuthInput) error {
	createInput, ok := input.(CreateBundleInstanceAuthInput)
	if !ok {
		return errors.New("while casting input to create input")
	}

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
		}`, createInput.BundleID, createInput.Credentials.ToString(), createInput.RuntimeID))

	if err := c.runWithTenant(ctx, gqlReq, tenant, nil); err != nil {
		errMsg := fmt.Sprintf("Error while creating bundle instance auth for bundle with id %q and runtime with id %q", createInput.BundleID, createInput.RuntimeID)
		log.C(ctx).WithError(err).Error(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

// UpdateBundleInstanceAuth updates bundle instance auth with a given input
func (c *client) UpdateBundleInstanceAuth(ctx context.Context, tenant string, input BundleInstanceAuthInput) error {
	updateInput, ok := input.(UpdateBundleInstanceAuthInput)
	if !ok {
		return errors.New("while casting input to update input")
	}

	updateMutationFormat := `mutation {
		result: updateBundleInstanceAuth(
			id: "%s"
			bundleID: "%s"
			in: {
			auth: {
				credential: { %s }
			}
		}) {
			id
		}
	}`
	for _, instanceAuth := range updateInput.Bundle.InstanceAuths {
		if *instanceAuth.RuntimeID == updateInput.RuntimeID {
			gqlReq := gcli.NewRequest(fmt.Sprintf(updateMutationFormat, instanceAuth.ID, updateInput.Bundle.ID, updateInput.Credentials.ToString()))

			if err := c.runWithTenant(ctx, gqlReq, tenant, nil); err != nil {
				errMsg := fmt.Sprintf("Error while updating bundle instance auth with id %q for bundle with id %q", instanceAuth.ID, updateInput.Bundle.ID)
				log.C(ctx).WithError(err).Error(errMsg)
				return errors.New(errMsg)
			}
		}
	}

	return nil
}

// DeleteBundleInstanceAuth deletes bundle instance auth with a given input
func (c *client) DeleteBundleInstanceAuth(ctx context.Context, tenant string, input BundleInstanceAuthInput) error {
	deleteInput, ok := input.(DeleteBundleInstanceAuthInput)
	if !ok {
		return errors.New("while casting input to delete input")
	}

	deleteMutationFormat := `mutation {
		result: deleteBundleInstanceAuth(
			authID: "%s"
		) {
    		id
		  }
		}`
	for _, instanceAuth := range deleteInput.Bundle.InstanceAuths {
		if *instanceAuth.RuntimeID == deleteInput.RuntimeID {
			gqlReq := gcli.NewRequest(fmt.Sprintf(deleteMutationFormat, instanceAuth.ID))

			if err := c.runWithTenant(ctx, gqlReq, tenant, nil); err != nil {
				errMsg := fmt.Sprintf("Error while deleting bundle instance auth with id %q", instanceAuth.ID)
				log.C(ctx).WithError(err).Error(errMsg)
				return errors.New(errMsg)
			}
		}
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
