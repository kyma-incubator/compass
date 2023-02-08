package application_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"
	"github.com/kyma-incubator/compass/components/director/internal/tokens"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application/automock"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var contextParam = txtest.CtxWithDBMatcher()

func TestResolver_RegisterApplication(t *testing.T) {
	// GIVEN
	modelApplication := fixModelApplication("foo", "tenant-foo", "Foo", "Lorem ipsum")
	gqlApplication := fixGQLApplication("foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.ApplicationRegisterInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.ApplicationRegisterInput{
		Name:        "Foo",
		Description: &desc,
	}

	modelInputWithLabel := model.ApplicationRegisterInput{
		Name:        modelInput.Name,
		Description: modelInput.Description,
		Labels:      map[string]interface{}{"managed": "false"},
	}
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		Input               graphql.ApplicationRegisterInput
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), "foo").Return(modelApplication, nil).Once()
				svc.On("Create", context.TODO(), modelInputWithLabel).Return("foo", nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("CreateInputFromGraphQL", mock.Anything, gqlInput).Return(modelInput, nil).Once()
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name:            "Returns error when application creation failed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Create", context.TODO(), modelInputWithLabel).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("CreateInputFromGraphQL", mock.Anything, gqlInput).Return(modelInput, nil).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Returns error when application fetch failed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Create", context.TODO(), modelInputWithLabel).Return("foo", nil).Once()
				svc.On("Get", context.TODO(), "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("CreateInputFromGraphQL", mock.Anything, gqlInput).Return(modelInput, nil).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")
			resolver.SetConverter(converter)

			// WHEN
			result, err := resolver.RegisterApplication(context.TODO(), testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transact.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateApplication(t *testing.T) {
	// GIVEN
	modelApplication := fixModelApplication("foo", "tenant-foo", "Foo", "Lorem ipsum")
	gqlApplication := fixGQLApplication("foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.ApplicationUpdateInput{
		Description: &desc,
	}
	modelInput := model.ApplicationUpdateInput{
		Description: &desc,
	}
	applicationID := "foo"

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		ApplicationID       string
		Input               graphql.ApplicationUpdateInput
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), "foo").Return(modelApplication, nil).Once()
				svc.On("Update", context.TODO(), applicationID, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("UpdateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name:            "Returns error when application update failed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Update", context.TODO(), applicationID, modelInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("UpdateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Update", context.TODO(), applicationID, modelInput).Return(nil).Once()
				svc.On("Get", context.TODO(), "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("UpdateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")
			resolver.SetConverter(converter)

			// WHEN
			result, err := resolver.UpdateApplication(context.TODO(), testCase.ApplicationID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transact.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_UnregisterApplication(t *testing.T) {
	// GIVEN
	appID := uuid.New()
	modelApplication := fixModelApplication(appID.String(), "tenant-foo", "Foo", "Bar")
	gqlApplication := fixGQLApplication(appID.String(), "Foo", "Bar")
	testErr := errors.New("Test error")
	testAuths := fixOAuths()
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		EventingSvcFn       func() *automock.EventingService
		SysAuthServiceFn    func() *automock.SystemAuthService
		OAuth20ServiceFn    func() *automock.OAuth20Service
		InputID             string
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), appID.String()).Return(modelApplication, nil).Once()
				svc.On("Delete", context.TODO(), appID.String()).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				svc.On("CleanupAfterUnregisteringApplication", context.TODO(), appID).Return(nil, nil).Once()
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", context.TODO(), pkgmodel.ApplicationReference, modelApplication.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", context.TODO(), testAuths).Return(nil)
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name:            "Returns error when application deletion failed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), appID.String()).Return(modelApplication, nil).Once()
				svc.On("Delete", context.TODO(), appID.String()).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				svc.On("CleanupAfterUnregisteringApplication", context.TODO(), appID).Return(nil, nil).Once()
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", context.TODO(), pkgmodel.ApplicationReference, modelApplication.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", context.TODO(), testAuths).Return(nil)

				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), appID.String()).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Return error when listing all auths failed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), appID.String()).Return(modelApplication, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				svc.On("CleanupAfterUnregisteringApplication", context.TODO(), appID).Return(nil, nil).Once()
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", context.TODO(), pkgmodel.ApplicationReference, modelApplication.ID).Return(nil, testErr)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Return error when removing oauth from hydra",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), appID.String()).Return(modelApplication, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				svc.On("CleanupAfterUnregisteringApplication", context.TODO(), appID).Return(nil, nil).Once()
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", context.TODO(), pkgmodel.ApplicationReference, modelApplication.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", context.TODO(), testAuths).Return(testErr)
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		}, {
			Name:            "Returns error when removing default eventing labels",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), appID.String()).Return(modelApplication, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				svc.On("CleanupAfterUnregisteringApplication", context.TODO(), appID).Return(nil, testErr).Once()
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			eventingSvc := testCase.EventingSvcFn()
			persistTx, transact := testCase.TransactionerFn()
			sysAuthSvc := testCase.SysAuthServiceFn()
			oAuth20Svc := testCase.OAuth20ServiceFn()
			resolver := application.NewResolver(transact, svc, nil, oAuth20Svc, sysAuthSvc, nil, nil, nil, eventingSvc, nil, nil, nil, "", "")
			resolver.SetConverter(converter)

			// WHEN
			result, err := resolver.UnregisterApplication(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			if testCase.ExpectedErr != nil {
				assert.EqualError(t, testCase.ExpectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, svc, converter, persistTx, transact, sysAuthSvc, oAuth20Svc, eventingSvc)
		})
	}
}

