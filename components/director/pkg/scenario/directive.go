package scenario

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"

	mp_bundle "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/99designs/gqlgen/graphql"
)

const (
	GetApplicationID                     = "GetApplicationID"
	GetApplicationIDByBundle             = "GetApplicationIDByBundle"
	GetApplicationIDByBundleInstanceAuth = "GetApplicationIDByBundleInstanceAuth"
)

var ErrMissingScenario = errors.New("Forbidden: Missing scenarios")

//go:generate mockery --name=ScenariosService --output=automock --outpkg=automock --case=underscore
type ScenariosService interface {
	GetScenarioNamesForApplication(ctx context.Context, applicationID string) ([]string, error)
	GetScenarioNamesForRuntime(ctx context.Context, runtimeId string) ([]string, error)
}

type directive struct {
	scenariosService ScenariosService
	transact         persistence.Transactioner

	applicationProviders map[string]func(context.Context, string, string) (string, error)
}

// NewDirective returns a new scenario directive
func NewDirective(transact persistence.Transactioner, scenariosService ScenariosService, bundleRepo mp_bundle.BundleRepository, bundleInstanceAuthRepo bundleinstanceauth.Repository) *directive {
	getApplicationIDByBundleFunc := func(ctx context.Context, tenantID, bundleID string) (string, error) {
		bndl, err := bundleRepo.GetByID(ctx, tenantID, bundleID)
		if err != nil {
			return "", errors.Wrapf(err, "while getting Bundle with id %s", bundleID)
		}
		return bndl.ApplicationID, nil
	}

	return &directive{
		transact:         transact,
		scenariosService: scenariosService,
		applicationProviders: map[string]func(context.Context, string, string) (string, error){
			GetApplicationID: func(ctx context.Context, tenantID string, appID string) (string, error) {
				return appID, nil
			},
			GetApplicationIDByBundle: getApplicationIDByBundleFunc,
			GetApplicationIDByBundleInstanceAuth: func(ctx context.Context, tenantID, bundleInstanceAuthID string) (string, error) {
				bundleInstanceAuth, err := bundleInstanceAuthRepo.GetByID(ctx, tenantID, bundleInstanceAuthID)
				if err != nil {
					return "", errors.Wrapf(err, "while getting Bundle instance auth with id %s", bundleInstanceAuthID)
				}

				return getApplicationIDByBundleFunc(ctx, tenantID, bundleInstanceAuth.BundleID)
			},
		},
	}
}

// HasScenario ensures that the runtime is in a scenario with the application which resources are being manipulated.
// If the caller is not a Runtime, then request is forwarded to the next resolver.
func (d *directive) HasScenario(ctx context.Context, _ interface{}, next graphql.Resolver, applicationProvider string, idField string) (res interface{}, err error) {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if consumerInfo.ConsumerType != consumer.Runtime {
		log.C(ctx).Debugf("Consumer type %v is not of type %v. Skipping verification directive...", consumerInfo.ConsumerType, consumer.Runtime)
		return next(ctx)
	}
	log.C(ctx).Infof("Attempting to verify that the requesting runtime is in scenario with the owning application entity")

	runtimeID := consumerInfo.ConsumerID
	log.C(ctx).Debugf("Found Runtime ID for the requesting runtime: %v", runtimeID)

	commonScenarios, err := d.extractCommonScenarios(ctx, runtimeID, applicationProvider, idField)
	if err != nil {
		return nil, err
	}

	if len(commonScenarios) == 0 {
		return nil, apperrors.NewInvalidOperationError("requesting runtime should be in same scenario as the requested application resource")
	}
	log.C(ctx).Debugf("Found the following common scenarios: %+v", commonScenarios)

	log.C(ctx).Infof("Runtime with ID %s is in scenario with the owning application entity", runtimeID)
	return next(ctx)
}

func (d *directive) extractCommonScenarios(ctx context.Context, runtimeID, applicationProvider, idField string) ([]string, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	resCtx := graphql.GetResolverContext(ctx)
	id, ok := resCtx.Args[idField].(string)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Could not get idField: %s from request context", idField))
	}

	appProviderFunc, ok := d.applicationProviders[applicationProvider]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Could not get app provider func: %s from provider list", applicationProvider))
	}

	tx, err := d.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while opening the db transaction: %v", err)
		return nil, err
	}
	defer d.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	appID, err := appProviderFunc(ctx, tenantID, id)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not derive app id, an error occurred")
	}
	log.C(ctx).Infof("Found owning Application ID based on the request parameter %s: %s", idField, appID)

	appScenarios, err := d.scenariosService.GetScenarioNamesForApplication(ctx, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching scenarios for application")
	}
	log.C(ctx).Debugf("Found the following application scenarios: %s", appScenarios)

	runtimeScenarios, err := d.scenariosService.GetScenarioNamesForRuntime(ctx, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching scenarios for runtime")
	}
	log.C(ctx).Debugf("Found the following runtime scenarios: %s", runtimeScenarios)

	if err := tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while committing transaction: %v", err)
		return nil, err
	}

	commonScenarios := stringsIntersection(appScenarios, runtimeScenarios)
	return commonScenarios, nil
}

// stringsIntersection returns the common elements in two string slices.
func stringsIntersection(str1, str2 []string) []string {
	var intersection []string
	strings := make(map[string]bool)
	for _, v := range str1 {
		strings[v] = true
	}
	for _, v := range str2 {
		if strings[v] {
			intersection = append(intersection, v)
		}
	}
	return intersection
}
