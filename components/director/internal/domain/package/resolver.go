package mp_package

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/package/mock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type Resolver struct{}

func NewResolver() *Resolver {
	return &Resolver{}
}

// TODO: Replace with real implementation
func (r *Resolver) AddPackage(ctx context.Context, applicationID string, in graphql.PackageCreateInput) (*graphql.Package, error) {
	return mock.FixPackage(), nil
}

// TODO: Replace with real implementation
func (r *Resolver) UpdatePackage(ctx context.Context, id string, in graphql.PackageUpdateInput) (*graphql.Package, error) {
	return mock.FixPackage(), nil
}

// TODO: Replace with real implementation
func (r *Resolver) DeletePackage(ctx context.Context, id string) (*graphql.Package, error) {
	return mock.FixPackage(), nil
}

// TODO: Replace with real implementation
func (r *Resolver) InstanceAuth(ctx context.Context, obj *graphql.Package, id string) (*graphql.PackageInstanceAuth, error) {
	var condition graphql.PackageInstanceAuthStatusCondition
	switch id {
	case "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb":
		condition = graphql.PackageInstanceAuthStatusConditionSucceeded
	case "cccccccc-cccc-cccc-cccc-cccccccccccc":
		condition = graphql.PackageInstanceAuthStatusConditionFailed
	default:
		condition = graphql.PackageInstanceAuthStatusConditionPending
	}

	return mock.FixPackageInstanceAuth(id, condition), nil
}

// TODO: Replace with real implementation
func (r *Resolver) InstanceAuths(ctx context.Context, obj *graphql.Package) ([]*graphql.PackageInstanceAuth, error) {
	return []*graphql.PackageInstanceAuth{
		mock.FixPackageInstanceAuth("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", graphql.PackageInstanceAuthStatusConditionPending),
		mock.FixPackageInstanceAuth("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", graphql.PackageInstanceAuthStatusConditionSucceeded),
		mock.FixPackageInstanceAuth("cccccccc-cccc-cccc-cccc-cccccccccccc", graphql.PackageInstanceAuthStatusConditionFailed),
	}, nil
}

var packageID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

// TODO: Replace with real implementation
func (r *Resolver) APIDefinitions(ctx context.Context, obj *graphql.Package, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
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
func (r *Resolver) EventDefinitions(ctx context.Context, obj *graphql.Package, group *string, first *int, after *graphql.PageCursor) (*graphql.EventDefinitionPage, error) {
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
func (r *Resolver) Documents(ctx context.Context, obj *graphql.Package, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
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
func (r *Resolver) APIDefinition(ctx context.Context, obj *graphql.Package, id string) (*graphql.APIDefinition, error) {
	return mock.FixAPIDefinition(id, packageID), nil
}

// TODO: Replace with real implementation
func (r *Resolver) EventDefinition(ctx context.Context, obj *graphql.Package, id string) (*graphql.EventDefinition, error) {
	return mock.FixEventDefinition(id, packageID), nil
}

// TODO: Replace with real implementation
func (r *Resolver) Document(ctx context.Context, obj *graphql.Package, id string) (*graphql.Document, error) {
	return mock.FixDocument(id, packageID), nil
}
