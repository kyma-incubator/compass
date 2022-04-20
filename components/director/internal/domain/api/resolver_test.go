package api_test

import (
	"context"
	"testing"
	"time"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_AddAPIToBundle(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "bar"

	modelAPI, spec, bundleRef := fixFullAPIDefinitionModel("test")
	modelBndl := &model.Bundle{
		ApplicationID: appID,
		BaseEntity: &model.BaseEntity{
			ID: bundleID,
		},
	}
	gqlAPI := fixFullGQLAPIDefinition("test")
	gqlAPIInput := fixGQLAPIDefinitionInput("name", "foo", "bar")
	modelAPIInput, specInput := fixModelAPIDefinitionInput("name", "foo", "bar")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                   string
		TransactionerFn        func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn              func() *automock.APIService
		BndlServiceFn          func() *automock.BundleService
		BndlReferenceServiceFn func() *automock.BundleReferenceService
		SpecServiceFn          func() *automock.SpecService
		ConverterFn            func() *automock.APIConverter
		AppServiceFn           func() *automock.ApplicationService
		ExpectedAPI            *graphql.APIDefinition
		ExpectedErr            error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), appID, bundleID, *modelAPIInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPI, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), bundleID).Return(modelBndl, nil)
				return appSvc
			},
			BndlReferenceServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &gqlAPI.ID, str.Ptr(bundleID)).Return(&bundleRef, nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, gqlAPI.ID).Return(&spec, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				conv.On("ToGraphQL", &modelAPI, &spec, &bundleRef).Return(gqlAPI, nil).Once()
				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("UpdateBaseURL", txtest.CtxWithDBMatcher(), appID, targetURL).Return(nil).Once()

				return svc
			},
			ExpectedAPI: gqlAPI,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				return &automock.APIService{}
			},
			BndlServiceFn: func() *automock.BundleService {
				return &automock.BundleService{}
			},
			BndlReferenceServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.AssertNotCalled(t, "UpdateBaseURL")

				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when bundle not exist",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				return &automock.APIService{}
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), bundleID).Return(nil, apperrors.NewNotFoundError(resource.Bundle, bundleID))
				return appSvc
			},
			BndlReferenceServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.AssertNotCalled(t, "UpdateBaseURL")

				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: errors.New("cannot add API to not existing bundle"),
		},
		{
			Name:            "Returns error when bundle existence check failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), bundleID).Return(nil, testErr)
				return appSvc
			},
			BndlReferenceServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.AssertNotCalled(t, "UpdateBaseURL")

				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when API creation failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), appID, bundleID, *modelAPIInput, specInput).Return("", testErr).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), bundleID).Return(modelBndl, nil)
				return appSvc
			},
			BndlReferenceServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.AssertNotCalled(t, "UpdateBaseURL")

				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when API retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), appID, bundleID, *modelAPIInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), bundleID).Return(modelBndl, nil)
				return appSvc
			},
			BndlReferenceServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.AssertNotCalled(t, "UpdateBaseURL")

				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},

		{
			Name:            "Returns error when UpdateBaseURL fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), appID, bundleID, *modelAPIInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPI, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), bundleID).Return(modelBndl, nil)
				return appSvc
			},
			BndlReferenceServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.AssertNotCalled(t, "GetByReferenceObjectID")
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("UpdateBaseURL", txtest.CtxWithDBMatcher(), appID, targetURL).Return(testErr).Once()

				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when Spec retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), appID, bundleID, *modelAPIInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPI, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), bundleID).Return(modelBndl, nil)
				return appSvc
			},
			BndlReferenceServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, gqlAPI.ID).Return(nil, testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("UpdateBaseURL", txtest.CtxWithDBMatcher(), appID, targetURL).Return(nil).Once()

				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when BundleReference retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), appID, bundleID, *modelAPIInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPI, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), bundleID).Return(modelBndl, nil)
				return appSvc
			},
			BndlReferenceServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &gqlAPI.ID, str.Ptr(bundleID)).Return(nil, testErr).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, gqlAPI.ID).Return(&spec, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("UpdateBaseURL", txtest.CtxWithDBMatcher(), appID, targetURL).Return(nil).Once()

				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when converting to graphql",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), appID, bundleID, *modelAPIInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPI, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), bundleID).Return(modelBndl, nil)
				return appSvc
			},
			BndlReferenceServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &gqlAPI.ID, str.Ptr(bundleID)).Return(&bundleRef, nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, gqlAPI.ID).Return(&spec, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				conv.On("ToGraphQL", &modelAPI, &spec, &bundleRef).Return(gqlAPI, testErr).Once()
				return conv
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("UpdateBaseURL", txtest.CtxWithDBMatcher(), appID, targetURL).Return(nil).Once()

				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), appID, bundleID, *modelAPIInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPI, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), bundleID).Return(modelBndl, nil)
				return appSvc
			},
			BndlReferenceServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &gqlAPI.ID, str.Ptr(bundleID)).Return(&bundleRef, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				conv.On("ToGraphQL", &modelAPI, &spec, &bundleRef).Return(gqlAPI, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, gqlAPI.ID).Return(&spec, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("UpdateBaseURL", txtest.CtxWithDBMatcher(), appID, targetURL).Return(nil).Once()

				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			bndlSvc := testCase.BndlServiceFn()
			bndlRefSvc := testCase.BndlReferenceServiceFn()
			specSvc := testCase.SpecServiceFn()
			appSvc := testCase.AppServiceFn()

			resolver := api.NewResolver(transact, svc, nil, bndlSvc, bndlRefSvc, converter, nil, specSvc, nil, appSvc)

			// WHEN
			result, err := resolver.AddAPIDefinitionToBundle(context.TODO(), bundleID, *gqlAPIInput)

			// THEN
			assert.Equal(t, testCase.ExpectedAPI, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			bndlSvc.AssertExpectations(t)
			bndlRefSvc.AssertExpectations(t)
			specSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
			appSvc.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteAPI(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "bar"
	var nilBundleID *string

	modelAPIDefinition, spec, bundleRef := fixFullAPIDefinitionModel("test")
	gqlAPIDefinition := fixFullGQLAPIDefinition("test")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn          func() *automock.APIService
		ConverterFn        func() *automock.APIConverter
		SpecServiceFn      func() *automock.SpecService
		BundleRefServiceFn func() *automock.BundleReferenceService
		ExpectedAPI        *graphql.APIDefinition
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", &modelAPIDefinition, &spec, &bundleRef).Return(gqlAPIDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&spec, nil).Once()
				return svc
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPIDefinition.ID, nilBundleID).Return(&bundleRef, nil).Once()
				return svc
			},
			ExpectedAPI: gqlAPIDefinition,
			ExpectedErr: nil,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				return &automock.APIService{}
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when API retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when Spec retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(nil, testErr).Once()
				return svc
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when BundleReference retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&spec, nil).Once()
				return svc
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPIDefinition.ID, nilBundleID).Return(nil, testErr).Once()
				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when API conversion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", &modelAPIDefinition, &spec, &bundleRef).Return(nil, testErr).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&spec, nil).Once()
				return svc
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPIDefinition.ID, nilBundleID).Return(&bundleRef, nil).Once()
				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when API deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", &modelAPIDefinition, &spec, &bundleRef).Return(gqlAPIDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&spec, nil).Once()
				return svc
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPIDefinition.ID, nilBundleID).Return(&bundleRef, nil).Once()
				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", &modelAPIDefinition, &spec, &bundleRef).Return(gqlAPIDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&spec, nil).Once()
				return svc
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPIDefinition.ID, nilBundleID).Return(&bundleRef, nil).Once()
				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			specService := testCase.SpecServiceFn()
			converter := testCase.ConverterFn()
			bundleRefService := testCase.BundleRefServiceFn()

			resolver := api.NewResolver(transact, svc, nil, nil, bundleRefService, converter, nil, specService, nil, nil)

			// WHEN
			result, err := resolver.DeleteAPIDefinition(context.TODO(), id)

			// THEN
			assert.Equal(t, testCase.ExpectedAPI, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			svc.AssertExpectations(t)
			specService.AssertExpectations(t)
			converter.AssertExpectations(t)
			bundleRefService.AssertExpectations(t)
			transact.AssertExpectations(t)
			persist.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateAPI(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "bar"
	var nilBundleID *string
	gqlAPIDefinitionInput := fixGQLAPIDefinitionInput(id, "foo", "bar")
	modelAPIDefinitionInput, modelSpecInput := fixModelAPIDefinitionInput(id, "foo", "bar")
	gqlAPIDefinition := fixFullGQLAPIDefinition("test")
	modelAPIDefinition, modelSpec, modelBundleRef := fixFullAPIDefinitionModel("test")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                  string
		TransactionerFn       func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn             func() *automock.APIService
		ConverterFn           func() *automock.APIConverter
		SpecServiceFn         func() *automock.SpecService
		BundleRefServiceFn    func() *automock.BundleReferenceService
		InputWebhookID        string
		InputAPI              graphql.APIDefinitionInput
		ExpectedAPIDefinition *graphql.APIDefinition
		ExpectedErr           error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				conv.On("ToGraphQL", &modelAPIDefinition, &modelSpec, &modelBundleRef).Return(gqlAPIDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&modelSpec, nil).Once()
				return svc
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPIDefinition.ID, nilBundleID).Return(&modelBundleRef, nil).Once()
				return svc
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: gqlAPIDefinition,
			ExpectedErr:           nil,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				return &automock.APIService{}
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name:            "Returns error when converting input to GraphQL fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				return &automock.APIService{}
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(nil, nil, testErr).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name:            "Returns error when API update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name:            "Returns error when API retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name:            "Returns error when Spec retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(nil, testErr).Once()
				return svc
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name:            "Returns error when BundleReference retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&modelSpec, nil).Once()
				return svc
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPIDefinition.ID, nilBundleID).Return(nil, testErr).Once()
				return svc
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				conv.On("ToGraphQL", &modelAPIDefinition, &modelSpec, &modelBundleRef).Return(nil, testErr).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&modelSpec, nil).Once()
				return svc
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPIDefinition.ID, nilBundleID).Return(&modelBundleRef, nil).Once()
				return svc
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				conv.On("ToGraphQL", &modelAPIDefinition, &modelSpec, &modelBundleRef).Return(gqlAPIDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&modelSpec, nil).Once()
				return svc
			},
			BundleRefServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPIDefinition.ID, nilBundleID).Return(&modelBundleRef, nil).Once()
				return svc
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			specService := testCase.SpecServiceFn()
			bundleRefService := testCase.BundleRefServiceFn()

			resolver := api.NewResolver(transact, svc, nil, nil, bundleRefService, converter, nil, specService, nil, nil)

			// WHEN
			result, err := resolver.UpdateAPIDefinition(context.TODO(), id, *gqlAPIDefinitionInput)

			// THEN
			assert.Equal(t, testCase.ExpectedAPIDefinition, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			specService.AssertExpectations(t)
			bundleRefService.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_RefetchAPISpec(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")

	apiID := "apiID"
	specID := "specID"

	dataBytes := "data"
	modelSpec := &model.Spec{
		ID:   specID,
		Data: &dataBytes,
	}

	clob := graphql.CLOB(dataBytes)
	gqlAPISpec := &graphql.APISpec{
		Data: &clob,
	}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.SpecService
		ConvFn          func() *automock.SpecConverter
		ExpectedAPISpec *graphql.APISpec
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, apiID).Return(modelSpec, nil).Once()
				svc.On("RefetchSpec", txtest.CtxWithDBMatcher(), specID, model.APISpecReference).Return(modelSpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLAPISpec", modelSpec).Return(gqlAPISpec, nil).Once()
				return conv
			},
			ExpectedAPISpec: gqlAPISpec,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ConvFn: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when getting spec failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, apiID).Return(nil, testErr).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when spec not found",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, apiID).Return(nil, nil).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     errors.Errorf("spec for API with id %q not found", apiID),
		},
		{
			Name:            "Returns error when refetching api spec failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, apiID).Return(modelSpec, nil).Once()
				svc.On("RefetchSpec", txtest.CtxWithDBMatcher(), specID, model.APISpecReference).Return(nil, testErr).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error converting to GraphQL fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, apiID).Return(modelSpec, nil).Once()
				svc.On("RefetchSpec", txtest.CtxWithDBMatcher(), specID, model.APISpecReference).Return(modelSpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLAPISpec", modelSpec).Return(nil, testErr).Once()
				return conv
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, apiID).Return(modelSpec, nil).Once()
				svc.On("RefetchSpec", txtest.CtxWithDBMatcher(), specID, model.APISpecReference).Return(modelSpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLAPISpec", modelSpec).Return(gqlAPISpec, nil).Once()
				return conv
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			svc := testCase.ServiceFn()
			conv := testCase.ConvFn()
			persist, transact := testCase.TransactionerFn()
			resolver := api.NewResolver(transact, nil, nil, nil, nil, nil, nil, svc, conv, nil)

			// WHEN
			result, err := resolver.RefetchAPISpec(context.TODO(), apiID)

			// then
			assert.Equal(t, testCase.ExpectedAPISpec, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			conv.AssertExpectations(t)
		})
	}
}

