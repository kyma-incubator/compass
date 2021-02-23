package osb_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/types/typesfakes"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/director"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestDeleteBinding(t *testing.T) {
	instanceID := "instanceID"
	bindingID := "bindingID"

	var (
		fakeCredentialsDeleter     *typesfakes.FakeBundleCredentialsDeleteRequester
		fakeCredentialsGetter      *typesfakes.FakeBundleCredentialsFetcher
		be                         *osb.UnbindEndpoint
		details                    domain.UnbindDetails
		bundleInstanceAuth         *director.BundleInstanceAuthOutput
		bundleInstanceAuthDeletion *director.BundleInstanceAuthDeletionOutput
	)

	setup := func() {
		fakeCredentialsDeleter = &typesfakes.FakeBundleCredentialsDeleteRequester{}
		fakeCredentialsGetter = &typesfakes.FakeBundleCredentialsFetcher{}

		be = osb.NewUnbindEndpoint(fakeCredentialsGetter, fakeCredentialsDeleter)

		details = domain.UnbindDetails{
			ServiceID: "serviceID",
			PlanID:    "planID",
		}

		bundleInstanceAuth = &director.BundleInstanceAuthOutput{
			InstanceAuth: &graphql.BundleInstanceAuth{
				ID: "instanceAuthID",
				Status: &graphql.BundleInstanceAuthStatus{
					Condition: graphql.BundleInstanceAuthStatusConditionSucceeded,
				},
			},
		}

		bundleInstanceAuthDeletion = &director.BundleInstanceAuthDeletionOutput{
			ID: "instanceAuthID",
		}
	}

	t.Run("Success", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		fakeCredentialsDeleter.RequestBundleInstanceCredentialsDeletionReturns(
			bundleInstanceAuthDeletion,
			nil,
		)
		binding, err := be.Unbind(context.TODO(), instanceID, bindingID, details, true)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceAuthCallCount())
		assert.Equal(t, 1, fakeCredentialsDeleter.RequestBundleInstanceCredentialsDeletionCallCount())
		assert.Equal(t, "unbind_operation", binding.OperationData)
	})

	t.Run("When credentials getter returns an error", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			nil,
			errors.New("some error"),
		)
		_, err := be.Unbind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while getting bundle instance credentials from director")
	})

	t.Run("When credentials are not found", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			nil,
			&NotFoundErr{},
		)
		_, err := be.Unbind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "binding does not exist")
	})

	t.Run("When credentials are already being deleted", func(t *testing.T) {
		setup()
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.BundleInstanceAuthStatusConditionUnused
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		_, err := be.Unbind(context.TODO(), instanceID, bindingID, details, true)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceAuthCallCount())
		assert.Equal(t, 0, fakeCredentialsDeleter.RequestBundleInstanceCredentialsDeletionCallCount())
	})

	t.Run("When credentials deleter returns not found", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		fakeCredentialsDeleter.RequestBundleInstanceCredentialsDeletionReturns(
			nil,
			&NotFoundErr{},
		)
		_, err := be.Unbind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "binding does not exist")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceAuthCallCount())
		assert.Equal(t, 1, fakeCredentialsDeleter.RequestBundleInstanceCredentialsDeletionCallCount())
	})

	t.Run("When credentials deleter returns error", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		fakeCredentialsDeleter.RequestBundleInstanceCredentialsDeletionReturns(
			nil,
			errors.New("some error"),
		)
		_, err := be.Unbind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while requesting bundle instance credentials deletion from director")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceAuthCallCount())
		assert.Equal(t, 1, fakeCredentialsDeleter.RequestBundleInstanceCredentialsDeletionCallCount())
	})

	t.Run("When async is not supported by platform", func(t *testing.T) {
		setup()
		_, err := be.Unbind(context.TODO(), instanceID, bindingID, details, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "This service plan requires client support for asynchronous service operations")
	})
}
