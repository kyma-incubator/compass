package scenario

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/99designs/gqlgen/graphql"
)

const (
	GetApplicationID                      = "GetApplicationID"
	GetApplicationIDByPackage             = "GetApplicationIDByPackage"
	GetApplicationIDByPackageInstanceAuth = "GetApplicationIDByPackageInstanceAuth"
)

// TODO check if logs are present
type directive struct {
	transact  persistence.Transactioner
	labelRepo label.LabelRepository

	applicationProviders map[string]func(context.Context, string, string) (string, error)
}

// NewDirective returns a new scenario directive
func NewDirective(repoBuilder func() (mp_package.PackageRepository, packageinstanceauth.Repository)) *directive {
	packageRepo, packageInstanceAuthRepo := repoBuilder()

	getApplicationIDByPackageFunc := func(ctx context.Context, tenantID, packageID string) (string, error) {
		pkg, err := packageRepo.GetByID(ctx, tenantID, packageID)
		if err != nil {
			return "", errors.Wrapf(err, "while getting Package with id %s", packageID)
		}
		return pkg.ApplicationID, nil
	}

	return &directive{
		labelRepo: label.NewRepository(label.NewConverter()),
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

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	runtimeID := consumerInfo.ConsumerID
	log.Debugf("Found Consumer ID for the requesting runtime: %v", runtimeID)

	resCtx := graphql.GetResolverContext(ctx)
	id, ok := resCtx.Args[idField].(string)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Could not get idField: %s from request context", idField))
	}

	appProviderFunc, ok := d.applicationProviders[applicationProvider]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Could not get app provider func: %s from provider list", applicationProvider))
	}

	appID, err := appProviderFunc(ctx, tenantID, id)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not derive app id, an error occurred")
	}
	log.Debugf("Found Application ID based on the request parameter %s: %s", idField, appID)

	tx, err := d.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer d.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	commonScenarios := stringsIntersection(appScenarios, runtimeScenarios)
	if len(commonScenarios) == 0 {
		return nil, errors.New("Forbidden: Missing scenarios")
	}

	/* TODO: leave or remove? or we could extract stringsIntersaction in pkg/utils of some sort
	if !hasCommonScenarios(appScenarios, runtimeScenarios) {
		return nil, errors.New("Forbidden: Missing scenarios")
	}
	*/

	log.Infof("Runtime with ID %s is in scenario with the owning application entity with id %s", runtimeID, appID)
	return next(ctx)
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

// hasCommonScenarios returns whether there are common elements in two string slices.
func hasCommonScenarios(str1, str2 []string) bool {
	strings := make(map[string]bool)
	for _, v := range str1 {
		strings[v] = true
	}
	for _, v := range str2 {
		if strings[v] {
			return true
		}
	}
	return false
}

func RepoBuilder() (mp_package.PackageRepository, packageinstanceauth.Repository) {
	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	eventAPIConverter := eventdef.NewConverter(frConverter, versionConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(frConverter, versionConverter)
	packageRepo := mp_package.NewRepository(mp_package.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter))
	packageInstanceAuthRepo := packageinstanceauth.NewRepository(packageinstanceauth.NewConverter(authConverter))

	return packageRepo, packageInstanceAuthRepo
}
