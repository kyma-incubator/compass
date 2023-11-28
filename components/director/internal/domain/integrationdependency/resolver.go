package integrationdependency

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"strings"
)

const (
	integrationDependencyKeyWord                     = "integrationDependency"
	manuallyAddedIntegrationDependenciesPackageOrdID = "package:manuallyAddedIntegrationDependencies"
	packageTitle                                     = "Integration Dependencies package"
	packageDescription                               = "This package contains manually added integration dependencies"
	packageShortDescription                          = "Manually added package"
	sapCorePolicyLevel                               = "sap:core:v1"
	packageVersion                                   = "1.0.0"
	defaultVersionValue                              = "v1"
	publicVisibility                                 = "public"
	activeReleaseStatus                              = "active"
)

// IntegrationDependencyService is responsible for the service-layer Integration Dependency operations.
//
//go:generate mockery --name=IntegrationDependencyService --output=automock --outpkg=automock --case=underscore --disable-version-string
type IntegrationDependencyService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, packageID *string, in model.IntegrationDependencyInput, integrationDependencyHash uint64) (string, error)
	ListByPackageID(ctx context.Context, packageID string) ([]*model.IntegrationDependency, error)
	Get(ctx context.Context, id string) (*model.IntegrationDependency, error)
	Delete(ctx context.Context, resourceType resource.Type, id string) error
}

// IntegrationDepConverter converts Integration Dependencies between the model.IntegrationDependency service-layer representation and the graphql-layer representation.
//
//go:generate mockery --name=IntegrationDependencyConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type IntegrationDepConverter interface {
	ToGraphQL(in *model.IntegrationDependency, aspects []*model.Aspect) (*graphql.IntegrationDependency, error)
	InputFromGraphQL(in *graphql.IntegrationDependencyInput) (*model.IntegrationDependencyInput, error)
}

// AspectService is responsible for the service-layer Aspect operations.
//
//go:generate mockery --name=AspectService --output=automock --outpkg=automock --case=underscore --disable-version-string
type AspectService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, integrationDependencyID string, in model.AspectInput) (string, error)
	ListByIntegrationDependencyID(ctx context.Context, integrationDependencyID string) ([]*model.Aspect, error)
	ListByApplicationIDs(ctx context.Context, applicationIDs []string, pageSize int, cursor string) ([]*model.Aspect, map[string]int, error)
}

// ApplicationService is responsible for the service-layer Application operations.
//
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	Get(ctx context.Context, id string) (*model.Application, error)
}

// ApplicationTemplateService is responsible for the service-layer Application Template operations.
//
//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateService interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
}

// PackageService is responsible for the service-layer Package operations.
//
//go:generate mockery --name=PackageService --output=automock --outpkg=automock --case=underscore --disable-version-string
type PackageService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.PackageInput, pkgHash uint64) (string, error)
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Package, error)
	Delete(ctx context.Context, resourceType resource.Type, id string) error
}

// Resolver is an object responsible for resolver-layer Integration Dependency operations
type Resolver struct {
	transact                       persistence.Transactioner
	integrationDependencySvc       IntegrationDependencyService
	integrationDependencyConverter IntegrationDepConverter
	aspectSvc                      AspectService
	appSvc                         ApplicationService
	appTemplateSvc                 ApplicationTemplateService
	packageSvc                     PackageService
}

// NewResolver returns a new object responsible for resolver-layer Integration Dependency operations.
func NewResolver(transact persistence.Transactioner, integrationDependencySvc IntegrationDependencyService, integrationDependencyConverter IntegrationDepConverter, aspectSvc AspectService, appSvc ApplicationService, appTemplateSvc ApplicationTemplateService, packageSvc PackageService) *Resolver {
	return &Resolver{
		transact:                       transact,
		integrationDependencySvc:       integrationDependencySvc,
		integrationDependencyConverter: integrationDependencyConverter,
		aspectSvc:                      aspectSvc,
		appSvc:                         appSvc,
		appTemplateSvc:                 appTemplateSvc,
		packageSvc:                     packageSvc,
	}
}

