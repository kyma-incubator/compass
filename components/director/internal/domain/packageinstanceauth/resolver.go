package packageinstanceauth

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/package/mock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type Resolver struct{}

func NewResolver() *Resolver {
	return &Resolver{}
}

var mockRequestTypeKey = "type"

// TODO: Replace with real implementation
func (r *Resolver) SetPackageInstanceAuth(ctx context.Context, packageID string, authID string, in graphql.AuthInput) (*graphql.PackageInstanceAuth, error) {
	return mock.FixPackageInstanceAuth(packageID, graphql.PackageInstanceAuthStatusConditionSucceeded), nil
}

// TODO: Replace with real implementation
func (r *Resolver) DeletePackageInstanceAuth(ctx context.Context, packageID string, authID string) (*graphql.PackageInstanceAuth, error) {
	return mock.FixPackageInstanceAuth(packageID, graphql.PackageInstanceAuthStatusConditionUnused), nil
}

// TODO: Replace with real implementation
func (r *Resolver) RequestPackageInstanceAuthCreation(ctx context.Context, packageID string, in graphql.PackageInstanceAuthRequestInput) (*graphql.PackageInstanceAuth, error) {
	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	if in.Context == nil {
		return mock.FixPackageInstanceAuth(id, graphql.PackageInstanceAuthStatusConditionPending), nil
	}

	data, ok := (*in.Context).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid context type: expected map[string]interface{}, actual %T", *in.Context)
	}

	if _, exists := data[mockRequestTypeKey]; !exists {
		return mock.FixPackageInstanceAuth(id, graphql.PackageInstanceAuthStatusConditionPending), nil
	}

	reqType, ok := data[mockRequestTypeKey].(string)
	if !ok {
		return nil, errors.New("invalid mock request type: expected string value (`success` or `error`)")
	}

	switch reqType {
	case "success":
		id = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	case "error":
		id = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	}

	return mock.FixPackageInstanceAuth(id, graphql.PackageInstanceAuthStatusConditionPending), nil
}

// TODO: Replace with real implementation
func (r *Resolver) RequestPackageInstanceAuthDeletion(ctx context.Context, packageID string, authID string) (*graphql.PackageInstanceAuth, error) {
	return mock.FixPackageInstanceAuth(packageID, graphql.PackageInstanceAuthStatusConditionUnused), nil
}