func TestResolver_FetchRequest(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	firstSpecID := "specID"
	secondSpecID := "specID2"
	specIDs := []string{firstSpecID, secondSpecID}
	firstFRID := "frID"
	secondFRID := "frID2"
	frURL := "foo.bar"
	timestamp := time.Now()

	frFirstSpec := fixModelFetchRequest(firstFRID, frURL, timestamp)
	frSecondSpec := fixModelFetchRequest(secondFRID, frURL, timestamp)
	fetchRequests := []*model.FetchRequest{frFirstSpec, frSecondSpec}

	gqlFRFirstSpec := fixGQLFetchRequest(frURL, timestamp)
	gqlFRSecondSpec := fixGQLFetchRequest(frURL, timestamp)
	gqlFetchRequests := []*graphql.FetchRequest{gqlFRFirstSpec, gqlFRSecondSpec}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.APIService
		ConverterFn     func() *automock.FetchRequestConverter
		ExpectedResult  []*graphql.FetchRequest
		ExpectedErr     []error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListFetchRequests", txtest.CtxWithDBMatcher(), specIDs).Return(fetchRequests, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("ToGraphQL", frFirstSpec).Return(gqlFRFirstSpec, nil).Once()
				conv.On("ToGraphQL", frSecondSpec).Return(gqlFRSecondSpec, nil).Once()
				return conv
			},
			ExpectedResult: gqlFetchRequests,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "FetchRequest doesn't exist",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListFetchRequests", txtest.CtxWithDBMatcher(), specIDs).Return(nil, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    nil,
		},
		{
			Name:            "Error when listing API FetchRequests",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListFetchRequests", txtest.CtxWithDBMatcher(), specIDs).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Error when converting FetchRequest to graphql",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListFetchRequests", txtest.CtxWithDBMatcher(), specIDs).Return(fetchRequests, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("ToGraphQL", frFirstSpec).Return(nil, testErr).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListFetchRequests", txtest.CtxWithDBMatcher(), specIDs).Return(fetchRequests, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("ToGraphQL", frFirstSpec).Return(gqlFRFirstSpec, nil).Once()
				conv.On("ToGraphQL", frSecondSpec).Return(gqlFRSecondSpec, nil).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			firstFRParams := dataloader.ParamFetchRequestAPIDef{ID: firstSpecID, Ctx: context.TODO()}
			secondFRParams := dataloader.ParamFetchRequestAPIDef{ID: secondSpecID, Ctx: context.TODO()}
			keys := []dataloader.ParamFetchRequestAPIDef{firstFRParams, secondFRParams}
			resolver := api.NewResolver(transact, svc, nil, nil, nil, nil, converter, nil, nil, nil)

			// WHEN
			result, err := resolver.FetchRequestAPIDefDataLoader(keys)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
		})
	}

	t.Run("Returns error when there are no Specs", func(t *testing.T) {
		resolver := api.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.FetchRequestAPIDefDataLoader([]dataloader.ParamFetchRequestAPIDef{})
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInternalError("No APIDef specs found").Error())
	})

	t.Run("Returns error when Specification ID is empty", func(t *testing.T) {
		params := dataloader.ParamFetchRequestAPIDef{ID: "", Ctx: context.TODO()}
		keys := []dataloader.ParamFetchRequestAPIDef{params}

		resolver := api.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.FetchRequestAPIDefDataLoader(keys)
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInternalError("Cannot fetch FetchRequest. APIDefinition Spec ID is empty").Error())
	})
}
