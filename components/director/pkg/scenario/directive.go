package scenario

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/99designs/gqlgen/graphql"
)

const (
	GetApplicationID                      = "GetApplicationID"
	GetApplicationIDByPackage             = "GetApplicationIDByPackage"
	GetApplicationIDByPackageInstanceAuth = "GetApplicationIDByPackageInstanceAuth"
)

var ErrMissingScenario = errors.New("Forbidden: Missing scenarios")

type directive struct {
	labelRepo label.LabelRepository
	transact  persistence.Transactioner

	applicationProviders map[string]func(context.Context, string, string) (string, error)
}

// NewDirective returns a new scenario directive
func NewDirective(transact persistence.Transactioner, labelRepo label.LabelRepository, packageRepo mp_package.PackageRepository, packageInstanceAuthRepo packageinstanceauth.Repository) *directive {
	getApplicationIDByPackageFunc := func(ctx context.Context, tenantID, packageID string) (string, error) {
		pkg, err := packageRepo.GetByID(ctx, tenantID, packageID)
		if err != nil {
			return "", errors.Wrapf(err, "while getting Package with id %s", packageID)
		}
		return pkg.ApplicationID, nil
	}

	return &directive{
		transact:  transact,
		labelRepo: labelRepo,
		applicationProviders: map[string]func(context.Context, string, string) (string, error){
			GetApplicationID: func(ctx context.Context, tenantID string, appID string) (string, error) {
				return appID, nil
			},
			GetApplicationIDByPackage: getApplicationIDByPackageFunc,
			GetApplicationIDByPackageInstanceAuth: func(ctx context.Context, tenantID, packageInstanceAuthID string) (string, error) {
				packageInstanceAuth, err := packageInstanceAuthRepo.GetByID(ctx, tenantID, packageInstanceAuthID)
				if err != nil {
					return "", errors.Wrapf(err, "while getting Package instance auth with id %s", packageInstanceAuthID)
				}

				return getApplicationIDByPackageFunc(ctx, tenantID, packageInstanceAuth.PackageID)
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
		log.Debugf("Consumer type %v is not of type %v. Skipping verification directive...", consumerInfo.ConsumerType, consumer.Runtime)
		return next(ctx)
	}
	log.Infof("Attempting to verify that the requesting runtime is in scenario with the owning application entity")

	runtimeID := consumerInfo.ConsumerID
	log.Debugf("Found Runtime ID for the requesting runtime: %v", runtimeID)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	commonScenarios, err := d.extractCommonScenarios(ctx, tenantID, runtimeID, applicationProvider, idField)
	if len(commonScenarios) == 0 {
		return nil, ErrMissingScenario
	}
	log.Debugf("Found the following common scenarios: %+v", commonScenarios)

	log.Infof("Runtime with ID %s is in scenario with the owning application entity", runtimeID)
	return next(ctx)
}

func (d *directive) extractCommonScenarios(ctx context.Context, tenantID, runtimeID, applicationProvider, idField string) ([]string, error) {
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
		log.Errorf("An error occurred while opening the db transaction: %s", err.Error())
		return nil, err
	}
	defer d.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	appID, err := appProviderFunc(ctx, tenantID, id)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not derive app id, an error occurred")
	}
	log.Infof("Found owning Application ID based on the request parameter %s: %s", idField, appID)

	appScenarios, err := d.getObjectScenarios(ctx, tenantID, model.ApplicationLabelableObject, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching scenarios for application")
	}
	log.Debugf("Found the following application scenarios: %s", appScenarios)

	runtimeScenarios, err := d.getObjectScenarios(ctx, tenantID, model.RuntimeLabelableObject, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching scenarios for runtime")
	}
	log.Debugf("Found the following runtime scenarios: %s", runtimeScenarios)

	if err := tx.Commit(); err != nil {
		log.Errorf("An error occurred while committing transaction: %s", err.Error())
		return nil, err
	}

	commonScenarios := stringsIntersection(appScenarios, runtimeScenarios)
	return commonScenarios, nil
}

func (d *directive) getObjectScenarios(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID string) ([]string, error) {
	scenariosLabel, err := d.labelRepo.GetByKey(ctx, tenantID, objectType, objectID, model.ScenariosKey)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching scenarios for object with id: %s and type: %s", objectID, objectType)
	}
	return label.ValueToStringsSlice(scenariosLabel.Value)
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
