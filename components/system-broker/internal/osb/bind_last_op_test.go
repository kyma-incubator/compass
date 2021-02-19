package osb_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/types/typesfakes"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestBindLastOp(t *testing.T) {
	instanceID := "instanceID"
	bindingID := "bindingID"

	var (
		fakeCredentialsGetter *typesfakes.FakeBundleCredentialsFetcher
		be                    *osb.BindLastOperationEndpoint
		bundleInstanceAuth    *director.BundleInstanceAuthOutput
		details               domain.PollDetails
	)

	setup := func() {
		fakeCredentialsGetter = &typesfakes.FakeBundleCredentialsFetcher{}

		be = osb.NewBindLastOperationEndpoint(fakeCredentialsGetter)

		bundleInstanceAuth = &director.BundleInstanceAuthOutput{
			InstanceAuth: &graphql.BundleInstanceAuth{
				ID: "instanceAuthID",
				Status: &graphql.BundleInstanceAuthStatus{
					Condition: graphql.BundleInstanceAuthStatusConditionSucceeded,
					Message:   "success",
				},
			},
		}

		details = domain.PollDetails{
			ServiceID:     "serviceID",
			PlanID:        "planID",
			OperationData: string(osb.BindOp),
		}
	}

	t.Run("Success", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		lastOp, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceAuthCallCount())
		assert.Equal(t, "success", lastOp.Description)
		assert.Equal(t, domain.Succeeded, lastOp.State)
	})

	t.Run("When credentials are pending", func(t *testing.T) {
		setup()
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.BundleInstanceAuthStatusConditionPending
		bundleInstanceAuth.InstanceAuth.Status.Message = "pending"
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		lastOp, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceAuthCallCount())
		assert.Equal(t, "pending", lastOp.Description)
		assert.Equal(t, domain.InProgress, lastOp.State)
	})

	t.Run("When credentials are failed", func(t *testing.T) {
		setup()
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.BundleInstanceAuthStatusConditionFailed
		bundleInstanceAuth.InstanceAuth.Status.Message = "failed"
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		lastOp, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceAuthCallCount())
		assert.Equal(t, "failed", lastOp.Description)
		assert.Equal(t, domain.Failed, lastOp.State)
	})

	t.Run("When credentials are unused", func(t *testing.T) {
		setup()
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.BundleInstanceAuthStatusConditionUnused
		bundleInstanceAuth.InstanceAuth.Status.Message = "unused"
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		_, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation reached unexpected state: op bind_operation, status")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceAuthCallCount())
	})

	t.Run("When no bundle instance auth exists", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			nil,
			&NotFoundErr{},
		)
		_, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "binding cannot be fetched")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceAuthCallCount())
	})

	t.Run("When credentials getter returns an error", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			nil,
			errors.New("some error"),
		)
		_, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while getting bundle instance credentials from director")
	})

	t.Run("When no bundle instance auth exists for unbind", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			nil,
			&NotFoundErr{},
		)
		details.OperationData = string(osb.UnbindOp)
		lastOp, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceAuthCallCount())
		assert.Equal(t, domain.Succeeded, lastOp.State)
	})

	t.Run("When bundle instance auth is still existing for unbind", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		details.OperationData = string(osb.UnbindOp)
		lastOp, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceAuthCallCount())
		assert.Equal(t, domain.InProgress, lastOp.State)
	})
}