func TestResolver_UnpairApplication(t *testing.T) {
	// GIVEN
	appID := uuid.New()
	modelApplication := fixModelApplication(appID.String(), "tenant-foo", "Foo", "Bar")
	gqlApplication := fixGQLApplication(appID.String(), "Foo", "Bar")
	testErr := errors.New("Test error")
	testAuths := fixOAuths()
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		EventingSvcFn       func() *automock.EventingService
		SysAuthServiceFn    func() *automock.SystemAuthService
		OAuth20ServiceFn    func() *automock.OAuth20Service
		InputID             string
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), appID.String()).Return(modelApplication, nil).Once()
				svc.On("Unpair", context.TODO(), appID.String()).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", context.TODO(), pkgmodel.ApplicationReference, modelApplication.ID).Return(testAuths, nil).Once()
				svc.On("DeleteMultipleByIDForObject", context.TODO(), testAuths).Return(nil).Once()
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", context.TODO(), testAuths).Return(nil).Once()
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name:            "Returns error when application unpairing failed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.AssertNotCalled(t, "Get")
				svc.On("Unpair", context.TODO(), appID.String()).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.AssertNotCalled(t, "DeleteMultipleByIDForObject")
				svc.AssertNotCalled(t, "ListForObject")
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.AssertNotCalled(t, "DeleteMultipleClientCredentials")

				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), appID.String()).Return(nil, testErr).Once()
				svc.On("Unpair", context.TODO(), appID.String()).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.AssertNotCalled(t, "DeleteMultipleByIDForObject")
				svc.AssertNotCalled(t, "ListForObject")
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.AssertNotCalled(t, "DeleteMultipleClientCredentials")
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Return error when listing all auths failed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), appID.String()).Return(modelApplication, nil).Once()
				svc.On("Unpair", context.TODO(), appID.String()).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.AssertNotCalled(t, "DeleteMultipleClientCredentials")
				svc.On("ListForObject", context.TODO(), pkgmodel.ApplicationReference, modelApplication.ID).Return(nil, testErr)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.AssertNotCalled(t, "DeleteMultipleClientCredentials")
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Return error when removing oauth from hydra",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), appID.String()).Return(modelApplication, nil).Once()
				svc.On("Unpair", context.TODO(), appID.String()).Return(nil).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")

				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("DeleteMultipleByIDForObject", context.TODO(), testAuths).Return(nil).Once()
				svc.On("ListForObject", context.TODO(), pkgmodel.ApplicationReference, modelApplication.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", context.TODO(), testAuths).Return(testErr)
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Return error when removing system auths",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), appID.String()).Return(modelApplication, nil).Once()
				svc.On("Unpair", context.TODO(), appID.String()).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("DeleteMultipleByIDForObject", context.TODO(), testAuths).Return(testErr).Once()
				svc.On("ListForObject", context.TODO(), pkgmodel.ApplicationReference, modelApplication.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.AssertNotCalled(t, "DeleteMultipleClientCredentials")
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			persistTx, transact := testCase.TransactionerFn()
			sysAuthSvc := testCase.SysAuthServiceFn()
			oAuth20Svc := testCase.OAuth20ServiceFn()
			resolver := application.NewResolver(transact, svc, nil, oAuth20Svc, sysAuthSvc, nil, nil, nil, nil, nil, nil, nil, "", "")
			resolver.SetConverter(converter)

			// WHEN
			result, err := resolver.UnpairApplication(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			if testCase.ExpectedErr != nil {
				assert.EqualError(t, testCase.ExpectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, svc, converter, persistTx, transact, sysAuthSvc, oAuth20Svc)
		})
	}
}

func TestResolver_MergeApplications(t *testing.T) {
	// GIVEN
	srcAppID := "srcID"
	destAppID := "destID"

	modelApplication := fixModelApplication(destAppID, "tenant-foo", "Foo", "Lorem ipsum")
	gqlApplication := fixGQLApplication(destAppID, "Foo", "Lorem ipsum")

	testErr := errors.New("Test error")

	testCases := []struct {
		Name                   string
		PersistenceFn          func() *persistenceautomock.PersistenceTx
		TransactionerFn        func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn              func() *automock.ApplicationService
		ApplicationConverterFn func() *automock.ApplicationConverter
		ExpectedResult         *graphql.Application
		ExpectedErr            error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ApplicationConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()

				return conv
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Merge", txtest.CtxWithDBMatcher(), destAppID, srcAppID).Return(modelApplication, nil).Once()

				return svc
			},
			ExpectedResult: gqlApplication,
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when webhook conversion to graphql fails",
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(nil, testErr).Once()
				return transact
			},
			PersistenceFn: txtest.PersistenceContextThatDoesntExpectCommit,
			ApplicationConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")

				return conv
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.AssertNotCalled(t, "Merge")

				return svc
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name: "Returns error on committing transaction",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(testErr).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ApplicationConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")

				return conv
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Merge", txtest.CtxWithDBMatcher(), destAppID, srcAppID).Return(modelApplication, nil).Once()

				return svc
			},
			ExpectedErr: testErr,
		},
		{
			Name:          "Returns error then Merge fails",
			PersistenceFn: txtest.PersistenceContextThatDoesntExpectCommit,
			ApplicationConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")

				return conv
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Merge", txtest.CtxWithDBMatcher(), destAppID, srcAppID).Return(nil, testErr).Once()

				return svc
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ExpectedResult:  nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ApplicationConverterFn()

			mockPersistence := testCase.PersistenceFn()
			mockTransactioner := testCase.TransactionerFn(mockPersistence)

			resolver := application.NewResolver(mockTransactioner, svc, nil, nil, nil, converter, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			result, err := resolver.MergeApplications(context.TODO(), destAppID, srcAppID)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			mockPersistence.AssertExpectations(t)
			mockTransactioner.AssertExpectations(t)
		})
	}
}

