package apiinstanceauth

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apipackage/mock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type Resolver struct{}

func NewResolver() *Resolver {
	return &Resolver{}
}

var mockRequestTypeKey = "type"

// TODO: Replace with real implementation
func (r *Resolver) SetAPIInstanceAuthForPackage(ctx context.Context, packageID string, authID string, in graphql.AuthInput) (*graphql.APIInstanceAuth, error) {
	return mock.FixAPIInstanceAuth(packageID, graphql.APIInstanceAuthStatusConditionSucceeded), nil
}

// TODO: Replace with real implementation
func (r *Resolver) DeleteAPIInstanceAuthForPackage(ctx context.Context, packageID string, authID string) (*graphql.APIInstanceAuth, error) {
	return mock.FixAPIInstanceAuth(packageID, graphql.APIInstanceAuthStatusConditionPending), nil
}

// TODO: Replace with real implementation
func (r *Resolver) RequestAPIInstanceAuthForPackage(ctx context.Context, packageID string, in graphql.APIInstanceAuthRequestInput) (*graphql.APIInstanceAuth, error) {
	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	if in.Context == nil {
		return mock.FixAPIInstanceAuth(id, graphql.APIInstanceAuthStatusConditionPending), nil
	}

	data, ok := (*in.Context).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid context type: expected map[string]interface{}, actual %T", *in.Context)
	}

	if _, exists := data[mockRequestTypeKey]; !exists {
		return mock.FixAPIInstanceAuth(id, graphql.APIInstanceAuthStatusConditionPending), nil
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

	return mock.FixAPIInstanceAuth(id, graphql.APIInstanceAuthStatusConditionPending), nil
}
