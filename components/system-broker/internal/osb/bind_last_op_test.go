package osb

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb/osbfakes"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestBindLastOp(t *testing.T) {
	instanceID := "instanceID"
	bindingID := "bindingID"

	var (
		fakeCredentialsGetter *osbfakes.FakePackageCredentialsFetcher
		be                    *BindLastOperationEndpoint
		bundleInstanceAuth    *director.BundleInstanceAuthOutput
		details               domain.PollDetails
	)

	setup := func() {
		fakeCredentialsGetter = &osbfakes.FakePackageCredentialsFetcher{}

		be = &BindLastOperationEndpoint{
			credentialsGetter: fakeCredentialsGetter,
		}

		bundleInstanceAuth = &director.BundleInstanceAuthOutput{
			InstanceAuth: &graphql.PackageInstanceAuth{
				ID: "instanceAuthID",
				Status: &graphql.PackageInstanceAuthStatus{
					Condition: graphql.PackageInstanceAuthStatusConditionSucceeded,
					Message:   "success",
				},
			},
		}

		details = domain.PollDetails{
			ServiceID:     "serviceID",
			PlanID:        "planID",
			OperationData: string(BindOp),
		}
	}

	t.Run("Success", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		lastOp, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
		assert.Equal(t, "success", lastOp.Description)
		assert.Equal(t, domain.Succeeded, lastOp.State)
	})

	t.Run("When credentials are pending", func(t *testing.T) {
		setup()
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.PackageInstanceAuthStatusConditionPending
		bundleInstanceAuth.InstanceAuth.Status.Message = "pending"
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		lastOp, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
		assert.Equal(t, "pending", lastOp.Description)
		assert.Equal(t, domain.InProgress, lastOp.State)
	})

	t.Run("When credentials are failed", func(t *testing.T) {
		setup()
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.PackageInstanceAuthStatusConditionFailed
		bundleInstanceAuth.InstanceAuth.Status.Message = "failed"
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		lastOp, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
		assert.Equal(t, "failed", lastOp.Description)
		assert.Equal(t, domain.Failed, lastOp.State)
	})

	t.Run("When credentials are unused", func(t *testing.T) {
		setup()
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.PackageInstanceAuthStatusConditionUnused
		bundleInstanceAuth.InstanceAuth.Status.Message = "unused"
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		_, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation reached unexpected state: op bind_operation, status")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
	})

	t.Run("When no package instance auth exists", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			nil,
			&notFoundErr{},
		)
		_, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "binding cannot be fetched")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
	})

	t.Run("When credentials getter returns an error", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			nil,
			errors.New("some error"),
		)
		_, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while getting package instance credentials from director")
	})

	t.Run("When no package instance auth exists for unbind", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			nil,
			&notFoundErr{},
		)
		details.OperationData = string(UnbindOp)
		lastOp, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
		assert.Equal(t, domain.Succeeded, lastOp.State)
	})

	t.Run("When package instance auth is still existing for unbind", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		details.OperationData = string(UnbindOp)
		lastOp, err := be.LastBindingOperation(context.TODO(), instanceID, bindingID, details)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
		assert.Equal(t, domain.InProgress, lastOp.State)
	})
}