func TestResolver_ApplicationBySystemNumber(t *testing.T) {
	// GIVEN
	systemNumber := "18"
	modelApplication := fixModelApplication("foo", "tenant-foo", appName, "Bar")
	gqlApplication := fixGQLApplication("foo", appName, "Bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name                string
		PersistenceFn       func() *persistenceautomock.PersistenceTx
		TransactionerFn     func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		SystemNumber        string
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetBySystemNumber", contextParam, systemNumber).Return(modelApplication, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			SystemNumber:        systemNumber,
			ExpectedApplication: gqlApplication,
		},
		{
			Name:            "GetBySystemNumber returns NotFound error",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetBySystemNumber", contextParam, systemNumber).Return(nil, apperrors.NewNotFoundError(resource.Application, "foo")).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			SystemNumber:        systemNumber,
			ExpectedApplication: nil,
			ExpectedErr:         nil,
		},
		{
			Name:            "GetBySystemNumber returns error",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetBySystemNumber", contextParam, systemNumber).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			SystemNumber:        systemNumber,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:          "Returns error when starting transaction failed",
			PersistenceFn: txtest.PersistenceContextThatExpectsCommit,
			ServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(nil, testErr).Once()
				return transact
			},
			SystemNumber:        systemNumber,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name: "Returns error when transaction commit failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(testErr).Once()
				return persistTx
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetBySystemNumber", contextParam, systemNumber).Return(modelApplication, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			TransactionerFn:     txtest.TransactionerThatSucceeds,
			SystemNumber:        systemNumber,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")
			resolver.SetConverter(converter)

			// WHEN
			result, err := resolver.ApplicationBySystemNumber(context.TODO(), testCase.SystemNumber)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Application(t *testing.T) {
	// GIVEN
	modelApplication := fixModelApplication("foo", "tenant-foo", "Foo", "Bar")
	gqlApplication := fixGQLApplication("foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name                string
		PersistenceFn       func() *persistenceautomock.PersistenceTx
		TransactionerFn     func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		InputID             string
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, "foo").Return(modelApplication, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name:            "Success returns nil when application not found",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, "foo").Return(nil, apperrors.NewNotFoundError(resource.Application, "foo")).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: nil,
			ExpectedErr:         nil,
		},
		{
			Name:            "Returns error when application retrieval failed",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")
			resolver.SetConverter(converter)

			// WHEN
			result, err := resolver.Application(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Applications(t *testing.T) {
	// GIVEN
	modelApplications := []*model.Application{
		fixModelApplication("foo", "tenant-foo", "Foo", "Lorem Ipsum"),
		fixModelApplication("bar", "tenant-bar", "Bar", "Lorem Ipsum"),
	}

	gqlApplications := []*graphql.Application{
		fixGQLApplication("foo", "Foo", "Lorem Ipsum"),
		fixGQLApplication("bar", "Bar", "Lorem Ipsum"),
	}

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"
	query := "foo"
	filter := []*labelfilter.LabelFilter{
		{Key: "", Query: &query},
	}
	gqlFilter := []*graphql.LabelFilter{
		{Key: "", Query: &query},
	}
	testErr := errors.New("Test error")

	testCases := []struct {
		Name              string
		PersistenceFn     func() *persistenceautomock.PersistenceTx
		TransactionerFn   func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn         func() *automock.ApplicationService
		ConverterFn       func() *automock.ApplicationConverter
		InputLabelFilters []*graphql.LabelFilter
		ExpectedResult    *graphql.ApplicationPage
		ExpectedErr       error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("List", contextParam, filter, first, after).Return(fixApplicationPage(modelApplications), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("MultipleToGraphQL", modelApplications).Return(gqlApplications).Once()
				return conv
			},
			InputLabelFilters: gqlFilter,
			ExpectedResult:    fixGQLApplicationPage(gqlApplications),
			ExpectedErr:       nil,
		},
		{
			Name:            "Returns error when application listing failed",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("List", contextParam, filter, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputLabelFilters: gqlFilter,
			ExpectedResult:    nil,
			ExpectedErr:       testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")
			resolver.SetConverter(converter)

			// WHEN
			ctx := consumer.SaveToContext(context.TODO(), consumer.Consumer{ConsumerID: "testConsumerID"})
			result, err := resolver.Applications(ctx, testCase.InputLabelFilters, &first, &gqlAfter)

			// then
			assert.Equal(t, testCase.ExpectedErr, err)
			assert.Equal(t, testCase.ExpectedResult, result)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Applications_DoubleAuthFlow(t *testing.T) {
	// GIVEN
	appTemplateID := "12345678-ae7e-4d1a-8027-520a96d5319d"

	modelApplication := fixModelApplication("foo", "tenant-foo", "Foo", "Bar")
	gqlApplication := fixGQLApplication("foo", "Foo", "Bar")

	modelApplication.ApplicationTemplateID = &appTemplateID
	gqlApplication.ApplicationTemplateID = &appTemplateID

	modelApplicationList := []*model.Application{
		modelApplication,
	}
	modelApplicationListWithNoMatchingRecord := []*model.Application{
		fixModelApplication("foo", "tenant-foo", "Foo", "Bar"),
	}

	consumerID := "abcd1122-ae7e-4d1a-8027-520a96d5319d"
	onBehalfOf := "a9653128-gs3e-4d1a-8sdj-52a96dd5301d"
	region := "eu-1"
	tokenClientID := "sb-token-client-id"
	strippedTokenClientID := "token-client-id"
	selfRegisterDistinguishLabelKey := "test-distinguish-label-key"

	certConsumer := consumer.Consumer{
		ConsumerID:    consumerID,
		ConsumerType:  consumer.ExternalCertificate,
		Flow:          oathkeeper.CertificateFlow,
		OnBehalfOf:    onBehalfOf,
		Region:        region,
		TokenClientID: tokenClientID,
	}
	ctxWithConsumerInfo := consumer.SaveToContext(context.TODO(), certConsumer)

	appTmplFilters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(scenarioassignment.SubaccountIDKey, fmt.Sprintf("\"%s\"", consumerID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
		labelfilter.NewForKeyWithQuery(selfRegisterDistinguishLabelKey, fmt.Sprintf("\"%s\"", strippedTokenClientID)),
	}

	appTemplate := fixModelApplicationTemplate(appTemplateID, "app-template")

	testErr := errors.New("Test error")

	testCases := []struct {
		Name                    string
		PersistenceFn           func() *persistenceautomock.PersistenceTx
		TransactionerFn         func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn               func() *automock.ApplicationService
		AppTemplateServiceFn    func() *automock.ApplicationTemplateService
		ConverterFn             func() *automock.ApplicationConverter
		Context                 context.Context
		ExpectedApplicationPage *graphql.ApplicationPage
		ExpectedErr             error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", contextParam).Return(modelApplicationList, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", contextParam, appTmplFilters).Return(appTemplate, nil).Once()

				return svc
			},
			Context: ctxWithConsumerInfo,
			ExpectedApplicationPage: &graphql.ApplicationPage{
				Data:       []*graphql.Application{gqlApplication},
				TotalCount: 1,
				PageInfo: &graphql.PageInfo{
					StartCursor: "1",
					EndCursor:   "1",
					HasNextPage: false,
				},
			},
			ExpectedErr: nil,
		},
		{
			Name:            "Error when no consumer is found in the context",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.NoopTransactioner,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.AssertNotCalled(t, "ListAll")

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.AssertNotCalled(t, "GetByFilters")

				return svc
			},
			Context:     context.TODO(),
			ExpectedErr: errors.New("cannot read consumer from context"),
		},
		{
			Name:            "Error when getting application template",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.AssertNotCalled(t, "ListAll")

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", contextParam, appTmplFilters).Return(nil, testErr).Once()

				return svc
			},
			Context:     ctxWithConsumerInfo,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when listing applications template",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", contextParam).Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", contextParam, appTmplFilters).Return(appTemplate, nil).Once()

				return svc
			},
			Context:     ctxWithConsumerInfo,
			ExpectedErr: testErr,
		},
		{
			Name:            "Error when no application found",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", contextParam).Return(modelApplicationListWithNoMatchingRecord, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.AssertNotCalled(t, "ToGraphQL")
				return conv
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", contextParam, appTmplFilters).Return(appTemplate, nil).Once()

				return svc
			},
			Context:     ctxWithConsumerInfo,
			ExpectedErr: errors.New("No application found for template with ID \"12345678-ae7e-4d1a-8027-520a96d5319d\""),
		},
		{
			Name: "Error when committing",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(testErr).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", contextParam).Return(modelApplicationList, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", contextParam, appTmplFilters).Return(appTemplate, nil).Once()

				return svc
			},
			Context:     ctxWithConsumerInfo,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			appTemplateSvc := testCase.AppTemplateServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, appTemplateSvc, selfRegisterDistinguishLabelKey, "sb-")
			resolver.SetConverter(converter)

			first := 2
			gqlAfter := graphql.PageCursor("test")
			query := "foo"
			gqlFilter := []*graphql.LabelFilter{
				{Key: "", Query: &query},
			}

			// WHEN
			result, err := resolver.Applications(testCase.Context, gqlFilter, &first, &gqlAfter)

			// then
			assert.Equal(t, testCase.ExpectedApplicationPage, result)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, svc, converter, appTemplateSvc)
		})
	}
}