// AddIntegrationDependencyToApplication adds an Integration Dependency in the context of an Application.
func (r *Resolver) AddIntegrationDependencyToApplication(ctx context.Context, appID string, in graphql.IntegrationDependencyInput) (*graphql.IntegrationDependency, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Adding Integration Dependency to application with id %q", appID)

	log.C(ctx).Infof("Getting application namespace for application with id %q", appID)
	appNamespace, err := r.getApplicationNamespace(ctx, appID)
	if err != nil {
		return nil, err
	}

	versionValue := defaultVersionValue
	if in.Version != nil {
		versionValue = in.Version.Value
	}
	// generate values which are mandatory by ORD spec if they are missing
	in.OrdID = getOrdID(in.OrdID, appNamespace, in.Name, versionValue)
	in.Visibility = getVisibility(in.Visibility)
	in.ReleaseStatus = getReleaseStatus(in.ReleaseStatus)
	in.Mandatory = getMandatory(in.Mandatory)

	var packageID string
	if in.PartOfPackage == nil {
		pkgOrdID := fmt.Sprintf("%s:%s:%s", appNamespace, manuallyAddedIntegrationDependenciesPackageOrdID, versionValue)
		log.C(ctx).Infof("Part of package field is missing. Creating a package with ordID %q for application with id %q", pkgOrdID, appID)

		packageID, err = r.createPackage(ctx, appID, pkgOrdID)
		if err != nil {
			return nil, err
		}
		in.PartOfPackage = &pkgOrdID
	} else {
		log.C(ctx).Infof("Part of package field is provided. Getting a package with ordID %q for application with id %q", *in.PartOfPackage, appID)
		packageID, err = r.getPackageID(ctx, appID, in.PartOfPackage)
		if err != nil {
			return nil, err
		}
	}

	convertedIn, err := r.integrationDependencyConverter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting GraphQL input to Integration Dependency input")
	}

	log.C(ctx).Infof("Creating integration dependency for application with id %q", appID)
	integrationDependencyID, err := r.integrationDependencySvc.Create(ctx, resource.Application, appID, &packageID, *convertedIn, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "error occurred while creating Integration Dependency for application with id %q", appID)
	}

	log.C(ctx).Infof("Creating aspects for integration dependency with id %q and application with id %q", integrationDependencyID, appID)
	if err = r.createAspects(ctx, resource.Application, appID, integrationDependencyID, convertedIn.Aspects); err != nil {
		return nil, errors.Wrapf(err, "error occurred while creating Aspects for Integration Dependency with id %q in the context of an application with id %q", integrationDependencyID, appID)
	}

	integrationDependency, err := r.integrationDependencySvc.Get(ctx, integrationDependencyID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Integration Depeendency with id %q", integrationDependencyID)
	}

	aspects, err := r.aspectSvc.ListByIntegrationDependencyID(ctx, integrationDependencyID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Aspects for Integration Depeendency with id %q", integrationDependencyID)
	}

	gqlIntegrationDependency, err := r.integrationDependencyConverter.ToGraphQL(integrationDependency, aspects)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Integration Dependecy with id %q to graphQL", integrationDependencyID)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Integration Depenedncy with id %q successfully added to application with id %q", integrationDependencyID, appID)
	return gqlIntegrationDependency, nil
}

