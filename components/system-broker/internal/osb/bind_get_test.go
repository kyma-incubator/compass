package osb_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/types/typesfakes"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/director"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestBindGet(t *testing.T) {
	instanceID := "instanceID"
	bindingID := "bindingID"

	var (
		fakeCredentialsGetter *typesfakes.FakeBundleCredentialsFetcherForInstance
		be                    *osb.GetBindingEndpoint
		bundleInstanceAuth    *director.BundleInstanceCredentialsOutput
	)

	setup := func() {
		fakeCredentialsGetter = &typesfakes.FakeBundleCredentialsFetcherForInstance{}

		be = osb.NewGetBindingEndpoint(fakeCredentialsGetter)

		bundleInstanceAuth = &director.BundleInstanceCredentialsOutput{
			InstanceAuth: &graphql.BundleInstanceAuth{
				Status: &graphql.BundleInstanceAuthStatus{
					Condition: graphql.BundleInstanceAuthStatusConditionSucceeded,
				},
				Auth: &graphql.Auth{
					Credential: &graphql.BasicCredentialData{
						Username: "username",
						Password: "password",
					},
				},
			},
			TargetURLs: map[string]string{},
		}
	}

	t.Run("Success", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceCredentialsReturns(
			bundleInstanceAuth,
			nil,
		)
		binding, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceCredentialsCallCount())
		creds, ok := binding.Credentials.(osb.BindingCredentials)
		assert.True(t, ok)
		basicCreds := creds.AuthDetails.Credentials.ToCredentials().BasicAuth
		assert.Equal(t, "username", basicCreds.Username)
		assert.Equal(t, "password", basicCreds.Password)
	})

	t.Run("When credentials mapper returns an error", func(t *testing.T) {
		setup()
		bundleInstanceAuth.InstanceAuth.Auth.Credential = graphql.BasicCredentialData{
			Username: "username",
			Password: "password",
		}
		fakeCredentialsGetter.FetchBundleInstanceCredentialsReturns(
			bundleInstanceAuth,
			nil,
		)
		_, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while mapping to binding credentials: got unknown credential type")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceCredentialsCallCount())
	})

	t.Run("When credentials getter returns an error", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceCredentialsReturns(
			nil,
			errors.New("some-error"),
		)
		_, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while getting bundle instance credentials from director")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceCredentialsCallCount())
	})

	t.Run("When credentials getter returns not found", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchBundleInstanceCredentialsReturns(
			nil,
			&NotFoundErr{},
		)
		_, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "binding cannot be fetched")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceCredentialsCallCount())
	})

	t.Run("When credentials getter is in pending state", func(t *testing.T) {
		setup()
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.BundleInstanceAuthStatusConditionPending
		fakeCredentialsGetter.FetchBundleInstanceCredentialsReturns(
			bundleInstanceAuth,
			nil,
		)
		_, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "binding cannot be fetched")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceCredentialsCallCount())
	})

	t.Run("When credentials getter is in unused state", func(t *testing.T) {
		setup()
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.BundleInstanceAuthStatusConditionUnused
		fakeCredentialsGetter.FetchBundleInstanceCredentialsReturns(
			bundleInstanceAuth,
			nil,
		)
		_, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "binding cannot be fetched")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceCredentialsCallCount())
	})

	t.Run("When credentials getter is in failed state", func(t *testing.T) {
		setup()
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.BundleInstanceAuthStatusConditionFailed
		fakeCredentialsGetter.FetchBundleInstanceCredentialsReturns(
			bundleInstanceAuth,
			nil,
		)
		_, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credentials status is not success")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchBundleInstanceCredentialsCallCount())
	})
}