func TestResolver_ApplicationsForRuntime(t *testing.T) {
	testError := errors.New("test error")

	modelApplications := []*model.Application{
		fixModelApplication("id1", "tenant-foo", "name", "desc"),
		fixModelApplication("id2", "tenant-bar", "name", "desc"),
	}

	applicationGraphQL := []*graphql.Application{
		fixGQLApplication("id1", "name", "desc"),
		fixGQLApplication("id2", "name", "desc"),
	}

	first := 10
	after := "test"
	gqlAfter := graphql.PageCursor(after)

	txGen := txtest.NewTransactionContextGenerator(testError)

	runtimeUUID := uuid.New()
	runtimeID := runtimeUUID.String()
	testCases := []struct {
		Name            string
		AppConverterFn  func() *automock.ApplicationConverter
		AppServiceFn    func() *automock.ApplicationService
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		InputRuntimeID  string
		ExpectedResult  *graphql.ApplicationPage
		ExpectedError   error
	}{
		{
			Name: "Success",
			AppServiceFn: func() *automock.ApplicationService {
				appService := &automock.ApplicationService{}
				appService.On("ListByRuntimeID", contextParam, runtimeUUID, first, after).Return(fixApplicationPage(modelApplications), nil).Once()
				return appService
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConverter := &automock.ApplicationConverter{}
				appConverter.On("MultipleToGraphQL", modelApplications).Return(applicationGraphQL).Once()
				return appConverter
			},
			TransactionerFn: txGen.ThatSucceeds,
			InputRuntimeID:  runtimeID,
			ExpectedResult:  fixGQLApplicationPage(applicationGraphQL),
			ExpectedError:   nil,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			AppServiceFn: func() *automock.ApplicationService {
				appService := &automock.ApplicationService{}
				appService.On("ListByRuntimeID", contextParam, runtimeUUID, first, after).Return(fixApplicationPage(modelApplications), nil).Once()
				return appService
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConverter := &automock.ApplicationConverter{}
				return appConverter
			},
			InputRuntimeID: runtimeID,
			ExpectedResult: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when application listing failed",
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListByRuntimeID", contextParam, runtimeUUID, first, after).Return(nil, testError).Once()
				return appSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConverter := &automock.ApplicationConverter{}
				return appConverter
			},
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputRuntimeID:  runtimeID,
			ExpectedResult:  nil,
			ExpectedError:   testError,
		},
		{
			Name: "Returns error when starting transaction failed",
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConverter := &automock.ApplicationConverter{}
				return appConverter
			},
			TransactionerFn: txGen.ThatFailsOnBegin,
			InputRuntimeID:  runtimeID,
			ExpectedResult:  nil,
			ExpectedError:   testError,
		},
		{
			Name: "Returns error when runtimeID is not UUID",
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConverter := &automock.ApplicationConverter{}
				return appConverter
			},
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputRuntimeID:  "blabla",
			ExpectedResult:  nil,
			ExpectedError:   errors.New("invalid UUID length"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			applicationSvc := testCase.AppServiceFn()
			applicationConverter := testCase.AppConverterFn()
			persistTx, transact := testCase.TransactionerFn()

			resolver := application.NewResolver(transact, applicationSvc, nil, nil, nil, applicationConverter, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			result, err := resolver.ApplicationsForRuntime(context.TODO(), testCase.InputRuntimeID, &first, &gqlAfter)

			// THEN
			if testCase.ExpectedError != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedResult, result)
			applicationSvc.AssertExpectations(t)
			applicationConverter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
			transact.AssertExpectations(t)
		})
	}
}

