package integrationdependency_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationdependency"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationdependency/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestResolver_AddIntegrationDependencyToApplication(t *testing.T) {
	// GIVEN
	modelIntDep := fixIntegrationDependencyModel(integrationDependencyID)

	aspectID := "aspectID"
	modelAspects := []*model.Aspect{
		{
			ApplicationID:                &appID,
			ApplicationTemplateVersionID: &appTemplateVersionID,
			IntegrationDependencyID:      integrationDependencyID,
			Title:                        title,
			Description:                  str.Ptr(description),
			Mandatory:                    &mandatory,
			SupportMultipleProviders:     &supportMultipleProviders,
			APIResources:                 json.RawMessage("[]"),
			EventResources:               json.RawMessage("[]"),
			BaseEntity: &model.BaseEntity{
				ID:        aspectID,
				Ready:     ready,
				CreatedAt: &fixedTimestamp,
				UpdatedAt: &time.Time{},
				DeletedAt: &time.Time{},
				Error:     nil,
			},
		},
	}

	appNamespace := "test.ns"
	modelApp := &model.Application{
		BaseEntity: &model.BaseEntity{
			ID: appID,
		},
		ApplicationNamespace: str.Ptr(appNamespace),
	}
	modelAppWithoutNamespace := &model.Application{
		BaseEntity: &model.BaseEntity{
			ID: appID,
		},
		ApplicationTemplateID: str.Ptr(appTemplateVersionID),
	}
	modelAppTemplate := &model.ApplicationTemplate{
		ID:                   appTemplateVersionID,
		ApplicationNamespace: str.Ptr(appNamespace),
	}
	modelAppTemplateWithoutNamespace := &model.ApplicationTemplate{
		ID: appTemplateVersionID,
	}

	gqlIntDep := fixGQLIntegrationDependency(integrationDependencyID)
	gqlIntDep.Aspects = []*graphql.Aspect{
		{
			Name:           title,
			Description:    str.Ptr(description),
			Mandatory:      &mandatory,
			APIResources:   []*graphql.AspectAPIDefinition{},
			EventResources: []*graphql.AspectEventDefinition{},
			BaseEntity: &graphql.BaseEntity{
				ID:        aspectID,
				Ready:     true,
				Error:     nil,
				CreatedAt: timeToTimestampPtr(fixedTimestamp),
				UpdatedAt: timeToTimestampPtr(time.Time{}),
				DeletedAt: timeToTimestampPtr(time.Time{}),
			},
		},
	}

	buildPackageOrdID := fmt.Sprintf("%s:package:manuallyAddedIntegrationDependencies:v1", appNamespace)
	gqlIntDepInputWithoutPackage := fixGQLIntegrationDependencyInputWithoutPackage()
	gqlIntDepInputWithPackage := fixGQLIntegrationDependencyInputWithPackageOrdID(buildPackageOrdID)
	gqlIntDepInputWithPackageAndWithoutGeneratedProps := fixGQLIntegrationDependencyInputWithPackageAndWithoutProperties(buildPackageOrdID)
	gqlIntDepInputWithPackageAndWithGeneratedProps := fixGQLIntegrationDependencyInputWithPackageAndWithProperties(appNamespace, buildPackageOrdID)
	gqlIntDepWithGeneratedProperties := fixGQLIntegrationDependencyWithGeneratedProperties(appNamespace, aspectID, buildPackageOrdID)
	modelIntDepInput := fixIntegrationDependencyInputModelWithPackageOrdID(buildPackageOrdID)
	gqlIntDepInputWithPackageAndVersion := fixGQLIntegrationDependencyInputWithPackageOrdID(buildPackageOrdID)
	gqlIntDepInputWithPackageAndVersion.Version = &graphql.VersionInput{Value: versionValue}
	modelIntDepInputWithVersion := fixIntegrationDependencyInputModelWithPackageOrdID(buildPackageOrdID)
	modelIntDepInputWithVersion.VersionInput = &model.VersionInput{Value: versionValue}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                           string
		TransactionerFn                func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		InputObject                    *graphql.IntegrationDependencyInput
		IntegrationDependencyServiceFn func() *automock.IntegrationDependencyService
		AspectServiceFn                func() *automock.AspectService
		ConverterFn                    func() *automock.IntegrationDepConverter
		AppServiceFn                   func() *automock.ApplicationService
		AppTemplateServiceFn           func() *automock.ApplicationTemplateService
		PackageServiceFn               func() *automock.PackageService
		ExpectedIntegrationDependency  *graphql.IntegrationDependency
		ExpectedErr                    error
	}{
		{
			Name:            "Success when Application has app namespace and part of package id is provided",
			TransactionerFn: txGen.ThatSucceeds,
			InputObject:     gqlIntDepInputWithPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), modelIntDepInput, mock.Anything).Return(integrationDependencyID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()
				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *modelIntDepInput.Aspects[0]).Return(mock.Anything, nil).Once()
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("InputFromGraphQL", gqlIntDepInputWithPackage).Return(&modelIntDepInput, nil).Once()
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(gqlIntDep, nil).Once()
				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.Package{{ID: packageID, OrdID: buildPackageOrdID}}, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: gqlIntDep,
			ExpectedErr:                   nil,
		},
		{
			Name:            "Success when Application has app namespace and part of package id and version are provided",
			TransactionerFn: txGen.ThatSucceeds,
			InputObject:     gqlIntDepInputWithPackageAndVersion,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), modelIntDepInputWithVersion, mock.Anything).Return(integrationDependencyID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()
				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *modelIntDepInputWithVersion.Aspects[0]).Return(mock.Anything, nil).Once()
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("InputFromGraphQL", gqlIntDepInputWithPackageAndVersion).Return(&modelIntDepInputWithVersion, nil).Once()
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(gqlIntDep, nil).Once()
				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.Package{{ID: packageID, OrdID: buildPackageOrdID}}, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: gqlIntDep,
			ExpectedErr:                   nil,
		},
		{
			Name:            "Success when Application has app namespace and part of package id is not provided",
			TransactionerFn: txGen.ThatSucceeds,
			InputObject:     gqlIntDepInputWithoutPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), modelIntDepInput, mock.Anything).Return(integrationDependencyID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()
				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *modelIntDepInput.Aspects[0]).Return(mock.Anything, nil).Once()
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("InputFromGraphQL", gqlIntDepInputWithPackage).Return(&modelIntDepInput, nil).Once()
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(gqlIntDep, nil).Once()
				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, mock.Anything, mock.Anything).Return(packageID, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: gqlIntDep,
			ExpectedErr:                   nil,
		},
		{
			Name:            "Success when Application has app namespace and part of package id is provided, but other properties are missing from input",
			TransactionerFn: txGen.ThatSucceeds,
			InputObject:     gqlIntDepInputWithPackageAndWithoutGeneratedProps,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				modelIntDepInput.OrdID = str.Ptr(fmt.Sprintf("%s:integrationDependency:%s:v1", appNamespace, strings.ToUpper(title)))
				modelIntDepInput.Mandatory = &mandatory
				modelIntDepInput.Visibility = publicVisibility
				modelIntDepInput.ReleaseStatus = str.Ptr(releaseStatus)
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), modelIntDepInput, mock.Anything).Return(integrationDependencyID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()
				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *modelIntDepInput.Aspects[0]).Return(mock.Anything, nil).Once()
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				modelIntDepInput.OrdID = str.Ptr(fmt.Sprintf("%s:integrationDependency:%s:v1", appNamespace, strings.ToUpper(title)))
				modelIntDepInput.Mandatory = &mandatory
				modelIntDepInput.Visibility = publicVisibility
				modelIntDepInput.ReleaseStatus = str.Ptr(releaseStatus)
				conv.On("InputFromGraphQL", gqlIntDepInputWithPackageAndWithGeneratedProps).Return(&modelIntDepInput, nil).Once()
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(gqlIntDepWithGeneratedProperties, nil).Once()
				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.Package{{ID: packageID, OrdID: buildPackageOrdID}}, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: gqlIntDepWithGeneratedProperties,
			ExpectedErr:                   nil,
		},
		{
			Name:            "Success when Application does not have app namespace, but Application Template does and part of package id is not provided",
			TransactionerFn: txGen.ThatSucceeds,
			InputObject:     gqlIntDepInputWithoutPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), modelIntDepInput, mock.Anything).Return(integrationDependencyID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()
				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *modelIntDepInput.Aspects[0]).Return(mock.Anything, nil).Once()
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("InputFromGraphQL", gqlIntDepInputWithPackage).Return(&modelIntDepInput, nil).Once()
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(gqlIntDep, nil).Once()
				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelAppWithoutNamespace, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), *modelAppWithoutNamespace.ApplicationTemplateID).Return(modelAppTemplate, nil).Once()

				return svc
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, mock.Anything, mock.Anything).Return(packageID, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: gqlIntDep,
			ExpectedErr:                   nil,
		},
		{
			Name:            "Error when Application does not have app namespace and Application Template does not have app namespace and part of package id is not provided",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     gqlIntDepInputWithoutPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				return &automock.IntegrationDependencyService{}
			},
			AspectServiceFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				return &automock.IntegrationDepConverter{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelAppWithoutNamespace, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), *modelAppWithoutNamespace.ApplicationTemplateID).Return(modelAppTemplateWithoutNamespace, nil).Once()

				return svc
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   errors.New("application namespace is missing for both"),
		},
		{
			Name:            "Error when getting Application template fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     &graphql.IntegrationDependencyInput{},
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				return &automock.IntegrationDependencyService{}
			},
			AspectServiceFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				return &automock.IntegrationDepConverter{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelAppWithoutNamespace, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), *modelAppWithoutNamespace.ApplicationTemplateID).Return(nil, testErr).Once()

				return svc
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when getting Application fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     &graphql.IntegrationDependencyInput{},
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				return &automock.IntegrationDependencyService{}
			},
			AspectServiceFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				return &automock.IntegrationDepConverter{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when Application does not have app namespace and app template",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     &graphql.IntegrationDependencyInput{},
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				return &automock.IntegrationDependencyService{}
			},
			AspectServiceFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				return &automock.IntegrationDepConverter{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				modelAppWithoutNamespace.ApplicationTemplateID = nil
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelAppWithoutNamespace, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   errors.New("application namespace is missing for application"),
		},
		{
			Name:            "Error when Create Package fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     gqlIntDepInputWithoutPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				return &automock.IntegrationDependencyService{}
			},
			AspectServiceFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				return &automock.IntegrationDepConverter{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, mock.Anything, mock.Anything).Return("", testErr).Once()

				return svc
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when listByApplicationID for package fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     gqlIntDepInputWithPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				return &automock.IntegrationDependencyService{}
			},
			AspectServiceFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				return &automock.IntegrationDepConverter{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.Package{}, testErr).Once()

				return svc
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when the input`s part of package id is provided, but it does not exist",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     gqlIntDepInputWithPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				return &automock.IntegrationDependencyService{}
			},
			AspectServiceFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				return &automock.IntegrationDepConverter{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.Package{{ID: packageID, OrdID: "wrongOrdID"}}, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   errors.New("does not exist"),
		},
		{
			Name:            "Error when InputFromGraphQL fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     gqlIntDepInputWithPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				return &automock.IntegrationDependencyService{}
			},
			AspectServiceFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("InputFromGraphQL", gqlIntDepInputWithPackage).Return(nil, testErr).Once()

				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.Package{{ID: packageID, OrdID: buildPackageOrdID}}, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when Create Integration Dependency fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     gqlIntDepInputWithPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), modelIntDepInput, mock.Anything).Return("", testErr).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("InputFromGraphQL", gqlIntDepInputWithPackage).Return(&modelIntDepInput, nil).Once()

				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.Package{{ID: packageID, OrdID: buildPackageOrdID}}, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when create Aspects fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     gqlIntDepInputWithPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), modelIntDepInput, mock.Anything).Return(integrationDependencyID, nil).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *modelIntDepInput.Aspects[0]).Return(mock.Anything, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("InputFromGraphQL", gqlIntDepInputWithPackage).Return(&modelIntDepInput, nil).Once()

				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.Package{{ID: packageID, OrdID: buildPackageOrdID}}, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when get Integration Dependency fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     gqlIntDepInputWithPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), modelIntDepInput, mock.Anything).Return(integrationDependencyID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(nil, testErr).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *modelIntDepInput.Aspects[0]).Return(mock.Anything, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("InputFromGraphQL", gqlIntDepInputWithPackage).Return(&modelIntDepInput, nil).Once()

				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.Package{{ID: packageID, OrdID: buildPackageOrdID}}, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when get Aspects by Integration Dependency id fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     gqlIntDepInputWithPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), modelIntDepInput, mock.Anything).Return(integrationDependencyID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *modelIntDepInput.Aspects[0]).Return(mock.Anything, nil).Once()
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("InputFromGraphQL", gqlIntDepInputWithPackage).Return(&modelIntDepInput, nil).Once()

				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.Package{{ID: packageID, OrdID: buildPackageOrdID}}, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when ToGraphQL fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputObject:     gqlIntDepInputWithPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), modelIntDepInput, mock.Anything).Return(integrationDependencyID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *modelIntDepInput.Aspects[0]).Return(mock.Anything, nil).Once()
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("InputFromGraphQL", gqlIntDepInputWithPackage).Return(&modelIntDepInput, nil).Once()
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(nil, testErr).Once()

				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.Package{{ID: packageID, OrdID: buildPackageOrdID}}, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when transaction commit fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			InputObject:     gqlIntDepInputWithPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), modelIntDepInput, mock.Anything).Return(integrationDependencyID, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *modelIntDepInput.Aspects[0]).Return(mock.Anything, nil).Once()
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("InputFromGraphQL", gqlIntDepInputWithPackage).Return(&modelIntDepInput, nil).Once()
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(gqlIntDep, nil).Once()

				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(modelApp, nil).Once()

				return svc
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.Package{{ID: packageID, OrdID: buildPackageOrdID}}, nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when transaction fails on begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			InputObject:     gqlIntDepInputWithPackage,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				return &automock.IntegrationDependencyService{}
			},
			AspectServiceFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				return &automock.IntegrationDepConverter{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				return &automock.ApplicationTemplateService{}
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			intDepSvc := testCase.IntegrationDependencyServiceFn()
			aspectSvc := testCase.AspectServiceFn()
			converter := testCase.ConverterFn()
			appSvc := testCase.AppServiceFn()
			appTemplateSvc := testCase.AppTemplateServiceFn()
			packageSvc := testCase.PackageServiceFn()
			defer mock.AssertExpectationsForObjects(t, persist, transact, intDepSvc, aspectSvc, converter, appSvc, appTemplateSvc, packageSvc)

			resolver := integrationdependency.NewResolver(transact, intDepSvc, converter, aspectSvc, appSvc, appTemplateSvc, packageSvc)

			// WHEN
			result, err := resolver.AddIntegrationDependencyToApplication(context.TODO(), appID, *testCase.InputObject)

			// THEN
			assert.Equal(t, testCase.ExpectedIntegrationDependency, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestResolver_DeleteIntegrationDependency(t *testing.T) {
	// GIVEN
	modelIntDep := fixIntegrationDependencyModel(integrationDependencyID)

	aspectID := "aspectID"
	modelAspects := []*model.Aspect{
		{
			ApplicationID:                &appID,
			ApplicationTemplateVersionID: &appTemplateVersionID,
			IntegrationDependencyID:      integrationDependencyID,
			Title:                        title,
			Description:                  str.Ptr(description),
			Mandatory:                    &mandatory,
			SupportMultipleProviders:     &supportMultipleProviders,
			APIResources:                 json.RawMessage("[]"),
			EventResources:               json.RawMessage("[]"),
			BaseEntity: &model.BaseEntity{
				ID:        aspectID,
				Ready:     ready,
				CreatedAt: &fixedTimestamp,
				UpdatedAt: &time.Time{},
				DeletedAt: &time.Time{},
				Error:     nil,
			},
		},
	}
	gqlIntDep := fixGQLIntegrationDependency(integrationDependencyID)
	gqlIntDep.Aspects = []*graphql.Aspect{
		{
			Name:           title,
			Description:    str.Ptr(description),
			Mandatory:      &mandatory,
			APIResources:   []*graphql.AspectAPIDefinition{},
			EventResources: []*graphql.AspectEventDefinition{},
			BaseEntity: &graphql.BaseEntity{
				ID:        aspectID,
				Ready:     true,
				Error:     nil,
				CreatedAt: timeToTimestampPtr(fixedTimestamp),
				UpdatedAt: timeToTimestampPtr(time.Time{}),
				DeletedAt: timeToTimestampPtr(time.Time{}),
			},
		},
	}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                           string
		TransactionerFn                func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		IntegrationDependencyServiceFn func() *automock.IntegrationDependencyService
		AspectServiceFn                func() *automock.AspectService
		ConverterFn                    func() *automock.IntegrationDepConverter
		PackageServiceFn               func() *automock.PackageService
		ExpectedIntegrationDependency  *graphql.IntegrationDependency
		ExpectedErr                    error
	}{
		{
			Name:            "Success there is only one Integration Dependency for Package",
			TransactionerFn: txGen.ThatSucceeds,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()
				svc.On("ListByPackageID", txtest.CtxWithDBMatcher(), *modelIntDep.PackageID).Return([]*model.IntegrationDependency{modelIntDep}, nil).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(gqlIntDep, nil).Once()
				return conv
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, *modelIntDep.PackageID).Return(nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: gqlIntDep,
			ExpectedErr:                   nil,
		},
		{
			Name:            "Success there is more than one Integration Dependency for Package",
			TransactionerFn: txGen.ThatSucceeds,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()
				svc.On("ListByPackageID", txtest.CtxWithDBMatcher(), *modelIntDep.PackageID).Return([]*model.IntegrationDependency{modelIntDep, modelIntDep}, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, integrationDependencyID).Return(nil).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(gqlIntDep, nil).Once()
				return conv
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ExpectedIntegrationDependency: gqlIntDep,
			ExpectedErr:                   nil,
		},
		{
			Name:            "Error when getting Integration Dependency fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(nil, testErr).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				return &automock.IntegrationDepConverter{}
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when getting Aspects by Integration Dependency id fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				return &automock.IntegrationDepConverter{}
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when ToGraphQL fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(nil, testErr).Once()
				return conv
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error getting Integration Dependencies by Package id fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()
				svc.On("ListByPackageID", txtest.CtxWithDBMatcher(), *modelIntDep.PackageID).Return(nil, testErr).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(gqlIntDep, nil).Once()
				return conv
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error getting Integration Dependencies by Package id fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()
				svc.On("ListByPackageID", txtest.CtxWithDBMatcher(), *modelIntDep.PackageID).Return([]*model.IntegrationDependency{modelIntDep}, nil).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(gqlIntDep, nil).Once()
				return conv
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, *modelIntDep.PackageID).Return(testErr).Once()

				return svc
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error getting Integration Dependencies by Package id fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()
				svc.On("ListByPackageID", txtest.CtxWithDBMatcher(), *modelIntDep.PackageID).Return([]*model.IntegrationDependency{modelIntDep, modelIntDep}, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, integrationDependencyID).Return(testErr).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(gqlIntDep, nil).Once()
				return conv
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when transaction fails on commit",
			TransactionerFn: txGen.ThatFailsOnCommit,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				svc := &automock.IntegrationDependencyService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelIntDep, nil).Once()
				svc.On("ListByPackageID", txtest.CtxWithDBMatcher(), *modelIntDep.PackageID).Return([]*model.IntegrationDependency{modelIntDep}, nil).Once()

				return svc
			},
			AspectServiceFn: func() *automock.AspectService {
				svc := &automock.AspectService{}
				svc.On("ListByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(modelAspects, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				conv := &automock.IntegrationDepConverter{}
				conv.On("ToGraphQL", modelIntDep, modelAspects).Return(gqlIntDep, nil).Once()
				return conv
			},
			PackageServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, *modelIntDep.PackageID).Return(nil).Once()

				return svc
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
		{
			Name:            "Error when transaction fails on begin",
			TransactionerFn: txGen.ThatFailsOnBegin,
			IntegrationDependencyServiceFn: func() *automock.IntegrationDependencyService {
				return &automock.IntegrationDependencyService{}
			},
			AspectServiceFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			ConverterFn: func() *automock.IntegrationDepConverter {
				return &automock.IntegrationDepConverter{}
			},
			PackageServiceFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			ExpectedIntegrationDependency: nil,
			ExpectedErr:                   testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			intDepSvc := testCase.IntegrationDependencyServiceFn()
			aspectSvc := testCase.AspectServiceFn()
			converter := testCase.ConverterFn()
			packageSvc := testCase.PackageServiceFn()
			defer mock.AssertExpectationsForObjects(t, persist, transact, intDepSvc, aspectSvc, converter, packageSvc)

			resolver := integrationdependency.NewResolver(transact, intDepSvc, converter, aspectSvc, nil, nil, packageSvc)

			// WHEN
			result, err := resolver.DeleteIntegrationDependency(context.TODO(), integrationDependencyID)

			// THEN
			assert.Equal(t, testCase.ExpectedIntegrationDependency, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}
		})
	}
}
