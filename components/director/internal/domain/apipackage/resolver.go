package apipackage

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apipackage/mock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type Resolver struct{}

func NewResolver() *Resolver {
	return &Resolver{}
}

// TODO: Replace with real implementation
func (r *Resolver) AddPackageDefinition(ctx context.Context, applicationID string, in graphql.PackageDefinitionCreateInput) (*graphql.PackageDefinition, error) {
	return mock.FixPackage(), nil
}

// TODO: Replace with real implementation
func (r *Resolver) UpdatePackageDefinition(ctx context.Context, id string, in graphql.PackageDefinitionUpdateInput) (*graphql.PackageDefinition, error) {
	return mock.FixPackage(), nil
}

// TODO: Replace with real implementation
func (r *Resolver) DeletePackageDefinition(ctx context.Context, id string) (*graphql.PackageDefinition, error) {
	return mock.FixPackage(), nil
}

// TODO: Replace with real implementation
func (r *Resolver) Auth(ctx context.Context, obj *graphql.PackageDefinition, id string) (*graphql.APIInstanceAuth, error) {
	var condition graphql.APIInstanceAuthStatusCondition
	switch id {
	case "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb":
		condition = graphql.APIInstanceAuthStatusConditionSucceeded
	case "cccccccc-cccc-cccc-cccc-cccccccccccc":
		condition = graphql.APIInstanceAuthStatusConditionFailed
	default:
		condition = graphql.APIInstanceAuthStatusConditionPending
	}

	return mock.FixAPIInstanceAuth(id, condition), nil
}

// TODO: Replace with real implementation
func (r *Resolver) Auths(ctx context.Context, obj *graphql.PackageDefinition) ([]*graphql.APIInstanceAuth, error) {
	return []*graphql.APIInstanceAuth{
		mock.FixAPIInstanceAuth("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", graphql.APIInstanceAuthStatusConditionPending),
		mock.FixAPIInstanceAuth("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", graphql.APIInstanceAuthStatusConditionSucceeded),
		mock.FixAPIInstanceAuth("cccccccc-cccc-cccc-cccc-cccccccccccc", graphql.APIInstanceAuthStatusConditionFailed),
	}, nil
}

var packageID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

// TODO: Replace with real implementation
func (r *Resolver) APIDefinitions(ctx context.Context, obj *graphql.PackageDefinition, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
	return &graphql.APIDefinitionPage{
		Data: []*graphql.APIDefinition{
			mock.FixAPIDefinition("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", obj.ID),
			mock.FixAPIDefinition("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", obj.ID),
			mock.FixAPIDefinition("cccccccc-cccc-cccc-cccc-cccccccccccc", obj.ID),
		},
		TotalCount: 3,
	}, nil
}

// TODO: Replace with real implementation
func (r *Resolver) EventDefinitions(ctx context.Context, obj *graphql.PackageDefinition, group *string, first *int, after *graphql.PageCursor) (*graphql.EventDefinitionPage, error) {
	return &graphql.EventDefinitionPage{
		Data: []*graphql.EventDefinition{
			mock.FixEventDefinition("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", obj.ID),
			mock.FixEventDefinition("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", obj.ID),
			mock.FixEventDefinition("cccccccc-cccc-cccc-cccc-cccccccccccc", obj.ID),
		},
		TotalCount: 3,
	}, nil
}

// TODO: Replace with real implementation
func (r *Resolver) Documents(ctx context.Context, obj *graphql.PackageDefinition, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	return &graphql.DocumentPage{
		Data: []*graphql.Document{
			mock.FixDocument("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", obj.ID),
			mock.FixDocument("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", obj.ID),
			mock.FixDocument("cccccccc-cccc-cccc-cccc-cccccccccccc", obj.ID),
		},
		TotalCount: 3,
	}, nil
}

// TODO: Replace with real implementation
func (r *Resolver) APIDefinition(ctx context.Context, obj *graphql.PackageDefinition, id string) (*graphql.APIDefinition, error) {
	return mock.FixAPIDefinition(id, packageID), nil
}

// TODO: Replace with real implementation
func (r *Resolver) EventDefinition(ctx context.Context, obj *graphql.PackageDefinition, id string) (*graphql.EventDefinition, error) {
	return mock.FixEventDefinition(id, packageID), nil
}

// TODO: Replace with real implementation
func (r *Resolver) Document(ctx context.Context, obj *graphql.PackageDefinition, id string) (*graphql.Document, error) {
	return mock.FixDocument(id, packageID), nil
}