func TestResolver_SetApplicationLabel(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	applicationID := "foo"
	gqlLabel := &graphql.Label{
		Key:   "key",
		Value: []string{"foo", "bar"},
	}
	modelLabel := &model.LabelInput{
		Key:        "key",
		Value:      []string{"foo", "bar"},
		ObjectID:   applicationID,
		ObjectType: model.ApplicationLabelableObject,
	}

	testCases := []struct {
		Name               string
		PersistenceFn      func() *persistenceautomock.PersistenceTx
		TransactionerFn    func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn          func() *automock.ApplicationService
		ConverterFn        func() *automock.ApplicationConverter
		InputApplicationID string
		InputKey           string
		InputValue         interface{}
		ExpectedLabel      *graphql.Label
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("SetLabel", contextParam, modelLabel).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			InputValue:         gqlLabel.Value,
			ExpectedLabel:      gqlLabel,
			ExpectedErr:        nil,
		},
		{
			Name:            "Returns error when adding label to application failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("SetLabel", contextParam, modelLabel).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			InputValue:         gqlLabel.Value,
			ExpectedLabel:      nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			persistTx := testCase.PersistenceFn()
			transactioner := testCase.TransactionerFn(persistTx)

			resolver := application.NewResolver(transactioner, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")
			resolver.SetConverter(converter)

			// WHEN
			result, err := resolver.SetApplicationLabel(context.TODO(), testCase.InputApplicationID, testCase.InputKey, testCase.InputValue)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}

	t.Run("Returns error when Label input validation failed", func(t *testing.T) {
		resolver := application.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

		// WHEN
		result, err := resolver.SetApplicationLabel(context.TODO(), "", "", "")

		// then
		require.Nil(t, result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "key=cannot be blank")
		assert.Contains(t, err.Error(), "value=cannot be blank")
		assert.Contains(t, err.Error(), "validation error for type LabelInput:")
	})
}

func TestResolver_DeleteApplicationLabel(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	applicationID := "foo"

	labelKey := "key"

	gqlLabel := &graphql.Label{
		Key:   labelKey,
		Value: []string{"foo", "bar"},
	}

	modelLabel := &model.Label{
		ID:         "b39ba24d-87fe-43fe-ac55-7f2e5ee04bcb",
		Tenant:     str.Ptr("tnt"),
		Key:        labelKey,
		Value:      []string{"foo", "bar"},
		ObjectID:   applicationID,
		ObjectType: model.ApplicationLabelableObject,
	}

	testCases := []struct {
		Name               string
		PersistenceFn      func() *persistenceautomock.PersistenceTx
		TransactionerFn    func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn          func() *automock.ApplicationService
		ConverterFn        func() *automock.ApplicationConverter
		InputApplicationID string
		InputKey           string
		ExpectedLabel      *graphql.Label
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetLabel", contextParam, applicationID, labelKey).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", contextParam, applicationID, labelKey).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			ExpectedLabel:      gqlLabel,
			ExpectedErr:        nil,
		},
		{
			Name:            "Returns error when label retrieval failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetLabel", contextParam, applicationID, labelKey).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			ExpectedLabel:      nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Returns error when deleting application's label failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetLabel", contextParam, applicationID, labelKey).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", contextParam, applicationID, labelKey).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			ExpectedLabel:      nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			persistTx := testCase.PersistenceFn()
			transactioner := testCase.TransactionerFn(persistTx)

			resolver := application.NewResolver(transactioner, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")
			resolver.SetConverter(converter)

			// WHEN
			result, err := resolver.DeleteApplicationLabel(context.TODO(), testCase.InputApplicationID, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transactioner.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_Webhooks(t *testing.T) {
	// GIVEN
	applicationID := "fooid"
	modelWebhooks := []*model.Webhook{
		fixModelWebhook(applicationID, "foo"),
		fixModelWebhook(applicationID, "bar"),
	}
	gqlWebhooks := []*graphql.Webhook{
		fixGQLWebhook("foo"),
		fixGQLWebhook("bar"),
	}
	app := fixGQLApplication(applicationID, "foo", "bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name               string
		PersistenceFn      func() *persistenceautomock.PersistenceTx
		TransactionerFn    func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn          func() *automock.WebhookService
		WebhookConverterFn func() *automock.WebhookConverter
		ExpectedResult     []*graphql.Webhook
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListAllApplicationWebhooks", contextParam, applicationID).Return(modelWebhooks, nil).Once()
				return svc
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks).Return(gqlWebhooks, nil).Once()
				return conv
			},
			ExpectedResult: gqlWebhooks,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when webhook listing failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListAllApplicationWebhooks", contextParam, applicationID).Return(nil, testErr).Once()
				return svc
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when webhook conversion to graphql fails",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListAllApplicationWebhooks", contextParam, applicationID).Return(modelWebhooks, nil).Once()
				return svc
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks).Return(nil, testErr).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name: "Returns error on starting transaction",
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(nil, testErr).Once()
				return transact
			},
			PersistenceFn: txtest.PersistenceContextThatDoesntExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error on committing transaction",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(testErr).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListAllApplicationWebhooks", contextParam, applicationID).Return(modelWebhooks, nil).Once()
				return svc
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks).Return(gqlWebhooks, nil).Once()
				return conv
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Webhook service returns not found error",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListAllApplicationWebhooks", contextParam, applicationID).Return(nil, apperrors.NewNotFoundError(resource.Webhook, "foo")).Once()
				return svc
			},
			WebhookConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.WebhookConverterFn()

			mockPersistence := testCase.PersistenceFn()
			mockTransactioner := testCase.TransactionerFn(mockPersistence)

			resolver := application.NewResolver(mockTransactioner, nil, svc, nil, nil, nil, converter, nil, nil, nil, nil, nil, "", "")

			// WHEN
			result, err := resolver.Webhooks(context.TODO(), app)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			mockPersistence.AssertExpectations(t)
			mockTransactioner.AssertExpectations(t)
		})
	}
}

