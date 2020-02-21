package packageinstanceauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/package/mock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=Service -output=automock -outpkg=automock -case=underscore
type Service interface {
	RequestDeletion(ctx context.Context, instanceAuth *model.PackageInstanceAuth, defaultPackageInstanceAuth *model.Auth) (bool, error)
	Get(ctx context.Context, id string) (*model.PackageInstanceAuth, error)
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=PackageService -output=automock -outpkg=automock -case=underscore
type PackageService interface {
	GetByInstanceAuthID(ctx context.Context, instanceAuthID string) (*model.Package, error)
}

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToGraphQL(in *model.PackageInstanceAuth) *graphql.PackageInstanceAuth
}

type Resolver struct {
	transact   persistence.Transactioner
	svc        Service
	packageSvc PackageService
	conv       Converter
}

func NewResolver(transact persistence.Transactioner, svc Service, packageSvc PackageService, conv Converter) *Resolver {
	return &Resolver{
		transact:   transact,
		svc:        svc,
		packageSvc: packageSvc,
		conv:       conv,
	}
}

var mockRequestTypeKey = "type"
var mockPackageID = "db5d3b2a-cf30-498b-9a66-29e60247c66b"

// TODO: Remove mock
func (r *Resolver) SetPackageInstanceAuthMock(ctx context.Context, authID string, in graphql.PackageInstanceAuthSetInput) (*graphql.PackageInstanceAuth, error) {
	return mock.FixPackageInstanceAuth(mockPackageID, graphql.PackageInstanceAuthStatusConditionSucceeded), nil
}

// TODO: Remove mock
func (r *Resolver) DeletePackageInstanceAuthMock(ctx context.Context, authID string) (*graphql.PackageInstanceAuth, error) {
	return mock.FixPackageInstanceAuth(mockPackageID, graphql.PackageInstanceAuthStatusConditionUnused), nil
}

// TODO: Remove mock
func (r *Resolver) RequestPackageInstanceAuthCreationMock(ctx context.Context, packageID string, in graphql.PackageInstanceAuthRequestInput) (*graphql.PackageInstanceAuth, error) {
	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	if in.Context == nil {
		return mock.FixPackageInstanceAuth(id, graphql.PackageInstanceAuthStatusConditionPending), nil
	}

	var data map[string]interface{}
	err := json.Unmarshal([]byte(*in.Context), &data)
	if err != nil {
		return nil, fmt.Errorf("invalid context type: expected JSON object, actual %+v", *in.Context)
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

	out := mock.FixPackageInstanceAuth(id, graphql.PackageInstanceAuthStatusConditionPending)

	out.Context = in.Context
	out.InputParams = in.InputParams

	return out, nil
}

// TODO: Remove mock
func (r *Resolver) RequestPackageInstanceAuthDeletionMock(ctx context.Context, authID string) (*graphql.PackageInstanceAuth, error) {
	return mock.FixPackageInstanceAuth(mockPackageID, graphql.PackageInstanceAuthStatusConditionUnused), nil
}

func (r *Resolver) DeletePackageInstanceAuth(ctx context.Context, authID string) (*graphql.PackageInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	instanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		return nil, err
	}

	err = r.svc.Delete(ctx, authID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth), nil
}

// TODO: Replace with real implementation
func (r *Resolver) SetPackageInstanceAuth(ctx context.Context, authID string, in graphql.PackageInstanceAuthSetInput) (*graphql.PackageInstanceAuth, error) {
	panic("not implemented")
}

// TODO: Replace with real implementation
func (r *Resolver) RequestPackageInstanceAuthCreation(ctx context.Context, packageID string, in graphql.PackageInstanceAuthRequestInput) (*graphql.PackageInstanceAuth, error) {
	panic("not implemented")
}

func (r *Resolver) RequestPackageInstanceAuthDeletion(ctx context.Context, authID string) (*graphql.PackageInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	instanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		return nil, err
	}

	pkg, err := r.packageSvc.GetByInstanceAuthID(ctx, authID)
	if err != nil {
		return nil, err
	}

	deleted, err := r.svc.RequestDeletion(ctx, instanceAuth, pkg.DefaultInstanceAuth)
	if err != nil {
		return nil, err
	}

	if !deleted {
		instanceAuth, err = r.svc.Get(ctx, authID) // get InstanceAuth once again for new status
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth), nil
}