// DeleteIntegrationDependency deletes an Integration Dependency by its ID.
func (r *Resolver) DeleteIntegrationDependency(ctx context.Context, id string) (*graphql.IntegrationDependency, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Deleting Integration Dependency with id %q", id)

	ctx = persistence.SaveToContext(ctx, tx)

	integrationDependency, err := r.integrationDependencySvc.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting for Integration Dependency with id %q", id)
	}

	aspects, err := r.aspectSvc.ListByIntegrationDependencyID(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting aspects for Integration Dependency with id %q", id)
	}

	gqlIntegrationDependency, err := r.integrationDependencyConverter.ToGraphQL(integrationDependency, aspects)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Integration Dependency with id %q to graphQL", id)
	}

	integrationDependencies, err := r.integrationDependencySvc.ListByPackageID(ctx, *integrationDependency.PackageID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Integration Dependencies for package with id %q", *integrationDependency.PackageID)
	}

	if len(integrationDependencies) == 1 {
		log.C(ctx).Infof("Deleting package with id %q for Integration Dependency with id %q", *integrationDependency.PackageID, id)
		// the deletion of the package will delete the integration dependency as well
		if err = r.packageSvc.Delete(ctx, resource.Application, *integrationDependency.PackageID); err != nil {
			return nil, errors.Wrapf(err, "while deleting package with id %q for Integration Dependency with id %q", *integrationDependency.PackageID, id)
		}
	} else {
		if err = r.integrationDependencySvc.Delete(ctx, resource.Application, id); err != nil {
			return nil, errors.Wrapf(err, "while deleting Integration Dependency with id %q", id)
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Integration Dependency with id %q successfully deleted.", id)
	return gqlIntegrationDependency, nil
}

func (r *Resolver) createAspects(ctx context.Context, resourceType resource.Type, resourceID string, integrationDependencyID string, aspects []*model.AspectInput) error {
	for _, aspect := range aspects {
		_, err := r.aspectSvc.Create(ctx, resourceType, resourceID, integrationDependencyID, *aspect)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) getApplicationNamespace(ctx context.Context, appID string) (string, error) {
	app, err := r.appSvc.Get(ctx, appID)
	if err != nil {
		return "", errors.Wrapf(err, "while getting application with id %q", appID)
	}

	if app.ApplicationNamespace != nil {
		return *app.ApplicationNamespace, nil
	}

	if app.ApplicationTemplateID != nil {
		appTemplate, err := r.appTemplateSvc.Get(ctx, *app.ApplicationTemplateID)
		if err != nil {
			return "", errors.Wrapf(err, "failed to retrieve application template for application with ID %q", appID)
		}
		if appTemplate.ApplicationNamespace != nil {
			return *appTemplate.ApplicationNamespace, nil
		}
		return "", errors.Errorf("application namespace is missing for both application template with ID %q and application with id %q", appTemplate.ID, appID)
	}

	return "", errors.Errorf("application namespace is missing for application %q", appID)
}

func (r *Resolver) createPackage(ctx context.Context, appID, pkgOrdID string) (string, error) {
	packageInput := model.PackageInput{
		OrdID:            pkgOrdID,
		Title:            packageTitle,
		Description:      packageDescription,
		ShortDescription: packageShortDescription,
		PolicyLevel:      str.Ptr(sapCorePolicyLevel),
		Version:          packageVersion,
	}
	packageID, err := r.packageSvc.Create(ctx, resource.Application, appID, packageInput, 0)
	if err != nil {
		return "", errors.Wrapf(err, "error occurred while creating package with ORD ID %q for application with id %q", pkgOrdID, appID)
	}

	return packageID, nil
}

func (r *Resolver) getPackageID(ctx context.Context, appID string, packageOrdID *string) (string, error) {
	packagesFromDB, err := r.packageSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return "", errors.Wrapf(err, "while listing packages for application with id %q", appID)
	}
	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return equalStrings(&packagesFromDB[i].OrdID, packageOrdID)
	}); found {
		return packagesFromDB[i].ID, nil
	}
	return "", errors.Errorf("package with ord ID: %q does not exist", *packageOrdID)
}

func searchInSlice(length int, f func(i int) bool) (int, bool) {
	for i := 0; i < length; i++ {
		if f(i) {
			return i, true
		}
	}
	return -1, false
}

func equalStrings(first, second *string) bool {
	return first != nil && second != nil && *first == *second
}

func getOrdID(inputOrdID *string, appNamespace, name, versionValue string) *string {
	if inputOrdID == nil {
		name = strings.ToUpper(strings.ReplaceAll(name, " ", ""))
		inputOrdID = str.Ptr(fmt.Sprintf("%s:%s:%s:%s", appNamespace, integrationDependencyKeyWord, name, versionValue))
	}
	return inputOrdID
}

func getVisibility(inputVisibility *string) *string {
	if inputVisibility == nil {
		inputVisibility = str.Ptr(publicVisibility)
	}
	return inputVisibility
}

func getReleaseStatus(inputReleaseStatus *string) *string {
	if inputReleaseStatus == nil {
		inputReleaseStatus = str.Ptr(activeReleaseStatus)
	}
	return inputReleaseStatus
}

func getMandatory(inputMandatory *bool) *bool {
	m := false
	if inputMandatory == nil {
		inputMandatory = &m
	}
	return inputMandatory
}