func TestResolver_Labels(t *testing.T) {
	// GIVEN

	id := "foo"
	tenant := "tenant"
	labelKey1 := "key1"
	labelValue1 := "val1"
	labelKey2 := "key2"
	labelValue2 := "val2"

	gqlApp := fixGQLApplication(id, "name", "desc")

	modelLabels := map[string]*model.Label{
		"abc": {
			ID:         "abc",
			Tenant:     str.Ptr(tenant),
			Key:        labelKey1,
			Value:      labelValue1,
			ObjectID:   id,
			ObjectType: model.ApplicationLabelableObject,
		},
		"def": {
			ID:         "def",
			Tenant:     str.Ptr(tenant),
			Key:        labelKey2,
			Value:      labelValue2,
			ObjectID:   id,
			ObjectType: model.ApplicationLabelableObject,
		},
	}

	gqlLabels := graphql.Labels{
		labelKey1: labelValue1,
		labelKey2: labelValue2,
	}

	gqlLabels1 := graphql.Labels{
		labelKey1: labelValue1,
	}

	testErr := errors.New("Test error")

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.ApplicationService
		InputApp        *graphql.Application
		InputKey        *string
		ExpectedResult  graphql.Labels
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListLabels", contextParam, id).Return(modelLabels, nil).Once()
				return svc
			},
			InputKey:       nil,
			ExpectedResult: gqlLabels,
			ExpectedErr:    nil,
		},
		{
			Name:            "Success when labels are filtered",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListLabels", contextParam, id).Return(modelLabels, nil).Once()
				return svc
			},
			InputKey:       &labelKey1,
			ExpectedResult: gqlLabels1,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when label listing failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListLabels", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			InputKey:       &labelKey1,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)

			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			result, err := resolver.Labels(context.TODO(), gqlApp, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			transact.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_Auths(t *testing.T) {
	// GIVEN
	id := "foo"
	auth := model.Auth{
		Credential: model.CredentialData{},
		OneTimeToken: &model.OneTimeToken{
			Token:     "sometoken",
			Type:      tokens.ApplicationToken,
			CreatedAt: time.Now(),
			Used:      false,
		},
	}
	gqlAuth := graphql.OneTimeTokenForApplication{
		TokenWithURL: graphql.TokenWithURL{
			Token:        auth.OneTimeToken.Token,
			ConnectorURL: auth.OneTimeToken.ConnectorURL,
		},
		LegacyConnectorURL: legacyConnectorURL,
	}
	testError := errors.New("error")
	gqlApp := fixGQLApplication(id, "name", "desc")
	txGen := txtest.NewTransactionContextGenerator(testError)

	sysAuthModels := []pkgmodel.SystemAuth{{ID: "id1", AppID: &id, Value: &auth}, {ID: "id2", AppID: &id, Value: &auth}}
	sysAuthModelCert := []pkgmodel.SystemAuth{{ID: "id1", AppID: &id, Value: nil}}
	sysAuthGQL := []*graphql.AppSystemAuth{{ID: "id1", Auth: &graphql.Auth{
		OneTimeToken: &gqlAuth,
	}}, {ID: "id2", Auth: &graphql.Auth{
		OneTimeToken: &gqlAuth,
	}}}
	sysAuthGQLCert := []*graphql.AppSystemAuth{{ID: "id1", Auth: nil}}
	sysAuthExpected := []*graphql.AppSystemAuth{{ID: "id1", Auth: &graphql.Auth{OneTimeToken: &gqlAuth}}, {ID: "id2", Auth: &graphql.Auth{OneTimeToken: &gqlAuth}}}
	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.SystemAuthService
		SysAuthConvFn   func() *automock.SystemAuthConverter
		InputApp        *graphql.Application
		ExpectedResult  []*graphql.AppSystemAuth
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), pkgmodel.ApplicationReference, id).Return(sysAuthModels, nil).Once()
				return svc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				sysAuthConv.On("ToGraphQL", &sysAuthModels[0]).Return(sysAuthGQL[0], nil).Once()
				sysAuthConv.On("ToGraphQL", &sysAuthModels[1]).Return(sysAuthGQL[1], nil).Once()
				return sysAuthConv
			},
			InputApp:       gqlApp,
			ExpectedResult: sysAuthExpected,
			ExpectedErr:    nil,
		},
		{
			Name:            "Success when System Auth is certificate",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), pkgmodel.ApplicationReference, id).Return(sysAuthModelCert, nil).Once()
				return svc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				sysAuthConv.On("ToGraphQL", &sysAuthModelCert[0]).Return(sysAuthGQLCert[0], nil).Once()
				return sysAuthConv
			},
			InputApp:       gqlApp,
			ExpectedResult: sysAuthGQLCert,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), pkgmodel.ApplicationReference, id).Return(sysAuthModels, nil).Once()
				svc.AssertNotCalled(t, "IsSystemAuthOneTimeTokenType")
				return svc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				sysAuthConv.AssertNotCalled(t, "ToGraphQL")
				return sysAuthConv
			},
			InputApp:       gqlApp,
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when list for SystemAuths failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), pkgmodel.ApplicationReference, id).Return([]pkgmodel.SystemAuth{}, testError).Once()
				svc.AssertNotCalled(t, "IsSystemAuthOneTimeTokenType")
				return svc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				sysAuthConv.AssertNotCalled(t, "IsSystemAuthOneTimeTokenType")
				return sysAuthConv
			},
			InputApp:       gqlApp,
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when conversion to graphql fails",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), pkgmodel.ApplicationReference, id).Return(sysAuthModels, nil).Once()
				svc.AssertNotCalled(t, "IsSystemAuthOneTimeTokenType")
				return svc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				sysAuthConv.On("ToGraphQL", &sysAuthModels[0]).Return(nil, testError).Once()
				return sysAuthConv
			},
			InputApp:       gqlApp,
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.AssertNotCalled(t, "IsSystemAuthOneTimeTokenType")
				return svc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			InputApp:       gqlApp,
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			persist, transact := testCase.TransactionerFn()
			conv := testCase.SysAuthConvFn()

			resolver := application.NewResolver(transact, nil, nil, nil, svc, nil, nil, conv, nil, nil, nil, nil, "", "")

			// WHEN
			result, err := resolver.Auths(context.TODO(), testCase.InputApp)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			conv.AssertExpectations(t)
			transact.AssertExpectations(t)
			persist.AssertExpectations(t)
		})
	}

	t.Run("Returns error when application is nil", func(t *testing.T) {
		resolver := application.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")
		// WHEN
		_, err := resolver.Auths(context.TODO(), nil)
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Application cannot be empty")
	})
}

