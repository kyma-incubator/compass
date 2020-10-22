package formation

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/99designs/gqlgen/graphql"
)

const (
	GetApplicationIDByPackage             = "GetApplicationIDByPackage"
	GetApplicationIDByPackageInstanceAuth = "GetApplicationIDByPackageInstanceAuth"
)

// TODO check order of directives
// TODO check if logs are present
type directive struct {
	transact  persistence.Transactioner
	labelRepo label.LabelRepository

	applicationProviders map[string]func(context.Context, string, string) (string, error)
	log                  *logrus.Logger
}

func NewDirective(log *logrus.Logger) *directive {
	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	eventAPIConverter := eventdef.NewConverter(frConverter, versionConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(frConverter, versionConverter)
	packageRepo := mp_package.NewRepository(mp_package.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter))
	packageInstanceAuthRepo := packageinstanceauth.NewRepository(packageinstanceauth.NewConverter(authConverter))

	getApplicationIDByPackageFunc := func(ctx context.Context, tenantID, packageID string) (string, error) {
		pkg, err := packageRepo.GetByID(ctx, tenantID, packageID)
		if err != nil {
			return "", errors.Wrapf(err, "while getting Package with id %s", packageID)
		}
		return pkg.ApplicationID, nil
	}

	return &directive{
		labelRepo: label.NewRepository(label.NewConverter()),
		log:       log,
		applicationProviders: map[string]func(context.Context, string, string) (string, error){
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

func (d *directive) HasFormation(ctx context.Context, _ interface{}, next graphql.Resolver, applicationProvider string, idField string) (res interface{}, err error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if consumerInfo.ConsumerType != consumer.Runtime {
		return next(ctx)
	}

	runtimeID := consumerInfo.ConsumerID

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

	runtimeScenarios, err := d.getObjectScenarios(ctx, tenantID, model.RuntimeLabelableObject, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching scenarios for runtime")
	}

	commonScenarios := stringsIntersection(appScenarios, runtimeScenarios)
	if len(commonScenarios) == 0 {
		return nil, errors.New("Forbidden: Missing formations")
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// TODO: check if can be reused for applications query
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