func TestResolver_EventingConfiguration(t *testing.T) {
	// GIVEN
	tnt := "tnt"
	externalTnt := "ex-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	applicationID := uuid.New()
	gqlApp := fixGQLApplication(applicationID.String(), "bar", "baz")
	app := fixModelApplication(applicationID.String(), tnt, "bar", "baz")

	converterMock := func() *automock.ApplicationConverter {
		converter := &automock.ApplicationConverter{}
		converter.On("GraphQLToModel", gqlApp, tnt).Return(app).Once()
		return converter
	}

	testErr := errors.New("this is a test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	defaultEveningURL := "https://eventing.domain.local"
	modelAppEventingCfg := fixModelApplicationEventingConfiguration(t, defaultEveningURL)
	gqlAppEventingCfg := fixGQLApplicationEventingConfiguration(defaultEveningURL)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		EventingSvcFn   func() *automock.EventingService
		ConverterFn     func() *automock.ApplicationConverter
		ExpectedOutput  *graphql.ApplicationEventingConfiguration
		ExpectedError   error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("GetForApplication", txtest.CtxWithDBMatcher(), *app).Return(modelAppEventingCfg, nil).Once()

				return eventingSvc
			},
			ConverterFn:    converterMock,
			ExpectedOutput: gqlAppEventingCfg,
			ExpectedError:  nil,
		}, {
			Name:            "Error when getting the configuration for runtime failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("GetForApplication", txtest.CtxWithDBMatcher(), *app).Return(nil, testErr).Once()
				return eventingSvc
			},
			ConverterFn:    converterMock,
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				return eventingSvc
			},
			ConverterFn:    converterMock,
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("GetForApplication", txtest.CtxWithDBMatcher(), *app).Return(modelAppEventingCfg, nil).Once()
				return eventingSvc
			},
			ConverterFn:    converterMock,
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			eventingSvc := testCase.EventingSvcFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, nil, nil, nil, nil, converter, nil, nil, eventingSvc, nil, nil, nil, "", "")

			// WHEN
			result, err := resolver.EventingConfiguration(ctx, gqlApp)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, eventingSvc, transact, persist, converter)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		// GIVEN
		resolver := application.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

		// WHEN
		_, err := resolver.EventingConfiguration(context.TODO(), &graphql.Application{})

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "cannot read tenant from context")
	})

	t.Run("Error when parent object is nil", func(t *testing.T) {
		// GIVEN
		resolver := application.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

		// WHEN
		result, err := resolver.EventingConfiguration(context.TODO(), nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Application cannot be empty")
		assert.Nil(t, result)
	})
}

func TestResolver_Bundles(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")

	tenantID := "1"
	firstAppID := "appID"
	secondAppID := "appID2"
	appIDs := []string{firstAppID, secondAppID}

	bundleFirstApp := fixModelBundle("foo", tenantID, firstAppID, "Foo", "Lorem Ipsum")
	bundleSecondApp := fixModelBundle("foo", tenantID, secondAppID, "Foo", "Lorem Ipsum")

	bundlesFirstApp := []*model.Bundle{bundleFirstApp}
	bundlesSecondApp := []*model.Bundle{bundleSecondApp}

	gqlBundleFirstApp := fixGQLBundle("foo", firstAppID, "Foo", "Lorem Ipsum")
	gqlBundleSecondApp := fixGQLBundle("foo", secondAppID, "Foo", "Lorem Ipsum")

	gqlBundlesFirstApp := []*graphql.Bundle{gqlBundleFirstApp}
	gqlBundlesSecondApp := []*graphql.Bundle{gqlBundleSecondApp}

	bundlePageFirstApp := fixBundlePage(bundlesFirstApp)
	bundlePageSecondApp := fixBundlePage(bundlesSecondApp)
	bundlePages := []*model.BundlePage{bundlePageFirstApp, bundlePageSecondApp}

	gqlBundlePageFirstApp := fixGQLBundlePage(gqlBundlesFirstApp)
	gqlBundlePageSecondApp := fixGQLBundlePage(gqlBundlesSecondApp)
	gqlBundlePages := []*graphql.BundlePage{gqlBundlePageFirstApp, gqlBundlePageSecondApp}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.BundleService
		ConverterFn     func() *automock.BundleConverter
		ExpectedResult  []*graphql.BundlePage
		ExpectedErr     []error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("ListByApplicationIDs", txtest.CtxWithDBMatcher(), appIDs, first, after).Return(bundlePages, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("MultipleToGraphQL", bundlesFirstApp).Return(gqlBundlesFirstApp, nil).Once()
				conv.On("MultipleToGraphQL", bundlesSecondApp).Return(gqlBundlesSecondApp, nil).Once()
				return conv
			},
			ExpectedResult: gqlBundlePages,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when Bundles listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("ListByApplicationIDs", txtest.CtxWithDBMatcher(), appIDs, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("ListByApplicationIDs", txtest.CtxWithDBMatcher(), appIDs, first, after).Return(bundlePages, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("MultipleToGraphQL", bundlesFirstApp).Return(nil, testErr).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("ListByApplicationIDs", txtest.CtxWithDBMatcher(), appIDs, first, after).Return(bundlePages, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("MultipleToGraphQL", bundlesFirstApp).Return(gqlBundlesFirstApp, nil).Once()
				conv.On("MultipleToGraphQL", bundlesSecondApp).Return(gqlBundlesSecondApp, nil).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, nil, nil, nil, nil, nil, nil, nil, nil, svc, converter, nil, "", "")
			firstAppParams := dataloader.ParamBundle{ID: firstAppID, Ctx: context.TODO(), First: &first, After: &gqlAfter}
			secondAppParams := dataloader.ParamBundle{ID: secondAppID, Ctx: context.TODO(), First: &first, After: &gqlAfter}
			keys := []dataloader.ParamBundle{firstAppParams, secondAppParams}

			// WHEN
			result, err := resolver.BundlesDataLoader(keys)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}

	t.Run("Returns error when there are no Applications", func(t *testing.T) {
		resolver := application.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")
		// WHEN
		_, err := resolver.BundlesDataLoader([]dataloader.ParamBundle{})
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInternalError("No Applications found").Error())
	})

	t.Run("Returns error when start cursor is nil", func(t *testing.T) {
		firstAppParams := dataloader.ParamBundle{ID: firstAppID, Ctx: context.TODO(), First: nil, After: &gqlAfter}
		keys := []dataloader.ParamBundle{firstAppParams}

		resolver := application.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")
		// WHEN
		_, err := resolver.BundlesDataLoader(keys)
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInvalidDataError("missing required parameter 'first'").Error())
	})
}

func TestResolver_Bundle(t *testing.T) {
	// GIVEN
	id := "foo"
	appID := "bar"
	tenantID := "baz"
	modelBundle := fixModelBundle(id, tenantID, appID, "name", "bar")
	gqlBundle := fixGQLBundle(id, appID, "name", "bar")
	app := fixGQLApplication("foo", "foo", "foo")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.BundleService
		ConverterFn     func() *automock.BundleConverter
		InputID         string
		Application     *graphql.Application
		ExpectedBundle  *graphql.Bundle
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelBundle, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			InputID:        "foo",
			Application:    app,
			ExpectedBundle: gqlBundle,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			InputID:        "foo",
			Application:    app,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when conversion to graphql fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("ToGraphQL", modelBundle).Return(nil, testErr).Once()
				return conv
			},
			InputID:        "foo",
			Application:    app,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns nil when bundle for application not found",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError(resource.Application, "")).Once()

				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			InputID:        "foo",
			Application:    app,
			ExpectedBundle: nil,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when commit begin error",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}

				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			InputID:        "foo",
			Application:    app,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			InputID:        "foo",
			Application:    app,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, nil, nil, nil, nil, nil, nil, nil, nil, svc, converter, nil, "", "")

			// WHEN
			result, err := resolver.Bundle(context.TODO(), testCase.Application, testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedBundle, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}

	t.Run("Returns error when application is nil", func(t *testing.T) {
		resolver := application.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")
		// WHEN
		_, err := resolver.Bundle(context.TODO(), nil, "")
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, apperrors.NewInternalError("Application cannot be empty").Error())
	})
}

func fixOAuths() []pkgmodel.SystemAuth {
	return []pkgmodel.SystemAuth{
		{
			ID:       "foo",
			TenantID: str.Ptr("foo"),
			Value: &model.Auth{
				Credential: model.CredentialData{
					Basic: nil,
					Oauth: &model.OAuthCredentialData{
						ClientID:     "foo",
						ClientSecret: "foo",
						URL:          "foo",
					},
				},
			},
		},
		{
			ID:       "bar",
			TenantID: str.Ptr("bar"),
			Value:    nil,
		},
		{
			ID:       "test",
			TenantID: str.Ptr("test"),
			Value: &model.Auth{
				Credential: model.CredentialData{
					Basic: &model.BasicCredentialData{
						Username: "test",
						Password: "test",
					},
					Oauth: nil,
				},
			},
		},
	}
}
